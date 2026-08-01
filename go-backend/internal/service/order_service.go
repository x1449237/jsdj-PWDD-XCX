package service

import (
	"errors"
	"fmt"
	"time"

	"gorm.io/gorm"

	"github.com/jisan/e-sports-platform/internal/model"
)

// CreateOrderInput 创建订单入参
type CreateOrderInput struct {
	Type            int8   `json:"type"`             // 订单类型
	ClubID          int64  `json:"club_id"`           // 俱乐部ID
	ServiceID       int64  `json:"service_id"`        // 服务项目ID
	Amount          int64  `json:"amount"`            // 金额(分)
	TeamCount       int    `json:"team_count"`        // 车队人数
	AppointmentTime *time.Time `json:"appointment_time"` // 预约时间
}

// CreateOrder 用户创建订单
// 含未成年人下单拦截、订单号生成、状态流转日志
func CreateOrder(userID int64, in *CreateOrderInput) (*model.Order, error) {
	if in.Amount <= 0 {
		return nil, errors.New("订单金额必须大于 0")
	}
	// 用户校验
	u, err := userRepo.FindByID(userID)
	if err != nil {
		return nil, err
	}
	if u == nil {
		return nil, errors.New("用户不存在")
	}
	if u.Status == 0 {
		return nil, errors.New("账号已被封禁")
	}
	// 未成年宵禁拦截(22:00-8:00 禁止下单)
	if u.IsMinor == 1 {
		hour := time.Now().Hour()
		if hour >= 22 || hour < 8 {
			_ = db.Create(&model.MinorCurfewLog{
				UserID: userID, Action: model.MinorActionOrder,
				BlockedAt: nowTimePtr(), CreatedAt: nowTimePtr(),
			}).Error
			return nil, errors.New("未成年人宵禁时段(22:00-8:00)禁止下单")
		}
	}
	clubID := in.ClubID
	if clubID == 0 {
		clubID = u.ClubID
	}

	o := &model.Order{
		OrderNo:        genOrderNo(),
		Type:           in.Type,
		UserID:         userID,
		ClubID:         clubID,
		ServiceID:      in.ServiceID,
		Amount:         in.Amount,
		Status:         model.OrderStatusPending,
		PayStatus:      0,
		TeamCount:      in.TeamCount,
		AppointmentTime: in.AppointmentTime,
		IsMinorOrder:   u.IsMinor,
		CreatedAt:      nowTimePtr(),
		UpdatedAt:      nowTimePtr(),
	}
	if o.Type == 0 {
		o.Type = model.OrderTypeInstant
	}
	if o.TeamCount <= 0 {
		o.TeamCount = 1
	}
	if err := orderRepo.Create(o); err != nil {
		return nil, err
	}
	// 写入状态流转日志
	_ = orderRepo.CreateStatusLog(&model.OrderStatusLog{
		OrderID:      o.ID,
		FromStatus:   -1,
		ToStatus:     model.OrderStatusPending,
		OperatorID:   userID,
		OperatorType: "user",
		Reason:       "用户下单",
		CreatedAt:    nowTimePtr(),
	})
	return o, nil
}

// genOrderNo 生成订单号: ES + yyyyMMddHHmmss + 6位随机
func genOrderNo() string {
	return fmt.Sprintf("ES%s%06d", time.Now().Format("20060102150405"), time.Now().UnixNano()%1000000)
}

// CreateAppointment 创建预约订单
func CreateAppointment(userID int64, in *CreateOrderInput) (*model.Order, error) {
	if in.AppointmentTime == nil {
		return nil, errors.New("预约时间不能为空")
	}
	if in.AppointmentTime.Before(time.Now()) {
		return nil, errors.New("预约时间不能早于当前时间")
	}
	in.Type = model.OrderTypeAppointment
	return CreateOrder(userID, in)
}

// GetOrderList 用户订单列表
func GetOrderList(userID int64, page, pageSize int, status int8) ([]model.Order, int64, error) {
	return orderRepo.ListUserOrders(userID, page, pageSize, status)
}

// GetOrderDetail 订单详情(校验归属:客户/打手/平台管理员)
func GetOrderDetail(orderID, userID int64, isAdmin bool) (*model.Order, error) {
	o, err := orderRepo.FindByID(orderID)
	if err != nil {
		return nil, err
	}
	if o == nil {
		return nil, errors.New("订单不存在")
	}
	if !isAdmin && o.UserID != userID && o.PlayerID != userID {
		return nil, errors.New("无权查看该订单")
	}
	return o, nil
}

// CancelOrder 用户取消订单(仅待接单/已接单可取消)
func CancelOrder(orderID, userID int64, reason string) error {
	o, err := orderRepo.FindByID(orderID)
	if err != nil {
		return err
	}
	if o == nil {
		return errors.New("订单不存在")
	}
	if o.UserID != userID {
		return errors.New("无权操作该订单")
	}
	if o.Status != model.OrderStatusPending && o.Status != model.OrderStatusAccepted {
		return errors.New("当前订单状态不可取消")
	}
	if err := orderRepo.Update(orderID, map[string]interface{}{
		"status":     model.OrderStatusTimeout,
		"updated_at": nowTimePtr(),
	}); err != nil {
		return err
	}
	_ = orderRepo.CreateStatusLog(&model.OrderStatusLog{
		OrderID: orderID, FromStatus: o.Status, ToStatus: model.OrderStatusTimeout,
		OperatorID: userID, OperatorType: "user", Reason: reason,
		CreatedAt: nowTimePtr(),
	})
	return nil
}

