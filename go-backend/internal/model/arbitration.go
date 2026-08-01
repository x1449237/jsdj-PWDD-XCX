package model

import (
	"encoding/json"
	"time"
)

// ArbitrationCase 仲裁案件表模型
type ArbitrationCase struct {
	ID            int64      `gorm:"primaryKey;autoIncrement;column:id" json:"id"`
	OrderID       int64      `gorm:"column:order_id;index:idx_order_id;not null;default:0" json:"order_id"`   // 订单ID
	SessionID     int64      `gorm:"column:session_id;index:idx_session_id;not null;default:0" json:"session_id"` // 会话ID
	Status        string     `gorm:"column:status;index:idx_status;size:32;not null;default:'pending'" json:"status"` // 状态 pending/arbitrating/closed
	ArbitratorID  int64      `gorm:"column:arbitrator_id;not null;default:0" json:"arbitrator_id"`            // 仲裁员ID
	Result        string     `gorm:"column:result;size:255;not null;default:''" json:"result"`               // 仲裁结果
	CreatedAt     *time.Time `gorm:"column:created_at;index:idx_created_at" json:"created_at"`
	UpdatedAt     *time.Time `gorm:"column:updated_at" json:"updated_at"`
}

// TableName 指定表名
func (ArbitrationCase) TableName() string {
	return "arbitration_cases"
}

// 仲裁案件状态常量
const (
	ArbitrationStatusPending     = "pending"
	ArbitrationStatusArbitrating = "arbitrating"
	ArbitrationStatusClosed      = "closed"
)

// ArbitrationEvidence 仲裁证据表
type ArbitrationEvidence struct {
	ID          int64      `gorm:"primaryKey;autoIncrement;column:id" json:"id"`
	CaseID      int64      `gorm:"column:case_id;index:idx_case_id;not null;default:0" json:"case_id"`      // 案件ID
	UploaderID  int64      `gorm:"column:uploader_id;index:idx_uploader_id;not null;default:0" json:"uploader_id"` // 上传人ID
	Type        string     `gorm:"column:type;size:32;not null;default:''" json:"type"`                    // 证据类型 video/screenshot/chat
	FileURL     string     `gorm:"column:file_url;size:512;not null;default:''" json:"file_url"`           // 文件URL
	Description string     `gorm:"column:description;size:255;not null;default:''" json:"description"`     // 描述
	CreatedAt   *time.Time `gorm:"column:created_at;index:idx_created_at" json:"created_at"`
}

// TableName 指定表名
func (ArbitrationEvidence) TableName() string {
	return "arbitration_evidences"
}

// 仲裁证据类型常量
const (
	ArbitrationEvidenceTypeVideo     = "video"
	ArbitrationEvidenceTypeScreenshot = "screenshot"
	ArbitrationEvidenceTypeChat      = "chat"
)

// ArbitrationRule 判责规则表
type ArbitrationRule struct {
	ID            int64           `gorm:"primaryKey;autoIncrement;column:id" json:"id"`
	Name          string          `gorm:"column:name;size:64;not null;default:''" json:"name"`                    // 规则名称
	Condition     json.RawMessage `gorm:"column:condition;type:json" json:"condition"`                            // 触发条件
	Responsibility string         `gorm:"column:responsibility;index:idx_responsibility;size:32;not null;default:''" json:"responsibility"` // 责任方 player/user/club/both
	Penalty       string          `gorm:"column:penalty;size:255;not null;default:''" json:"penalty"`             // 处罚措施
	CreatedAt     *time.Time      `gorm:"column:created_at" json:"created_at"`
	UpdatedAt     *time.Time      `gorm:"column:updated_at" json:"updated_at"`
}

// TableName 指定表名
func (ArbitrationRule) TableName() string {
	return "arbitration_rules"
}

// 判责责任方常量
const (
	ArbitrationResponsibilityPlayer = "player"
	ArbitrationResponsibilityUser   = "user"
	ArbitrationResponsibilityClub   = "club"
	ArbitrationResponsibilityBoth   = "both"
)
