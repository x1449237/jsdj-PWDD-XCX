package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"net/http"
	_ "net/http/pprof"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	redislib "github.com/redis/go-redis/v9"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"github.com/jisan/e-sports-platform/internal/config"
	"github.com/jisan/e-sports-platform/internal/middleware"
	"github.com/jisan/e-sports-platform/internal/model"
	"github.com/jisan/e-sports-platform/internal/routes"
	"github.com/jisan/e-sports-platform/internal/service"
	"github.com/jisan/e-sports-platform/internal/utils"
	"github.com/jisan/e-sports-platform/pkg/cache"
	"github.com/jisan/e-sports-platform/pkg/queue"
	"github.com/jisan/e-sports-platform/pkg/websocket"
	"github.com/robfig/cron/v3"
)

func main() {
	// 配置文件路径(支持命令行参数与环境变量)
	configPath := flag.String("config", "./internal/config/config.yaml", "配置文件路径")
	flag.Parse()
	if envPath := os.Getenv("CONFIG_PATH"); envPath != "" {
		*configPath = envPath
	}

	// 1. 初始化 zap 日志(配置加载前先用基础日志)
	baseLogger, _ := zap.NewProduction()
	baseLogger.Info("戟三电竞平台后端启动中...", zap.String("config", *configPath))

	// 2. 加载配置(支持热加载)
	cfg, err := config.Load(*configPath, baseLogger)
	if err != nil {
		baseLogger.Fatal("加载配置失败", zap.Error(err))
	}

	// 3. 根据配置重新初始化日志
	appLogger := initLogger(cfg.Log)
	defer appLogger.Sync()
	appLogger.Info("日志初始化完成",
		zap.String("env", cfg.App.Env),
		zap.String("level", cfg.Log.Level))

	// 根据环境设置 Gin 模式
	if cfg.App.Env == "prod" {
		gin.SetMode(gin.ReleaseMode)
	}

	// 4. 初始化 MySQL(GORM，支持读写分离配置)
	db, readDB, err := initMySQL(cfg.MySQL, appLogger)
	if err != nil {
		appLogger.Fatal("初始化 MySQL 失败", zap.Error(err))
	}
	appLogger.Info("MySQL 初始化完成",
		zap.String("host", cfg.MySQL.Host),
		zap.Bool("read_replica_enabled", readDB != nil))

	// 5. 初始化 Redis
	redisClient, err := cache.NewRedisClient(cfg.Redis)
	if err != nil {
		appLogger.Fatal("初始化 Redis 失败", zap.Error(err))
	}
	defer redisClient.Close()
	appLogger.Info("Redis 初始化完成", zap.String("addr", cfg.Redis.Addr()))

	cacheWrapper := cache.NewCache(redisClient)

	// 6. 初始化 WebSocket Hub
	hub := websocket.NewHub(redisClient, appLogger)
	appLogger.Info("WebSocket Hub 初始化完成")

	// 7. 初始化异步队列客户端
	queueClient := queue.NewClient(cfg.Redis)
	defer queueClient.Close()
	appLogger.Info("异步队列客户端初始化完成")

	// 7.1 初始化异步队列消费端 Server
	queueServer, err := queue.NewServer(cfg.Redis)
	if err != nil {
		appLogger.Fatal("初始化异步队列 Server 失败", zap.Error(err))
	}
	var rdb *redislib.Client
	if rc, ok := any(redisClient).(*cache.RedisClient); ok {
		rdb = rc.Client()
	}
	queueServer.RegisterHandlers(db, rdb, hub)
	go func() {
		if err := queueServer.Start(); err != nil {
			appLogger.Error("异步队列 Server 异常退出", zap.Error(err))
		}
	}()
	appLogger.Info("异步队列 Server 初始化完成")

	// 8. 初始化 JWT 管理器
	jwtManager := utils.NewJWTManager(cfg.JWT.Secret, cfg.JWT.ExpireHours, cfg.JWT.AdminExpireHours, cfg.JWT.Issuer)

	// 9. 初始化中间件
	authMW := middleware.NewAuthMiddleware(jwtManager)
	rateLimitMW := middleware.NewRateLimitMiddleware(redisClient, cfg.RateLimit)
	clubScopeMW := middleware.NewClubScopeMiddleware()
	opLogMW := middleware.NewOperationLogMiddleware(db, appLogger)
	apiVersionMW := middleware.NewAPIVersionMiddleware("v1")

	// 10. 创建 Gin 引擎并注册全局中间件
	r := gin.New()
	r.Use(
		middleware.TraceID(),                       // 请求追踪ID
		middleware.Recover(appLogger),              // 全局异常恢复
		middleware.CORS(middleware.DefaultCORSConfig()), // CORS 跨域
		apiVersionMW.APIVersion(),                  // API 版本控制(灰度发布)
		rateLimitMW.RateLimit(),                    // 接口限流
	)

	// 11. 注册业务路由(由 routes 包负责，其他代理实现)
	//     routes.Deps 为依赖容器，字段由 routes 包定义
	deps := &routes.Deps{
		Config:     cfg,
		Logger:     appLogger,
		DB:         db,
		ReadDB:     readDB,
		Redis:      redisClient,
		Cache:      cacheWrapper,
		Hub:        hub,
		Queue:      queueClient,
		JWT:        jwtManager,
		AuthMW:     authMW,
		ClubScopeMW: clubScopeMW,
		OpLogMW:    opLogMW,
	}
	routes.RegisterRoutes(r, deps)

	// 12. 启动 cron 定时任务
	cronScheduler := startCronJobs(appLogger, db, redisClient, queueClient)
	defer cronScheduler.Stop()
	appLogger.Info("Cron 定时任务已启动")

	// 13. 启动 pprof 性能监控(独立端口)
	pprofServer := startPprof(cfg.App.PprofPort, appLogger)

	// 14. 启动 HTTP 服务
	httpServer := &http.Server{
		Addr:         fmt.Sprintf(":%d", cfg.App.Port),
		Handler:      r,
		ReadTimeout:  time.Duration(cfg.App.ReadTimeout) * time.Second,
		WriteTimeout: time.Duration(cfg.App.WriteTimeout) * time.Second,
	}

	go func() {
		appLogger.Info("HTTP 服务启动", zap.Int("port", cfg.App.Port))
		if err := httpServer.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			appLogger.Fatal("HTTP 服务异常退出", zap.Error(err))
		}
	}()

	// 15. 优雅关闭(捕获 SIGINT/SIGTERM)
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	sig := <-quit
	appLogger.Info("收到退出信号，开始优雅关闭", zap.String("signal", sig.String()))

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(cfg.App.GracefulTimeout)*time.Second)
	defer cancel()

	if err := httpServer.Shutdown(ctx); err != nil {
		appLogger.Error("HTTP 服务关闭失败", zap.Error(err))
	}
	if err := pprofServer.Shutdown(ctx); err != nil {
		appLogger.Error("pprof 服务关闭失败", zap.Error(err))
	}
	queueServer.Shutdown()
	appLogger.Info("异步队列 Server 已关闭")

	appLogger.Info("戟三电竞平台后端已安全退出")
}

