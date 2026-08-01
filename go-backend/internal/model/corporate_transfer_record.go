package model

import "time"

// CorporateTransferRecord 企业对公小额打款记录表
type CorporateTransferRecord struct {
	ID          int64      `gorm:"primaryKey;autoIncrement;column:id" json:"id"`
	ClubID      int64      `gorm:"column:club_id;index:idx_club_id;not null;default:0" json:"club_id"`        // 俱乐部ID
	VerifyID    int64      `gorm:"column:verify_id;index:idx_verify_id;not null;default:0" json:"verify_id"`  // 验证流程ID
	Amount      int64      `gorm:"column:amount;not null;default:0" json:"amount"`                            // 打款金额(分)
	Direction   string     `gorm:"column:direction;size:16;not null;default:'out'" json:"direction"`         // 方向 out/refund
	BankName    string     `gorm:"column:bank_name;size:64;not null;default:''" json:"bank_name"`            // 开户行
	BankAccount string     `gorm:"column:bank_account;size:64;not null;default:''" json:"bank_account"`     // 银行账号
	TransferAt  *time.Time `gorm:"column:transfer_at" json:"transfer_at"`                                    // 打款时间
	Status      string     `gorm:"column:status;index:idx_status;size:32;not null;default:'pending'" json:"status"` // 状态 pending/success/failed
	CreatedAt   *time.Time `gorm:"column:created_at;index:idx_created_at" json:"created_at"`
}

// TableName 指定表名
func (CorporateTransferRecord) TableName() string {
	return "corporate_transfer_records"
}

// 打款方向常量
const (
	CorporateTransferDirectionOut    = "out"    // 打出
	CorporateTransferDirectionRefund = "refund" // 退回
)
