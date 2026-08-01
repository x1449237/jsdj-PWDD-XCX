package model

import "time"

// EnterpriseRegistration 企业入驻申请表模型
type EnterpriseRegistration struct {
	ID                  int64      `gorm:"primaryKey;autoIncrement;column:id" json:"id"`
	ClubID              int64      `gorm:"column:club_id;index:idx_club_id;not null;default:0" json:"club_id"`        // 俱乐部ID
	BusinessLicenseURL  string     `gorm:"column:business_license_url;size:512;not null;default:''" json:"business_license_url"` // 营业执照URL
	LegalPersonName     string     `gorm:"column:legal_person_name;size:64;not null;default:''" json:"legal_person_name"` // 法人姓名
	LegalPersonIDCard   string     `gorm:"column:legal_person_id_card;size:18;not null;default:''" json:"legal_person_id_card"` // 法人身份证号
	LegalPersonIDFront  string     `gorm:"column:legal_person_id_front;size:512;not null;default:''" json:"legal_person_id_front"` // 法人身份证正面URL
	LegalPersonIDBack   string     `gorm:"column:legal_person_id_back;size:512;not null;default:''" json:"legal_person_id_back"` // 法人身份证反面URL
	ContactPhone        string     `gorm:"column:contact_phone;size:20;not null;default:''" json:"contact_phone"`   // 联系电话
	ContactEmail        string     `gorm:"column:contact_email;size:128;not null;default:''" json:"contact_email"`   // 联系邮箱
	Address             string     `gorm:"column:address;size:255;not null;default:''" json:"address"`              // 企业地址
	BankName            string     `gorm:"column:bank_name;size:64;not null;default:''" json:"bank_name"`           // 开户行
	BankAccount         string     `gorm:"column:bank_account;size:64;not null;default:''" json:"bank_account"`     // 银行账号
	ElectronicLicenseURL string    `gorm:"column:electronic_license_url;size:512;not null;default:''" json:"electronic_license_url"` // 电子营业执照URL
	AuthLetterURL       string     `gorm:"column:auth_letter_url;size:512;not null;default:''" json:"auth_letter_url"` // 授权书URL
	AgentIDCardFront    string     `gorm:"column:agent_id_card_front;size:512;not null;default:''" json:"agent_id_card_front"` // 代理人身份证正面URL
	AgentIDCardBack     string     `gorm:"column:agent_id_card_back;size:512;not null;default:''" json:"agent_id_card_back"` // 代理人身份证反面URL
	Status              int8       `gorm:"column:status;index:idx_status;not null;default:0" json:"status"`          // 状态 0待审核 1通过 2驳回
	ReviewerID          int64      `gorm:"column:reviewer_id;not null;default:0" json:"reviewer_id"`                // 审核人ID
	ReviewedAt          *time.Time `gorm:"column:reviewed_at" json:"reviewed_at"`                                   // 审核时间
	RejectReason        string     `gorm:"column:reject_reason;size:255;not null;default:''" json:"reject_reason"`  // 驳回原因
	CreatedAt           *time.Time `gorm:"column:created_at;index:idx_created_at" json:"created_at"`
	UpdatedAt           *time.Time `gorm:"column:updated_at" json:"updated_at"`
}

// TableName 指定表名
func (EnterpriseRegistration) TableName() string {
	return "enterprise_registrations"
}

// 入驻审核状态常量
const (
	RegistrationStatusPending int8 = 0 // 待审核
	RegistrationStatusApproved int8 = 1 // 通过
	RegistrationStatusRejected int8 = 2 // 驳回
)

// PersonalRegistration 个人入驻申请表
type PersonalRegistration struct {
	ID               int64      `gorm:"primaryKey;autoIncrement;column:id" json:"id"`
	ClubID           int64      `gorm:"column:club_id;index:idx_club_id;not null;default:0" json:"club_id"`        // 俱乐部ID
	RealName         string     `gorm:"column:real_name;size:32;not null;default:''" json:"real_name"`            // 真实姓名
	IDCard           string     `gorm:"column:id_card;size:18;not null;default:''" json:"id_card"`                // 身份证号
	Phone            string     `gorm:"column:phone;size:20;not null;default:''" json:"phone"`                    // 手机号
	IDCardFront      string     `gorm:"column:id_card_front;size:512;not null;default:''" json:"id_card_front"`   // 身份证正面URL
	IDCardBack       string     `gorm:"column:id_card_back;size:512;not null;default:''" json:"id_card_back"`     // 身份证反面URL
	HandheldIDCard   string     `gorm:"column:handheld_id_card;size:512;not null;default:''" json:"handheld_id_card"` // 手持身份证URL
	BankCard         string     `gorm:"column:bank_card;size:32;not null;default:''" json:"bank_card"`            // 银行卡号
	BankName         string     `gorm:"column:bank_name;size:64;not null;default:''" json:"bank_name"`            // 开户行
	BankPhone        string     `gorm:"column:bank_phone;size:20;not null;default:''" json:"bank_phone"`          // 银行预留手机号
	FaceVerifyStatus int8       `gorm:"column:face_verify_status;not null;default:0" json:"face_verify_status"`   // 活体检测状态 0未检测 1通过 2失败
	SelfDeclarationURL string   `gorm:"column:self_declaration_url;size:512;not null;default:''" json:"self_declaration_url"` // 自我声明URL
	Status           int8       `gorm:"column:status;index:idx_status;not null;default:0" json:"status"`          // 状态 0待审核 1通过 2驳回
	ReviewerID       int64      `gorm:"column:reviewer_id;not null;default:0" json:"reviewer_id"`                 // 审核人ID
	ReviewedAt       *time.Time `gorm:"column:reviewed_at" json:"reviewed_at"`                                    // 审核时间
	RejectReason     string     `gorm:"column:reject_reason;size:255;not null;default:''" json:"reject_reason"`   // 驳回原因
	CreatedAt        *time.Time `gorm:"column:created_at;index:idx_created_at" json:"created_at"`
	UpdatedAt        *time.Time `gorm:"column:updated_at" json:"updated_at"`
}

// TableName 指定表名
func (PersonalRegistration) TableName() string {
	return "personal_registrations"
}