// SubmitAppeal 提交订单申诉
func SubmitAppeal(orderID, userID int64, appealType, description string, evidenceURLs []string) (*model.Appeal, error) {
	o, err := orderRepo.FindByID(orderID)
	if err != nil {
		return nil, err
	}
	if o == nil {
		return nil, errors.New("订单不存在")
	}
	if o.UserID != userID && o.PlayerID != userID {
		return nil, errors.New("无权对该订单申诉")
	}
	var evJSON []byte
	if len(evidenceURLs) > 0 {
		evJSON = mustMarshal(evidenceURLs)
	}
	a := &model.Appeal{
		OrderID:      orderID,
		UserID:       userID,
		Type:         appealType,
		Description:  description,
		Status:       model.AppealStatusPending,
		EvidenceURLs: evJSON,
		CreatedAt:    nowTimePtr(),
		UpdatedAt:    nowTimePtr(),
	}
	if err := db.Create(a).Error; err != nil {
		return nil, err
	}
	return a, nil
}

// SubmitEvaluation 提交订单评价(仅已完成可评价)
func SubmitEvaluation(orderID, userID int64, score int, content string) (*model.Evaluation, error) {
	if !validScore(score) {
		return nil, errors.New("评分必须在 1-5 之间")
	}
	o, err := orderRepo.FindByID(orderID)
	if err != nil {
		return nil, err
	}
	if o == nil {
		return nil, errors.New("订单不存在")
	}
	if o.UserID != userID {
		return nil, errors.New("无权评价该订单")
	}
	if o.Status != model.OrderStatusCompleted && o.Status != model.OrderStatusSettled && o.Status != model.OrderStatusToSettle {
		return nil, errors.New("订单未完成，不可评价")
	}
	// 防重:同一订单只能评价一次
	var cnt int64
	_ = db.Model(&model.Evaluation{}).Where("order_id = ?", orderID).Count(&cnt).Error
	if cnt > 0 {
		return nil, errors.New("该订单已评价")
	}
	e := &model.Evaluation{
		OrderID:   orderID,
		UserID:    userID,
		PlayerID:  o.PlayerID,
		Score:     score,
		Content:   content,
		Status:    model.EvaluationStatusDisplayed,
		DisplayAt: nowTimePtr(),
		CreatedAt: nowTimePtr(),
		UpdatedAt: nowTimePtr(),
	}
	if err := db.Create(e).Error; err != nil {
		return nil, err
	}
	return e, nil
}

// validScore 校验评分
func validScore(s int) bool {
	return s >= 1 && s <= 5
}

// SendReward 用户给打手打赏
func SendReward(orderID, userID int64, amount int64, giftType string) (*model.Reward, error) {
	if amount <= 0 {
		return nil, errors.New("打赏金额必须大于 0")
	}
	o, err := orderRepo.FindByID(orderID)
	if err != nil {
		return nil, err
	}
	if o == nil {
		return nil, errors.New("订单不存在")
	}
	if o.UserID != userID {
		return nil, errors.New("无权对该订单打赏")
	}
	if o.PlayerID == 0 {
		return nil, errors.New("订单暂未匹配打手")
	}
	// 扣减用户余额
	u, err := userRepo.FindByID(userID)
	if err != nil {
		return nil, err
	}
	if u == nil {
		return nil, errors.New("用户不存在")
	}
	if u.IsMinor == 1 {
		_ = db.Create(&model.MinorCurfewLog{
			UserID: userID, Action: model.MinorActionReward,
			BlockedAt: nowTimePtr(), CreatedAt: nowTimePtr(),
		}).Error
		return nil, errors.New("未成年人禁止打赏")
	}
	if u.Balance < amount {
		return nil, errors.New("余额不足")
	}
	err = db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Model(&model.User{}).Where("id = ? AND balance >= ?", userID, amount).
			UpdateColumn("balance", gorm.Expr("balance - ?", amount)).Error; err != nil {
			return err
		}
		r := &model.Reward{
			OrderID:  orderID,
			UserID:   userID,
			PlayerID: o.PlayerID,
			Amount:   amount,
			GiftType: giftType,
			CreatedAt: nowTimePtr(),
		}
		if err := tx.Create(r).Error; err != nil {
			return err
		}
		// 打手余额入账(待结算)
		if err := tx.Model(&model.User{}).Where("id = ?", o.PlayerID).
			UpdateColumn("balance", gorm.Expr("balance + ?", amount)).Error; err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	r := &model.Reward{OrderID: orderID, UserID: userID, PlayerID: o.PlayerID, Amount: amount, GiftType: giftType}
	return r, nil
}

