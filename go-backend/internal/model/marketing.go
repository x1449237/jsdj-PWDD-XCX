package model

import (
	"time"

	"gorm.io/datatypes"
)

// CouponTemplate 优惠券模板表模型
type CouponTemplate struct {
	ID            int64      `gorm:"primaryKey;autoIncrement;column:id" json:"id"`
	Name          string     `gorm:"column:name;size:64;not null;default:''" json:"name"`                                 // 券名称
	Type          string     `gorm:"column:type;index:idx_type;size:32;not null;default:''" json:"type"`                  // 类型 newuser/fullcut/discount/recharge/compensation
	Amount        int64      `gorm:"column:amount;not null;default:0" json:"amount"`                                      // 优惠金额(分)
	MinSpend      int64      `gorm:"column:min_spend;not null;default:0" json:"min_spend"`                                // 最低消费(分)
	DiscountRatio float64    `gorm:"column:discount_ratio;type:decimal(3,2);not null;default:0.00" json:"discount_ratio"` // 折扣比例
	ValidDays     int        `gorm:"column:valid_days;not null;default:0" json:"valid_days"`                              // 有效天数
	TotalCount    int        `gorm:"column:total_count;not null;default:0" json:"total_count"`                            // 发放总量
	IssuedCount   int        `gorm:"column:issued_count;not null;default:0" json:"issued_count"`                          // 已发放数
	Status        int8       `gorm:"column:status;index:idx_status;not null;default:1" json:"status"`                     // 状态 1启用 0停用
	CreatedAt     *time.Time `gorm:"column:created_at;index:idx_created_at" json:"created_at"`
	UpdatedAt     *time.Time `gorm:"column:updated_at" json:"updated_at"`
}

// TableName 指定表名
func (CouponTemplate) TableName() string {
	return "coupon_templates"
}

// 优惠券类型常量
const (
	CouponTypeNewuser      = "newuser"      // 新人券
	CouponTypeFullCut      = "fullcut"      // 满减券
	CouponTypeDiscount     = "discount"     // 折扣券
	CouponTypeRecharge     = "recharge"     // 充值券
	CouponTypeCompensation = "compensation" // 补偿券
)

// UserCoupon 用户优惠券表
type UserCoupon struct {
	ID         int64      `gorm:"primaryKey;autoIncrement;column:id" json:"id"`
	UserID     int64      `gorm:"column:user_id;index:idx_user_id;not null;default:0" json:"user_id"`             // 用户ID
	TemplateID int64      `gorm:"column:template_id;index:idx_template_id;not null;default:0" json:"template_id"` // 模板ID
	Status     string     `gorm:"column:status;index:idx_status;size:32;not null;default:'unused'" json:"status"` // 状态 unused/used/expired
	UsedAt     *time.Time `gorm:"column:used_at" json:"used_at"`                                                  // 使用时间
	ExpireAt   *time.Time `gorm:"column:expire_at" json:"expire_at"`                                              // 过期时间
	CreatedAt  *time.Time `gorm:"column:created_at;index:idx_created_at" json:"created_at"`
	UpdatedAt  *time.Time `gorm:"column:updated_at" json:"updated_at"`
}

// TableName 指定表名
func (UserCoupon) TableName() string {
	return "user_coupons"
}

// 用户优惠券状态常量
const (
	UserCouponStatusUnused  = "unused"
	UserCouponStatusUsed    = "used"
	UserCouponStatusExpired = "expired"
)

// RechargeActivity 充值活动表
type RechargeActivity struct {
	ID              int64      `gorm:"primaryKey;autoIncrement;column:id" json:"id"`
	Name            string     `gorm:"column:name;size:64;not null;default:''" json:"name"`                // 活动名称
	ThresholdAmount int64      `gorm:"column:threshold_amount;not null;default:0" json:"threshold_amount"` // 充值门槛(分)
	BonusAmount     int64      `gorm:"column:bonus_amount;not null;default:0" json:"bonus_amount"`         // 赠送金额(分)
	Status          int8       `gorm:"column:status;index:idx_status;not null;default:1" json:"status"`    // 状态 1启用 0停用
	StartAt         *time.Time `gorm:"column:start_at" json:"start_at"`                                    // 开始时间
	EndAt           *time.Time `gorm:"column:end_at" json:"end_at"`                                        // 结束时间
	CreatedAt       *time.Time `gorm:"column:created_at;index:idx_created_at" json:"created_at"`
	UpdatedAt       *time.Time `gorm:"column:updated_at" json:"updated_at"`
}

// TableName 指定表名
func (RechargeActivity) TableName() string {
	return "recharge_activities"
}

// LotteryActivity 抽奖活动表
type LotteryActivity struct {
	ID        int64      `gorm:"primaryKey;autoIncrement;column:id" json:"id"`
	Name      string     `gorm:"column:name;size:64;not null;default:''" json:"name"`             // 活动名称
	Status    int8       `gorm:"column:status;index:idx_status;not null;default:1" json:"status"` // 状态 1启用 0停用
	StartAt   *time.Time `gorm:"column:start_at" json:"start_at"`                                 // 开始时间
	EndAt     *time.Time `gorm:"column:end_at" json:"end_at"`                                     // 结束时间
	CreatedAt *time.Time `gorm:"column:created_at;index:idx_created_at" json:"created_at"`
	UpdatedAt *time.Time `gorm:"column:updated_at" json:"updated_at"`
}

// TableName 指定表名
func (LotteryActivity) TableName() string {
	return "lottery_activities"
}

// ClubCoupon 俱乐部优惠券表
type ClubCoupon struct {
	ID                 int64          `gorm:"primaryKey;autoIncrement;column:id" json:"id"`
	Status             int8           `gorm:"column:status;index:idx_status;not null;default:1" json:"status"`    // 状态 1启用 0停用
	IsInternalOnly     int8           `gorm:"column:is_internal_only;not null;default:0" json:"is_internal_only"` // 仅内部可见
	IsPaused           int8           `gorm:"column:is_paused;not null;default:0" json:"is_paused"`               // 暂停发放
	ApplicableGames    datatypes.JSON `gorm:"column:applicable_games" json:"applicable_games"`                    // 适用游戏
	ApplicableServices datatypes.JSON `gorm:"column:applicable_services" json:"applicable_services"`              // 适用服务
	CreatedAt          *time.Time     `gorm:"column:created_at;index:idx_created_at" json:"created_at"`
	UpdatedAt          *time.Time     `gorm:"column:updated_at" json:"updated_at"`
}

// TableName 指定表名
func (ClubCoupon) TableName() string {
	return "club_coupons"
}
