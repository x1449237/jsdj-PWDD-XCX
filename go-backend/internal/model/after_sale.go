package model

import "time"

// AfterSaleSession 售后会话表模型
type AfterSaleSession struct {
	ID                  int64      `gorm:"primaryKey;autoIncrement;column:id" json:"id"`
	OrderID             int64      `gorm:"column:order_id;index:idx_order_id;not null;default:0" json:"order_id"`     // 订单ID
	UserID              int64      `gorm:"column:user_id;index:idx_user_id;not null;default:0" json:"user_id"`       // 用户ID
	ClubID              int64      `gorm:"column:club_id;not null;default:0" json:"club_id"`                         // 俱乐部ID
	Status              int8       `gorm:"column:status;index:idx_status;not null;default:1" json:"status"`          // 状态 1处理中 2已关闭
	InterventionStatus  int8       `gorm:"column:intervention_status;index:idx_intervention_status;not null;default:0" json:"intervention_status"` // 介入状态 0=未介入 1=待介入 2=已介入
	InterventionType    string     `gorm:"column:intervention_type;size:32;not null;default:''" json:"intervention_type"` // 介入类型 keyword/manual
	CreatedAt           *time.Time `gorm:"column:created_at" json:"created_at"`
	UpdatedAt           *time.Time `gorm:"column:updated_at" json:"updated_at"`
}

// TableName 指定表名
func (AfterSaleSession) TableName() string {
	return "after_sale_sessions"
}

// 售后介入状态常量
const (
	AfterSaleInterventionNone        int8 = 0 // 未介入
	AfterSaleInterventionPending     int8 = 1 // 待介入
	AfterSaleInterventionInvolved    int8 = 2 // 已介入
)

// AfterSaleMessage 售后消息表
type AfterSaleMessage struct {
	ID        int64      `gorm:"primaryKey;autoIncrement;column:id" json:"id"`
	SessionID int64      `gorm:"column:session_id;index:idx_session_id;not null;default:0" json:"session_id"` // 售后会话ID
	SenderID  int64      `gorm:"column:sender_id;index:idx_sender_id;not null;default:0" json:"sender_id"`   // 发送者ID
	Content   string     `gorm:"column:content;type:text" json:"content"`                                   // 消息内容
	MsgType   string     `gorm:"column:msg_type;size:32;not null;default:'text'" json:"msg_type"`            // 消息类型
	MediaURL  string     `gorm:"column:media_url;size:512;not null;default:''" json:"media_url"`            // 媒体URL
	CreatedAt *time.Time `gorm:"column:created_at;index:idx_created_at" json:"created_at"`
}

// TableName 指定表名
func (AfterSaleMessage) TableName() string {
	return "after_sale_messages"
}

// AfterSaleKeyword 售后风控关键词表
type AfterSaleKeyword struct {
	ID         int64      `gorm:"primaryKey;autoIncrement;column:id" json:"id"`
	Keyword    string     `gorm:"column:keyword;size:128;not null;default:''" json:"keyword"`              // 关键词
	MatchType  string     `gorm:"column:match_type;size:32;not null;default:'exact'" json:"match_type"`   // 匹配类型 exact/fuzzy
	Enabled    int8       `gorm:"column:enabled;index:idx_enabled;not null;default:1" json:"enabled"`     // 是否启用 0否 1是
	CreatedAt  *time.Time `gorm:"column:created_at" json:"created_at"`
	UpdatedAt  *time.Time `gorm:"column:updated_at" json:"updated_at"`
}

// TableName 指定表名
func (AfterSaleKeyword) TableName() string {
	return "after_sale_keywords"
}

// 关键词匹配类型常量
const (
	KeywordMatchExact = "exact"
	KeywordMatchFuzzy = "fuzzy"
)
