package middleware

import (
	"github.com/gin-gonic/gin"
	"github.com/jisan/e-sports-platform/internal/utils"
)

// ClubScopeMiddleware 俱乐部范围限制中间件
// 内置管理端(shop_admin)只能访问本俱乐部数据，强制注入 club_id 过滤
type ClubScopeMiddleware struct{}

// NewClubScopeMiddleware 创建俱乐部范围限制中间件
func NewClubScopeMiddleware() *ClubScopeMiddleware {
	return &ClubScopeMiddleware{}
}

// ClubScope 强制校验内置管理端的俱乐部归属
// 要求请求方为 shop_admin 且携带 club_id；将 club_id 注入 context 供 service 层使用
func (m *ClubScopeMiddleware) ClubScope() gin.HandlerFunc {
	return func(c *gin.Context) {
		userType, _ := c.Get(utils.ContextKeyUserType)
		ut, _ := userType.(string)

		// 平台管理员不受俱乐部范围限制
		if ut == utils.JWTUserTypeAdmin {
			c.Next()
			return
		}

		// 内置管理端必须有 club_id
		clubIDAny, exists := c.Get(utils.ContextKeyClubID)
		if !exists {
			utils.Fail(c, utils.CodeForbidden, "未绑定俱乐部，无法访问")
			c.Abort()
			return
		}
		clubID, ok := clubIDAny.(int64)
		if !ok || clubID == 0 {
			utils.Fail(c, utils.CodeForbidden, "俱乐部信息异常")
			c.Abort()
			return
		}

		// 将 club_id 写入 query/body 可读取的位置标记，供 service 层强制过滤
		c.Set(ContextKeyClubScope, clubID)
		c.Next()
	}
}

// ContextKeyClubScope 俱乐部范围限制在 gin.Context 中的键
const ContextKeyClubScope = "club_scope_id"

// GetClubScopeID 从 gin.Context 获取俱乐部范围 ID(供 service 层调用)
// 返回 0 表示不受俱乐部范围限制(平台管理员)
func GetClubScopeID(c *gin.Context) int64 {
	if v, exists := c.Get(ContextKeyClubScope); exists {
		if id, ok := v.(int64); ok {
			return id
		}
	}
	// 若无 scope 限制，回退到 claims 中的 club_id
	if v, exists := c.Get(utils.ContextKeyClubID); exists {
		if id, ok := v.(int64); ok {
			return id
		}
	}
	return 0
}

// IsPlatformAdmin 判断当前请求方是否为平台管理员(不受俱乐部范围限制)
func IsPlatformAdmin(c *gin.Context) bool {
	userType, _ := c.Get(utils.ContextKeyUserType)
	ut, _ := userType.(string)
	return ut == utils.JWTUserTypeAdmin
}
