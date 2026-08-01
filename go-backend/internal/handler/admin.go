package handler

import (
	"github.com/gin-gonic/gin"
	"github.com/jisan/e-sports-platform/internal/model"
	"github.com/jisan/e-sports-platform/internal/service"
	"github.com/jisan/e-sports-platform/internal/utils"
)

// AdminHandler 平台管理员处理器(仪表盘/用户/管理员/订单)
type AdminHandler struct{}

// NewAdminHandler 创建平台管理员处理器
func NewAdminHandler() *AdminHandler { return &AdminHandler{} }

// Dashboard 平台仪表盘
// GET /api/v1/admin/dashboard
func (h *AdminHandler) Dashboard(c *gin.Context) {
	res, err := service.AdminDashboard()
	if err != nil {
		utils.Fail(c, utils.CodeServerError, err.Error())
		return
	}
	utils.Success(c, res)
}

// BigScreenData 大屏数据
// GET /api/v1/admin/big-screen
func (h *AdminHandler) BigScreenData(c *gin.Context) {
	res, err := service.AdminBigScreenData()
	if err != nil {
		utils.Fail(c, utils.CodeServerError, err.Error())
		return
	}
	utils.Success(c, res)
}

// GetUsers 平台用户列表
// GET /api/v1/admin/users
func (h *AdminHandler) GetUsers(c *gin.Context) {
	page, pageSize := getPage(c)
	status := parseInt8Query(c, "status", -1)
	role := parseInt8Query(c, "role", -1)
	keyword := c.Query("keyword")
	list, total, err := service.AdminGetUsers(page, pageSize, status, role, keyword)
	if err != nil {
		utils.Fail(c, utils.CodeServerError, err.Error())
		return
	}
	utils.SuccessWithTotal(c, list, total)
}

// GetUserDetail 用户详情
// GET /api/v1/admin/users/:id
func (h *AdminHandler) GetUserDetail(c *gin.Context) {
	userID := parseInt64Path(c, "id")
	res, err := service.AdminGetUserDetail(userID)
	if err != nil {
		utils.Fail(c, utils.CodeNotFound, err.Error())
		return
	}
	utils.Success(c, res)
}

// GetNormalUsers 正常用户列表
// GET /api/v1/admin/users/normal
func (h *AdminHandler) GetNormalUsers(c *gin.Context) {
	page, pageSize := getPage(c)
	keyword := c.Query("keyword")
	list, total, err := service.AdminGetNormalUsers(page, pageSize, keyword)
	if err != nil {
		utils.Fail(c, utils.CodeServerError, err.Error())
		return
	}
	utils.SuccessWithTotal(c, list, total)
}

// GetFailedVerificationUsers 实名验证失败用户
// GET /api/v1/admin/users/failed-verification
func (h *AdminHandler) GetFailedVerificationUsers(c *gin.Context) {
	page, pageSize := getPage(c)
	list, total, err := service.AdminGetFailedVerificationUsers(page, pageSize)
	if err != nil {
		utils.Fail(c, utils.CodeServerError, err.Error())
		return
	}
	utils.SuccessWithTotal(c, list, total)
}

// banUserRequest 封禁用户请求
type banUserRequest struct {
	Reason string `json:"reason" binding:"required"`
}

// BanUser 封禁用户
// POST /api/v1/admin/users/:id/ban
func (h *AdminHandler) BanUser(c *gin.Context) {
	userID := parseInt64Path(c, "id")
	var req banUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Fail(c, utils.CodeBadRequest, "参数错误: "+err.Error())
		return
	}
	adminID := getCurrentUserID(c)
	if err := service.AdminBanUser(userID, adminID, req.Reason); err != nil {
		utils.Fail(c, utils.CodeBadRequest, err.Error())
		return
	}
	utils.Success(c, gin.H{"msg": "ok"})
}

// UnbanUser 解封用户
// POST /api/v1/admin/users/:id/unban
func (h *AdminHandler) UnbanUser(c *gin.Context) {
	userID := parseInt64Path(c, "id")
	adminID := getCurrentUserID(c)
	if err := service.AdminUnbanUser(userID, adminID); err != nil {
		utils.Fail(c, utils.CodeBadRequest, err.Error())
		return
	}
	utils.Success(c, gin.H{"msg": "ok"})
}

// ExportUsers 导出用户
// GET /api/v1/admin/users/export
func (h *AdminHandler) ExportUsers(c *gin.Context) {
	status := parseInt8Query(c, "status", -1)
	role := parseInt8Query(c, "role", -1)
	keyword := c.Query("keyword")
	list, err := service.AdminExportUsers(status, role, keyword)
	if err != nil {
		utils.Fail(c, utils.CodeServerError, err.Error())
		return
	}
	utils.Success(c, list)
}

// GetManagers 管理员列表
// GET /api/v1/admin/managers
func (h *AdminHandler) GetManagers(c *gin.Context) {
	page, pageSize := getPage(c)
	keyword := c.Query("keyword")
	list, total, err := service.AdminGetManagers(page, pageSize, keyword)
	if err != nil {
		utils.Fail(c, utils.CodeServerError, err.Error())
		return
	}
	utils.SuccessWithTotal(c, list, total)
}

// addManagerRequest 新增管理员请求
type addManagerRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
	Nickname string `json:"nickname"`
	Email    string `json:"email"`
	Role     int8   `json:"role"`
}

