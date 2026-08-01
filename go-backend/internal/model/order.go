package model

import "time"

// Order 订单表模型
type Order struct {
	ID              int64      `gorm:"primaryKey;autoIncrement;column:id" json:"id"`
	OrderNo         string     `gorm:"column:order_no;uniqueIndex:uk_order_no;size:64;not null;default:''" json:"order_no"` // 订单号
	Type            int8       `gorm:"column:type;not null;default:1" json:"type"`                            // 订单类型 1=即时 2=预约 3=车队 4=教学
	UserID          int64      `gorm:"column:user_id;index:idx_user_id;not null;default:0" json:"user_id"`   // 客户ID
	PlayerID        int64      `gorm:"column:player_id;index:idx_player_id;not null;default:0" json:"player_id"` // 打手ID
	ClubID          int64      `gorm:"column:club_id;index:idx_club_id;not null;default:0" json:"club_id"`   // 俱乐部ID
	ServiceID       int64      `gorm:"column:service_id;not null;default:0" json:"service_id"`               // 服务项目ID
	Amount          int64      `gorm:"column:amount;not null;default:0" json:"amount"`                        // 订单金额(分)
	Status          int8       `gorm:"column:status;index:idx_status;not null;default:0" json:"status"`       // 状态 0=待接单 1=已接单 2=进行中 3=待验收 4=已完成 5=待结算 6=已结算 10=超时取消 11=大额验证失败 12=已退款
	PayStatus       int8       `gorm:"column:pay_status;not null;default:0" json:"pay_status"`                // 支付状态 0未支付 1已支付 2已退款 3部分退款
	PaidAt          *time.Time `gorm:"column:paid_at" json:"paid_at"`                                        // 支付时间
	AcceptedAt      *time.Time `gorm:"column:accepted_at" json:"accepted_at"`                                // 接单时间
	StartedAt       *time.Time `gorm:"column:started_at" json:"started_at"`                                  // 开始服务时间
	EndedAt         *time.Time `gorm:"column:ended_at" json:"ended_at"`                                      // 结束服务时间
	SettledAt       *time.Time `gorm:"column:settled_at" json:"settled_at"`                                  // 结算时间
	RefundAmount    int64      `gorm:"column:refund_amount;not null;default:0" json:"refund_amount"`          // 退款金额(分)
	IsMinorOrder    int8       `gorm:"column:is_minor_order;not null;default:0" json:"is_minor_order"`        // 是否未成年人订单 0否 1是
	FaceSessionID   string     `gorm:"column:face_session_id;size:64;not null;default:''" json:"face_session_id"` // 活体会话ID
	AppointmentTime *time.Time `gorm:"column:appointment_time" json:"appointment_time"`                       // 预约时间
	TeamCount       int        `gorm:"column:team_count;not null;default:1" json:"team_count"`               // 车队人数
	CreatedAt       *time.Time `gorm:"column:created_at;index:idx_created_at" json:"created_at"`
	UpdatedAt       *time.Time `gorm:"column:updated_at" json:"updated_at"`
}

// TableName 指定表名
func (Order) TableName() string {
	return "orders"
}

// 订单状态常量
const (
	OrderStatusPending     int8 = 0  // 待接单
	OrderStatusAccepted    int8 = 1  // 已接单
	OrderStatusInProgress  int8 = 2  // 进行中
	OrderStatusToVerify    int8 = 3  // 待验收
	OrderStatusCompleted   int8 = 4  // 已完成
	OrderStatusToSettle    int8 = 5  // 待结算
	OrderStatusSettled     int8 = 6  // 已结算
	OrderStatusTeamPending int8 = 7  // 车队匹配中(等待满员)
	OrderStatusTimeout     int8 = 10 // 超时取消
	OrderStatusVerifyFail  int8 = 11 // 大额验证失败
	OrderStatusRefunded    int8 = 12 // 已退款
	OrderStatusCanceled    int8 = 13 // 已取消(用户主动取消)
)

// 订单类型常量
const (
	OrderTypeInstant    int8 = 1 // 即时
	OrderTypeAppointment int8 = 2 // 预约
	OrderTypeTeam       int8 = 3 // 车队
	OrderTypeTeaching   int8 = 4 // 教学
)

// OrderStatusLog 订单状态流转日志表
type OrderStatusLog struct {
	ID           int64      `gorm:"primaryKey;autoIncrement;column:id" json:"id"`
	OrderID      int64      `gorm:"column:order_id;index:idx_order_id;not null;default:0" json:"order_id"`    // 订单ID
	FromStatus   int8       `gorm:"column:from_status;not null;default:0" json:"from_status"`                // 原状态
	ToStatus     int8       `gorm:"column:to_status;not null;default:0" json:"to_status"`                    // 新状态
	OperatorID   int64      `gorm:"column:operator_id;not null;default:0" json:"operator_id"`                // 操作人ID
	OperatorType string     `gorm:"column:operator_type;size:32;not null;default:''" json:"operator_type"`   // 操作人类型 user/admin/system
	Reason       string     `gorm:"column:reason;size:255;not null;default:''" json:"reason"`               // 变更原因
	CreatedAt    *time.Time `gorm:"column:created_at;index:idx_created_at" json:"created_at"`
}

// TableName 指定表名
func (OrderStatusLog) TableName() string {
	return "order_status_logs"
}

// OrderEvidence 订单履约凭证表
type OrderEvidence struct {
	ID        int64      `gorm:"primaryKey;autoIncrement;column:id" json:"id"`
	OrderID   int64      `gorm:"column:order_id;index:idx_order_id;not null;default:0" json:"order_id"`   // 订单ID
	UserID    int64      `gorm:"column:user_id;index:idx_user_id;not null;default:0" json:"user_id"`     // 上传用户ID
	Type      string     `gorm:"column:type;size:32;not null;default:''" json:"type"`                    // 类型 video/screenshot
	FileURL   string     `gorm:"column:file_url;size:512;not null;default:''" json:"file_url"`           // 文件URL
	CreatedAt *time.Time `gorm:"column:created_at;index:idx_created_at" json:"created_at"`
}

// TableName 指定表名
func (OrderEvidence) TableName() string {
	return "order_evidence"
}
