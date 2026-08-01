package repository

import (
	"errors"

	"gorm.io/gorm"

	"github.com/jisan/e-sports-platform/internal/model"
)

// AdminRepo 平台管理员数据访问仓储
type AdminRepo struct {
	db *gorm.DB
}

// NewAdminRepo 创建管理员仓储
func NewAdminRepo(db *gorm.DB) *AdminRepo {
	return &AdminRepo{db: db}
}

// FindByUsername 根据账号查询管理员
func (r *AdminRepo) FindByUsername(username string) (*model.Admin, error) {
	var a model.Admin
	if err := r.db.Where("username = ?", username).First(&a).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &a, nil
}

// FindByID 根据ID查询管理员
func (r *AdminRepo) FindByID(id int64) (*model.Admin, error) {
	var a model.Admin
	if err := r.db.First(&a, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &a, nil
}

// Create 创建管理员
func (r *AdminRepo) Create(a *model.Admin) error {
	return r.db.Create(a).Error
}

// Update 更新管理员字段
func (r *AdminRepo) Update(id int64, fields map[string]interface{}) error {
	return r.db.Model(&model.Admin{}).Where("id = ?", id).Updates(fields).Error
}

// Delete 删除管理员
func (r *AdminRepo) Delete(id int64) error {
	return r.db.Delete(&model.Admin{}, id).Error
}

// List 分页查询管理员
func (r *AdminRepo) List(page, pageSize int, keyword string) ([]model.Admin, int64, error) {
	var admins []model.Admin
	var total int64
	q := r.db.Model(&model.Admin{})
	if keyword != "" {
		like := "%" + keyword + "%"
		q = q.Where("username LIKE ? OR nickname LIKE ? OR email LIKE ?", like, like, like)
	}
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	err := q.Scopes(Paginate(page, pageSize)).Order("id DESC").Find(&admins).Error
	return admins, total, err
}

// UpdateLastLogin 更新最后登录信息
func (r *AdminRepo) UpdateLastLogin(id int64, ip string) error {
	return r.db.Model(&model.Admin{}).Where("id = ?", id).
		Updates(map[string]interface{}{"last_login_at": nowPtr(), "last_login_ip": ip}).Error
}

// CountSuperAdmin 统计超级管理员数量
func (r *AdminRepo) CountSuperAdmin() (int64, error) {
	var count int64
	err := r.db.Model(&model.Admin{}).Where("role & ? > 0", model.AdminRoleSuper).Count(&count).Error
	return count, err
}

// CreatePasswordHistory 写入密码历史
func (r *AdminRepo) CreatePasswordHistory(h *model.AdminPasswordHistory) error {
	return r.db.Create(h).Error
}

// ListPasswordHistory 查询最近 N 条密码历史
func (r *AdminRepo) ListPasswordHistory(adminID int64, limit int) ([]model.AdminPasswordHistory, error) {
	var list []model.AdminPasswordHistory
	if limit <= 0 {
		limit = 5
	}
	err := r.db.Where("admin_id = ?", adminID).Order("id DESC").Limit(limit).Find(&list).Error
	return list, err
}

// CreateWebauthn 创建 WebAuthn 通行密钥
func (r *AdminRepo) CreateWebauthn(w *model.AdminWebauthn) error {
	return r.db.Create(w).Error
}

// FindWebauthnByCredential 根据凭证ID查询通行密钥
func (r *AdminRepo) FindWebauthnByCredential(credID string) (*model.AdminWebauthn, error) {
	var w model.AdminWebauthn
	if err := r.db.Where("credential_id = ?", credID).First(&w).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &w, nil
}

// ListWebauthnByAdmin 查询管理员绑定的通行密钥
func (r *AdminRepo) ListWebauthnByAdmin(adminID int64) ([]model.AdminWebauthn, error) {
	var list []model.AdminWebauthn
	err := r.db.Where("admin_id = ?", adminID).Find(&list).Error
	return list, err
}
