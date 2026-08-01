package queue

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"github.com/hibiken/asynq"
	redislib "github.com/redis/go-redis/v9"
	"gorm.io/gorm"

	"github.com/jisan/e-sports-platform/internal/config"
	"github.com/jisan/e-sports-platform/internal/model"
	"github.com/jisan/e-sports-platform/pkg/websocket"
)

const (
	TaskMessagePush         = "queue:message:push"
	TaskAIScan              = "queue:ai:scan"
	TaskOrderSettle         = "queue:order:settle"
	TaskOrderTimeoutClose   = "queue:order:timeout_close"
	TaskOrderRemind         = "queue:order:remind"
	TaskWithdrawProcess     = "queue:withdraw:process"
	TaskWithdrawPaid        = "queue:withdraw:paid"
	TaskOfflineMessagePush  = "queue:offline:push"
)

const (
	QueueCritical = "critical"
	QueueDefault  = "default"
	QueueLow      = "low"
)

type MessagePushPayload struct {
	UserID  int64  `json:"user_id"`
	Type    string `json:"type"`
	Title   string `json:"title"`
	Content string `json:"content"`
}

type AIScanPayload struct {
	TargetType string `json:"target_type"`
	TargetID   int64  `json:"target_id"`
	Content    string `json:"content"`
}

type OrderSettlePayload struct {
	OrderID int64 `json:"order_id"`
}

type OrderTimeoutPayload struct {
	OrderID  int64 `json:"order_id"`
	Timeout  int   `json:"timeout"`
}

type WithdrawProcessPayload struct {
	WithdrawID int64 `json:"withdraw_id"`
}

type Client struct {
	client *asynq.Client
}

func NewClient(cfg config.RedisConfig) *Client {
	client := asynq.NewClient(asynq.RedisClientOpt{
		Addr:     cfg.Addr(),
		Password: cfg.Password,
		DB:       cfg.DB,
	})
	return &Client{client: client}
}

func (c *Client) Close() error {
	return c.client.Close()
}

func (c *Client) enqueue(ctx context.Context, typeName, queue string, payload []byte, opts ...asynq.Option) error {
	defaultOpts := []asynq.Option{
		asynq.Queue(queue),
		asynq.MaxRetry(3),
		asynq.Timeout(5 * time.Minute),
		asynq.Retention(24 * time.Hour),
	}
	task := asynq.NewTask(typeName, payload, append(defaultOpts, opts...)...)
	info, err := c.client.EnqueueContext(ctx, task)
	if err != nil {
		return fmt.Errorf("投递任务 %s 失败: %w", typeName, err)
	}
	_ = info
	return nil
}

func (c *Client) EnqueueMessagePush(ctx context.Context, p MessagePushPayload) error {
	payload, err := json.Marshal(p)
	if err != nil {
		return err
	}
	return c.enqueue(ctx, TaskMessagePush, QueueDefault, payload)
}

func (c *Client) EnqueueAIScan(ctx context.Context, p AIScanPayload) error {
	payload, err := json.Marshal(p)
	if err != nil {
		return err
	}
	return c.enqueue(ctx, TaskAIScan, QueueLow, payload)
}

func (c *Client) EnqueueOrderSettle(ctx context.Context, p OrderSettlePayload) error {
	payload, err := json.Marshal(p)
	if err != nil {
		return err
	}
	return c.enqueue(ctx, TaskOrderSettle, QueueCritical, payload)
}

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

func (c *Client) EnqueueWithdrawProcess(ctx context.Context, p WithdrawProcessPayload) error {
	payload, err := json.Marshal(p)
	if err != nil {
		return err
	}
	return c.enqueue(ctx, TaskWithdrawProcess, QueueCritical, payload)
}

func (c *Client) EnqueueOrderTimeoutCloseByOrderNo(ctx context.Context, orderNo string, delay time.Duration) error {
	return c.enqueue(ctx, TaskOrderTimeoutClose, QueueCritical, []byte(orderNo),
		asynq.ProcessIn(delay),
	)
}

func (c *Client) EnqueueOrderSettleDelayed(ctx context.Context, orderID int64, delay time.Duration) error {
	payload := strconv.FormatInt(orderID, 10)
	return c.enqueue(ctx, TaskOrderSettle, QueueCritical, []byte(payload),
		asynq.ProcessIn(delay),
	)
}

type Server struct {
	srv *asynq.Server
	mux *asynq.ServeMux
}

func NewServer(cfg config.RedisConfig) (*Server, error) {
	srv := asynq.NewServer(
		asynq.RedisClientOpt{
			Addr:     fmt.Sprintf("%s:%d", cfg.Host, cfg.Port),
			Password: cfg.Password,
			DB:       cfg.DB + 1,
		},
		asynq.Config{Concurrency: 20},
	)
	return &Server{srv: srv, mux: asynq.NewServeMux()}, nil
}

func (s *Server) RegisterHandlers(db *gorm.DB, rdb *redislib.Client, hub *websocket.Hub) {
	s.mux.HandleFunc(TaskOrderTimeoutClose, HandleOrderTimeoutClose(db))
	s.mux.HandleFunc(TaskOrderSettle, HandleOrderSettle(db))
	s.mux.HandleFunc(TaskWithdrawPaid, HandleWithdrawPaid(db))
	s.mux.HandleFunc(TaskAIScan, HandleAIScan(db, rdb))
	s.mux.HandleFunc(TaskMessagePush, HandlePushMessage(db, hub))
	s.mux.HandleFunc(TaskOfflineMessagePush, HandleOfflineMessagePush(db, rdb, hub))
}

