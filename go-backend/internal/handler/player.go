package handler

import (
	"github.com/gin-gonic/gin"
	"github.com/jisan/e-sports-platform/internal/service"
	"github.com/jisan/e-sports-platform/internal/utils"
)

// PlayerHandler 打手处理器(打手侧操作)
type PlayerHandler struct{}

// NewPlayerHandler 创建打手处理器
func NewPlayerHandler() *PlayerHandler { return &PlayerHandler{} }

// GetGrabList 可抢单列表
// GET /api/v1/player/grab-orders
func (h *PlayerHandler) GetGrabList(c *gin.Context) {
	userID := getCurrentUserID(c)
	clubID := parseInt64Query(c, "club_id", 0)
	page, pageSize := getPage(c)
	list, total, err := service.GetGrabOrderList(userID, clubID, page, pageSize)
	if err != nil {
		utils.Fail(c, utils.CodeServerError, err.Error())
		return
	}
	utils.SuccessWithTotal(c, list, total)
}

// GrabOrder 打手抢单
// POST /api/v1/player/grab-orders/:id
func (h *PlayerHandler) GrabOrder(c *gin.Context) {
	orderID := parseInt64Path(c, "id")
	playerID := getCurrentUserID(c)
	if err := service.GrabOrder(orderID, playerID); err != nil {
		utils.Fail(c, utils.CodeBadRequest, err.Error())
		return
	}
	utils.Success(c, gin.H{"msg": "ok"})
}

// GetPlayerOrders 打手接单列表
// GET /api/v1/player/orders
func (h *PlayerHandler) GetPlayerOrders(c *gin.Context) {
	playerID := getCurrentUserID(c)
	page, pageSize := getPage(c)
	status := parseInt8Query(c, "status", -1)
	list, total, err := service.GetPlayerOrders(playerID, page, pageSize, status)
	if err != nil {
		utils.Fail(c, utils.CodeServerError, err.Error())
		return
	}
	utils.SuccessWithTotal(c, list, total)
}

// StartService 开始服务
// POST /api/v1/player/orders/:id/start
func (h *PlayerHandler) StartService(c *gin.Context) {
	orderID := parseInt64Path(c, "id")
	playerID := getCurrentUserID(c)
	if err := service.StartService(orderID, playerID); err != nil {
		utils.Fail(c, utils.CodeBadRequest, err.Error())
		return
	}
	utils.Success(c, gin.H{"msg": "ok"})
}

// CompleteService 完成服务
// POST /api/v1/player/orders/:id/complete
func (h *PlayerHandler) CompleteService(c *gin.Context) {
	orderID := parseInt64Path(c, "id")
	playerID := getCurrentUserID(c)
	if err := service.CompleteService(orderID, playerID); err != nil {
		utils.Fail(c, utils.CodeBadRequest, err.Error())
		return
	}
	utils.Success(c, gin.H{"msg": "ok"})
}

// transferOrderRequest 转单请求
type transferOrderRequest struct {
	ToPlayerID int64 `json:"to_player_id" binding:"required"`
}

// TransferOrder 转单
// POST /api/v1/player/orders/:id/transfer
func (h *PlayerHandler) TransferOrder(c *gin.Context) {
	orderID := parseInt64Path(c, "id")
	var req transferOrderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Fail(c, utils.CodeBadRequest, "参数错误: "+err.Error())
		return
	}
	playerID := getCurrentUserID(c)
	if err := service.TransferOrder(orderID, playerID, req.ToPlayerID); err != nil {
		utils.Fail(c, utils.CodeBadRequest, err.Error())
		return
	}
	utils.Success(c, gin.H{"msg": "ok"})
}

// GetEarnings 打手收益概览
// GET /api/v1/player/earnings
func (h *PlayerHandler) GetEarnings(c *gin.Context) {
	playerID := getCurrentUserID(c)
	res, err := service.GetEarnings(playerID)
	if err != nil {
		utils.Fail(c, utils.CodeServerError, err.Error())
		return
	}
	utils.Success(c, res)
}

// GetFrozenEarnings 冻结收益明细
// GET /api/v1/player/earnings/frozen
func (h *PlayerHandler) GetFrozenEarnings(c *gin.Context) {
	playerID := getCurrentUserID(c)
	list, err := service.GetFrozenEarnings(playerID)
	if err != nil {
		utils.Fail(c, utils.CodeServerError, err.Error())
		return
	}
	utils.Success(c, list)
}

// applyWithdrawRequest 申请提现请求
type applyWithdrawRequest struct {
	Amount  int64             `json:"amount" binding:"required"`
	Channel string            `json:"channel" binding:"required"`
	Bank    map[string]string `json:"bank"`
}

// ApplyWithdraw 申请提现
// POST /api/v1/player/withdraw
func (h *PlayerHandler) ApplyWithdraw(c *gin.Context) {
	var req applyWithdrawRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Fail(c, utils.CodeBadRequest, "参数错误: "+err.Error())
		return
	}
	userID := getCurrentUserID(c)
	w, err := service.ApplyWithdraw(userID, req.Amount, req.Channel, req.Bank)
	if err != nil {
		utils.Fail(c, utils.CodeBadRequest, err.Error())
		return
	}
	utils.Success(c, w)
}

