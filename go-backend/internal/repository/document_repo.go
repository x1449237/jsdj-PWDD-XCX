package repository

import (
	"errors"

	"gorm.io/gorm"

	"github.com/jisan/e-sports-platform/internal/model"
)

// DocumentRepo 协议文档数据访问仓储
type DocumentRepo struct {
	db *gorm.DB
}

// NewDocumentRepo 创建文档仓储
func NewDocumentRepo(db *gorm.DB) *DocumentRepo {
	return &DocumentRepo{db: db}
}

// ============ Document 文档 ============

// CreateDocument 创建文档
func (r *DocumentRepo) CreateDocument(d *model.PlatformDocument) error {
	return r.db.Create(d).Error
}

// FindDocumentByID 根据ID查询文档（含未删除）
func (r *DocumentRepo) FindDocumentByID(id int64) (*model.PlatformDocument, error) {
	var d model.PlatformDocument
	if err := r.db.First(&d, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &d, nil
}

// ListDocuments 查询文档列表（默认过滤已删除）
func (r *DocumentRepo) ListDocuments(includeDeleted bool) ([]model.PlatformDocument, error) {
	var list []model.PlatformDocument
	q := r.db.Model(&model.PlatformDocument{})
	if !includeDeleted {
		q = q.Where("is_deleted = 0")
	}
	err := q.Order("id DESC").Find(&list).Error
	return list, err
}

// UpdateDocument 更新文档
func (r *DocumentRepo) UpdateDocument(id int64, fields map[string]interface{}) error {
	return r.db.Model(&model.PlatformDocument{}).Where("id = ?", id).Updates(fields).Error
}

// SoftDeleteDocument 软删除文档
func (r *DocumentRepo) SoftDeleteDocument(id int64) error {
	return r.db.Model(&model.PlatformDocument{}).Where("id = ?", id).
		Updates(map[string]interface{}{"is_deleted": 1, "updated_at": gorm.Expr("COALESCE(updated_at, NOW())")}).Error
}

// ============ Version 版本 ============

// CreateVersion 创建版本记录
func (r *DocumentRepo) CreateVersion(v *model.DocumentVersion) error {
	return r.db.Create(v).Error
}

// ListVersionsByDocument 按文档列出所有版本
func (r *DocumentRepo) ListVersionsByDocument(documentID int64) ([]model.DocumentVersion, error) {
	var list []model.DocumentVersion
	err := r.db.Where("document_id = ?", documentID).Order("id DESC").Find(&list).Error
	return list, err
}

// ============ SignLog 签署流水 ============

// CreateSignLog 写签署流水
func (r *DocumentRepo) CreateSignLog(l *model.AgreementSignLog) error {
	return r.db.Create(l).Error
}

// CountSignedByUserAndDoc 查询用户是否已签某文档
func (r *DocumentRepo) CountSignedByUserAndDoc(uid, documentID int64) (int64, error) {
	var cnt int64
	err := r.db.Model(&model.AgreementSignLog{}).
		Where("uid = ? AND document_id = ?", uid, documentID).Count(&cnt).Error
	return cnt, err
}
