package model

import (
	"encoding/json"
	"time"
)

// ClubArchive 俱乐部注销资料归档表
type ClubArchive struct {
	ID             int64           `gorm:"primaryKey;autoIncrement;column:id" json:"id"`
	ClubID         int64           `gorm:"column:club_id;index:idx_club_id;not null;default:0" json:"club_id"`        // 俱乐部ID
	ArchiveData    json.RawMessage `gorm:"column:archive_data;type:json" json:"archive_data"`                          // 归档资料 JSON
	Encrypted      bool            `gorm:"column:encrypted;not null;default:0" json:"encrypted"`                       // 是否已加密
	FileHash       string          `gorm:"column:file_hash;index:idx_file_hash;size:64;not null;default:''" json:"file_hash"` // 加密后哈希
	BlockchainTxID string          `gorm:"column:blockchain_tx_id;size:128;not null;default:''" json:"blockchain_tx_id"` // 上链交易ID
	ArchivedAt     *time.Time      `gorm:"column:archived_at" json:"archived_at"`                                     // 归档时间
	CreatedAt      *time.Time      `gorm:"column:created_at;index:idx_created_at" json:"created_at"`
}

// TableName 指定表名
func (ClubArchive) TableName() string {
	return "club_archives"
}
