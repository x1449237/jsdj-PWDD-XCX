package model

import "time"

// PlatformDocument 平台文档/协议表模型
type PlatformDocument struct {
	ID        int64      `gorm:"primaryKey;autoIncrement;column:id" json:"id"`
	Name      string     `gorm:"column:name;size:128;not null;default:''" json:"name"`
	DocType   string     `gorm:"column:doc_type;index:idx_doc_type;size:32;not null;default:''" json:"doc_type"`
	FileUrl   string     `gorm:"column:file_url;size:512;not null;default:''" json:"file_url"`
	Version   string     `gorm:"column:version;size:32;not null;default:'1.0.0'" json:"version"`
	Role      string     `gorm:"column:role;index:idx_role;size:32;not null;default:'player'" json:"role"`
	IsDeleted int8       `gorm:"column:is_deleted;index:idx_is_deleted;not null;default:0" json:"is_deleted"`
	CreatedBy int64      `gorm:"column:created_by;not null;default:0" json:"created_by"`
	CreatedAt *time.Time `gorm:"column:created_at;index:idx_created_at" json:"created_at"`
	UpdatedAt *time.Time `gorm:"column:updated_at" json:"updated_at"`
}

// TableName 指定表名
func (PlatformDocument) TableName() string {
	return "platform_documents"
}

// 文档类型常量
const (
	DocTypeProtocol = "protocol"
	DocTypePolicy   = "policy"
	DocTypeContract = "contract"
)

// 文档适用角色常量
const (
	DocRolePlayer      = "player"
	DocRoleCustomer    = "customer"
	DocRoleDistributor = "distributor"
	DocRoleClub        = "club"
)

// DocumentVersion 文档版本历史
type DocumentVersion struct {
	ID         int64      `gorm:"primaryKey;autoIncrement;column:id" json:"id"`
	DocumentID int64      `gorm:"column:document_id;index:idx_document_id;not null;default:0" json:"document_id"`
	FileUrl    string     `gorm:"column:file_url;size:512;not null;default:''" json:"file_url"`
	Version    string     `gorm:"column:version;size:32;not null;default:'1.0.0'" json:"version"`
	CreatedBy  int64      `gorm:"column:created_by;not null;default:0" json:"created_by"`
	CreatedAt  *time.Time `gorm:"column:created_at;index:idx_created_at" json:"created_at"`
}

// TableName 指定表名
func (DocumentVersion) TableName() string {
	return "document_versions"
}

// AgreementSignLog 协议签署流水
type AgreementSignLog struct {
	ID         int64      `gorm:"primaryKey;autoIncrement;column:id" json:"id"`
	UID        int64      `gorm:"column:uid;index:idx_uid;not null;default:0" json:"uid"`
	DocumentID int64      `gorm:"column:document_id;index:idx_document_id;not null;default:0" json:"document_id"`
	Role       string     `gorm:"column:role;size:32;not null;default:''" json:"role"`
	SignIP     string     `gorm:"column:sign_ip;size:64;not null;default:''" json:"sign_ip"`
	SignedAt   *time.Time `gorm:"column:signed_at" json:"signed_at"`
}

// TableName 指定表名
func (AgreementSignLog) TableName() string {
	return "agreement_sign_logs"
}