// UploadEvidence 上传订单履约凭证
func UploadEvidence(orderID, userID int64, fileType, fileURL string) error {
	o, err := orderRepo.FindByID(orderID)
	if err != nil {
		return err
	}
	if o == nil {
		return errors.New("订单不存在")
	}
	if o.UserID != userID && o.PlayerID != userID {
		return errors.New("无权上传该订单凭证")
	}
	return orderRepo.CreateEvidence(&model.OrderEvidence{
		OrderID: orderID,
		UserID:  userID,
		Type:    fileType,
		FileURL: fileURL,
		CreatedAt: nowTimePtr(),
	})
}

// mustMarshal JSON 序列化，失败返回空字节数组
func mustMarshal(v interface{}) []byte {
	if v == nil {
		return nil
	}
	b, _ := jsonMarshal(v)
	return b
}

// AdminGetOrders 平台订单列表
func AdminGetOrders(page, pageSize int, status int8, keyword string) ([]model.Order, int64, error) {
	return orderRepo.ListAll(page, pageSize, status, keyword)
}

// AdminForceUpdateOrderStatus 平台强制更新订单状态
func AdminForceUpdateOrderStatus(orderID, adminID int64, status int8, reason string) error {
	o, err := orderRepo.FindByID(orderID)
	if err != nil {
		return err
	}
	if o == nil {
		return errors.New("订单不存在")
	}
	old := o.Status
	if err := orderRepo.Update(orderID, map[string]interface{}{
		"status":     status,
		"updated_at": nowTimePtr(),
	}); err != nil {
		return err
	}
	return orderRepo.CreateStatusLog(&model.OrderStatusLog{
		OrderID: orderID, FromStatus: old, ToStatus: status,
		OperatorID: adminID, OperatorType: "admin", Reason: reason,
		CreatedAt: nowTimePtr(),
	})
}

// AdminGetFailedOrders 大额验证失败订单列表
func AdminGetFailedOrders(page, pageSize int) ([]model.Order, int64, error) {
	return orderRepo.ListAll(page, pageSize, model.OrderStatusVerifyFail, "")
}

// AdminBatchOrderOperation 批量订单操作
func AdminBatchOrderOperation(adminID int64, orderIDs []int64, action string) (int, error) {
	if len(orderIDs) == 0 {
		return 0, errors.New("未选择订单")
	}
	success := 0
	for _, oid := range orderIDs {
		var fields map[string]interface{}
		switch action {
		case "complete":
			fields = map[string]interface{}{"status": model.OrderStatusCompleted, "ended_at": nowTimePtr(), "updated_at": nowTimePtr()}
		case "cancel":
			fields = map[string]interface{}{"status": model.OrderStatusTimeout, "updated_at": nowTimePtr()}
		default:
			continue
		}
		if err := orderRepo.Update(oid, fields); err == nil {
			success++
		}
	}
	return success, nil
}

// ShopGetOrders 俱乐部订单列表
func ShopGetOrders(clubID int64, page, pageSize int, status int8, keyword string) ([]model.Order, int64, error) {
	return orderRepo.ListByClub(clubID, page, pageSize, status, keyword)
}

// ShopGetOrderDetail 俱乐部订单详情(校验俱乐部归属)
func ShopGetOrderDetail(orderID, clubID int64) (*model.Order, error) {
	o, err := orderRepo.FindByID(orderID)
	if err != nil {
		return nil, err
	}
	if o == nil {
		return nil, errors.New("订单不存在")
	}
	if clubID > 0 && o.ClubID != clubID {
		return nil, errors.New("无权查看该订单")
	}
	return o, nil
}

// ShopGetFailedOrders 俱乐部大额验证失败订单
func ShopGetFailedOrders(clubID int64, page, pageSize int) ([]model.Order, int64, error) {
	return orderRepo.ListByClub(clubID, page, pageSize, model.OrderStatusVerifyFail, "")
}

// ShopUpdateOrderStatus 内置管理端更新订单状态(创始人专属由 handler 校验角色)
func ShopUpdateOrderStatus(orderID, adminID int64, status int8, reason string) error {
	return AdminForceUpdateOrderStatus(orderID, adminID, status, reason)
}

// ShopGetAfterSaleOrders 售后订单列表
func ShopGetAfterSaleOrders(clubID int64, page, pageSize int) ([]model.AfterSaleSession, int64, error) {
	var list []model.AfterSaleSession
	var total int64
	q := db.Model(&model.AfterSaleSession{})
	if clubID > 0 {
		q = q.Where("club_id = ?", clubID)
	}
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	err := q.Scopes(Paginate2(page, pageSize)).Order("id DESC").Find(&list).Error
	return list, total, err
}

// Paginate2 service 层分页辅助(复用 repository.Paginate 但避免循环引用同名)
func Paginate2(page, pageSize int) func(*gorm.DB) *gorm.DB {
	return paginateScope(page, pageSize)
}

// paginateScope 实际分页 scope
func paginateScope(page, pageSize int) func(*gorm.DB) *gorm.DB {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 10
	}
	if pageSize > 100 {
		pageSize = 100
	}
	return func(db *gorm.DB) *gorm.DB {
		return db.Offset((page - 1) * pageSize).Limit(pageSize)
	}
}
