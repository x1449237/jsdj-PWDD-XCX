package model

import (
	"time"

	"gorm.io/datatypes"
)

// ============================================================
// 平台方管理人员 IM 会话归类清单 - 专用模型
// ============================================================

// 会话优先级常量
const (
	SessPriorityNormal       int8 = 0
	SessPriorityNewSale      int8 = 1
	SessPriorityTimeoutEvid  int8 = 2
	SessPriorityPlayerInt    int8 = 3
	SessPrioritySensitiveTop int8 = 4
)

// 会话风险标识
const (
	RiskFlagNone      int8 = 0
	RiskFlagSensitive int8 = 1
	RiskFlagPlayerInt int8 = 2
	RiskFlagTimeout   int8 = 3
	RiskFlagClosed    int8 = 4
)

// 会话分组桶
const (
	BucketRiskTop     = "risk_top"      // 敏感词预警
	BucketPlayerInt   = "player_int"    // 玩家介入
	BucketTimeout     = "timeout"       // 超时举证
	BucketNormalSale  = "normal_sale"   // 新消息售后
	BucketClubChat    = "club_chat"     // 俱乐部队群聊
	BucketSilentSale  = "silent_sale"   // 沉寂售后(>7天)
	BucketSilentClub  = "silent_club"   // 长期沉寂群(>15天)
	BucketClosed      = "closed"        // 已办结
	BucketStarred     = "starred"       // 稍后处理星标(虚拟分组)
)

// 工作台任务桶
const (
	TaskBucketEmergency = "emergency" // 今日紧急
	TaskBucketTodo      = "todo"      // 待跟进
	TaskBucketYesterday = "yesterday" // 昨日遗留
)

// 快捷话术分类
const (
	QuickReplySoothe   = "soothe"   // 安抚玩家
	QuickReplyEvidence = "evidence" // 要求俱乐部举证
	QuickReplyNotice   = "notice"   // 仲裁时限
)

// 头像框标识
const (
	FramePlatformGold = "platform_gold"
	FrameFounderBlue  = "founder_blue"
	FrameAdminGray    = "admin_gray"
)

// 气泡角色标识
const (
	BubbleRolePlatform   = "platform"
	BubbleRoleClubAdmin  = "club_admin"
	BubbleRoleUser       = "user"
	BubbleRolePlayer     = "player"
)

// -------- 会话自定义标签 --------

// ImSessionTag 会话-标签关联
type ImSessionTag struct {
	ID        int64      `gorm:"primaryKey;autoIncrement;column:id" json:"id"`
	SessionID int64      `gorm:"column:session_id;uniqueIndex:uk_session_tag;not null;default:0" json:"session_id"`
	TagID     int64      `gorm:"column:tag_id;uniqueIndex:uk_session_tag;not null;default:0" json:"tag_id"`
	TagName   string     `gorm:"column:tag_name;size:64;not null;default:''" json:"tag_name"`
	TagColor  string     `gorm:"column:tag_color;size:16;not null;default:''" json:"tag_color"`
	CreatedBy int64      `gorm:"column:created_by;not null;default:0" json:"created_by"`
	CreatedAt *time.Time `gorm:"column:created_at" json:"created_at"`
}

func (ImSessionTag) TableName() string { return "im_session_tags" }

// ImTagDefinition 标签定义
type ImTagDefinition struct {
	ID        int64      `gorm:"primaryKey;autoIncrement;column:id" json:"id"`
	Name      string     `gorm:"column:name;size:64;index:idx_name;not null;default:''" json:"name"`
	Color     string     `gorm:"column:color;size:16;not null;default:'#409EFF'" json:"color"`
	Sort      int        `gorm:"column:sort;not null;default:0" json:"sort"`
	CreatedBy int64      `gorm:"column:created_by;not null;default:0" json:"created_by"`
	IsSystem  int8       `gorm:"column:is_system;not null;default:0" json:"is_system"`
	CreatedAt *time.Time `gorm:"column:created_at" json:"created_at"`
	UpdatedAt *time.Time `gorm:"column:updated_at" json:"updated_at"`
}

func (ImTagDefinition) TableName() string { return "im_tag_definitions" }

