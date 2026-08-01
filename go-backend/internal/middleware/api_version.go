package middleware

import (
	"context"
	"encoding/json"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"github.com/jisan/e-sports-platform/internal/model"
	"github.com/jisan/e-sports-platform/internal/utils"
	"github.com/jisan/e-sports-platform/pkg/cache"
)

const (
	APIVersionV1 = "v1"
	APIVersionV2 = "v2"
)

const (
	GrayFeatureDefault = "api_v2_default"
	grayCacheKey       = "jisan:gray:config:default"
	grayCacheTTL       = 5 * time.Minute
)

var (
	grayConfigMu   sync.RWMutex
	cachedGrayCfg  = &model.GrayConfig{Whitelist: []int64{}, RolloutPercent: 0}
	grayDB         *gorm.DB
	grayCache      *cache.Cache
	grayRedis      *cache.RedisClient
	grayInited     bool
)

// InitGrayMiddleware 初始化灰度中间件依赖(在service初始化后调用)
func InitGrayMiddleware(db *gorm.DB, cc *cache.Cache, rc *cache.RedisClient) {
	grayConfigMu.Lock()
	defer grayConfigMu.Unlock()
	grayDB = db
	grayCache = cc
	grayRedis = rc
	grayInited = true
	_ = loadGrayConfigFromDB()
}

// loadGrayConfigFromDB 从DB或Redis加载灰度配置
func loadGrayConfigFromDB() *model.GrayConfig {
	cfg := &model.GrayConfig{Whitelist: []int64{}, RolloutPercent: 0}
	if !grayInited || grayDB == nil {
		return cfg
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	// 1. 尝试从Redis缓存读取
	if grayRedis != nil {
		raw, err := grayRedis.Get(ctx, grayCacheKey)
		if err == nil && raw != "" {
			if jerr := json.Unmarshal([]byte(raw), cfg); jerr == nil && len(cfg.Whitelist) >= 0 {
				grayConfigMu.Lock()
				cachedGrayCfg = cfg
				grayConfigMu.Unlock()
				return cfg
			}
		}
	}

	// 2. 优先从 GrayRelease 表读取
	var gr model.GrayRelease
	err := grayDB.Where("feature_name = ? AND enabled = 1", GrayFeatureDefault).First(&gr).Error
	if err == nil {
		cfg.RolloutPercent = gr.RolloutPercent
		if len(gr.Whitelist) > 0 {
			_ = json.Unmarshal(gr.Whitelist, &cfg.Whitelist)
		}
	} else {
		// 3. 兜底从 system_configs 表 JSON 读取
		var sc model.SystemConfig
		er2 := grayDB.Where("`key` = ?", "gray_release_config").First(&sc).Error
		if er2 == nil && sc.Value != "" {
			_ = json.Unmarshal([]byte(sc.Value), cfg)
		} else {
			// 4. 旧格式分开读 rollout/whitelist
			var sc1, sc2 model.SystemConfig
			if er3 := grayDB.Where("`key` = ?", "gray_rollout_percent").First(&sc1).Error; er3 == nil {
				n := 0
				if _, serr := jsonUnmarshalFallback([]byte(sc1.Value), &n); serr == nil {
					cfg.RolloutPercent = n
				}
			}
			if er4 := grayDB.Where("`key` = ?", "gray_whitelist").First(&sc2).Error; er4 == nil {
				_ = json.Unmarshal([]byte(sc2.Value), &cfg.Whitelist)
			}
		}
	}

	if cfg.RolloutPercent < 0 {
		cfg.RolloutPercent = 0
	}
	if cfg.RolloutPercent > 100 {
		cfg.RolloutPercent = 100
	}
	if cfg.Whitelist == nil {
		cfg.Whitelist = []int64{}
	}

	// 写回 Redis 缓存
	if grayRedis != nil {
		if b, merr := json.Marshal(cfg); merr == nil {
			_ = grayRedis.Set(ctx, grayCacheKey, string(b), grayCacheTTL)
		}
	}

	grayConfigMu.Lock()
	cachedGrayCfg = cfg
	grayConfigMu.Unlock()
	return cfg
}

// ReloadGrayConfig 触发灰度配置重新加载(在admin更新灰度后调用)
func ReloadGrayConfig() error {
	if grayRedis != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()
		_ = grayRedis.Del(ctx, grayCacheKey)
	}
	loadGrayConfigFromDB()
	return nil
}

