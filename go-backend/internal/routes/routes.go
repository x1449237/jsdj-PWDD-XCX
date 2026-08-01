package routes

import (
	"github.com/gin-gonic/gin"
	"github.com/jisan/e-sports-platform/internal/config"
	"github.com/jisan/e-sports-platform/internal/handler"
	"github.com/jisan/e-sports-platform/internal/middleware"
	"github.com/jisan/e-sports-platform/internal/service"
	"github.com/jisan/e-sports-platform/internal/utils"
	"github.com/jisan/e-sports-platform/pkg/cache"
	"github.com/jisan/e-sports-platform/pkg/queue"
	"github.com/jisan/e-sports-platform/pkg/websocket"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// Deps 依赖容器(由 main 注入)
type Deps struct {
	Config      *config.Config
	Logger      *zap.Logger
	DB          *gorm.DB
	ReadDB      *gorm.DB
	Redis       *cache.RedisClient
	Cache       *cache.Cache
	Hub         *websocket.Hub
	Queue       *queue.Client
	JWT         *utils.JWTManager
	AuthMW      *middleware.AuthMiddleware
	ClubScopeMW *middleware.ClubScopeMiddleware
	OpLogMW     *middleware.OperationLogMiddleware
}

// RegisterRoutes 注册业务路由
func RegisterRoutes(r *gin.Engine, deps *Deps) {
	// 1. 初始化 service 层依赖与仓储实例
	service.Init(&service.Deps{
		DB:     deps.DB,
		ReadDB: deps.ReadDB,
		Redis:  deps.Redis,
		Cache:  deps.Cache,
		Hub:    deps.Hub,
		Queue:  deps.Queue,
		JWT:    deps.JWT,
		Logger: deps.Logger,
		Config: deps.Config,
	})

	// 2. 实例化各 handler
	authH := handler.NewAuthHandler()
	userH := handler.NewUserHandler()
	orderH := handler.NewOrderHandler()
	playerH := handler.NewPlayerHandler()
	clubH := handler.NewClubHandler()
	chatH := handler.NewChatHandler()
	paymentH := handler.NewPaymentHandler()
	distributorH := handler.NewDistributorHandler()
	dispatcherH := handler.NewDispatcherHandler()
	guardianH := handler.NewGuardianHandler()
	marketingH := handler.NewMarketingHandler()

	// 平台管理员 handler
	adminH := handler.NewAdminHandler()
	adminAuditH := handler.NewAdminAuditHandler()
	adminFinanceH := handler.NewAdminFinanceHandler()
	adminAppealH := handler.NewAdminAppealHandler()
	adminConfigH := handler.NewAdminConfigHandler()
	adminRiskH := handler.NewAdminRiskHandler()
	adminInviteH := handler.NewAdminInviteHandler()
	adminMarketingH := handler.NewAdminMarketingHandler()

	// 内置管理端(俱乐部)handler
	shopH := handler.NewShopHandler()

	// WebSocket 升级处理器
	wsHandler := websocket.NewHandler(deps.Hub, deps.Logger)

	// 中间件快捷引用
	authReq := deps.AuthMW.AuthRequired()
	adminReq := deps.AuthMW.AdminRequired()
	shopReq := deps.AuthMW.ShopAdminRequired()
	clubScope := deps.ClubScopeMW.ClubScope()

	// 3. 健康检查(无需认证)
	r.GET("/health", func(c *gin.Context) {
		utils.Success(c, gin.H{"status": "ok"})
	})

	api := r.Group("/api/v1")

	// ---------------- 认证(公开) ----------------
	auth := api.Group("/auth")
	{
		auth.POST("/wx-login", authH.WxLogin)
		// 注册/解密手机号需要登录态
		auth.POST("/register", authReq, authH.Register)
		auth.POST("/decode-phone", authReq, authH.DecodePhone)
	}

	// 平台管理员登录(公开)
	api.POST("/admin/login", authH.AdminLogin)
	// 内置管理端登录(公开)
	api.POST("/shop/login", authH.ShopAdminLogin)

	// 微信支付回调(公开，签名校验在 service 内)
	api.POST("/webhook/wxpay", paymentH.WxPayCallback)

	// ---------------- 俱乐部浏览(公开) ----------------
	clubs := api.Group("/clubs")
	{
		clubs.GET("", clubH.GetClubList)
		clubs.GET("/:id", clubH.GetClubDetail)
		// 入驻开关查询(公开，前端入驻页打开时先查询)
		clubs.GET("/join-switch", clubH.CheckClubSwitch)
		clubs.POST("/join-click", authReq, clubH.RecordClubJoinClick)

		// 入驻流程(需登录)
		clubs.POST("/abbr", authReq, clubH.GenerateAbbr)
		clubs.POST("/upload-pdf", authReq, clubH.UploadPDF)
		clubs.POST("/registrations/personal", authReq, clubH.SubmitPersonal)
		clubs.POST("/registrations/enterprise", authReq, clubH.SubmitEnterprise)
		clubs.POST("/draft", authReq, clubH.SaveDraft)
		clubs.GET("/draft", authReq, clubH.GetDraft)
		clubs.GET("/seal-remind", authReq, clubH.GetSealRemind)
	}

	// ---------------- 用户侧(需登录) ----------------
	userAPI := api.Group("", authReq)
	{
		// 用户资料
		userAPI.GET("/user/profile", userH.GetProfile)
		userAPI.PUT("/user/profile", userH.UpdateProfile)
		userAPI.POST("/user/realname", userH.SubmitRealname)
		userAPI.GET("/user/realname/status", userH.GetRealnameStatus)
		userAPI.POST("/user/face-verify", userH.FaceVerify)
		userAPI.GET("/user/favorites", userH.ListFavorites)
		userAPI.POST("/user/favorites/:player_id", userH.ToggleFavorite)

		// 订单
		userAPI.POST("/orders", orderH.CreateOrder)
		userAPI.GET("/orders", orderH.GetOrderList)
		userAPI.GET("/orders/:id", orderH.GetOrderDetail)
		userAPI.POST("/orders/:id/cancel", orderH.CancelOrder)
		userAPI.POST("/orders/:id/confirm", orderH.ConfirmOrderAcceptance)
		userAPI.POST("/orders/:id/appeal", orderH.SubmitAppeal)
		userAPI.POST("/orders/:id/evaluation", orderH.SubmitEvaluation)
		userAPI.POST("/orders/:id/reward", orderH.SendReward)
		userAPI.POST("/orders/:id/evidence", orderH.UploadEvidence)

		// 申诉
		userAPI.GET("/appeals", orderH.GetAppealList)
		userAPI.GET("/appeals/:id", orderH.GetAppealDetail)
		userAPI.POST("/appeals/:id/materials", orderH.UploadAppealMaterials)

		// 聊天
		userAPI.GET("/chat/sessions", chatH.GetSessions)
		userAPI.GET("/chat/sessions/:id/messages", chatH.GetMessages)
		userAPI.POST("/chat/messages", chatH.SendMessage)
		userAPI.POST("/chat/messages/:id/revoke", chatH.RevokeMessage)
		userAPI.POST("/chat/files", chatH.UploadChatFile)
		userAPI.POST("/chat/sessions/:id/intervention", chatH.RequestIntervention)

		// 支付
		userAPI.POST("/payments", paymentH.CreatePayment)

		// 打手
		player := userAPI.Group("/player")
		{
			player.GET("/grab-orders", playerH.GetGrabList)
			player.POST("/grab-orders/:id", playerH.GrabOrder)
			player.GET("/orders", playerH.GetPlayerOrders)
			player.POST("/orders/:id/start", playerH.StartService)
			player.POST("/orders/:id/complete", playerH.CompleteService)
			player.POST("/orders/:id/transfer", playerH.TransferOrder)
			player.GET("/earnings", playerH.GetEarnings)
			player.GET("/earnings/frozen", playerH.GetFrozenEarnings)
			player.POST("/withdraw", playerH.ApplyWithdraw)
			player.GET("/services", playerH.GetMyServices)
			player.POST("/services", playerH.CreateService)
			player.GET("/evaluations", playerH.GetMyEvaluations)
			player.POST("/evaluations/:id/appeal", playerH.AppealEvaluation)
			player.POST("/join-application", playerH.SubmitJoinApplication)
			player.GET("/join-applications", playerH.GetMyApplications)
		}
		// 打手详情/列表(可公开浏览)
		userAPI.GET("/player/:id", playerH.GetPlayerDetail)

		// 打手列表
		userAPI.GET("/players", playerH.GetPlayerList)

		// 分销商
		distributor := userAPI.Group("/distributor")
		{
			distributor.GET("/subordinates", distributorH.GetSubordinates)
			distributor.GET("/commissions", distributorH.GetCommissionList)
			distributor.GET("/ranking", distributorH.GetRanking)
		}

		// 派单员
		dispatcher := userAPI.Group("/dispatcher")
		{
			dispatcher.GET("/orders", dispatcherH.GetDispatchOrders)
			dispatcher.POST("/orders/:id/dispatch", dispatcherH.DispatchOrder)
			dispatcher.GET("/history", dispatcherH.GetDispatchHistory)
		}

		// 家长(未成年守护)
		guardian := userAPI.Group("/guardian")
		{
			guardian.POST("/bind", guardianH.BindGuardian)
			guardian.GET("/children/:id/report", guardianH.GetChildReport)
			guardian.PUT("/children/:id/settings", guardianH.UpdateChildSettings)
			guardian.POST("/children/:id/freeze", guardianH.FreezeChild)
		}

		// 营销活动
		marketing := userAPI.Group("/marketing")
		{
			marketing.GET("/coupons", marketingH.GetMyCoupons)
			marketing.GET("/recharge-activities", marketingH.GetRechargeActivities)
			marketing.POST("/recharge", marketingH.Recharge)
			marketing.GET("/lottery-activities", marketingH.GetLotteryActivities)
			marketing.POST("/lottery/draw", marketingH.DrawLottery)
			marketing.GET("/group-buy-activities", marketingH.GetGroupBuyActivities)
			marketing.POST("/group-buy/join", marketingH.JoinGroupBuy)
			marketing.GET("/invite-qrcode", marketingH.GenerateInviteQRCode)
			marketing.POST("/redeem", marketingH.RedeemInviteCode)
		}
	}

	// ---------------- WebSocket(需登录) ----------------
	api.GET("/ws", authReq, wsHandler.Handle)

	// ---------------- 平台管理端(需平台管理员认证) ----------------
	admin := api.Group("/admin", adminReq)
	{
		// 认证相关(已登录态)
		admin.POST("/change-password", authH.AdminChangePassword)
		admin.POST("/bind-email", authH.AdminBindEmail)
		admin.POST("/verify-email", authH.AdminVerifyEmail)
		admin.GET("/init-status", authH.AdminGetInitStatus)
		admin.POST("/forgot-account", authH.ForgotAccount)
		admin.POST("/forgot-password", authH.ForgotPassword)
		admin.POST("/webauthn/begin", authH.WebauthnBegin)
		admin.POST("/webauthn/finish", authH.WebauthnFinish)

		// 仪表盘
		admin.GET("/dashboard", adminH.Dashboard)
		admin.GET("/big-screen", adminH.BigScreenData)

		// 用户管理
		admin.GET("/users", adminH.GetUsers)
		admin.GET("/users/normal", adminH.GetNormalUsers)
		admin.GET("/users/failed-verification", adminH.GetFailedVerificationUsers)
		admin.GET("/users/export", adminH.ExportUsers)
		admin.GET("/users/:id", adminH.GetUserDetail)
		admin.POST("/users/:id/ban", adminH.BanUser)
		admin.POST("/users/:id/unban", adminH.UnbanUser)

		// 管理员管理
		admin.GET("/managers", adminH.GetManagers)
		admin.POST("/managers", adminH.AddManager)
		admin.PUT("/managers/:id", adminH.UpdateManager)
		admin.DELETE("/managers/:id", adminH.DeleteManager)
		admin.POST("/managers/:id/reset-password", adminH.ResetManagerPassword)

		// 订单管理
		admin.GET("/orders", adminH.GetOrders)
		admin.GET("/orders/failed", adminH.GetFailedOrders)
		admin.POST("/orders/batch", adminH.BatchOrderOperation)
		admin.POST("/orders/:id/status", adminH.ForceUpdateOrderStatus)
		admin.POST("/orders/:id/retry-face", adminH.RetryFaceVerify)
		admin.GET("/orders/:id/profit-share", adminFinanceH.GetProfitShare)
		admin.POST("/orders/:id/settle", adminFinanceH.SettleOrder)
		admin.POST("/orders/:id/refund", adminFinanceH.ProcessRefund)

		// 审核(俱乐部/打手/分销商/派单员/管理端账号)
		admin.GET("/clubs/audit", adminAuditH.AuditClubs)
		admin.GET("/clubs/audit-filter", adminAuditH.AuditClubsFiltered)
		admin.POST("/clubs/:id/approve", adminAuditH.ApproveClub)
		admin.POST("/clubs/:id/reject", adminAuditH.RejectClub)
		admin.POST("/clubs/:id/freeze", adminAuditH.FreezeClub)
		admin.POST("/clubs/:id/unfreeze", adminAuditH.UnfreezeClub)
		admin.POST("/clubs/:id/cancel", adminAuditH.CancelClub)
		admin.POST("/clubs/:id/vbadge/hide", adminAuditH.HideVBadge)
		admin.POST("/clubs/:id/vbadge/restore", adminAuditH.RestoreVBadge)
		admin.GET("/clubs/:id/change-logs", adminAuditH.GetClubChangeLogs)
		admin.POST("/fine-rules/:id/review", adminAuditH.ReviewFineRule)
		admin.GET("/players/audit", adminAuditH.AuditPlayers)
		admin.POST("/players/:id/approve", adminAuditH.ApprovePlayer)
		admin.POST("/players/:id/reject", adminAuditH.RejectPlayer)
		admin.GET("/distributors/audit", adminAuditH.AuditDistributors)
		admin.POST("/distributors/:id/approve", adminAuditH.ApproveDistributor)
		admin.GET("/dispatchers/audit", adminAuditH.AuditDispatchers)
		admin.POST("/dispatchers/:id/approve", adminAuditH.ApproveDispatcher)
		admin.GET("/shop-admins", adminAuditH.GetShopAdmins)
		admin.POST("/shop-admins", adminAuditH.AdminAddShopAccount)
		admin.POST("/shop-admins/:id/disable", adminAuditH.DisableShopAdmin)
		admin.POST("/shop-admins/:id/enable", adminAuditH.EnableShopAdmin)
		admin.POST("/shop-admins/:id/reset-password", adminAuditH.ResetShopAdminPassword)

		// 平台官方账号
		admin.GET("/platform-accounts", adminAuditH.GetPlatformAccounts)
		admin.POST("/platform-accounts", adminAuditH.CreatePlatformAccount)
		admin.PUT("/platform-accounts/:id", adminAuditH.UpdatePlatformAccount)
		admin.POST("/platform-accounts/:id/disable", adminAuditH.DisablePlatformAccount)

		// 财务(提现/保证金)
		admin.GET("/withdrawals", adminFinanceH.GetWithdrawals)
		admin.POST("/withdrawals/batch", adminFinanceH.BatchWithdraw)
		admin.POST("/withdrawals/:id/approve", adminFinanceH.ApproveWithdrawal)
		admin.POST("/withdrawals/:id/reject", adminFinanceH.RejectWithdrawal)
		admin.GET("/deposits", adminFinanceH.GetDeposits)
		admin.POST("/deposits/:club_id/confirm", adminFinanceH.ConfirmDeposit)
		admin.POST("/deposits/:club_id/refund", adminFinanceH.RefundDeposit)
		admin.PUT("/deposits/config", adminFinanceH.UpdateDepositConfig)

		// 对公小额打款验证(企业入驻)
		admin.GET("/corporate-transfers", adminFinanceH.TransferList)
		admin.POST("/corporate-transfers", adminFinanceH.GenerateTransfer)
		admin.POST("/corporate-transfers/:id/verify", adminFinanceH.VerifyTransfer)

		// 申诉/售后
		admin.GET("/appeals", adminAppealH.GetAppeals)
		admin.GET("/appeals/:id", adminAppealH.GetAppealDetail)
		admin.POST("/appeals/:id/reply", adminAppealH.ReplyAppeal)
		admin.POST("/appeals/:id/close", adminAppealH.CloseAppeal)

		// 仲裁
		admin.GET("/arbitration/cases", adminAppealH.GetArbitrationCases)
		admin.GET("/arbitration/cases/:id", adminAppealH.GetArbitrationCaseDetail)
		admin.POST("/arbitration/cases/:id/judge", adminAppealH.JudgeArbitration)
		admin.GET("/arbitration/rules", adminAppealH.GetArbitrationRules)
		admin.POST("/arbitration/rules", adminAppealH.AddArbitrationRule)
		admin.GET("/arbitration/evidence-templates", adminAppealH.GetEvidenceTemplates)
		admin.POST("/arbitration/evidence-templates", adminAppealH.AddEvidenceTemplate)

		// 风控
		admin.GET("/risk/users", adminRiskH.GetRiskUsers)
		admin.GET("/risk/alerts", adminRiskH.GetRiskAlerts)
		admin.POST("/risk/alerts/:id/handle", adminRiskH.HandleRiskAlert)
		admin.GET("/up-master/certs", adminRiskH.GetUpMasterCerts)
		admin.POST("/up-master/:id/approve", adminRiskH.ApproveUpMaster)
		admin.POST("/up-master/:id/revoke", adminRiskH.RevokeUpMaster)
		admin.GET("/punishments", adminRiskH.GetPunishments)
		admin.POST("/punishments", adminRiskH.CreatePunishment)

		// 聊天审计
		admin.GET("/chat/audit", adminRiskH.GetChatAuditList)
		admin.GET("/chat/sessions/:id/messages", adminRiskH.GetChatMessages)
		admin.GET("/chat/risk-sessions", adminRiskH.GetRiskSessions)
		admin.GET("/chat/keywords", adminRiskH.GetChatKeywords)
		admin.POST("/chat/keywords", adminRiskH.AddChatKeyword)

		// 未成年监管
		admin.GET("/minor/curfew-logs", adminRiskH.GetMinorCurfewLogs)
		admin.GET("/minor/users", adminRiskH.ListMinors)

		// 邀请码
		admin.GET("/invite-codes", adminInviteH.GetInviteCodes)
		admin.GET("/invite-codes/export", adminInviteH.ExportInviteCodes)
		admin.POST("/invite-codes/club", adminInviteH.GenerateClubCode)
		admin.POST("/invite-codes/platform", adminInviteH.GeneratePlatformCode)
		admin.POST("/invite-codes/:id/revoke", adminInviteH.RevokeInviteCode)

		// 营销活动
		admin.GET("/coupon-templates", adminMarketingH.GetCouponTemplates)
		admin.POST("/coupon-templates", adminMarketingH.CreateCouponTemplate)
		admin.GET("/recharge-activities", adminMarketingH.GetRechargeActivities)
		admin.POST("/recharge-activities", adminMarketingH.CreateRechargeActivity)
		admin.GET("/lottery-activities", adminMarketingH.GetLotteryActivities)
		admin.POST("/lottery-activities", adminMarketingH.CreateLotteryActivity)
		admin.GET("/group-buy-activities", adminMarketingH.GetGroupBuyActivities)
		admin.POST("/group-buy-activities", adminMarketingH.CreateGroupBuyActivity)
		admin.GET("/invite-reward-config", adminMarketingH.GetInviteRewardConfig)
		admin.PUT("/invite-reward-config", adminMarketingH.UpdateInviteRewardConfig)

		// 系统/配置/日志
		admin.GET("/system-configs", adminConfigH.GetSystemConfig)
		admin.PUT("/system-configs", adminConfigH.UpdateSystemConfig)
		admin.GET("/operation-logs", adminConfigH.GetOperationLogs)
		admin.GET("/api-monitor", adminConfigH.GetApiMonitor)
		admin.GET("/backups", adminConfigH.GetBackupList)
		admin.POST("/backups", adminConfigH.CreateBackup)
		admin.POST("/backups/restore", adminConfigH.RestoreBackup)
		admin.GET("/gray-release", adminConfigH.GetGrayRelease)
		admin.PUT("/gray-release", adminConfigH.UpdateGrayRelease)
		admin.GET("/agreements", adminConfigH.GetAgreements)
		admin.POST("/agreements", adminConfigH.CreateAgreement)
		admin.GET("/anti-boosting-rules", adminConfigH.GetAntiBoostingRules)
		admin.POST("/anti-boosting-rules", adminConfigH.AddAntiBoostingRule)
		admin.GET("/notifications", adminConfigH.GetNotifications)
		admin.POST("/notifications", adminConfigH.SendNotification)
		admin.GET("/subscribe-templates", adminConfigH.GetSubscribeTemplates)
		admin.POST("/subscribe-templates", adminConfigH.AddSubscribeTemplate)
		admin.GET("/shop-decorations", adminConfigH.GetShopDecorations)
		admin.PUT("/shop-decorations/:shop_id", adminConfigH.UpdateShopDecoration)
		admin.GET("/timeout-rules", adminConfigH.GetTimeoutRules)
		admin.POST("/timeout-rules", adminConfigH.AddTimeoutRule)
		admin.PUT("/timeout-rules/:id", adminConfigH.UpdateTimeoutRule)
		admin.GET("/documents", adminConfigH.GetDocuments)
		admin.POST("/documents", adminConfigH.UploadDocument)
		admin.GET("/documents/:id/versions", adminConfigH.GetDocumentVersions)
		admin.PUT("/documents/:id", adminConfigH.ReplaceDocument)
		admin.DELETE("/documents/:id", adminConfigH.DeleteDocument)
	}

	// ---------------- 内置管理端(俱乐部，需 shop_admin 认证 + 俱乐部范围) ----------------
	shop := api.Group("/shop", shopReq, clubScope)
	{
		shop.GET("/club", shopH.GetClubInfo)
		shop.PUT("/club", shopH.UpdateClubInfo)

		// 入会申请
		shop.GET("/applications", shopH.GetApplications)
		shop.GET("/applications/:id", shopH.GetApplicationDetail)
		shop.POST("/applications/:id/exam/start", shopH.StartExam)
		shop.POST("/applications/:id/exam/submit", shopH.SubmitExamResult)
		shop.POST("/applications/:id/approve", shopH.ApproveApplication)
		shop.POST("/applications/:id/reject", shopH.RejectApplication)

		// 打手管理
		shop.GET("/gamers", shopH.GetGamers)
		shop.GET("/gamers/:id", shopH.GetGamerDetail)
		shop.POST("/gamers/:id/approve", shopH.ApproveGamer)
		shop.POST("/gamers/:id/remove", shopH.RemoveGamer)
		shop.GET("/gamers/:id/evaluations", shopH.GetGamerEvaluations)

		// 管理员管理
		shop.GET("/admins", shopH.GetAdmins)
		shop.POST("/admins", shopH.AddAdmin)
		shop.POST("/admins/:id/remove", shopH.RemoveAdmin)
		shop.POST("/admins/:id/reset-password", shopH.ResetAdminPassword)

		// 群聊管理
		shop.GET("/groups", shopH.GetGroups)
		shop.POST("/groups", shopH.CreateGroup)
		shop.GET("/groups/:id/members", shopH.GetGroupMembers)
		shop.POST("/groups/:id/messages", shopH.SendGroupMessage)
		shop.PUT("/groups/:id/announcement", shopH.PublishAnnouncement)
		shop.GET("/groups/:id/announcement/stats", shopH.GetAnnouncementReadStats)
		shop.POST("/groups/:id/announcement/read", shopH.MarkAnnouncementRead)

		// 抽成比例 + 罚款规则
		shop.PUT("/club/commission", shopH.UpdateCommissionRate)
		shop.GET("/fine-rules", shopH.ListFineRules)
		shop.POST("/fine-rules", shopH.CreateFineRule)
		shop.POST("/fine-rules/:id/revoke", shopH.RevokeFineRule)

		// 风控
		shop.GET("/risk/users", shopH.GetRiskUsers)
		shop.GET("/risk/orders", shopH.GetRiskOrders)

		// 订单管理
		shop.GET("/orders", shopH.GetOrders)
		shop.GET("/orders/failed", shopH.GetFailedOrders)
		shop.GET("/orders/:id", shopH.GetOrderDetail)
		shop.POST("/orders/:id/status", shopH.UpdateOrderStatus)
		shop.POST("/orders/:id/refund", shopH.ProcessRefund)

		// 售后
		shop.GET("/after-sales", shopH.GetAfterSaleList)
		shop.GET("/after-sales/:id", shopH.GetAfterSaleDetail)
		shop.POST("/after-sales/:id/reply", shopH.ReplyAfterSale)
		shop.POST("/after-sales/:id/evidence", shopH.UploadAfterSaleEvidence)

		// 财务
		shop.GET("/withdrawals", shopH.GetWithdrawals)
		shop.GET("/finance/overview", shopH.GetFinanceOverview)
		shop.GET("/finance/details", shopH.GetFinanceDetails)

		// 邀请码
		shop.GET("/invite-codes", shopH.GetInviteCodes)
		shop.POST("/invite-codes", shopH.GenerateInviteCode)
		shop.POST("/invite-codes/:id/revoke", shopH.RevokeInviteCode)
	}
}