// -------- 会话个人备注与星标 --------

// ImSessionNote 会话备注
type ImSessionNote struct {
	ID         int64      `gorm:"primaryKey;autoIncrement;column:id" json:"id"`
	SessionID  int64      `gorm:"column:session_id;not null;default:0" json:"session_id"`
	PlatformUID int64     `gorm:"column:platform_uid;not null;default:0" json:"platform_uid"`
	Content    string     `gorm:"column:content;size:1024;not null;default:''" json:"content"`
	IsStarred  int8       `gorm:"column:is_starred;index:idx_is_starred;not null;default:0" json:"is_starred"`
	CreatedAt  *time.Time `gorm:"column:created_at" json:"created_at"`
	UpdatedAt  *time.Time `gorm:"column:updated_at" json:"updated_at"`
}

func (ImSessionNote) TableName() string { return "im_session_notes" }

// -------- 搜索历史 --------

// ImSearchHistory 搜索历史
type ImSearchHistory struct {
	ID          int64      `gorm:"primaryKey;autoIncrement;column:id" json:"id"`
	PlatformUID int64      `gorm:"column:platform_uid;index:idx_platform_uid;not null;default:0" json:"platform_uid"`
	Keyword     string     `gorm:"column:keyword;size:255;not null;default:''" json:"keyword"`
	SearchType  string     `gorm:"column:search_type;size:32;not null;default:'all'" json:"search_type"`
	CreatedAt   *time.Time `gorm:"column:created_at;index:idx_created_at" json:"created_at"`
}

func (ImSearchHistory) TableName() string { return "im_search_histories" }

// -------- 工作台 --------

// ImWorkbenchTask 工作台任务
type ImWorkbenchTask struct {
	ID          int64      `gorm:"primaryKey;autoIncrement;column:id" json:"id"`
	PlatformUID int64      `gorm:"column:platform_uid;not null;default:0" json:"platform_uid"`
	SessionID   int64      `gorm:"column:session_id;not null;default:0" json:"session_id"`
	TaskBucket  string     `gorm:"column:task_bucket;size:32;index:idx_platform_bucket;not null;default:'pending'" json:"task_bucket"`
	Title       string     `gorm:"column:title;size:255;not null;default:''" json:"title"`
	TaskType    string     `gorm:"column:task_type;size:32;not null;default:''" json:"task_type"`
	DeadlineAt  *time.Time `gorm:"column:deadline_at" json:"deadline_at"`
	IsDone      int8       `gorm:"column:is_done;index:idx_is_done;not null;default:0" json:"is_done"`
	CreatedAt   *time.Time `gorm:"column:created_at" json:"created_at"`
	UpdatedAt   *time.Time `gorm:"column:updated_at" json:"updated_at"`
}

func (ImWorkbenchTask) TableName() string { return "im_workbench_tasks" }

// ImWorkbenchLayout 工作台布局
type ImWorkbenchLayout struct {
	ID          int64          `gorm:"primaryKey;autoIncrement;column:id" json:"id"`
	PlatformUID int64          `gorm:"column:platform_uid;uniqueIndex:uk_platform_uid;not null;default:0" json:"platform_uid"`
	BucketOrder datatypes.JSON `gorm:"column:bucket_order" json:"bucket_order"`
	UpdatedAt   *time.Time     `gorm:"column:updated_at" json:"updated_at"`
}

func (ImWorkbenchLayout) TableName() string { return "im_workbench_layouts" }

// -------- 快捷话术 --------

// ImQuickReply 快捷话术
type ImQuickReply struct {
	ID        int64      `gorm:"primaryKey;autoIncrement;column:id" json:"id"`
	Category  string     `gorm:"column:category;index:idx_category;size:32;not null;default:'soothe'" json:"category"`
	Title     string     `gorm:"column:title;size:128;not null;default:''" json:"title"`
	Content   string     `gorm:"column:content;type:text" json:"content"`
	Sort      int        `gorm:"column:sort;not null;default:0" json:"sort"`
	IsSystem  int8       `gorm:"column:is_system;not null;default:1" json:"is_system"`
	CreatedAt *time.Time `gorm:"column:created_at" json:"created_at"`
	UpdatedAt *time.Time `gorm:"column:updated_at" json:"updated_at"`
}

