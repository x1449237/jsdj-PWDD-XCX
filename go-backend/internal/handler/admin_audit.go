package handler

import (
	"github.com/gin-gonic/gin"
	"github.com/jisan/e-sports-platform/internal/model"
	"github.com/jisan/e-sports-platform/internal/service"
	"github.com/jisan/e-sports-platform/internal/utils"
)

// AdminAuditHandler 平台审核处理器(俱乐部/打手/分销商/派单员/管理端账号)
type AdminAuditHandler struct{}

// NewAdminAuditHandler 创建平台审核处理器
func NewAdminAuditHandler() *AdminAuditHandler { return &AdminAuditHandler{} }

// AuditClubs 俱乐部审核列表
// GET /api/v1/admin/clubs/audit
func (h *AdminAuditHandler) AuditClubs(c *gin.Context) {
	page, pageSize := getPage(c)
	status := parseInt8Query(c, "status", -1)
	keyword := c.Query("keyword")
	list, total, err := service.AdminAuditClubs(page, pageSize, status, keyword)
	if err != nil {
		utils.Fail(c, utils.CodeServerError, err.Error())
		return
	}
	utils.SuccessWithTotal(c, list, total)
}

// ApproveClub 审核通过俱乐部
// POST /api/v1/admin/clubs/:id/approve
func (h *AdminAuditHandler) ApproveClub(c *gin.Context) {
	clubID := parseInt64Path(c, "id")
	adminID := getCurrentUserID(c)
	if err := service.AdminApproveClub(clubID, adminID); err != nil {
		utils.Fail(c, utils.CodeBadRequest, err.Error())
		return
	}
	utils.Success(c, gin.H{"msg": "ok"})
}

// rejectClubRequest 驳回俱乐部请求
type rejectClubRequest struct {
	Reason string `json:"reason"`
}

// RejectClub 驳回俱乐部
// POST /api/v1/admin/clubs/:id/reject
func (h *AdminAuditHandler) RejectClub(c *gin.Context) {
	clubID := parseInt64Path(c, "id")
	var req rejectClubRequest
	_ = c.ShouldBindJSON(&req)
	if err := service.AdminRejectClub(clubID, req.Reason); err != nil {
		utils.Fail(c, utils.CodeBadRequest, err.Error())
		return
	}
	utils.Success(c, gin.H{"msg": "ok"})
}

// FreezeClub 冻结俱乐部(级联禁用功能)
// POST /api/v1/admin/clubs/:id/freeze
func (h *AdminAuditHandler) FreezeClub(c *gin.Context) {
	clubID := parseInt64Path(c, "id")
	var req rejectClubRequest
	_ = c.ShouldBindJSON(&req)
	if err := service.AdminFreezeClub(clubID, req.Reason); err != nil {
		utils.Fail(c, utils.CodeBadRequest, err.Error())
		return
	}
	utils.Success(c, gin.H{"msg": "ok"})
}

// CancelClub 安全注销俱乐部
// POST /api/v1/admin/clubs/:id/cancel
func (h *AdminAuditHandler) CancelClub(c *gin.Context) {
	clubID := parseInt64Path(c, "id")
	var req rejectClubRequest
	_ = c.ShouldBindJSON(&req)
	if err := service.AdminCancelClub(clubID, req.Reason); err != nil {
		utils.Fail(c, utils.CodeBadRequest, err.Error())
		return
	}
	utils.Success(c, gin.H{"msg": "ok"})
}

// UnfreezeClub 解冻俱乐部(级联恢复)
// POST /api/v1/admin/clubs/:id/unfreeze
func (h *AdminAuditHandler) UnfreezeClub(c *gin.Context) {
	clubID := parseInt64Path(c, "id")
	if err := service.AdminUnfreezeClub(clubID); err != nil {
		utils.Fail(c, utils.CodeBadRequest, err.Error())
		return
	}
	utils.Success(c, gin.H{"msg": "ok"})
}

