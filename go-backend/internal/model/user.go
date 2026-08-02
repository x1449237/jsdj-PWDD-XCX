package model

import "time"

// User 用户表模型
type User struct {
	ID                 int64      `gorm:"primaryKey;autoIncrement;column:id" json:"id"`                                            // 主键ID
	OpenID             string     `gorm:"column:openid;uniqueIndex:uk_openid;size:64;not null;default:''" json:"openid"`           // 微信openid
	UnionID            string     `gorm:"column:unionid;size:64;not null;default:''" json:"union_id"`                              // 微信unionid
	Phone              string     `gorm:"column:phone;size:20;index:idx_phone;not null;default:''" json:"phone"`                   // 手机号
	Nickname           string     `gorm:"column:nickname;size:64;not null;default:''" json:"nickname"`                             // 昵称
	Avatar             string     `gorm:"column:avatar;size:512;not null;default:''" json:"avatar"`                                // 头像URL
	Role               int8       `gorm:"column:role;index:idx_role;not null;default:1" json:"role"`                               // 角色 1=客户 2=打手 4=分销商 8=派单员(位运算)
	InviteCode         string     `gorm:"column:invite_code;index:idx_invite_code;size:64;not null;default:''" json:"invite_code"` // 绑定的邀请码
	ClubID             int64      `gorm:"column:club_id;index:idx_club_id;not null;default:0" json:"club_id"`                      // 所属俱乐部ID
	CreditScore        int        `gorm:"column:credit_score;not null;default:100" json:"credit_score"`                            // 信用分 初始100
	IsRealname         int8       `gorm:"column:is_realname;not null;default:0" json:"is_realname"`                                // 是否已实名 0否 1是
	IsMinor            int8       `gorm:"column:is_minor;not null;default:0" json:"is_minor"`                                      // 是否未成年 0否 1是
	RealName           string     `gorm:"column:real_name;size:32;not null;default:''" json:"real_name"`                           // 真实姓名
	IDCard             string     `gorm:"column:id_card;size:18;not null;default:''" json:"id_card"`                               // 身份证号
	IsPhoneAbandoned   int8       `gorm:"column:is_phone_abandoned;not null;default:0" json:"is_phone_abandoned"`                  // 手机号是否被二次放号回收 0否 1是
	Status             int8       `gorm:"column:status;index:idx_status;not null;default:1" json:"status"`                         // 状态 1=正常 0=封禁
	Balance            int64      `gorm:"column:balance;not null;default:0" json:"balance"`                                        // 余额(分)
	Points             int        `gorm:"column:points;not null;default:0" json:"points"`                                          // 积分
	DarkMode           int8       `gorm:"column:dark_mode;not null;default:0" json:"dark_mode"`                                    // 深色模式
	AgreementDismissed int8       `gorm:"column:agreement_dismissed;not null;default:0" json:"agreement_dismissed"`                // 协议永久关闭
	CreatedAt          *time.Time `gorm:"column:created_at;index:idx_created_at" json:"created_at"`                                // 创建时间
	UpdatedAt          *time.Time `gorm:"column:updated_at" json:"updated_at"`                                                     // 更新时间
}

// TableName 指定表名
func (User) TableName() string {
	return "users"
}

// 角色位运算常量
const (
	RoleCustomer    int8 = 1 // 客户
	RolePlayer      int8 = 2 // 打手
	RoleDistributor int8 = 4 // 分销商
	RoleDispatcher  int8 = 8 // 派单员
)

// HasRole 判断用户是否拥有指定角色(位运算)
func (u *User) HasRole(role int8) bool {
	return u.Role&role != 0
}
