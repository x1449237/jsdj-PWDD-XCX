package handler

import (
	"github.com/gin-gonic/gin"
	"github.com/jisan/e-sports-platform/internal/service"
	"github.com/jisan/e-sports-platform/internal/utils"
)

// PaymentHandler 支付处理器
type PaymentHandler struct{}

// NewPaymentHandler 创建支付处理器
func NewPaymentHandler() *PaymentHandler { return &PaymentHandler{} }

// createPaymentRequest 创建支付请求
// 注意:不再信任客户端 amount,后端以订单实际金额为准
type createPaymentRequest struct {
	OrderID   int64  `json:"order_id" binding:"required"`
	PayMethod string `json:"pay_method" binding:"required"`
}

// CreatePayment 创建支付记录
// POST /api/v1/payments
// 安全修复:传入 userID 校验订单归属,金额以订单为准(忽略客户端 amount)
func (h *PaymentHandler) CreatePayment(c *gin.Context) {
	var req createPaymentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Fail(c, utils.CodeBadRequest, "参数错误: "+err.Error())
		return
	}
	userID := getCurrentUserID(c)
	result, err := service.CreatePayment(userID, req.OrderID, req.PayMethod)
	if err != nil {
		utils.Fail(c, utils.CodeBadRequest, err.Error())
		return
	}
	utils.Success(c, result)
}

// WxPayCallback 微信支付回调
// POST /api/v1/webhook/wxpay
func (h *PaymentHandler) WxPayCallback(c *gin.Context) {
	if err := service.ProxyWxPayCallback(c.Request); err != nil {
		utils.Fail(c, utils.CodeBadRequest, err.Error())
		return
	}
	utils.Success(c, gin.H{"msg": "ok"})
}
