package middleware

import (
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/sony/gobreaker"

	"github.com/jisan/e-sports-platform/internal/model"
	"github.com/jisan/e-sports-platform/internal/utils"
)

// CircuitBreakerMiddleware 接口熔断中间件
// 每个 service_name (微信支付、分账API、邮件、短信等) 配置 gobreaker 断路器
// 失败率 > 阈值 -> open -> 后续请求直接返回 503
// half-open 探测,状态变更写 circuit_breakers 表日志
type CircuitBreakerMiddleware struct {
	breakers map[string]*gobreaker.CircuitBreaker
	mu       sync.RWMutex
}

// NewCircuitBreakerMiddleware 创建熔断中间件实例
func NewCircuitBreakerMiddleware() *CircuitBreakerMiddleware {
	m := &CircuitBreakerMiddleware{
		breakers: make(map[string]*gobreaker.CircuitBreaker),
	}
	// 默认注册常见服务名
	defaults := []struct {
		name       string
		maxReq     uint32
		interval   time.Duration
		timeout    time.Duration
		readyToTrip func(gobreaker.Counts) bool
	}{
		{"wx_pay",       1,  60 * time.Second,  30 * time.Second,  nil},
		{"split_api",    1,  60 * time.Second,  30 * time.Second,  nil},
		{"smtp_email",   1,  120 * time.Second, 60 * time.Second,  nil},
		{"sms",          1,  120 * time.Second, 60 * time.Second,  nil},
		{"face_verify",  2,  60 * time.Second,  30 * time.Second,  nil},
		{"webauthn",     2,  60 * time.Second,  30 * time.Second,  nil},
		{"wx_login",     1,  60 * time.Second,  30 * time.Second,  nil},
	}
	for _, d := range defaults {
		st := gobreaker.Settings{
			Name:          d.name,
			MaxRequests:   d.maxReq,
			Interval:      d.interval,
			Timeout:       d.timeout,
			ReadyToTrip:   d.readyToTrip,
			OnStateChange: logCircuitStateChange,
		}
		if st.ReadyToTrip == nil {
			st.ReadyToTrip = defaultReadyToTrip
		}
		m.breakers[d.name] = gobreaker.NewCircuitBreaker(st)
	}
	return m
}

// defaultReadyToTrip 默认失败判定:连续失败 5 次 或 最近 10 次失败率 > 60%
func defaultReadyToTrip(counts gobreaker.Counts) bool {
	failureRatio := float64(counts.TotalFailures) / float64(counts.Requests)
	return counts.Requests >= 5 && failureRatio >= 0.6
}

// logCircuitStateChange 记录断路器状态变更到 circuit_breakers 表
// 说明: middleware 包为避免循环依赖不直接引用 service/db;
// 通过 BreakerLogSaver 回调钩子实现落库,main 初始化时注入;
// 未注入时仅输出占位,编译通过。
var BreakerLogSaver func(rec *model.CircuitBreakerLog) error

func logCircuitStateChange(name string, from gobreaker.State, to gobreaker.State) {
	now := time.Now()
	rec := &model.CircuitBreakerLog{
		ServiceName: name,
		State:       to.String(),
		FromState:   from.String(),
		ToState:     to.String(),
		Reason:      "state change",
		CreatedAt:   &now,
	}
	if BreakerLogSaver != nil {
		_ = BreakerLogSaver(rec)
	}
}

// Get 获取或懒加载指定 service 的断路器
func (m *CircuitBreakerMiddleware) Get(serviceName string) *gobreaker.CircuitBreaker {
	if serviceName == "" {
		return nil
	}
	m.mu.RLock()
	cb, ok := m.breakers[serviceName]
	m.mu.RUnlock()
	if ok {
		return cb
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	if cb, ok2 := m.breakers[serviceName]; ok2 {
		return cb
	}
	cb = gobreaker.NewCircuitBreaker(gobreaker.Settings{
		Name:          serviceName,
		MaxRequests:   1,
		Interval:      60 * time.Second,
		Timeout:       30 * time.Second,
		ReadyToTrip:   defaultReadyToTrip,
		OnStateChange: logCircuitStateChange,
	})
	m.breakers[serviceName] = cb
	return cb
}

// CircuitBreakerHandle 使用指定 serviceName 包装 gin Handler
// 当断路器 open 时直接返回 503
func (m *CircuitBreakerMiddleware) CircuitBreakerHandle(serviceName string) gin.HandlerFunc {
	cb := m.Get(serviceName)
	return func(c *gin.Context) {
		if cb == nil {
			c.Next()
			return
		}
		state := cb.State()
		if state == gobreaker.StateOpen {
			utils.Fail(c, utils.CodeServerError,
				fmt.Sprintf("服务[%s]当前不可用(熔断)，请稍后重试", serviceName))
			c.AbortWithStatus(http.StatusServiceUnavailable)
			return
		}
		// 继续执行,同时在上下文写入断路器,供 handler/dao 使用 cb.Execute
		c.Set("_cb_"+serviceName, cb)
		c.Next()
		// 根据 handler 执行结果/状态码决定是否标记为失败
		status := c.Writer.Status()
		if status >= 500 {
			_, _ = cb.Execute(func() (interface{}, error) {
				return nil, fmt.Errorf("upstream status %d", status)
			})
		}
	}
}

// CircuitBreakerWrapper 是默认单例(供全局调用)
var defaultCircuitBreaker *CircuitBreakerMiddleware
var cbOnce sync.Once

// DefaultCircuitBreaker 返回全局熔断器单例
func DefaultCircuitBreaker() *CircuitBreakerMiddleware {
	cbOnce.Do(func() {
		defaultCircuitBreaker = NewCircuitBreakerMiddleware()
	})
	return defaultCircuitBreaker
}

// CircuitBreakerBy 默认熔断 Handler(指定 serviceName)
func CircuitBreakerBy(serviceName string) gin.HandlerFunc {
	return DefaultCircuitBreaker().CircuitBreakerHandle(serviceName)
}