// initLogger 根据 LogConfig 初始化 zap 日志
func initLogger(cfg config.LogConfig) *zap.Logger {
	var level zapcore.Level
	if err := level.UnmarshalText([]byte(cfg.Level)); err != nil {
		level = zapcore.InfoLevel
	}

	encCfg := zapcore.EncoderConfig{
		TimeKey:        "ts",
		LevelKey:       "level",
		NameKey:        "logger",
		CallerKey:      "caller",
		FunctionKey:    zapcore.OmitKey,
		MessageKey:     "msg",
		StacktraceKey:  "stacktrace",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    zapcore.LowercaseLevelEncoder,
		EncodeTime:     zapcore.ISO8601TimeEncoder,
		EncodeDuration: zapcore.SecondsDurationEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder,
	}

	var encoder zapcore.Encoder
	if cfg.Encoding == "console" {
		encoder = zapcore.NewConsoleEncoder(encCfg)
	} else {
		encoder = zapcore.NewJSONEncoder(encCfg)
	}

	// 输出到 stdout；如配置目录则同时写文件
	var ws zapcore.WriteSyncer = os.Stdout
	if cfg.Directory != "" {
		if err := os.MkdirAll(cfg.Directory, 0o755); err == nil {
			if f, err := os.OpenFile(cfg.Directory+"/app.log",
				os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o644); err == nil {
				ws = zapcore.NewMultiWriteSyncer(os.Stdout, f)
			}
		}
	}

	core := zapcore.NewCore(encoder, ws, level)
	return zap.New(core, zap.AddCaller(), zap.AddStacktrace(zap.ErrorLevel))
}

