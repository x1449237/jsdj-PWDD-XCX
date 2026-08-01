package handler

import (
	"github.com/gin-gonic/gin"
	"github.com/jisan/e-sports-platform/internal/model"
	"github.com/jisan/e-sports-platform/internal/service"
	"github.com/jisan/e-sports-platform/internal/utils"
)

// AdminInviteHandler 平台邀请码处理器
type AdminInviteHandler struct{}

// NewAdminInviteHandler 创建平台邀请码处理器
func NewAdminInviteHandler() *AdminInviteHandler { return &AdminInviteHandler{} }

// GetInviteCodes 平台邀请码列表
// GET /api/v1/admin/invite-codes
func (h *AdminInviteHandler) GetInviteCodes(c *gin.Context) {
	page, pageSize := getPage(c)
	codeType := c.Query("type")
	status := c.Query("status")
	list, total, err := service.AdminGetInviteCodes(page, pageSize, codeType, status)
	if err != nil {
		utils.Fail(c, utils.CodeServerError, err.Error())
		return
	}
	utils.SuccessWithTotal(c, list, total)
}

// generateClubCodeRequest 生成俱乐部邀请码请求
type generateClubCodeRequest struct {
	ClubID     int64  `json:"club_id" binding:"required"`
	Role       string `json:"role"`
	MaxUses    int    `json:"max_uses"`
	ExpireDays int    `json:"expire_days"`
}

// GenerateClubCode 平台生成俱乐部邀请码
// POST /api/v1/admin/invite-codes/club
func (h *AdminInviteHandler) GenerateClubCode(c *gin.Context) {
	var req generateClubCodeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Fail(c, utils.CodeBadRequest, "参数错误: "+err.Error())
		return
	}
	adminID := getCurrentUserID(c)
	code, err := service.AdminGenerateClubCode(req.ClubID, req.Role, req.MaxUses, req.ExpireDays, adminID)
	if err != nil {
		utils.Fail(c, utils.CodeBadRequest, err.Error())
		return
	}
	utils.Success(c, code)
}

// generatePlatformCodeRequest 生成通用邀请码请求
type generatePlatformCodeRequest struct {
	MaxUses    int `json:"max_uses"`
	ExpireDays int `json:"expire_days"`
}

// GeneratePlatformCode 平台生成通用邀请码
// POST /api/v1/admin/invite-codes/platform
func (h *AdminInviteHandler) GeneratePlatformCode(c *gin.Context) {
	var req generatePlatformCodeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Fail(c, utils.CodeBadRequest, "参数错误: "+err.Error())
		return
	}
	adminID := getCurrentUserID(c)
	code, err := service.AdminGeneratePlatformCode(req.MaxUses, req.ExpireDays, adminID)
	if err != nil {
		utils.Fail(c, utils.CodeBadRequest, err.Error())
		return
	}
	utils.Success(c, code)
}

// RevokeInviteCode 平台撤销邀请码
// POST /api/v1/admin/invite-codes/:id/revoke
func (h *AdminInviteHandler) RevokeInviteCode(c *gin.Context) {
	id := parseInt64Path(c, "id")
	if err := service.AdminRevokeInviteCode(id); err != nil {
		utils.Fail(c, utils.CodeBadRequest, err.Error())
		return
	}
	utils.Success(c, gin.H{"msg": "ok"})
}

// ExportInviteCodes 导出邀请码
// GET /api/v1/admin/invite-codes/export
func (h *AdminInviteHandler) ExportInviteCodes(c *gin.Context) {
	codeType := c.Query("type")
	status := c.Query("status")
	list, err := service.AdminExportInviteCodes(codeType, status)
	if err != nil {
		utils.Fail(c, utils.CodeServerError, err.Error())
		return
	}
	utils.Success(c, list)
}

// AdminMarketingHandler 平台营销活动处理器
type AdminMarketingHandler struct{}

// NewAdminMarketingHandler 创建平台营销处理器
func NewAdminMarketingHandler() *AdminMarketingHandler { return &AdminMarketingHandler{} }

// GetCouponTemplates 优惠券模板列表
// GET /api/v1/admin/coupon-templates
func (h *AdminMarketingHandler) GetCouponTemplates(c *gin.Context) {
	page, pageSize := getPage(c)
	list, total, err := service.AdminGetCouponTemplates(page, pageSize)
	if err != nil {
		utils.Fail(c, utils.CodeServerError, err.Error())
		return
	}
	utils.SuccessWithTotal(c, list, total)
}

