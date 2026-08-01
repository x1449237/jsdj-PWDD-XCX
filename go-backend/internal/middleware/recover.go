package middleware

import (
	"runtime/debug"

	"github.com/gin-gonic/gin"
	"github.com/jisan/e-sports-platform/internal/utils"
	"go.uber.org/zap"
)

// Recover 全局异常恢复中间件
// 捕获 panic，返回统一错误格式，不向客户端暴露服务器内部错误信息
func Recover(logger *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if rec := recover(); rec != nil {
				// 记录完整堆栈到日志(包含 trace_id 便于排查)
				traceID := utils.GetTraceID(c)
				logger.Error("服务异常 panic",
					zap.Any("recover", rec),
					zap.String("trace_id", traceID),
					zap.String("method", c.Request.Method),
					zap.String("path", c.Request.URL.Path),
					zap.String("client_ip", c.ClientIP()),
					zap.ByteString("stack", debug.Stack()),
				)

				// 返回统一错误格式，不暴露内部信息
				c.AbortWithStatusJSON(200, utils.Response{
					Code:    utils.CodeServerError,
					Msg:     "服务器开小差了，请稍后重试",
					Data:    struct{}{},
					TraceID: traceID,
				})
			}
		}()
		c.Next()
	}
}
