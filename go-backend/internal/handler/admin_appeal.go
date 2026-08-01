package handler

import (
	"github.com/gin-gonic/gin"
	"github.com/jisan/e-sports-platform/internal/model"
	"github.com/jisan/e-sports-platform/internal/service"
	"github.com/jisan/e-sports-platform/internal/utils"
)

// AdminAppealHandler 平台申诉/售后处理器(含仲裁能力)
type AdminAppealHandler struct{}

// NewAdminAppealHandler 创建平台申诉处理器
func NewAdminAppealHandler() *AdminAppealHandler { return &AdminAppealHandler{} }

// GetAppeals 平台申诉列表
// GET /api/v1/admin/appeals
func (h *AdminAppealHandler) GetAppeals(c *gin.Context) {
	page, pageSize := getPage(c)
	status := c.Query("status")
	list, total, err := service.AdminGetAppeals(page, pageSize, status)
	if err != nil {
		utils.Fail(c, utils.CodeServerError, err.Error())
		return
	}
	utils.SuccessWithTotal(c, list, total)
}

// GetAppealDetail 平台申诉详情
// GET /api/v1/admin/appeals/:id
func (h *AdminAppealHandler) GetAppealDetail(c *gin.Context) {
	id := parseInt64Path(c, "id")
	res, err := service.AdminGetAppealDetail(id)
	if err != nil {
		utils.Fail(c, utils.CodeNotFound, err.Error())
		return
	}
	utils.Success(c, res)
}

// replyAppealRequest 回复申诉请求
type replyAppealRequest struct {
	Content string `json:"content" binding:"required"`
}

// ReplyAppeal 平台回复申诉
// POST /api/v1/admin/appeals/:id/reply
func (h *AdminAppealHandler) ReplyAppeal(c *gin.Context) {
	id := parseInt64Path(c, "id")
	var req replyAppealRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Fail(c, utils.CodeBadRequest, "参数错误: "+err.Error())
		return
	}
	adminID := getCurrentUserID(c)
	if err := service.AdminReplyAppeal(id, adminID, req.Content); err != nil {
		utils.Fail(c, utils.CodeBadRequest, err.Error())
		return
	}
	utils.Success(c, gin.H{"msg": "ok"})
}

// closeAppealRequest 关闭申诉请求
type closeAppealRequest struct {
	Resolved bool `json:"resolved"`
}

// CloseAppeal 关闭申诉
// POST /api/v1/admin/appeals/:id/close
func (h *AdminAppealHandler) CloseAppeal(c *gin.Context) {
	id := parseInt64Path(c, "id")
	var req closeAppealRequest
	_ = c.ShouldBindJSON(&req)
	if err := service.AdminCloseAppeal(id, req.Resolved); err != nil {
		utils.Fail(c, utils.CodeBadRequest, err.Error())
		return
	}
	utils.Success(c, gin.H{"msg": "ok"})
}

// ---------------- 仲裁 ----------------

// GetArbitrationCases 仲裁案件列表
// GET /api/v1/admin/arbitration/cases
func (h *AdminAppealHandler) GetArbitrationCases(c *gin.Context) {
	page, pageSize := getPage(c)
	status := c.Query("status")
	list, total, err := service.AdminGetArbitrationCases(page, pageSize, status)
	if err != nil {
		utils.Fail(c, utils.CodeServerError, err.Error())
		return
	}
	utils.SuccessWithTotal(c, list, total)
}

// GetArbitrationCaseDetail 仲裁案件详情
// GET /api/v1/admin/arbitration/cases/:id
func (h *AdminAppealHandler) GetArbitrationCaseDetail(c *gin.Context) {
	id := parseInt64Path(c, "id")
	res, err := service.AdminGetArbitrationCaseDetail(id)
	if err != nil {
		utils.Fail(c, utils.CodeNotFound, err.Error())
		return
	}
	utils.Success(c, res)
}

// judgeArbitrationRequest 仲裁判决请求
type judgeArbitrationRequest struct {
	Result string `json:"result" binding:"required"`
}

// JudgeArbitration 仲裁判决
// POST /api/v1/admin/arbitration/cases/:id/judge
func (h *AdminAppealHandler) JudgeArbitration(c *gin.Context) {
	id := parseInt64Path(c, "id")
	var req judgeArbitrationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Fail(c, utils.CodeBadRequest, "参数错误: "+err.Error())
		return
	}
	adminID := getCurrentUserID(c)
	if err := service.AdminJudgeArbitration(id, adminID, req.Result); err != nil {
		utils.Fail(c, utils.CodeBadRequest, err.Error())
		return
	}
	utils.Success(c, gin.H{"msg": "ok"})
}

// GetArbitrationRules 判责规则列表
// GET /api/v1/admin/arbitration/rules
func (h *AdminAppealHandler) GetArbitrationRules(c *gin.Context) {
	page, pageSize := getPage(c)
	list, total, err := service.AdminGetArbitrationRules(page, pageSize)
	if err != nil {
		utils.Fail(c, utils.CodeServerError, err.Error())
		return
	}
	utils.SuccessWithTotal(c, list, total)
}

// addArbitrationRuleRequest 新增判责规则请求
type addArbitrationRuleRequest struct {
	Name           string `json:"name" binding:"required"`
	Responsibility string `json:"responsibility"`
	Penalty        string `json:"penalty"`
}

// AddArbitrationRule 新增判责规则
// POST /api/v1/admin/arbitration/rules
func (h *AdminAppealHandler) AddArbitrationRule(c *gin.Context) {
	var req addArbitrationRuleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Fail(c, utils.CodeBadRequest, "参数错误: "+err.Error())
		return
	}
	rule := &model.ArbitrationRule{
		Name:           req.Name,
		Responsibility: req.Responsibility,
		Penalty:        req.Penalty,
	}
	if err := service.AdminAddArbitrationRule(rule); err != nil {
		utils.Fail(c, utils.CodeBadRequest, err.Error())
		return
	}
	utils.Success(c, rule)
}

// GetEvidenceTemplates 证据模板列表
// GET /api/v1/admin/arbitration/evidence-templates
func (h *AdminAppealHandler) GetEvidenceTemplates(c *gin.Context) {
	page, pageSize := getPage(c)
	list, total, err := service.AdminGetEvidenceTemplates(page, pageSize)
	if err != nil {
		utils.Fail(c, utils.CodeServerError, err.Error())
		return
	}
	utils.SuccessWithTotal(c, list, total)
}

// addEvidenceTemplateRequest 新增证据模板请求
type addEvidenceTemplateRequest struct {
	Name           string `json:"name" binding:"required"`
	Responsibility string `json:"responsibility"`
	Penalty        string `json:"penalty"`
}

// AddEvidenceTemplate 新增证据模板
// POST /api/v1/admin/arbitration/evidence-templates
func (h *AdminAppealHandler) AddEvidenceTemplate(c *gin.Context) {
	var req addEvidenceTemplateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Fail(c, utils.CodeBadRequest, "参数错误: "+err.Error())
		return
	}
	t := &model.ArbitrationRule{
		Name:           req.Name,
		Responsibility: req.Responsibility,
		Penalty:        req.Penalty,
	}
	if err := service.AdminAddEvidenceTemplate(t); err != nil {
		utils.Fail(c, utils.CodeBadRequest, err.Error())
		return
	}
	utils.Success(c, t)
}
