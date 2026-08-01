package model

import "time"

// RiskUser 风险用户表模型
type RiskUser struct {
	ID        int64      `gorm:"primaryKey;autoIncrement;column:id" json:"id"`
	UserID    int64      `gorm:"column:user_id;index:idx_user_id;not null;default:0" json:"user_id"`     // 用户ID
	RiskLevel string     `gorm:"column:risk_level;index:idx_risk_level;size:32;not null;default:'low'" json:"risk_level"` // 风险等级 low/medium/high/critical
	RiskType  string     `gorm:"column:risk_type;index:idx_risk_type;size:32;not null;default:''" json:"risk_type"` // 风险类型
	MarkedAt  *time.Time `gorm:"column:marked_at" json:"marked_at"`                                     // 标记时间
	CreatedAt *time.Time `gorm:"column:created_at" json:"created_at"`
	UpdatedAt *time.Time `gorm:"column:updated_at" json:"updated_at"`
}

// TableName 指定表名
func (RiskUser) TableName() string {
	return "risk_users"
}

// 风险等级常量
const (
	RiskLevelLow      = "low"
	RiskLevelMedium   = "medium"
	RiskLevelHigh     = "high"
	RiskLevelCritical = "critical"
)

// AiRiskAlert AI风险预警表
type AiRiskAlert struct {
	ID          int64      `gorm:"primaryKey;autoIncrement;column:id" json:"id"`
	AlertType   string     `gorm:"column:alert_type;index:idx_alert_type;size:32;not null;default:''" json:"alert_type"` // 预警类型
	TargetType  string     `gorm:"column:target_type;index:idx_target_type;size:32;not null;default:''" json:"target_type"` // 目标类型 user/order/club
	TargetID    int64      `gorm:"column:target_id;index:idx_target_id;not null;default:0" json:"target_id"` // 目标ID
	Description string     `gorm:"column:description;size:255;not null;default:''" json:"description"`     // 风险描述
	Level       int8       `gorm:"column:level;not null;default:1" json:"level"`                           // 风险等级 1低 2中 3高 4最高
	Status      int8       `gorm:"column:status;index:idx_status;not null;default:0" json:"status"`        // 状态 0待处理 1已处理 2已忽略
	CreatedAt   *time.Time `gorm:"column:created_at;index:idx_created_at" json:"created_at"`
	UpdatedAt   *time.Time `gorm:"column:updated_at" json:"updated_at"`
}

// TableName 指定表名
func (AiRiskAlert) TableName() string {
	return "ai_risk_alerts"
}

// SensitiveWord 敏感词库表
type SensitiveWord struct {
	ID        int64      `gorm:"primaryKey;autoIncrement;column:id" json:"id"`
	Word      string     `gorm:"column:word;size:128;not null;default:''" json:"word"`                   // 敏感词
	Category  string     `gorm:"column:category;index:idx_category;size:32;not null;default:''" json:"category"` // 分类 fraud/boosting/gambling/etc
	Enabled   int8       `gorm:"column:enabled;index:idx_enabled;not null;default:1" json:"enabled"`     // 是否启用 0否 1是
	CreatedAt *time.Time `gorm:"column:created_at" json:"created_at"`
	UpdatedAt *time.Time `gorm:"column:updated_at" json:"updated_at"`
}

// TableName 指定表名
func (SensitiveWord) TableName() string {
	return "sensitive_words"
}

// 敏感词分类常量
const (
	SensitiveCategoryFraud    = "fraud"    // 诈骗
	SensitiveCategoryBoosting = "boosting" // 代练
	SensitiveCategoryGambling = "gambling" // 赌博
)
