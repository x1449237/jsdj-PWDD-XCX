package model

import (
	"time"

	"gorm.io/datatypes"
)

// GroupChat 群聊表模型
type GroupChat struct {
	ID                   int64      `gorm:"primaryKey;autoIncrement;column:id" json:"id"`
	GroupName            string     `gorm:"column:group_name;size:128;not null;default:''" json:"group_name"`
	PinnedAnnouncementID int64      `gorm:"column:pinned_announcement_id;not null;default:0" json:"pinned_announcement_id"`
	GroupType            string     `gorm:"column:group_type;index:idx_group_type;size:32;not null;default:''" json:"group_type"`
	ClubID               int64      `gorm:"column:club_id;index:idx_club_id;not null;default:0" json:"club_id"`
	CategoryType         string     `gorm:"column:category_type;size:32;not null;default:''" json:"category_type"`
	CreatorID            int64      `gorm:"column:creator_id;not null;default:0" json:"creator_id"`
	Announcement         string     `gorm:"column:announcement;type:text" json:"announcement"`
	AnnouncementAt       *time.Time `gorm:"column:announcement_at" json:"announcement_at"`
	Status               int8       `gorm:"column:status;index:idx_status;not null;default:1" json:"status"`
	GroupTypeLabel       string     `gorm:"column:group_type_label;size:32;not null;default:''" json:"group_type_label"`
	WelcomeMsg           string     `gorm:"column:welcome_msg;type:text" json:"welcome_msg"`
	PrevOwnerUID         int64      `gorm:"column:prev_owner_uid;not null;default:0" json:"prev_owner_uid"`
	MuteNotify           int8       `gorm:"column:mute_notify;not null;default:0" json:"mute_notify"`             // 免打扰
	IsHidden             int8       `gorm:"column:is_hidden;not null;default:0" json:"is_hidden"`                 // 隐藏废弃
	AggregateSameClub    int8       `gorm:"column:aggregate_same_club;not null;default:0" json:"aggregate_same_club"` // 同俱乐部聚合
	LastMsgAt            *time.Time `gorm:"column:last_msg_at" json:"last_msg_at"`                                // 最后发言时间
	LastMsgPreview       string     `gorm:"column:last_msg_preview;size:512;not null;default:''" json:"last_msg_preview"`
	UnreadCount          int        `gorm:"column:unread_count;not null;default:0" json:"unread_count"`
	CreatedAt            *time.Time `gorm:"column:created_at;index:idx_created_at" json:"created_at"`
	UpdatedAt            *time.Time `gorm:"column:updated_at" json:"updated_at"`
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
	CategoryTypeChat      = "chat"      // 普通交流
	CategoryTypeWelfare   = "welfare"   // 福利群
	CategoryTypeAfterSale = "aftersale" // 售后群
)

// GroupChatMember 群成员表
type GroupChatMember struct {
	ID          int64      `gorm:"primaryKey;autoIncrement;column:id" json:"id"`
	GroupID     int64      `gorm:"column:group_id;index:idx_group_id;not null;default:0" json:"group_id"` // 群ID
	UserID      int64      `gorm:"column:user_id;index:idx_user_id;not null;default:0" json:"user_id"`    // 用户ID
	Role        string     `gorm:"column:role;size:32;not null;default:'member'" json:"role"`             // 角色 member/admin/owner/platform
	MemberLevel int8       `gorm:"column:member_level;not null;default:0" json:"member_level"`            // 成员级别
	IsMuted     int8       `gorm:"column:is_muted;not null;default:0" json:"is_muted"`                    // 是否禁言 0否 1是
	JoinedAt    *time.Time `gorm:"column:joined_at" json:"joined_at"`                                     // 加入时间
	CreatedAt   *time.Time `gorm:"column:created_at" json:"created_at"`
	UpdatedAt   *time.Time `gorm:"column:updated_at" json:"updated_at"`
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
	ID                   int64          `gorm:"primaryKey;autoIncrement;column:id" json:"id"`
	GroupID              int64          `gorm:"column:group_id;index:idx_group_id;not null;default:0" json:"group_id"`
	SenderID             int64          `gorm:"column:sender_id;index:idx_sender_id;not null;default:0" json:"sender_id"`
	MsgType              string         `gorm:"column:msg_type;size:32;not null;default:'text'" json:"msg_type"`
	AtAll                int8           `gorm:"column:at_all;not null;default:0" json:"at_all"`
	Content              string         `gorm:"column:content;type:text" json:"content"`
	MediaURL             string         `gorm:"column:media_url;size:512;not null;default:''" json:"media_url"`
	Duration             int            `gorm:"column:duration;not null;default:0" json:"duration"`
	AsrText              string         `gorm:"column:asr_text;type:text" json:"asr_text"`
	IsRevoked            int8           `gorm:"column:is_revoked;not null;default:0" json:"is_revoked"`
	AtUIDs               datatypes.JSON `gorm:"column:at_uids" json:"at_uids"`
	IsForcePopup         int8           `gorm:"column:is_force_popup;not null;default:0" json:"is_force_popup"`
	IsKeyMessage         int8           `gorm:"column:is_key_message;not null;default:0" json:"is_key_message"`                 // 关键消息
	IsImportantDirective int8           `gorm:"column:is_important_directive;not null;default:0" json:"is_important_directive"` // 重要指令
	CreatedAt            *time.Time     `gorm:"column:created_at;index:idx_created_at" json:"created_at"`
}

// TableName 指定表名
func (GroupChatMessage) TableName() string {
	return "group_chat_messages"
}
