package handler

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jisan/e-sports-platform/internal/service"
	"github.com/jisan/e-sports-platform/internal/utils"
)

// OrderHandler 订单处理器(用户侧)
type OrderHandler struct{}

// NewOrderHandler 创建订单处理器
func NewOrderHandler() *OrderHandler { return &OrderHandler{} }

// createOrderRequest 创建订单请求
type createOrderRequest struct {
	Type            int8   `json:"type"`              // 订单类型
	ClubID          int64  `json:"club_id"`            // 俱乐部ID
	ServiceID       int64  `json:"service_id"`         // 服务项目ID
	Amount          int64  `json:"amount" binding:"required"` // 金额(分)
	TeamCount       int    `json:"team_count"`         // 车队人数
	AppointmentTime string `json:"appointment_time"`   // 预约时间(ISO8601)
}

// CreateOrder 用户创建订单
// POST /api/v1/orders
func (h *OrderHandler) CreateOrder(c *gin.Context) {
	var req createOrderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Fail(c, utils.CodeBadRequest, "参数错误: "+err.Error())
		return
	}
	userID := getCurrentUserID(c)
	in := &service.CreateOrderInput{
		Type:      req.Type,
		ClubID:    req.ClubID,
		ServiceID: req.ServiceID,
		Amount:    req.Amount,
		TeamCount: req.TeamCount,
	}
	if req.AppointmentTime != "" {
		t, err := time.Parse(time.RFC3339, req.AppointmentTime)
		if err == nil {
			in.Type = 2 // 预约单
			in.AppointmentTime = &t
		}
	}
	o, err := service.CreateOrder(userID, in)
	if err != nil {
		utils.Fail(c, utils.CodeBadRequest, err.Error())
		return
	}
	utils.Success(c, o)
}

// GetOrderList 用户订单列表
// GET /api/v1/orders
func (h *OrderHandler) GetOrderList(c *gin.Context) {
	userID := getCurrentUserID(c)
	page, pageSize := getPage(c)
	status := parseInt8Query(c, "status", -1)
	list, total, err := service.GetOrderList(userID, page, pageSize, status)
	if err != nil {
		utils.Fail(c, utils.CodeServerError, err.Error())
		return
	}
	utils.SuccessWithTotal(c, list, total)
}

// GetOrderDetail 订单详情
// GET /api/v1/orders/:id
func (h *OrderHandler) GetOrderDetail(c *gin.Context) {
	orderID := parseInt64Path(c, "id")
	if orderID == 0 {
		utils.Fail(c, utils.CodeBadRequest, "订单ID不能为空")
		return
	}
	userID := getCurrentUserID(c)
	o, err := service.GetOrderDetail(orderID, userID, isPlatformAdmin(c))
	if err != nil {
		utils.Fail(c, utils.CodeNotFound, err.Error())
		return
	}
	utils.Success(c, o)
}

// cancelOrderRequest 取消订单请求
type cancelOrderRequest struct {
	Reason string `json:"reason"`
}

// CancelOrder 用户取消订单
// POST /api/v1/orders/:id/cancel
func (h *OrderHandler) CancelOrder(c *gin.Context) {
	orderID := parseInt64Path(c, "id")
	userID := getCurrentUserID(c)
	var req cancelOrderRequest
	_ = c.ShouldBindJSON(&req)
	if err := service.CancelOrder(orderID, userID, req.Reason); err != nil {
		utils.Fail(c, utils.CodeBadRequest, err.Error())
		return
	}
	utils.Success(c, gin.H{"msg": "ok"})
}

// submitAppealRequest 提交申诉请求
type submitAppealRequest struct {
	Type         string   `json:"type" binding:"required"`
	Description  string   `json:"description" binding:"required"`
	EvidenceURLs []string `json:"evidence_urls"`
}

// SubmitAppeal 提交订单申诉
// POST /api/v1/orders/:id/appeal
func (h *OrderHandler) SubmitAppeal(c *gin.Context) {
	orderID := parseInt64Path(c, "id")
	var req submitAppealRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Fail(c, utils.CodeBadRequest, "参数错误: "+err.Error())
		return
	}
	userID := getCurrentUserID(c)
	a, err := service.SubmitAppeal(orderID, userID, req.Type, req.Description, req.EvidenceURLs)
	if err != nil {
		utils.Fail(c, utils.CodeBadRequest, err.Error())
		return
	}
	utils.Success(c, a)
}

