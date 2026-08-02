package model

import (
	"time"

	"gorm.io/datatypes"
)

// ============================================================
// 俱乐部扩展模块
// ============================================================

// ClubHomeDecoration 俱乐部主页装修
type ClubHomeDecoration struct {
	ID           int64          `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	ClubID       int64          `gorm:"column:club_id;not null;default:0" json:"club_id"`
	Banners      datatypes.JSON `gorm:"column:banners" json:"banners"`
	Intro        string         `gorm:"column:intro;not null;default:''" json:"intro"`
	MainGames    datatypes.JSON `gorm:"column:main_games" json:"main_games"`
	PriceDisplay datatypes.JSON `gorm:"column:price_display" json:"price_display"`
	IsPublished  int8           `gorm:"column:is_published;not null;default:0" json:"is_published"`
	CreatedAt    *time.Time     `gorm:"column:created_at" json:"created_at"`
	UpdatedAt    *time.Time     `gorm:"column:updated_at" json:"updated_at"`
}

func (ClubHomeDecoration) TableName() string {
	return "club_home_decorations"
}

// ClubService 俱乐部服务项
type ClubService struct {
	ID        int64      `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	ClubID    int64      `gorm:"column:club_id;not null;default:0" json:"club_id"`
	Name      string     `gorm:"column:name;not null;default:''" json:"name"`
	Duration  int        `gorm:"column:duration;not null;default:0" json:"duration"`
	Price     int64      `gorm:"column:price;not null;default:0" json:"price"`
	Intro     string     `gorm:"column:intro;not null;default:''" json:"intro"`
	GameID    int        `gorm:"column:game_id;not null;default:0" json:"game_id"`
	Status    int8       `gorm:"column:status;not null;default:0" json:"status"`
	Sort      int        `gorm:"column:sort;not null;default:0" json:"sort"`
	CreatedAt *time.Time `gorm:"column:created_at" json:"created_at"`
	UpdatedAt *time.Time `gorm:"column:updated_at" json:"updated_at"`
}

func (ClubService) TableName() string {
	return "club_services"
}

// ClubMemberCard 俱乐部成员名片
type ClubMemberCard struct {
	ID         int64      `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	ClubID     int64      `gorm:"column:club_id;not null;default:0" json:"club_id"`
	MemberUID  int64      `gorm:"column:member_uid;not null;default:0" json:"member_uid"`
	PosterURL  string     `gorm:"column:poster_url;not null;default:''" json:"poster_url"`
	ShareTitle string     `gorm:"column:share_title;not null;default:''" json:"share_title"`
	ShareDesc  string     `gorm:"column:share_desc;not null;default:''" json:"share_desc"`
	QrCodeURL  string     `gorm:"column:qr_code_url;not null;default:''" json:"qr_code_url"`
	CreatedAt  *time.Time `gorm:"column:created_at" json:"created_at"`
	UpdatedAt  *time.Time `gorm:"column:updated_at" json:"updated_at"`
}

func (ClubMemberCard) TableName() string {
	return "club_member_cards"
}

// ClubMemberProfile 俱乐部成员资料
type ClubMemberProfile struct {
	ID         int64          `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	ClubID     int64          `gorm:"column:club_id;not null;default:0" json:"club_id"`
	MemberUID  int64          `gorm:"column:member_uid;not null;default:0" json:"member_uid"`
	GameID     int            `gorm:"column:game_id;not null;default:0" json:"game_id"`
	RankInfo   string         `gorm:"column:rank_info;not null;default:''" json:"rank_info"`
	WinRate    float64        `gorm:"column:win_rate;not null;default:0" json:"win_rate"`
	TotalGames int            `gorm:"column:total_games;not null;default:0" json:"total_games"`
	HeroTags   datatypes.JSON `gorm:"column:hero_tags" json:"hero_tags"`
	CreatedAt  *time.Time     `gorm:"column:created_at" json:"created_at"`
	UpdatedAt  *time.Time     `gorm:"column:updated_at" json:"updated_at"`
}

func (ClubMemberProfile) TableName() string {
	return "club_member_profiles"
}

// ClubPermissionLog 俱乐部权限操作日志
type ClubPermissionLog struct {
	ID           int64          `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	ClubID       int64          `gorm:"column:club_id;not null;default:0" json:"club_id"`
	OperatorUID  int64          `gorm:"column:operator_uid;not null;default:0" json:"operator_uid"`
	OperatorRole int8           `gorm:"column:operator_role;not null;default:0" json:"operator_role"`
	Action       string         `gorm:"column:action;not null;default:''" json:"action"`
	TargetUID    int64          `gorm:"column:target_uid;not null;default:0" json:"target_uid"`
	Detail       datatypes.JSON `gorm:"column:detail" json:"detail"`
	CreatedAt    *time.Time     `gorm:"column:created_at" json:"created_at"`
	UpdatedAt    *time.Time     `gorm:"column:updated_at" json:"updated_at"`
}

func (ClubPermissionLog) TableName() string {
	return "club_permission_logs"
}

// ClubMemberResignation 俱乐部成员退会申请
type ClubMemberResignation struct {
	ID         int64      `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	ClubID     int64      `gorm:"column:club_id;not null;default:0" json:"club_id"`
	MemberUID  int64      `gorm:"column:member_uid;not null;default:0" json:"member_uid"`
	Reason     string     `gorm:"column:reason;not null;default:''" json:"reason"`
	Status     int8       `gorm:"column:status;not null;default:0" json:"status"`
	AuditorUID int64      `gorm:"column:auditor_uid;not null;default:0" json:"auditor_uid"`
	AuditedAt  *time.Time `gorm:"column:audited_at" json:"audited_at"`
	CreatedAt  *time.Time `gorm:"column:created_at" json:"created_at"`
	UpdatedAt  *time.Time `gorm:"column:updated_at" json:"updated_at"`
}