func (ImQuickReply) TableName() string { return "im_quick_replies" }

// -------- 证据打包预览 --------

// ImEvidencePack 证据打包记录
type ImEvidencePack struct {
	ID         int64          `gorm:"primaryKey;autoIncrement;column:id" json:"id"`
	SessionID  int64          `gorm:"column:session_id;index:idx_session_id;not null;default:0" json:"session_id"`
	MessageIDs datatypes.JSON `gorm:"column:message_ids" json:"message_ids"`
	CreatedBy  int64          `gorm:"column:created_by;not null;default:0" json:"created_by"`
	CreatedAt  *time.Time     `gorm:"column:created_at" json:"created_at"`
}

func (ImEvidencePack) TableName() string { return "im_evidence_packs" }

// -------- 聊天气泡样式配置 (后端统管) --------

// ImBubbleStyle 气泡样式
type ImBubbleStyle struct {
	ID                    int64      `gorm:"primaryKey;autoIncrement;column:id" json:"id"`
	RoleKey               string     `gorm:"column:role_key;uniqueIndex:uk_role_key;size:32;not null;default:''" json:"role_key"`
	BubbleBG              string     `gorm:"column:bubble_bg;size:64;not null;default:'#FFFFFF'" json:"bubble_bg"`
	BubbleRadius          int        `gorm:"column:bubble_radius;not null;default:12" json:"bubble_radius"`
	BubbleShadow          string     `gorm:"column:bubble_shadow;size:128;not null;default:''" json:"bubble_shadow"`
	TextColor             string     `gorm:"column:text_color;size:32;not null;default:'#303133'" json:"text_color"`
	TextStrokeColor       string     `gorm:"column:text_stroke_color;size:32;not null;default:''" json:"text_stroke_color"`
	TextStrokeWidth       float64    `gorm:"column:text_stroke_width;type:decimal(3,1);not null;default:0" json:"text_stroke_width"`
	TextBoldImportant     int8       `gorm:"column:text_bold_important;not null;default:0" json:"text_bold_important"`
	ImportantTextColor    string     `gorm:"column:important_text_color;size:32;not null;default:'#FFFFFF'" json:"important_text_color"`
	ImportantStrokeWidth  float64    `gorm:"column:important_stroke_width;type:decimal(3,1);not null;default:0" json:"important_stroke_width"`
	VoiceWaveColor        string     `gorm:"column:voice_wave_color;size:32;not null;default:'#C0C4CC'" json:"voice_wave_color"`
	LongTextLineHeight    float64    `gorm:"column:long_text_line_height;type:decimal(3,1);not null;default:1.6" json:"long_text_line_height"`
	BrightnessLocked      int8       `gorm:"column:brightness_locked;not null;default:1" json:"brightness_locked"`
	LockFontOpacity       int8       `gorm:"column:lock_font_opacity;not null;default:1" json:"lock_font_opacity"`
	VBadgeSizeMultiple    float64    `gorm:"column:v_badge_size_multiple;type:decimal(3,2);not null;default:1.00" json:"v_badge_size_multiple"`
	NickFontWeight        string     `gorm:"column:nick_font_weight;size:16;not null;default:'normal'" json:"nick_font_weight"`
	NickFontColor         string     `gorm:"column:nick_font_color;size:32;not null;default:'#303133'" json:"nick_font_color"`
	NickFontSizePlus      int        `gorm:"column:nick_font_size_plus;not null;default:0" json:"nick_font_size_plus"`
	OfficialTagText       string     `gorm:"column:official_tag_text;size:32;not null;default:''" json:"official_tag_text"`
	OfficialTagBG         string     `gorm:"column:official_tag_bg;size:32;not null;default:''" json:"official_tag_bg"`
	NoticePopupEnable     int8       `gorm:"column:notice_popup_enable;not null;default:0" json:"notice_popup_enable"`
	KeyMsgLeftBarColor    string     `gorm:"column:key_msg_left_bar_color;size:32;not null;default:''" json:"key_msg_left_bar_color"`
	SessionBadge          string     `gorm:"column:session_badge;size:32;not null;default:''" json:"session_badge"`
	SessionBadgeColor     string     `gorm:"column:session_badge_color;size:32;not null;default:''" json:"session_badge_color"`
	SessionHornEnable     int8       `gorm:"column:session_horn_enable;not null;default:0" json:"session_horn_enable"`
	GroupTopNoticeEnable  int8       `gorm:"column:group_top_notice_enable;not null;default:0" json:"group_top_notice_enable"`
	GroupTopNoticeText    string     `gorm:"column:group_top_notice_text;size:255;not null;default:''" json:"group_top_notice_text"`
	NickVBadgeColor       string     `gorm:"column:nick_v_badge_color;size:32;not null;default:''" json:"nick_v_badge_color"`
	CreatedAt             *time.Time `gorm:"column:created_at" json:"created_at"`
	UpdatedAt             *time.Time `gorm:"column:updated_at" json:"updated_at"`
}

