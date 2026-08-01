package lock

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/jisan/e-sports-platform/pkg/cache"
)

var (
	// ErrLockNotAcquired 未获取到锁
	ErrLockNotAcquired = errors.New("锁已被占用，获取失败")
	// ErrLockNotHeld 释放未持有的锁
	ErrLockNotHeld = errors.New("未持有该锁，释放失败")
)

// 默认参数
const (
	defaultTTL          = 10 * time.Second // 默认锁有效期
	defaultRenewInterval = 3 * time.Second // 默认续期间隔
)

// 释放锁的 Lua 脚本：仅当 value 匹配时才删除(防止误删他人持有的锁)
const unlockScript = `
if redis.call("GET", KEYS[1]) == ARGV[1] then
    return redis.call("DEL", KEYS[1])
else
    return 0
end
`

// 续期锁的 Lua 脚本：仅当 value 匹配时才续期
const renewScript = `
if redis.call("GET", KEYS[1]) == ARGV[1] then
    return redis.call("PEXPIRE", KEYS[1], ARGV[2])
else
    return 0
end
`

// DistributedLock 基于 Redis SetNX 的分布式锁
type DistributedLock struct {
	redis *cache.RedisClient
}

// NewDistributedLock 创建分布式锁实例
func NewDistributedLock(r *cache.RedisClient) *DistributedLock {
	return &DistributedLock{redis: r}
}

// lockToken 持有的锁信息
type lockToken struct {
	key   string
	value string
}

// Lock 阻塞式获取锁，会一直重试直到成功或 ctx 取消
// 统一用于：抢单、分账、提现、活体校验等关键业务
func (d *DistributedLock) Lock(ctx context.Context, key string, ttl time.Duration) (*lockToken, error) {
	if ttl <= 0 {
		ttl = defaultTTL
	}
	value := uuid.New().String()

	// 自旋获取锁
	backoff := 50 * time.Millisecond
	for {
		ok, err := d.redis.SetNX(ctx, key, value, ttl)
		if err != nil {
			return nil, err
		}
		if ok {
			return &lockToken{key: key, value: value}, nil
		}
		// 等待重试
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-time.After(backoff):
		}
		if backoff < 500*time.Millisecond {
			backoff = backoff * 2
		}
	}
}

// TryLock 尝试获取锁一次，失败立即返回
func (d *DistributedLock) TryLock(ctx context.Context, key string, ttl time.Duration) (*lockToken, error) {
	if ttl <= 0 {
		ttl = defaultTTL
	}
	value := uuid.New().String()
	ok, err := d.redis.SetNX(ctx, key, value, ttl)
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, ErrLockNotAcquired
	}
	return &lockToken{key: key, value: value}, nil
}

// Unlock 释放锁
func (d *DistributedLock) Unlock(ctx context.Context, token *lockToken) error {
	if token == nil {
		return ErrLockNotHeld
	}
	res, err := d.redis.Eval(ctx, unlockScript, []string{token.key}, token.value)
	if err != nil {
		return err
	}
	if res.(int64) == 0 {
		return ErrLockNotHeld
	}
	return nil
}

// Renew 续期锁
func (d *DistributedLock) Renew(ctx context.Context, token *lockToken, ttl time.Duration) (bool, error) {
	if token == nil {
		return false, ErrLockNotHeld
	}
	if ttl <= 0 {
		ttl = defaultTTL
	}
	res, err := d.redis.Eval(ctx, renewScript, []string{token.key}, token.value, ttl.Milliseconds())
	if err != nil {
		return false, err
	}
	return res.(int64) == 1, nil
}

// AutoRenew 启动自动续期协程，返回停止函数
// 调用方应在业务完成后调用 stop 以停止续期
func (d *DistributedLock) AutoRenew(ctx context.Context, token *lockToken, ttl time.Duration) (stop func(), err error) {
	if token == nil {
		return nil, ErrLockNotHeld
	}
	if ttl <= 0 {
		ttl = defaultTTL
	}
	// 续期间隔为 TTL 的 1/3
	interval := ttl / 3
	if interval <= 0 {
		interval = defaultRenewInterval
	}

	stopCtx, cancel := context.WithCancel(ctx)
	var once sync.Once
	stop = func() { once.Do(cancel) }

	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()
		for {
			select {
			case <-stopCtx.Done():
				return
			case <-ticker.C:
				ok, err := d.Renew(stopCtx, token, ttl)
				if err != nil || !ok {
					// 续期失败，锁已丢失，停止续期
					return
				}
			}
		}
	}()

	return stop, nil
}

// WithLock 在锁保护下执行函数，执行完毕自动释放锁并停止续期
// ttl 为锁有效期，会自动续期保证业务执行期间锁不丢失
func (d *DistributedLock) WithLock(ctx context.Context, key string, ttl time.Duration, fn func() error) error {
	token, err := d.Lock(ctx, key, ttl)
	if err != nil {
		return fmt.Errorf("获取分布式锁失败: %w", err)
	}
	defer d.Unlock(context.Background(), token)

	stop, err := d.AutoRenew(ctx, token, ttl)
	if err != nil {
		return fmt.Errorf("启动锁自动续期失败: %w", err)
	}
	defer stop()

	return fn()
}
