package model

import (
	"encoding/json"
	"time"
)

// Appeal 申诉工单表模型
type Appeal struct {
	ID          int64           `gorm:"primaryKey;autoIncrement;column:id" json:"id"`
	OrderID     int64           `gorm:"column:order_id;index:idx_order_id;not null;default:0" json:"order_id"`   // 订单ID
	UserID      int64           `gorm:"column:user_id;index:idx_user_id;not null;default:0" json:"user_id"`     // 用户ID
	Type        string          `gorm:"column:type;size:32;not null;default:''" json:"type"`                    // 申诉类型
	Description string          `gorm:"column:description;type:text" json:"description"`                        // 问题描述
	Status      string          `gorm:"column:status;index:idx_status;size:32;not null;default:'pending'" json:"status"` // 状态 pending/processing/resolved/rejected
	EvidenceURLs json.RawMessage `gorm:"column:evidence_urls;type:json" json:"evidence_urls"`                    // 证据URL数组
	CreatedAt   *time.Time      `gorm:"column:created_at;index:idx_created_at" json:"created_at"`
	UpdatedAt   *time.Time      `gorm:"column:updated_at" json:"updated_at"`
}

// TableName 指定表名
func (Appeal) TableName() string {
	return "appeals"
}

// 申诉状态常量
const (
	AppealStatusPending    = "pending"
	AppealStatusProcessing = "processing"
	AppealStatusResolved   = "resolved"
	AppealStatusRejected   = "rejected"
)

// AppealCommunication 申诉沟通表
type AppealCommunication struct {
	ID        int64      `gorm:"primaryKey;autoIncrement;column:id" json:"id"`
	AppealID  int64      `gorm:"column:appeal_id;index:idx_appeal_id;not null;default:0" json:"appeal_id"` // 申诉ID
	SenderID  int64      `gorm:"column:sender_id;index:idx_sender_id;not null;default:0" json:"sender_id"` // 发送者ID
	Content   string     `gorm:"column:content;type:text" json:"content"`                                 // 沟通内容
	CreatedAt *time.Time `gorm:"column:created_at;index:idx_created_at" json:"created_at"`
}

// TableName 指定表名
func (AppealCommunication) TableName() string {
	return "appeal_communications"
}
