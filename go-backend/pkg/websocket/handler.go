package websocket

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/jisan/e-sports-platform/internal/utils"
	"go.uber.org/zap"
)

// Handler WebSocket 升级处理器
type Handler struct {
	hub    *Hub
	logger *zap.Logger
	upgrader websocket.Upgrader
}

// NewHandler 创建 WebSocket 升级处理器
func NewHandler(hub *Hub, logger *zap.Logger) *Handler {
	return &Handler{
		hub:    hub,
		logger: logger,
		upgrader: websocket.Upgrader{
			ReadBufferSize:  10240,
			WriteBufferSize: 10240,
			// 允许跨域(生产环境应按域名白名单校验)
			CheckOrigin: func(r *http.Request) bool {
				return true
			},
			HandshakeTimeout: 10 * time.Second,
		},
	}
}

// Handle 连接处理入口(供路由注册调用)
// 需在 auth 中间件之后调用，从 context 读取已认证的用户信息
func (h *Handler) Handle(c *gin.Context) {
	// 从上下文获取用户信息(由 auth 中间件注入)
	userID := utils.GetCurrentUserID(c)
	if userID == 0 {
		c.JSON(http.StatusUnauthorized, gin.H{"code": 401, "msg": "未认证"})
		return
	}

	userType := UserTypeUser
	if ut, exists := c.Get(utils.ContextKeyUserType); exists {
		if s, ok := ut.(string); ok && s != "" {
			userType = s
		}
	}

	// 升级为 WebSocket 连接
	conn, err := h.upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		h.logger.Warn("WebSocket 升级失败",
			zap.Int64("user_id", userID), zap.Error(err))
		return
	}

	// 创建客户端并绑定 Hub
	client := NewClientWithHub(h.hub, conn, userID, userType, h.logger)

	h.logger.Info("WebSocket 新连接",
		zap.Int64("user_id", userID),
		zap.String("user_type", userType),
		zap.String("conn_id", client.ConnID()))

	// 启动客户端读写循环(阻塞直到连接关闭)
	ctx := context.WithValue(c.Request.Context(), utils.ContextKeyUserID, userID)
	client.Start(ctx)
}

// Upgrade 原始升级方法(供非 Gin 场景使用)
func (h *Handler) Upgrade(w http.ResponseWriter, r *http.Request, userID int64, userType string) {
	conn, err := h.upgrader.Upgrade(w, r, nil)
	if err != nil {
		h.logger.Warn("WebSocket 升级失败",
			zap.Int64("user_id", userID), zap.Error(err))
		return
	}
	client := NewClientWithHub(h.hub, conn, userID, userType, h.logger)
	client.Start(r.Context())
}
