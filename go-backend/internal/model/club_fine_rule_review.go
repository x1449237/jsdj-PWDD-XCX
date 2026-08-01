package model

import "time"

// ClubFineRuleReview 俱乐部罚款规则平台备案审核表
type ClubFineRuleReview struct {
	ID           int64      `gorm:"primaryKey;autoIncrement;column:id" json:"id"`
	RuleID       int64      `gorm:"column:rule_id;index:idx_rule_id;not null;default:0" json:"rule_id"`        // 罚款规则ID
	ClubID       int64      `gorm:"column:club_id;index:idx_club_id;not null;default:0" json:"club_id"`        // 俱乐部ID
	ReviewStatus string     `gorm:"column:review_status;index:idx_review_status;size:32;not null;default:'pending'" json:"review_status"` // 审核状态 pending/approved/revoked
	ReviewerID   int64      `gorm:"column:reviewer_id;not null;default:0" json:"reviewer_id"`                  // 审核人ID
	ReviewNote  string     `gorm:"column:review_note;size:255;not null;default:''" json:"review_note"`        // 审核备注
	ReviewedAt  *time.Time `gorm:"column:reviewed_at" json:"reviewed_at"`                                    // 审核时间
	CreatedAt   *time.Time `gorm:"column:created_at;index:idx_created_at" json:"created_at"`
}

// TableName 指定表名
func (ClubFineRuleReview) TableName() string {
	return "club_fine_rule_reviews"
}

// 罚款规则审核状态常量
const (
	ClubFineRuleReviewPending  = "pending"  // 待审核
	ClubFineRuleReviewApproved = "approved" // 已备案
	ClubFineRuleReviewRevoked  = "revoked"  // 已下架
)
