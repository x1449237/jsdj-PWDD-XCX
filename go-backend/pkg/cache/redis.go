package cache

import (
	"context"
	"fmt"
	"time"

	"github.com/jisan/e-sports-platform/internal/config"
	"github.com/redis/go-redis/v9"
)

// RedisClient 全局 Redis 客户端封装
type RedisClient struct {
	client *redis.Client
}

// NewRedisClient 根据配置创建 Redis 客户端
func NewRedisClient(cfg config.RedisConfig) (*RedisClient, error) {
	client := redis.NewClient(&redis.Options{
		Addr:         cfg.Addr(),
		Password:     cfg.Password,
		DB:           cfg.DB,
		PoolSize:     cfg.PoolSize,
		MinIdleConns: cfg.MinIdleConns,
		DialTimeout:  5 * time.Second,
		ReadTimeout:  3 * time.Second,
		WriteTimeout: 3 * time.Second,
	})

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("Redis 连接失败: %w", err)
	}

	return &RedisClient{client: client}, nil
}

// Client 返回底层 go-redis 客户端
func (r *RedisClient) Client() *redis.Client {
	return r.client
}

// Close 关闭连接
func (r *RedisClient) Close() error {
	return r.client.Close()
}

// Ping 健康检查
func (r *RedisClient) Ping(ctx context.Context) error {
	return r.client.Ping(ctx).Err()
}

// ---------------- 基础 KV 操作 ----------------

// Set 设置字符串缓存
func (r *RedisClient) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	return r.client.Set(ctx, key, value, expiration).Err()
}

// Get 获取字符串缓存
func (r *RedisClient) Get(ctx context.Context, key string) (string, error) {
	return r.client.Get(ctx, key).Result()
}

// GetBytes 获取字节数组缓存
func (r *RedisClient) GetBytes(ctx context.Context, key string) ([]byte, error) {
	return r.client.Get(ctx, key).Bytes()
}

// Del 删除缓存
func (r *RedisClient) Del(ctx context.Context, keys ...string) error {
	return r.client.Del(ctx, keys...).Err()
}

// Exists 判断 key 是否存在
func (r *RedisClient) Exists(ctx context.Context, key string) (bool, error) {
	n, err := r.client.Exists(ctx, key).Result()
	return n > 0, err
}

// Expire 设置过期时间
func (r *RedisClient) Expire(ctx context.Context, key string, expiration time.Duration) error {
	return r.client.Expire(ctx, key, expiration).Err()
}

// ---------------- Sorted Set 操作(离线消息等场景) ----------------

// ZAdd 向 Sorted Set 添加成员
func (r *RedisClient) ZAdd(ctx context.Context, key string, score float64, member interface{}) error {
	return r.client.ZAdd(ctx, key, redis.Z{Score: score, Member: member}).Err()
}

// ZRangeByScore 按分数范围获取 Sorted Set 成员
func (r *RedisClient) ZRangeByScore(ctx context.Context, key string, min, max string) ([]string, error) {
	return r.client.ZRangeByScore(ctx, key, &redis.ZRangeBy{Min: min, Max: max}).Result()
}

// ZRemRangeByScore 按分数范围删除 Sorted Set 成员
func (r *RedisClient) ZRemRangeByScore(ctx context.Context, key, min, max string) error {
	return r.client.ZRemRangeByScore(ctx, key, min, max).Err()
}

// ZCard 获取 Sorted Set 成员数
func (r *RedisClient) ZCard(ctx context.Context, key string) (int64, error) {
	return r.client.ZCard(ctx, key).Result()
}

// ---------------- SetNX(分布式锁底层) ----------------

// SetNX 仅在 key 不存在时设置
func (r *RedisClient) SetNX(ctx context.Context, key string, value interface{}, expiration time.Duration) (bool, error) {
	return r.client.SetNX(ctx, key, value, expiration).Result()
}

// Eval 执行 Lua 脚本
func (r *RedisClient) Eval(ctx context.Context, script string, keys []string, args ...interface{}) (interface{}, error) {
	return r.client.Eval(ctx, script, keys, args...).Result()
}
