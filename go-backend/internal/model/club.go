package model

import "time"

// Club 俱乐部表模型
type Club struct {
	ID             int64      `gorm:"primaryKey;autoIncrement;column:id" json:"id"`
	Name           string     `gorm:"column:name;size:128;not null;default:''" json:"name"`                     // 俱乐部名称
	Abbreviation   string     `gorm:"column:abbreviation;uniqueIndex:uk_abbreviation;size:10;not null;default:''" json:"abbreviation"` // 缩写(唯一封存)
	Type           int8       `gorm:"column:type;not null;default:2" json:"type"`                               // 类型 1=企业 2=个人
	Status         int8       `gorm:"column:status;index:idx_status;not null;default:0" json:"status"`          // 状态 0=审核中 1=审核通过 2=驳回 3=冻结 4=注销
	FounderUID     int64      `gorm:"column:founder_uid;index:idx_founder_uid;not null;default:0" json:"founder_uid"` // 创始人用户ID
	VBadgeType     int8       `gorm:"column:v_badge_type;not null;default:0" json:"v_badge_type"`               // V标类型 0=无 1=蓝V(企业) 2=绿V(个人)
	DepositStatus  int8       `gorm:"column:deposit_status;not null;default:0" json:"deposit_status"`           // 保证金状态 0未缴 1已缴 2已退
	DepositAmount  int64      `gorm:"column:deposit_amount;not null;default:0" json:"deposit_amount"`           // 保证金金额(分)
	Description    string     `gorm:"column:description;type:text" json:"description"`                          // 俱乐部简介
	Logo           string     `gorm:"column:logo;size:512;not null;default:''" json:"logo"`                     // Logo URL
	Background     string     `gorm:"column:background;size:512;not null;default:''" json:"background"`         // 背景图URL
	ContactPhone   string     `gorm:"column:contact_phone;size:20;not null;default:''" json:"contact_phone"`   // 联系电话
	ContactWechat  string     `gorm:"column:contact_wechat;size:64;not null;default:''" json:"contact_wechat"`  // 联系微信
	ContactQQ      string     `gorm:"column:contact_qq;size:32;not null;default:''" json:"contact_qq"`          // 联系QQ
	BusinessHours  string     `gorm:"column:business_hours;size:64;not null;default:''" json:"business_hours"`  // 营业时间
	Rating         float64    `gorm:"column:rating;type:decimal(2,1);not null;default:5.0" json:"rating"`       // 评分
	TotalOrders    int        `gorm:"column:total_orders;not null;default:0" json:"total_orders"`               // 总订单数
	RejectCount    int        `gorm:"column:reject_count;not null;default:0" json:"reject_count"`               // 入驻驳回次数
	LockedUntil    *time.Time `gorm:"column:locked_until" json:"locked_until"`                                 // 入驻锁定截止时间(驳回3次后锁定7天)
	CommissionRate int8       `gorm:"column:commission_rate;not null;default:0" json:"commission_rate"`         // 创始人抽成比例(0-100)
	IsArchived     int8       `gorm:"column:is_archived;not null;default:0" json:"is_archived"`                 // 是否已注销归档 0否 1是
	CreatedAt      *time.Time `gorm:"column:created_at;index:idx_created_at" json:"created_at"`
	UpdatedAt      *time.Time `gorm:"column:updated_at" json:"updated_at"`
}

// TableName 指定表名
func (Club) TableName() string {
	return "clubs"
}

// 俱乐部状态常量
const (
	ClubStatusReviewing int8 = 0 // 审核中
	ClubStatusApproved  int8 = 1 // 审核通过
	ClubStatusRejected  int8 = 2 // 驳回
	ClubStatusFrozen    int8 = 3 // 冻结
	ClubStatusCanceled  int8 = 4 // 注销
)

// 俱乐部类型常量
const (
	ClubTypeEnterprise int8 = 1 // 企业
	ClubTypePersonal   int8 = 2 // 个人
)

// ClubMember 俱乐部成员表
type ClubMember struct {
	ID        int64      `gorm:"primaryKey;autoIncrement;column:id" json:"id"`
	ClubID    int64      `gorm:"column:club_id;index:idx_club_id;not null;default:0" json:"club_id"`    // 俱乐部ID
	UserID    int64      `gorm:"column:user_id;index:idx_user_id;not null;default:0" json:"user_id"`  // 用户ID
	Role      int8       `gorm:"column:role;not null;default:3" json:"role"`                          // 角色 1=创始人 2=管理员 3=打手
	JoinedAt  *time.Time `gorm:"column:joined_at" json:"joined_at"`                                  // 加入时间
	Status    int8       `gorm:"column:status;index:idx_status;not null;default:1" json:"status"`    // 状态 1正常 0已移除
	CreatedAt *time.Time `gorm:"column:created_at" json:"created_at"`
	UpdatedAt *time.Time `gorm:"column:updated_at" json:"updated_at"`
}

// TableName 指定表名
func (ClubMember) TableName() string {
	return "club_members"
}

// 俱乐部成员角色常量
const (
	ClubMemberRoleFounder int8 = 1 // 创始人
	ClubMemberRoleAdmin   int8 = 2 // 管理员
	ClubMemberRolePlayer  int8 = 3 // 打手
)

// ClubAbbreviation 俱乐部缩写封存表
type ClubAbbreviation struct {
	ID          int64      `gorm:"primaryKey;autoIncrement;column:id" json:"id"`
	Abbreviation string    `gorm:"column:abbreviation;uniqueIndex:uk_abbreviation;size:10;not null;default:''" json:"abbreviation"` // 缩写
	ClubID      int64      `gorm:"column:club_id;index:idx_club_id;not null;default:0" json:"club_id"`      // 原俱乐部ID
	AbandonedAt *time.Time `gorm:"column:abandoned_at" json:"abandoned_at"`                                // 封存时间
	CreatedAt   *time.Time `gorm:"column:created_at" json:"created_at"`
}

// TableName 指定表名
func (ClubAbbreviation) TableName() string {
	return "club_abbreviations"
}
