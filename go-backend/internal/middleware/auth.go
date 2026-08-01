package middleware

import (
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/jisan/e-sports-platform/internal/utils"
)

// AuthMiddleware JWT 认证中间件
// 区分普通用户(user)/平台管理员(admin)/内置管理端(shop_admin)
type AuthMiddleware struct {
	jwt *utils.JWTManager
}

// NewAuthMiddleware 创建认证中间件
func NewAuthMiddleware(jwt *utils.JWTManager) *AuthMiddleware {
	return &AuthMiddleware{jwt: jwt}
}

// extractToken 从请求头提取 token，支持 Authorization: Bearer <token> 与 query 参数 token
func extractToken(c *gin.Context) string {
	auth := c.GetHeader("Authorization")
	if auth != "" {
		// 优先 Bearer token
		if strings.HasPrefix(auth, "Bearer ") {
			return strings.TrimPrefix(auth, "Bearer ")
		}
		return auth
	}
	// 兼容 query 参数(WebSocket 等场景)
	if token := c.Query("token"); token != "" {
		return token
	}
	return ""
}

// setClaims 将 claims 注入 gin.Context
func setClaims(c *gin.Context, claims *utils.Claims) {
	c.Set(utils.ContextKeyClaims, claims)
	c.Set(utils.ContextKeyUserID, claims.UserID)
	c.Set(utils.ContextKeyUserRole, claims.Role)
	c.Set(utils.ContextKeyUserType, claims.UserType)
	c.Set(utils.ContextKeyClubID, claims.ClubID)
}

// AuthRequired 普通用户认证(任何已登录用户)
func (m *AuthMiddleware) AuthRequired() gin.HandlerFunc {
	return func(c *gin.Context) {
		token := extractToken(c)
		if token == "" {
			utils.Fail(c, utils.CodeUnauthorized, "未登录或登录已过期")
			c.Abort()
			return
		}

		claims, err := m.jwt.ParseToken(token)
		if err != nil {
			utils.Fail(c, utils.CodeUnauthorized, "登录凭证无效，请重新登录")
			c.Abort()
			return
		}

		setClaims(c, claims)
		c.Next()
	}
}

// AdminRequired 平台管理员认证(user_type 必须为 admin)
func (m *AuthMiddleware) AdminRequired() gin.HandlerFunc {
	return func(c *gin.Context) {
		token := extractToken(c)
		if token == "" {
			utils.Fail(c, utils.CodeUnauthorized, "管理员未登录")
			c.Abort()
			return
		}

		claims, err := m.jwt.ParseToken(token)
		if err != nil {
			utils.Fail(c, utils.CodeUnauthorized, "登录凭证无效，请重新登录")
			c.Abort()
			return
		}

		if claims.UserType != utils.JWTUserTypeAdmin {
			utils.Fail(c, utils.CodeForbidden, "无管理员权限")
			c.Abort()
			return
		}

		setClaims(c, claims)
		c.Next()
	}
}

// ShopAdminRequired 内置管理端认证(user_type 必须为 shop_admin，且需携带 club_id)
func (m *AuthMiddleware) ShopAdminRequired() gin.HandlerFunc {
	return func(c *gin.Context) {
		token := extractToken(c)
		if token == "" {
			utils.Fail(c, utils.CodeUnauthorized, "管理端未登录")
			c.Abort()
			return
		}

		claims, err := m.jwt.ParseToken(token)
		if err != nil {
			utils.Fail(c, utils.CodeUnauthorized, "登录凭证无效，请重新登录")
			c.Abort()
			return
		}

		if claims.UserType != utils.JWTUserTypeShopAdmin {
			utils.Fail(c, utils.CodeForbidden, "无内置管理端权限")
			c.Abort()
			return
		}

		if claims.ClubID == 0 {
			utils.Fail(c, utils.CodeForbidden, "未绑定俱乐部")
			c.Abort()
			return
		}

		setClaims(c, claims)
		c.Next()
	}
}

// SuperAdminRequired 超级管理员认证(角色位含超级管理员 1)
func (m *AuthMiddleware) SuperAdminRequired() gin.HandlerFunc {
	return func(c *gin.Context) {
		token := extractToken(c)
		if token == "" {
			utils.Fail(c, utils.CodeUnauthorized, "管理员未登录")
			c.Abort()
			return
		}

		claims, err := m.jwt.ParseToken(token)
		if err != nil {
			utils.Fail(c, utils.CodeUnauthorized, "登录凭证无效，请重新登录")
			c.Abort()
			return
		}

		if claims.UserType != utils.JWTUserTypeAdmin || claims.Role&1 == 0 {
			utils.Fail(c, utils.CodeForbidden, "无超级管理员权限")
			c.Abort()
			return
		}

		setClaims(c, claims)
		c.Next()
	}
}

// OptionalAuth 可选认证：未携带 token 也能访问，但若携带则解析注入
func (m *AuthMiddleware) OptionalAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		token := extractToken(c)
		if token == "" {
			c.Next()
			return
		}
		if claims, err := m.jwt.ParseToken(token); err == nil {
			setClaims(c, claims)
		}
		c.Next()
	}
}
