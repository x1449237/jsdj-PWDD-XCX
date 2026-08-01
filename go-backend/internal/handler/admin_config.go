package handler

import (
	"github.com/gin-gonic/gin"
	"github.com/jisan/e-sports-platform/internal/service"
	"github.com/jisan/e-sports-platform/internal/utils"
)

// AdminConfigHandler 平台配置/日志/系统处理器
type AdminConfigHandler struct{}

// NewAdminConfigHandler 创建平台配置处理器
func NewAdminConfigHandler() *AdminConfigHandler { return &AdminConfigHandler{} }

// GetSystemConfig 获取系统配置(全部)
// GET /api/v1/admin/system-configs
func (h *AdminConfigHandler) GetSystemConfig(c *gin.Context) {
	list, err := service.AdminGetSystemConfig()
	if err != nil {
		utils.Fail(c, utils.CodeServerError, err.Error())
		return
	}
	utils.Success(c, list)
}

// updateSystemConfigRequest 更新系统配置请求
type updateSystemConfigRequest struct {
	Key         string `json:"key" binding:"required"`
	Value       string `json:"value" binding:"required"`
	Description string `json:"description"`
}

// UpdateSystemConfig 更新系统配置项
// PUT /api/v1/admin/system-configs
func (h *AdminConfigHandler) UpdateSystemConfig(c *gin.Context) {
	var req updateSystemConfigRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Fail(c, utils.CodeBadRequest, "参数错误: "+err.Error())
		return
	}
	if err := service.AdminUpdateSystemConfig(req.Key, req.Value, req.Description); err != nil {
		utils.Fail(c, utils.CodeBadRequest, err.Error())
		return
	}
	utils.Success(c, gin.H{"msg": "ok"})
}

// GetOperationLogs 操作日志列表
// GET /api/v1/admin/operation-logs
func (h *AdminConfigHandler) GetOperationLogs(c *gin.Context) {
	page, pageSize := getPage(c)
	action := c.Query("action")
	list, total, err := service.AdminGetOperationLogs(page, pageSize, action)
	if err != nil {
		utils.Fail(c, utils.CodeServerError, err.Error())
		return
	}
	utils.SuccessWithTotal(c, list, total)
}

// GetApiMonitor API 监控数据
// GET /api/v1/admin/api-monitor
func (h *AdminConfigHandler) GetApiMonitor(c *gin.Context) {
	res, err := service.AdminGetApiMonitor()
	if err != nil {
		utils.Fail(c, utils.CodeServerError, err.Error())
		return
	}
	utils.Success(c, res)
}

// createBackupRequest 创建备份请求
type createBackupRequest struct {
	Name string `json:"name" binding:"required"`
}

// CreateBackup 创建备份
// POST /api/v1/admin/backups
func (h *AdminConfigHandler) CreateBackup(c *gin.Context) {
	var req createBackupRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Fail(c, utils.CodeBadRequest, "参数错误: "+err.Error())
		return
	}
	adminID := getCurrentUserID(c)
	res, err := service.AdminCreateBackup(adminID, req.Name)
	if err != nil {
		utils.Fail(c, utils.CodeBadRequest, err.Error())
		return
	}
	utils.Success(c, res)
}

// restoreBackupRequest 恢复备份请求
type restoreBackupRequest struct {
	Name string `json:"name" binding:"required"`
}

// RestoreBackup 恢复备份
// POST /api/v1/admin/backups/restore
func (h *AdminConfigHandler) RestoreBackup(c *gin.Context) {
	var req restoreBackupRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Fail(c, utils.CodeBadRequest, "参数错误: "+err.Error())
		return
	}
	adminID := getCurrentUserID(c)
	if err := service.AdminRestoreBackup(adminID, req.Name); err != nil {
		utils.Fail(c, utils.CodeBadRequest, err.Error())
		return
	}
	utils.Success(c, gin.H{"msg": "ok"})
}

// GetBackupList 备份列表
// GET /api/v1/admin/backups
func (h *AdminConfigHandler) GetBackupList(c *gin.Context) {
	list, err := service.AdminGetBackupList()
	if err != nil {
		utils.Fail(c, utils.CodeServerError, err.Error())
		return
	}
	utils.Success(c, list)
}

// GetGrayRelease 灰度发布配置
// GET /api/v1/admin/gray-release
func (h *AdminConfigHandler) GetGrayRelease(c *gin.Context) {
	res, err := service.AdminGetGrayRelease()
	if err != nil {
		utils.Fail(c, utils.CodeServerError, err.Error())
		return
	}
	utils.Success(c, res)
}

// updateGrayReleaseRequest 更新灰度发布请求
type updateGrayReleaseRequest struct {
	Percent   int     `json:"percent"`
	Whitelist []int64 `json:"whitelist"`
}