func (ImBubbleStyle) TableName() string { return "im_bubble_styles" }

// -------- 三级头像框样式配置 --------

// ImAvatarFrameStyle 头像框
type ImAvatarFrameStyle struct {
	ID                int64          `gorm:"primaryKey;autoIncrement;column:id" json:"id"`
	FrameKey          string         `gorm:"column:frame_key;uniqueIndex:uk_frame_key;size:32;not null;default:''" json:"frame_key"`
	FrameThickness    int            `gorm:"column:frame_thickness;not null;default:4" json:"frame_thickness"`
	FrameColor        string         `gorm:"column:frame_color;size:32;not null;default:''" json:"frame_color"`
	CornerDecorEnable int8           `gorm:"column:corner_decor_enable;not null;default:0" json:"corner_decor_enable"`
	CornerDecorColor  string         `gorm:"column:corner_decor_color;size:32;not null;default:''" json:"corner_decor_color"`
	AnimationEnable   int8           `gorm:"column:animation_enable;not null;default:0" json:"animation_enable"`
	AnimationType     string         `gorm:"column:animation_type;size:32;not null;default:'none'" json:"animation_type"`
	AnimationParams   datatypes.JSON `gorm:"column:animation_params" json:"animation_params"`
	GlowEnable        int8           `gorm:"column:glow_enable;not null;default:0" json:"glow_enable"`
	LayerCount        int            `gorm:"column:layer_count;not null;default:1" json:"layer_count"`
	OnlySelfClub      int8           `gorm:"column:only_self_club;not null;default:0" json:"only_self_club"`
	RequireVBadge     int8           `gorm:"column:require_v_badge;not null;default:0" json:"require_v_badge"`
	Priority          int            `gorm:"column:priority;not null;default:0" json:"priority"`
	CreatedAt         *time.Time     `gorm:"column:created_at" json:"created_at"`
	UpdatedAt         *time.Time     `gorm:"column:updated_at" json:"updated_at"`
}

func (ImAvatarFrameStyle) TableName() string { return "im_avatar_frame_styles" }

// -------- 账号样式授权映射 --------

// ImStyleGrant 样式授权
type ImStyleGrant struct {
	ID            int64          `gorm:"primaryKey;autoIncrement;column:id" json:"id"`
	UserID        int64          `gorm:"column:user_id;index:idx_user_id;not null;default:0" json:"user_id"`
	BubbleRoleKey string         `gorm:"column:bubble_role_key;size:32;not null;default:''" json:"bubble_role_key"`
	AvatarFrameKey string        `gorm:"column:avatar_frame_key;size:32;not null;default:''" json:"avatar_frame_key"`
	ScopeClubIDs  datatypes.JSON `gorm:"column:scope_club_ids" json:"scope_club_ids"`
	GrantedBy     int64          `gorm:"column:granted_by;not null;default:0" json:"granted_by"`
	GrantedAt     *time.Time     `gorm:"column:granted_at" json:"granted_at"`
	ExpiresAt     *time.Time     `gorm:"column:expires_at" json:"expires_at"`
}

func (ImStyleGrant) TableName() string { return "im_style_grants" }
