package model

import "time"

// DistributorRelation 分销关系表模型
type DistributorRelation struct {
	ID            int64      `gorm:"primaryKey;autoIncrement;column:id" json:"id"`
	SuperiorID    int64      `gorm:"column:superior_id;index:idx_superior_id;not null;default:0" json:"superior_id"` // 上级分销商ID
	SubordinateID int64      `gorm:"column:subordinate_id;index:idx_subordinate_id;not null;default:0" json:"subordinate_id"` // 下级用户ID
	Level         int8       `gorm:"column:level;index:idx_level;not null;default:1" json:"level"`             // 级别 1=一级 2=二级
	IsValid       int8       `gorm:"column:is_valid;not null;default:0" json:"is_valid"`                       // 是否有效下级 0否 1是
	CreatedAt     *time.Time `gorm:"column:created_at" json:"created_at"`
	UpdatedAt     *time.Time `gorm:"column:updated_at" json:"updated_at"`
}

// TableName 指定表名
func (DistributorRelation) TableName() string {
	return "distributor_relations"
}

// 分销级别常量
const (
	DistributorLevelOne   int8 = 1 // 一级
	DistributorLevelTwo   int8 = 2 // 二级
)

// DistributorCommission 分销佣金记录表
type DistributorCommission struct {
	ID           int64      `gorm:"primaryKey;autoIncrement;column:id" json:"id"`
	OrderID      int64      `gorm:"column:order_id;index:idx_order_id;not null;default:0" json:"order_id"`    // 订单ID
	DistributorID int64     `gorm:"column:distributor_id;index:idx_distributor_id;not null;default:0" json:"distributor_id"` // 分销商ID
	Amount       int64      `gorm:"column:amount;not null;default:0" json:"amount"`                          // 佣金金额(分)
	Ratio        float64    `gorm:"column:ratio;type:decimal(5,2);not null;default:0.00" json:"ratio"`       // 佣金比例%
	Level        int8       `gorm:"column:level;not null;default:1" json:"level"`                            // 级别 1一级 2二级
	Status       int8       `gorm:"column:status;index:idx_status;not null;default:0" json:"status"`         // 状态 0待结算 1已结算 2已回滚
	CreatedAt    *time.Time `gorm:"column:created_at;index:idx_created_at" json:"created_at"`
	UpdatedAt    *time.Time `gorm:"column:updated_at" json:"updated_at"`
}

// TableName 指定表名
func (DistributorCommission) TableName() string {
	return "distributor_commissions"
}

// 分销佣金状态常量
const (
	DistributorCommissionStatusPending  int8 = 0 // 待结算
	DistributorCommissionStatusSettled  int8 = 1 // 已结算
	DistributorCommissionStatusRollback int8 = 2 // 已回滚
)
