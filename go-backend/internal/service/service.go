package service

import (
	"go.uber.org/zap"
	"gorm.io/gorm"

	"github.com/jisan/e-sports-platform/internal/config"
	"github.com/jisan/e-sports-platform/internal/repository"
	"github.com/jisan/e-sports-platform/internal/utils"
	"github.com/jisan/e-sports-platform/pkg/cache"
	"github.com/jisan/e-sports-platform/pkg/lock"
	"github.com/jisan/e-sports-platform/pkg/queue"
	"github.com/jisan/e-sports-platform/pkg/websocket"
)

// Deps service 层依赖容器
type Deps struct {
	DB     *gorm.DB
	ReadDB *gorm.DB
	Redis  *cache.RedisClient
	Cache  *cache.Cache
	Hub    *websocket.Hub
	Queue  *queue.Client
	JWT    *utils.JWTManager
	Logger *zap.Logger
	Config *config.Config
}

// 全局依赖(由 Init 初始化，供各 service 包级函数调用)
var (
	db     *gorm.DB
	readDB *gorm.DB
	redis  *cache.RedisClient
	cacheC *cache.Cache
	hub    *websocket.Hub
	queueC *queue.Client
	jwtMgr *utils.JWTManager
	logger *zap.Logger
	cfg    *config.Config
	distLock *lock.DistributedLock // 安全的分布式锁(基于 Redis SetNX + Lua 释放)

	userRepo       *repository.UserRepo
	adminRepo      *repository.AdminRepo
	orderRepo      *repository.OrderRepo
	chatRepo       *repository.ChatRepo
	clubRepo       *repository.ClubRepo
	paymentRepo    *repository.PaymentRepo
	inviteCodeRepo *repository.InviteCodeRepo
)

// Init 初始化 service 层依赖与各仓储实例
// 应在路由注册前由 routes 包调用一次
func Init(d *Deps) {
	db = d.DB
	readDB = d.ReadDB
	redis = d.Redis
	cacheC = d.Cache
	hub = d.Hub
	queueC = d.Queue
	jwtMgr = d.JWT
	logger = d.Logger
	cfg = d.Config

	// 初始化安全的分布式锁
	if redis != nil {
		distLock = lock.NewDistributedLock(redis)
	}

	userRepo = repository.NewUserRepo(db)
	adminRepo = repository.NewAdminRepo(db)
	orderRepo = repository.NewOrderRepo(db)
	chatRepo = repository.NewChatRepo(db)
	clubRepo = repository.NewClubRepo(db)
	paymentRepo = repository.NewPaymentRepo(db)
	inviteCodeRepo = repository.NewInviteCodeRepo(db)

	// 注入文件安全模块默认 DB(水印/加密/上链存证)
	utils.SetFileSecurityDB(db)
}

// readDBOr 主库兜底:未配置只读库时使用主库
func readDBOr() *gorm.DB {
	if readDB != nil {
		return readDB
	}
	return db
}
