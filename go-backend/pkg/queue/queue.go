package queue

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/hibiken/asynq"
	"github.com/jisan/e-sports-platform/internal/config"
)

// 任务类型定义
const (
	// TaskMessagePush 消息推送(订阅消息/通知推送)
	TaskMessagePush = "queue:message:push"
	// TaskAIScan AI 风控扫描(消息内容/订单风险)
	TaskAIScan = "queue:ai:scan"
	// TaskOrderSettle 订单结算(分账/佣金结算)
	TaskOrderSettle = "queue:order:settle"
	// TaskOrderTimeoutClose 订单超时关闭
	TaskOrderTimeoutClose = "queue:order:timeout_close"
	// TaskOrderRemind 预约订单提醒
	TaskOrderRemind = "queue:order:remind"
	// TaskWithdrawProcess 提现处理
	TaskWithdrawProcess = "queue:withdraw:process"
	// TaskOfflineMessagePush 离线消息推送补偿
	TaskOfflineMessagePush = "queue:offline:push"
)

// 队列名称(按优先级分组)
const (
	QueueCritical = "critical" // 高优先级：订单结算、提现
	QueueDefault  = "default"  // 默认：消息推送、提醒
	QueueLow      = "low"      // 低优先级：AI 扫描、离线补偿
)

// 消息推送任务载荷
type MessagePushPayload struct {
	UserID  int64  `json:"user_id"`  // 接收用户ID
	Type    string `json:"type"`     // 通知类型
	Title   string `json:"title"`    // 标题
	Content string `json:"content"`  // 内容
}

// AI 扫描任务载荷
type AIScanPayload struct {
	TargetType string `json:"target_type"` // 目标类型 message/order/user
	TargetID   int64  `json:"target_id"`   // 目标ID
	Content    string `json:"content"`     // 待扫描内容
}

// 订单结算任务载荷
type OrderSettlePayload struct {
	OrderID int64 `json:"order_id"` // 订单ID
}

// 订单超时关闭任务载荷
type OrderTimeoutPayload struct {
	OrderID  int64 `json:"order_id"`  // 订单ID
	Timeout  int   `json:"timeout"`   // 超时分钟数
}

// 提现处理任务载荷
type WithdrawProcessPayload struct {
	WithdrawID int64 `json:"withdraw_id"` // 提现记录ID
}

// Client Asynq 任务投递客户端封装
type Client struct {
	client *asynq.Client
}

// NewClient 创建任务投递客户端
func NewClient(cfg config.RedisConfig) *Client {
	client := asynq.NewClient(asynq.RedisClientOpt{
		Addr:     cfg.Addr(),
		Password: cfg.Password,
		DB:       cfg.DB,
	})
	return &Client{client: client}
}

// Close 关闭客户端
func (c *Client) Close() error {
	return c.client.Close()
}

// enqueue 投递任务
func (c *Client) enqueue(ctx context.Context, typeName, queue string, payload []byte, opts ...asynq.Option) error {
	defaultOpts := []asynq.Option{
		asynq.Queue(queue),
		asynq.MaxRetry(3),                   // 最大重试次数
		asynq.Timeout(5 * time.Minute),      // 单任务超时
		asynq.Retention(24 * time.Hour),     // 完成后保留时长
	}
	task := asynq.NewTask(typeName, payload, append(defaultOpts, opts...)...)
	info, err := c.client.EnqueueContext(ctx, task)
	if err != nil {
		return fmt.Errorf("投递任务 %s 失败: %w", typeName, err)
	}
	_ = info
	return nil
}

// EnqueueMessagePush 投递消息推送任务
func (c *Client) EnqueueMessagePush(ctx context.Context, p MessagePushPayload) error {
	payload, err := json.Marshal(p)
	if err != nil {
		return err
	}
	return c.enqueue(ctx, TaskMessagePush, QueueDefault, payload)
}

// EnqueueAIScan 投递 AI 扫描任务
func (c *Client) EnqueueAIScan(ctx context.Context, p AIScanPayload) error {
	payload, err := json.Marshal(p)
	if err != nil {
		return err
	}
	return c.enqueue(ctx, TaskAIScan, QueueLow, payload)
}

