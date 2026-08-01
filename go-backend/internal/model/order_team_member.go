package model

import "time"

// OrderTeamMember 车队订单成员表
type OrderTeamMember struct {
	ID        int64      `gorm:"primaryKey;autoIncrement;column:id" json:"id"`
	OrderID   int64      `gorm:"column:order_id;index:idx_order_id;not null;default:0" json:"order_id"`
	PlayerID  int64      `gorm:"column:player_id;index:idx_player_id;not null;default:0" json:"player_id"`
	JoinedAt  *time.Time `gorm:"column:joined_at" json:"joined_at"`
	Status    string     `gorm:"column:status;size:32;not null;default:'pending'" json:"status"`
	CreatedAt *time.Time `gorm:"column:created_at;index:idx_created_at" json:"created_at"`
}

// TableName 指定表名
func (OrderTeamMember) TableName() string {
	return "order_team_members"
}

// 车队订单成员状态常量
const (
	TeamMemberStatusPending = "pending"
	TeamMemberStatusJoined  = "joined"
	TeamMemberStatusLeft    = "left"
)

// 车队订单状态扩展（在 Order 中使用 Status，或单独扩展 team_status）
const (
	OrderTeamStatusPending = "team_pending" // 等车手组队中
)