// UpdateGrayRelease 更新灰度发布配置
// PUT /api/v1/admin/gray-release
func (h *AdminConfigHandler) UpdateGrayRelease(c *gin.Context) {
	var req updateGrayReleaseRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Fail(c, utils.CodeBadRequest, "参数错误: "+err.Error())
		return
	}
	if err := service.AdminUpdateGrayRelease(req.Percent, req.Whitelist); err != nil {
		utils.Fail(c, utils.CodeBadRequest, err.Error())
		return
	}
	utils.Success(c, gin.H{"msg": "ok"})
}

// GetAgreements 合规协议列表
// GET /api/v1/admin/agreements
func (h *AdminConfigHandler) GetAgreements(c *gin.Context) {
	list, err := service.AdminGetAgreements()
	if err != nil {
		utils.Fail(c, utils.CodeServerError, err.Error())
		return
	}
	utils.Success(c, list)
}

// agreementRequest 合规协议请求
type agreementRequest struct {
	Name    string `json:"name" binding:"required"`
	Content string `json:"content" binding:"required"`
}

// CreateAgreement 创建合规协议
// POST /api/v1/admin/agreements
func (h *AdminConfigHandler) CreateAgreement(c *gin.Context) {
	var req agreementRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Fail(c, utils.CodeBadRequest, "参数错误: "+err.Error())
		return
	}
	if err := service.AdminCreateAgreement(req.Name, req.Content); err != nil {
		utils.Fail(c, utils.CodeBadRequest, err.Error())
		return
	}
	utils.Success(c, gin.H{"msg": "ok"})
}

// GetAntiBoostingRules 防代练规则
// GET /api/v1/admin/anti-boosting-rules
func (h *AdminConfigHandler) GetAntiBoostingRules(c *gin.Context) {
	list, err := service.AdminGetAntiBoostingRules()
	if err != nil {
		utils.Fail(c, utils.CodeServerError, err.Error())
		return
	}
	utils.Success(c, list)
}

// AddAntiBoostingRule 新增防代练规则
// POST /api/v1/admin/anti-boosting-rules
func (h *AdminConfigHandler) AddAntiBoostingRule(c *gin.Context) {
	var req agreementRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Fail(c, utils.CodeBadRequest, "参数错误: "+err.Error())
		return
	}
	if err := service.AdminAddAntiBoostingRule(req.Name, req.Content); err != nil {
		utils.Fail(c, utils.CodeBadRequest, err.Error())
		return
	}
	utils.Success(c, gin.H{"msg": "ok"})
}

// GetNotifications 通知列表
// GET /api/v1/admin/notifications
func (h *AdminConfigHandler) GetNotifications(c *gin.Context) {
	page, pageSize := getPage(c)
	list, total, err := service.AdminGetNotifications(page, pageSize)
	if err != nil {
		utils.Fail(c, utils.CodeServerError, err.Error())
		return
	}
	utils.SuccessWithTotal(c, list, total)
}

// sendNotificationRequest 发送通知请求
type sendNotificationRequest struct {
	UserID   int64  `json:"user_id" binding:"required"`
	Type     string `json:"type" binding:"required"`
	Title    string `json:"title" binding:"required"`
	Content  string `json:"content" binding:"required"`
	Category string `json:"category"`
}

// SendNotification 发送通知
// POST /api/v1/admin/notifications
func (h *AdminConfigHandler) SendNotification(c *gin.Context) {
	var req sendNotificationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Fail(c, utils.CodeBadRequest, "参数错误: "+err.Error())
		return
	}
	if err := service.AdminSendNotification(req.UserID, req.Type, req.Title, req.Content, req.Category); err != nil {
		utils.Fail(c, utils.CodeBadRequest, err.Error())
		return
	}
	utils.Success(c, gin.H{"msg": "ok"})
}

// GetSubscribeTemplates 订阅消息模板列表
// GET /api/v1/admin/subscribe-templates
func (h *AdminConfigHandler) GetSubscribeTemplates(c *gin.Context) {
	list, err := service.AdminGetSubscribeTemplates()
	if err != nil {
		utils.Fail(c, utils.CodeServerError, err.Error())
		return
	}
	utils.Success(c, list)
}

// addSubscribeTemplateRequest 新增订阅消息模板请求
type addSubscribeTemplateRequest struct {
	Name       string `json:"name" binding:"required"`
	TemplateID string `json:"template_id" binding:"required"`
}

// AddSubscribeTemplate 新增订阅消息模板
// POST /api/v1/admin/subscribe-templates
func (h *AdminConfigHandler) AddSubscribeTemplate(c *gin.Context) {
	var req addSubscribeTemplateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Fail(c, utils.CodeBadRequest, "参数错误: "+err.Error())
		return
	}
	if err := service.AdminAddSubscribeTemplate(req.Name, req.TemplateID); err != nil {
		utils.Fail(c, utils.CodeBadRequest, err.Error())
		return
	}
	utils.Success(c, gin.H{"msg": "ok"})
}

