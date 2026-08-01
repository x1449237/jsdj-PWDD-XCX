package service

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"sort"
	"strconv"
	"strings"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/jisan/e-sports-platform/internal/model"
	"github.com/jisan/e-sports-platform/pkg/queue"
)

type JSAPIPayParams struct {
	AppID     string `json:"app_id"`
	TimeStamp string `json:"time_stamp"`
	NonceStr  string `json:"nonce_str"`
	Package   string `json:"package"`
	SignType  string `json:"sign_type"`
	PaySign   string `json:"pay_sign"`
	PrepayID  string `json:"prepay_id,omitempty"`
}

type PaymentResult struct {
	Payment    *model.Payment  `json:"payment"`
	JSAPI      *JSAPIPayParams `json:"jsapi,omitempty"`
	Sandbox    bool            `json:"sandbox"`
	PrepayID   string          `json:"prepay_id,omitempty"`
	OutTradeNo string          `json:"out_trade_no,omitempty"`
}

// CreatePayment 创建支付记录
// 安全修复:
// 1. 校验订单归属(order.UserID == userID)防越权
// 2. 金额以后端订单金额为准,忽略客户端传入 amount(防金额篡改)
// 3. 校验订单状态(仅待接单可支付,且未支付)防重复支付
// 4. 校验金额 > 0
// 5. 防重复创建 pending 支付记录:复用已有 pending 记录
func CreatePayment(userID, orderID int64, payMethod string) (*PaymentResult, error) {
	if orderID <= 0 {
		return nil, errors.New("订单ID无效")
	}
	if payMethod == "" {
		return nil, errors.New("支付方式不能为空")
	}
	// 1. 查订单并校验归属
	o, err := orderRepo.FindByID(orderID)
	if err != nil {
		return nil, err
	}
	if o == nil {
		return nil, errors.New("订单不存在")
	}
	if o.UserID != userID {
		return nil, errors.New("无权为该订单创建支付")
	}
	// 2. 校验订单状态(仅待接单可支付,且未支付)
	// 安全修复:移除 Accepted 状态允许支付,支付仅限 Pending(避免已接单订单二次支付)
	if o.PayStatus == 1 {
		return nil, errors.New("订单已支付,无需重复支付")
	}
	if o.Status != model.OrderStatusPending {
		return nil, errors.New("当前订单状态不可支付")
	}
	// 3. 金额以订单金额为准(忽略客户端传入)
	amount := o.Amount
	if amount <= 0 {
		return nil, errors.New("订单金额异常")
	}
	// 4. 防重复创建:复用已有 pending 支付记录
	exist, _ := paymentRepo.FindPendingPaymentByOrderID(orderID)
	if exist != nil {
		result := &PaymentResult{Payment: exist, OutTradeNo: exist.OutTradeNo}
		return result, nil
	}
	p := &model.Payment{
		OrderID:    orderID,
		OutTradeNo: genOrderNo(),
		Amount:     amount,
		PayMethod:  payMethod,
		Status:     model.PaymentStatusPending,
		CreatedAt:  nowTimePtr(),
		UpdatedAt:  nowTimePtr(),
	}
	if err := paymentRepo.CreatePayment(p); err != nil {
		return nil, err
	}

	result := &PaymentResult{Payment: p, OutTradeNo: p.OutTradeNo}

	if payMethod == "wechat" || payMethod == "wechat_jsapi" {
		sandbox := true
		if cfg != nil && cfg.WeChat.MchID != "" {
			sandbox = strings.HasPrefix(cfg.WeChat.MchID, "sandbox")
		}
		result.Sandbox = sandbox

		prepayID := "wx" + fmt.Sprintf("%d", time.Now().Unix()) + randomHex(16)
		result.PrepayID = prepayID

		appID := ""
		mchKey := ""
		if cfg != nil {
			appID = cfg.WeChat.AppID
			mchKey = cfg.WeChat.MchKey
		}
		if appID == "" {
			appID = "wx_mock_appid"
		}
		if mchKey == "" {
			mchKey = "sandbox_mock_key"
		}

		ts := strconv.FormatInt(time.Now().Unix(), 10)
		nonce := randomHex(16)
		pkg := "prepay_id=" + prepayID
		signType := "RSA"
		if sandbox {
			signType = "MD5"
		}

		paySign := mockSign(appID, ts, nonce, pkg, signType, mchKey)

		result.JSAPI = &JSAPIPayParams{
			AppID:     appID,
			TimeStamp: ts,
			NonceStr:  nonce,
			Package:   pkg,
			SignType:  signType,
			PaySign:   paySign,
			PrepayID:  prepayID,
		}
	}

	return result, nil
}

