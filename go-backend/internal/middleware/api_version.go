package middleware

import (
	"github.com/gin-gonic/gin"
	"github.com/jisan/e-sports-platform/internal/utils"
)

// 当前支持的 API 版本
const (
	APIVersionV1 = "v1"
	APIVersionV2 = "v2"
)

// 默认灰度发布配置
var (
	// grayWhitelist v2 版本灰度白名单(用户ID集合)
	grayWhitelist = map[int64]bool{}
	// grayRolloutPercent v2 灰度比例(0-100)
	grayRolloutPercent = 0
)

// APIVersionMiddleware API 版本控制中间件(支持灰度发布)
// 通过 X-Api-Version 请求头或 URL 前缀识别版本
// 灰度策略：白名单用户优先，其次按灰度比例随机命中
type APIVersionMiddleware struct {
	defaultVersion string
}

// NewAPIVersionMiddleware 创建 API 版本控制中间件
// defaultVersion 为默认版本(如 v1)
func NewAPIVersionMiddleware(defaultVersion string) *APIVersionMiddleware {
	if defaultVersion == "" {
		defaultVersion = APIVersionV1
	}
	return &APIVersionMiddleware{defaultVersion: defaultVersion}
}

// SetGrayConfig 设置灰度发布配置(线程不安全，应在启动时调用)
func SetGrayConfig(whitelist []int64, rolloutPercent int) {
	grayWhitelist = make(map[int64]bool, len(whitelist))
	for _, uid := range whitelist {
		grayWhitelist[uid] = true
	}
	if rolloutPercent < 0 {
		rolloutPercent = 0
	}
	if rolloutPercent > 100 {
		rolloutPercent = 100
	}
	grayRolloutPercent = rolloutPercent
}

// APIVersion 版本控制与灰度发布
func (m *APIVersionMiddleware) APIVersion() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 解析请求版本
		version := c.GetHeader("X-Api-Version")
		if version == "" {
			version = m.defaultVersion
		}

		// 未显式指定版本时，按灰度策略决定是否升级到 v2
		if c.GetHeader("X-Api-Version") == "" {
			userID := utils.GetCurrentUserID(c)
			if m.shouldGrayV2(userID) {
				version = APIVersionV2
			}
		}

		// 版本不支持的兜底
		if version != APIVersionV1 && version != APIVersionV2 {
			version = m.defaultVersion
		}

		// 注入版本信息
		c.Set(ContextKeyAPIVersion, version)
		c.Header("X-Api-Version", version)

		c.Next()
	}
}

// shouldGrayV2 判断是否命中 v2 灰度
func (m *APIVersionMiddleware) shouldGrayV2(userID int64) bool {
	// 白名单优先
	if userID > 0 && grayWhitelist[userID] {
		return true
	}
	// 按灰度比例(基于用户ID取模，保证同用户结果稳定)
	if grayRolloutPercent > 0 && userID > 0 {
		return int(userID%100) < grayRolloutPercent
	}
	return false
}

// ContextKeyAPIVersion API 版本在 gin.Context 中的键
const ContextKeyAPIVersion = "api_version"

// GetAPIVersion 从 gin.Context 获取当前 API 版本
func GetAPIVersion(c *gin.Context) string {
	if v, exists := c.Get(ContextKeyAPIVersion); exists {
		if s, ok := v.(string); ok {
			return s
		}
	}
	return APIVersionV1
}