func (ClubMemberResignation) TableName() string {
	return "club_member_resignations"
}

// ClubBlacklist 俱乐部黑名单
type ClubBlacklist struct {
	ID          int64      `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	ClubID      int64      `gorm:"column:club_id;not null;default:0" json:"club_id"`
	UserID      int64      `gorm:"column:user_id;not null;default:0" json:"user_id"`
	Reason      string     `gorm:"column:reason;not null;default:''" json:"reason"`
	OperatorUID int64      `gorm:"column:operator_uid;not null;default:0" json:"operator_uid"`
	Type        int8       `gorm:"column:type;not null;default:0" json:"type"`
	CreatedAt   *time.Time `gorm:"column:created_at" json:"created_at"`
	UpdatedAt   *time.Time `gorm:"column:updated_at" json:"updated_at"`
}

func (ClubBlacklist) TableName() string {
	return "club_blacklists"
}

// ClubPointRule 俱乐部积分规则
type ClubPointRule struct {
	ID          int64      `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	ClubID      int64      `gorm:"column:club_id;not null;default:0" json:"club_id"`
	Action      string     `gorm:"column:action;not null;default:''" json:"action"`
	Points      int        `gorm:"column:points;not null;default:0" json:"points"`
	Description string     `gorm:"column:description;not null;default:''" json:"description"`
	CreatedAt   *time.Time `gorm:"column:created_at" json:"created_at"`
	UpdatedAt   *time.Time `gorm:"column:updated_at" json:"updated_at"`
}

func (ClubPointRule) TableName() string {
	return "club_point_rules"
}

// ClubPointLog 俱乐部积分流水
type ClubPointLog struct {
	ID        int64      `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	ClubID    int64      `gorm:"column:club_id;not null;default:0" json:"club_id"`
	MemberUID int64      `gorm:"column:member_uid;not null;default:0" json:"member_uid"`
	Action    string     `gorm:"column:action;not null;default:''" json:"action"`
	Points    int        `gorm:"column:points;not null;default:0" json:"points"`
	Balance   int        `gorm:"column:balance;not null;default:0" json:"balance"`
	Remark    string     `gorm:"column:remark;not null;default:''" json:"remark"`
	CreatedAt *time.Time `gorm:"column:created_at" json:"created_at"`
	UpdatedAt *time.Time `gorm:"column:updated_at" json:"updated_at"`
}

func (ClubPointLog) TableName() string {
	return "club_point_logs"
}

// ClubFeeRule 俱乐部费用规则
type ClubFeeRule struct {
	ID        int64      `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	ClubID    int64      `gorm:"column:club_id;not null;default:0" json:"club_id"`
	FeeType   int8       `gorm:"column:fee_type;not null;default:0" json:"fee_type"`
	Amount    int64      `gorm:"column:amount;not null;default:0" json:"amount"`
	DeductDay int8       `gorm:"column:deduct_day;not null;default:0" json:"deduct_day"`
	Status    int8       `gorm:"column:status;not null;default:0" json:"status"`
	CreatedAt *time.Time `gorm:"column:created_at" json:"created_at"`
	UpdatedAt *time.Time `gorm:"column:updated_at" json:"updated_at"`
}

func (ClubFeeRule) TableName() string {
	return "club_fee_rules"
}

// ClubRecruitCard 俱乐部招募卡
type ClubRecruitCard struct {
	ID        int64      `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	ClubID    int64      `gorm:"column:club_id;not null;default:0" json:"club_id"`
	Title     string     `gorm:"column:title;not null;default:''" json:"title"`
	Content   string     `gorm:"column:content;not null;default:''" json:"content"`
	PosterURL string     `gorm:"column:poster_url;not null;default:''" json:"poster_url"`
	ShareLink string     `gorm:"column:share_link;not null;default:''" json:"share_link"`
	Status    int8       `gorm:"column:status;not null;default:0" json:"status"`
	CreatedAt *time.Time `gorm:"column:created_at" json:"created_at"`
	UpdatedAt *time.Time `gorm:"column:updated_at" json:"updated_at"`
}

func (ClubRecruitCard) TableName() string {
	return "club_recruit_cards"
}

// ClubAdminTodo 俱乐部管理员待办
type ClubAdminTodo struct {
	ID        int64      `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	ClubID    int64      `gorm:"column:club_id;not null;default:0" json:"club_id"`
	AdminUID  int64      `gorm:"column:admin_uid;not null;default:0" json:"admin_uid"`
	Type      string     `gorm:"column:type;not null;default:''" json:"type"`
	TargetID  int64      `gorm:"column:target_id;not null;default:0" json:"target_id"`
	Title     string     `gorm:"column:title;not null;default:''" json:"title"`
	Status    int8       `gorm:"column:status;not null;default:0" json:"status"`
	CreatedAt *time.Time `gorm:"column:created_at" json:"created_at"`
	UpdatedAt *time.Time `gorm:"column:updated_at" json:"updated_at"`
}

func (ClubAdminTodo) TableName() string {
	return "club_admin_todos"
}

