package model

import "time"

// ChatSession 聊天会话表模型
type ChatSession struct {
	ID          int64      `gorm:"primaryKey;autoIncrement;column:id" json:"id"`
	SessionType string     `gorm:"column:session_type;index:idx_session_type;size:32;not null;default:''" json:"session_type"` // 会话类型 order/after_sale/group_internal/group_category
	RefID       int64      `gorm:"column:ref_id;index:idx_ref_id;not null;default:0" json:"ref_id"`        // 关联ID(订单ID或群ID)
	Status      int8       `gorm:"column:status;not null;default:1" json:"status"`                       // 状态 1正常 0关闭
	CreatedAt   *time.Time `gorm:"column:created_at;index:idx_created_at" json:"created_at"`
	UpdatedAt   *time.Time `gorm:"column:updated_at" json:"updated_at"`
}

// TableName 指定表名
func (ChatSession) TableName() string {
	return "chat_sessions"
}

// 会话类型常量
const (
	SessionTypeOrder          = "order"          // 订单会话
	SessionTypeAfterSale      = "after_sale"      // 售后会话
	SessionTypeGroupInternal  = "group_internal"  // 俱乐部内部群
	SessionTypeGroupCategory  = "group_category"  // 分类群
)

// ChatMessage 聊天消息表
type ChatMessage struct {
	ID         int64      `gorm:"primaryKey;autoIncrement;column:id" json:"id"`
	SessionID  int64      `gorm:"column:session_id;index:idx_session_id;not null;default:0" json:"session_id"` // 会话ID
	SenderID   int64      `gorm:"column:sender_id;index:idx_sender_id;not null;default:0" json:"sender_id"`   // 发送者ID
	SenderType string     `gorm:"column:sender_type;size:32;not null;default:''" json:"sender_type"`          // 发送者类型 user/player/club_admin/platform
	MsgType    string     `gorm:"column:msg_type;size:32;not null;default:'text'" json:"msg_type"`            // 消息类型 text/image/voice/file/card
	Content    string     `gorm:"column:content;type:text" json:"content"`                                   // 消息内容
	MediaURL   string     `gorm:"column:media_url;size:512;not null;default:''" json:"media_url"`            // 媒体URL
	Duration   int        `gorm:"column:duration;not null;default:0" json:"duration"`                        // 语音时长(秒)
	AsrText    string     `gorm:"column:asr_text;type:text" json:"asr_text"`                                 // ASR转文字结果
	IsRead     int8       `gorm:"column:is_read;not null;default:0" json:"is_read"`                          // 是否已读 0否 1是
	IsRevoked  int8       `gorm:"column:is_revoked;not null;default:0" json:"is_revoked"`                    // 是否撤回 0否 1是
	RiskLevel  int8       `gorm:"column:risk_level;not null;default:0" json:"risk_level"`                    // 风险等级 0无 1低 2中 3高
	CreatedAt  *time.Time `gorm:"column:created_at;index:idx_created_at" json:"created_at"`
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
	SessionID  int64      `gorm:"column:session_id;index:idx_session_id;not null;default:0" json:"session_id"` // 会话ID
	UploaderID int64      `gorm:"column:uploader_id;index:idx_uploader_id;not null;default:0" json:"uploader_id"` // 上传者ID
	FileURL    string     `gorm:"column:file_url;size:512;not null;default:''" json:"file_url"`            // 文件URL
	FileName   string     `gorm:"column:file_name;size:255;not null;default:''" json:"file_name"`           // 文件名
	FileSize   int64      `gorm:"column:file_size;not null;default:0" json:"file_size"`                     // 文件大小(字节)
	FileType   string     `gorm:"column:file_type;size:32;not null;default:''" json:"file_type"`            // 文件类型
	CreatedAt  *time.Time `gorm:"column:created_at;index:idx_created_at" json:"created_at"`
}

// TableName 指定表名
func (ChatFile) TableName() string {
	return "chat_files"
}