// initMySQL 初始化 MySQL 连接池(支持读写分离配置)
// 返回主库与只读库(未配置只读库时 readDB 为 nil)
func initMySQL(cfg config.MySQLConfig, log *zap.Logger) (db *gorm.DB, readDB *gorm.DB, err error) {
	gormLogLevel := logger.Warn
	switch cfg.LogLevel {
	case "silent":
		gormLogLevel = logger.Silent
	case "error":
		gormLogLevel = logger.Error
	case "warn":
		gormLogLevel = logger.Warn
	case "info":
		gormLogLevel = logger.Info
	}

	db, err = gorm.Open(mysql.Open(cfg.DSN()), &gorm.Config{
		Logger: logger.Default.LogMode(gormLogLevel),
	})
	if err != nil {
		return nil, nil, fmt.Errorf("连接 MySQL 主库失败: %w", err)
	}

	// 配置主库连接池
	sqlDB, err := db.DB()
	if err != nil {
		return nil, nil, err
	}
	sqlDB.SetMaxIdleConns(cfg.MaxIdleConns)
	sqlDB.SetMaxOpenConns(cfg.MaxOpenConns)
	sqlDB.SetConnMaxLifetime(time.Duration(cfg.ConnMaxLifetime) * time.Second)

	// 读写分离：配置了只读库则建立只读连接
	if cfg.ReadHost != "" {
		readDB, err = gorm.Open(mysql.Open(cfg.ReadDSN()), &gorm.Config{
			Logger: logger.Default.LogMode(gormLogLevel),
		})
		if err != nil {
			log.Warn("连接 MySQL 只读库失败，回退到主库", zap.Error(err))
			readDB = nil
		} else {
			readSQLDB, err := readDB.DB()
			if err == nil {
				readSQLDB.SetMaxIdleConns(cfg.MaxIdleConns)
				readSQLDB.SetMaxOpenConns(cfg.MaxOpenConns)
				readSQLDB.SetConnMaxLifetime(time.Duration(cfg.ConnMaxLifetime) * time.Second)
			}
		}
	}

	return db, readDB, nil
}

