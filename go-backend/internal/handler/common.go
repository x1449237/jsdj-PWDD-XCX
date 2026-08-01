package handler

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/jisan/e-sports-platform/internal/middleware"
	"github.com/jisan/e-sports-platform/internal/utils"
)

// handler 通用辅助函数:仅做参数解析与上下文读取，不含业务逻辑。

// getCurrentUserID 从 gin.Context 获取当前登录用户ID
func getCurrentUserID(c *gin.Context) int64 {
	return utils.GetCurrentUserID(c)
}

// getClubScopeID 获取俱乐部范围限制ID(0 表示平台管理员不受限)
func getClubScopeID(c *gin.Context) int64 {
	return middleware.GetClubScopeID(c)
}

// isPlatformAdmin 判断当前请求方是否为平台管理员
func isPlatformAdmin(c *gin.Context) bool {
	return middleware.IsPlatformAdmin(c)
}

// getPage 解析分页参数(page 默认 1，pageSize 默认 10，最大 100)
func getPage(c *gin.Context) (int, int) {
	page := parseIntQuery(c, "page", 1)
	pageSize := parseIntQuery(c, "page_size", 10)
	return utils.ValidatePage(page, pageSize)
}

// parseIntQuery 从 query 解析整数参数，解析失败返回默认值
func parseIntQuery(c *gin.Context, key string, def int) int {
	s := c.Query(key)
	if s == "" {
		return def
	}
	n, err := strconv.Atoi(s)
	if err != nil {
		return def
	}
	return n
}

// parseInt8Query 从 query 解析 int8 参数，解析失败返回默认值
func parseInt8Query(c *gin.Context, key string, def int8) int8 {
	s := c.Query(key)
	if s == "" {
		return def
	}
	n, err := strconv.ParseInt(s, 10, 8)
	if err != nil {
		return def
	}
	return int8(n)
}

// parseInt64Path 从 path 参数解析 int64
func parseInt64Path(c *gin.Context, key string) int64 {
	n, _ := strconv.ParseInt(c.Param(key), 10, 64)
	return n
}

// parseInt64Query 从 query 解析 int64 参数，解析失败返回默认值
func parseInt64Query(c *gin.Context, key string, def int64) int64 {
	s := c.Query(key)
	if s == "" {
		return def
	}
	n, err := strconv.ParseInt(s, 10, 64)
	if err != nil {
		return def
	}
	return n
}

// getClientIP 获取客户端 IP
func getClientIP(c *gin.Context) string {
	return c.ClientIP()
}
