package handler

import (
	"github.com/gin-gonic/gin"
	"github.com/jisan/e-sports-platform/internal/service"
	"github.com/jisan/e-sports-platform/internal/utils"
)

// AdminRiskHandler 平台风控处理器(含聊天审计与未成年监管)
type AdminRiskHandler struct{}

// NewAdminRiskHandler 创建平台风控处理器
func NewAdminRiskHandler() *AdminRiskHandler { return &AdminRiskHandler{} }

// ---------------- 风险用户与预警 ----------------

// GetRiskUsers 平台风险用户列表
// GET /api/v1/admin/risk/users
func (h *AdminRiskHandler) GetRiskUsers(c *gin.Context) {
	page, pageSize := getPage(c)
	level := c.Query("level")
	list, total, err := service.AdminGetRiskUsers(page, pageSize, level)
	if err != nil {
		utils.Fail(c, utils.CodeServerError, err.Error())
		return
	}
	utils.SuccessWithTotal(c, list, total)
}

// GetRiskAlerts AI 风险预警列表
// GET /api/v1/admin/risk/alerts
func (h *AdminRiskHandler) GetRiskAlerts(c *gin.Context) {
	page, pageSize := getPage(c)
	alertType := c.Query("alert_type")
	list, total, err := service.AdminGetRiskAlerts(page, pageSize, alertType)
	if err != nil {
		utils.Fail(c, utils.CodeServerError, err.Error())
		return
	}
	utils.SuccessWithTotal(c, list, total)
}

// handleRiskAlertRequest 处理风险预警请求
type handleRiskAlertRequest struct {
	Status int8   `json:"status" binding:"required"`
	Result string `json:"result"`
}

// HandleRiskAlert 处理 AI 风险预警
// POST /api/v1/admin/risk/alerts/:id/handle
func (h *AdminRiskHandler) HandleRiskAlert(c *gin.Context) {
	id := parseInt64Path(c, "id")
	var req handleRiskAlertRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Fail(c, utils.CodeBadRequest, "参数错误: "+err.Error())
		return
	}
	adminID := getCurrentUserID(c)
	if err := service.AdminHandleRiskAlert(id, adminID, req.Status, req.Result); err != nil {
		utils.Fail(c, utils.CodeBadRequest, err.Error())
		return
	}
	utils.Success(c, gin.H{"msg": "ok"})
}

// ---------------- UP主认证 ----------------

// GetUpMasterCerts UP主认证列表
// GET /api/v1/admin/up-master/certs
func (h *AdminRiskHandler) GetUpMasterCerts(c *gin.Context) {
	page, pageSize := getPage(c)
	list, total, err := service.AdminGetUpMasterCerts(page, pageSize)
	if err != nil {
		utils.Fail(c, utils.CodeServerError, err.Error())
		return
	}
	utils.SuccessWithTotal(c, list, total)
}

// ApproveUpMaster 审核通过UP主
// POST /api/v1/admin/up-master/:id/approve
func (h *AdminRiskHandler) ApproveUpMaster(c *gin.Context) {
	id := parseInt64Path(c, "id")
	if err := service.AdminApproveUpMaster(id); err != nil {
		utils.Fail(c, utils.CodeBadRequest, err.Error())
		return
	}
	utils.Success(c, gin.H{"msg": "ok"})
}

// RevokeUpMaster 撤销UP主认证
// POST /api/v1/admin/up-master/:id/revoke
func (h *AdminRiskHandler) RevokeUpMaster(c *gin.Context) {
	id := parseInt64Path(c, "id")
	if err := service.AdminRevokeUpMaster(id); err != nil {
		utils.Fail(c, utils.CodeBadRequest, err.Error())
		return
	}
	utils.Success(c, gin.H{"msg": "ok"})
}

// ---------------- 处罚记录 ----------------

// createPunishmentRequest 创建处罚记录请求
type createPunishmentRequest struct {
	UserID      int64  `json:"user_id" binding:"required"`
	RiskType    string `json:"risk_type" binding:"required"`
	Description string `json:"description"`
}

// CreatePunishment 创建处罚记录
// POST /api/v1/admin/punishments
func (h *AdminRiskHandler) CreatePunishment(c *gin.Context) {
	var req createPunishmentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Fail(c, utils.CodeBadRequest, "参数错误: "+err.Error())
		return
	}
	if err := service.AdminCreatePunishment(req.UserID, req.RiskType, req.Description); err != nil {
		utils.Fail(c, utils.CodeBadRequest, err.Error())
		return
	}
	utils.Success(c, gin.H{"msg": "ok"})
}