// ClubGameZone 俱乐部游戏分区
type ClubGameZone struct {
	ID             int64      `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	ClubID         int64      `gorm:"column:club_id;not null;default:0" json:"club_id"`
	GameID         int        `gorm:"column:game_id;not null;default:0" json:"game_id"`
	GameName       string     `gorm:"column:game_name;not null;default:''" json:"game_name"`
	CommissionRate float64    `gorm:"column:commission_rate;not null;default:0" json:"commission_rate"`
	ManagerUID     int64      `gorm:"column:manager_uid;not null;default:0" json:"manager_uid"`
	Status         int8       `gorm:"column:status;not null;default:0" json:"status"`
	CreatedAt      *time.Time `gorm:"column:created_at" json:"created_at"`
	UpdatedAt      *time.Time `gorm:"column:updated_at" json:"updated_at"`
}

func (ClubGameZone) TableName() string {
	return "club_game_zones"
}

// ClubTempCommissionRule 俱乐部临时抽成规则
type ClubTempCommissionRule struct {
	ID             int64      `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	ClubID         int64      `gorm:"column:club_id;not null;default:0" json:"club_id"`
	GameID         int        `gorm:"column:game_id;not null;default:0" json:"game_id"`
	CommissionRate float64    `gorm:"column:commission_rate;not null;default:0" json:"commission_rate"`
	StartTime      *time.Time `gorm:"column:start_time" json:"start_time"`
	EndTime        *time.Time `gorm:"column:end_time" json:"end_time"`
	Reason         string     `gorm:"column:reason;not null;default:''" json:"reason"`
	Status         int8       `gorm:"column:status;not null;default:0" json:"status"`
	CreatedAt      *time.Time `gorm:"column:created_at" json:"created_at"`
	UpdatedAt      *time.Time `gorm:"column:updated_at" json:"updated_at"`
}

func (ClubTempCommissionRule) TableName() string {
	return "club_temp_commission_rules"
}

// ClubMemberLeave 俱乐部成员请假
type ClubMemberLeave struct {
	ID         int64      `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	ClubID     int64      `gorm:"column:club_id;not null;default:0" json:"club_id"`
	MemberUID  int64      `gorm:"column:member_uid;not null;default:0" json:"member_uid"`
	LeaveType  int8       `gorm:"column:leave_type;not null;default:0" json:"leave_type"`
	StartTime  *time.Time `gorm:"column:start_time" json:"start_time"`
	EndTime    *time.Time `gorm:"column:end_time" json:"end_time"`
	Reason     string     `gorm:"column:reason;not null;default:''" json:"reason"`
	Status     int8       `gorm:"column:status;not null;default:0" json:"status"`
	AuditorUID int64      `gorm:"column:auditor_uid;not null;default:0" json:"auditor_uid"`
	AuditedAt  *time.Time `gorm:"column:audited_at" json:"audited_at"`
	CreatedAt  *time.Time `gorm:"column:created_at" json:"created_at"`
	UpdatedAt  *time.Time `gorm:"column:updated_at" json:"updated_at"`
}

func (ClubMemberLeave) TableName() string {
	return "club_member_leaves"
}

// ClubMemberChangeRequest 俱乐部成员变更申请
type ClubMemberChangeRequest struct {
	ID         int64      `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	ClubID     int64      `gorm:"column:club_id;not null;default:0" json:"club_id"`
	MemberUID  int64      `gorm:"column:member_uid;not null;default:0" json:"member_uid"`
	Field      string     `gorm:"column:field;not null;default:''" json:"field"`
	OldValue   string     `gorm:"column:old_value;not null;default:''" json:"old_value"`
	NewValue   string     `gorm:"column:new_value;not null;default:''" json:"new_value"`
	Status     int8       `gorm:"column:status;not null;default:0" json:"status"`
	AuditorUID int64      `gorm:"column:auditor_uid;not null;default:0" json:"auditor_uid"`
	AuditedAt  *time.Time `gorm:"column:audited_at" json:"audited_at"`
	CreatedAt  *time.Time `gorm:"column:created_at" json:"created_at"`
	UpdatedAt  *time.Time `gorm:"column:updated_at" json:"updated_at"`
}

func (ClubMemberChangeRequest) TableName() string {
	return "club_member_change_requests"
}

// ClubPriorityDispatch 俱乐部优先派单
type ClubPriorityDispatch struct {
	ID        int64      `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	ClubID    int64      `gorm:"column:club_id;not null;default:0" json:"club_id"`
	MemberUID int64      `gorm:"column:member_uid;not null;default:0" json:"member_uid"`
	Priority  int        `gorm:"column:priority;not null;default:0" json:"priority"`
	CreatedAt *time.Time `gorm:"column:created_at" json:"created_at"`
	UpdatedAt *time.Time `gorm:"column:updated_at" json:"updated_at"`
}

func (ClubPriorityDispatch) TableName() string {
	return "club_priority_dispatch"
}

// ClubInternalResource 俱乐部内部资源
type ClubInternalResource struct {
	ID          int64      `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	ClubID      int64      `gorm:"column:club_id;not null;default:0" json:"club_id"`
	MemberUID   int64      `gorm:"column:member_uid;not null;default:0" json:"member_uid"`
	Type        int8       `gorm:"column:type;not null;default:0" json:"type"`
	Amount      int64      `gorm:"column:amount;not null;default:0" json:"amount"`
	Description string     `gorm:"column:description;not null;default:''" json:"description"`
	OrderID     int64      `gorm:"column:order_id;not null;default:0" json:"order_id"`
	CreatedAt   *time.Time `gorm:"column:created_at" json:"created_at"`
	UpdatedAt   *time.Time `gorm:"column:updated_at" json:"updated_at"`
}

func (ClubInternalResource) TableName() string {
	return "club_internal_resources"
}