// HideVBadge 平台手动隐藏俱乐部 V 标
// POST /api/v1/admin/clubs/:id/vbadge/hide
func (h *AdminAuditHandler) HideVBadge(c *gin.Context) {
	clubID := parseInt64Path(c, "id")
	if err := service.AdminHideVBadge(clubID); err != nil {
		utils.Fail(c, utils.CodeBadRequest, err.Error())
		return
	}
	utils.Success(c, gin.H{"msg": "ok"})
}

// RestoreVBadge 平台恢复被手动隐藏的俱乐部 V 标
// POST /api/v1/admin/clubs/:id/vbadge/restore
func (h *AdminAuditHandler) RestoreVBadge(c *gin.Context) {
	clubID := parseInt64Path(c, "id")
	if err := service.AdminRestoreVBadge(clubID); err != nil {
		utils.Fail(c, utils.CodeBadRequest, err.Error())
		return
	}
	utils.Success(c, gin.H{"msg": "ok"})
}

// AuditClubsFiltered 俱乐部审核列表(多条件筛选)
// GET /api/v1/admin/clubs/audit-filter
func (h *AdminAuditHandler) AuditClubsFiltered(c *gin.Context) {
	page, pageSize := getPage(c)
	f := service.ClubAuditFilter{
		Status:        parseInt8Query(c, "status", -1),
		Type:          parseInt8Query(c, "type", -1),
		VBadgeType:    parseInt8Query(c, "v_badge_type", -1),
		DepositStatus: parseInt8Query(c, "deposit_status", -1),
		Keyword:       c.Query("keyword"),
	}
	list, total, err := service.AdminAuditClubsFiltered(page, pageSize, f)
	if err != nil {
		utils.Fail(c, utils.CodeServerError, err.Error())
		return
	}
	utils.SuccessWithTotal(c, list, total)
}

// GetClubChangeLogs 俱乐部资料修改日志(入驻/资料变更审计溯源)
// GET /api/v1/admin/clubs/:id/change-logs
func (h *AdminAuditHandler) GetClubChangeLogs(c *gin.Context) {
	clubID := parseInt64Path(c, "id")
	page, pageSize := getPage(c)
	list, total, err := service.AdminGetClubChangeLogs(clubID, page, pageSize)
	if err != nil {
		utils.Fail(c, utils.CodeServerError, err.Error())
		return
	}
	utils.SuccessWithTotal(c, list, total)
}

