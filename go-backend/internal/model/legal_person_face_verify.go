package model

import "time"

// 法人活体认证状态常量
const (
	LegalPersonFaceStatusPending = "pending" // 待验证
	LegalPersonFaceStatusPassed  = "passed"  // 通过
	LegalPersonFaceStatusFailed  = "failed"  // 失败
	LegalPersonFaceStatusExpired = "expired" // 已过期
)

// LegalPersonFaceVerify 法人活体认证表
type LegalPersonFaceVerify struct {
	ID              int64      `gorm:"primaryKey;autoIncrement;column:id" json:"id"`
	ClubID          int64      `gorm:"column:club_id;index:idx_club_id;not null;default:0" json:"club_id"` // 俱乐部ID
	LegalPersonName string     `gorm:"column:legal_person_name;size:64;not null;default:''" json:"legal_person_name"` // 法人姓名
	LegalPersonIDCard string   `gorm:"column:legal_person_id_card;size:18;not null;default:''" json:"legal_person_id_card"` // 法人身份证号
	VerifyToken     string     `gorm:"column:verify_token;size:255;not null;default:''" json:"verify_token"` // 活体认证 token
	VerifyAt        *time.Time `gorm:"column:verify_at" json:"verify_at"`                                     // 验证时间
	ExpireAt        *time.Time `gorm:"column:expire_at;index:idx_expire_at" json:"expire_at"`                // 过期时间(verify_at + 72h)
	Status          string     `gorm:"column:status;size:32;index:idx_status;not null;default:'pending'" json:"status"` // 状态 pending/passed/failed/expired
	CreatedAt       *time.Time `gorm:"column:created_at" json:"created_at"`
	UpdatedAt       *time.Time `gorm:"column:updated_at" json:"updated_at"`
}

// TableName 指定表名
func (LegalPersonFaceVerify) TableName() string {
	return "legal_person_face_verifies"
}