// startCronJobs 启动 cron 定时任务
func startCronJobs(log *zap.Logger, db *gorm.DB, rdb *cache.RedisClient, qc *queue.Client) *cron.Cron {
	c := cron.New(cron.WithSeconds())

	// 1. 每分钟扫超时订单兜底
	_, _ = c.AddFunc("0 * * * * *", func() {
		log.Info("[cron] 执行超时订单兜底检查")
		cutoff := time.Now().Add(-10 * time.Minute)
		var pendingOrders []model.Order
		db.Where("status = ? AND created_at < ?", model.OrderStatusPending, cutoff).Find(&pendingOrders)
		for _, o := range pendingOrders {
			db.Model(&o).Update("status", 10)
			db.Create(&model.OrderStatusLog{OrderID: o.ID, FromStatus: 0, ToStatus: 10, Reason: "兜底超时关闭", OperatorType: "system"})
		}
		if len(pendingOrders) > 0 {
			log.Info("[cron] 兜底关闭超时订单", zap.Int("count", len(pendingOrders)))
		}
	})

	// 2. 每小时自动结算：检查待结算状态且已过72h的订单直接置为已结算
	_, _ = c.AddFunc("0 0 * * * *", func() {
		log.Info("[cron] 执行自动结算检查")
		cutoff := time.Now().Add(-72 * time.Hour)
		var toSettleOrders []model.Order
		db.Where("status = ? AND updated_at < ?", model.OrderStatusToSettle, cutoff).Find(&toSettleOrders)
		for _, o := range toSettleOrders {
			db.Model(&o).Update("status", model.OrderStatusSettled)
			db.Model(&o).Update("settled_at", time.Now())
			db.Create(&model.OrderStatusLog{OrderID: o.ID, FromStatus: 5, ToStatus: 6, Reason: "Cron自动结算", OperatorType: "system"})
		}
		if len(toSettleOrders) > 0 {
			log.Info("[cron] 自动结算订单", zap.Int("count", len(toSettleOrders)))
		}
	})

	// 3. 宵禁1小时前提醒：21:00 提醒未成年用户即将宵禁
	_, _ = c.AddFunc("0 0 21 * * *", func() {
		log.Info("[cron] 执行未成年人宵禁前1小时提醒")
		var minors []model.User
		db.Where("is_minor = 1 AND status = 1").Find(&minors)
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		for _, u := range minors {
			_ = qc.EnqueueMessagePush(ctx, queue.MessagePushPayload{
				UserID:  u.ID,
				Type:    "curfew_reminder",
				Title:   "宵禁提醒",
				Content: "距离22:00未成年人宵禁还有1小时，请合理安排时间",
			})
		}
		if len(minors) > 0 {
			log.Info("[cron] 发送未成年人宵禁提醒", zap.Int("count", len(minors)))
		}
	})

	// 4. 每天凌晨3点 NTP 时间同步 + AES备份 + 72h自动结算兜底
	_, _ = c.AddFunc("0 0 3 * * *", func() {
		log.Info("[cron] 执行每日3点综合任务(NTP+备份+72h结算兜底)")

		// 4.1 NTP 时间同步(沙箱模式: 读 remote NTP; 失败时使用系统时间)
		go func() {
			log.Info("[cron] NTP 同步开始")
			// 真实项目: 使用 github.com/beevik/ntp 或 systemd-timesyncd
			//   t, err := ntp.Time("pool.ntp.org"); if err==nil { 设置系统时间(需要root) }
			// 沙箱: 仅打印并记录到 system_configs
			_ = db.Where("`key` = ?", "ntp_last_sync").
				Assign(map[string]interface{}{
					"value":       time.Now().Format(time.RFC3339),
					"description": "NTP 最后同步时间(沙箱使用系统时间)",
					"updated_at":  time.Now(),
				}).FirstOrCreate(&model.SystemConfig{Key: "ntp_last_sync", Value: time.Now().Format(time.RFC3339)}).Error
			log.Info("[cron] NTP 同步结束(沙箱)")
		}()

		// 4.2 AES 备份 (调用 service.AdminCreateBackup)
		go func() {
			log.Info("[cron] AES 自动备份开始")
			if rdb != nil {
				// 分布式锁: 避免多实例重复备份
				lockKey := "lock:backup:daily"
				ok, _ := rdb.SetNX(context.Background(), lockKey, "1", 2*time.Hour)
				if !ok {
					log.Info("[cron] 跳过备份: 已被其他实例持有锁")
					return
				}
				defer func() { _ = rdb.Del(context.Background(), lockKey) }()
			}
			_, berr := service.AdminCreateBackup(0, "auto_daily_"+time.Now().Format("20060102"))
			if berr != nil {
				log.Warn("[cron] 自动备份失败", zap.Error(berr))
			} else {
				log.Info("[cron] AES 自动备份完成")
			}
		}()

		// 4.3 72h自动结算兜底(对每小时任务的二次保障)
		go func() {
			log.Info("[cron] 72h自动结算兜底开始")
			cutoff := time.Now().Add(-72 * time.Hour)
			res := db.Model(&model.Order{}).
				Where("status = ? AND updated_at < ?", model.OrderStatusToSettle, cutoff).
				Updates(map[string]interface{}{
					"status":     model.OrderStatusSettled,
					"settled_at": time.Now(),
					"updated_at": time.Now(),
				})
			if res.Error != nil {
				log.Warn("[cron] 72h自动结算兜底失败", zap.Error(res.Error))
				return
			}
			if res.RowsAffected > 0 {
				log.Info("[cron] 72h自动结算兜底订单", zap.Int64("count", res.RowsAffected))
			}
		}()
	})

	// 5. 每月1号 UP主 升降级 (比较 follower_count 与 tier_config)
	_, _ = c.AddFunc("0 0 0 1 * *", func() {
		log.Info("[cron] 执行每月1号 UP主升降级任务")
		up, down, err := service.MonthlyCronUpMasterTier()
		if err != nil {
			log.Warn("[cron] UP主升降级任务异常", zap.Error(err))
			return
		}
		log.Info("[cron] UP主升降级任务完成", zap.Int("upgrade_count", up), zap.Int("downgrade_count", down))
	})

	// 6. 每天23:59 清理过期日志 (>log_retention_days)
	_, _ = c.AddFunc("0 59 23 * * *", func() {
		log.Info("[cron] 执行每日过期日志清理")
		retentionDays := 30
		// 从 system_configs 读取 log_retention_days
		var sc model.SystemConfig
		if err := db.Where("`key` = ?", "log_retention_days").First(&sc).Error; err == nil && sc.Value != "" {
			if n := atoiSafe(sc.Value); n > 0 {
				retentionDays = n
			}
		}
		cutoff := time.Now().AddDate(0, 0, -retentionDays)
		log.Info("[cron] 日志保留天数", zap.Int("days", retentionDays), zap.Time("cutoff", cutoff))
		// 依次清理多张日志表(忽略错误继续)
		tables := []string{
			"login_logs",
			"access_logs",
			"anti_boosting_log",
			"audit_logs",
			"circuit_breaker_log",
			"notify_logs",
			"chat_read_receipts",
		}
		for _, tbl := range tables {
			sql := fmt.Sprintf("DELETE FROM %s WHERE created_at < ? LIMIT 2000", tbl)
			res := db.Exec(sql, cutoff)
			if res.Error != nil {
				log.Debug("[cron] 清理跳过(表可能不存在)", zap.String("table", tbl), zap.Error(res.Error))
				continue
			}
			if res.RowsAffected > 0 {
				log.Info("[cron] 清理过期日志", zap.String("table", tbl), zap.Int64("rows", res.RowsAffected))
			}
		}
		log.Info("[cron] 每日过期日志清理完成")
	})

	c.Start()
	return c
}

// atoiSafe 字符串转 int 安全函数
func atoiSafe(s string) int {
	n := 0
	for i := 0; i < len(s); i++ {
		if s[i] < '0' || s[i] > '9' {
			break
		}
		n = n*10 + int(s[i]-'0')
	}
	return n
}

// startPprof 启动 pprof 性能监控服务
func startPprof(port int, log *zap.Logger) *http.Server {
	srv := &http.Server{
		Addr:    fmt.Sprintf(":%d", port),
		Handler: http.DefaultServeMux, // net/http/pprof 已注册到 DefaultServeMux
	}
	go func() {
		log.Info("pprof 性能监控服务启动", zap.Int("port", port))
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Error("pprof 服务异常", zap.Error(err))
		}
	}()
	return srv
}