// GetPunishments 处罚记录列表
// GET /api/v1/admin/punishments
func (h *AdminRiskHandler) GetPunishments(c *gin.Context) {
	page, pageSize := getPage(c)
	list, total, err := service.AdminGetPunishments(page, pageSize)
	if err != nil {
		utils.Fail(c, utils.CodeServerError, err.Error())
		return
	}
	utils.SuccessWithTotal(c, list, total)
}

// ---------------- 聊天审计 ----------------

// GetChatAuditList 聊天审计列表
// GET /api/v1/admin/chat/audit
func (h *AdminRiskHandler) GetChatAuditList(c *gin.Context) {
	page, pageSize := getPage(c)
	list, total, err := service.AdminGetChatAuditList(page, pageSize)
	if err != nil {
		utils.Fail(c, utils.CodeServerError, err.Error())
		return
	}
	utils.SuccessWithTotal(c, list, total)
}

// GetChatMessages 管理端查看会话消息
// GET /api/v1/admin/chat/sessions/:id/messages
func (h *AdminRiskHandler) GetChatMessages(c *gin.Context) {
	sessionID := parseInt64Path(c, "id")
	page, pageSize := getPage(c)
	list, total, err := service.AdminGetChatMessages(sessionID, page, pageSize)
	if err != nil {
		utils.Fail(c, utils.CodeServerError, err.Error())
		return
	}
	utils.SuccessWithTotal(c, list, total)
}

// GetRiskSessions 风险会话列表
// GET /api/v1/admin/chat/risk-sessions
func (h *AdminRiskHandler) GetRiskSessions(c *gin.Context) {
	page, pageSize := getPage(c)
	list, total, err := service.AdminGetRiskSessions(page, pageSize)
	if err != nil {
		utils.Fail(c, utils.CodeServerError, err.Error())
		return
	}
	utils.SuccessWithTotal(c, list, total)
}

// addChatKeywordRequest 新增聊天关键词请求
type addChatKeywordRequest struct {
	Keyword   string `json:"keyword" binding:"required"`
	MatchType string `json:"match_type"`
}

// AddChatKeyword 新增聊天关键词
// POST /api/v1/admin/chat/keywords
func (h *AdminRiskHandler) AddChatKeyword(c *gin.Context) {
	var req addChatKeywordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Fail(c, utils.CodeBadRequest, "参数错误: "+err.Error())
		return
	}
	if err := service.AdminAddChatKeyword(req.Keyword, req.MatchType); err != nil {
		utils.Fail(c, utils.CodeBadRequest, err.Error())
		return
	}
	utils.Success(c, gin.H{"msg": "ok"})
}

// GetChatKeywords 关键词列表
// GET /api/v1/admin/chat/keywords
func (h *AdminRiskHandler) GetChatKeywords(c *gin.Context) {
	page, pageSize := getPage(c)
	list, total, err := service.AdminGetChatKeywords(page, pageSize)
	if err != nil {
		utils.Fail(c, utils.CodeServerError, err.Error())
		return
	}
	utils.SuccessWithTotal(c, list, total)
}

// ---------------- 未成年监管 ----------------

// GetMinorCurfewLogs 未成年宵禁拦截日志
// GET /api/v1/admin/minor/curfew-logs
func (h *AdminRiskHandler) GetMinorCurfewLogs(c *gin.Context) {
	page, pageSize := getPage(c)
	userID := parseInt64Query(c, "user_id", 0)
	list, total, err := service.GetMinorCurfewLogs(userID, page, pageSize)
	if err != nil {
		utils.Fail(c, utils.CodeServerError, err.Error())
		return
	}
	utils.SuccessWithTotal(c, list, total)
}

// ListMinors 未成年用户列表
// GET /api/v1/admin/minor/users
func (h *AdminRiskHandler) ListMinors(c *gin.Context) {
	page, pageSize := getPage(c)
	list, total, err := service.ListMinors(page, pageSize)
	if err != nil {
		utils.Fail(c, utils.CodeServerError, err.Error())
		return
	}
	utils.SuccessWithTotal(c, list, total)
}
