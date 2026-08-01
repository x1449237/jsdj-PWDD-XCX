package model

import (
	"encoding/json"
	"time"
)

// SystemConfig 系统配置表模型
type SystemConfig struct {
	ID          int64     `gorm:"primaryKey;autoIncrement;column:id" json:"id"`
	Key         string    `gorm:"column:key;uniqueIndex:uk_key;size:128;not null;default:''" json:"key"` // 配置键
	Value       string    `gorm:"column:value;type:text" json:"value"`                                  // 配置值
	Description string    `gorm:"column:description;size:255;not null;default:''" json:"description"`  // 配置描述
	UpdatedAt   *time.Time `gorm:"column:updated_at" json:"updated_at"`
}

// TableName 指定表名
func (SystemConfig) TableName() string {
	return "system_configs"
}

// OperationLog 操作日志表
type OperationLog struct {
	ID           int64           `gorm:"primaryKey;autoIncrement;column:id" json:"id"`
	OperatorID   int64           `gorm:"column:operator_id;index:idx_operator_id;not null;default:0" json:"operator_id"` // 操作人ID
	OperatorType string          `gorm:"column:operator_type;size:32;not null;default:''" json:"operator_type"`        // 操作人类型 admin/shop_admin
	Action       string          `gorm:"column:action;index:idx_action;size:64;not null;default:''" json:"action"`    // 操作动作
	TargetType   string          `gorm:"column:target_type;index:idx_target_type;size:64;not null;default:''" json:"target_type"` // 操作对象类型
	TargetID     int64           `gorm:"column:target_id;not null;default:0" json:"target_id"`                         // 操作对象ID
	Content      json.RawMessage `gorm:"column:content;type:json" json:"content"`                                      // 操作内容
	IP           string          `gorm:"column:ip;size:64;not null;default:''" json:"ip"`                             // IP地址
	DeviceInfo   string          `gorm:"column:device_info;size:255;not null;default:''" json:"device_info"`          // 设备信息
	CreatedAt    *time.Time      `gorm:"column:created_at;index:idx_created_at" json:"created_at"`
}

// TableName 指定表名
func (OperationLog) TableName() string {
	return "operation_logs"
}

// Notification 通知表
type Notification struct {
	ID        int64      `gorm:"primaryKey;autoIncrement;column:id" json:"id"`
	UserID    int64      `gorm:"column:user_id;index:idx_user_id;not null;default:0" json:"user_id"`     // 用户ID
	Type      string     `gorm:"column:type;size:32;not null;default:''" json:"type"`                    // 通知类型
	Title     string     `gorm:"column:title;size:128;not null;default:''" json:"title"`                 // 标题
	Content   string     `gorm:"column:content;type:text" json:"content"`                                // 内容
	IsRead    int8       `gorm:"column:is_read;index:idx_is_read;not null;default:0" json:"is_read"`     // 是否已读 0否 1是
	Category  string     `gorm:"column:category;index:idx_category;size:32;not null;default:'system'" json:"category"` // 分类 pending/supervision/system
	CreatedAt *time.Time `gorm:"column:created_at;index:idx_created_at" json:"created_at"`
	UpdatedAt *time.Time `gorm:"column:updated_at" json:"updated_at"`
}

// TableName 指定表名
func (Notification) TableName() string {
	return "notifications"
}

// 通知分类常量
const (
	NotificationCategoryPending     = "pending"     // 待办
	NotificationCategorySupervision = "supervision" // 监管
	NotificationCategorySystem      = "system"      // 系统
)
