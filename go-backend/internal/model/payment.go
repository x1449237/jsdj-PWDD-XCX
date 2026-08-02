package model

import (
	"encoding/json"
	"time"
)

// Payment 支付记录表模型
type Payment struct {
	ID            int64      `gorm:"primaryKey;autoIncrement;column:id" json:"id"`
	OrderID       int64      `gorm:"column:order_id;index:idx_order_id;not null;default:0" json:"order_id"`                            // 订单ID
	OutTradeNo    string     `gorm:"column:out_trade_no;uniqueIndex:uk_out_trade_no;size:64;not null;default:''" json:"out_trade_no"`  // 商户订单号
	TransactionID string     `gorm:"column:transaction_id;index:idx_transaction_id;size:64;not null;default:''" json:"transaction_id"` // 第三方交易号
	Amount        int64      `gorm:"column:amount;not null;default:0" json:"amount"`                                                   // 支付金额(分)
	PayMethod     string     `gorm:"column:pay_method;size:32;not null;default:''" json:"pay_method"`                                  // 支付方式 wechat/ios
	PayTime       *time.Time `gorm:"column:pay_time" json:"pay_time"`                                                                  // 支付时间
	RefundAmount  int64      `gorm:"column:refund_amount;not null;default:0" json:"refund_amount"`                                     // 退款金额(分)
	RefundTime    *time.Time `gorm:"column:refund_time" json:"refund_time"`                                                            // 退款时间
	Status        string     `gorm:"column:status;index:idx_status;size:32;not null;default:'pending'" json:"status"`                  // 状态 pending/paid/refunded/partial_refund
	CreatedAt     *time.Time `gorm:"column:created_at;index:idx_created_at" json:"created_at"`
	UpdatedAt     *time.Time `gorm:"column:updated_at" json:"updated_at"`
}

// TableName 指定表名
func (Payment) TableName() string {
	return "payments"
}

// 支付状态常量
const (
	PaymentStatusPending    = "pending"
	PaymentStatusPaid       = "paid"
	PaymentStatusRefunded   = "refunded"
	PaymentStatusPartialRef = "partial_refund"
)

// Withdraw 提现记录表
type Withdraw struct {
	ID            int64      `gorm:"primaryKey;autoIncrement;column:id" json:"id"`
	UserID        int64      `gorm:"column:user_id;index:idx_user_id;not null;default:0" json:"user_id"`              // 用户ID
	Amount        int64      `gorm:"column:amount;not null;default:0" json:"amount"`                                  // 提现金额(分)
	Fee           int64      `gorm:"column:fee;not null;default:0" json:"fee"`                                        // 手续费(分)
	Tax           int64      `gorm:"column:tax;not null;default:0" json:"tax"`                                        // 个税(分)
	NetAmount     int64      `gorm:"column:net_amount;not null;default:0" json:"net_amount"`                          // 到账金额(分)
	BankCard      string     `gorm:"column:bank_card;size:32;not null;default:''" json:"bank_card"`                   // 银行卡号
	BankName      string     `gorm:"column:bank_name;size:64;not null;default:''" json:"bank_name"`                   // 开户行
	BankPhone     string     `gorm:"column:bank_phone;size:20;not null;default:''" json:"bank_phone"`                 // 银行预留手机号
	IDCard        string     `gorm:"column:id_card;size:18;not null;default:''" json:"id_card"`                       // 身份证号
	RealName      string     `gorm:"column:real_name;size:32;not null;default:''" json:"real_name"`                   // 真实姓名
	Channel       string     `gorm:"column:channel;size:32;not null;default:''" json:"channel"`                       // 提现渠道 wechat/bank
	Status        string     `gorm:"column:status;index:idx_status;size:32;not null;default:'pending'" json:"status"` // 状态 pending/approved/rejected/paid
	FailReason    string     `gorm:"column:fail_reason;size:255;not null;default:''" json:"fail_reason"`              // 失败原因
	ReviewerID    int64      `gorm:"column:reviewer_id;not null;default:0" json:"reviewer_id"`                        // 审核人ID
	ReviewedAt    *time.Time `gorm:"column:reviewed_at" json:"reviewed_at"`                                           // 审核时间
	PaidAt        *time.Time `gorm:"column:paid_at" json:"paid_at"`                                                   // 打款时间
	RetryCount    int        `gorm:"column:retry_count;not null;default:0" json:"retry_count"`                        // 重试次数
	MergeSettleID int64      `gorm:"column:merge_settle_id;not null;default:0" json:"merge_settle_id"`                // 合并结算ID
	CreatedAt     *time.Time `gorm:"column:created_at;index:idx_created_at" json:"created_at"`
	UpdatedAt     *time.Time `gorm:"column:updated_at" json:"updated_at"`
}