// EnqueueOrderSettle 投递订单结算任务
func (c *Client) EnqueueOrderSettle(ctx context.Context, p OrderSettlePayload) error {
	payload, err := json.Marshal(p)
	if err != nil {
		return err
	}
	return c.enqueue(ctx, TaskOrderSettle, QueueCritical, payload)
}

// EnqueueOrderTimeoutClose 投递订单超时关闭任务(延迟执行)
func (c *Client) EnqueueOrderTimeoutClose(ctx context.Context, p OrderTimeoutPayload) error {
	payload, err := json.Marshal(p)
	if err != nil {
		return err
	}
	delay := time.Duration(p.Timeout) * time.Minute
	return c.enqueue(ctx, TaskOrderTimeoutClose, QueueCritical, payload,
		asynq.ProcessIn(delay),
	)
}

// EnqueueWithdrawProcess 投递提现处理任务
func (c *Client) EnqueueWithdrawProcess(ctx context.Context, p WithdrawProcessPayload) error {
	payload, err := json.Marshal(p)
	if err != nil {
		return err
	}
	return c.enqueue(ctx, TaskWithdrawProcess, QueueCritical, payload)
}

// ---------------- 消费者服务 ----------------

// HandlerFunc 任务处理函数类型
type HandlerFunc func(context.Context, []byte) error

// Server Asynq 消费者服务封装
type Server struct {
	server *asynq.Server
	mux    *asynq.ServeMux
}

// HandlerRegistry 消费者注册表
type HandlerRegistry struct {
	handlers map[string]HandlerFunc
}

// NewHandlerRegistry 创建消费者注册表
func NewHandlerRegistry() *HandlerRegistry {
	return &HandlerRegistry{handlers: make(map[string]HandlerFunc)}
}

// Register 注册任务处理器
func (r *HandlerRegistry) Register(typeName string, handler HandlerFunc) {
	r.handlers[typeName] = handler
}

// NewServer 创建消费者服务
// cfg Redis 配置，concurrency 并发数
func NewServer(cfg config.RedisConfig, concurrency int) *Server {
	if concurrency <= 0 {
		concurrency = 10
	}
	srv := asynq.NewServer(
		asynq.RedisClientOpt{
			Addr:     cfg.Addr(),
			Password: cfg.Password,
			DB:       cfg.DB,
		},
		asynq.Config{
			Concurrency: concurrency,
			Queues: map[string]int{
				QueueCritical: 6, // 高优先级权重最大
				QueueDefault:  3,
				QueueLow:      1,
			},
			RetryDelayFunc: func(n int, _ error, _ *asynq.Task) time.Duration {
				// 指数退避重试
				return time.Duration(n*n) * time.Minute
			},
		},
	)
	mux := asynq.NewServeMux()
	return &Server{server: srv, mux: mux}
}

// RegisterHandlers 注册所有消费者处理器
func (s *Server) RegisterHandlers(registry *HandlerRegistry) {
	for typeName, handler := range registry.handlers {
		h := handler // 捕获循环变量
		s.mux.HandleFunc(typeName, func(ctx context.Context, t *asynq.Task) error {
			return h(ctx, t.Payload())
		})
	}
}

// HandleFunc 注册单个任务处理器
func (s *Server) HandleFunc(typeName string, handler HandlerFunc) {
	s.mux.HandleFunc(typeName, func(ctx context.Context, t *asynq.Task) error {
		return handler(ctx, t.Payload())
	})
}

// Start 启动消费者服务(阻塞)
func (s *Server) Start() error {
	return s.server.Run(s.mux)
}

// Shutdown 优雅关闭消费者服务
func (s *Server) Shutdown() {
	s.server.Shutdown()
}

// RedisOptions 返回 asynq 使用的 Redis 客户端选项(供外部诊断)
func RedisOptions(cfg config.RedisConfig) asynq.RedisClientOpt {
	return asynq.RedisClientOpt{
		Addr:     cfg.Addr(),
		Password: cfg.Password,
		DB:       cfg.DB,
	}
}
