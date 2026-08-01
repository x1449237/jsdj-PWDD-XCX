package model

import "time"

// FileBlockchainRecord 文件上链存证记录表
type FileBlockchainRecord struct {
	ID              int64      `gorm:"primaryKey;autoIncrement;column:id" json:"id"`
	FileHash        string     `gorm:"column:file_hash;index:idx_file_hash;size:64;not null;default:''" json:"file_hash"` // 文件 SHA-256 哈希
	FileType        string     `gorm:"column:file_type;size:32;not null;default:''" json:"file_type"`                    // 文件类型
	RefType         string     `gorm:"column:ref_type;index:idx_ref_type;size:32;not null;default:''" json:"ref_type"`   // 关联类型 club_join/club_archive
	RefID           int64      `gorm:"column:ref_id;index:idx_ref_id;not null;default:0" json:"ref_id"`                  // 关联ID
	OSSUrl          string     `gorm:"column:oss_url;size:512;not null;default:''" json:"oss_url"`                      // OSS 存储地址
	WatermarkText   string     `gorm:"column:watermark_text;size:128;not null;default:''" json:"watermark_text"`        // 水印文本
	BlockchainTxID  string     `gorm:"column:blockchain_tx_id;index:idx_blockchain_tx_id;size:128;not null;default:''" json:"blockchain_tx_id"` // 上链交易ID
	CreatedAt       *time.Time `gorm:"column:created_at;index:idx_created_at" json:"created_at"`
}

// TableName 指定表名
func (FileBlockchainRecord) TableName() string {
	return "file_blockchain_records"
}
