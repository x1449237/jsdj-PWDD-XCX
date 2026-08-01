package websocket

import (
	"context"
	"encoding/json"
	"sync"
	"sync/atomic"
	"time"

	"github.com/jisan/e-sports-platform/pkg/cache"
	"go.uber.org/zap"
)

// Hub 连接管理中心，负责注册/注销/广播/点对点推送
type Hub struct {
	// clients 用户连接表: userID -> 该用户的多个客户端连接
	clients map[int64]map[*Client]struct{}
	mu      sync.RWMutex

	// redis 缓存客户端(用于离线消息存储)
	redis *cache.RedisClient

	logger *zap.Logger

	// onMessage 业务消息处理回调(由 service 层注入会话转发逻辑)
	onMessage func(*Client, *Message)

	// 统计
	connCount    int64 // 当前总连接数
	totalCount   int64 // 历史连接总数
	maxConnLimit int64 // 最大连接数限制
}

// NewHub 创建连接管理中心
func NewHub(redis *cache.RedisClient, logger *zap.Logger) *Hub {
	return &Hub{
		clients:      make(map[int64]map[*Client]struct{}),
		redis:        redis,
		logger:       logger,
		maxConnLimit: MaxConnections,
	}
}

// Register 注册客户端连接
func (h *Hub) Register(client *Client) {
	h.mu.Lock()
	defer h.mu.Unlock()

	// 安全:系统总连接数限制(防海量连接耗尽资源/DoS)
	if h.maxConnLimit > 0 && atomic.LoadInt64(&h.connCount) >= h.maxConnLimit {
		h.logger.Warn("WebSocket 总连接数已达上限,拒绝新连接",
			zap.Int64("user_id", client.userID),
			zap.Int64("current", atomic.LoadInt64(&h.connCount)),
			zap.Int64("max", h.maxConnLimit))
		_ = client.conn.Close()
		return
	}

	// 单用户连接数限制
	if len(h.clients[client.userID]) >= MaxConnPerUser {
		// 踢掉最早的一个连接(简单实现：随机取一个关闭)
		for c := range h.clients[client.userID] {
			_ = c.conn.Close()
			break
		}
	}

	if _, ok := h.clients[client.userID]; !ok {
		h.clients[client.userID] = make(map[*Client]struct{})
	}
	h.clients[client.userID][client] = struct{}{}

	atomic.AddInt64(&h.connCount, 1)
	atomic.AddInt64(&h.totalCount, 1)

	h.logger.Info("WebSocket 连接注册",
		zap.Int64("user_id", client.userID),
		zap.String("conn_id", client.connID),
		zap.Int64("current_conns", atomic.LoadInt64(&h.connCount)))

	// 连接数告警检测
	h.checkAlert()
}

// Unregister 注销客户端连接
func (h *Hub) Unregister(client *Client) {
	h.mu.Lock()
	defer h.mu.Unlock()

	if conns, ok := h.clients[client.userID]; ok {
		if _, exists := conns[client]; exists {
			delete(conns, client)
			atomic.AddInt64(&h.connCount, -1)
			if len(conns) == 0 {
				delete(h.clients, client.userID)
			}
		}
	}
}

// SendToUser 点对点推送：用户在线则直接推送，离线则存入离线消息队列
func (h *Hub) SendToUser(ctx context.Context, userID int64, msg *Message) error {
	data, err := msg.Encode()
	if err != nil {
		return err
	}

	// 在线直接推送
	if h.isUserOnline(userID) {
		h.mu.RLock()
		conns := h.clients[userID]
		// 复制一份避免持锁过久
		targets := make([]*Client, 0, len(conns))
		for c := range conns {
			targets = append(targets, c)
		}
		h.mu.RUnlock()

		delivered := false
		for _, c := range targets {
			if c.Send(data) {
				delivered = true
			}
		}
		if delivered {
			return nil
		}
	}

	// 用户离线或投递失败，存入 Redis Sorted Set(以时间戳为 score)
	return h.storeOffline(ctx, userID, data)
}

// luaTrimOffline 离线消息裁剪+TTL Lua 脚本(原子执行)
// 保留最新的 max 条(score 升序删除多余),并刷新 TTL
const luaTrimOffline = `
local cnt = redis.call('ZCARD', KEYS[1])
local max = tonumber(ARGV[1])
if cnt > max then
  redis.call('ZREMRANGEBYRANK', KEYS[1], 0, cnt - max - 1)
end
return redis.call('EXPIRE', KEYS[1], ARGV[2])
`

