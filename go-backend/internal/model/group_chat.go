package model

import "time"

// GroupChat 群聊表模型
type GroupChat struct {
	ID             int64      `gorm:"primaryKey;autoIncrement;column:id" json:"id"`
	GroupName      string     `gorm:"column:group_name;size:128;not null;default:''" json:"group_name"`         // 群名称
	GroupType      string     `gorm:"column:group_type;index:idx_group_type;size:32;not null;default:''" json:"group_type"` // 群类型 internal/category
	ClubID         int64      `gorm:"column:club_id;index:idx_club_id;not null;default:0" json:"club_id"`       // 俱乐部ID
	CategoryType   string     `gorm:"column:category_type;size:32;not null;default:''" json:"category_type"`   // 分类类型 chat/welfare/aftersale
	CreatorID      int64      `gorm:"column:creator_id;not null;default:0" json:"creator_id"`                  // 创建人ID
	Announcement   string     `gorm:"column:announcement;type:text" json:"announcement"`                       // 群公告
	AnnouncementAt *time.Time `gorm:"column:announcement_at" json:"announcement_at"`                           // 公告更新时间
	Status         int8       `gorm:"column:status;index:idx_status;not null;default:1" json:"status"`         // 状态 1正常 0解散
	CreatedAt      *time.Time `gorm:"column:created_at;index:idx_created_at" json:"created_at"`
	UpdatedAt      *time.Time `gorm:"column:updated_at" json:"updated_at"`
}

// TableName 指定表名
func (GroupChat) TableName() string {
	return "group_chats"
}

// 群类型常量
const (
	GroupTypeInternal = "internal" // 俱乐部内部群
	GroupTypeCategory = "category" // 分类群
)

// 分类类型常量
const (
	CategoryTypeChat     = "chat"     // 普通交流
	CategoryTypeWelfare  = "welfare"  // 福利群
	CategoryTypeAfterSale = "aftersale" // 售后群
)

// GroupChatMember 群成员表
type GroupChatMember struct {
	ID        int64      `gorm:"primaryKey;autoIncrement;column:id" json:"id"`
	GroupID   int64      `gorm:"column:group_id;index:idx_group_id;not null;default:0" json:"group_id"`     // 群ID
	UserID    int64      `gorm:"column:user_id;index:idx_user_id;not null;default:0" json:"user_id"`       // 用户ID
	Role      string     `gorm:"column:role;size:32;not null;default:'member'" json:"role"`                // 角色 member/admin/owner/platform
	IsMuted   int8       `gorm:"column:is_muted;not null;default:0" json:"is_muted"`                       // 是否禁言 0否 1是
	JoinedAt  *time.Time `gorm:"column:joined_at" json:"joined_at"`                                       // 加入时间
	CreatedAt *time.Time `gorm:"column:created_at" json:"created_at"`
	UpdatedAt *time.Time `gorm:"column:updated_at" json:"updated_at"`
}

// TableName 指定表名
func (GroupChatMember) TableName() string {
	return "group_chat_members"
}

// 群成员角色常量
const (
	GroupRoleMember   = "member"
	GroupRoleAdmin    = "admin"
	GroupRoleOwner    = "owner"
	GroupRolePlatform = "platform"
)

// GroupChatMessage 群消息表
type GroupChatMessage struct {
	ID        int64      `gorm:"primaryKey;autoIncrement;column:id" json:"id"`
	GroupID   int64      `gorm:"column:group_id;index:idx_group_id;not null;default:0" json:"group_id"`     // 群ID
	SenderID  int64      `gorm:"column:sender_id;index:idx_sender_id;not null;default:0" json:"sender_id"` // 发送者ID
	MsgType   string     `gorm:"column:msg_type;size:32;not null;default:'text'" json:"msg_type"`          // 消息类型
	Content   string     `gorm:"column:content;type:text" json:"content"`                                  // 消息内容
	MediaURL  string     `gorm:"column:media_url;size:512;not null;default:''" json:"media_url"`           // 媒体URL
	Duration  int        `gorm:"column:duration;not null;default:0" json:"duration"`                       // 时长(秒)
	AsrText   string     `gorm:"column:asr_text;type:text" json:"asr_text"`                                // 语音转文字
	IsRevoked int8       `gorm:"column:is_revoked;not null;default:0" json:"is_revoked"`                   // 是否撤回 0否 1是
	CreatedAt *time.Time `gorm:"column:created_at;index:idx_created_at" json:"created_at"`
}

// TableName 指定表名
func (GroupChatMessage) TableName() string {
	return "group_chat_messages"
}
