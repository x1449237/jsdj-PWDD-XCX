package model

import "time"

// GroupBuyActivity 拼团活动表模型
type GroupBuyActivity struct {
	ID            int64      `gorm:"primaryKey;autoIncrement;column:id" json:"id"`
	Name          string     `gorm:"column:name;size:128;not null;default:''" json:"name"`
	MinMembers    int        `gorm:"column:min_members;not null;default:2" json:"min_members"`
	MaxMembers    int        `gorm:"column:max_members;not null;default:5" json:"max_members"`
	DiscountRatio float64    `gorm:"column:discount_ratio;type:decimal(3,2);not null;default:0.00" json:"discount_ratio"`
	ServiceTypeID int64      `gorm:"column:service_type_id;not null;default:0" json:"service_type_id"`
	ServiceID     int64      `gorm:"column:service_id;not null;default:0" json:"service_id"`
	MinSpend      int64      `gorm:"column:min_spend_bigint;not null;default:0" json:"min_spend"`
	Status        int8       `gorm:"column:status;index:idx_status;not null;default:1" json:"status"`
	StartAt       *time.Time `gorm:"column:start_at" json:"start_at"`
	EndAt         *time.Time `gorm:"column:end_at" json:"end_at"`
	CreatedAt     *time.Time `gorm:"column:created_at;index:idx_created_at" json:"created_at"`
	UpdatedAt     *time.Time `gorm:"column:updated_at" json:"updated_at"`
}

// TableName 指定表名
func (GroupBuyActivity) TableName() string {
	return "group_buy_activities"
}

// 拼团活动状态常量
const (
	GroupBuyActivityStatusDisabled int8 = 0
	GroupBuyActivityStatusEnabled  int8 = 1
)

// GroupBuyGroup 拼团团组表模型
type GroupBuyGroup struct {
	ID          int64      `gorm:"primaryKey;autoIncrement;column:id" json:"id"`
	ActivityID  int64      `gorm:"column:activity_id;index:idx_activity_id;not null;default:0" json:"activity_id"`
	LeaderUID   int64      `gorm:"column:leader_uid;index:idx_leader_uid;not null;default:0" json:"leader_uid"`
	MemberCount int        `gorm:"column:member_count;not null;default:0" json:"member_count"`
	Status      string     `gorm:"column:status;index:idx_status;size:32;not null;default:'pending'" json:"status"`
	ExpiredAt   *time.Time `gorm:"column:expired_at" json:"expired_at"`
	CreatedAt   *time.Time `gorm:"column:created_at;index:idx_created_at" json:"created_at"`
	UpdatedAt   *time.Time `gorm:"column:updated_at" json:"updated_at"`
}

// TableName 指定表名
func (GroupBuyGroup) TableName() string {
	return "group_buy_groups"
}

// 拼团团组状态常量
const (
	GroupBuyGroupStatusPending = "pending"
	GroupBuyGroupStatusFull    = "full"
	GroupBuyGroupStatusSuccess = "success"
	GroupBuyGroupStatusFailed  = "failed"
)

// GroupBuyMember 拼团成员表模型
type GroupBuyMember struct {
	ID         int64      `gorm:"primaryKey;autoIncrement;column:id" json:"id"`
	GroupID    int64      `gorm:"column:group_id;index:idx_group_id;not null;default:0" json:"group_id"`
	ActivityID int64      `gorm:"column:activity_id;index:idx_activity_id;not null;default:0" json:"activity_id"`
	UID        int64      `gorm:"column:uid;index:idx_uid;not null;default:0" json:"uid"`
	JoinedAt   *time.Time `gorm:"column:joined_at" json:"joined_at"`
	IsLeader   int8       `gorm:"column:is_leader;not null;default:0" json:"is_leader"`
	OrderID    int64      `gorm:"column:order_id;not null;default:0" json:"order_id"`
	CreatedAt  *time.Time `gorm:"column:created_at;index:idx_created_at" json:"created_at"`
}

// TableName 指定表名
func (GroupBuyMember) TableName() string {
	return "group_buy_members"
}
