package service

import (
	"errors"

	"github.com/jisan/e-sports-platform/internal/model"
)

// GetDispatchOrders 派单员待派订单列表(待接单且无打手)
func GetDispatchOrders(page, pageSize int) ([]model.Order, int64, error) {
	return orderRepo.ListAll(page, pageSize, model.OrderStatusPending, "")
}

// DispatchOrder 派单员指定打手接单
func DispatchOrder(orderID, dispatcherID, playerID int64) error {
	o, err := orderRepo.FindByID(orderID)
	if err != nil {
		return err
	}
	if o == nil {
		return errors.New("订单不存在")
	}
	if o.Status != model.OrderStatusPending {
		return errors.New("订单已被接单，无法派单")
	}
	// 直接指派打手
	if err := orderRepo.Update(orderID, map[string]interface{}{
		"player_id":   playerID,
		"status":      model.OrderStatusAccepted,
		"accepted_at": nowTimePtr(),
		"updated_at":  nowTimePtr(),
	}); err != nil {
		return err
	}
	return orderRepo.CreateStatusLog(&model.OrderStatusLog{
		OrderID: orderID, FromStatus: model.OrderStatusPending, ToStatus: model.OrderStatusAccepted,
		OperatorID: dispatcherID, OperatorType: "user", Reason: "派单员指派",
		CreatedAt: nowTimePtr(),
	})
}

// GetDispatchHistory 派单历史(派单员维度的订单流转)
func GetDispatchHistory(dispatcherID int64, page, pageSize int) ([]model.OrderStatusLog, int64, error) {
	var list []model.OrderStatusLog
	var total int64
	q := db.Model(&model.OrderStatusLog{}).Where("operator_id = ? AND operator_type = ?", dispatcherID, "user")
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	err := q.Scopes(Paginate2(page, pageSize)).Order("id DESC").Find(&list).Error
	return list, total, err
}

// AuditDispatchers 派单员审核列表
func AuditDispatchers(page, pageSize int, keyword string) ([]model.User, int64, error) {
	return userRepo.List(page, pageSize, -1, model.RoleDispatcher, keyword)
}

// ApproveDispatcher 审核通过派单员
func ApproveDispatcher(dispatcherID int64) error {
	return userRepo.Update(dispatcherID, map[string]interface{}{
		"updated_at": nowTimePtr(),
	})
}