// GetShopDecorations 店铺装饰列表
// GET /api/v1/admin/shop-decorations
func (h *AdminConfigHandler) GetShopDecorations(c *gin.Context) {
	list, err := service.AdminGetShopDecorations()
	if err != nil {
		utils.Fail(c, utils.CodeServerError, err.Error())
		return
	}
	utils.Success(c, list)
}

// updateShopDecorationRequest 更新店铺装饰请求
type updateShopDecorationRequest struct {
	Content string `json:"content" binding:"required"`
}

// UpdateShopDecoration 更新店铺装饰
// PUT /api/v1/admin/shop-decorations/:shop_id
func (h *AdminConfigHandler) UpdateShopDecoration(c *gin.Context) {
	shopID := parseInt64Path(c, "shop_id")
	var req updateShopDecorationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Fail(c, utils.CodeBadRequest, "参数错误: "+err.Error())
		return
	}
	if err := service.AdminUpdateShopDecoration(shopID, req.Content); err != nil {
		utils.Fail(c, utils.CodeBadRequest, err.Error())
		return
	}
	utils.Success(c, gin.H{"msg": "ok"})
}

// GetTimeoutRules 超时规则列表
// GET /api/v1/admin/timeout-rules
func (h *AdminConfigHandler) GetTimeoutRules(c *gin.Context) {
	list, err := service.AdminGetTimeoutRules()
	if err != nil {
		utils.Fail(c, utils.CodeServerError, err.Error())
		return
	}
	utils.Success(c, list)
}

// AddTimeoutRule 新增超时规则
// POST /api/v1/admin/timeout-rules
func (h *AdminConfigHandler) AddTimeoutRule(c *gin.Context) {
	var req agreementRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Fail(c, utils.CodeBadRequest, "参数错误: "+err.Error())
		return
	}
	if err := service.AdminAddTimeoutRule(req.Name, req.Content); err != nil {
		utils.Fail(c, utils.CodeBadRequest, err.Error())
		return
	}
	utils.Success(c, gin.H{"msg": "ok"})
}

// updateTimeoutRuleRequest 更新超时规则请求
type updateTimeoutRuleRequest struct {
	Content string `json:"content" binding:"required"`
}

// UpdateTimeoutRule 更新超时规则
// PUT /api/v1/admin/timeout-rules/:id
func (h *AdminConfigHandler) UpdateTimeoutRule(c *gin.Context) {
	id := parseInt64Path(c, "id")
	var req updateTimeoutRuleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Fail(c, utils.CodeBadRequest, "参数错误: "+err.Error())
		return
	}
	if err := service.AdminUpdateTimeoutRule(id, req.Content); err != nil {
		utils.Fail(c, utils.CodeBadRequest, err.Error())
		return
	}
	utils.Success(c, gin.H{"msg": "ok"})
}

// GetDocuments 文档列表
// GET /api/v1/admin/documents
func (h *AdminConfigHandler) GetDocuments(c *gin.Context) {
	list, err := service.AdminGetDocuments()
	if err != nil {
		utils.Fail(c, utils.CodeServerError, err.Error())
		return
	}
	utils.Success(c, list)
}

// UploadDocument 上传文档
// POST /api/v1/admin/documents
func (h *AdminConfigHandler) UploadDocument(c *gin.Context) {
	var req agreementRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Fail(c, utils.CodeBadRequest, "参数错误: "+err.Error())
		return
	}
	if err := service.AdminUploadDocument(req.Name, req.Content); err != nil {
		utils.Fail(c, utils.CodeBadRequest, err.Error())
		return
	}
	utils.Success(c, gin.H{"msg": "ok"})
}

// GetDocumentVersions 文档版本列表
// GET /api/v1/admin/documents/:id/versions
func (h *AdminConfigHandler) GetDocumentVersions(c *gin.Context) {
	id := parseInt64Path(c, "id")
	list, err := service.AdminGetDocumentVersions(id)
	if err != nil {
		utils.Fail(c, utils.CodeServerError, err.Error())
		return
	}
	utils.Success(c, list)
}

// ReplaceDocument 替换文档
// PUT /api/v1/admin/documents/:id
func (h *AdminConfigHandler) ReplaceDocument(c *gin.Context) {
	id := parseInt64Path(c, "id")
	var req updateTimeoutRuleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Fail(c, utils.CodeBadRequest, "参数错误: "+err.Error())
		return
	}
	if err := service.AdminReplaceDocument(id, req.Content); err != nil {
		utils.Fail(c, utils.CodeBadRequest, err.Error())
		return
	}
	utils.Success(c, gin.H{"msg": "ok"})
}

// DeleteDocument 删除文档
// DELETE /api/v1/admin/documents/:id
func (h *AdminConfigHandler) DeleteDocument(c *gin.Context) {
	id := parseInt64Path(c, "id")
	if err := service.AdminDeleteDocument(id); err != nil {
		utils.Fail(c, utils.CodeBadRequest, err.Error())
		return
	}
	utils.Success(c, gin.H{"msg": "ok"})
}