func (s *Server) Start() error {
	return s.srv.Run(s.mux)
}

func (s *Server) Shutdown() {
	s.srv.Shutdown()
}

func HandleOrderTimeoutClose(db *gorm.DB) func(context.Context, *asynq.Task) error {
	return func(ctx context.Context, t *asynq.Task) error {
		orderNo := string(t.Payload())
		var order model.Order
		if err := db.Where("order_no = ?", orderNo).First(&order).Error; err != nil {
			return err
		}
		if order.Status != 0 {
			return nil
		}
		db.Model(&order).Update("status", 10)
		db.Create(&model.OrderStatusLog{OrderID: order.ID, FromStatus: 0, ToStatus: 10, Reason: "超时自动关闭"})
		if order.PayStatus == 1 {
		}
		return nil
	}
}

func HandleOrderSettle(db *gorm.DB) func(context.Context, *asynq.Task) error {
	return func(ctx context.Context, t *asynq.Task) error {
		oid, _ := strconv.ParseInt(string(t.Payload()), 10, 64)
		var order model.Order
		if err := db.First(&order, oid).Error; err != nil {
			return err
		}
		if order.Status != 5 {
			return nil
		}
		db.Model(&order).Update("status", 6)
		db.Create(&model.OrderStatusLog{OrderID: order.ID, FromStatus: 5, ToStatus: 6, Reason: "T+3自动结算"})
		return nil
	}
}

func HandleWithdrawPaid(db *gorm.DB) func(context.Context, *asynq.Task) error {
	return func(ctx context.Context, t *asynq.Task) error {
		wid, _ := strconv.ParseInt(string(t.Payload()), 10, 64)
		db.Model(&model.Withdraw{}).Where("id=?", wid).Updates(map[string]any{"status": "paid", "paid_at": time.Now()})
		return nil
	}
}

func HandleAIScan(db *gorm.DB, rdb *redislib.Client) func(context.Context, *asynq.Task) error {
	return func(ctx context.Context, t *asynq.Task) error {
		uid, _ := strconv.ParseInt(string(t.Payload()), 10, 64)
		db.Create(&model.AiRiskAlert{AlertType: "user_scan", TargetType: "user", TargetID: uid, Level: 1, Status: 0})
		return nil
	}
}

func HandlePushMessage(db *gorm.DB, hub *websocket.Hub) func(context.Context, *asynq.Task) error {
	return func(ctx context.Context, t *asynq.Task) error {
		var p model.Notification
		_ = json.Unmarshal(t.Payload(), &p)
		msg, _ := websocket.NewMessage("notification", 0, p.UserID, p)
		if hub != nil {
			_ = hub.SendToUser(ctx, p.UserID, msg)
		}
		return nil
	}
}

func HandleOfflineMessagePush(db *gorm.DB, rdb *redislib.Client, hub *websocket.Hub) func(context.Context, *asynq.Task) error {
	return func(ctx context.Context, t *asynq.Task) error {
		uid, _ := strconv.ParseInt(string(t.Payload()), 10, 64)
		if hub != nil && rdb != nil && hub.IsUserOnline(uid) {
			key := fmt.Sprintf("jisan:offline:%d", uid)
			for {
				results, err := rdb.ZPopMin(ctx, key, 100).Result()
				if err != nil || len(results) == 0 {
					break
				}
				for _, z := range results {
					if dataStr, ok := z.Member.(string); ok {
						msg := &websocket.Message{}
						if json.Unmarshal([]byte(dataStr), msg) == nil {
							_ = hub.SendToUser(ctx, uid, msg)
						}
					}
				}
			}
		}
		return nil
	}
}

type HandlerFunc func(context.Context, []byte) error

type ServerLegacy struct {
	server *asynq.Server
	mux    *asynq.ServeMux
}

type HandlerRegistry struct {
	handlers map[string]HandlerFunc
}

func NewHandlerRegistry() *HandlerRegistry {
	return &HandlerRegistry{handlers: make(map[string]HandlerFunc)}
}

func (r *HandlerRegistry) Register(typeName string, handler HandlerFunc) {
	r.handlers[typeName] = handler
}

func NewServerLegacy(cfg config.RedisConfig, concurrency int) *ServerLegacy {
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
				QueueCritical: 6,
				QueueDefault:  3,
				QueueLow:      1,
			},
			RetryDelayFunc: func(n int, _ error, _ *asynq.Task) time.Duration {
				return time.Duration(n*n) * time.Minute
			},
		},
	)
	mux := asynq.NewServeMux()
	return &ServerLegacy{server: srv, mux: mux}
}

func (s *ServerLegacy) RegisterHandlers(registry *HandlerRegistry) {
	for typeName, handler := range registry.handlers {
		h := handler
		s.mux.HandleFunc(typeName, func(ctx context.Context, t *asynq.Task) error {
			return h(ctx, t.Payload())
		})
	}
}

func (s *ServerLegacy) HandleFunc(typeName string, handler HandlerFunc) {
	s.mux.HandleFunc(typeName, func(ctx context.Context, t *asynq.Task) error {
		return handler(ctx, t.Payload())
	})
}

func (s *ServerLegacy) Start() error {
	return s.server.Run(s.mux)
}

func (s *ServerLegacy) Shutdown() {
	s.server.Shutdown()
}

func RedisOptions(cfg config.RedisConfig) asynq.RedisClientOpt {
	return asynq.RedisClientOpt{
		Addr:     cfg.Addr(),
		Password: cfg.Password,
		DB:       cfg.DB,
	}
}
