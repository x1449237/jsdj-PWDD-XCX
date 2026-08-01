package repository

import (
	"errors"

	"gorm.io/gorm"

	"github.com/jisan/e-sports-platform/internal/model"
)

// OrderRepo 订单数据访问仓储
type OrderRepo struct {
	db *gorm.DB
}

// NewOrderRepo 创建订单仓储
func NewOrderRepo(db *gorm.DB) *OrderRepo {
	return &OrderRepo{db: db}
}

// Create 创建订单
func (r *OrderRepo) Create(o *model.Order) error {
	return r.db.Create(o).Error
}

// FindByID 根据ID查询订单
func (r *OrderRepo) FindByID(id int64) (*model.Order, error) {
	var o model.Order
	if err := r.db.First(&o, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &o, nil
}

// FindByOrderNo 根据订单号查询订单
func (r *OrderRepo) FindByOrderNo(orderNo string) (*model.Order, error) {
	var o model.Order
	if err := r.db.Where("order_no = ?", orderNo).First(&o).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &o, nil
}

// Update 更新订单字段
func (r *OrderRepo) Update(id int64, fields map[string]interface{}) error {
	return r.db.Model(&model.Order{}).Where("id = ?", id).Updates(fields).Error
}

// GrabOrder 抢单：原子更新订单状态(仅待接单可被抢)
func (r *OrderRepo) GrabOrder(orderID, playerID int64) (bool, error) {
	// 通过条件更新保证原子性，避免并发抢单冲突
	res := r.db.Model(&model.Order{}).
		Where("id = ? AND status = ?", orderID, model.OrderStatusPending).
		Updates(map[string]interface{}{
			"player_id":   playerID,
			"status":      model.OrderStatusAccepted,
			"accepted_at": nowPtr(),
		})
	if res.Error != nil {
		return false, res.Error
	}
	return res.RowsAffected > 0, nil
}

// ListUserOrders 查询客户订单列表(分页)
func (r *OrderRepo) ListUserOrders(userID int64, page, pageSize int, status int8) ([]model.Order, int64, error) {
	var orders []model.Order
	var total int64
	q := r.db.Model(&model.Order{}).Where("user_id = ?", userID)
	if status >= 0 {
		q = q.Where("status = ?", status)
	}
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	err := q.Scopes(Paginate(page, pageSize)).Order("id DESC").Find(&orders).Error
	return orders, total, err
}

// ListPlayerOrders 查询打手接单列表(分页)
func (r *OrderRepo) ListPlayerOrders(playerID int64, page, pageSize int, status int8) ([]model.Order, int64, error) {
	var orders []model.Order
	var total int64
	q := r.db.Model(&model.Order{}).Where("player_id = ?", playerID)
	if status >= 0 {
		q = q.Where("status = ?", status)
	}
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	err := q.Scopes(Paginate(page, pageSize)).Order("id DESC").Find(&orders).Error
	return orders, total, err
}

// ListGrabOrders 查询可抢单列表(待接单)
func (r *OrderRepo) ListGrabOrders(clubID int64, page, pageSize int) ([]model.Order, int64, error) {
	var orders []model.Order
	var total int64
	q := r.db.Model(&model.Order{}).Where("status = ? AND (club_id = ? OR ? = 0)",
		model.OrderStatusPending, clubID, clubID)
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	err := q.Scopes(Paginate(page, pageSize)).Order("id DESC").Find(&orders).Error
	return orders, total, err
}

// ListByClub 俱乐部订单查询(支持多状态/关键词)
func (r *OrderRepo) ListByClub(clubID int64, page, pageSize int, status int8, keyword string) ([]model.Order, int64, error) {
	var orders []model.Order
	var total int64
	q := r.db.Model(&model.Order{}).Where("club_id = ?", clubID)
	if status >= 0 {
		q = q.Where("status = ?", status)
	}
	if keyword != "" {
		q = q.Where("order_no LIKE ?", "%"+keyword+"%")
	}
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	err := q.Scopes(Paginate(page, pageSize)).Order("id DESC").Find(&orders).Error
	return orders, total, err
}

// ListAll 平台维度订单查询
func (r *OrderRepo) ListAll(page, pageSize int, status int8, keyword string) ([]model.Order, int64, error) {
	var orders []model.Order
	var total int64
	q := r.db.Model(&model.Order{})
	if status >= 0 {
		q = q.Where("status = ?", status)
	}
	if keyword != "" {
		q = q.Where("order_no LIKE ?", "%"+keyword+"%")
	}
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	err := q.Scopes(Paginate(page, pageSize)).Order("id DESC").Find(&orders).Error
	return orders, total, err
}

// CountByStatus 统计俱乐部各状态订单数
func (r *OrderRepo) CountByStatus(clubID int64) (map[int8]int64, error) {
	type row struct {
		Status int8 `gorm:"column:status"`
		Cnt    int64 `gorm:"column:cnt"`
	}
	var rows []row
	q := r.db.Model(&model.Order{}).Select("status, COUNT(*) AS cnt")
	if clubID > 0 {
		q = q.Where("club_id = ?", clubID)
	}
	if err := q.Group("status").Scan(&rows).Error; err != nil {
		return nil, err
	}
	m := make(map[int8]int64, len(rows))
	for _, r := range rows {
		m[r.Status] = r.Cnt
	}
	return m, nil
}

// CreateStatusLog 写入订单状态流转日志
func (r *OrderRepo) CreateStatusLog(l *model.OrderStatusLog) error {
	return r.db.Create(l).Error
}

// CreateEvidence 写入履约凭证
func (r *OrderRepo) CreateEvidence(e *model.OrderEvidence) error {
	return r.db.Create(e).Error
}

// ListEvidence 查询订单凭证
func (r *OrderRepo) ListEvidence(orderID int64) ([]model.OrderEvidence, error) {
	var list []model.OrderEvidence
	err := r.db.Where("order_id = ?", orderID).Order("id DESC").Find(&list).Error
	return list, err
}

// SumAmount 统计俱乐部订单总金额(已完成)
func (r *OrderRepo) SumAmount(clubID int64) (int64, error) {
	var total int64
	q := r.db.Model(&model.Order{}).Where("status IN ?", []int8{model.OrderStatusCompleted, model.OrderStatusSettled})
	if clubID > 0 {
		q = q.Where("club_id = ?", clubID)
	}
	err := q.Select("COALESCE(SUM(amount),0)").Scan(&total).Error
	return total, err
}
