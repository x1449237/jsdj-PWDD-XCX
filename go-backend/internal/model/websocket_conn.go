package model

import (
	"encoding/json"
	"time"
)

// WsConnection WebSocket连接表模型
type WsConnection struct {
	ID         int64      `gorm:"primaryKey;autoIncrement;column:id" json:"id"`
	UserID     int64      `gorm:"column:user_id;index:idx_user_id;not null;default:0" json:"user_id"`     // 用户ID
	UserType   string     `gorm:"column:user_type;size:32;not null;default:''" json:"user_type"`          // 用户类型 user/admin/platform
	ConnID     string     `gorm:"column:conn_id;index:idx_conn_id;size:64;not null;default:''" json:"conn_id"` // 连接ID
	LastPingAt *time.Time `gorm:"column:last_ping_at" json:"last_ping_at"`                                // 最后心跳时间
	IsActive   int8       `gorm:"column:is_active;index:idx_is_active;not null;default:1" json:"is_active"` // 是否活跃 0否 1是
	CreatedAt  *time.Time `gorm:"column:created_at" json:"created_at"`
	UpdatedAt  *time.Time `gorm:"column:updated_at" json:"updated_at"`
}

// TableName 指定表名
func (WsConnection) TableName() string {
	return "ws_connections"
}

// WebSocket 用户类型常量
const (
	WsUserTypeUser     = "user"     // 普通用户
	WsUserTypeAdmin    = "admin"    // 管理员
	WsUserTypePlatform = "platform" // 平台官方账号
)

// OfflineMessage 离线消息表
type OfflineMessage struct {
	ID          int64           `gorm:"primaryKey;autoIncrement;column:id" json:"id"`
	UserID      int64           `gorm:"column:user_id;index:idx_user_id;not null;default:0" json:"user_id"`     // 用户ID
	SessionID   int64           `gorm:"column:session_id;index:idx_session_id;not null;default:0" json:"session_id"` // 会话ID
	MessageData json.RawMessage `gorm:"column:message_data;type:json" json:"message_data"`                       // 消息数据
	CreatedAt   *time.Time      `gorm:"column:created_at;index:idx_created_at" json:"created_at"`
}

// TableName 指定表名
func (OfflineMessage) TableName() string {
	return "offline_messages"
}
