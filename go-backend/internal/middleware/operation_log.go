package middleware

import (
	"encoding/json"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jisan/e-sports-platform/internal/model"
	"github.com/jisan/e-sports-platform/internal/utils"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// OperationLogMiddleware 操作日志记录中间件
// 异步记录管理端操作行为到 operation_logs 表，含操作人、动作、对象、IP、设备信息、结果、模块
type OperationLogMiddleware struct {
	db     *gorm.DB
	logger *zap.Logger
}

// NewOperationLogMiddleware 创建操作日志中间件
func NewOperationLogMiddleware(db *gorm.DB, logger *zap.Logger) *OperationLogMiddleware {
	return &OperationLogMiddleware{db: db, logger: logger}
}

// ContextKeyOpLogResult 操作结果在 gin.Context 中的键(success/fail)
const ContextKeyOpLogResult = "op_log_result"

// ContextKeyOpLogModule 操作模块在 gin.Context 中的键(club_join/deposit/vbadge 等)
const ContextKeyOpLogModule = "op_log_module"

// sensitiveQueryKeys 操作日志中需要脱敏的查询参数键(小写匹配)
// 这些参数可能携带 token、密码、身份证、银行卡等敏感信息
var sensitiveQueryKeys = map[string]bool{
	"token":          true,
	"access_token":   true,
	"password":       true,
	"old_password":   true,
	"new_password":   true,
	"id_card":        true,
	"idcard":         true,
	"bank_card":      true,
	"bankcard":       true,
	"card_no":        true,
	"cardno":         true,
	"phone":          true,
	"mobile":         true,
	"real_name":      true,
	"realname":       true,
	"verify_code":    true,
	"verifycode":     true,
	"sms_code":       true,
	"smscode":        true,
	"signature":      true,
	"sign":           true,
	"secret":         true,
	"api_key":        true,
	"apikey":         true,
	"private_key":    true,
	"privatekey":     true,
	"wechat_pay_key": true,
}

// sanitizeQueryMap 对 query 参数进行脱敏处理
// 敏感参数值替换为 "***REDACTED***"
func sanitizeQueryMap(m map[string][]string) map[string][]string {
	if len(m) == 0 {
		return m
	}
	out := make(map[string][]string, len(m))
	for k, v := range m {
		if sensitiveQueryKeys[strings.ToLower(k)] {
			out[k] = []string{"***REDACTED***"}
		} else {
			out[k] = v
		}
	}
	return out
}

// OperationLog 操作日志记录(异步写入，模块为空)
// action 为操作动作标识，targetType 为操作对象类型
func (m *OperationLogMiddleware) OperationLog(action, targetType string) gin.HandlerFunc {
	return m.OperationLogWithModule(action, targetType, "")
}

// OperationLogWithModule 操作日志记录(异步写入，带业务模块标识)
// module 用于细化业务域，如 club_join/deposit/vbadge/club_fine/transfer 等
func (m *OperationLogMiddleware) OperationLogWithModule(action, targetType, module string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 先执行业务
		c.Next()

		// 异步记录日志，不阻塞响应
		go func(c *gin.Context, action, targetType, module string) {
			operatorID := utils.GetCurrentUserID(c)
			if operatorID == 0 {
				return
			}

			userType, _ := c.Get(utils.ContextKeyUserType)
			ut, _ := userType.(string)

			// 操作对象 ID 优先从路径参数取
			var targetID int64
			if id := c.Param("id"); id != "" {
				// 简单解析，忽略错误
				var n int64
				for _, ch := range id {
					if ch < '0' || ch > '9' {
						break
					}
					n = n*10 + int64(ch-'0')
				}
				targetID = n
			}

			// 查询参数脱敏(避免 token/密码/身份证/银行卡等敏感信息写入日志)
			sanitizedQuery := sanitizeQueryMap(c.Request.URL.Query())
			content, _ := json.Marshal(sanitizedQuery)

			// 操作结果：默认 success，handler 可通过 context 覆盖为 fail
			result := "success"
			if v, exists := c.Get(ContextKeyOpLogResult); exists {
				if r, ok := v.(string); ok && r != "" {
					result = r
				}
			}
			// 模块优先取 handler 注入的 context 值
			if v, exists := c.Get(ContextKeyOpLogModule); exists {
				if mod, ok := v.(string); ok && mod != "" {
					module = mod
				}
			}

			// HTTP 状态码 >= 400 视为失败
			if c.Writer.Status() >= 400 {
				result = "fail"
			}

			log := &model.OperationLog{
				OperatorID:   operatorID,
				OperatorType: ut,
				Action:       action,
				TargetType:   targetType,
				TargetID:     targetID,
				Content:      content,
				IP:           c.ClientIP(),
				DeviceInfo:   c.GetHeader("X-Device-Info"),
				Result:       result,
				Module:       module,
				CreatedAt:    nowPtr(),
			}

			if err := m.db.Create(log).Error; err != nil {
				m.logger.Warn("写入操作日志失败",
					zap.Int64("operator_id", operatorID),
					zap.String("action", action),
					zap.Error(err))
			}
		}(c.Copy(), action, targetType, module)
	}
}

// nowPtr 返回当前时间的指针
func nowPtr() *time.Time {
	now := time.Now()
	return &now
}