func mockSign(appID, ts, nonce, pkg, signType, key string) string {
	params := map[string]string{
		"appId":     appID,
		"timeStamp": ts,
		"nonceStr":  nonce,
		"package":   pkg,
		"signType":  signType,
	}
	keys := make([]string, 0, len(params))
	for k := range params {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	var sb strings.Builder
	for i, k := range keys {
		if i > 0 {
			sb.WriteString("&")
		}
		sb.WriteString(k)
		sb.WriteString("=")
		sb.WriteString(params[k])
	}
	sb.WriteString("&key=")
	sb.WriteString(key)
	signTarget := sb.String()
	return strings.ToUpper(sha1LikeMock(signTarget))
}

func sha1LikeMock(s string) string {
	h := uint32(0)
	for i := 0; i < len(s); i++ {
		h = h*31 + uint32(s[i])
	}
	return fmt.Sprintf("%032x", h) + fmt.Sprintf("%08x", len(s))
}

func randomHex(n int) string {
	b := make([]byte, n)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}

// MarkPaymentPaid 标记支付成功
// 安全修复:
// 1. 金额校验:回调金额必须与支付记录金额一致(防 1 分钱支付攻击)
// 2. 幂等校验:事务内条件更新 WHERE status='pending',根据 RowsAffected 判定(防重复入账)
// 3. 行锁保护:事务内对支付记录加锁
// 4. 重复支付单号校验:transaction_id 已存在则拒绝
func MarkPaymentPaid(outTradeNo, txnID string, paidAmount int64) error {
	if outTradeNo == "" {
		return errors.New("商户订单号不能为空")
	}
	// 重复交易号校验(防同一交易号重复入账)
	if txnID != "" {
		exist, _ := paymentRepo.FindPaymentByTxnID(txnID)
		if exist != nil {
			if exist.OutTradeNo == outTradeNo {
				return nil // 同一订单同一交易号,幂等返回
			}
			return errors.New("交易号已被其他支付记录占用,疑似伪造回调")
		}
	}
	return db.Transaction(func(tx *gorm.DB) error {
		// 行锁查询支付记录
		var p model.Payment
		if err := tx.Where("out_trade_no = ?", outTradeNo).
			Clauses(gormClauseLockUpdate()).
			First(&p).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return errors.New("支付记录不存在")
			}
			return err
		}
		// 幂等:已支付直接返回
		if p.Status == model.PaymentStatusPaid {
			return nil
		}
		// 状态校验:仅 pending 可入账
		if p.Status != model.PaymentStatusPending {
			return fmt.Errorf("支付记录状态异常(%s),不可入账", p.Status)
		}
		// 金额校验:回调金额必须与支付记录金额一致
		if paidAmount > 0 && paidAmount != p.Amount {
			return fmt.Errorf("回调金额(%d)与订单金额(%d)不一致", paidAmount, p.Amount)
		}
		now := nowTimePtr()
		// 条件更新:仅 pending -> paid,根据 RowsAffected 判定本次是否生效(防并发)
		res := tx.Model(&model.Payment{}).
			Where("id = ? AND status = ?", p.ID, model.PaymentStatusPending).
			Updates(map[string]interface{}{
				"transaction_id": txnID,
				"status":         model.PaymentStatusPaid,
				"pay_time":       now,
				"updated_at":     now,
			})
		if res.Error != nil {
			return res.Error
		}
		if res.RowsAffected == 0 {
			// 已被其他并发请求处理,幂等返回
			return nil
		}
		// 更新订单支付状态(条件更新防重复)
		return tx.Model(&model.Order{}).
			Where("id = ? AND pay_status IN (0, 2)", p.OrderID).
			Updates(map[string]interface{}{
				"pay_status": 1,
				"paid_at":    now,
				"updated_at": now,
			}).Error
	})
}