// GetMyServices 打手服务项目列表
// GET /api/v1/player/services
func (h *PlayerHandler) GetMyServices(c *gin.Context) {
	playerID := getCurrentUserID(c)
	list, err := service.GetMyServices(playerID)
	if err != nil {
		utils.Fail(c, utils.CodeServerError, err.Error())
		return
	}
	utils.Success(c, list)
}

// createServiceRequest 创建服务项目请求
type createServiceRequest struct {
	ClubID int64  `json:"club_id" binding:"required"`
	Name   string `json:"name" binding:"required"`
	Amount int64  `json:"amount" binding:"required"`
}

// CreateService 打手创建服务项目
// POST /api/v1/player/services
func (h *PlayerHandler) CreateService(c *gin.Context) {
	var req createServiceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Fail(c, utils.CodeBadRequest, "参数错误: "+err.Error())
		return
	}
	playerID := getCurrentUserID(c)
	if err := service.CreateService(playerID, req.ClubID, req.Name, req.Amount); err != nil {
		utils.Fail(c, utils.CodeBadRequest, err.Error())
		return
	}
	utils.Success(c, gin.H{"msg": "ok"})
}

// GetMyEvaluations 打手收到的评价列表
// GET /api/v1/player/evaluations
func (h *PlayerHandler) GetMyEvaluations(c *gin.Context) {
	playerID := getCurrentUserID(c)
	page, pageSize := getPage(c)
	list, total, err := service.GetMyEvaluations(playerID, page, pageSize)
	if err != nil {
		utils.Fail(c, utils.CodeServerError, err.Error())
		return
	}
	utils.SuccessWithTotal(c, list, total)
}

// appealEvaluationRequest 评价申诉请求
type appealEvaluationRequest struct {
	Reason string `json:"reason" binding:"required"`
}

// AppealEvaluation 打手对评价申诉
// POST /api/v1/player/evaluations/:id/appeal
func (h *PlayerHandler) AppealEvaluation(c *gin.Context) {
	evaluationID := parseInt64Path(c, "id")
	var req appealEvaluationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Fail(c, utils.CodeBadRequest, "参数错误: "+err.Error())
		return
	}
	playerID := getCurrentUserID(c)
	a, err := service.AppealEvaluation(evaluationID, playerID, req.Reason)
	if err != nil {
		utils.Fail(c, utils.CodeBadRequest, err.Error())
		return
	}
	utils.Success(c, a)
}

// GetPlayerDetail 打手详情
// GET /api/v1/player/:id
func (h *PlayerHandler) GetPlayerDetail(c *gin.Context) {
	playerID := parseInt64Path(c, "id")
	res, err := service.GetPlayerDetail(playerID)
	if err != nil {
		utils.Fail(c, utils.CodeNotFound, err.Error())
		return
	}
	utils.Success(c, res)
}

// GetPlayerList 打手列表
// GET /api/v1/players
func (h *PlayerHandler) GetPlayerList(c *gin.Context) {
	clubID := parseInt64Query(c, "club_id", 0)
	page, pageSize := getPage(c)
	list, total, err := service.GetPlayerList(clubID, page, pageSize)
	if err != nil {
		utils.Fail(c, utils.CodeServerError, err.Error())
		return
	}
	utils.SuccessWithTotal(c, list, total)
}

// submitJoinApplicationRequest 提交入会申请请求
type submitJoinApplicationRequest struct {
	ClubID       int64  `json:"club_id" binding:"required"`
	RealName     string `json:"real_name" binding:"required"`
	GameAccount  string `json:"game_account" binding:"required"`
	GameRegion   string `json:"game_region"`
	GoodPosition string `json:"good_position"`
	RankLevel    string `json:"rank_level"`
	Intro        string `json:"intro"`
}

// SubmitJoinApplication 提交入会申请
// POST /api/v1/player/join-application
func (h *PlayerHandler) SubmitJoinApplication(c *gin.Context) {
	var req submitJoinApplicationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Fail(c, utils.CodeBadRequest, "参数错误: "+err.Error())
		return
	}
	userID := getCurrentUserID(c)
	a, err := service.SubmitJoinApplication(userID, req.ClubID, req.RealName, req.GameAccount, req.GameRegion, req.GoodPosition, req.RankLevel, req.Intro)
	if err != nil {
		utils.Fail(c, utils.CodeBadRequest, err.Error())
		return
	}
	utils.Success(c, a)
}

// GetMyApplications 我的入会申请列表
// GET /api/v1/player/join-applications
func (h *PlayerHandler) GetMyApplications(c *gin.Context) {
	userID := getCurrentUserID(c)
	page, pageSize := getPage(c)
	list, total, err := service.GetMyApplications(userID, page, pageSize)
	if err != nil {
		utils.Fail(c, utils.CodeServerError, err.Error())
		return
	}
	utils.SuccessWithTotal(c, list, total)
}
