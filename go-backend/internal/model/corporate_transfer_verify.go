package model

import (
	"strconv"
	"time"
)

// 对公打款验证状态常量
const (
	CorporateTransferStatusPending  = "pending"  // 待验证
	CorporateTransferStatusVerified = "verified" // 已验证
	CorporateTransferStatusFailed   = "failed"   // 验证失败
	CorporateTransferStatusExpired  = "expired"  // 已过期
)

// CorporateTransferVerify 对公小额打款验证表
type CorporateTransferVerify struct {
	ID           int64      `gorm:"primaryKey;autoIncrement;column:id" json:"id"`
	ClubID       int64      `gorm:"column:club_id;index:idx_club_id;not null;default:0" json:"club_id"` // 俱乐部ID
	BankName     string     `gorm:"column:bank_name;size:64;not null;default:''" json:"bank_name"`     // 开户行
	BankAccount  string     `gorm:"column:bank_account;size:64;not null;default:''" json:"bank_account"` // 银行账号
	AccountName  string     `gorm:"column:account_name;size:64;not null;default:''" json:"account_name"` // 账户名
	VerifyAmount string     `gorm:"column:verify_amount;type:decimal(4,1);not null;default:0.0" json:"verify_amount"` // 验证金额(0.0-0.9，1位小数)
	GeneratedAt  *time.Time `gorm:"column:generated_at" json:"generated_at"`                             // 生成打款时间
	ExpireAt     *time.Time `gorm:"column:expire_at;index:idx_expire_at" json:"expire_at"`               // 过期时间(generated_at + 48h)
	VerifyCount  int        `gorm:"column:verify_count;not null;default:0" json:"verify_count"`          // 已用验证次数
	Status       string     `gorm:"column:status;size:32;index:idx_status;not null;default:'pending'" json:"status"` // 状态 pending/verified/failed/expired
	LockedUntil  *time.Time `gorm:"column:locked_until;index:idx_locked_until" json:"locked_until"`      // 锁定时间(失败次数达上限后 15 天内禁止提交)
	CreatedAt    *time.Time `gorm:"column:created_at" json:"created_at"`
	UpdatedAt    *time.Time `gorm:"column:updated_at" json:"updated_at"`
}

// TableName 指定表名
func (CorporateTransferVerify) TableName() string {
	return "corporate_transfer_verifies"
}

// GetVerifyAmountFloat 返回验证金额的浮点表示
func (c *CorporateTransferVerify) GetVerifyAmountFloat() float64 {
	f, _ := strconv.ParseFloat(c.VerifyAmount, 64)
	return f
}
