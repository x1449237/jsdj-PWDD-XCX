package model

import "time"

// 保证金扣款类型常量
const (
	DepositDeductTypeFine         = "fine"         // 罚款
	DepositDeductTypeCompensation = "compensation" // 赔偿
)

// ClubDepositDeduction 俱乐部保证金扣款记录表
type ClubDepositDeduction struct {
	ID        int64      `gorm:"primaryKey;autoIncrement;column:id" json:"id"`
	ClubID    int64      `gorm:"column:club_id;index:idx_club_id;not null;default:0" json:"club_id"` // 俱乐部ID
	Amount    int64      `gorm:"column:amount;not null;default:0" json:"amount"`                     // 扣除金额(分)
	Type      string     `gorm:"column:type;size:32;index:idx_type;not null;default:'fine'" json:"type"` // 类型 fine/compensation
	Reason    string     `gorm:"column:reason;size:255;not null;default:''" json:"reason"`           // 扣款原因
	OperatorID int64     `gorm:"column:operator_id;not null;default:0" json:"operator_id"`            // 操作人ID
	CreatedAt *time.Time `gorm:"column:created_at;index:idx_created_at" json:"created_at"`
}

// TableName 指定表名
func (ClubDepositDeduction) TableName() string {
	return "club_deposit_deductions"
}
