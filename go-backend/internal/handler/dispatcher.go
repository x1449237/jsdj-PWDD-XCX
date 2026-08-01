package handler

import (
	"github.com/gin-gonic/gin"
	"github.com/jisan/e-sports-platform/internal/service"
	"github.com/jisan/e-sports-platform/internal/utils"
)

// DispatcherHandler 派单员处理器
type DispatcherHandler struct{}

// NewDispatcherHandler 创建派单员处理器
func NewDispatcherHandler() *DispatcherHandler { return &DispatcherHandler{} }

// GetDispatchOrders 派单员待派订单列表
// GET /api/v1/dispatcher/orders
func (h *DispatcherHandler) GetDispatchOrders(c *gin.Context) {
	page, pageSize := getPage(c)
	list, total, err := service.GetDispatchOrders(page, pageSize)
	if err != nil {
		utils.Fail(c, utils.CodeServerError, err.Error())
		return
	}
	utils.SuccessWithTotal(c, list, total)
}

// dispatchOrderRequest 派单请求
type dispatchOrderRequest struct {
	PlayerID int64 `json:"player_id" binding:"required"`
}

// DispatchOrder 派单员指定打手接单
// POST /api/v1/dispatcher/orders/:id/dispatch
func (h *DispatcherHandler) DispatchOrder(c *gin.Context) {
	orderID := parseInt64Path(c, "id")
	var req dispatchOrderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Fail(c, utils.CodeBadRequest, "参数错误: "+err.Error())
		return
	}
	dispatcherID := getCurrentUserID(c)
	if err := service.DispatchOrder(orderID, dispatcherID, req.PlayerID); err != nil {
		utils.Fail(c, utils.CodeBadRequest, err.Error())
		return
	}
	utils.Success(c, gin.H{"msg": "ok"})
}

// GetDispatchHistory 派单历史
// GET /api/v1/dispatcher/history
func (h *DispatcherHandler) GetDispatchHistory(c *gin.Context) {
	dispatcherID := getCurrentUserID(c)
	page, pageSize := getPage(c)
	list, total, err := service.GetDispatchHistory(dispatcherID, page, pageSize)
	if err != nil {
		utils.Fail(c, utils.CodeServerError, err.Error())
		return
	}
	utils.SuccessWithTotal(c, list, total)
}