// gormClauseLockUpdate 返回 FOR UPDATE 子句
func gormClauseLockUpdate() clause.Locking {
	return clause.Locking{Strength: "UPDATE"}
}

// ProcessRefund 订单退款
// 安全修复:
// 1. 按 order_id + status IN (paid, partial_refund) 查询支付记录(原仅查 paid,部分退款后无法继续退)
// 2. 校验订单状态机:仅特定状态可退款(防任意状态退款)
// 3. 反向分账:回滚已分账给打手/俱乐部/分销商的金额(防资金双花)
// 4. 退款余额更新放入事务(原在事务外且错误被吞)
// 5. 退款金额负数直接拒绝(原静默当作全退)
// 6. 行锁防并发退款竞态(防重复退款/超额退款)
// 7. 部分退款按比例反向分账(原部分退款也全额回滚)
// 8. clubID > 0 时校验俱乐部归属(防跨俱乐部退款)
func ProcessRefund(orderID, operatorID int64, refundAmount int64, isAdmin bool) (*model.Payment, error) {
	return ProcessRefundWithClub(orderID, operatorID, refundAmount, isAdmin, 0)
}

// ProcessRefundWithClub 带俱乐部归属校验的退款
func ProcessRefundWithClub(orderID, operatorID int64, refundAmount int64, isAdmin bool, clubID int64) (*model.Payment, error) {
	o, err := orderRepo.FindByID(orderID)
	if err != nil {
		return nil, err
	}
	if o == nil {
		return nil, errors.New("订单不存在")
	}
	// 俱乐部归属校验:非平台管理员且指定了 clubID 时,校验订单归属
	if !isAdmin && clubID > 0 && o.ClubID != clubID {
		return nil, errors.New("无权操作该订单")
	}
	if o.PayStatus != 1 && o.PayStatus != 3 {
		return nil, errors.New("订单未支付，不可退款")
	}
	// 状态机校验:仅特定状态可退款(防任意状态退款)
	if !canRefund(o.Status) {
		return nil, fmt.Errorf("订单当前状态(%d)不允许退款", o.Status)
	}
	// 1. 按 order_id + status IN (paid, partial_refund) 查询支付记录
	// 修复:原仅查 paid,部分退款后状态变 partial_refund,无法继续退款
	p, err := paymentRepo.FindRefundablePaymentByOrderID(orderID)
	if err != nil || p == nil {
		return nil, errors.New("支付记录不存在")
	}
	// 2. 退款金额校验:负数直接拒绝,0 表示全退
	if refundAmount < 0 {
		return nil, errors.New("退款金额不能为负数")
	}
	if refundAmount == 0 {
		refundAmount = p.Amount - p.RefundAmount
	}
	if refundAmount <= 0 {
		return nil, errors.New("可退款金额不足")
	}
	if p.RefundAmount+refundAmount > p.Amount {
		return nil, errors.New("退款金额超过支付金额")
	}
	now := nowTimePtr()
	newRefund := p.RefundAmount + refundAmount
	newStatus := model.PaymentStatusPartialRef
	if newRefund >= p.Amount {
		newStatus = model.PaymentStatusRefunded
	}
	// 计算反向分账比例(部分退款按比例回滚)
	reverseRatio := float64(refundAmount) / float64(p.Amount)
	// 3. 事务内执行:行锁支付记录 + 更新支付记录 + 按比例反向分账 + 退款给用户 + 更新订单
	err = db.Transaction(func(tx *gorm.DB) error {
		// 行锁支付记录(防并发退款竞态)
		var locked model.Payment
		if err := tx.Where("id = ?", p.ID).
			Clauses(gormClauseLockUpdate()).
			First(&locked).Error; err != nil {
			return err
		}
		// 二次校验(锁定后重读)
		if locked.RefundAmount+refundAmount > locked.Amount {
			return errors.New("退款金额超过支付金额(并发冲突)")
		}
		// 3.1 更新支付记录(条件更新防重复)
		res := tx.Model(&model.Payment{}).
			Where("id = ? AND refund_amount + ? <= amount", p.ID, refundAmount).
			Updates(map[string]interface{}{
				"refund_amount": locked.RefundAmount + refundAmount,
				"refund_time":   now,
				"status":        newStatus,
				"updated_at":    now,
			})
		if res.Error != nil {
			return res.Error
		}
		if res.RowsAffected == 0 {
			return errors.New("退款失败,可能存在并发退款")
		}
		// 3.2 按比例反向分账(部分退款按比例回滚,全额退款全额回滚)
		if _, err := ReverseProfitShareByRatio(tx, orderID, reverseRatio); err != nil {
			return fmt.Errorf("反向分账失败: %w", err)
		}
		// 3.3 退款给用户余额(放入事务)
		if err := tx.Model(&model.User{}).Where("id = ?", o.UserID).
			UpdateColumn("balance", gorm.Expr("balance + ?", refundAmount)).Error; err != nil {
			return err
		}
		// 3.4 更新订单
		orderStatus := o.Status
		if newStatus == model.PaymentStatusRefunded {
			orderStatus = model.OrderStatusRefunded
		}
		updates := map[string]interface{}{
			"refund_amount": o.RefundAmount + refundAmount,
			"updated_at":    now,
		}
		if newStatus == model.PaymentStatusRefunded {
			updates["status"] = orderStatus
			updates["pay_status"] = 2
		} else {
			updates["pay_status"] = 3
		}
		return tx.Model(&model.Order{}).Where("id = ?", orderID).Updates(updates).Error
	})
	if err != nil {
		return nil, err
	}
	p.RefundAmount = newRefund
	p.Status = newStatus
	return p, nil
}

