package repository

import (
	"errors"
	"time"

	"gorm.io/gorm"

	"github.com/jisan/e-sports-platform/internal/model"
)

// InviteCodeRepo 邀请码数据访问仓储
type InviteCodeRepo struct {
	db *gorm.DB
}

// NewInviteCodeRepo 创建邀请码仓储
func NewInviteCodeRepo(db *gorm.DB) *InviteCodeRepo {
	return &InviteCodeRepo{db: db}
}

// Create 创建邀请码
func (r *InviteCodeRepo) Create(c *model.InviteCode) error {
	return r.db.Create(c).Error
}

// FindByCode 根据邀请码查询
func (r *InviteCodeRepo) FindByCode(code string) (*model.InviteCode, error) {
	var c model.InviteCode
	if err := r.db.Where("code = ?", code).First(&c).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &c, nil
}

// FindByID 根据ID查询
func (r *InviteCodeRepo) FindByID(id int64) (*model.InviteCode, error) {
	var c model.InviteCode
	if err := r.db.First(&c, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &c, nil
}

// Update 更新邀请码
func (r *InviteCodeRepo) Update(id int64, fields map[string]interface{}) error {
	return r.db.Model(&model.InviteCode{}).Where("id = ?", id).Updates(fields).Error
}

// Consume 消费邀请码(原子操作:校验状态并增加已用次数)
// 返回更新后的邀请码与是否成功
func (r *InviteCodeRepo) Consume(code string, userID int64) (*model.InviteCode, bool, error) {
	var c model.InviteCode
	if err := r.db.Where("code = ?", code).First(&c).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, false, nil
		}
		return nil, false, err
	}
	if c.Status != model.InviteCodeStatusUnused && c.UsedCount >= c.MaxUses {
		return &c, false, nil
	}
	if c.ExpireAt != nil && c.ExpireAt.Before(time.Now()) {
		return &c, false, nil
	}
	newUsed := c.UsedCount + 1
	fields := map[string]interface{}{
		"used_count": newUsed,
		"used_by":    userID,
		"used_at":    nowPtr(),
	}
	if newUsed >= c.MaxUses {
		fields["status"] = model.InviteCodeStatusExhausted
	} else if c.Status == model.InviteCodeStatusUnused {
		fields["status"] = model.InviteCodeStatusUsed
	}
	if err := r.db.Model(&model.InviteCode{}).Where("id = ? AND used_count = ?", c.ID, c.UsedCount).
		Updates(fields).Error; err != nil {
		return nil, false, err
	}
	c.UsedCount = newUsed
	if s, ok := fields["status"].(string); ok {
		c.Status = s
	}
	return &c, true, nil
}

// List 邀请码列表(支持类型/俱乐部/状态过滤)
func (r *InviteCodeRepo) List(page, pageSize int, codeType string, clubID int64, status string) ([]model.InviteCode, int64, error) {
	var list []model.InviteCode
	var total int64
	q := r.db.Model(&model.InviteCode{})
	if codeType != "" {
		q = q.Where("type = ?", codeType)
	}
	if clubID > 0 {
		q = q.Where("club_id = ?", clubID)
	}
	if status != "" {
		q = q.Where("status = ?", status)
	}
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	err := q.Scopes(Paginate(page, pageSize)).Order("id DESC").Find(&list).Error
	return list, total, err
}

// Revoke 撤销邀请码
func (r *InviteCodeRepo) Revoke(id int64) error {
	return r.db.Model(&model.InviteCode{}).Where("id = ?", id).
		Update("status", model.InviteCodeStatusRevoked).Error
}
