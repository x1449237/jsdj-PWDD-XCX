package service

import (
	"regexp"
	"sync"
	"time"

	"github.com/jisan/e-sports-platform/internal/model"
)

// 防代练规则内存缓存
var (
	antiBoostingMu       sync.RWMutex
	antiBoostingPatterns []compiledRule
	antiBoostingLoadedAt time.Time
	antiBoostingCacheDur = 5 * time.Minute
)

type compiledRule struct {
	id      int64
	pattern *regexp.Regexp
	raw     string
	severity int8
}

// 内容类型常量
const (
	AntiBoostingContentTypeChat         int64 = 1 // 聊天消息
	AntiBoostingContentTypeOrderDesc    int64 = 2 // 订单描述
	AntiBoostingContentTypeAnnouncement int64 = 3 // 俱乐部/公告
	AntiBoostingContentTypePost         int64 = 4 // 用户动态
)

// reloadAntiBoostingRules 懒加载防代练规则(带缓存有效期)
func reloadAntiBoostingRules() []compiledRule {
	antiBoostingMu.RLock()
	loaded := antiBoostingLoadedAt
	hasCached := len(antiBoostingPatterns) > 0
	cached := make([]compiledRule, len(antiBoostingPatterns))
	copy(cached, antiBoostingPatterns)
	antiBoostingMu.RUnlock()

	now := time.Now()
	if hasCached && now.Sub(loaded) < antiBoostingCacheDur {
		return cached
	}

	// 重新加载
	var rules []model.AntiBoostingRule
	if err := db.Where("enabled = 1").Find(&rules).Error; err != nil {
		// 加载失败返回旧缓存
		return cached
	}

	newCompiled := make([]compiledRule, 0, len(rules))
	for _, r := range rules {
		if r.Pattern == "" {
			continue
		}
		reg, cerr := regexp.Compile(r.Pattern)
		if cerr != nil {
			// 编译失败的正则跳过
			continue
		}
		newCompiled = append(newCompiled, compiledRule{
			id:       r.ID,
			pattern:  reg,
			raw:      r.Pattern,
			severity: r.Severity,
		})
	}

	antiBoostingMu.Lock()
	antiBoostingPatterns = newCompiled
	antiBoostingLoadedAt = now
	antiBoostingMu.Unlock()
	return newCompiled
}

// InvalidateAntiBoostingCache 强制清空防代练规则缓存(在管理端增删改规则后调用)
func InvalidateAntiBoostingCache() {
	antiBoostingMu.Lock()
	antiBoostingPatterns = nil
	antiBoostingLoadedAt = time.Time{}
	antiBoostingMu.Unlock()
}

// CheckContentAntiBoosting 防代练内容检测
// contentType: 1 chat 2 order_desc 3 announcement 4 post
// 返回: hit 是否命中, patterns 命中的规则列表, err 错误
func CheckContentAntiBoosting(contentType, uid int64, text string) (bool, []string, error) {
	if text == "" {
		return false, nil, nil
	}
	rules := reloadAntiBoostingRules()
	if len(rules) == 0 {
		return false, nil, nil
	}

	hitPatterns := make([]string, 0, 2)
	hasHit := false
	for _, r := range rules {
		if r.pattern.MatchString(text) {
			hitPatterns = append(hitPatterns, r.raw)
			hasHit = true
			// severity>=2 且命中至少 1 条就先记录，继续收集其他命中
		}
	}

	if hasHit {
		now := nowTimePtr()
		for _, p := range hitPatterns {
			_ = db.Create(&model.AntiBoostingLog{
				UID:            uid,
				ContentType:    contentType,
				MatchedPattern: p,
				Content:        text,
				CreatedAt:      now,
			}).Error
		}

		// 24h 内命中超过 3 次自动标记 risk_user
		dayAgo := time.Now().Add(-24 * time.Hour)
		var cnt int64
		_ = db.Model(&model.AntiBoostingLog{}).
			Where("uid = ? AND created_at >= ?", uid, dayAgo).
			Count(&cnt).Error
		if cnt >= 3 {
			// 查询是否已有风险标记
			var existing model.RiskUser
			err := db.Where("user_id = ? AND risk_type = ?", uid, "anti_boosting").First(&existing).Error
			if err != nil {
				// 不存在则创建
				_ = db.Create(&model.RiskUser{
					UserID:    uid,
					RiskLevel: model.RiskLevelHigh,
					RiskType:  "anti_boosting",
					MarkedAt:  now,
					CreatedAt: now,
					UpdatedAt: now,
				}).Error
			} else {
				// 存在则升级风险等级
				_ = db.Model(&existing).Updates(map[string]interface{}{
					"risk_level": model.RiskLevelCritical,
					"marked_at":  now,
					"updated_at": now,
				}).Error
			}
		}
	}

	return hasHit, hitPatterns, nil
}

// AntiBoostingHitCount 用户 24h 内命中次数(供查询)
func AntiBoostingHitCount(uid int64) int64 {
	dayAgo := time.Now().Add(-24 * time.Hour)
	var cnt int64
	_ = db.Model(&model.AntiBoostingLog{}).
		Where("uid = ? AND created_at >= ?", uid, dayAgo).
		Count(&cnt).Error
	return cnt
}
