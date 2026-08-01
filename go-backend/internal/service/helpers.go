package service

import (
	"context"
	"crypto/aes"
	"crypto/cipher"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	redislib "github.com/redis/go-redis/v9"
	"go.uber.org/zap"

	"github.com/jisan/e-sports-platform/internal/utils"
)

// nowTimePtr 返回当前时间的指针(供写入 created_at 等字段)
func nowTimePtr() *time.Time {
	now := time.Now()
	return &now
}

// contextWithTimeout 创建一个 3 秒超时的 context(简化 Redis 调用)
func contextWithTimeout() (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.Background(), 3*time.Second)
}

// itoa int64 转字符串
func itoa(n int64) string {
	return strconv.FormatInt(n, 10)
}

// atoi 字符串转 int64
func atoi(s string) int64 {
	n, _ := strconv.ParseInt(s, 10, 64)
	return n
}

// fmtSscanf 包装 fmt.Sscanf(供解析整数)
func fmtSscanf(s string, v *int64) (int, error) {
	return fmt.Sscanf(s, "%d", v)
}

// cacheKey 构造带统一前缀的 Redis 缓存键
func cacheKey(suffix string) string {
	return "jisan:" + suffix
}

// zapField 包装 zap.String 字段(简化日志调用)
func zapField(key, val string) zap.Field {
	return zap.String(key, val)
}

// newTraceID 生成一个新的 trace ID(UUID 去横线)
func newTraceID() string {
	return strings.ReplaceAll(uuid.New().String(), "-", "")
}

// DecryptWxPhoneNumber 解密微信小程序手机号(AES-128-CBC)
// encData 为 base64 编码的加密数据，iv 为 base64 的初始向量，sessionKey 为 base64 的会话密钥
// 微信手机号加密格式: AES-128-CBC，输出为 JSON {phoneNumber: "..."}
func DecryptWxPhoneNumber(encData, iv, sessionKey string) (string, error) {
	keyBytes, err := base64.StdEncoding.DecodeString(sessionKey)
	if err != nil {
		return "", fmt.Errorf("sessionKey base64 解码失败: %w", err)
	}
	ivBytes, err := base64.StdEncoding.DecodeString(iv)
	if err != nil {
		return "", fmt.Errorf("iv base64 解码失败: %w", err)
	}
	cipherBytes, err := base64.StdEncoding.DecodeString(encData)
	if err != nil {
		return "", fmt.Errorf("encData base64 解码失败: %w", err)
	}

	block, err := aes.NewCipher(keyBytes)
	if err != nil {
		return "", err
	}
	if len(cipherBytes)%block.BlockSize() != 0 {
		return "", errors.New("密文长度不是块大小的整数倍")
	}
	decrypted := make([]byte, len(cipherBytes))
	mode := cipher.NewCBCDecrypter(block, ivBytes)
	mode.CryptBlocks(decrypted, cipherBytes)

	// 去除 PKCS#7 填充
	padLen := int(decrypted[len(decrypted)-1])
	if padLen <= 0 || padLen > block.BlockSize() {
		return "", errors.New("PKCS7 填充长度异常")
	}
	decrypted = decrypted[:len(decrypted)-padLen]

	// 解析 JSON 提取手机号
	var payload struct {
		PhoneNumber     string `json:"phoneNumber"`
		PurePhoneNumber string `json:"purePhoneNumber"`
		CountryCode     string `json:"countryCode"`
	}
	if err := jsonUnmarshal(decrypted, &payload); err != nil {
		return "", fmt.Errorf("解析手机号 JSON 失败: %w", err)
	}
	if payload.PhoneNumber != "" {
		return payload.PhoneNumber, nil
	}
	return payload.PurePhoneNumber, nil
}

// jsonUnmarshal 包装 json.Unmarshal(避免在多处重复 import json)
func jsonUnmarshal(data []byte, v interface{}) error {
	return json.Unmarshal(data, v)
}

// jsonMarshal 包装 json.Marshal
func jsonMarshal(v interface{}) ([]byte, error) {
	return json.Marshal(v)
}

// hashPasswordUtil 包装 utils.HashPassword(避免各 service 重复 import utils)
func hashPasswordUtil(pwd string) (string, error) {
	return utils.HashPassword(pwd)
}

// checkPasswordUtil 包装 utils.CheckPassword
func checkPasswordUtil(pwd, hash string) bool {
	return utils.CheckPassword(pwd, hash)
}

// parseBoolStr 将 "1"/"true" 等字符串解析为 bool
func parseBoolStr(s string) bool {
	return s == "1" || strings.EqualFold(s, "true") || strings.EqualFold(s, "yes")
}

// redisGet 简化 Redis Get 调用，缓存不存在时返回空串
func redisGet(ctx context.Context, key string) string {
	if redis == nil {
		return ""
	}
	val, err := redis.Get(ctx, key)
	if err != nil && !errors.Is(err, redisNil) {
		return ""
	}
	return val
}

// redisNil 本地 redis.Nil 引用
var redisNil = redislib.Nil

// absInt64 取绝对值
func absInt64(n int64) int64 {
	if n < 0 {
		return -n
	}
	return n
}