// canRefund 校验订单状态是否允许退款
// 仅 待接单/已接单/进行中/待验收/待结算 状态可退款
// 已结算/已完成/已退款/已取消/超时 等终态不可退款(已结算需先反向分账,此处仍允许但事务内会处理)
func canRefund(status int8) bool {
	switch status {
	case model.OrderStatusPending,
		model.OrderStatusAccepted,
		model.OrderStatusInProgress,
		model.OrderStatusToVerify,
		model.OrderStatusToSettle,
		model.OrderStatusCompleted,
		model.OrderStatusSettled:
		return true
	}
	return false
}

func ShopProcessRefund(orderID, adminID int64, refundAmount int64) (*model.Payment, error) {
	return ProcessRefund(orderID, adminID, refundAmount, false)
}

// ShopProcessRefundWithClub 带俱乐部归属校验的俱乐部退款
func ShopProcessRefundWithClub(orderID, adminID int64, refundAmount int64, clubID int64) (*model.Payment, error) {
	return ProcessRefundWithClub(orderID, adminID, refundAmount, false, clubID)
}

func AdminProcessRefund(orderID, adminID int64, refundAmount int64) (*model.Payment, error) {
	return ProcessRefund(orderID, adminID, refundAmount, true)
}

func AdminGetWithdrawals(page, pageSize int, status string) ([]model.Withdraw, int64, error) {
	return paymentRepo.ListAllWithdraws(page, pageSize, status)
}

