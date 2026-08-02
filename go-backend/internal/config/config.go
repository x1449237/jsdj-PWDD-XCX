package config

import (
	"fmt"
	"sync"
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/spf13/viper"
	"go.uber.org/zap"
)

// Config 全局配置根结构体
type Config struct {
	App       AppConfig       `mapstructure:"app"`
	MySQL     MySQLConfig     `mapstructure:"mysql"`
	Redis     RedisConfig     `mapstructure:"redis"`
	JWT       JWTConfig       `mapstructure:"jwt"`
	OSS       OSSConfig       `mapstructure:"oss"`
	WeChat    WeChatConfig    `mapstructure:"wechat"`
	WebSocket WSConfig        `mapstructure:"websocket"`
	Log       LogConfig       `mapstructure:"log"`
	RateLimit RateLimitConfig `mapstructure:"rate_limit"`
	SMTP      SMTPConfig      `mapstructure:"smtp"`
}

// SMTPConfig 邮件SMTP配置
type SMTPConfig struct {
	Host     string `mapstructure:"host"`
	Port     int    `mapstructure:"port"`
	User     string `mapstructure:"user"`
	Password string `mapstructure:"pass"`
	From     string `mapstructure:"from"`
	Sandbox  bool   `mapstructure:"sandbox"`
}

// AppConfig 应用基础配置
type AppConfig struct {
	Name        string `mapstructure:"name"`         // 应用名称
	Env         string `mapstructure:"env"`          // 运行环境 dev/test/prod
	Port        int    `mapstructure:"port"`         // HTTP 服务端口
	PprofPort   int    `mapstructure:"pprof_port"`   // pprof 监控端口
	ReadTimeout int    `mapstructure:"read_timeout"` // 读超时(秒)
	WriteTimeout int   `mapstructure:"write_timeout"` // 写超时(秒)
	GracefulTimeout int `mapstructure:"graceful_timeout"` // 优雅关闭等待时间(秒)

	// HTTPS 配置(启用后服务以 HTTPS 方式启动，不启用则回退 HTTP)
	TLSEnable  bool   `mapstructure:"tls_enable"`   // 是否启用 HTTPS
	TLSCert    string `mapstructure:"tls_cert"`     // 证书文件路径(.crt/.pem)
	TLSKey     string `mapstructure:"tls_key"`      // 私钥文件路径(.key)
	TLSPort    int    `mapstructure:"tls_port"`     // HTTPS 监听端口(为0则复用 port)
}

// IsTLS 是否启用 HTTPS
func (a AppConfig) IsTLS() bool {
	return a.TLSEnable && a.TLSCert != "" && a.TLSKey != ""
}

// MySQLConfig MySQL 数据库配置(支持读写分离)
type MySQLConfig struct {
	Host            string `mapstructure:"host"`
	Port            int    `mapstructure:"port"`
	Username        string `mapstructure:"username"`
	Password        string `mapstructure:"password"`
	Database        string `mapstructure:"database"`
	Charset         string `mapstructure:"charset"`
	MaxIdleConns    int    `mapstructure:"max_idle_conns"`
	MaxOpenConns    int    `mapstructure:"max_open_conns"`
	ConnMaxLifetime int    `mapstructure:"conn_max_lifetime"` // 连接最大存活时间(秒)
	LogLevel        string `mapstructure:"log_level"`         // silent/error/warn/info

	// 读写分离配置
	ReadHost string `mapstructure:"read_host"` // 只读库地址(为空则不启用读写分离)
	ReadPort int    `mapstructure:"read_port"`
}

// DSN 返回主库连接 DSN
func (m MySQLConfig) DSN() string {
	return fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=%s&parseTime=True&loc=Local&timeout=5s",
		m.Username, m.Password, m.Host, m.Port, m.Database, m.Charset)
}

// ReadDSN 返回只读库连接 DSN(未配置则返回主库 DSN)
func (m MySQLConfig) ReadDSN() string {
	host, port := m.Host, m.Port
	if m.ReadHost != "" {
		host, port = m.ReadHost, m.ReadPort
	}
	return fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=%s&parseTime=True&loc=Local&timeout=5s",
		m.Username, m.Password, host, port, m.Database, m.Charset)
}

