package service

import (
	"context"
	"errors"
	"time"

	"gorm.io/gorm"

	"github.com/jisan/e-sports-platform/internal/model"
	"github.com/jisan/e-sports-platform/pkg/queue"
)

// CreatePayment 创建支付记录(下单后发起支付)
func CreatePayment(orderID, amount int64, payMethod string) (*model.Payment, error) {
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
	return p, nil
}

// MarkPaymentPaid 标记支付成功(webhook 回调调用)
func MarkPaymentPaid(outTradeNo, txnID string) error {
	p, err := paymentRepo.FindPaymentByOutTradeNo(outTradeNo)
	if err != nil {
		return err
	}
	if p == nil {
		return errors.New("支付记录不存在")
	}
	if p.Status == model.PaymentStatusPaid {
		return nil // 幂等
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
		// 更新订单支付状态
		return tx.Model(&model.Order{}).Where("id = ?", p.OrderID).
			Updates(map[string]interface{}{
				"pay_status": 1,
				"paid_at":    now,
				"updated_at": now,
			}).Error
	})
}

// ProcessRefund 处理退款(全额/部分)
// refundAmount 为 0 表示全额退款
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
		// 没有支付记录时按订单金额构造
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
			orderStatus = o.Status // 部分退款保持原订单状态
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
	// 退还用户余额(若原支付走余额)
	_ = userRepo.UpdateBalance(o.UserID, refundAmount)
	_ = operatorID
	_ = isAdmin
	p.RefundAmount = newRefund
	p.Status = newStatus
	return p, nil
}

// ShopProcessRefund 内置管理端处理退款
func ShopProcessRefund(orderID, adminID int64, refundAmount int64) (*model.Payment, error) {
	return ProcessRefund(orderID, adminID, refundAmount, false)
}

// AdminProcessRefund 平台处理退款
func AdminProcessRefund(orderID, adminID int64, refundAmount int64) (*model.Payment, error) {
	return ProcessRefund(orderID, adminID, refundAmount, true)
}

// AdminGetWithdrawals 平台提现记录列表
func AdminGetWithdrawals(page, pageSize int, status string) ([]model.Withdraw, int64, error) {
	return paymentRepo.ListAllWithdraws(page, pageSize, status)
}

// AdminApproveWithdrawal 审核通过提现
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
	return paymentRepo.UpdateWithdraw(withdrawID, map[string]interface{}{
		"status":      model.WithdrawStatusApproved,
		"reviewer_id": adminID,
		"reviewed_at": nowTimePtr(),
		"updated_at":  nowTimePtr(),
	})
}

// AdminRejectWithdrawal 驳回提现
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

// AdminBatchWithdraw 批量提现处理
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

// ShopGetWithdrawals 俱乐部提现记录(按俱乐部过滤打手)
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

// ShopGetFinanceOverview 俱乐部财务概览
func ShopGetFinanceOverview(clubID int64) (map[string]interface{}, error) {
	totalAmount, _ := orderRepo.SumAmount(clubID)
	statusCnt, _ := orderRepo.CountByStatus(clubID)
	var totalWithdraw int64
	_ = db.Model(&model.Withdraw{}).
		Joins("JOIN users u ON u.id = withdrawals.user_id").
		Where("u.club_id = ? AND withdrawals.status = ?", clubID, model.WithdrawStatusPaid).
		Select("COALESCE(SUM(withdrawals.amount),0)").Scan(&totalWithdraw).Error
	return map[string]interface{}{
		"total_amount":     totalAmount,
		"status_count":     statusCnt,
		"total_withdrawn":  totalWithdraw,
	}, nil
}

// ShopGetFinanceDetails 俱乐部财务明细
func ShopGetFinanceDetails(clubID int64, page, pageSize int) ([]model.Order, int64, error) {
	return orderRepo.ListByClub(clubID, page, pageSize, -1, "")
}

// HandleWxPayNotify 处理微信支付回调
func HandleWxPayNotify(outTradeNo, txnID string) error {
	return MarkPaymentPaid(outTradeNo, txnID)
}

// enqueueWithdrawPaidTask 占位:投递提现打款任务(实际由 queue 处理)
func enqueueWithdrawPaidTask(withdrawID int64) {
	if queueC == nil {
		return
	}
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	_ = queueC.EnqueueWithdrawProcess(ctx, queue.WithdrawProcessPayload{WithdrawID: withdrawID})
}