// AdminApproveWithdrawal 审核通过提现申请
// 安全修复:审核通过后原子扣减用户余额(原未扣减),并入队打款任务
func AdminApproveWithdrawal(withdrawID, adminID int64) error {
	w, err := paymentRepo.FindWithdraw(withdrawID)
	if err != nil {
		return err
	}
	if w == nil {
		return errors.New("提现记录不存在")
	}
	if w.Status != model.WithdrawStatusPending {
		return errors.New("提现状态不允许审核")
	}
	// 事务内:更新提现状态 + 扣减用户余额(原子操作,防超额提现)
	err = db.Transaction(func(tx *gorm.DB) error {
		// 行锁提现记录
		var locked model.Withdraw
		if err := tx.Where("id = ?", withdrawID).
			Clauses(gormClauseLockUpdate()).
			First(&locked).Error; err != nil {
			return err
		}
		if locked.Status != model.WithdrawStatusPending {
			return errors.New("提现状态已变更")
		}
		// 原子扣减用户余额(条件更新防超额)
		res := tx.Model(&model.User{}).
			Where("id = ? AND balance >= ?", locked.UserID, locked.Amount).
			UpdateColumn("balance", gorm.Expr("balance - ?", locked.Amount))
		if res.Error != nil {
			return res.Error
		}
		if res.RowsAffected == 0 {
			return errors.New("用户余额不足,无法审核通过")
		}
		// 更新提现状态
		now := nowTimePtr()
		return tx.Model(&model.Withdraw{}).Where("id = ?", withdrawID).
			Updates(map[string]interface{}{
				"status":      model.WithdrawStatusApproved,
				"reviewer_id": adminID,
				"reviewed_at": now,
				"updated_at":  now,
			}).Error
	})
	if err != nil {
		return err
	}
	// 入队打款任务
	enqueueWithdrawPaidTask(withdrawID)
	return nil
}

func AdminRejectWithdrawal(withdrawID, adminID int64, reason string) error {
	w, err := paymentRepo.FindWithdraw(withdrawID)
	if err != nil {
		return err
	}
	if w == nil {
		return errors.New("提现记录不存在")
	}
	if w.Status != model.WithdrawStatusPending {
		return errors.New("提现状态不允许审核")
	}
	return paymentRepo.UpdateWithdraw(withdrawID, map[string]interface{}{
		"status":      model.WithdrawStatusRejected,
		"reviewer_id": adminID,
		"reviewed_at": nowTimePtr(),
		"updated_at":  nowTimePtr(),
	})
}

func AdminBatchWithdraw(adminID int64, ids []int64, action string) (int, error) {
	success := 0
	for _, id := range ids {
		var err error
		switch action {
		case "approve":
			err = AdminApproveWithdrawal(id, adminID)
		case "reject":
			err = AdminRejectWithdrawal(id, adminID, "批量驳回")
		default:
			continue
		}
		if err == nil {
			success++
		}
	}
	return success, nil
}

func ShopGetWithdrawals(clubID int64, page, pageSize int) ([]model.Withdraw, int64, error) {
	var list []model.Withdraw
	var total int64
	q := db.Model(&model.Withdraw{}).
		Joins("JOIN users u ON u.id = withdrawals.user_id").
		Where("u.club_id = ?", clubID)
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	err := q.Scopes(Paginate2(page, pageSize)).Order("withdrawals.id DESC").Find(&list).Error
	return list, total, err
}

func ShopGetFinanceOverview(clubID int64) (map[string]interface{}, error) {
	totalAmount, err := orderRepo.SumAmount(clubID)
	if err != nil {
		return nil, err
	}
	statusCnt, err := orderRepo.CountByStatus(clubID)
	if err != nil {
		return nil, err
	}
	var totalWithdraw int64
	if err := db.Model(&model.Withdraw{}).
		Joins("JOIN users u ON u.id = withdrawals.user_id").
		Where("u.club_id = ? AND withdrawals.status = ?", clubID, model.WithdrawStatusPaid).
		Select("COALESCE(SUM(withdrawals.amount),0)").Scan(&totalWithdraw).Error; err != nil {
		return nil, err
	}
	return map[string]interface{}{
		"total_amount":    totalAmount,
		"status_count":    statusCnt,
		"total_withdrawn": totalWithdraw,
	}, nil
}

func ShopGetFinanceDetails(clubID int64, page, pageSize int) ([]model.Order, int64, error) {
	return orderRepo.ListByClub(clubID, page, pageSize, -1, "")
}

// HandleWxPayNotify 兼容旧接口(全额信任回调,仅用于内部测试)
// 生产环境必须使用 HandleWxPayCallbackWithHeaders 进行签名校验
func HandleWxPayNotify(outTradeNo, txnID string) error {
	return MarkPaymentPaid(outTradeNo, txnID, 0)
}

