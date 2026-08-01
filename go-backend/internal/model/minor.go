package model

import "time"

// MinorCurfewLog 未成年宵禁日志表模型
type MinorCurfewLog struct {
	ID        int64      `gorm:"primaryKey;autoIncrement;column:id" json:"id"`
	UserID    int64      `gorm:"column:user_id;index:idx_user_id;not null;default:0" json:"user_id"`     // 用户ID
	Action    string     `gorm:"column:action;size:32;not null;default:''" json:"action"`                // 尝试操作 order/pay/reward
	BlockedAt *time.Time `gorm:"column:blocked_at" json:"blocked_at"`                                    // 拦截时间
	CreatedAt *time.Time `gorm:"column:created_at;index:idx_created_at" json:"created_at"`
}

// TableName 指定表名
func (MinorCurfewLog) TableName() string {
	return "minor_curfew_logs"
}

// 宵禁拦截操作类型常量
const (
	MinorActionOrder  = "order"  // 下单
	MinorActionPay    = "pay"    // 支付
	MinorActionReward = "reward" // 打赏
)

// ParentGuardianBind 家长绑定表
type ParentGuardianBind struct {
	ID         int64      `gorm:"primaryKey;autoIncrement;column:id" json:"id"`
	ParentUID  int64      `gorm:"column:parent_uid;index:idx_parent_uid;not null;default:0" json:"parent_uid"` // 家长用户ID
	ChildUID   int64      `gorm:"column:child_uid;index:idx_child_uid;not null;default:0" json:"child_uid"`   // 未成年用户ID
	VerifiedAt *time.Time `gorm:"column:verified_at" json:"verified_at"`                                  // 验证时间
	CreatedAt  *time.Time `gorm:"column:created_at" json:"created_at"`
	UpdatedAt  *time.Time `gorm:"column:updated_at" json:"updated_at"`
}

// TableName 指定表名
func (ParentGuardianBind) TableName() string {
	return "parent_guardian_binds"
}

// ParentGuardianSetting 家长设置表
type ParentGuardianSetting struct {
	ID           int64      `gorm:"primaryKey;autoIncrement;column:id" json:"id"`
	ParentUID    int64      `gorm:"column:parent_uid;index:idx_parent_uid;not null;default:0" json:"parent_uid"` // 家长用户ID
	ChildUID     int64      `gorm:"column:child_uid;index:idx_child_uid;not null;default:0" json:"child_uid"`   // 未成年用户ID
	MonthlyLimit int64      `gorm:"column:monthly_limit;not null;default:0" json:"monthly_limit"`              // 月度消费限额(分)
	AllowOrder   int8       `gorm:"column:allow_order;not null;default:1" json:"allow_order"`                  // 允许下单 0否 1是
	AllowReward  int8       `gorm:"column:allow_reward;not null;default:1" json:"allow_reward"`                // 允许打赏 0否 1是
	IsFrozen     int8       `gorm:"column:is_frozen;not null;default:0" json:"is_frozen"`                      // 是否冻结账户 0否 1是
	CreatedAt    *time.Time `gorm:"column:created_at" json:"created_at"`
	UpdatedAt    *time.Time `gorm:"column:updated_at" json:"updated_at"`
}

// TableName 指定表名
func (ParentGuardianSetting) TableName() string {
	return "parent_guardian_settings"
}