// storeOffline 存储离线消息到 Redis Sorted Set
// 安全修复:
// 1. 限制单用户离线消息数量(MaxOfflineMsgPerUser),超过则淘汰最早消息(防内存膨胀/恶意灌入)
// 2. 设置 TTL(OfflineMsgTTL=7天),到期自动清理(原无限期保留)
// 3. 用 Lua 原子执行裁剪+TTL,避免并发竞态
func (h *Hub) storeOffline(ctx context.Context, userID int64, data []byte) error {
	if h.redis == nil {
		return nil
	}
	key := h.offlineKey(userID)
	score := float64(time.Now().UnixMilli())
	if err := h.redis.ZAdd(ctx, key, score, string(data)); err != nil {
		h.logger.Warn("存储离线消息失败", zap.Int64("user_id", userID), zap.Error(err))
		return err
	}
	// 裁剪超额 + 刷新 TTL(原子)
	ttlSec := int64(OfflineMsgTTL / time.Second)
	_, _ = h.redis.Eval(ctx, luaTrimOffline, []string{key}, MaxOfflineMsgPerUser, ttlSec)
	return nil
}

// Broadcast 广播消息给所有在线用户
func (h *Hub) Broadcast(msg *Message) {
	data, err := msg.Encode()
	if err != nil {
		return
	}
	h.mu.RLock()
	defer h.mu.RUnlock()

	for _, conns := range h.clients {
		for c := range conns {
			c.Send(data)
		}
	}
}

// BroadcastByUserIDs 广播消息给指定用户集合
func (h *Hub) BroadcastByUserIDs(userIDs []int64, msg *Message) {
	data, err := msg.Encode()
	if err != nil {
		return
	}
	h.mu.RLock()
	defer h.mu.RUnlock()

	for _, uid := range userIDs {
		if conns, ok := h.clients[uid]; ok {
			for c := range conns {
				c.Send(data)
			}
		}
	}
}

// handleClientMessage 处理客户端发来的业务消息(转发到对应会话接收方)
func (h *Hub) handleClientMessage(sender *Client, msg *Message) {
	// 此处仅做基础转发框架：将消息转发给会话中其他参与者
	// 具体会话成员解析由 service 层注入回调实现
	if h.onMessage != nil {
		h.onMessage(sender, msg)
	}
}

// onMessage 业务消息处理回调(由 service 层注入)
// SetMessageHandler 设置业务消息处理回调(供 service 层注入会话转发逻辑)
func (h *Hub) SetMessageHandler(handler func(*Client, *Message)) {
	h.onMessage = handler
}

// MarkInactive 标记连接为非活跃并关闭
func (h *Hub) MarkInactive(client *Client) {
	h.logger.Info("标记连接非活跃",
		zap.Int64("user_id", client.userID), zap.String("conn_id", client.connID))
	_ = client.conn.Close()
}

// isUserOnline 判断用户是否在线
func (h *Hub) isUserOnline(userID int64) bool {
	h.mu.RLock()
	defer h.mu.RUnlock()
	conns, ok := h.clients[userID]
	return ok && len(conns) > 0
}

// IsUserOnline 线程安全地判断用户是否在线(导出)
func (h *Hub) IsUserOnline(userID int64) bool {
	return h.isUserOnline(userID)
}

// GetOnlineUsers 获取所有在线用户ID
func (h *Hub) GetOnlineUsers() []int64 {
	h.mu.RLock()
	defer h.mu.RUnlock()
	ids := make([]int64, 0, len(h.clients))
	for uid := range h.clients {
		ids = append(ids, uid)
	}
	return ids
}

// ConnectionCount 当前总连接数
func (h *Hub) ConnectionCount() int64 {
	return atomic.LoadInt64(&h.connCount)
}

// TotalCount 历史连接总数
func (h *Hub) TotalCount() int64 {
	return atomic.LoadInt64(&h.totalCount)
}

// OnlineUserCount 在线用户数
func (h *Hub) OnlineUserCount() int {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return len(h.clients)
}

// checkAlert 连接数告警检测(达到最大连接数的 80% 告警)
func (h *Hub) checkAlert() {
	current := atomic.LoadInt64(&h.connCount)
	threshold := h.maxConnLimit * int64(AlertThresholdPercent) / 100
	if current >= threshold {
		h.logger.Warn("WebSocket 连接数达到告警阈值",
			zap.Int64("current", current),
			zap.Int64("threshold", threshold),
			zap.Int64("max", h.maxConnLimit))
	}
}

// offlineKey 离线消息 Sorted Set 键
func (h *Hub) offlineKey(userID int64) string {
	return cache.OfflineMsgKey(userID)
}

// EncodeMessage 便捷方法：编码消息为 JSON
func EncodeMessage(msg *Message) ([]byte, error) {
	return json.Marshal(msg)
}
