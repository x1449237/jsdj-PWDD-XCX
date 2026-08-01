package model

import "time"

// JoinApplication 入会申请表模型
type JoinApplication struct {
	ID           int64      `gorm:"primaryKey;autoIncrement;column:id" json:"id"`
	ClubID       int64      `gorm:"column:club_id;index:idx_club_id;not null;default:0" json:"club_id"`      // 俱乐部ID
	UserID       int64      `gorm:"column:user_id;index:idx_user_id;not null;default:0" json:"user_id"`      // 用户ID
	RealName     string     `gorm:"column:real_name;size:32;not null;default:''" json:"real_name"`           // 真实姓名
	GameAccount  string     `gorm:"column:game_account;size:128;not null;default:''" json:"game_account"`   // 游戏账号
	GameRegion   string     `gorm:"column:game_region;size:64;not null;default:''" json:"game_region"`       // 游戏大区
	GoodPosition string     `gorm:"column:good_position;size:64;not null;default:''" json:"good_position"`   // 擅长位置
	RankLevel    string     `gorm:"column:rank_level;size:32;not null;default:''" json:"rank_level"`         // 段位
	Intro        string     `gorm:"column:intro;size:500;not null;default:''" json:"intro"`                  // 自我介绍
	Status       string     `gorm:"column:status;index:idx_status;size:32;not null;default:'pending'" json:"status"` // 状态 pending/examining/approved/rejected
	CreatedAt    *time.Time `gorm:"column:created_at;index:idx_created_at" json:"created_at"`
	UpdatedAt    *time.Time `gorm:"column:updated_at" json:"updated_at"`
}

// TableName 指定表名
func (JoinApplication) TableName() string {
	return "join_applications"
}

// 入会申请状态常量
const (
	JoinStatusPending   = "pending"
	JoinStatusExamining = "examining"
	JoinStatusApproved  = "approved"
	JoinStatusRejected  = "rejected"
)

// ExamRecord 考核记录表
type ExamRecord struct {
	ID            int64      `gorm:"primaryKey;autoIncrement;column:id" json:"id"`
	ApplicationID int64      `gorm:"column:application_id;index:idx_application_id;not null;default:0" json:"application_id"` // 申请ID
	ExaminerID    int64      `gorm:"column:examiner_id;index:idx_examiner_id;not null;default:0" json:"examiner_id"` // 考核人ID
	Requirement   string     `gorm:"column:requirement;size:255;not null;default:''" json:"requirement"`     // 考核要求
	Result        string     `gorm:"column:result;index:idx_result;size:32;not null;default:''" json:"result"` // 考核结果 pass/fail
	Remark        string     `gorm:"column:remark;size:500;not null;default:''" json:"remark"`              // 备注
	VideoURL      string     `gorm:"column:video_url;size:512;not null;default:''" json:"video_url"`         // 考核视频URL
	CreatedAt     *time.Time `gorm:"column:created_at" json:"created_at"`
	UpdatedAt     *time.Time `gorm:"column:updated_at" json:"updated_at"`
}

// TableName 指定表名
func (ExamRecord) TableName() string {
	return "exam_records"
}

// 考核结果常量
const (
	ExamResultPass = "pass"
	ExamResultFail = "fail"
)

// ExamTemplate 考核模板表
type ExamTemplate struct {
	ID        int64      `gorm:"primaryKey;autoIncrement;column:id" json:"id"`
	ClubID    int64      `gorm:"column:club_id;index:idx_club_id;not null;default:0" json:"club_id"`       // 俱乐部ID
	Game      string     `gorm:"column:game;index:idx_game;size:64;not null;default:''" json:"game"`      // 游戏
	RankLevel string     `gorm:"column:rank_level;size:32;not null;default:''" json:"rank_level"`         // 段位
	Standard  string     `gorm:"column:standard;type:text" json:"standard"`                               // 考核标准
	CreatedAt *time.Time `gorm:"column:created_at" json:"created_at"`
	UpdatedAt *time.Time `gorm:"column:updated_at" json:"updated_at"`
}

// TableName 指定表名
func (ExamTemplate) TableName() string {
	return "exam_templates"
}
