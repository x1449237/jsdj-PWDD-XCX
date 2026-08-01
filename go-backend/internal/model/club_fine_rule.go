package model

import "time"

// ClubFineRule 俱乐部内部罚款规则表
type ClubFineRule struct {
	ID          int64      `gorm:"primaryKey;autoIncrement;column:id" json:"id"`
	ClubID      int64      `gorm:"column:club_id;index:idx_club_id;not null;default:0" json:"club_id"`        // 俱乐部ID
	Name        string     `gorm:"column:name;size:64;not null;default:''" json:"name"`                     // 规则名称
	Description string     `gorm:"column:description;type:text" json:"description"`                          // 规则描述
	Amount      int64      `gorm:"column:amount;not null;default:0" json:"amount"`                            // 罚款金额(分)
	Status      string     `gorm:"column:status;index:idx_status;size:32;not null;default:'active'" json:"status"` // 状态 active/revoked
	HasUnpaid   int8       `gorm:"column:has_unpaid;not null;default:0" json:"has_unpaid"`                    // 是否存在未赔付罚款 0否 1是
	CreatedBy   int64      `gorm:"column:created_by;not null;default:0" json:"created_by"`                    // 创建人ID
	CreatedAt   *time.Time `gorm:"column:created_at;index:idx_created_at" json:"created_at"`
	UpdatedAt   *time.Time `gorm:"column:updated_at" json:"updated_at"`
}

// TableName 指定表名
func (ClubFineRule) TableName() string {
	return "club_fine_rules"
}

// 罚款规则状态常量
const (
	ClubFineRuleStatusActive  = "active"  // 生效
	ClubFineRuleStatusRevoked = "revoked" // 已下架
)
