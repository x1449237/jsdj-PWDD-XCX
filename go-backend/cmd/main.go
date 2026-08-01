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
	"github.com/jisan/e-sports-platform/internal/config"
	"github.com/jisan/e-sports-platform/internal/middleware"
	"github.com/jisan/e-sports-platform/internal/routes"
	"github.com/jisan/e-sports-platform/internal/utils"
	"github.com/jisan/e-sports-platform/pkg/cache"
	"github.com/jisan/e-sports-platform/pkg/queue"
	"github.com/jisan/e-sports-platform/pkg/websocket"
	"github.com/robfig/cron/v3"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
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
	cronScheduler := startCronJobs(appLogger, db)
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
func startCronJobs(log *zap.Logger, db *gorm.DB) *cron.Cron {
	c := cron.New(cron.WithSeconds())

	// 示例：每 5 分钟检查订单超时关闭
	_, _ = c.AddFunc("0 */5 * * * *", func() {
		log.Info("[cron] 执行订单超时检查任务")
		// 实际逻辑由 service 层注入
	})

	// 示例：每天凌晨 2 点清理过期日志
	_, _ = c.AddFunc("0 0 2 * * *", func() {
		log.Info("[cron] 执行日志清理任务")
	})

	// 示例：每分钟推送预约订单提醒
	_, _ = c.AddFunc("0 * * * * *", func() {
		log.Info("[cron] 执行预约订单提醒任务")
	})

	c.Start()
	return c
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
