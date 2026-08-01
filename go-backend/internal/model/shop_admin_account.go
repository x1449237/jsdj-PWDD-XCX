package model

import "time"

// ShopAdminAccount 内置管理端账号表模型
type ShopAdminAccount struct {
	ID          int64      `gorm:"primaryKey;autoIncrement;column:id" json:"id"`
	Username    string     `gorm:"column:username;uniqueIndex:uk_username;size:64;not null;default:''" json:"username"` // 账号 wxadmin_缩写序号
	Password    string     `gorm:"column:password;size:255;not null;default:''" json:"password"`              // 密码(bcrypt加密)
	ClubID      int64      `gorm:"column:club_id;index:idx_club_id;not null;default:0" json:"club_id"`         // 俱乐部ID
	Role        int8       `gorm:"column:role;not null;default:2" json:"role"`                                // 角色 1=创始人 2=管理员
	RealName    string     `gorm:"column:real_name;size:32;not null;default:''" json:"real_name"`             // 真实姓名
	Phone       string     `gorm:"column:phone;size:20;not null;default:''" json:"phone"`                     // 手机号
	Status      int8       `gorm:"column:status;index:idx_status;not null;default:1" json:"status"`           // 状态 1正常 0封禁
	LastLoginAt *time.Time `gorm:"column:last_login_at" json:"last_login_at"`                                // 最后登录时间
	LastLoginIP string     `gorm:"column:last_login_ip;size:64;not null;default:''" json:"last_login_ip"`    // 最后登录IP
	CreatedAt   *time.Time `gorm:"column:created_at;index:idx_created_at" json:"created_at"`
	UpdatedAt   *time.Time `gorm:"column:updated_at" json:"updated_at"`
}

// TableName 指定表名
func (ShopAdminAccount) TableName() string {
	return "shop_admin_accounts"
}

// 内置管理端角色常量
const (
	ShopAdminRoleFounder int8 = 1 // 创始人
	ShopAdminRoleAdmin   int8 = 2 // 管理员
)
