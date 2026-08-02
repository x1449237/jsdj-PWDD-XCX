package model

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"time"
)

// JSONString 自定义 JSON 类型，实现 sql.Scanner / driver.Valuer
type JSONString json.RawMessage

// Value 实现 driver.Valuer
func (j JSONString) Value() (driver.Value, error) {
	if len(j) == 0 {
		return nil, nil
	}
	return json.RawMessage(j).MarshalJSON()
}

// Scan 实现 sql.Scanner
func (j *JSONString) Scan(value interface{}) error {
	bytes, ok := value.([]byte)
	if !ok {
		return errors.New("failed to unmarshal JSONB value")
	}
	result := json.RawMessage{}
	err := json.Unmarshal(bytes, &result)
	*j = JSONString(result)
	return err
}

// MarshalJSON 自定义 JSON 序列化
func (j JSONString) MarshalJSON() ([]byte, error) {
	if len(j) == 0 {
		return []byte("null"), nil
	}
	b, err := json.RawMessage(j).MarshalJSON()
	if err != nil {
		return []byte("null"), err
	}
	return b, nil
}

// UnmarshalJSON 自定义 JSON 反序列化
func (j *JSONString) UnmarshalJSON(data []byte) error {
	result := json.RawMessage{}
	err := json.Unmarshal(data, &result)
	*j = JSONString(result)
	return err
}

// UpMasterCertification UP主认证表模型
type UpMasterCertification struct {
	ID             int64       `gorm:"primaryKey;autoIncrement;column:id" json:"id"`
	UID            int64       `gorm:"column:uid;index:idx_uid;not null;default:0" json:"uid"`
	Platform       string      `gorm:"column:platform;index:idx_platform;size:32;not null;default:''" json:"platform"` // douyin/bilibili/kuaishou/xhs
	AccountName    string      `gorm:"column:account_name;size:128;not null;default:''" json:"account_name"`
	FollowerCount  int         `gorm:"column:follower_count;index:idx_follower_count;not null;default:0" json:"follower_count"`
	LinkUrl        string      `gorm:"column:link_url;size:512;not null;default:''" json:"link_url"`
	ScreenshotUrls JSONString  `gorm:"column:screenshot_urls;type:json" json:"screenshot_urls"`
	Status         int8        `gorm:"column:status;index:idx_status;not null;default:0" json:"status"` // 0=pending 1=approved 2=rejected 3=revoked
	Tier           int         `gorm:"column:tier;index:idx_tier;not null;default:0" json:"tier"`       // 1~6档
	VerifiedAt     *time.Time  `gorm:"column:verified_at" json:"verified_at"`
	ReviewerID     int64       `gorm:"column:reviewer_id;not null;default:0" json:"reviewer_id"`
	RejectReason   string      `gorm:"column:reject_reason;size:255;not null;default:''" json:"reject_reason"`
	CreatedAt      *time.Time  `gorm:"column:created_at;index:idx_created_at" json:"created_at"`
	UpdatedAt      *time.Time  `gorm:"column:updated_at" json:"updated_at"`
}

// TableName 指定表名
func (UpMasterCertification) TableName() string {
	return "up_master_certifications"
}

// UP主认证状态常量
const (
	UpMasterStatusPending  int8 = 0 // 待审核
	UpMasterStatusApproved int8 = 1 // 已通过
	UpMasterStatusRejected int8 = 2 // 已驳回
	UpMasterStatusRevoked  int8 = 3 // 已撤销
)

// UP主平台常量
const (
	UpMasterPlatformDouyin   = "douyin"
	UpMasterPlatformBilibili = "bilibili"
	UpMasterPlatformKuaishou = "kuaishou"
	UpMasterPlatformXHS      = "xhs"
)

// UpMasterTierConfig UP主档位配置表模型
type UpMasterTierConfig struct {
	ID           int         `gorm:"primaryKey;autoIncrement;column:id" json:"id"`
	TierName     string      `gorm:"column:tier_name;size:64;not null;default:''" json:"tier_name"`
	MinFollowers int         `gorm:"column:min_followers;not null;default:0" json:"min_followers"`
	MaxFollowers int         `gorm:"column:max_followers;not null;default:0" json:"max_followers"`
	BadgeIcon    string      `gorm:"column:badge_icon;size:512;not null;default:''" json:"badge_icon"`
	Benefits     JSONString  `gorm:"column:benefits;type:json" json:"benefits"`
}

// TableName 指定表名
func (UpMasterTierConfig) TableName() string {
	return "up_master_tier_configs"
}