// RedisConfig Redis 配置
type RedisConfig struct {
	Host         string `mapstructure:"host"`
	Port         int    `mapstructure:"port"`
	Password     string `mapstructure:"password"`
	DB           int    `mapstructure:"db"`
	PoolSize     int    `mapstructure:"pool_size"`
	MinIdleConns int    `mapstructure:"min_idle_conns"`
}

// Addr 返回 Redis 地址 host:port
func (r RedisConfig) Addr() string {
	return fmt.Sprintf("%s:%d", r.Host, r.Port)
}

// JWTConfig JWT 配置
type JWTConfig struct {
	Secret        string `mapstructure:"secret"`         // 签名密钥
	ExpireHours   int    `mapstructure:"expire_hours"`   // 普通用户 token 有效期(小时)
	AdminExpireHours int `mapstructure:"admin_expire_hours"` // 管理员 token 有效期(小时)
	Issuer        string `mapstructure:"issuer"`         // 签发者
}

// OSSConfig 对象存储配置
type OSSConfig struct {
	Provider   string `mapstructure:"provider"`   // 服务商 aliyun/tencent
	Endpoint   string `mapstructure:"endpoint"`   // 接入点
	AccessKey  string `mapstructure:"access_key"` // AccessKey
	SecretKey  string `mapstructure:"secret_key"` // SecretKey
	Bucket     string `mapstructure:"bucket"`     // 存储桶
	Region     string `mapstructure:"region"`     // 区域
	CDNDomain  string `mapstructure:"cdn_domain"` // CDN 加速域名
	SignExpire int    `mapstructure:"sign_expire"` // 签名 URL 有效期(秒)
}

// WeChatConfig 微信小程序配置
type WeChatConfig struct {
	AppID           string `mapstructure:"app_id"`
	AppSecret       string `mapstructure:"app_secret"`
	MchID           string `mapstructure:"mch_id"`           // 商户号
	MchKey          string `mapstructure:"mch_key"`          // 商户密钥
	NotifyURL       string `mapstructure:"notify_url"`       // 支付回调地址
	ApiV3Key        string `mapstructure:"api_v3_key"`       // APIv3 密钥(用于回调解密)
	PlatformCertPath string `mapstructure:"platform_cert_path"` // 微信平台证书路径(用于回调验签)
	SerialNo        string `mapstructure:"serial_no"`        // 商户证书序列号
}

// WSConfig WebSocket 配置
type WSConfig struct {
	HeartbeatSeconds     int `mapstructure:"heartbeat_seconds"`     // 心跳间隔(秒)
	TimeoutSeconds       int `mapstructure:"timeout_seconds"`       // 连接超时(秒)
	MaxConnPerUser       int `mapstructure:"max_conn_per_user"`     // 单用户最大连接数
	ProbeIntervalSeconds int `mapstructure:"probe_interval_seconds"` // 活跃度探针间隔(秒)
	ProbeMaxMiss         int `mapstructure:"probe_max_miss"`        // 连续无响应次数阈值
	WriteBufferSize      int `mapstructure:"write_buffer_size"`     // 写缓冲区大小
	ReadBufferSize       int `mapstructure:"read_buffer_size"`      // 读缓冲区大小
	AlertThreshold       int `mapstructure:"alert_threshold"`       // 连接数告警阈值(%)
}

// LogConfig 日志配置
type LogConfig struct {
	Level       string `mapstructure:"level"`         // debug/info/warn/error
	Encoding    string `mapstructure:"encoding"`      // json/console
	Directory   string `mapstructure:"directory"`     // 日志目录
	MaxSize     int    `mapstructure:"max_size"`      // 单文件最大大小(MB)
	MaxBackups  int    `mapstructure:"max_backups"`   // 保留备份数
	MaxAge      int    `mapstructure:"max_age"`       // 保留天数
}

// RateLimitConfig 限流配置(按角色差异化 QPS)
type RateLimitConfig struct {
	UserQPS    int `mapstructure:"user_qps"`    // 普通用户 QPS
	PlayerQPS  int `mapstructure:"player_qps"`  // 打手 QPS
	AdminQPS   int `mapstructure:"admin_qps"`   // 管理员 QPS
	Burst      int `mapstructure:"burst"`       // 突发桶容量
}

