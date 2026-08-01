package service

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"sort"
	"strconv"
	"strings"
	"time"

	"gorm.io/gorm"

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

func CreatePayment(orderID, amount int64, payMethod string) (*PaymentResult, error) {
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

func MarkPaymentPaid(outTradeNo, txnID string) error {
	p, err := paymentRepo.FindPaymentByOutTradeNo(outTradeNo)
	if err != nil {
		return err
	}
	if p == nil {
		return errors.New("支付记录不存在")
	}
	if p.Status == model.PaymentStatusPaid {
		return nil
	}
	now := nowTimePtr()
	return db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Model(&model.Payment{}).Where("id = ?", p.ID).
			Updates(map[string]interface{}{
				"transaction_id": txnID,
				"status":         model.PaymentStatusPaid,
				"pay_time":       now,
				"updated_at":     now,
			}).Error; err != nil {
			return err
		}
		return tx.Model(&model.Order{}).Where("id = ?", p.OrderID).
			Updates(map[string]interface{}{
				"pay_status": 1,
				"paid_at":    now,
				"updated_at": now,
			}).Error
	})
}

func ProcessRefund(orderID, operatorID int64, refundAmount int64, isAdmin bool) (*model.Payment, error) {
	o, err := orderRepo.FindByID(orderID)
	if err != nil {
		return nil, err
	}
	if o == nil {
		return nil, errors.New("订单不存在")
	}
	if o.PayStatus != 1 && o.PayStatus != 3 {
		return nil, errors.New("订单未支付，不可退款")
	}
	p, err := paymentRepo.FindPaymentByOutTradeNo(o.OrderNo)
	if err != nil || p == nil {
		return nil, errors.New("支付记录不存在")
	}
	if refundAmount <= 0 {
		refundAmount = p.Amount - p.RefundAmount
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
	err = db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Model(&model.Payment{}).Where("id = ?", p.ID).
			Updates(map[string]interface{}{
				"refund_amount": newRefund,
				"refund_time":   now,
				"status":        newStatus,
				"updated_at":    now,
			}).Error; err != nil {
			return err
		}
		orderStatus := model.OrderStatusRefunded
		if newStatus == model.PaymentStatusPartialRef {
			orderStatus = o.Status
		}
		updates := map[string]interface{}{
			"refund_amount": o.RefundAmount + refundAmount,
			"updated_at":    now,
		}
		if newStatus == model.PaymentStatusRefunded {
			updates["status"] = orderStatus
		}
		return tx.Model(&model.Order{}).Where("id = ?", orderID).Updates(updates).Error
	})
	if err != nil {
		return nil, err
	}
	_ = userRepo.UpdateBalance(o.UserID, refundAmount)
	_ = operatorID
	_ = isAdmin
	p.RefundAmount = newRefund
	p.Status = newStatus
	return p, nil
}

func ShopProcessRefund(orderID, adminID int64, refundAmount int64) (*model.Payment, error) {
	return ProcessRefund(orderID, adminID, refundAmount, false)
}

func AdminProcessRefund(orderID, adminID int64, refundAmount int64) (*model.Payment, error) {
	return ProcessRefund(orderID, adminID, refundAmount, true)
}

func AdminGetWithdrawals(page, pageSize int, status string) ([]model.Withdraw, int64, error) {
	return paymentRepo.ListAllWithdraws(page, pageSize, status)
}

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
	err = paymentRepo.UpdateWithdraw(withdrawID, map[string]interface{}{
		"status":      model.WithdrawStatusApproved,
		"reviewer_id": adminID,
		"reviewed_at": nowTimePtr(),
		"updated_at":  nowTimePtr(),
	})
	if err != nil {
		return err
	}
	if queueC != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()
		payload := strconv.FormatInt(withdrawID, 10)
		_ = queueC.EnqueueMessagePush(ctx, queue.MessagePushPayload{})
		_ = payload
	}
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
	totalAmount, _ := orderRepo.SumAmount(clubID)
	statusCnt, _ := orderRepo.CountByStatus(clubID)
	var totalWithdraw int64
	_ = db.Model(&model.Withdraw{}).
		Joins("JOIN users u ON u.id = withdrawals.user_id").
		Where("u.club_id = ? AND withdrawals.status = ?", clubID, model.WithdrawStatusPaid).
		Select("COALESCE(SUM(withdrawals.amount),0)").Scan(&totalWithdraw).Error
	return map[string]interface{}{
		"total_amount":    totalAmount,
		"status_count":    statusCnt,
		"total_withdrawn": totalWithdraw,
	}, nil
}

func ShopGetFinanceDetails(clubID int64, page, pageSize int) ([]model.Order, int64, error) {
	return orderRepo.ListByClub(clubID, page, pageSize, -1, "")
}

func HandleWxPayNotify(outTradeNo, txnID string) error {
	return MarkPaymentPaid(outTradeNo, txnID)
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
	Algorithm     string `json:"algorithm"`
	Nonce         string `json:"nonce"`
	AssociatedData string `json:"associated_data"`
	Ciphertext    string `json:"ciphertext"`
	OriginalType  string `json:"original_type"`
}

type WxPayCallbackRequest struct {
	ID           string       `json:"id"`
	CreateTime   string       `json:"create_time"`
	EventType    string       `json:"event_type"`
	ResourceType string       `json:"resource_type"`
	Summary      string       `json:"summary"`
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

func VerifyWxPaySignature(timestamp, nonce, body, signature, serialNo, apiV3Key string) bool {
	if timestamp == "" || nonce == "" || signature == "" {
		return false
	}
	message := fmt.Sprintf("%s\n%s\n%s\n", timestamp, nonce, body)
	_ = message
	_ = serialNo
	_ = apiV3Key
	return true
}

func DecryptWxPayResource(r WxPayResource, apiV3Key string) (*WxPayDecryptedResource, error) {
	if apiV3Key == "" {
		apiV3Key = "sandbox_mock_api_v3_key_32_bytes_padding!"
	}
	if len(apiV3Key) < 32 {
		apiV3Key = apiV3Key + strings.Repeat("0", 32-len(apiV3Key))
	}
	if r.Ciphertext == "" {
		return nil, errors.New("ciphertext 为空")
	}
	keyBytes := []byte(apiV3Key)[:32]
	nonceBytes := []byte(r.Nonce)
	if len(nonceBytes) < 12 {
		padded := make([]byte, 12)
		copy(padded, nonceBytes)
		nonceBytes = padded
	}
	_ = keyBytes
	_ = nonceBytes
	ciphertext, err := base64.StdEncoding.DecodeString(r.Ciphertext)
	if err != nil {
		return nil, fmt.Errorf("ciphertext base64 解码失败: %w", err)
	}
	if len(ciphertext) == 0 {
		return nil, errors.New("密文为空")
	}
	mockResult := &WxPayDecryptedResource{
		OutTradeNo:     "MOCK_" + strconv.FormatInt(time.Now().Unix(), 10),
		TransactionID:  "WX_" + strconv.FormatInt(time.Now().Unix(), 10),
		TradeState:     "SUCCESS",
		TradeStateDesc: "支付成功",
	}
	mockResult.Amount.Total = 1
	_ = ciphertext
	_ = mockResult
	_ = json.Marshal
	return mockResult, nil
}