// ClubCustomerRelation 俱乐部客户关系
type ClubCustomerRelation struct {
	ID             int64      `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	ClubID         int64      `gorm:"column:club_id;not null;default:0" json:"club_id"`
	CustomerUID    int64      `gorm:"column:customer_uid;not null;default:0" json:"customer_uid"`
	ReceptionistUID int64     `gorm:"column:receptionist_uid;not null;default:0" json:"receptionist_uid"`
	RebateRate     float64    `gorm:"column:rebate_rate;not null;default:0" json:"rebate_rate"`
	TotalOrders    int        `gorm:"column:total_orders;not null;default:0" json:"total_orders"`
	TotalAmount    int64      `gorm:"column:total_amount;not null;default:0" json:"total_amount"`
	CreatedAt      *time.Time `gorm:"column:created_at" json:"created_at"`
	UpdatedAt      *time.Time `gorm:"column:updated_at" json:"updated_at"`
}

func (ClubCustomerRelation) TableName() string {
	return "club_customer_relations"
}

// ClubTemplatePhrase 俱乐部话术模板
type ClubTemplatePhrase struct {
	ID        int64      `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	ClubID    int64      `gorm:"column:club_id;not null;default:0" json:"club_id"`
	Category  string     `gorm:"column:category;not null;default:''" json:"category"`
	Title     string     `gorm:"column:title;not null;default:''" json:"title"`
	Content   string     `gorm:"column:content;not null;default:''" json:"content"`
	Sort      int        `gorm:"column:sort;not null;default:0" json:"sort"`
	Status    int8       `gorm:"column:status;not null;default:0" json:"status"`
	CreatedAt *time.Time `gorm:"column:created_at" json:"created_at"`
	UpdatedAt *time.Time `gorm:"column:updated_at" json:"updated_at"`
}

func (ClubTemplatePhrase) TableName() string {
	return "club_template_phrases"
}

// ClubMemberRanking 俱乐部成员排行
type ClubMemberRanking struct {
	ID          int64      `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	ClubID      int64      `gorm:"column:club_id;not null;default:0" json:"club_id"`
	MemberUID   int64      `gorm:"column:member_uid;not null;default:0" json:"member_uid"`
	PeriodType  int8       `gorm:"column:period_type;not null;default:0" json:"period_type"`
	PeriodDate  string     `gorm:"column:period_date;not null;default:''" json:"period_date"`
	OrderCount  int        `gorm:"column:order_count;not null;default:0" json:"order_count"`
	Income      int64      `gorm:"column:income;not null;default:0" json:"income"`
	Rank        int        `gorm:"column:rank;not null;default:0" json:"rank"`
	CreatedAt   *time.Time `gorm:"column:created_at" json:"created_at"`
	UpdatedAt   *time.Time `gorm:"column:updated_at" json:"updated_at"`
}

func (ClubMemberRanking) TableName() string {
	return "club_member_rankings"
}

// ============================================================
// 订单扩展模块
// ============================================================

// OrderTemplate 订单模板
type OrderTemplate struct {
	ID         int64      `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	UserID     int64      `gorm:"column:user_id;not null;default:0" json:"user_id"`
	ClubID     int64      `gorm:"column:club_id;not null;default:0" json:"club_id"`
	Name       string     `gorm:"column:name;not null;default:''" json:"name"`
	Type       int8       `gorm:"column:type;not null;default:0" json:"type"`
	GameID     int        `gorm:"column:game_id;not null;default:0" json:"game_id"`
	GameZone   string     `gorm:"column:game_zone;not null;default:''" json:"game_zone"`
	GameIDText string     `gorm:"column:game_id_text;not null;default:''" json:"game_id_text"`
	Requirement string    `gorm:"column:requirement;not null;default:''" json:"requirement"`
	CreatedAt  *time.Time `gorm:"column:created_at" json:"created_at"`
	UpdatedAt  *time.Time `gorm:"column:updated_at" json:"updated_at"`
}

func (OrderTemplate) TableName() string {
	return "order_templates"
}

// OrderSupplement 订单补充说明
type OrderSupplement struct {
	ID        int64          `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	OrderID   int64          `gorm:"column:order_id;not null;default:0" json:"order_id"`
	UserID    int64          `gorm:"column:user_id;not null;default:0" json:"user_id"`
	Content   string         `gorm:"column:content;not null;default:''" json:"content"`
	FileURLs  datatypes.JSON `gorm:"column:file_urls" json:"file_urls"`
	CreatedAt *time.Time     `gorm:"column:created_at" json:"created_at"`
	UpdatedAt *time.Time     `gorm:"column:updated_at" json:"updated_at"`
}

func (OrderSupplement) TableName() string {
	return "order_supplements"
}

// OrderPartialRefund 订单部分退款
type OrderPartialRefund struct {
	ID           int64      `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	OrderID      int64      `gorm:"column:order_id;not null;default:0" json:"order_id"`
	Amount       int64      `gorm:"column:amount;not null;default:0" json:"amount"`
	Reason       string     `gorm:"column:reason;not null;default:''" json:"reason"`
	OperatorID   int64      `gorm:"column:operator_id;not null;default:0" json:"operator_id"`
	OperatorType string     `gorm:"column:operator_type;not null;default:''" json:"operator_type"`
	Status       int8       `gorm:"column:status;not null;default:0" json:"status"`
	CreatedAt    *time.Time `gorm:"column:created_at" json:"created_at"`
	UpdatedAt    *time.Time `gorm:"column:updated_at" json:"updated_at"`
}

func (OrderPartialRefund) TableName() string {
	return "order_partial_refunds"
}

