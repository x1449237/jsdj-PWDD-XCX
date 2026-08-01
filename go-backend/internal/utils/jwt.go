package utils

import (
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

// weakSecretBlacklist 默认/弱密钥黑名单(禁止使用)
var weakSecretBlacklist = map[string]bool{
	"":                true,
	"secret":          true,
	"your-secret-key": true,
	"change-me":       true,
	"jwt-secret":      true,
	"jwt_secret":      true,
	"default":         true,
	"test":            true,
	"123456":          true,
}

// JWT 用户类型常量
const (
	JWTUserTypeUser      = "user"      // 普通用户
	JWTUserTypeAdmin     = "admin"     // 平台管理员
	JWTUserTypeShopAdmin = "shop_admin" // 内置管理端
)

// Claims 自定义 JWT 声明
type Claims struct {
	UserID   int64  `json:"user_id"`            // 用户/管理员ID
	Role     int8   `json:"role"`               // 角色(位运算)
	UserType string `json:"user_type"`          // 用户类型
	ClubID   int64  `json:"club_id,omitempty"`  // 俱乐部ID(内置管理端使用)
	jwt.RegisteredClaims
}

// JWTManager JWT 管理器
type JWTManager struct {
	secret          string
	expireHours     int        // 普通用户有效期(小时)
	adminExpireHours int       // 管理员有效期(小时)
	issuer          string
}

// NewJWTManager 创建 JWT 管理器
// 安全修复:
// 1. 优先从环境变量 JWT_SECRET 读取密钥(避免硬编码在配置文件/提交到 git)
// 2. 启动期强制校验:密钥长度 >= 32 字节,且不在弱密钥黑名单中
// 3. 不合规直接 panic(fail-fast,密钥不合规不应启动服务)
func NewJWTManager(secret string, expireHours, adminExpireHours int, issuer string) *JWTManager {
	// 优先环境变量(生产环境密钥不应落入配置文件/git)
	if envSecret := os.Getenv("JWT_SECRET"); envSecret != "" {
		secret = envSecret
	}
	// 启动期密钥校验
	if weakSecretBlacklist[secret] {
		panic("JWT 密钥不合规: 禁止使用默认/弱密钥,请通过环境变量 JWT_SECRET 配置高强度密钥(>=32字节)")
	}
	if len(secret) < 32 {
		panic(fmt.Sprintf("JWT 密钥不合规: 长度仅 %d 字节,要求至少 32 字节,请通过环境变量 JWT_SECRET 配置高强度随机密钥", len(secret)))
	}
	return &JWTManager{
		secret:           secret,
		expireHours:      expireHours,
		adminExpireHours: adminExpireHours,
		issuer:           issuer,
	}
}

// GenerateToken 生成 JWT token
// userType 决定有效期: admin/shop_admin 使用 adminExpireHours, 其他使用 expireHours
func (m *JWTManager) GenerateToken(userID int64, role int8, userType string, clubID int64) (string, error) {
	hours := m.expireHours
	if userType == JWTUserTypeAdmin || userType == JWTUserTypeShopAdmin {
		hours = m.adminExpireHours
	}

	now := time.Now()
	claims := Claims{
		UserID:   userID,
		Role:     role,
		UserType: userType,
		ClubID:   clubID,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    m.issuer,
			Subject:   userType,
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(time.Duration(hours) * time.Hour)),
			NotBefore: jwt.NewNumericDate(now),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(m.secret))
}

// ParseToken 解析并校验 JWT token
func (m *JWTManager) ParseToken(tokenString string) (*Claims, error) {
	claims := &Claims{}
	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("非预期的签名方法")
		}
		return []byte(m.secret), nil
	})
	if err != nil {
		return nil, err
	}
	if !token.Valid {
		return nil, errors.New("无效的 token")
	}
	return claims, nil
}

// 上下文键
const (
	// ContextKeyClaims JWT claims 在 gin.Context 中的键
	ContextKeyClaims = "jwt_claims"
	// ContextKeyUserID 用户ID 在 gin.Context 中的键
	ContextKeyUserID = "user_id"
	// ContextKeyUserRole 用户角色在 gin.Context 中的键
	ContextKeyUserRole = "user_role"
	// ContextKeyUserType 用户类型在 gin.Context 中的键
	ContextKeyUserType = "user_type"
	// ContextKeyClubID 俱乐部ID 在 gin.Context 中的键
	ContextKeyClubID = "club_id"
)

// GetClaimsFromContext 从 gin.Context 获取 claims
func GetClaimsFromContext(c *gin.Context) (*Claims, bool) {
	val, exists := c.Get(ContextKeyClaims)
	if !exists {
		return nil, false
	}
	claims, ok := val.(*Claims)
	return claims, ok
}

// GetCurrentUserID 从 gin.Context 获取当前用户ID
func GetCurrentUserID(c *gin.Context) int64 {
	if claims, ok := GetClaimsFromContext(c); ok {
		return claims.UserID
	}
	if uid, exists := c.Get(ContextKeyUserID); exists {
		if id, ok := uid.(int64); ok {
			return id
		}
	}
	return 0
}
