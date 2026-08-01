package model

import "time"

// AnnouncementReadLog 公告已读日志表
type AnnouncementReadLog struct {
	ID             int64      `gorm:"primaryKey;autoIncrement;column:id" json:"id"`
	AnnouncementID int64      `gorm:"column:announcement_id;index:idx_announcement_id;not null;default:0" json:"announcement_id"` // 公告ID(对应 group_chats.id)
	UserID         int64      `gorm:"column:user_id;index:idx_user_id;not null;default:0" json:"user_id"`        // 用户ID
	ReadAt         *time.Time `gorm:"column:read_at" json:"read_at"`                                            // 阅读时间
	CreatedAt      *time.Time `gorm:"column:created_at;index:idx_created_at" json:"created_at"`
}

// TableName 指定表名
func (AnnouncementReadLog) TableName() string {
	return "announcement_read_logs"
}
