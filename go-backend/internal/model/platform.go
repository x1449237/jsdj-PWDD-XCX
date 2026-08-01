package model

import "time"

// PlatformOfficialAccount 平台官方账号表模型
type PlatformOfficialAccount struct {
	ID        int64      `gorm:"primaryKey;autoIncrement;column:id" json:"id"`
	Username  string     `gorm:"column:username;uniqueIndex:uk_username;size:64;not null;default:''" json:"username"` // 账号
	Nickname  string     `gorm:"column:nickname;size:64;not null;default:''" json:"nickname"`               // 昵称
	Avatar    string     `gorm:"column:avatar;size:512;not null;default:''" json:"avatar"`                  // 头像URL
	Status    int8       `gorm:"column:status;index:idx_status;not null;default:1" json:"status"`           // 状态 1正常 0停用
	CreatedAt *time.Time `gorm:"column:created_at" json:"created_at"`
	UpdatedAt *time.Time `gorm:"column:updated_at" json:"updated_at"`
}

// TableName 指定表名
func (PlatformOfficialAccount) TableName() string {
	return "platform_official_accounts"
}

// PlatformInterventionLog 平台介入日志表
type PlatformInterventionLog struct {
	ID          int64      `gorm:"primaryKey;autoIncrement;column:id" json:"id"`
	SessionID   int64      `gorm:"column:session_id;index:idx_session_id;not null;default:0" json:"session_id"` // 会话ID
	OrderID     int64      `gorm:"column:order_id;index:idx_order_id;not null;default:0" json:"order_id"`     // 订单ID
	TriggerType string     `gorm:"column:trigger_type;size:32;not null;default:''" json:"trigger_type"`       // 触发类型 keyword/manual
	HandlerID   int64      `gorm:"column:handler_id;not null;default:0" json:"handler_id"`                    // 处理人ID
	Result      string     `gorm:"column:result;size:255;not null;default:''" json:"result"`                 // 处理结果
	CreatedAt   *time.Time `gorm:"column:created_at;index:idx_created_at" json:"created_at"`
}

// TableName 指定表名
func (PlatformInterventionLog) TableName() string {
	return "platform_intervention_logs"
}

// 介入触发类型常量
const (
	InterventionTriggerKeyword = "keyword" // 关键词触发
	InterventionTriggerManual  = "manual"  // 人工触发
)
