package repository

import (
	"errors"

	"gorm.io/gorm"

	"github.com/jisan/e-sports-platform/internal/model"
)

// PaymentRepo 支付/提现数据访问仓储
type PaymentRepo struct {
	db *gorm.DB
}

// NewPaymentRepo 创建支付仓储
func NewPaymentRepo(db *gorm.DB) *PaymentRepo {
	return &PaymentRepo{db: db}
}

// CreatePayment 创建支付记录
func (r *PaymentRepo) CreatePayment(p *model.Payment) error {
	return r.db.Create(p).Error
}

// FindPaymentByOutTradeNo 根据商户订单号查询支付记录
func (r *PaymentRepo) FindPaymentByOutTradeNo(no string) (*model.Payment, error) {
	var p model.Payment
	if err := r.db.Where("out_trade_no = ?", no).First(&p).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &p, nil
}

// FindPaidPaymentByOrderID 根据订单ID查询已支付的支付记录(用于退款)
// 退款应按 order_id + status=paid 查询,而非 out_trade_no(订单号与支付单号不一致)
func (r *PaymentRepo) FindPaidPaymentByOrderID(orderID int64) (*model.Payment, error) {
	var p model.Payment
	if err := r.db.Where("order_id = ? AND status = ?", orderID, model.PaymentStatusPaid).
		Order("id DESC").First(&p).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &p, nil
}

// FindRefundablePaymentByOrderID 根据订单ID查询可退款的支付记录
// 安全修复:原 FindPaidPaymentByOrderID 仅查 status=paid,
// 部分退款后状态变为 partial_refund,导致无法继续退款至全额
// 现改为查询 status IN (paid, partial_refund)
func (r *PaymentRepo) FindRefundablePaymentByOrderID(orderID int64) (*model.Payment, error) {
	var p model.Payment
	if err := r.db.Where("order_id = ? AND status IN ?", orderID,
		[]string{model.PaymentStatusPaid, model.PaymentStatusPartialRef}).
		Order("id DESC").First(&p).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &p, nil
}

// FindPendingPaymentByOrderID 根据订单ID查询待支付的支付记录(防重复创建)
func (r *PaymentRepo) FindPendingPaymentByOrderID(orderID int64) (*model.Payment, error) {
	var p model.Payment
	if err := r.db.Where("order_id = ? AND status = ?", orderID, model.PaymentStatusPending).
		Order("id DESC").First(&p).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &p, nil
}

// FindPaymentByTxnID 根据第三方交易号查询支付记录
func (r *PaymentRepo) FindPaymentByTxnID(txnID string) (*model.Payment, error) {
	var p model.Payment
	if err := r.db.Where("transaction_id = ?", txnID).First(&p).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &p, nil
}

// UpdatePayment 更新支付记录
func (r *PaymentRepo) UpdatePayment(id int64, fields map[string]interface{}) error {
	return r.db.Model(&model.Payment{}).Where("id = ?", id).Updates(fields).Error
}

// CreateWithdraw 创建提现记录
func (r *PaymentRepo) CreateWithdraw(w *model.Withdraw) error {
	return r.db.Create(w).Error
}

// FindWithdraw 根据ID查询提现记录
func (r *PaymentRepo) FindWithdraw(id int64) (*model.Withdraw, error) {
	var w model.Withdraw
	if err := r.db.First(&w, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &w, nil
}

// UpdateWithdraw 更新提现记录
func (r *PaymentRepo) UpdateWithdraw(id int64, fields map[string]interface{}) error {
	return r.db.Model(&model.Withdraw{}).Where("id = ?", id).Updates(fields).Error
}

// ListUserWithdraws 用户提现记录列表
func (r *PaymentRepo) ListUserWithdraws(userID int64, page, pageSize int) ([]model.Withdraw, int64, error) {
	var list []model.Withdraw
	var total int64
	q := r.db.Model(&model.Withdraw{}).Where("user_id = ?", userID)
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	err := q.Scopes(Paginate(page, pageSize)).Order("id DESC").Find(&list).Error
	return list, total, err
}

// ListAllWithdraws 平台提现记录列表(按状态过滤)
func (r *PaymentRepo) ListAllWithdraws(page, pageSize int, status string) ([]model.Withdraw, int64, error) {
	var list []model.Withdraw
	var total int64
	q := r.db.Model(&model.Withdraw{})
	if status != "" {
		q = q.Where("status = ?", status)
	}
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	err := q.Scopes(Paginate(page, pageSize)).Order("id DESC").Find(&list).Error
	return list, total, err
}

// SumFrozenEarnings 统计用户待结算(冻结)收益
// 简化口径:已接单/进行中/待验收状态订单金额 - 已提现金额
func (r *PaymentRepo) SumFrozenEarnings(userID int64) (int64, error) {
	var total int64
	err := r.db.Raw(`
		SELECT COALESCE(SUM(amount), 0)
		FROM orders
		WHERE player_id = ? AND status IN (1,2,3)
	`, userID).Scan(&total).Error
	return total, err
}

// SumSettledEarnings 统计用户已结算收益
func (r *PaymentRepo) SumSettledEarnings(userID int64) (int64, error) {
	var total int64
	err := r.db.Raw(`
		SELECT COALESCE(SUM(amount), 0)
		FROM orders
		WHERE player_id = ? AND status IN (4,5,6)
	`, userID).Scan(&total).Error
	return total, err
}

// CreateProfitShareRule 创建分账规则
func (r *PaymentRepo) CreateProfitShareRule(rule *model.ProfitShareRule) error {
	return r.db.Create(rule).Error
}

// ListProfitShareRules 查询启用的分账规则
func (r *PaymentRepo) ListProfitShareRules() ([]model.ProfitShareRule, error) {
	var list []model.ProfitShareRule
	err := r.db.Where("status = 1").Find(&list).Error
	return list, err
}

// CreateProfitShareRecord 创建分账记录
func (r *PaymentRepo) CreateProfitShareRecord(rec *model.ProfitShareRecord) error {
	return r.db.Create(rec).Error
}

// ListProfitShareByOrder 查询订单的分账记录
func (r *PaymentRepo) ListProfitShareByOrder(orderID int64) ([]model.ProfitShareRecord, error) {
	var list []model.ProfitShareRecord
	err := r.db.Where("order_id = ?", orderID).Find(&list).Error
	return list, err
}

// UpdateProfitShareStatus 批量更新分账状态
func (r *PaymentRepo) UpdateProfitShareStatus(orderID int64, status int8) error {
	return r.db.Model(&model.ProfitShareRecord{}).
		Where("order_id = ?", orderID).Update("status", status).Error
}
