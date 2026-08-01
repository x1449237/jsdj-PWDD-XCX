package cache

import (
	"context"
	"encoding/json"
	"errors"
	"math/rand"
	"sync"
	"time"

	"github.com/redis/go-redis/v9"
)

// 缓存空值标记，防止缓存穿透
const nullFlag = "__NULL_CACHE__"

// redisNil redis.Nil 错误，用于判断缓存不存在
var redisNil = redis.Nil

// Cache 分层缓存封装，提供缓存击穿/雪崩/穿透防护
type Cache struct {
	redis *RedisClient
	// 单飞锁表，防止缓存击穿(同一 key 并发回源合并)
	sfMu   sync.Mutex
	sfLocks map[string]*sync.Mutex
}

// NewCache 创建缓存封装
func NewCache(r *RedisClient) *Cache {
	return &Cache{
		redis:   r,
		sfLocks: make(map[string]*sync.Mutex),
	}
}

// singleflightLock 获取指定 key 的单飞锁(防止击穿)
func (c *Cache) singleflightLock(key string) *sync.Mutex {
	c.sfMu.Lock()
	defer c.sfMu.Unlock()
	if l, ok := c.sfLocks[key]; ok {
		return l
	}
	l := &sync.Mutex{}
	c.sfLocks[key] = l
	return l
}

// releaseSfLock 释放单飞锁表项(简单回收，避免无限增长)
func (c *Cache) releaseSfLock(key string) {
	c.sfMu.Lock()
	defer c.sfMu.Unlock()
	delete(c.sfLocks, key)
}

// GetOrSet 缓存击穿防护：缓存不存在时执行 loader 回源并写入缓存
// dest 必须为指针，结果将反序列化至 dest
func (c *Cache) GetOrSet(ctx context.Context, key string, dest interface{}, ttl time.Duration, loader func() (interface{}, error)) error {
	// 1. 先读缓存
	hit, err := c.getJSON(ctx, key, dest)
	if err != nil {
		return err
	}
	if hit {
		return nil
	}

	// 2. 缓存未命中，加单飞锁防止击穿(同一 key 并发只回源一次)
	l := c.singleflightLock(key)
	l.Lock()
	defer func() {
		l.Unlock()
		c.releaseSfLock(key)
	}()

	// 双重检查：可能在等待锁期间已被其他请求写入
	hit, err = c.getJSON(ctx, key, dest)
	if err != nil {
		return err
	}
	if hit {
		return nil
	}

	// 3. 回源
	data, err := loader()
	if err != nil {
		return err
	}

	// 4. 空值缓存(防穿透)，使用较短 TTL
	if data == nil {
		return c.Set(ctx, key, nullFlag, TTLShort)
	}

	// 5. 写入缓存，TTL 加随机抖动(防雪崩)
	return c.setJSON(ctx, key, data, jitterTTL(ttl))
}

// getJSON 读取并反序列化，返回是否命中
func (c *Cache) getJSON(ctx context.Context, key string, dest interface{}) (bool, error) {
	raw, err := c.redis.Get(ctx, key)
	if err != nil {
		if errors.Is(err, redisNil) {
			return false, nil
		}
		return false, err
	}
	// 空值标记命中，视为缓存存在但数据为空
	if raw == nullFlag {
		return true, nil
	}
	if err := json.Unmarshal([]byte(raw), dest); err != nil {
		return false, err
	}
	return true, nil
}

// setJSON 序列化并写入
func (c *Cache) setJSON(ctx context.Context, key string, value interface{}, ttl time.Duration) error {
	bytes, err := json.Marshal(value)
	if err != nil {
		return err
	}
	return c.redis.Set(ctx, key, string(bytes), ttl)
}

// Set 写入缓存
func (c *Cache) Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error {
	return c.redis.Set(ctx, key, value, ttl)
}

// SetJSON 写入 JSON 缓存(带抖动 TTL)
func (c *Cache) SetJSON(ctx context.Context, key string, value interface{}, ttl time.Duration) error {
	return c.setJSON(ctx, key, value, jitterTTL(ttl))
}

// Get 读取原始字符串
func (c *Cache) Get(ctx context.Context, key string) (string, error) {
	return c.redis.Get(ctx, key)
}

// GetJSON 读取并反序列化
func (c *Cache) GetJSON(ctx context.Context, key string, dest interface{}) (bool, error) {
	return c.getJSON(ctx, key, dest)
}

// Del 删除缓存
func (c *Cache) Del(ctx context.Context, keys ...string) error {
	return c.redis.Del(ctx, keys...)
}

// SetHot 永久缓存热点数据(不设置过期时间)
func (c *Cache) SetHot(ctx context.Context, key string, value interface{}) error {
	return c.setJSON(ctx, key, value, 0)
}

// jitterTTL 在基础 TTL 上加随机抖动，防止大量 key 同时过期导致雪崩
// jitter 范围: [0, ttl*0.1]
func jitterTTL(ttl time.Duration) time.Duration {
	if ttl <= 0 {
		return ttl
	}
	jitter := time.Duration(rand.Int63n(int64(ttl) / 10))
	return ttl + jitter
}

// IsNil 判断是否为缓存不存在错误
func IsNil(err error) bool {
	return errors.Is(err, redisNil)
}
