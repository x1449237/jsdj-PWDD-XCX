package middleware

import (
	"context"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jisan/e-sports-platform/internal/utils"
)

// ContextKeyTraceIDReq 标准 context 中 trace_id 的键类型
type ContextKeyTraceIDReq struct{}

// TraceID 请求追踪ID中间件：为每个请求生成 UUID 注入 context 与响应头
// 统一响应格式中 trace_id 字段即来源于此
func TraceID() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 优先使用上游传入的 X-Trace-Id
		traceID := c.GetHeader("X-Trace-Id")
		if traceID == "" {
			traceID = uuid.New().String()
		}
		// 注入 gin.Context(供 utils.GetTraceID 读取)
		c.Set(utils.ContextKeyTraceID, traceID)
		// 注入标准 context(供 service/repository 层传递)
		ctx := context.WithValue(c.Request.Context(), ContextKeyTraceIDReq{}, traceID)
		c.Request = c.Request.WithContext(ctx)
		// 写入响应头便于链路追踪
		c.Header("X-Trace-Id", traceID)
		c.Header("X-Response-Time-Start", time.Now().Format(time.RFC3339Nano))

		c.Next()
	}
}

// TraceIDFromContext 从标准 context 中获取 trace_id
func TraceIDFromContext(ctx context.Context) string {
	if v, ok := ctx.Value(ContextKeyTraceIDReq{}).(string); ok {
		return v
	}
	return ""
}
