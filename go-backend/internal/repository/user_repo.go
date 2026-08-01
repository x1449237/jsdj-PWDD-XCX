package repository

import (
	"errors"

	"gorm.io/gorm"

	"github.com/jisan/e-sports-platform/internal/model"
)

// UserRepo 用户数据访问仓储
type UserRepo struct {
	db *gorm.DB
}

// NewUserRepo 创建用户仓储
func NewUserRepo(db *gorm.DB) *UserRepo {
	return &UserRepo{db: db}
}

// FindByOpenID 根据 openid 查询用户
func (r *UserRepo) FindByOpenID(openid string) (*model.User, error) {
	var u model.User
	if err := r.db.Where("openid = ?", openid).First(&u).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &u, nil
}

// FindByID 根据ID查询用户
func (r *UserRepo) FindByID(id int64) (*model.User, error) {
	var u model.User
	if err := r.db.First(&u, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &u, nil
}

// FindByPhone 根据手机号查询用户
func (r *UserRepo) FindByPhone(phone string) (*model.User, error) {
	var u model.User
	if err := r.db.Where("phone = ?", phone).First(&u).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &u, nil
}

// Create 创建用户
func (r *UserRepo) Create(u *model.User) error {
	return r.db.Create(u).Error
}

// Update 更新用户指定字段
func (r *UserRepo) Update(id int64, fields map[string]interface{}) error {
	return r.db.Model(&model.User{}).Where("id = ?", id).Updates(fields).Error
}

// UpdateBalance 增减用户余额(正数增加，负数扣减)
func (r *UserRepo) UpdateBalance(id int64, delta int64) error {
	return r.db.Model(&model.User{}).Where("id = ?", id).
		UpdateColumn("balance", gorm.Expr("balance + ?", delta)).Error
}

// UpdatePoints 增减用户积分
func (r *UserRepo) UpdatePoints(id int64, delta int) error {
	return r.db.Model(&model.User{}).Where("id = ?", id).
		UpdateColumn("points", gorm.Expr("points + ?", delta)).Error
}

// ListByIDs 根据ID列表批量查询
func (r *UserRepo) ListByIDs(ids []int64) ([]model.User, error) {
	var users []model.User
	if len(ids) == 0 {
		return users, nil
	}
	err := r.db.Where("id IN ?", ids).Find(&users).Error
	return users, err
}

// List 分页查询用户(支持状态/角色/关键词过滤)
func (r *UserRepo) List(page, pageSize int, status, role int8, keyword string) ([]model.User, int64, error) {
	var users []model.User
	var total int64
	q := r.db.Model(&model.User{})
	if status >= 0 {
		q = q.Where("status = ?", status)
	}
	if role > 0 {
		q = q.Where("role & ? > 0", role)
	}
	if keyword != "" {
		like := "%" + keyword + "%"
		q = q.Where("nickname LIKE ? OR phone LIKE ? OR real_name LIKE ?", like, like, like)
	}
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	err := q.Scopes(Paginate(page, pageSize)).Order("id DESC").Find(&users).Error
	return users, total, err
}