// UpMasterLevelLog UP主升降级日志表模型
type UpMasterLevelLog struct {
	ID            int64      `gorm:"primaryKey;autoIncrement;column:id" json:"id"`
	UID           int64      `gorm:"column:uid;index:idx_uid;not null;default:0" json:"uid"`
	CertID        int64      `gorm:"column:cert_id;index:idx_cert_id;not null;default:0" json:"cert_id"`
	OldTier       int        `gorm:"column:old_tier;not null;default:0" json:"old_tier"`
	NewTier       int        `gorm:"column:new_tier;index:idx_new_tier;not null;default:0" json:"new_tier"`
	FollowerCount int        `gorm:"column:follower_count;not null;default:0" json:"follower_count"`
	ChangeType    string     `gorm:"column:change_type;size:16;not null;default:''" json:"change_type"` // upgrade/downgrade
	Reason        string     `gorm:"column:reason;size:255;not null;default:''" json:"reason"`
	CreatedAt     *time.Time `gorm:"column:created_at;index:idx_created_at" json:"created_at"`
}

// TableName 指定表名
func (UpMasterLevelLog) TableName() string {
	return "up_master_level_logs"
}

// GrayRelease 灰度发布配置表模型
type GrayRelease struct {
	ID             int64      `gorm:"primaryKey;autoIncrement;column:id" json:"id"`
	FeatureName    string     `gorm:"column:feature_name;uniqueIndex:uk_feature_name;size:64;not null;default:''" json:"feature_name"`
	RolloutPercent int        `gorm:"column:rollout_percent;not null;default:0" json:"rollout_percent"` // 0~100
	Whitelist      JSONString `gorm:"column:whitelist;type:json" json:"whitelist"`                     // []int64 uid列表
	Enabled        int8       `gorm:"column:enabled;index:idx_enabled;not null;default:1" json:"enabled"`
	Description    string     `gorm:"column:description;size:255;not null;default:''" json:"description"`
	CreatedAt      *time.Time `gorm:"column:created_at" json:"created_at"`
	UpdatedAt      *time.Time `gorm:"column:updated_at" json:"updated_at"`
}

// TableName 指定表名
func (GrayRelease) TableName() string {
	return "gray_releases"
}

// GrayConfig 灰度配置结构体(内存中使用)
type GrayConfig struct {
	Whitelist      []int64 `json:"whitelist"`
	RolloutPercent int     `json:"rollout_percent"`
}

// AntiBoostingRule 防代练规则表模型
type AntiBoostingRule struct {
	ID         int64      `gorm:"primaryKey;autoIncrement;column:id" json:"id"`
	Pattern    string     `gorm:"column:pattern;size:512;not null;default:''" json:"pattern"` // 正则表达式
	RuleType   string     `gorm:"column:rule_type;size:32;not null;default:''" json:"rule_type"`
	Enabled    int8       `gorm:"column:enabled;index:idx_enabled;not null;default:1" json:"enabled"`
	Severity   int8       `gorm:"column:severity;not null;default:1" json:"severity"` // 1=提示 2=拦截 3=高风险
	CreatedAt  *time.Time `gorm:"column:created_at" json:"created_at"`
	UpdatedAt  *time.Time `gorm:"column:updated_at" json:"updated_at"`
}

// TableName 指定表名
func (AntiBoostingRule) TableName() string {
	return "anti_boosting_rules"
}

// AntiBoostingLog 防代练命中日志表模型
type AntiBoostingLog struct {
	ID             int64      `gorm:"primaryKey;autoIncrement;column:id" json:"id"`
	UID            int64      `gorm:"column:uid;index:idx_uid;not null;default:0" json:"uid"`
	ContentType    int64      `gorm:"column:content_type;index:idx_content_type;not null;default:0" json:"content_type"` // 1=chat 2=order_desc 3=announcement 4=post
	MatchedPattern string     `gorm:"column:matched_pattern;size:512;not null;default:''" json:"matched_pattern"`
	Content        string     `gorm:"column:content;type:text" json:"content"`
	CreatedAt      *time.Time `gorm:"column:created_at;index:idx_created_at" json:"created_at"`
}

// TableName 指定表名
func (AntiBoostingLog) TableName() string {
	return "anti_boosting_logs"
}