// OrderTag 订单标签
type OrderTag struct {
	ID        int64      `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	Name      string     `gorm:"column:name;not null;default:''" json:"name"`
	Color     string     `gorm:"column:color;not null;default:''" json:"color"`
	CreatedAt *time.Time `gorm:"column:created_at" json:"created_at"`
	UpdatedAt *time.Time `gorm:"column:updated_at" json:"updated_at"`
}

func (OrderTag) TableName() string {
	return "order_tags"
}

// OrderTagRelation 订单标签关联
type OrderTagRelation struct {
	ID        int64      `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	OrderID   int64      `gorm:"column:order_id;not null;default:0" json:"order_id"`
	TagID     int64      `gorm:"column:tag_id;not null;default:0" json:"tag_id"`
	CreatedAt *time.Time `gorm:"column:created_at" json:"created_at"`
	UpdatedAt *time.Time `gorm:"column:updated_at" json:"updated_at"`
}

func (OrderTagRelation) TableName() string {
	return "order_tag_relations"
}

// OrderRemark 订单备注
type OrderRemark struct {
	ID        int64      `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	OrderID   int64      `gorm:"column:order_id;not null;default:0" json:"order_id"`
	UserID    int64      `gorm:"column:user_id;not null;default:0" json:"user_id"`
	UserType  string     `gorm:"column:user_type;not null;default:''" json:"user_type"`
	Content   string     `gorm:"column:content;not null;default:''" json:"content"`
	IsPinned  int8       `gorm:"column:is_pinned;not null;default:0" json:"is_pinned"`
	CreatedAt *time.Time `gorm:"column:created_at" json:"created_at"`
	UpdatedAt *time.Time `gorm:"column:updated_at" json:"updated_at"`
}

func (OrderRemark) TableName() string {
	return "order_remarks"
}

// OrderExtension 订单加时
type OrderExtension struct {
	ID            int64      `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	OrderID       int64      `gorm:"column:order_id;not null;default:0" json:"order_id"`
	PlayerID      int64      `gorm:"column:player_id;not null;default:0" json:"player_id"`
	ExtendMinutes int        `gorm:"column:extend_minutes;not null;default:0" json:"extend_minutes"`
	Reason        string     `gorm:"column:reason;not null;default:''" json:"reason"`
	Status        int8       `gorm:"column:status;not null;default:0" json:"status"`
	AuditorUID    int64      `gorm:"column:auditor_uid;not null;default:0" json:"auditor_uid"`
	AuditedAt     *time.Time `gorm:"column:audited_at" json:"audited_at"`
	CreatedAt     *time.Time `gorm:"column:created_at" json:"created_at"`
	UpdatedAt     *time.Time `gorm:"column:updated_at" json:"updated_at"`
}

func (OrderExtension) TableName() string {
	return "order_extensions"
}

// OrderRefundLedger 订单退款台账
type OrderRefundLedger struct {
	ID               int64      `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	OrderID          int64      `gorm:"column:order_id;not null;default:0" json:"order_id"`
	RefundType       int8       `gorm:"column:refund_type;not null;default:0" json:"refund_type"`
	Amount           int64      `gorm:"column:amount;not null;default:0" json:"amount"`
	Reason           string     `gorm:"column:reason;not null;default:''" json:"reason"`
	OperatorID       int64      `gorm:"column:operator_id;not null;default:0" json:"operator_id"`
	ArbitrationCaseID int64     `gorm:"column:arbitration_case_id;not null;default:0" json:"arbitration_case_id"`
	CreatedAt        *time.Time `gorm:"column:created_at" json:"created_at"`
	UpdatedAt        *time.Time `gorm:"column:updated_at" json:"updated_at"`
}

func (OrderRefundLedger) TableName() string {
	return "order_refund_ledger"
}

// OrderPriceLog 订单价格变更日志
type OrderPriceLog struct {
	ID           int64      `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	OrderID      int64      `gorm:"column:order_id;not null;default:0" json:"order_id"`
	OriginalPrice int64     `gorm:"column:original_price;not null;default:0" json:"original_price"`
	NewPrice     int64      `gorm:"column:new_price;not null;default:0" json:"new_price"`
	OperatorID   int64      `gorm:"column:operator_id;not null;default:0" json:"operator_id"`
	OperatorType string     `gorm:"column:operator_type;not null;default:''" json:"operator_type"`
	Reason       string     `gorm:"column:reason;not null;default:''" json:"reason"`
	CreatedAt    *time.Time `gorm:"column:created_at" json:"created_at"`
	UpdatedAt    *time.Time `gorm:"column:updated_at" json:"updated_at"`
}

func (OrderPriceLog) TableName() string {
	return "order_price_logs"
}

// OrderTransfer 订单转接
type OrderTransfer struct {
	ID           int64      `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	OrderID      int64      `gorm:"column:order_id;not null;default:0" json:"order_id"`
	FromPlayerID int64      `gorm:"column:from_player_id;not null;default:0" json:"from_player_id"`
	ToPlayerID   int64      `gorm:"column:to_player_id;not null;default:0" json:"to_player_id"`
	ClubID       int64      `gorm:"column:club_id;not null;default:0" json:"club_id"`
	Reason       string     `gorm:"column:reason;not null;default:''" json:"reason"`
	Status       int8       `gorm:"column:status;not null;default:0" json:"status"`
	CreatedAt    *time.Time `gorm:"column:created_at" json:"created_at"`
	UpdatedAt    *time.Time `gorm:"column:updated_at" json:"updated_at"`
}

func (OrderTransfer) TableName() string {
	return "order_transfers"
}

// OrderPriceChange 订单改价申请
type OrderPriceChange struct {
	ID            int64      `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	OrderID       int64      `gorm:"column:order_id;not null;default:0" json:"order_id"`
	ApplicantID   int64      `gorm:"column:applicant_id;not null;default:0" json:"applicant_id"`
	OriginalPrice int64      `gorm:"column:original_price;not null;default:0" json:"original_price"`
	NewPrice      int64      `gorm:"column:new_price;not null;default:0" json:"new_price"`
	Reason        string     `gorm:"column:reason;not null;default:''" json:"reason"`
	Status        int8       `gorm:"column:status;not null;default:0" json:"status"`
	AuditorUID    int64      `gorm:"column:auditor_uid;not null;default:0" json:"auditor_uid"`
	AuditedAt     *time.Time `gorm:"column:audited_at" json:"audited_at"`
	CreatedAt     *time.Time `gorm:"column:created_at" json:"created_at"`
	UpdatedAt     *time.Time `gorm:"column:updated_at" json:"updated_at"`
}

