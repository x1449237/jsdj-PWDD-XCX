package middleware

import (
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

// CORSConfig CORS 跨域配置
type CORSConfig struct {
	AllowOrigins     []string      // 允许的源
	AllowMethods     []string      // 允许的方法
	AllowHeaders     []string      // 允许的请求头
	ExposeHeaders    []string      // 暴露的响应头
	AllowCredentials bool          // 是否允许携带凭证
	MaxAge           time.Duration // 预检请求缓存时长
}

// DefaultCORSConfig 默认 CORS 配置(开发环境宽松，生产环境需收紧 AllowOrigins)
func DefaultCORSConfig() CORSConfig {
	return CORSConfig{
		AllowOrigins: []string{"*"},
		AllowMethods: []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders: []string{
			"Origin",
			"Content-Type",
			"Accept",
			"Authorization",
			"X-Trace-Id",
			"X-Api-Version",
			"X-Device-Info",
			"X-Club-Id",
		},
		ExposeHeaders: []string{
			"Content-Length",
			"X-Trace-Id",
			"X-Response-Time-Start",
		},
		AllowCredentials: false,
		MaxAge:           12 * time.Hour,
	}
}

// CORS CORS 跨域中间件
func CORS(cfg CORSConfig) gin.HandlerFunc {
	if len(cfg.AllowOrigins) == 0 {
		cfg = DefaultCORSConfig()
	}
	if cfg.MaxAge == 0 {
		cfg.MaxAge = 12 * time.Hour
	}
	if len(cfg.AllowMethods) == 0 {
		cfg.AllowMethods = []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"}
	}
	if len(cfg.AllowHeaders) == 0 {
		cfg.AllowHeaders = []string{"Origin", "Content-Type", "Accept", "Authorization", "X-Trace-Id", "X-Api-Version"}
	}

	return cors.New(cors.Config{
		AllowOrigins:     cfg.AllowOrigins,
		AllowMethods:     cfg.AllowMethods,
		AllowHeaders:     cfg.AllowHeaders,
		ExposeHeaders:    cfg.ExposeHeaders,
		AllowCredentials: cfg.AllowCredentials,
		MaxAge:           cfg.MaxAge,
	})
}