// submitEvaluationRequest 提交评价请求
type submitEvaluationRequest struct {
	Score   int    `json:"score" binding:"required"`
	Content string `json:"content"`
}

// SubmitEvaluation 提交订单评价
// POST /api/v1/orders/:id/evaluation
func (h *OrderHandler) SubmitEvaluation(c *gin.Context) {
	orderID := parseInt64Path(c, "id")
	var req submitEvaluationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Fail(c, utils.CodeBadRequest, "参数错误: "+err.Error())
		return
	}
	userID := getCurrentUserID(c)
	e, err := service.SubmitEvaluation(orderID, userID, req.Score, req.Content)
	if err != nil {
		utils.Fail(c, utils.CodeBadRequest, err.Error())
		return
	}
	utils.Success(c, e)
}

// sendRewardRequest 打赏请求
type sendRewardRequest struct {
	Amount  int64  `json:"amount" binding:"required"`
	GiftType string `json:"gift_type"`
}

// SendReward 用户给打手打赏
// POST /api/v1/orders/:id/reward
func (h *OrderHandler) SendReward(c *gin.Context) {
	orderID := parseInt64Path(c, "id")
	var req sendRewardRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Fail(c, utils.CodeBadRequest, "参数错误: "+err.Error())
		return
	}
	userID := getCurrentUserID(c)
	r, err := service.SendReward(orderID, userID, req.Amount, req.GiftType)
	if err != nil {
		utils.Fail(c, utils.CodeBadRequest, err.Error())
		return
	}
	utils.Success(c, r)
}

// uploadEvidenceRequest 上传凭证请求
type uploadEvidenceRequest struct {
	FileType string `json:"file_type" binding:"required"`
	FileURL  string `json:"file_url" binding:"required"`
}

// UploadEvidence 上传订单履约凭证
// POST /api/v1/orders/:id/evidence
func (h *OrderHandler) UploadEvidence(c *gin.Context) {
	orderID := parseInt64Path(c, "id")
	var req uploadEvidenceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Fail(c, utils.CodeBadRequest, "参数错误: "+err.Error())
		return
	}
	userID := getCurrentUserID(c)
	if err := service.UploadEvidence(orderID, userID, req.FileType, req.FileURL); err != nil {
		utils.Fail(c, utils.CodeBadRequest, err.Error())
		return
	}
	utils.Success(c, gin.H{"msg": "ok"})
}

// GetAppealList 用户申诉列表
// GET /api/v1/appeals
func (h *OrderHandler) GetAppealList(c *gin.Context) {
	userID := getCurrentUserID(c)
	page, pageSize := getPage(c)
	list, total, err := service.GetAppealList(userID, page, pageSize)
	if err != nil {
		utils.Fail(c, utils.CodeServerError, err.Error())
		return
	}
	utils.SuccessWithTotal(c, list, total)
}

// GetAppealDetail 申诉详情
// GET /api/v1/appeals/:id
func (h *OrderHandler) GetAppealDetail(c *gin.Context) {
	appealID := parseInt64Path(c, "id")
	userID := getCurrentUserID(c)
	a, err := service.GetAppealDetail(appealID, userID)
	if err != nil {
		utils.Fail(c, utils.CodeNotFound, err.Error())
		return
	}
	utils.Success(c, a)
}

// uploadAppealMaterialsRequest 上传申诉补充材料请求
type uploadAppealMaterialsRequest struct {
	URLs []string `json:"urls" binding:"required"`
}

// UploadAppealMaterials 上传申诉补充材料
// POST /api/v1/appeals/:id/materials
func (h *OrderHandler) UploadAppealMaterials(c *gin.Context) {
	appealID := parseInt64Path(c, "id")
	var req uploadAppealMaterialsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Fail(c, utils.CodeBadRequest, "参数错误: "+err.Error())
		return
	}
	userID := getCurrentUserID(c)
	if err := service.UploadAppealMaterials(appealID, userID, req.URLs); err != nil {
		utils.Fail(c, utils.CodeBadRequest, err.Error())
		return
	}
	utils.Success(c, gin.H{"msg": "ok"})
}