var (
	globalConfig *Config
	once         sync.Once
	mu           sync.RWMutex
)

// Load 加载配置文件，支持热加载
func Load(path string, logger *zap.Logger) (*Config, error) {
	v := viper.New()
	v.SetConfigFile(path)
	v.SetConfigType("yaml")

	// 设置默认值
	setDefaults(v)

	if err := v.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("读取配置文件失败: %w", err)
	}

	cfg := &Config{}
	if err := v.Unmarshal(cfg); err != nil {
		return nil, fmt.Errorf("解析配置文件失败: %w", err)
	}

	once.Do(func() {
		globalConfig = cfg
	})
	mu.Lock()
	globalConfig = cfg
	mu.Unlock()

	// 配置热加载
	v.OnConfigChange(func(e fsnotify.Event) {
		logger.Info("检测到配置文件变更，开始热加载", zap.String("file", e.Name))
		newCfg := &Config{}
		if err := v.Unmarshal(newCfg); err != nil {
			logger.Error("配置热加载失败", zap.Error(err))
			return
		}
		mu.Lock()
		globalConfig = newCfg
		mu.Unlock()
		logger.Info("配置热加载完成")
	})
	v.WatchConfig()

	return cfg, nil
}

// GetGlobal 获取全局配置(线程安全)
func GetGlobal() *Config {
	mu.RLock()
	defer mu.RUnlock()
	return globalConfig
}

// setDefaults 设置默认值
func setDefaults(v *viper.Viper) {
	v.SetDefault("app.port", 8080)
	v.SetDefault("app.pprof_port", 6060)
	v.SetDefault("app.env", "dev")
	v.SetDefault("app.read_timeout", 15)
	v.SetDefault("app.write_timeout", 15)
	v.SetDefault("app.graceful_timeout", 30)

	v.SetDefault("mysql.charset", "utf8mb4")
	v.SetDefault("mysql.max_idle_conns", 10)
	v.SetDefault("mysql.max_open_conns", 100)
	v.SetDefault("mysql.conn_max_lifetime", 3600)
	v.SetDefault("mysql.log_level", "warn")

	v.SetDefault("redis.db", 0)
	v.SetDefault("redis.pool_size", 100)
	v.SetDefault("redis.min_idle_conns", 10)

	v.SetDefault("jwt.expire_hours", 168)
	v.SetDefault("jwt.admin_expire_hours", 24)
	v.SetDefault("jwt.issuer", "jisan-esports")

	v.SetDefault("oss.sign_expire", 300)

	v.SetDefault("websocket.heartbeat_seconds", 25)
	v.SetDefault("websocket.timeout_seconds", 70)
	v.SetDefault("websocket.max_conn_per_user", 3)
	v.SetDefault("websocket.probe_interval_seconds", 300)
	v.SetDefault("websocket.probe_max_miss", 3)
	v.SetDefault("websocket.write_buffer_size", 10240)
	v.SetDefault("websocket.read_buffer_size", 10240)
	v.SetDefault("websocket.alert_threshold", 80)

	v.SetDefault("rate_limit.user_qps", 20)
	v.SetDefault("rate_limit.player_qps", 30)
	v.SetDefault("rate_limit.admin_qps", 100)
	v.SetDefault("rate_limit.burst", 50)
}

// fsnotifyEvent 已由 github.com/fsnotify/fsnotify.Event 替代

// HeartbeatDuration 返回心跳间隔
func (w WSConfig) HeartbeatDuration() time.Duration {
	return time.Duration(w.HeartbeatSeconds) * time.Second
}

// TimeoutDuration 返回连接超时时长
func (w WSConfig) TimeoutDuration() time.Duration {
	return time.Duration(w.TimeoutSeconds) * time.Second
}

// ProbeIntervalDuration 返回活跃度探针间隔
func (w WSConfig) ProbeIntervalDuration() time.Duration {
	return time.Duration(w.ProbeIntervalSeconds) * time.Second
}
