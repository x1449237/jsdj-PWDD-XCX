package model

import (
	"encoding/json"
	"time"
)

// ClubJoinDraft 俱乐部入驻草稿表
type ClubJoinDraft struct {
	ID        int64           `gorm:"primaryKey;autoIncrement;column:id" json:"id"`
	UserID    int64           `gorm:"column:user_id;index:idx_user_id;not null;default:0" json:"user_id"` // 用户ID
	DraftData json.RawMessage `gorm:"column:draft_data;type:json" json:"draft_data"`                       // 草稿数据(JSON)
	ExpireAt  *time.Time      `gorm:"column:expire_at;index:idx_expire_at" json:"expire_at"`               // 过期时间
	CreatedAt *time.Time      `gorm:"column:created_at" json:"created_at"`
	UpdatedAt *time.Time      `gorm:"column:updated_at" json:"updated_at"`
}

// TableName 指定表名
func (ClubJoinDraft) TableName() string {
	return "club_join_drafts"
}