// GetGrayConfig 获取当前灰度配置
func GetGrayConfig() *model.GrayConfig {
	grayConfigMu.RLock()
	defer grayConfigMu.RUnlock()
	// 深拷贝返回，避免外部修改
	cfg := &model.GrayConfig{
		RolloutPercent: cachedGrayCfg.RolloutPercent,
		Whitelist:      make([]int64, len(cachedGrayCfg.Whitelist)),
	}
	copy(cfg.Whitelist, cachedGrayCfg.Whitelist)
	return cfg
}

// jsonUnmarshalFallback 兼容纯数字字符串的反序列化
func jsonUnmarshalFallback(b []byte, n *int) (interface{}, error) {
	if len(b) == 0 {
		*n = 0
		return nil, nil
	}
	err := json.Unmarshal(b, n)
	if err == nil {
		return nil, nil
	}
	// 纯数字字符串
	var s string
	if err2 := json.Unmarshal(b, &s); err2 == nil {
		if p, perr := parseIntString(s); perr == nil {
			*n = p
			return nil, nil
		}
	}
	// 直接解析字符串
	if p, perr := parseIntString(string(b)); perr == nil {
		*n = p
		return nil, nil
	}
	*n = 0
	return nil, nil
}

func parseIntString(s string) (int, error) {
	n := 0
	mult := 1
	for i := len(s) - 1; i >= 0; i-- {
		c := s[i]
		if c < '0' || c > '9' {
			if i == 0 && c == '-' {
				n = -n
				break
			}
			continue
		}
		n += int(c-'0') * mult
		mult *= 10
	}
	return n, nil
}

type APIVersionMiddleware struct {
	defaultVersion string
}

func NewAPIVersionMiddleware(defaultVersion string) *APIVersionMiddleware {
	if defaultVersion == "" {
		defaultVersion = APIVersionV1
	}
	return &APIVersionMiddleware{defaultVersion: defaultVersion}
}

// SetGrayConfig 兼容旧API(线程安全)
func SetGrayConfig(whitelist []int64, rolloutPercent int) {
	grayConfigMu.Lock()
	defer grayConfigMu.Unlock()
	cachedGrayCfg = &model.GrayConfig{
		Whitelist:      append([]int64{}, whitelist...),
		RolloutPercent: rolloutPercent,
	}
	if cachedGrayCfg.RolloutPercent < 0 {
		cachedGrayCfg.RolloutPercent = 0
	}
	if cachedGrayCfg.RolloutPercent > 100 {
		cachedGrayCfg.RolloutPercent = 100
	}
}

func (m *APIVersionMiddleware) APIVersion() gin.HandlerFunc {
	return func(c *gin.Context) {
		version := c.GetHeader("X-Api-Version")
		if version == "" {
			version = m.defaultVersion
		}

		if c.GetHeader("X-Api-Version") == "" {
			userID := utils.GetCurrentUserID(c)
			if m.shouldGrayV2(userID) {
				version = APIVersionV2
			}
		}

		if version != APIVersionV1 && version != APIVersionV2 {
			version = m.defaultVersion
		}

		c.Set(ContextKeyAPIVersion, version)
		c.Header("X-Api-Version", version)

		c.Next()
	}
}

func (m *APIVersionMiddleware) shouldGrayV2(userID int64) bool {
	// 懒加载配置(如果还没初始化)
	var cfg *model.GrayConfig
	grayConfigMu.RLock()
	cfg = cachedGrayCfg
	grayConfigMu.RUnlock()
	if cfg == nil {
		cfg = loadGrayConfigFromDB()
	}

	if userID <= 0 {
		return false
	}

	// 白名单优先
	for _, wid := range cfg.Whitelist {
		if wid == userID {
			return true
		}
	}
	// 基于 uid hash 判断百分比，保证同一用户稳定命中
	if cfg.RolloutPercent > 0 {
		mod := int(hashUID(userID) % 100)
		return mod < cfg.RolloutPercent
	}
	return false
}

// hashUID 简单 hash，使用 FNV-1a 变体使分布更均匀
func hashUID(uid int64) uint64 {
	h := uint64(1469598103934665603)
	for i := 0; i < 8; i++ {
		h ^= uint64((uid >> (i * 8)) & 0xff)
		h *= 1099511628211
	}
	return h
}

const ContextKeyAPIVersion = "api_version"

func GetAPIVersion(c *gin.Context) string {
	if v, exists := c.Get(ContextKeyAPIVersion); exists {
		if s, ok := v.(string); ok {
			return s
		}
	}
	return APIVersionV1
}
