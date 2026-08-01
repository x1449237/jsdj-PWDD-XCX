package middleware

import (
	"context"
	"fmt"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jisan/e-sports-platform/internal/config"
	"github.com/jisan/e-sports-platform/internal/utils"
	"github.com/jisan/e-sports-platform/pkg/cache"
)

// RateLimitMiddleware 接口限流中间件
// 区分普通用户/打手/管理员差异化 QPS，基于 Redis 固定窗口计数实现
type RateLimitMiddleware struct {
	redis *cache.RedisClient
	cfg   config.RateLimitConfig
}

// NewRateLimitMiddleware 创建限流中间件
func NewRateLimitMiddleware(redis *cache.RedisClient, cfg config.RateLimitConfig) *RateLimitMiddleware {
	return &RateLimitMiddleware{redis: redis, cfg: cfg}
}

// qpsByUserType 根据用户类型返回对应 QPS 限制
func (m *RateLimitMiddleware) qpsByUserType(userType string, role int8) int {
	switch userType {
	case utils.JWTUserTypeAdmin, utils.JWTUserTypeShopAdmin:
		if m.cfg.AdminQPS > 0 {
			return m.cfg.AdminQPS
		}
	case utils.JWTUserTypeUser:
		// 打手(角色位含 2)使用打手 QPS
		if role&2 != 0 && m.cfg.PlayerQPS > 0 {
			return m.cfg.PlayerQPS
		}
		if m.cfg.UserQPS > 0 {
			return m.cfg.UserQPS
		}
	}
	// 默认普通用户 QPS
	if m.cfg.UserQPS > 0 {
		return m.cfg.UserQPS
	}
	return 20
}

// RateLimit 限流中间件(按用户维度)
func (m *RateLimitMiddleware) RateLimit() gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := utils.GetCurrentUserID(c)
		userType, _ := c.Get(utils.ContextKeyUserType)
		ut, _ := userType.(string)
		roleAny, _ := c.Get(utils.ContextKeyUserRole)
		role, _ := roleAny.(int8)

		qps := m.qpsByUserType(ut, role)

		// 限流标识：未登录用户按 IP 维度
		var key string
		if userID > 0 {
			key = cache.RateLimitKey(userID, c.FullPath())
		} else {
			key = cache.RateLimitIPKey(c.ClientIP(), c.FullPath())
		}

		ctx, cancel := context.WithTimeout(c.Request.Context(), time.Second)
		defer cancel()

		allowed, count, err := m.allow(ctx, key, qps)
		if err != nil {
			// Redis 异常时放行，避免影响可用性
			c.Next()
			return
		}
		if !allowed {
			utils.FailWithHTTP(c, 429, utils.CodeTooMany, "请求过于频繁，请稍后再试")
			c.Abort()
			return
		}

		// 设置限流响应头
		c.Header("X-RateLimit-Limit", fmt.Sprintf("%d", qps))
		c.Header("X-RateLimit-Remaining", fmt.Sprintf("%d", int64(qps)-count))
		c.Next()
	}
}

// allow 基于 Redis 固定窗口计数判断是否放行
// 返回 allowed(是否放行), count(当前窗口计数), err
func (m *RateLimitMiddleware) allow(ctx context.Context, key string, qps int) (bool, int64, error) {
	client := m.redis.Client()
	// 使用 INCR + EXPIRE 实现固定窗口(首次设置过期时间 1 秒)
	pipe := client.TxPipeline()
	incr := pipe.Incr(ctx, key)
	pipe.Expire(ctx, key, time.Second)
	if _, err := pipe.Exec(ctx); err != nil {
		return false, 0, err
	}
	count := incr.Val()
	return count <= int64(qps), count, nil
}
