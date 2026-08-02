package model

import "time"

// ChatSession 聊天会话表模型
type ChatSession struct {
	ID                int64      `gorm:"primaryKey;autoIncrement;column:id" json:"id"`
	SessionType       string     `gorm:"column:session_type;index:idx_session_type;size:32;not null;default:''" json:"session_type"`
	IsPinned          int8       `gorm:"column:is_pinned;not null;default:0" json:"is_pinned"`
	RefID             int64      `gorm:"column:ref_id;index:idx_ref_id;not null;default:0" json:"ref_id"`
	ClubID            int64      `gorm:"column:club_id;index:idx_club_id;not null;default:0" json:"club_id"`
	Status            int8       `gorm:"column:status;not null;default:1" json:"status"`
	PriorityLevel     int8       `gorm:"column:priority_level;index:idx_priority;not null;default:0" json:"priority_level"`       // 会话优先级 0~4
	RiskFlag          int8       `gorm:"column:risk_flag;index:idx_risk_flag;not null;default:0" json:"risk_flag"`                 // 风险标识 0~4
	LastMsgAt         *time.Time `gorm:"column:last_msg_at;index:idx_last_msg_at" json:"last_msg_at"`                              // 最后消息时间
	LastMsgPreview    string     `gorm:"column:last_msg_preview;size:512;not null;default:''" json:"last_msg_preview"`             // 预览
	UnreadCount       int        `gorm:"column:unread_count;not null;default:0" json:"unread_count"`                               // 未读数
	GroupBucket       string     `gorm:"column:group_bucket;index:idx_group_bucket;size:32;not null;default:'normal'" json:"group_bucket"`
	SilentDays        int        `gorm:"column:silent_days;not null;default:0" json:"silent_days"`                                 // 沉寂天数
	OfficialHasReply  int8       `gorm:"column:official_has_reply;not null;default:0" json:"official_has_reply"`                   // 官方已介入
	HasOfficialNotice int8       `gorm:"column:has_official_notice;not null;default:0" json:"has_official_notice"`                 // 金色小喇叭
	GameID            int        `gorm:"column:game_id;not null;default:0" json:"game_id"`                                         // 游戏ID
	CreatedAt         *time.Time `gorm:"column:created_at;index:idx_created_at" json:"created_at"`
	UpdatedAt         *time.Time `gorm:"column:updated_at" json:"updated_at"`
}

// TableName 指定表名
func (ChatSession) TableName() string {
	return "chat_sessions"
}

// 会话类型常量
const (
	SessionTypeOrder         = "order"          // 订单会话
	SessionTypeAfterSale     = "after_sale"     // 售后会话
	SessionTypeGroupInternal = "group_internal" // 俱乐部内部群
	SessionTypeGroupCategory = "group_category" // 分类群
)

// ChatMessage 聊天消息表
type ChatMessage struct {
	ID                  int64      `gorm:"primaryKey;autoIncrement;column:id" json:"id"`
	SessionID           int64      `gorm:"column:session_id;index:idx_session_id;not null;default:0" json:"session_id"`
	SenderID            int64      `gorm:"column:sender_id;index:idx_sender_id;not null;default:0" json:"sender_id"`
	SenderType          string     `gorm:"column:sender_type;size:32;not null;default:''" json:"sender_type"`
	MsgType             string     `gorm:"column:msg_type;size:32;not null;default:'text'" json:"msg_type"`
	Content             string     `gorm:"column:content;type:text" json:"content"`
	MediaURL            string     `gorm:"column:media_url;size:512;not null;default:''" json:"media_url"`
	Duration            int        `gorm:"column:duration;not null;default:0" json:"duration"`
	AsrText             string     `gorm:"column:asr_text;type:text" json:"asr_text"`
	IsRead              int8       `gorm:"column:is_read;not null;default:0" json:"is_read"`
	IsRevoked           int8       `gorm:"column:is_revoked;not null;default:0" json:"is_revoked"`
	RiskLevel           int8       `gorm:"column:risk_level;not null;default:0" json:"risk_level"`
	IsKeyMessage        int8       `gorm:"column:is_key_message;not null;default:0" json:"is_key_message"`               // 关键消息左侧竖线
	IsImportantDirective int8      `gorm:"column:is_important_directive;not null;default:0" json:"is_important_directive"` // 重要指令双描边
	CreatedAt           *time.Time `gorm:"column:created_at;index:idx_created_at" json:"created_at"`
}

// TableName 指定表名
func (ChatMessage) TableName() string {
	return "chat_messages"
}

// 消息类型常量
const (
	MsgTypeText  = "text"
	MsgTypeImage = "image"
	MsgTypeVoice = "voice"
	MsgTypeFile  = "file"
	MsgTypeCard  = "card"
)

// ChatFile 聊天文件传输表
type ChatFile struct {
	ID         int64      `gorm:"primaryKey;autoIncrement;column:id" json:"id"`
	SessionID  int64      `gorm:"column:session_id;index:idx_session_id;not null;default:0" json:"session_id"`    // 会话ID
	UploaderID int64      `gorm:"column:uploader_id;index:idx_uploader_id;not null;default:0" json:"uploader_id"` // 上传者ID
	FileURL    string     `gorm:"column:file_url;size:512;not null;default:''" json:"file_url"`                   // 文件URL
	FileName   string     `gorm:"column:file_name;size:255;not null;default:''" json:"file_name"`                 // 文件名
	FileSize   int64      `gorm:"column:file_size;not null;default:0" json:"file_size"`                           // 文件大小(字节)
	FileType   string     `gorm:"column:file_type;size:32;not null;default:''" json:"file_type"`                  // 文件类型
	CreatedAt  *time.Time `gorm:"column:created_at;index:idx_created_at" json:"created_at"`
}

// TableName 指定表名
func (ChatFile) TableName() string {
	return "chat_files"
}
