package cache

import "fmt"

// 缓存键前缀
const (
	keyPrefix       = "jisan:"
	prefixUser      = "user:"       // 用户信息
	prefixSession   = "session:"    // 会话/登录态
	prefixConfig    = "config:"     // 系统配置
	prefixClub      = "club:"       // 俱乐部信息
	prefixOrder     = "order:"      // 订单信息
	prefixPlayer    = "player:"     // 打手信息
	prefixRateLimit = "ratelimit:"  // 限流计数
	prefixLock      = "lock:"       // 分布式锁
	prefixOffline   = "offline:"    // 离线消息(Sorted Set)
	prefixWS        = "ws:"         // WebSocket 连接
	prefixSensitive = "sensitive:"  // 敏感词
	prefixCaptcha   = "captcha:"    // 验证码
	prefixToken     = "token:"      // token 黑名单
)

// 缓存默认过期时间
const (
	TTLShort      = 60 * 5        // 短时缓存 5 分钟
	TTLNormal     = 60 * 30       // 普通缓存 30 分钟
	TTLLong       = 60 * 60 * 2   // 长时缓存 2 小时
	TTLVerifyCode = 60 * 5        // 验证码 5 分钟
	TTLOfflineMsg = 60 * 60 * 24 * 7 // 离线消息 7 天
)

// ---------------- 用户相关 ----------------

// UserKey 用户信息缓存键(热点数据)
func UserKey(userID int64) string {
	return fmt.Sprintf("%s%s%d", keyPrefix, prefixUser, userID)
}

// UserTokenKey 用户 token 缓存键
func UserTokenKey(userID int64) string {
	return fmt.Sprintf("%s%stoken:%d", keyPrefix, prefixUser, userID)
}

// UserSessionKey 用户会话缓存键
func UserSessionKey(sessionID string) string {
	return fmt.Sprintf("%s%s%s", keyPrefix, prefixSession, sessionID)
}

// ---------------- 俱乐部相关 ----------------

// ClubKey 俱乐部信息缓存键
func ClubKey(clubID int64) string {
	return fmt.Sprintf("%s%s%d", keyPrefix, prefixClub, clubID)
}

// ClubAbbreviationKey 俱乐部缩写缓存键
func ClubAbbreviationKey(abbreviation string) string {
	return fmt.Sprintf("%s%sabbr:%s", keyPrefix, prefixClub, abbreviation)
}

// ---------------- 订单相关 ----------------

// OrderKey 订单信息缓存键
func OrderKey(orderID int64) string {
	return fmt.Sprintf("%s%s%d", keyPrefix, prefixOrder, orderID)
}

// OrderNoKey 订单号到订单ID的映射缓存键
func OrderNoKey(orderNo string) string {
	return fmt.Sprintf("%s%sno:%s", keyPrefix, prefixOrder, orderNo)
}

// ---------------- 打手相关 ----------------

// PlayerKey 打手信息缓存键
func PlayerKey(playerID int64) string {
	return fmt.Sprintf("%s%s%d", keyPrefix, prefixPlayer, playerID)
}

// ---------------- 系统配置 ----------------

// SystemConfigKey 系统配置缓存键(永久缓存)
func SystemConfigKey(key string) string {
	return fmt.Sprintf("%s%s%s", keyPrefix, prefixConfig, key)
}

// ---------------- 限流 ----------------

// RateLimitKey 接口限流计数键
func RateLimitKey(userID int64, route string) string {
	return fmt.Sprintf("%s%s%d:%s", keyPrefix, prefixRateLimit, userID, route)
}

// RateLimitIPKey 基于 IP 的限流计数键
func RateLimitIPKey(ip, route string) string {
	return fmt.Sprintf("%s%sip:%s:%s", keyPrefix, prefixRateLimit, ip, route)
}

// ---------------- 分布式锁 ----------------

// LockKey 分布式锁键
func LockKey(business string, id int64) string {
	return fmt.Sprintf("%s%s%s:%d", keyPrefix, prefixLock, business, id)
}

// ---------------- 离线消息 ----------------

// OfflineMsgKey 离线消息 Sorted Set 键(以时间戳为 score)
func OfflineMsgKey(userID int64) string {
	return fmt.Sprintf("%s%s%d", keyPrefix, prefixOffline, userID)
}

// ---------------- WebSocket ----------------

// WSConnKey WebSocket 用户在线连接键(记录连接ID)
func WSConnKey(userID int64) string {
	return fmt.Sprintf("%s%sonline:%d", keyPrefix, prefixWS, userID)
}

// WSConnCountKey WebSocket 全局连接数计数键
func WSConnCountKey() string {
	return fmt.Sprintf("%s%sconn_count", keyPrefix, prefixWS)
}

// ---------------- 敏感词 ----------------

// SensitiveWordsKey 敏感词库缓存键
func SensitiveWordsKey() string {
	return fmt.Sprintf("%s%sall", keyPrefix, prefixSensitive)
}

// ---------------- 验证码 ----------------

// CaptchaKey 验证码缓存键
func CaptchaKey(scene, target string) string {
	return fmt.Sprintf("%s%s%s:%s", keyPrefix, prefixCaptcha, scene, target)
}

// ---------------- Token 黑名单 ----------------

// TokenBlacklistKey token 黑名单键
func TokenBlacklistKey(tokenID string) string {
	return fmt.Sprintf("%s%s%s", keyPrefix, prefixToken, tokenID)
}
