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

	// 平台方管理人员 handler
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

	// 俱乐部小助手APP扩展功能handler
	extH := handler.NewExtensionHandler()

	// 平台方管理人员 IM 会话归类清单 handler
	platformIMH := handler.NewPlatformIMHandler()

	// 99~582 需求扩展 handler
	platform99H := handler.NewPlatform99582Handler()

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

		// ===== 俱乐部小助手APP扩展 - 俱乐部内置管理端 =====
		// 主页装修
		shop.GET("/club-ext/home-decoration", extH.GetHomeDecoration)
		shop.PUT("/club-ext/home-decoration", extH.UpdateHomeDecoration)
		// 技能项目
		shop.GET("/club-ext/services", extH.ListClubServices)
		shop.POST("/club-ext/services", extH.CreateClubService)
		shop.PUT("/club-ext/services/:id", extH.UpdateClubService)
		shop.DELETE("/club-ext/services/:id", extH.DeleteClubService)
		// 成员名片/档案
		shop.GET("/club-ext/member-card", extH.GetMemberCard)
		shop.PUT("/club-ext/member-card", extH.UpdateMemberCard)
		shop.GET("/club-ext/member-profile", extH.GetMemberProfile)
		shop.PUT("/club-ext/member-profile", extH.UpdateMemberProfile)
		// 权限操作日志
		shop.GET("/club-ext/permission-logs", extH.ListPermissionLogs)
		// 退会申请
		shop.GET("/club-ext/resignations", extH.ListResignations)
		shop.PUT("/club-ext/resignations/:id/audit", extH.AuditResignation)
		// 黑名单
		shop.GET("/club-ext/blacklists", extH.ListBlacklists)
		shop.POST("/club-ext/blacklists", extH.AddBlacklist)
		shop.DELETE("/club-ext/blacklists/:userId", extH.RemoveBlacklist)
		// 积分体系
		shop.GET("/club-ext/point-rules", extH.ListPointRules)
		shop.PUT("/club-ext/point-rules/:id", extH.UpdatePointRule)
		shop.GET("/club-ext/point-logs", extH.ListPointLogs)
		// 团费规则
		shop.GET("/club-ext/fee-rule", extH.GetFeeRule)
		shop.PUT("/club-ext/fee-rule", extH.SaveFeeRule)
		// 招募卡片
		shop.GET("/club-ext/recruit-cards", extH.ListRecruitCards)
		shop.POST("/club-ext/recruit-cards", extH.SaveRecruitCard)
		// 管理员待办
		shop.GET("/club-ext/todos", extH.ListAdminTodos)
		shop.PUT("/club-ext/todos/:id/complete", extH.CompleteAdminTodo)
		// 多游戏分区
		shop.GET("/club-ext/game-zones", extH.ListGameZones)
		shop.POST("/club-ext/game-zones", extH.SaveGameZone)
		// 临时抽成
		shop.GET("/club-ext/temp-commission", extH.ListTempCommissionRules)
		shop.POST("/club-ext/temp-commission", extH.CreateTempCommissionRule)
		// 请假
		shop.GET("/club-ext/leaves", extH.ListLeaves)
		shop.PUT("/club-ext/leaves/:id/audit", extH.AuditLeave)
		// 资料修改审核
		shop.GET("/club-ext/change-requests", extH.ListChangeRequests)
		shop.PUT("/club-ext/change-requests/:id/audit", extH.AuditChangeRequest)
		// 优先派单
		shop.GET("/club-ext/priority-dispatch", extH.ListPriorityDispatch)
		shop.POST("/club-ext/priority-dispatch", extH.SetPriority)
		// 内部资源
		shop.GET("/club-ext/internal-resources", extH.ListInternalResources)
		shop.POST("/club-ext/internal-resources", extH.CreateInternalResource)
		// 客户归属
		shop.GET("/club-ext/customer-relations", extH.ListCustomerRelations)
		shop.POST("/club-ext/customer-relations", extH.SaveCustomerRelation)
		// 模板话术
		shop.GET("/club-ext/template-phrases", extH.ListTemplatePhrases)
		shop.POST("/club-ext/template-phrases", extH.SaveTemplatePhrase)
		shop.DELETE("/club-ext/template-phrases/:id", extH.DeleteTemplatePhrase)
		// 业绩排行
		shop.GET("/club-ext/rankings", extH.GetRanking)
		// 俱乐部设置(营业时间/最低价/接单门槛/招募大厅/展示开关等)
		shop.PUT("/club-ext/settings", extH.UpdateClubSettings)
		// 创始人转移
		shop.POST("/club-ext/transfer-founder", extH.TransferFounder)
		// 成员禁单/角色
		shop.PUT("/club-ext/members/:userId/ban", extH.SetMemberBan)
		shop.PUT("/club-ext/members/:userId/role", extH.SetMemberRole)
		// 快捷回复话术
		shop.GET("/club-ext/quick-replies", extH.ListQuickReplies)
		shop.POST("/club-ext/quick-replies", extH.SaveQuickReply)
		// 群成员禁言
		shop.POST("/club-ext/groups/mute", extH.MuteGroupMember)
		shop.DELETE("/club-ext/groups/unmute", extH.UnmuteGroupMember)
		// 订单审核(改价/转单/延时/部分退款/异常/归档)
		shop.PUT("/club-ext/orders/partial-refunds/:id/audit", extH.AuditPartialRefund)
		shop.PUT("/club-ext/orders/extensions/:id/audit", extH.AuditOrderExtension)
		shop.PUT("/club-ext/orders/transfers/:id/audit", extH.AuditOrderTransfer)
		shop.PUT("/club-ext/orders/price-changes/:id/audit", extH.AuditOrderPriceChange)
		shop.PUT("/club-ext/orders/:id/abnormal", extH.MarkOrderAbnormal)
		shop.PUT("/club-ext/orders/:id/archive", extH.ArchiveOrder)
		// 活动弹窗(俱乐部管理端)
		shop.GET("/club-ext/activity-popups", extH.AdminListActivityPopups)
		shop.POST("/club-ext/activity-popups", extH.AdminSaveActivityPopup)
		// 财务台账/月结/返点
		shop.GET("/finance-ext/ledger", extH.ListFinanceLedger)
		shop.GET("/finance-ext/settlements", extH.ListMonthlySettlements)
		shop.GET("/finance-ext/rebates", extH.ListRebateRecords)
		shop.POST("/finance-ext/rebates", extH.CreateRebateRecord)
		shop.PUT("/finance-ext/rebates/:id/audit", extH.AuditRebateRecord)
	}

	// ===== 俱乐部小助手APP扩展 - 玩家端(需登录) =====
	userAPI.POST("/order-ext/templates", extH.CreateOrderTemplate)
	userAPI.GET("/order-ext/templates", extH.ListOrderTemplates)
	userAPI.DELETE("/order-ext/templates/:id", extH.DeleteOrderTemplate)
	userAPI.POST("/order-ext/supplements", extH.CreateSupplement)
	userAPI.GET("/order-ext/supplements", extH.ListSupplements)
	userAPI.POST("/order-ext/partial-refunds", extH.CreatePartialRefund)
	userAPI.GET("/order-ext/partial-refunds", extH.ListPartialRefunds)
	userAPI.POST("/order-ext/remarks", extH.AddOrderRemark)
	userAPI.GET("/order-ext/remarks", extH.ListOrderRemarks)
	userAPI.POST("/order-ext/extensions", extH.CreateOrderExtension)
	userAPI.POST("/order-ext/transfers", extH.CreateOrderTransfer)
	userAPI.POST("/order-ext/price-changes", extH.CreateOrderPriceChange)
	userAPI.GET("/order-ext/price-logs", extH.ListOrderPriceLogs)
	userAPI.POST("/order-ext/tags", extH.AddOrderTag)
	userAPI.DELETE("/order-ext/tags", extH.RemoveOrderTag)
	userAPI.GET("/order-ext/tags", extH.ListOrderTags)
	userAPI.GET("/order-ext/refund-ledger", extH.ListRefundLedger)
	// 退会/请假/资料变更申报
	userAPI.POST("/user/club-resignation", extH.CreateResignation)
	userAPI.POST("/user/club-leave", extH.CreateLeave)
	userAPI.POST("/user/club-change-request", extH.CreateChangeRequest)
	// 收藏俱乐部
	userAPI.POST("/user/favorite-clubs", extH.FavoriteClub)
	userAPI.DELETE("/user/favorite-clubs/:clubId", extH.UnfavoriteClub)
	userAPI.GET("/user/favorite-clubs", extH.ListFavoriteClubs)
	// 钱包/存单
	userAPI.GET("/user/wallet-logs", extH.ListWalletChangeLogs)
	userAPI.GET("/user/deposits", extH.ListDeposits)
	userAPI.POST("/user/deposits", extH.CreateDeposit)
	// 意见反馈
	userAPI.POST("/user/feedbacks", extH.CreateFeedback)
	userAPI.GET("/user/feedbacks", extH.ListUserFeedbacks)
	// 拉黑打手
	userAPI.POST("/user/blocklist", extH.BlockPlayer)
	userAPI.DELETE("/user/blocklist/:playerId", extH.UnblockPlayer)
	// 通知设置
	userAPI.GET("/user/notification-settings", extH.GetNotificationSettings)
	userAPI.PUT("/user/notification-settings", extH.UpdateNotificationSettings)
	// 活动弹窗(玩家端)
	userAPI.GET("/activity-popups", extH.ListActivityPopups)
	// 聊天扩展(玩家端)
        userAPI.GET("/chat-ext/group-files", extH.ListGroupFiles)
        userAPI.POST("/chat-ext/reports", extH.CreateChatReport)
        userAPI.PUT("/chat-ext/sessions/:id/pin", extH.TogglePinSession)

        // ===== 平台方管理人员 IM 会话归类清单 =====
        pim := admin.Group("/platform-im")
        {
                // 一~二~三 会话列表+筛选+搜索
                pim.GET("/sessions", platformIMH.ListSessions)
                pim.POST("/sessions/:id/close", platformIMH.MarkSessionClosed)
                // 四 标签+备注+星标
                pim.GET("/tags", platformIMH.ListTags)
                pim.POST("/tags", platformIMH.CreateTag)
                pim.PUT("/tags/:id", platformIMH.UpdateTag)
                pim.DELETE("/tags/:id", platformIMH.DeleteTag)
                pim.POST("/sessions/:id/tags", platformIMH.TagSession)
                pim.DELETE("/sessions/:id/tags/:tag_id", platformIMH.UntagSession)
                pim.PUT("/sessions/:id/note", platformIMH.UpsertNote)
                // 三 搜索历史
                pim.GET("/search-history", platformIMH.ListSearchHistory)
                pim.POST("/search-history/clear", platformIMH.ClearSearchHistory)
                // 七 工作台
                pim.GET("/workbench", platformIMH.GetWorkbench)
                pim.PUT("/workbench/layout", platformIMH.SaveWorkbenchLayout)
                // 八~九 快捷话术+证据
                pim.GET("/quick-replies", platformIMH.ListQuickReplies)
                pim.POST("/quick-replies", platformIMH.CreateQuickReply)
                pim.PUT("/quick-replies/:id", platformIMH.UpdateQuickReply)
                pim.DELETE("/quick-replies/:id", platformIMH.DeleteQuickReply)
                pim.GET("/sessions/:id/evidence", platformIMH.ListEvidenceMessages)
                // 十~十一 样式下发(后端统管样式,前端仅渲染)
                pim.GET("/styles/me", platformIMH.PullMyStyle)
                pim.GET("/styles/all", platformIMH.PullAllStyles)
                // 六 群聊批量管理
                pim.PUT("/groups/:id/setting", platformIMH.UpdateGroupSetting)
                pim.POST("/groups/batch-mute", platformIMH.BatchGroupMute)
        }

        // ===== 俱乐部小助手APP扩展 - 平台超级管理员端 =====
	admin.GET("/punishment-templates", extH.ListPunishmentTemplates)
	admin.POST("/punishment-templates", extH.SavePunishmentTemplate)
	admin.GET("/feedbacks", extH.ListAllFeedbacks)
	admin.PUT("/feedbacks/:id/reply", extH.ReplyFeedback)
	admin.GET("/festival-templates", extH.ListFestivalTemplates)
	admin.POST("/festival-templates", extH.SaveFestivalTemplate)
	admin.GET("/promo-channels", extH.ListPromoChannels)
	admin.POST("/promo-channels", extH.CreatePromoChannel)
	admin.GET("/chat-reports", extH.ListChatReports)
	admin.PUT("/chat-reports/:id/handle", extH.HandleChatReport)

	// ===== 99~582 需求配套 HTTP 接口 =====
	// 公开接口
	pub := api.Group("")
	{
		pub.GET("/platform/join-switch", platform99H.GetClubJoinSwitch)
		pub.GET("/public/abbr-help", platform99H.GetAbbrHelpPage)
		pub.GET("/badge-configs", platform99H.GetAllBadgeRenderConfigs)
	}
	// 登录态: 用户侧 99~582 配套
	userAPI.POST("/orders/gen-club-order-no", platform99H.GenerateClubOrderNo)
	userAPI.POST("/registrations/personal/files", platform99H.UpsertPersonalRegFiles)
	userAPI.POST("/registrations/enterprise/files", platform99H.UpsertEnterpriseRegFiles)
	userAPI.POST("/im/typing", platform99H.UpsertTypingStatus)
	userAPI.GET("/im/:session_id/typing", platform99H.GetTypingUsers)
	userAPI.GET("/im/preference", platform99H.GetUserIMPreference)
	userAPI.POST("/im/preference", platform99H.SaveUserIMPreference)
	userAPI.POST("/favorites", platform99H.ToggleFavorite)
	userAPI.GET("/favorites", platform99H.ListFavorites)
	userAPI.GET("/announcements", platform99H.ListAnnouncements)
	userAPI.POST("/announcements/:id/read", platform99H.MarkAnnouncementRead)
	userAPI.GET("/global-group/:group_id/config", platform99H.GetGlobalGroupConfig)
	userAPI.POST("/global-group/:group_id/at-all-consume", platform99H.IncrementAtAllUsed)
	userAPI.POST("/after-sale/:session_id/marks", platform99H.MarkAfterSaleMessage)
	userAPI.GET("/after-sale/:session_id/marks", platform99H.ListAfterSaleMarks)
	// 公告创建(俱乐部管理员/平台超管)
	userAPI.POST("/announcements", platform99H.CreateAnnouncement)
	userAPI.POST("/export/watermark-log", platform99H.CreateExportWatermarkLog)
	userAPI.GET("/export/watermark-log/:export_no", platform99H.QueryExportWatermark)
	userAPI.POST("/watermark/detect-log", platform99H.InsertDetectLog)

	// Web 超管侧
	admin.POST("/platform/join-switch", platform99H.SetClubJoinSwitch)
	admin.GET("/admin/enterprise/verify-amount/random", platform99H.GenerateRandomVerifyAmount)
	admin.GET("/admin/global-param", platform99H.GetGlobalParam)
	admin.POST("/admin/global-param", platform99H.SetGlobalParam)
	admin.POST("/admin/global-group/config", platform99H.UpsertGlobalGroupConfig)
	admin.POST("/admin/badge-config", platform99H.UpdateBadgeRenderConfig)
	admin.POST("/admin/announcements/:id/status", platform99H.UpdateAnnouncementStatus)
}