// LotteryPrize 抽奖奖品表模型
type LotteryPrize struct {
	ID            int64      `gorm:"primaryKey;autoIncrement;column:id" json:"id"`
	ActivityID    int64      `gorm:"column:activity_id;index:idx_activity_id;not null;default:0" json:"activity_id"`
	PrizeName     string     `gorm:"column:prize_name;size:128;not null;default:''" json:"prize_name"`
	PrizeType     string     `gorm:"column:prize_type;size:32;not null;default:''" json:"prize_type"` // coupon/recharge/points/balance/thankyou
	PrizeValue    int64      `gorm:"column:prize_value;not null;default:0" json:"prize_value"`       // 券id/充值金额/积分/余额
	Probability   int        `gorm:"column:probability;not null;default:0" json:"probability"`       // 权重
	Stock         int        `gorm:"column:stock;not null;default:0" json:"stock"`                   // 库存
	DailyMax      int        `gorm:"column:daily_max;not null;default:0" json:"daily_max"`           // 活动每日最大抽奖次数
	DisplayOrder  int        `gorm:"column:display_order;not null;default:0" json:"display_order"`
	Status        int8       `gorm:"column:status;index:idx_status;not null;default:1" json:"status"`
	CreatedAt     *time.Time `gorm:"column:created_at" json:"created_at"`
	UpdatedAt     *time.Time `gorm:"column:updated_at" json:"updated_at"`
}

// TableName 指定表名
func (LotteryPrize) TableName() string {
	return "lottery_prizes"
}

// LotteryRecord 抽奖记录表模型
type LotteryRecord struct {
	ID         int64      `gorm:"primaryKey;autoIncrement;column:id" json:"id"`
	UID        int64      `gorm:"column:uid;index:idx_uid;not null;default:0" json:"uid"`
	ActivityID int64      `gorm:"column:activity_id;index:idx_activity_id;not null;default:0" json:"activity_id"`
	PrizeID    int64      `gorm:"column:prize_id;index:idx_prize_id;not null;default:0" json:"prize_id"`
	DrawIP     string     `gorm:"column:draw_ip;size:64;not null;default:''" json:"draw_ip"`
	IsWon      int8       `gorm:"column:is_won;index:idx_is_won;not null;default:0" json:"is_won"`
	Delivered  int8       `gorm:"column:delivered;index:idx_delivered;not null;default:0" json:"delivered"` // 是否已发奖
	CreatedAt  *time.Time `gorm:"column:created_at;index:idx_created_at" json:"created_at"`
}

// TableName 指定表名
func (LotteryRecord) TableName() string {
	return "lottery_records"
}

// 抽奖活动扩展字段
const (
	LotteryDailyMaxKey = "daily_max" // 活动每日最大抽奖次数配置在activity JSON或单独字段
)

// UserRechargeLog 用户充值日志模型(用于抽奖发充值券等)
type UserRechargeLog struct {
	ID          int64      `gorm:"primaryKey;autoIncrement;column:id" json:"id"`
	UserID      int64      `gorm:"column:user_id;index:idx_user_id;not null;default:0" json:"user_id"`
	Amount      int64      `gorm:"column:amount;not null;default:0" json:"amount"`
	Source      string     `gorm:"column:source;size:32;not null;default:''" json:"source"` // lottery/recharge/compensation
	RefID       int64      `gorm:"column:ref_id;index:idx_ref_id;not null;default:0" json:"ref_id"`
	BalanceBefore int64    `gorm:"column:balance_before;not null;default:0" json:"balance_before"`
	BalanceAfter  int64    `gorm:"column:balance_after;not null;default:0" json:"balance_after"`
	CreatedAt   *time.Time `gorm:"column:created_at;index:idx_created_at" json:"created_at"`
}

// TableName 指定表名
func (UserRechargeLog) TableName() string {
	return "user_recharge_logs"
}

// CircuitBreakerLog 熔断器日志表模型
type CircuitBreakerLog struct {
	ID          int64      `gorm:"primaryKey;autoIncrement;column:id" json:"id"`
	ServiceName string     `gorm:"column:service_name;index:idx_service_name;size:64;not null;default:''" json:"service_name"`
	State       string     `gorm:"column:state;size:16;not null;default:''" json:"state"` // closed/open/half-open
	FromState   string     `gorm:"column:from_state;size:16;not null;default:''" json:"from_state"`
	ToState     string     `gorm:"column:to_state;size:16;not null;default:''" json:"to_state"`
	Reason      string     `gorm:"column:reason;size:512;not null;default:''" json:"reason"`
	Counts      int64      `gorm:"column:counts;not null;default:0" json:"counts"`
	CreatedAt   *time.Time `gorm:"column:created_at;index:idx_created_at" json:"created_at"`
}

// TableName 指定表名
func (CircuitBreakerLog) TableName() string {
	return "circuit_breaker_logs"
}