// AddManager 新增管理员
// POST /api/v1/admin/managers
func (h *AdminHandler) AddManager(c *gin.Context) {
	var req addManagerRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Fail(c, utils.CodeBadRequest, "参数错误: "+err.Error())
		return
	}
	a := &model.Admin{
		Username: req.Username,
		Password: req.Password,
		Nickname: req.Nickname,
		Email:    req.Email,
		Role:     req.Role,
	}
	if err := service.AdminAddManager(a); err != nil {
		utils.Fail(c, utils.CodeBadRequest, err.Error())
		return
	}
	utils.Success(c, a)
}

// updateManagerRequest 更新管理员请求
type updateManagerRequest struct {
	Nickname string `json:"nickname"`
	Email    string `json:"email"`
	Role     int8   `json:"role"`
	Status   int8   `json:"status"`
}

// UpdateManager 更新管理员
// PUT /api/v1/admin/managers/:id
func (h *AdminHandler) UpdateManager(c *gin.Context) {
	id := parseInt64Path(c, "id")
	var req updateManagerRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Fail(c, utils.CodeBadRequest, "参数错误: "+err.Error())
		return
	}
	fields := map[string]interface{}{
		"nickname": req.Nickname,
		"email":    req.Email,
		"role":     req.Role,
		"status":   req.Status,
	}
	if err := service.AdminUpdateManager(id, fields); err != nil {
		utils.Fail(c, utils.CodeBadRequest, err.Error())
		return
	}
	utils.Success(c, gin.H{"msg": "ok"})
}

// DeleteManager 删除管理员
// DELETE /api/v1/admin/managers/:id
func (h *AdminHandler) DeleteManager(c *gin.Context) {
	id := parseInt64Path(c, "id")
	operatorID := getCurrentUserID(c)
	if err := service.AdminDeleteManager(id, operatorID); err != nil {
		utils.Fail(c, utils.CodeBadRequest, err.Error())
		return
	}
	utils.Success(c, gin.H{"msg": "ok"})
}

// resetManagerPasswordRequest 重置管理员密码请求
type resetManagerPasswordRequest struct {
	Password string `json:"password" binding:"required"`
}

// ResetManagerPassword 重置管理员密码
// POST /api/v1/admin/managers/:id/reset-password
func (h *AdminHandler) ResetManagerPassword(c *gin.Context) {
	id := parseInt64Path(c, "id")
	var req resetManagerPasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Fail(c, utils.CodeBadRequest, "参数错误: "+err.Error())
		return
	}
	if err := service.AdminResetManagerPassword(id, req.Password); err != nil {
		utils.Fail(c, utils.CodeBadRequest, err.Error())
		return
	}
	utils.Success(c, gin.H{"msg": "ok"})
}

// GetOrders 平台订单列表
// GET /api/v1/admin/orders
func (h *AdminHandler) GetOrders(c *gin.Context) {
	page, pageSize := getPage(c)
	status := parseInt8Query(c, "status", -1)
	keyword := c.Query("keyword")
	list, total, err := service.AdminGetOrders(page, pageSize, status, keyword)
	if err != nil {
		utils.Fail(c, utils.CodeServerError, err.Error())
		return
	}
	utils.SuccessWithTotal(c, list, total)
}

// forceUpdateOrderStatusRequest 强制更新订单状态请求
type forceUpdateOrderStatusRequest struct {
	Status int8   `json:"status" binding:"required"`
	Reason string `json:"reason"`
}

// ForceUpdateOrderStatus 平台强制更新订单状态
// POST /api/v1/admin/orders/:id/status
func (h *AdminHandler) ForceUpdateOrderStatus(c *gin.Context) {
	orderID := parseInt64Path(c, "id")
	var req forceUpdateOrderStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Fail(c, utils.CodeBadRequest, "参数错误: "+err.Error())
		return
	}
	adminID := getCurrentUserID(c)
	if err := service.AdminForceUpdateOrderStatus(orderID, adminID, req.Status, req.Reason); err != nil {
		utils.Fail(c, utils.CodeBadRequest, err.Error())
		return
	}
	utils.Success(c, gin.H{"msg": "ok"})
}

// GetFailedOrders 大额验证失败订单
// GET /api/v1/admin/orders/failed
func (h *AdminHandler) GetFailedOrders(c *gin.Context) {
	page, pageSize := getPage(c)
	list, total, err := service.AdminGetFailedOrders(page, pageSize)
	if err != nil {
		utils.Fail(c, utils.CodeServerError, err.Error())
		return
	}
	utils.SuccessWithTotal(c, list, total)
}

// batchOrderOperationRequest 批量订单操作请求
type batchOrderOperationRequest struct {
	OrderIDs []int64 `json:"order_ids" binding:"required"`
	Action   string  `json:"action" binding:"required"`
}

// BatchOrderOperation 批量订单操作
// POST /api/v1/admin/orders/batch
func (h *AdminHandler) BatchOrderOperation(c *gin.Context) {
	var req batchOrderOperationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Fail(c, utils.CodeBadRequest, "参数错误: "+err.Error())
		return
	}
	adminID := getCurrentUserID(c)
	success, err := service.AdminBatchOrderOperation(adminID, req.OrderIDs, req.Action)
	if err != nil {
		utils.Fail(c, utils.CodeBadRequest, err.Error())
		return
	}
	utils.Success(c, gin.H{"success": success})
}

// RetryFaceVerify 重试活体校验
// POST /api/v1/admin/orders/:id/retry-face
func (h *AdminHandler) RetryFaceVerify(c *gin.Context) {
	orderID := parseInt64Path(c, "id")
	if err := service.AdminRetryFaceVerify(orderID); err != nil {
		utils.Fail(c, utils.CodeBadRequest, err.Error())
		return
	}
	utils.Success(c, gin.H{"msg": "ok"})
}