// ReviewFineRule 平台审核罚款规则备案
// POST /api/v1/admin/fine-rules/:id/review
func (h *AdminAuditHandler) ReviewFineRule(c *gin.Context) {
	ruleID := parseInt64Path(c, "id")
	reviewerID := getCurrentUserID(c)
	var req struct {
		Approve bool   `json:"approve"`
		Note    string `json:"note"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Fail(c, utils.CodeBadRequest, "参数错误: "+err.Error())
		return
	}
	if err := service.AdminReviewFineRule(ruleID, reviewerID, req.Approve, req.Note); err != nil {
		utils.Fail(c, utils.CodeBadRequest, err.Error())
		return
	}
	utils.Success(c, gin.H{"msg": "ok"})
}

// AuditPlayers 打手审核列表
// GET /api/v1/admin/players/audit
func (h *AdminAuditHandler) AuditPlayers(c *gin.Context) {
	page, pageSize := getPage(c)
	keyword := c.Query("keyword")
	list, total, err := service.AdminAuditPlayers(page, pageSize, keyword)
	if err != nil {
		utils.Fail(c, utils.CodeServerError, err.Error())
		return
	}
	utils.SuccessWithTotal(c, list, total)
}

// ApprovePlayer 审核通过打手
// POST /api/v1/admin/players/:id/approve
func (h *AdminAuditHandler) ApprovePlayer(c *gin.Context) {
	playerID := parseInt64Path(c, "id")
	if err := service.AdminApprovePlayer(playerID); err != nil {
		utils.Fail(c, utils.CodeBadRequest, err.Error())
		return
	}
	utils.Success(c, gin.H{"msg": "ok"})
}

// rejectPlayerRequest 驳回打手请求
type rejectPlayerRequest struct {
	Reason string `json:"reason"`
}

// RejectPlayer 驳回打手
// POST /api/v1/admin/players/:id/reject
func (h *AdminAuditHandler) RejectPlayer(c *gin.Context) {
	playerID := parseInt64Path(c, "id")
	var req rejectPlayerRequest
	_ = c.ShouldBindJSON(&req)
	if err := service.AdminRejectPlayer(playerID, req.Reason); err != nil {
		utils.Fail(c, utils.CodeBadRequest, err.Error())
		return
	}
	utils.Success(c, gin.H{"msg": "ok"})
}

// AuditDistributors 分销商审核列表
// GET /api/v1/admin/distributors/audit
func (h *AdminAuditHandler) AuditDistributors(c *gin.Context) {
	page, pageSize := getPage(c)
	keyword := c.Query("keyword")
	list, total, err := service.AdminAuditDistributors(page, pageSize, keyword)
	if err != nil {
		utils.Fail(c, utils.CodeServerError, err.Error())
		return
	}
	utils.SuccessWithTotal(c, list, total)
}

// ApproveDistributor 审核通过分销商
// POST /api/v1/admin/distributors/:id/approve
func (h *AdminAuditHandler) ApproveDistributor(c *gin.Context) {
	id := parseInt64Path(c, "id")
	if err := service.AdminApproveDistributor(id); err != nil {
		utils.Fail(c, utils.CodeBadRequest, err.Error())
		return
	}
	utils.Success(c, gin.H{"msg": "ok"})
}

// AuditDispatchers 派单员审核列表
// GET /api/v1/admin/dispatchers/audit
func (h *AdminAuditHandler) AuditDispatchers(c *gin.Context) {
	page, pageSize := getPage(c)
	keyword := c.Query("keyword")
	list, total, err := service.AdminAuditDispatchers(page, pageSize, keyword)
	if err != nil {
		utils.Fail(c, utils.CodeServerError, err.Error())
		return
	}
	utils.SuccessWithTotal(c, list, total)
}

// ApproveDispatcher 审核通过派单员
// POST /api/v1/admin/dispatchers/:id/approve
func (h *AdminAuditHandler) ApproveDispatcher(c *gin.Context) {
	id := parseInt64Path(c, "id")
	if err := service.AdminApproveDispatcher(id); err != nil {
		utils.Fail(c, utils.CodeBadRequest, err.Error())
		return
	}
	utils.Success(c, gin.H{"msg": "ok"})
}

// GetShopAdmins 内置管理端账号列表
// GET /api/v1/admin/shop-admins
func (h *AdminAuditHandler) GetShopAdmins(c *gin.Context) {
	page, pageSize := getPage(c)
	list, total, err := service.AdminGetShopAdmins(page, pageSize)
	if err != nil {
		utils.Fail(c, utils.CodeServerError, err.Error())
		return
	}
	utils.SuccessWithTotal(c, list, total)
}

// DisableShopAdmin 禁用内置管理端账号
// POST /api/v1/admin/shop-admins/:id/disable
func (h *AdminAuditHandler) DisableShopAdmin(c *gin.Context) {
	id := parseInt64Path(c, "id")
	if err := service.AdminDisableShopAdmin(id); err != nil {
		utils.Fail(c, utils.CodeBadRequest, err.Error())
		return
	}
	utils.Success(c, gin.H{"msg": "ok"})
}

// EnableShopAdmin 启用内置管理端账号
// POST /api/v1/admin/shop-admins/:id/enable
func (h *AdminAuditHandler) EnableShopAdmin(c *gin.Context) {
	id := parseInt64Path(c, "id")
	if err := service.AdminEnableShopAdmin(id); err != nil {
		utils.Fail(c, utils.CodeBadRequest, err.Error())
		return
	}
	utils.Success(c, gin.H{"msg": "ok"})
}

// resetShopAdminPasswordRequest 重置内置管理端密码请求
type resetShopAdminPasswordRequest struct {
	Password string `json:"password" binding:"required"`
}

// ResetShopAdminPassword 平台重置内置管理端密码
// POST /api/v1/admin/shop-admins/:id/reset-password
func (h *AdminAuditHandler) ResetShopAdminPassword(c *gin.Context) {
	id := parseInt64Path(c, "id")
	var req resetShopAdminPasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Fail(c, utils.CodeBadRequest, "参数错误: "+err.Error())
		return
	}
	if err := service.AdminResetShopAdminPassword(id, req.Password); err != nil {
		utils.Fail(c, utils.CodeBadRequest, err.Error())
		return
	}
	utils.Success(c, gin.H{"msg": "ok"})
}

// AddAdminShopAccountRequest 平台代建内置管理端账号请求
type AddAdminShopAccountRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
	ClubID   int64  `json:"club_id" binding:"required"`
	RealName string `json:"real_name"`
	Phone    string `json:"phone"`
	Role     int8   `json:"role"`
}

// AdminAddShopAccount 平台代建内置管理端账号(复用 ShopAddAdmin)
// POST /api/v1/admin/shop-admins
func (h *AdminAuditHandler) AdminAddShopAccount(c *gin.Context) {
	var req AddAdminShopAccountRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Fail(c, utils.CodeBadRequest, "参数错误: "+err.Error())
		return
	}
	a, err := service.ShopAddAdmin(req.ClubID, req.Username, req.Password, req.RealName, req.Phone, req.Role)
	if err != nil {
		utils.Fail(c, utils.CodeBadRequest, err.Error())
		return
	}
	utils.Success(c, a)
}

// GetPlatformAccounts 平台官方账号列表
// GET /api/v1/admin/platform-accounts
func (h *AdminAuditHandler) GetPlatformAccounts(c *gin.Context) {
	list, err := service.AdminGetPlatformAccounts()
	if err != nil {
		utils.Fail(c, utils.CodeServerError, err.Error())
		return
	}
	utils.Success(c, list)
}

// createPlatformAccountRequest 创建平台官方账号请求
type createPlatformAccountRequest struct {
	Username string `json:"username" binding:"required"`
	Nickname string `json:"nickname"`
	Avatar   string `json:"avatar"`
}

// CreatePlatformAccount 创建平台官方账号
// POST /api/v1/admin/platform-accounts
func (h *AdminAuditHandler) CreatePlatformAccount(c *gin.Context) {
	var req createPlatformAccountRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Fail(c, utils.CodeBadRequest, "参数错误: "+err.Error())
		return
	}
	a := &model.PlatformOfficialAccount{
		Username: req.Username,
		Nickname: req.Nickname,
		Avatar:   req.Avatar,
	}
	if err := service.AdminCreatePlatformAccount(a); err != nil {
		utils.Fail(c, utils.CodeBadRequest, err.Error())
		return
	}
	utils.Success(c, a)
}

// UpdatePlatformAccount 更新平台官方账号
// PUT /api/v1/admin/platform-accounts/:id
func (h *AdminAuditHandler) UpdatePlatformAccount(c *gin.Context) {
	id := parseInt64Path(c, "id")
	var req createPlatformAccountRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Fail(c, utils.CodeBadRequest, "参数错误: "+err.Error())
		return
	}
	fields := map[string]interface{}{
		"username": req.Username, "nickname": req.Nickname, "avatar": req.Avatar,
	}
	if err := service.AdminUpdatePlatformAccount(id, fields); err != nil {
		utils.Fail(c, utils.CodeBadRequest, err.Error())
		return
	}
	utils.Success(c, gin.H{"msg": "ok"})
}

// DisablePlatformAccount 停用平台官方账号
// POST /api/v1/admin/platform-accounts/:id/disable
func (h *AdminAuditHandler) DisablePlatformAccount(c *gin.Context) {
	id := parseInt64Path(c, "id")
	if err := service.AdminDisablePlatformAccount(id); err != nil {
		utils.Fail(c, utils.CodeBadRequest, err.Error())
		return
	}
	utils.Success(c, gin.H{"msg": "ok"})
}
