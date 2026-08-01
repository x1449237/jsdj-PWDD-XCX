package handler

import (
	"github.com/gin-gonic/gin"
	"github.com/jisan/e-sports-platform/internal/service"
	"github.com/jisan/e-sports-platform/internal/utils"
)

// MarketingHandler 营销活动处理器
type MarketingHandler struct{}

// NewMarketingHandler 创建营销处理器
func NewMarketingHandler() *MarketingHandler { return &MarketingHandler{} }

// GetMyCoupons 用户优惠券列表
// GET /api/v1/marketing/coupons
func (h *MarketingHandler) GetMyCoupons(c *gin.Context) {
	userID := getCurrentUserID(c)
	status := c.Query("status")
	list, err := service.GetMyCoupons(userID, status)
	if err != nil {
		utils.Fail(c, utils.CodeServerError, err.Error())
		return
	}
	utils.Success(c, list)
}

// GetRechargeActivities 充值活动列表
// GET /api/v1/marketing/recharge-activities
func (h *MarketingHandler) GetRechargeActivities(c *gin.Context) {
	list, err := service.GetRechargeActivities()
	if err != nil {
		utils.Fail(c, utils.CodeServerError, err.Error())
		return
	}
	utils.Success(c, list)
}

// rechargeRequest 充值请求
type rechargeRequest struct {
	Amount     int64 `json:"amount" binding:"required"`
	ActivityID int64 `json:"activity_id"`
}

// Recharge 用户充值
// POST /api/v1/marketing/recharge
func (h *MarketingHandler) Recharge(c *gin.Context) {
	var req rechargeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Fail(c, utils.CodeBadRequest, "参数错误: "+err.Error())
		return
	}
	userID := getCurrentUserID(c)
	res, err := service.Recharge(userID, req.Amount, req.ActivityID)
	if err != nil {
		utils.Fail(c, utils.CodeBadRequest, err.Error())
		return
	}
	utils.Success(c, res)
}

// GetLotteryActivities 抽奖活动列表
// GET /api/v1/marketing/lottery-activities
func (h *MarketingHandler) GetLotteryActivities(c *gin.Context) {
	list, err := service.GetLotteryActivities()
	if err != nil {
		utils.Fail(c, utils.CodeServerError, err.Error())
		return
	}
	utils.Success(c, list)
}

// drawLotteryRequest 抽奖请求
type drawLotteryRequest struct {
	ActivityID int64 `json:"activity_id" binding:"required"`
}

// DrawLottery 抽奖
// POST /api/v1/marketing/lottery/draw
func (h *MarketingHandler) DrawLottery(c *gin.Context) {
	var req drawLotteryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Fail(c, utils.CodeBadRequest, "参数错误: "+err.Error())
		return
	}
	userID := getCurrentUserID(c)
	drawIP := c.ClientIP()
	res, err := service.DrawLottery(userID, req.ActivityID, drawIP)
	if err != nil {
		utils.Fail(c, utils.CodeBadRequest, err.Error())
		return
	}
	utils.Success(c, res)
}

// GetGroupBuyActivities 拼团活动列表
// GET /api/v1/marketing/group-buy-activities
func (h *MarketingHandler) GetGroupBuyActivities(c *gin.Context) {
	list, err := service.GetGroupBuyActivities()
	if err != nil {
		utils.Fail(c, utils.CodeServerError, err.Error())
		return
	}
	utils.Success(c, list)
}

// joinGroupBuyRequest 参与拼团请求
type joinGroupBuyRequest struct {
	ActivityID int64 `json:"activity_id" binding:"required"`
}

// JoinGroupBuy 参与拼团
// POST /api/v1/marketing/group-buy/join
func (h *MarketingHandler) JoinGroupBuy(c *gin.Context) {
	var req joinGroupBuyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Fail(c, utils.CodeBadRequest, "参数错误: "+err.Error())
		return
	}
	userID := getCurrentUserID(c)
	res, err := service.JoinGroupBuy(userID, req.ActivityID)
	if err != nil {
		utils.Fail(c, utils.CodeBadRequest, err.Error())
		return
	}
	utils.Success(c, res)
}

// GenerateInviteQRCode 生成用户专属邀请二维码
// GET /api/v1/marketing/invite-qrcode
func (h *MarketingHandler) GenerateInviteQRCode(c *gin.Context) {
	userID := getCurrentUserID(c)
	content, err := service.GenerateInviteQRCode(userID)
	if err != nil {
		utils.Fail(c, utils.CodeServerError, err.Error())
		return
	}
	utils.Success(c, gin.H{"qrcode": content})
}

// redeemInviteCodeRequest 福利兑换请求
type redeemInviteCodeRequest struct {
	Code string `json:"code" binding:"required"`
}

// RedeemInviteCode 福利兑换(平台通用邀请码)
// POST /api/v1/marketing/redeem
func (h *MarketingHandler) RedeemInviteCode(c *gin.Context) {
	var req redeemInviteCodeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Fail(c, utils.CodeBadRequest, "参数错误: "+err.Error())
		return
	}
	userID := getCurrentUserID(c)
	res, err := service.RedeemInviteCode(userID, req.Code)
	if err != nil {
		utils.Fail(c, utils.CodeBadRequest, err.Error())
		return
	}
	utils.Success(c, res)
}