func (OrderPriceChange) TableName() string {
	return "order_price_changes"
}

// ClubFavorite 俱乐部收藏
type ClubFavorite struct {
	ID        int64      `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	UserID    int64      `gorm:"column:user_id;not null;default:0" json:"user_id"`
	ClubID    int64      `gorm:"column:club_id;not null;default:0" json:"club_id"`
	CreatedAt *time.Time `gorm:"column:created_at" json:"created_at"`
	UpdatedAt *time.Time `gorm:"column:updated_at" json:"updated_at"`
}

func (ClubFavorite) TableName() string {
	return "club_favorites"
}

// ============================================================
// IM扩展模块
// ============================================================

// GroupChatFile 群聊文件
type GroupChatFile struct {
	ID          int64      `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	GroupID     int64      `gorm:"column:group_id;not null;default:0" json:"group_id"`
	UploaderUID int64      `gorm:"column:uploader_uid;not null;default:0" json:"uploader_uid"`
	FileName    string     `gorm:"column:file_name;not null;default:''" json:"file_name"`
	FileURL     string     `gorm:"column:file_url;not null;default:''" json:"file_url"`
	FileSize    int64      `gorm:"column:file_size;not null;default:0" json:"file_size"`
	FileType    string     `gorm:"column:file_type;not null;default:''" json:"file_type"`
	CreatedAt   *time.Time `gorm:"column:created_at" json:"created_at"`
	UpdatedAt   *time.Time `gorm:"column:updated_at" json:"updated_at"`
}

func (GroupChatFile) TableName() string {
	return "group_chat_files"
}

// ChatQuickReply 快捷回复
type ChatQuickReply struct {
	ID        int64      `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	ClubID    int64      `gorm:"column:club_id;not null;default:0" json:"club_id"`
	Category  string     `gorm:"column:category;not null;default:''" json:"category"`
	Title     string     `gorm:"column:title;not null;default:''" json:"title"`
	Content   string     `gorm:"column:content;not null;default:''" json:"content"`
	Sort      int        `gorm:"column:sort;not null;default:0" json:"sort"`
	Status    int8       `gorm:"column:status;not null;default:0" json:"status"`
	CreatedAt *time.Time `gorm:"column:created_at" json:"created_at"`
	UpdatedAt *time.Time `gorm:"column:updated_at" json:"updated_at"`
}

func (ChatQuickReply) TableName() string {
	return "chat_quick_replies"
}

// ChatMessageRead 消息已读记录
type ChatMessageRead struct {
	ID        int64      `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	SessionID string     `gorm:"column:session_id;not null;default:''" json:"session_id"`
	MessageID int64      `gorm:"column:message_id;not null;default:0" json:"message_id"`
	UserID    int64      `gorm:"column:user_id;not null;default:0" json:"user_id"`
	ReadAt    *time.Time `gorm:"column:read_at" json:"read_at"`
	CreatedAt *time.Time `gorm:"column:created_at" json:"created_at"`
	UpdatedAt *time.Time `gorm:"column:updated_at" json:"updated_at"`
}

func (ChatMessageRead) TableName() string {
	return "chat_message_reads"
}

// GroupChatMute 群聊禁言
type GroupChatMute struct {
	ID          int64      `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	GroupID     int64      `gorm:"column:group_id;not null;default:0" json:"group_id"`
	MemberUID   int64      `gorm:"column:member_uid;not null;default:0" json:"member_uid"`
	OperatorUID int64      `gorm:"column:operator_uid;not null;default:0" json:"operator_uid"`
	Duration    int        `gorm:"column:duration;not null;default:0" json:"duration"`
	ExpireAt    *time.Time `gorm:"column:expire_at" json:"expire_at"`
	Reason      string     `gorm:"column:reason;not null;default:''" json:"reason"`
	CreatedAt   *time.Time `gorm:"column:created_at" json:"created_at"`
	UpdatedAt   *time.Time `gorm:"column:updated_at" json:"updated_at"`
}

func (GroupChatMute) TableName() string {
	return "group_chat_mutes"
}

// ChatReport 聊天举报
type ChatReport struct {
	ID           int64          `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	ReporterUID  int64          `gorm:"column:reporter_uid;not null;default:0" json:"reporter_uid"`
	SessionID    string         `gorm:"column:session_id;not null;default:''" json:"session_id"`
	MessageID    int64          `gorm:"column:message_id;not null;default:0" json:"message_id"`
	Reason       string         `gorm:"column:reason;not null;default:''" json:"reason"`
	EvidenceURLs datatypes.JSON `gorm:"column:evidence_urls" json:"evidence_urls"`
	Status       int8           `gorm:"column:status;not null;default:0" json:"status"`
	HandlerUID   int64          `gorm:"column:handler_uid;not null;default:0" json:"handler_uid"`
	HandleResult string         `gorm:"column:handle_result;not null;default:''" json:"handle_result"`
	CreatedAt    *time.Time     `gorm:"column:created_at" json:"created_at"`
	UpdatedAt    *time.Time     `gorm:"column:updated_at" json:"updated_at"`
}

