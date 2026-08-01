package model

import "time"

// UserVBadge 用户V标表模型
type UserVBadge struct {
	ID        int64      `gorm:"primaryKey;autoIncrement;column:id" json:"id"`
	UserID    int64      `gorm:"column:user_id;index:idx_user_id;not null;default:0" json:"user_id"`       // 用户ID
	ClubID    int64      `gorm:"column:club_id;index:idx_club_id;not null;default:0" json:"club_id"`       // 俱乐部ID
	BadgeType string     `gorm:"column:badge_type;index:idx_badge_type;size:32;not null;default:'none'" json:"badge_type"` // V标类型 none/blue/green/gold
	Status    int8       `gorm:"column:status;not null;default:1" json:"status"`                          // 状态 1有效 0失效
	GrantedAt *time.Time `gorm:"column:granted_at" json:"granted_at"`                                     // 授予时间
	CreatedAt *time.Time `gorm:"column:created_at" json:"created_at"`
	UpdatedAt *time.Time `gorm:"column:updated_at" json:"updated_at"`
}

// TableName 指定表名
func (UserVBadge) TableName() string {
	return "user_v_badges"
}

// V标类型常量
const (
	VBadgeTypeNone  = "none"  // 无
	VBadgeTypeBlue  = "blue"  // 蓝V(企业)
	VBadgeTypeGreen = "green" // 绿V(个人)
	VBadgeTypeGold  = "gold"  // 金V(平台)
)
