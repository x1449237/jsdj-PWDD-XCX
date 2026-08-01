package model

import "time"

// CreditLog 信用分变更流水表
type CreditLog struct {
	ID        int64      `gorm:"primaryKey;autoIncrement;column:id" json:"id"`
	UID       int64      `gorm:"column:uid;index:idx_uid;not null;default:0" json:"uid"`
	Delta     int        `gorm:"column:delta;not null;default:0" json:"delta"`
	After     int        `gorm:"column:after;not null;default:0" json:"after"`
	Reason    string     `gorm:"column:reason;size:255;not null;default:''" json:"reason"`
	RefID     int64      `gorm:"column:ref_id;index:idx_ref_id;not null;default:0" json:"ref_id"`
	CreatedAt *time.Time `gorm:"column:created_at;index:idx_created_at" json:"created_at"`
}

// TableName 指定表名
func (CreditLog) TableName() string {
	return "credit_logs"
}