// BackupRecord 备份记录表模型
type BackupRecord struct {
	ID           int64      `gorm:"primaryKey;autoIncrement;column:id" json:"id"`
	Name         string     `gorm:"column:name;uniqueIndex:uk_name;size:128;not null;default:''" json:"name"`
	BackupType   string     `gorm:"column:backup_type;size:32;not null;default:'manual'" json:"backup_type"` // auto/manual
	FileSize     int64      `gorm:"column:file_size;not null;default:0" json:"file_size"`
	OSSUrl       string     `gorm:"column:oss_url;size:512;not null;default:''" json:"oss_url"`
	Encrypted    int8       `gorm:"column:encrypted;not null;default:1" json:"encrypted"`
	Status       string     `gorm:"column:status;size:16;not null;default:'success'" json:"status"` // success/failed
	OperatorID   int64      `gorm:"column:operator_id;not null;default:0" json:"operator_id"`
	ErrorMessage string     `gorm:"column:error_message;size:512;not null;default:''" json:"error_message"`
	CreatedAt    *time.Time `gorm:"column:created_at;index:idx_created_at" json:"created_at"`
}

// TableName 指定表名
func (BackupRecord) TableName() string {
	return "backup_records"
}

// RestoreRecord 恢复记录表模型
type RestoreRecord struct {
	ID           int64      `gorm:"primaryKey;autoIncrement;column:id" json:"id"`
	BackupName   string     `gorm:"column:backup_name;index:idx_backup_name;size:128;not null;default:''" json:"backup_name"`
	BackupID     int64      `gorm:"column:backup_id;not null;default:0" json:"backup_id"`
	Status       string     `gorm:"column:status;size:16;not null;default:'success'" json:"status"`
	OperatorID   int64      `gorm:"column:operator_id;not null;default:0" json:"operator_id"`
	ErrorMessage string     `gorm:"column:error_message;size:512;not null;default:''" json:"error_message"`
	CreatedAt    *time.Time `gorm:"column:created_at;index:idx_created_at" json:"created_at"`
}

// TableName 指定表名
func (RestoreRecord) TableName() string {
	return "restore_records"
}

// 活体检测相关(在risk.go或user.go已有的，这里补充完整)
// FaceVerifyRateLimit 已在 schema.sql 中定义，对应 model 补充
type FaceVerifyRateLimit struct {
	ID        int64      `gorm:"primaryKey;autoIncrement;column:id" json:"id"`
	UserID    int64      `gorm:"column:user_id;index:idx_user_date;not null;default:0" json:"user_id"`
	IP        string     `gorm:"column:ip;index:idx_ip;size:64;not null;default:''" json:"ip"`
	Count     int        `gorm:"column:count;not null;default:0" json:"count"`
	Date      string     `gorm:"column:date;index:idx_user_date;size:16;not null;default:''" json:"date"` // yyyy-MM-dd
	Fee       int64      `gorm:"column:fee;not null;default:0" json:"fee"`                               // 本次活体认证扣费金额(分),0=未扣费(如缓存命中)
	CreatedAt *time.Time `gorm:"column:created_at" json:"created_at"`
	UpdatedAt *time.Time `gorm:"column:updated_at" json:"updated_at"`
}

// TableName 指定表名
func (FaceVerifyRateLimit) TableName() string {
	return "face_verify_rate_limits"
}

// RealnameCache 活体检测缓存表模型
type RealnameCache struct {
	ID             int64      `gorm:"primaryKey;autoIncrement;column:id" json:"id"`
	UserID         int64      `gorm:"column:user_id;index:idx_user_id;not null;default:0" json:"user_id"`
	LastVerifyTime *time.Time `gorm:"column:last_verify_time" json:"last_verify_time"`
	ExpireTime     *time.Time `gorm:"column:expire_time;index:idx_expire_time" json:"expire_time"`
	VerifySession  string     `gorm:"column:verify_session;size:128;not null;default:''" json:"verify_session"`
	CreatedAt      *time.Time `gorm:"column:created_at" json:"created_at"`
	UpdatedAt      *time.Time `gorm:"column:updated_at" json:"updated_at"`
}

// TableName 指定表名
func (RealnameCache) TableName() string {
	return "realname_caches"
}

// OperationLog 已在 system.go 定义
// RiskUser 已在 risk.go 定义

// 给 LotteryActivity 补充 DailyMax 字段(通过扩展方式，不破坏原 model)
// 通过 JSON 扩展实现，或在实际业务中使用单独表关联
