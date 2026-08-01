package utils

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// 统一响应格式: {"code": 200, "msg": "成功", "data": {}, "trace_id": "uuid"}

// ContextKeyTraceID trace_id 在 gin.Context 中的键(供 trace_id 中间件写入)
const ContextKeyTraceID = "trace_id"

// GetTraceID 从 gin.Context 获取 trace_id，不存在则返回空串
func GetTraceID(c *gin.Context) string {
	if v, exists := c.Get(ContextKeyTraceID); exists {
		if id, ok := v.(string); ok {
			return id
		}
	}
	return ""
}

// 业务码定义
const (
	CodeSuccess       = 200 // 成功
	CodeBadRequest    = 400 // 参数错误
	CodeUnauthorized  = 401 // 未认证
	CodeForbidden     = 403 // 无权限
	CodeNotFound      = 404 // 资源不存在
	CodeTooMany       = 429 // 限流
	CodeServerError   = 500 // 服务器错误
	CodeNotImplemented = 501 // 功能未实现
)

// Response 统一响应结构体
type Response struct {
	Code    int         `json:"code"`    // 业务码
	Msg     string      `json:"msg"`     // 提示信息
	Data    interface{} `json:"data"`    // 数据
	TraceID string      `json:"trace_id"` // 请求追踪ID
}

// ResponseWithTotal 带总数的列表响应结构体
type ResponseWithTotal struct {
	List  interface{} `json:"list"`
	Total int64       `json:"total"`
}

// Success 成功响应
func Success(c *gin.Context, data interface{}) {
	c.JSON(http.StatusOK, Response{
		Code:    CodeSuccess,
		Msg:     "成功",
		Data:    data,
		TraceID: GetTraceID(c),
	})
}

// SuccessWithTotal 成功响应(带分页总数)
func SuccessWithTotal(c *gin.Context, list interface{}, total int64) {
	c.JSON(http.StatusOK, Response{
		Code: CodeSuccess,
		Msg:  "成功",
		Data: ResponseWithTotal{
			List:  list,
			Total: total,
		},
		TraceID: GetTraceID(c),
	})
}

// Fail 失败响应
func Fail(c *gin.Context, code int, msg string) {
	c.JSON(http.StatusOK, Response{
		Code:    code,
		Msg:     msg,
		Data:    struct{}{},
		TraceID: GetTraceID(c),
	})
}

// FailWithTrace 失败响应(显式传入 traceID)
func FailWithTrace(c *gin.Context, code int, msg string, traceID string) {
	c.JSON(http.StatusOK, Response{
		Code:    code,
		Msg:     msg,
		Data:    struct{}{},
		TraceID: traceID,
	})
}

// FailWithHTTP 失败响应(同时设置 HTTP 状态码)
func FailWithHTTP(c *gin.Context, httpCode int, code int, msg string) {
	c.JSON(httpCode, Response{
		Code:    code,
		Msg:     msg,
		Data:    struct{}{},
		TraceID: GetTraceID(c),
	})
}
