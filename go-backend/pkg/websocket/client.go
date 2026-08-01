package websocket

import (
	"context"
	"time"

	"github.com/gorilla/websocket"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

// Client WebSocket 客户端连接
type Client struct {
	hub      *Hub
	conn     *websocket.Conn
	userID   int64    // 用户ID
	userType string   // 用户类型 user/admin/platform
	connID   string   // 连接唯一ID
	send     chan []byte // 待发送消息缓冲
	logger   *zap.Logger

	// 活跃度探针相关
	probeMissCount int // 连续无响应次数
	isActive       bool
	probeTimer     *time.Timer
}

// NewClientWithHub 创建客户端并绑定 Hub
func NewClientWithHub(hub *Hub, conn *websocket.Conn, userID int64, userType string, logger *zap.Logger) *Client {
	return &Client{
		hub:      hub,
		conn:     conn,
		userID:   userID,
		userType: userType,
		connID:   uuid.New().String(),
		send:     make(chan []byte, 256),
		logger:   logger,
		isActive: true,
	}
}

// UserID 返回用户ID
func (c *Client) UserID() int64 { return c.userID }

// ConnID 返回连接ID
func (c *Client) ConnID() string { return c.connID }

// IsActive 返回是否活跃
func (c *Client) IsActive() bool { return c.isActive }

// Send 直接向客户端发送缓冲投递消息(非阻塞，缓冲满则关闭连接)
func (c *Client) Send(msg []byte) bool {
	select {
	case c.send <- msg:
		return true
	default:
		// 缓冲已满，关闭连接
		c.logger.Warn("WebSocket 发送缓冲已满，关闭连接",
			zap.Int64("user_id", c.userID), zap.String("conn_id", c.connID))
		return false
	}
}

// readPump 读取客户端消息的循环
func (c *Client) readPump() {
	defer func() {
		c.hub.Unregister(c)
		_ = c.conn.Close()
	}()

	c.conn.SetReadLimit(MaxMessageSize)
	// 设置初始读超时
	_ = c.conn.SetReadDeadline(time.Now().Add(TimeoutDuration))

	// Pong 处理：收到 Pong 即重置读超时并标记活跃
	c.conn.SetPongHandler(func(string) error {
		_ = c.conn.SetReadDeadline(time.Now().Add(TimeoutDuration))
		c.isActive = true
		c.probeMissCount = 0
		return nil
	})

	for {
		_, message, err := c.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseNormalClosure) {
				c.logger.Warn("WebSocket 读取异常",
					zap.Int64("user_id", c.userID), zap.Error(err))
			}
			return
		}

		// 解析消息处理心跳与探针响应
		msg, err := DecodeMessage(message)
		if err != nil {
			continue
		}
		switch msg.Type {
		case MsgTypePong, MsgTypeProbeAck:
			// 客户端主动 Pong/探针响应，标记活跃
			c.isActive = true
			c.probeMissCount = 0
			_ = c.conn.SetReadDeadline(time.Now().Add(TimeoutDuration))
		case MsgTypePing:
			// 响应客户端心跳
			pong, _ := (&Message{Type: MsgTypePong, Timestamp: time.Now().UnixMilli()}).Encode()
			c.Send(pong)
		default:
			// 业务消息交由 Hub 处理(如转发到对应会话)
			c.hub.handleClientMessage(c, msg)
		}
	}
}

// writePump 向客户端写消息的循环，并按心跳间隔发送 Ping
func (c *Client) writePump() {
	ticker := time.NewTicker(HeartbeatInterval) // 25 秒 Ping 一次
	probeTicker := time.NewTicker(ProbeInterval) // 5 分钟探针一次
	defer func() {
		ticker.Stop()
		probeTicker.Stop()
		_ = c.conn.Close()
	}()

	for {
		select {
		case message, ok := <-c.send:
			_ = c.conn.SetWriteDeadline(time.Now().Add(WriteWait))
			if !ok {
				// Hub 关闭了 send 通道
				_ = c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}
			if err := c.conn.WriteMessage(websocket.TextMessage, message); err != nil {
				return
			}

		case <-ticker.C:
			// 心跳 Ping-Pong：发送 Ping
			_ = c.conn.SetWriteDeadline(time.Now().Add(WriteWait))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}

		case <-probeTicker.C:
			// 活跃度探针：每 5 分钟 Ping-Pong，连续 3 次无响应标记非活跃
			c.probe()
		}
	}
}

// probe 活跃度探针检测
func (c *Client) probe() {
	prevActive := c.isActive
	// 如果上一轮探针后未收到任何 Pong/响应，计为一次 miss
	if !prevActive {
		c.probeMissCount++
	}
	// 重置本轮活跃标记，等待下一次 Pong 响应
	c.isActive = false

	if c.probeMissCount >= ProbeMaxMiss {
		// 连续 3 次无响应，标记非活跃并断开连接
		c.logger.Info("WebSocket 连续探针无响应，标记非活跃并断开",
			zap.Int64("user_id", c.userID), zap.Int("miss", c.probeMissCount))
		c.hub.MarkInactive(c)
		return
	}

	// 发送探针消息
	probeMsg, _ := (&Message{Type: MsgTypeProbe, Timestamp: time.Now().UnixMilli()}).Encode()
	c.Send(probeMsg)
}

// pullOfflineMessages 重连后从 Redis Sorted Set 拉取离线消息并推送
func (c *Client) pullOfflineMessages(ctx context.Context) {
	if c.hub == nil || c.hub.redis == nil {
		return
	}
	messages, err := c.hub.redis.ZRangeByScore(ctx, c.hub.offlineKey(c.userID), "-inf", "+inf")
	if err != nil {
		c.logger.Warn("拉取离线消息失败", zap.Int64("user_id", c.userID), zap.Error(err))
		return
	}
	for _, msg := range messages {
		c.Send([]byte(msg))
	}
	// 拉取后清空离线消息
	if err := c.hub.redis.ZRemRangeByScore(ctx, c.hub.offlineKey(c.userID), "-inf", "+inf"); err != nil {
		c.logger.Warn("清理离线消息失败", zap.Int64("user_id", c.userID), zap.Error(err))
	}
}

// Start 启动客户端读写循环
func (c *Client) Start(ctx context.Context) {
	// 注册到 Hub
	c.hub.Register(c)
	// 拉取离线消息
	go c.pullOfflineMessages(ctx)
	// 启动读写循环
	go c.writePump()
	c.readPump()
}