// TableName 指定表名
func (Withdraw) TableName() string {
	return "withdrawals"
}

// 提现状态常量
const (
	WithdrawStatusPending  = "pending"
	WithdrawStatusApproved = "approved"
	WithdrawStatusRejected = "rejected"
	WithdrawStatusPaid     = "paid"
)

// ProfitShareRule 分账规则表
type ProfitShareRule struct {
	ID               int64           `gorm:"primaryKey;autoIncrement;column:id" json:"id"`
	Name             string          `gorm:"column:name;size:64;not null;default:''" json:"name"`                                       // 规则名称
	PlayerRatio      float64         `gorm:"column:player_ratio;type:decimal(5,2);not null;default:0.00" json:"player_ratio"`           // 打手分账比例%
	ClubRatio        float64         `gorm:"column:club_ratio;type:decimal(5,2);not null;default:0.00" json:"club_ratio"`               // 俱乐部分账比例%
	DistributorRatio float64         `gorm:"column:distributor_ratio;type:decimal(5,2);not null;default:0.00" json:"distributor_ratio"` // 分销商分账比例%
	PlatformRatio    float64         `gorm:"column:platform_ratio;type:decimal(5,2);not null;default:0.00" json:"platform_ratio"`       // 平台分账比例%
	Conditions       json.RawMessage `gorm:"column:conditions;type:json" json:"conditions"`                                             // 适用条件
	Status           int8            `gorm:"column:status;index:idx_status;not null;default:1" json:"status"`                           // 状态 1启用 0停用
	CreatedAt        *time.Time      `gorm:"column:created_at" json:"created_at"`
	UpdatedAt        *time.Time      `gorm:"column:updated_at" json:"updated_at"`
}

// TableName 指定表名
func (ProfitShareRule) TableName() string {
	return "profit_share_rules"
}

// ProfitShareRecord 分账记录表
type ProfitShareRecord struct {
	ID                int64      `gorm:"primaryKey;autoIncrement;column:id" json:"id"`
	OrderID           int64      `gorm:"column:order_id;index:idx_order_id;not null;default:0" json:"order_id"`  // 订单ID
	UserID            int64      `gorm:"column:user_id;index:idx_user_id;not null;default:0" json:"user_id"`     // 收款方用户ID
	Role              string     `gorm:"column:role;size:32;not null;default:''" json:"role"`                    // 角色 player/club/distributor/platform
	Amount            int64      `gorm:"column:amount;not null;default:0" json:"amount"`                         // 分账金额(分)
	Ratio             float64    `gorm:"column:ratio;type:decimal(5,2);not null;default:0.00" json:"ratio"`      // 分账比例%
	Status            int8       `gorm:"column:status;index:idx_status;not null;default:0" json:"status"`        // 状态 0待分账 1已分账 2已回滚
	IsFrozen          int8       `gorm:"column:is_frozen;not null;default:0" json:"is_frozen"`                   // 是否冻结
	FrozenReason      string     `gorm:"column:frozen_reason;size:255;not null;default:''" json:"frozen_reason"` // 冻结原因
	UnfrozenAt        *time.Time `gorm:"column:unfrozen_at" json:"unfrozen_at"`                                  // 解冻时间
	EstimatedSettleAt *time.Time `gorm:"column:estimated_settle_at" json:"estimated_settle_at"`                  // 预计结算时间
	CreatedAt         *time.Time `gorm:"column:created_at;index:idx_created_at" json:"created_at"`
	UpdatedAt         *time.Time `gorm:"column:updated_at" json:"updated_at"`
}

// TableName 指定表名
func (ProfitShareRecord) TableName() string {
	return "profit_share_records"
}

// 分账角色常量
const (
	ProfitShareRolePlayer      = "player"
	ProfitShareRoleClub        = "club"
	ProfitShareRoleDistributor = "distributor"
	ProfitShareRolePlatform    = "platform"
)