func (ChatReport) TableName() string {
	return "chat_reports"
}

// ============================================================
// 财务扩展模块
// ============================================================

// ClubFinanceLedger 俱乐部财务台账
type ClubFinanceLedger struct {
	ID          int64      `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	ClubID      int64      `gorm:"column:club_id;not null;default:0" json:"club_id"`
	Type        string     `gorm:"column:type;not null;default:''" json:"type"`
	Amount      int64      `gorm:"column:amount;not null;default:0" json:"amount"`
	OrderID     int64      `gorm:"column:order_id;not null;default:0" json:"order_id"`
	MemberUID   int64      `gorm:"column:member_uid;not null;default:0" json:"member_uid"`
	Description string     `gorm:"column:description;not null;default:''" json:"description"`
	CreatedAt   *time.Time `gorm:"column:created_at" json:"created_at"`
	UpdatedAt   *time.Time `gorm:"column:updated_at" json:"updated_at"`
}

func (ClubFinanceLedger) TableName() string {
	return "club_finance_ledgers"
}

// UserDeposit 用户存款
type UserDeposit struct {
	ID        int64      `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	UserID    int64      `gorm:"column:user_id;not null;default:0" json:"user_id"`
	Amount    int64      `gorm:"column:amount;not null;default:0" json:"amount"`
	Balance   int64      `gorm:"column:balance;not null;default:0" json:"balance"`
	ExpireAt  *time.Time `gorm:"column:expire_at" json:"expire_at"`
	Status    int8       `gorm:"column:status;not null;default:0" json:"status"`
	Source    string     `gorm:"column:source;not null;default:''" json:"source"`
	CreatedAt *time.Time `gorm:"column:created_at" json:"created_at"`
	UpdatedAt *time.Time `gorm:"column:updated_at" json:"updated_at"`
}

func (UserDeposit) TableName() string {
	return "user_deposits"
}

// MonthlySettlement 月度结算
type MonthlySettlement struct {
	ID            int64      `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	ClubID        int64      `gorm:"column:club_id;not null;default:0" json:"club_id"`
	MemberUID     int64      `gorm:"column:member_uid;not null;default:0" json:"member_uid"`
	SettleMonth   string     `gorm:"column:settle_month;not null;default:''" json:"settle_month"`
	TotalIncome   int64      `gorm:"column:total_income;not null;default:0" json:"total_income"`
	TotalRefund   int64      `gorm:"column:total_refund;not null;default:0" json:"total_refund"`
	NetIncome     int64      `gorm:"column:net_income;not null;default:0" json:"net_income"`
	PlatformFee   int64      `gorm:"column:platform_fee;not null;default:0" json:"platform_fee"`
	ClubFee       int64      `gorm:"column:club_fee;not null;default:0" json:"club_fee"`
	PlayerIncome  int64      `gorm:"column:player_income;not null;default:0" json:"player_income"`
	PdfURL        string     `gorm:"column:pdf_url;not null;default:''" json:"pdf_url"`
	Status        int8       `gorm:"column:status;not null;default:0" json:"status"`
	CreatedAt     *time.Time `gorm:"column:created_at" json:"created_at"`
	UpdatedAt     *time.Time `gorm:"column:updated_at" json:"updated_at"`
}

func (MonthlySettlement) TableName() string {
	return "monthly_settlements"
}

// RebateRecord 返利记录
type RebateRecord struct {
	ID             int64      `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	ClubID         int64      `gorm:"column:club_id;not null;default:0" json:"club_id"`
	CustomerUID    int64      `gorm:"column:customer_uid;not null;default:0" json:"customer_uid"`
	ReceptionistUID int64     `gorm:"column:receptionist_uid;not null;default:0" json:"receptionist_uid"`
	Amount         int64      `gorm:"column:amount;not null;default:0" json:"amount"`
	OrderID        int64      `gorm:"column:order_id;not null;default:0" json:"order_id"`
	Status         int8       `gorm:"column:status;not null;default:0" json:"status"`
	CreatedAt      *time.Time `gorm:"column:created_at" json:"created_at"`
	UpdatedAt      *time.Time `gorm:"column:updated_at" json:"updated_at"`
}

func (RebateRecord) TableName() string {
	return "rebate_records"
}

// WalletChangeLog 钱包变动日志
type WalletChangeLog struct {
	ID            int64      `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	UserID        int64      `gorm:"column:user_id;not null;default:0" json:"user_id"`
	ChangeType    string     `gorm:"column:change_type;not null;default:''" json:"change_type"`
	Amount        int64      `gorm:"column:amount;not null;default:0" json:"amount"`
	BalanceAfter  int64      `gorm:"column:balance_after;not null;default:0" json:"balance_after"`
	OrderID       int64      `gorm:"column:order_id;not null;default:0" json:"order_id"`
	Description   string     `gorm:"column:description;not null;default:''" json:"description"`
	Hash          string     `gorm:"column:hash;not null;default:''" json:"hash"`
	CreatedAt     *time.Time `gorm:"column:created_at" json:"created_at"`
	UpdatedAt     *time.Time `gorm:"column:updated_at" json:"updated_at"`
}

func (WalletChangeLog) TableName() string {
	return "wallet_change_logs"
}

// ============================================================
// 管理后台扩展模块
// ============================================================

// PunishmentTemplate 处罚模板
type PunishmentTemplate struct {
	ID            int64      `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	Name          string     `gorm:"column:name;not null;default:''" json:"name"`
	Type          int8       `gorm:"column:type;not null;default:0" json:"type"`
	Content       string     `gorm:"column:content;not null;default:''" json:"content"`
	DurationDays  int        `gorm:"column:duration_days;not null;default:0" json:"duration_days"`
	Status        int8       `gorm:"column:status;not null;default:0" json:"status"`
	CreatedAt     *time.Time `gorm:"column:created_at" json:"created_at"`
	UpdatedAt     *time.Time `gorm:"column:updated_at" json:"updated_at"`
}