// CreateCouponTemplate 创建优惠券模板
// POST /api/v1/admin/coupon-templates
func (h *AdminMarketingHandler) CreateCouponTemplate(c *gin.Context) {
	var t model.CouponTemplate
	if err := c.ShouldBindJSON(&t); err != nil {
		utils.Fail(c, utils.CodeBadRequest, "参数错误: "+err.Error())
		return
	}
	if err := service.AdminCreateCouponTemplate(&t); err != nil {
		utils.Fail(c, utils.CodeBadRequest, err.Error())
		return
	}
	utils.Success(c, t)
}

// GetRechargeActivities 充值活动列表(管理端)
// GET /api/v1/admin/recharge-activities
func (h *AdminMarketingHandler) GetRechargeActivities(c *gin.Context) {
	page, pageSize := getPage(c)
	list, total, err := service.AdminGetRechargeActivities(page, pageSize)
	if err != nil {
		utils.Fail(c, utils.CodeServerError, err.Error())
		return
	}
	utils.SuccessWithTotal(c, list, total)
}

// CreateRechargeActivity 创建充值活动
// POST /api/v1/admin/recharge-activities
func (h *AdminMarketingHandler) CreateRechargeActivity(c *gin.Context) {
	var a model.RechargeActivity
	if err := c.ShouldBindJSON(&a); err != nil {
		utils.Fail(c, utils.CodeBadRequest, "参数错误: "+err.Error())
		return
	}
	if err := service.AdminCreateRechargeActivity(&a); err != nil {
		utils.Fail(c, utils.CodeBadRequest, err.Error())
		return
	}
	utils.Success(c, a)
}

// GetLotteryActivities 抽奖活动列表(管理端)
// GET /api/v1/admin/lottery-activities
func (h *AdminMarketingHandler) GetLotteryActivities(c *gin.Context) {
	page, pageSize := getPage(c)
	list, total, err := service.AdminGetLotteryActivities(page, pageSize)
	if err != nil {
		utils.Fail(c, utils.CodeServerError, err.Error())
		return
	}
	utils.SuccessWithTotal(c, list, total)
}

// CreateLotteryActivity 创建抽奖活动
// POST /api/v1/admin/lottery-activities
func (h *AdminMarketingHandler) CreateLotteryActivity(c *gin.Context) {
	var a model.LotteryActivity
	if err := c.ShouldBindJSON(&a); err != nil {
		utils.Fail(c, utils.CodeBadRequest, "参数错误: "+err.Error())
		return
	}
	if err := service.AdminCreateLotteryActivity(&a); err != nil {
		utils.Fail(c, utils.CodeBadRequest, err.Error())
		return
	}
	utils.Success(c, a)
}

// GetGroupBuyActivities 拼团活动列表(管理端)
// GET /api/v1/admin/group-buy-activities
func (h *AdminMarketingHandler) GetGroupBuyActivities(c *gin.Context) {
	page, pageSize := getPage(c)
	list, total, err := service.AdminGetGroupBuyActivities(page, pageSize)
	if err != nil {
		utils.Fail(c, utils.CodeServerError, err.Error())
		return
	}
	utils.SuccessWithTotal(c, list, total)
}

// CreateGroupBuyActivity 创建拼团活动
// POST /api/v1/admin/group-buy-activities
func (h *AdminMarketingHandler) CreateGroupBuyActivity(c *gin.Context) {
	var a model.GroupBuyActivity
	if err := c.ShouldBindJSON(&a); err != nil {
		utils.Fail(c, utils.CodeBadRequest, "参数错误: "+err.Error())
		return
	}
	if err := service.AdminCreateGroupBuyActivity(&a); err != nil {
		utils.Fail(c, utils.CodeBadRequest, err.Error())
		return
	}
	utils.Success(c, a)
}

// GetInviteRewardConfig 邀请奖励配置
// GET /api/v1/admin/invite-reward-config
func (h *AdminMarketingHandler) GetInviteRewardConfig(c *gin.Context) {
	res, err := service.AdminGetInviteRewardConfig()
	if err != nil {
		utils.Fail(c, utils.CodeServerError, err.Error())
		return
	}
	utils.Success(c, res)
}

// updateInviteRewardConfigRequest 更新邀请奖励配置请求
type updateInviteRewardConfigRequest struct {
	Points  int64 `json:"points"`
	Balance int64 `json:"balance"`
}

// UpdateInviteRewardConfig 更新邀请奖励配置
// PUT /api/v1/admin/invite-reward-config
func (h *AdminMarketingHandler) UpdateInviteRewardConfig(c *gin.Context) {
	var req updateInviteRewardConfigRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Fail(c, utils.CodeBadRequest, "参数错误: "+err.Error())
		return
	}
	if err := service.AdminUpdateInviteRewardConfig(req.Points, req.Balance); err != nil {
		utils.Fail(c, utils.CodeBadRequest, err.Error())
		return
	}
	utils.Success(c, gin.H{"msg": "ok"})
}
