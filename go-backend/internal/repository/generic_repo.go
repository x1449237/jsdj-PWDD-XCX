package repository

import (
	"time"

	"gorm.io/gorm"
)

// GenericRepo 通用 CRUD 仓储，封装常见的数据库操作
// 通过嵌入 GenericRepo，领域仓储可复用基础 CRUD 能力
type GenericRepo struct {
	DB *gorm.DB
}

// NewGenericRepo 创建通用仓储
func NewGenericRepo(db *gorm.DB) *GenericRepo {
	return &GenericRepo{DB: db}
}

// Create 创建记录(模型指针)
func (r *GenericRepo) Create(model interface{}) error {
	return r.DB.Create(model).Error
}

// FindByID 根据主键查询
func (r *GenericRepo) FindByID(model interface{}, id int64) error {
	return r.DB.First(model, id).Error
}

// Updates 更新指定字段(map 形式)
func (r *GenericRepo) Updates(model interface{}, id int64, fields map[string]interface{}) error {
	return r.DB.Model(model).Where("id = ?", id).Updates(fields).Error
}

// Delete 物理删除
func (r *GenericRepo) Delete(model interface{}, id int64) error {
	return r.DB.Where("id = ?", id).Delete(model).Error
}

// SoftDelete 软删除(更新 status 字段为 0)
func (r *GenericRepo) SoftDelete(model interface{}, id int64) error {
	return r.DB.Model(model).Where("id = ?", id).Update("status", 0).Error
}

// Count 统计符合条件的记录数
func (r *GenericRepo) Count(model interface{}, query string, args ...interface{}) (int64, error) {
	var count int64
	err := r.DB.Model(model).Where(query, args...).Count(&count).Error
	return count, err
}

// Paginate 分页查询辅助
// page 从 1 开始，pageSize 每页条数；返回 offset
func Paginate(page, pageSize int) func(db *gorm.DB) *gorm.DB {
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
		offset := (page - 1) * pageSize
		return db.Offset(offset).Limit(pageSize)
	}
}

// nowPtr 返回当前时间的指针(供写入 created_at 等字段)
func nowPtr() *time.Time {
	now := time.Now()
	return &now
}

// txFrom 从可选的 *gorm.DB 获取事务句柄，为空则使用默认 db
func txFrom(db, tx *gorm.DB) *gorm.DB {
	if tx != nil {
		return tx
	}
	return db
}
