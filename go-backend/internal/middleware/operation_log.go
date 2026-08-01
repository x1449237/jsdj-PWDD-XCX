package middleware

import (
	"encoding/json"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jisan/e-sports-platform/internal/model"
	"github.com/jisan/e-sports-platform/internal/utils"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// OperationLogMiddleware 操作日志记录中间件
// 异步记录管理端操作行为到 operation_logs 表，含操作人、动作、对象、IP、设备信息
type OperationLogMiddleware struct {
	db     *gorm.DB
	logger *zap.Logger
}

// NewOperationLogMiddleware 创建操作日志中间件
func NewOperationLogMiddleware(db *gorm.DB, logger *zap.Logger) *OperationLogMiddleware {
	return &OperationLogMiddleware{db: db, logger: logger}
}

// OperationLog 操作日志记录(异步写入)
// action 为操作动作标识，targetType 为操作对象类型
func (m *OperationLogMiddleware) OperationLog(action, targetType string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 先执行业务
		c.Next()

		// 异步记录日志，不阻塞响应
		go func(c *gin.Context, action, targetType string) {
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

			content, _ := json.Marshal(c.Request.URL.Query())

			log := &model.OperationLog{
				OperatorID:   operatorID,
				OperatorType: ut,
				Action:       action,
				TargetType:   targetType,
				TargetID:     targetID,
				Content:      content,
				IP:           c.ClientIP(),
				DeviceInfo:   c.GetHeader("X-Device-Info"),
				CreatedAt:    nowPtr(),
			}

			if err := m.db.Create(log).Error; err != nil {
				m.logger.Warn("写入操作日志失败",
					zap.Int64("operator_id", operatorID),
					zap.String("action", action),
					zap.Error(err))
			}
		}(c.Copy(), action, targetType)
	}
}

// nowPtr 返回当前时间的指针
func nowPtr() *time.Time {
	now := time.Now()
	return &now
}