func (PunishmentTemplate) TableName() string {
	return "punishment_templates"
}

// ============================================================
// UX扩展模块
// ============================================================

// UserFeedback 用户反馈
type UserFeedback struct {
	ID         int64          `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	UserID     int64          `gorm:"column:user_id;not null;default:0" json:"user_id"`
	Type       string         `gorm:"column:type;not null;default:''" json:"type"`
	Content    string         `gorm:"column:content;not null;default:''" json:"content"`
	Images     datatypes.JSON `gorm:"column:images" json:"images"`
	Contact    string         `gorm:"column:contact;not null;default:''" json:"contact"`
	Status     int8           `gorm:"column:status;not null;default:0" json:"status"`
	Reply      string         `gorm:"column:reply;not null;default:''" json:"reply"`
	HandlerUID int64          `gorm:"column:handler_uid;not null;default:0" json:"handler_uid"`
	CreatedAt  *time.Time     `gorm:"column:created_at" json:"created_at"`
	UpdatedAt  *time.Time     `gorm:"column:updated_at" json:"updated_at"`
}

func (UserFeedback) TableName() string {
	return "user_feedbacks"
}

// UserBlocklist 用户屏蔽列表
type UserBlocklist struct {
	ID         int64      `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	UserID     int64      `gorm:"column:user_id;not null;default:0" json:"user_id"`
	BlockedUID int64      `gorm:"column:blocked_uid;not null;default:0" json:"blocked_uid"`
	CreatedAt  *time.Time `gorm:"column:created_at" json:"created_at"`
	UpdatedAt  *time.Time `gorm:"column:updated_at" json:"updated_at"`
}

func (UserBlocklist) TableName() string {
	return "user_blocklist"
}

// UserNotificationSetting 用户通知设置
type UserNotificationSetting struct {
	ID                int64      `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	UserID            int64      `gorm:"column:user_id;not null;default:0" json:"user_id"`
	OrderNotify       int8       `gorm:"column:order_notify;not null;default:0" json:"order_notify"`
	AfterSaleNotify   int8       `gorm:"column:after_sale_notify;not null;default:0" json:"after_sale_notify"`
	MarketingNotify   int8       `gorm:"column:marketing_notify;not null;default:0" json:"marketing_notify"`
	SystemNotify      int8       `gorm:"column:system_notify;not null;default:0" json:"system_notify"`
	CreatedAt         *time.Time `gorm:"column:created_at" json:"created_at"`
	UpdatedAt         *time.Time `gorm:"column:updated_at" json:"updated_at"`
}

func (UserNotificationSetting) TableName() string {
	return "user_notification_settings"
}

// ============================================================
// 营销扩展模块
// ============================================================

// ActivityPopup 活动弹窗
type ActivityPopup struct {
	ID             int64      `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	ClubID         int64      `gorm:"column:club_id;not null;default:0" json:"club_id"`
	Title          string     `gorm:"column:title;not null;default:''" json:"title"`
	ImageURL       string     `gorm:"column:image_url;not null;default:''" json:"image_url"`
	LinkURL        string     `gorm:"column:link_url;not null;default:''" json:"link_url"`
	StartTime      *time.Time `gorm:"column:start_time" json:"start_time"`
	EndTime        *time.Time `gorm:"column:end_time" json:"end_time"`
	IsInternalOnly int8       `gorm:"column:is_internal_only;not null;default:0" json:"is_internal_only"`
	Status         int8       `gorm:"column:status;not null;default:0" json:"status"`
	CreatedAt      *time.Time `gorm:"column:created_at" json:"created_at"`
	UpdatedAt      *time.Time `gorm:"column:updated_at" json:"updated_at"`
}

func (ActivityPopup) TableName() string {
	return "activity_popups"
}

// FestivalTemplate 节日模板
type FestivalTemplate struct {
	ID           int64      `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	Name         string     `gorm:"column:name;not null;default:''" json:"name"`
	Content      string     `gorm:"column:content;not null;default:''" json:"content"`
	FestivalDate string     `gorm:"column:festival_date;not null;default:''" json:"festival_date"`
	Status       int8       `gorm:"column:status;not null;default:0" json:"status"`
	CreatedAt    *time.Time `gorm:"column:created_at" json:"created_at"`
	UpdatedAt    *time.Time `gorm:"column:updated_at" json:"updated_at"`
}

func (FestivalTemplate) TableName() string {
	return "festival_templates"
}

// PromoChannel 推广渠道
type PromoChannel struct {
	ID            int64      `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	ChannelName   string     `gorm:"column:channel_name;not null;default:''" json:"channel_name"`
	ChannelCode   string     `gorm:"column:channel_code;not null;default:''" json:"channel_code"`
	ShareLink     string     `gorm:"column:share_link;not null;default:''" json:"share_link"`
	Registrations int        `gorm:"column:registrations;not null;default:0" json:"registrations"`
	Orders        int        `gorm:"column:orders;not null;default:0" json:"orders"`
	CreatedAt     *time.Time `gorm:"column:created_at" json:"created_at"`
	UpdatedAt     *time.Time `gorm:"column:updated_at" json:"updated_at"`
}

func (PromoChannel) TableName() string {
	return "promo_channels"
}
