package model

import "time"

// Admin 管理员表模型
type Admin struct {
	ID          int64      `gorm:"primaryKey;autoIncrement;column:id" json:"id"`
	Username    string     `gorm:"column:username;uniqueIndex:uk_username;size:64;not null;default:''" json:"username"` // 管理员账号
	Password    string     `gorm:"column:password;size:255;not null;default:''" json:"password"`               // 密码(bcrypt加密)
	Nickname    string     `gorm:"column:nickname;size:64;not null;default:''" json:"nickname"`                // 昵称
	Role        int8       `gorm:"column:role;index:idx_role;not null;default:2" json:"role"`                  // 角色 1=超级管理员 2=运营 4=财务 8=风控(位运算)
	Email       string     `gorm:"column:email;size:128;not null;default:''" json:"email"`                     // 绑定邮箱
	Phone       string     `gorm:"column:phone;size:20;not null;default:''" json:"phone"`                      // 手机号
	IsInit      int8       `gorm:"column:is_init;not null;default:0" json:"is_init"`                           // 是否完成首次初始化 0否 1是
	Status      int8       `gorm:"column:status;index:idx_status;not null;default:1" json:"status"`            // 状态 1正常 0封禁
	LastLoginAt *time.Time `gorm:"column:last_login_at" json:"last_login_at"`                                 // 最后登录时间
	LastLoginIP string     `gorm:"column:last_login_ip;size:64;not null;default:''" json:"last_login_ip"`     // 最后登录IP
	CreatedAt   *time.Time `gorm:"column:created_at;index:idx_created_at" json:"created_at"`
	UpdatedAt   *time.Time `gorm:"column:updated_at" json:"updated_at"`
}

// TableName 指定表名
func (Admin) TableName() string {
	return "admins"
}

// 管理员角色位运算常量
const (
	AdminRoleSuper  int8 = 1 // 超级管理员
	AdminRoleOps    int8 = 2 // 运营
	AdminRoleFinance int8 = 4 // 财务
	AdminRoleRisk   int8 = 8 // 风控
)

// AdminPasswordHistory 管理员密码历史表
type AdminPasswordHistory struct {
	ID           int64      `gorm:"primaryKey;autoIncrement;column:id" json:"id"`
	AdminID      int64      `gorm:"column:admin_id;index:idx_admin_id;not null;default:0" json:"admin_id"` // 管理员ID
	PasswordHash string     `gorm:"column:password_hash;size:255;not null;default:''" json:"password_hash"` // 历史密码hash
	CreatedAt    *time.Time `gorm:"column:created_at;index:idx_created_at" json:"created_at"`
}

// TableName 指定表名
func (AdminPasswordHistory) TableName() string {
	return "admin_password_history"
}

// AdminWebauthn 管理员WebAuthn通行密钥表
type AdminWebauthn struct {
	ID           int64      `gorm:"primaryKey;autoIncrement;column:id" json:"id"`
	AdminID      int64      `gorm:"column:admin_id;index:idx_admin_id;not null;default:0" json:"admin_id"`      // 管理员ID
	CredentialID string     `gorm:"column:credential_id;uniqueIndex:uk_credential_id;size:255;not null;default:''" json:"credential_id"` // 凭证ID
	PublicKey    string     `gorm:"column:public_key;type:text" json:"public_key"`                             // 公钥
	DeviceInfo   string     `gorm:"column:device_info;size:255;not null;default:''" json:"device_info"`        // 设备信息
	CreatedAt    *time.Time `gorm:"column:created_at" json:"created_at"`
}

// TableName 指定表名
func (AdminWebauthn) TableName() string {
	return "admin_webauthn"
}
