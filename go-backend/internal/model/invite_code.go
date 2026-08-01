package model

import (
	"encoding/json"
	"time"
)

// InviteCode 邀请码表模型
type InviteCode struct {
	ID          int64           `gorm:"primaryKey;autoIncrement;column:id" json:"id"`
	Code        string          `gorm:"column:code;uniqueIndex:uk_code;index:idx_code;size:64;not null;default:''" json:"code"` // 邀请码
	Type        string          `gorm:"column:type;index:idx_type;size:32;not null;default:'platform'" json:"type"` // 类型 club/platform
	ClubID      int64           `gorm:"column:club_id;index:idx_club_id;not null;default:0" json:"club_id"`        // 指定俱乐部ID
	Role        string          `gorm:"column:role;size:32;not null;default:''" json:"role"`                       // 角色 DS/FXS/空
	MaxUses     int             `gorm:"column:max_uses;not null;default:1" json:"max_uses"`                        // 最大使用次数
	UsedCount   int             `gorm:"column:used_count;not null;default:0" json:"used_count"`                    // 已使用次数
	ExpireAt    *time.Time      `gorm:"column:expire_at" json:"expire_at"`                                         // 过期时间
	Benefits    json.RawMessage `gorm:"column:benefits;type:json" json:"benefits"`                                 // 福利配置
	Status      string          `gorm:"column:status;index:idx_status;size:32;not null;default:'unused'" json:"status"` // 状态 unused/used/exhausted/expired/revoked
	CreatorID   int64           `gorm:"column:creator_id;not null;default:0" json:"creator_id"`                    // 创建人ID
	CreatorType string          `gorm:"column:creator_type;size:32;not null;default:''" json:"creator_type"`       // 创建人类型 admin/club
	UsedBy      int64           `gorm:"column:used_by;not null;default:0" json:"used_by"`                          // 使用人ID
	UsedAt      *time.Time      `gorm:"column:used_at" json:"used_at"`                                             // 使用时间
	CreatedAt   *time.Time      `gorm:"column:created_at;index:idx_created_at" json:"created_at"`
	UpdatedAt   *time.Time      `gorm:"column:updated_at" json:"updated_at"`
}

// TableName 指定表名
func (InviteCode) TableName() string {
	return "invite_codes"
}

// 邀请码类型常量
const (
	InviteCodeTypeClub    = "club"    // 俱乐部邀请码
	InviteCodeTypePlatform = "platform" // 平台通用码(QPT_ 前缀)
)

// 邀请码角色标识常量
const (
	InviteCodeRoleDS  = "DS"  // 打手
	InviteCodeRoleFXS = "FXS" // 分销商
)

// 邀请码状态常量
const (
	InviteCodeStatusUnused     = "unused"     // 未使用
	InviteCodeStatusUsed       = "used"       // 已使用
	InviteCodeStatusExhausted  = "exhausted"  // 已用尽
	InviteCodeStatusExpired    = "expired"    // 已过期
	InviteCodeStatusRevoked    = "revoked"    // 已撤销
)