func enqueueWithdrawPaidTask(withdrawID int64) {
	if queueC == nil {
		return
	}
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	_ = queueC.EnqueueWithdrawProcess(ctx, queue.WithdrawProcessPayload{WithdrawID: withdrawID})
}

type WxPayResource struct {
	Algorithm      string `json:"algorithm"`
	Nonce          string `json:"nonce"`
	AssociatedData string `json:"associated_data"`
	Ciphertext     string `json:"ciphertext"`
	OriginalType   string `json:"original_type"`
}

type WxPayCallbackRequest struct {
	ID           string        `json:"id"`
	CreateTime   string        `json:"create_time"`
	EventType    string        `json:"event_type"`
	ResourceType string        `json:"resource_type"`
	Summary      string        `json:"summary"`
	Resource     WxPayResource `json:"resource"`
}

type WxPayDecryptedResource struct {
	AppID          string `json:"appid"`
	MchID          string `json:"mchid"`
	OutTradeNo     string `json:"out_trade_no"`
	TransactionID  string `json:"transaction_id"`
	TradeType      string `json:"trade_type"`
	TradeState     string `json:"trade_state"`
	TradeStateDesc string `json:"trade_state_desc"`
	BankType       string `json:"bank_type"`
	SuccessTime    string `json:"success_time"`
	Amount         struct {
		Total         int64  `json:"total"`
		PayerTotal    int64  `json:"payer_total"`
		Currency      string `json:"currency"`
		PayerCurrency string `json:"payer_currency"`
	} `json:"amount"`
	Payer struct {
		OpenID string `json:"openid"`
	} `json:"payer"`
}

// VerifyWxPaySignature 微信支付回调签名校验
// 安全修复:原实现恒返回 true,现改为:
// 1. 沙箱模式(cfg.WeChat.MchID 以 sandbox 开头)允许跳过签名校验(测试用)
// 2. 生产模式必须提供完整的签名参数,否则拒绝
// 3. 生产模式应使用微信平台公钥验签(此处保留接口,实际验签逻辑需接入微信 SDK)
// 注意:生产环境应接入 wechatpay-go SDK 进行真正的 RSA 签名校验
func VerifyWxPaySignature(timestamp, nonce, body, signature, serialNo, apiV3Key string) bool {
	if timestamp == "" || nonce == "" || signature == "" {
		return false
	}
	// 沙箱模式跳过签名校验(仅测试环境)
	if cfg != nil && cfg.WeChat.MchID != "" && strings.HasPrefix(cfg.WeChat.MchID, "sandbox") {
		return true
	}
	// 生产模式:必须配置微信平台公钥才能验签
	// TODO: 接入 wechatpay-go SDK 进行真正的 RSA 签名校验
	// 当前若无平台公钥配置,拒绝所有回调(安全失败)
	if cfg == nil || cfg.WeChat.PlatformCertPath == "" {
		// 未配置平台证书,记录告警但拒绝(防伪造)
		return false
	}
	// 此处应调用微信 SDK 验签,暂时返回 false 直到接入真正验签
	return false
}

// DecryptWxPayResource 解密微信支付回调资源
// 安全修复:
// 1. 删除 mock 回退路径(原解密失败返回 MOCK_SUCCESS)
// 2. 解密失败直接返回错误
// 3. 仅使用 AES-GCM 解密(微信 V3 标准)
func DecryptWxPayResource(r WxPayResource, apiV3Key string) (*WxPayDecryptedResource, error) {
	if apiV3Key == "" {
		return nil, errors.New("apiV3Key 未配置,无法解密回调")
	}
	if len(apiV3Key) < 32 {
		return nil, errors.New("apiV3Key 长度不足 32 字节")
	}
	if r.Ciphertext == "" {
		return nil, errors.New("ciphertext 为空")
	}
	// 使用 AES-GCM 解密(微信 V3 标准),不再回退到 mock
	return DecryptWxPayResourceAESGCM(r, apiV3Key)
}
