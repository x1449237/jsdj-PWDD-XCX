package model

import "time"

// ClubInfoChangeLog 俱乐部资料修改日志表
type ClubInfoChangeLog struct {
	ID         int64      `gorm:"primaryKey;autoIncrement;column:id" json:"id"`
	ClubID     int64      `gorm:"column:club_id;index:idx_club_id;not null;default:0" json:"club_id"`        // 俱乐部ID
	Field      string     `gorm:"column:field;size:64;not null;default:''" json:"field"`                    // 修改字段名
	OldValue   string     `gorm:"column:old_value;type:text" json:"old_value"`                                // 旧值
	NewValue   string     `gorm:"column:new_value;type:text" json:"new_value"`                                // 新值
	OperatorID int64      `gorm:"column:operator_id;not null;default:0" json:"operator_id"`                  // 操作人ID
	CreatedAt  *time.Time `gorm:"column:created_at;index:idx_created_at" json:"created_at"`
}

// TableName 指定表名
func (ClubInfoChangeLog) TableName() string {
	return "club_info_change_logs"
}
