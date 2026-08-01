package model

import "time"

// Evaluation 评价表模型
type Evaluation struct {
	ID        int64      `gorm:"primaryKey;autoIncrement;column:id" json:"id"`
	OrderID   int64      `gorm:"column:order_id;index:idx_order_id;not null;default:0" json:"order_id"`   // 订单ID
	UserID    int64      `gorm:"column:user_id;index:idx_user_id;not null;default:0" json:"user_id"`     // 客户ID
	PlayerID  int64      `gorm:"column:player_id;index:idx_player_id;not null;default:0" json:"player_id"` // 打手ID
	Score     int        `gorm:"column:score;not null;default:5" json:"score"`                            // 评分 1-5
	Content   string     `gorm:"column:content;size:500;not null;default:''" json:"content"`             // 评价内容
	Status    string     `gorm:"column:status;index:idx_status;size:32;not null;default:'pending'" json:"status"` // 状态 pending/displayed/deleted
	CreatedAt *time.Time `gorm:"column:created_at;index:idx_created_at" json:"created_at"`
	DisplayAt *time.Time `gorm:"column:display_at" json:"display_at"`                                    // 展示时间
	UpdatedAt *time.Time `gorm:"column:updated_at" json:"updated_at"`
}

// TableName 指定表名
func (Evaluation) TableName() string {
	return "evaluations"
}

// 评价状态常量
const (
	EvaluationStatusPending   = "pending"
	EvaluationStatusDisplayed = "displayed"
	EvaluationStatusDeleted   = "deleted"
)

// Reward 打赏表
type Reward struct {
	ID        int64      `gorm:"primaryKey;autoIncrement;column:id" json:"id"`
	OrderID   int64      `gorm:"column:order_id;index:idx_order_id;not null;default:0" json:"order_id"`   // 订单ID
	UserID    int64      `gorm:"column:user_id;index:idx_user_id;not null;default:0" json:"user_id"`     // 客户ID
	PlayerID  int64      `gorm:"column:player_id;index:idx_player_id;not null;default:0" json:"player_id"` // 打手ID
	Amount    int64      `gorm:"column:amount;not null;default:0" json:"amount"`                         // 打赏金额(分)
	GiftType  string     `gorm:"column:gift_type;size:32;not null;default:''" json:"gift_type"`          // 礼物类型
	CreatedAt *time.Time `gorm:"column:created_at;index:idx_created_at" json:"created_at"`
}

// TableName 指定表名
func (Reward) TableName() string {
	return "rewards"
}
