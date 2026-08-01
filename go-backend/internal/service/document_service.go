package service

import (
	"errors"

	"github.com/jisan/e-sports-platform/internal/model"
)

// AdminGetDocuments 管理端列出所有文档（含软删除）
func AdminGetDocuments() ([]model.PlatformDocument, error) {
	var list []model.PlatformDocument
	err := db.Model(&model.PlatformDocument{}).
		Order("id DESC").Find(&list).Error
	return list, err
}

// AdminUploadDocument 管理端上传新文档
func AdminUploadDocument(adminID int64, name, docType, fileURL, version, role string) (*model.PlatformDocument, error) {
	if name == "" || docType == "" || fileURL == "" {
		return nil, errors.New("文档名称/类型/文件地址不能为空")
	}
	if version == "" {
		version = "1.0.0"
	}
	if role == "" {
		role = model.DocRolePlayer
	}
	now := nowTimePtr()
	doc := &model.PlatformDocument{
		Name:      name,
		DocType:   docType,
		FileUrl:   fileURL,
		Version:   version,
		Role:      role,
		IsDeleted: 0,
		CreatedBy: adminID,
		CreatedAt: now,
		UpdatedAt: now,
	}
	if err := db.Create(doc).Error; err != nil {
		return nil, err
	}
	// 同步写版本记录
	_ = db.Create(&model.DocumentVersion{
		DocumentID: doc.ID,
		FileUrl:    fileURL,
		Version:    version,
		CreatedBy:  adminID,
		CreatedAt:  now,
	}).Error
	return doc, nil
}

// AdminReplaceDocument 替换/更新文档（升级版本+写版本历史）
func AdminReplaceDocument(documentID, adminID int64, name, fileURL, newVersion, role string) (*model.PlatformDocument, error) {
	doc := &model.PlatformDocument{}
	if err := db.First(doc, documentID).Error; err != nil {
		return nil, errors.New("文档不存在")
	}
	updates := map[string]interface{}{"updated_at": nowTimePtr()}
	if name != "" {
		updates["name"] = name
	}
	if fileURL != "" {
		updates["file_url"] = fileURL
	}
	if newVersion != "" {
		updates["version"] = newVersion
	}
	if role != "" {
		updates["role"] = role
	}
	if err := db.Model(doc).Updates(updates).Error; err != nil {
		return nil, err
	}
	// 新版本入库（记录真实更新后的 version & file_url）
	reloaded := &model.PlatformDocument{}
	_ = db.First(reloaded, documentID).Error
	_ = db.Create(&model.DocumentVersion{
		DocumentID: reloaded.ID,
		FileUrl:    reloaded.FileUrl,
		Version:    reloaded.Version,
		CreatedBy:  adminID,
		CreatedAt:  nowTimePtr(),
	}).Error
	return reloaded, nil
}

// AdminDeleteDocument 管理端软删除文档
func AdminDeleteDocument(documentID, adminID int64) error {
	_ = adminID
	return db.Model(&model.PlatformDocument{}).Where("id = ?", documentID).
		Updates(map[string]interface{}{
			"is_deleted": 1,
			"updated_at": nowTimePtr(),
		}).Error
}

// ====== 兼容 handler 简化签名（老接口兼容层） ======

// AdminUploadDocumentSimple 兼容签名（无 adminID，无版本）
func AdminUploadDocumentSimple(name, content string) error {
	if name == "" {
		name = "未命名文档"
	}
	docType := model.DocTypeProtocol
	// 简化：content 直接作为 file_url 保存
	_, err := AdminUploadDocument(0, name, docType, content, "1.0.0", model.DocRolePlayer)
	return err
}

// AdminReplaceDocumentSimple 兼容签名（仅替换内容）
func AdminReplaceDocumentSimple(documentID int64, content string) error {
	if content == "" {
		return errors.New("文件地址不能为空")
	}
	_, err := AdminReplaceDocument(documentID, 0, "", content, "", "")
	return err
}

// AdminDeleteDocumentSimple 兼容签名（单个 id 参数）
func AdminDeleteDocumentSimple(documentID int64) error {
	return AdminDeleteDocument(documentID, 0)
}

// AdminGetDocumentVersions 管理端查询文档的所有历史版本
func AdminGetDocumentVersions(documentID int64) ([]model.DocumentVersion, error) {
	var list []model.DocumentVersion
	err := db.Where("document_id = ?", documentID).Order("id DESC").Find(&list).Error
	return list, err
}
