package handler

import (
	"github.com/gin-gonic/gin"
	"github.com/jisan/e-sports-platform/internal/service"
	"github.com/jisan/e-sports-platform/internal/utils"
)

// UserHandler 用户资料处理器
type UserHandler struct{}

// NewUserHandler 创建用户处理器
func NewUserHandler() *UserHandler { return &UserHandler{} }

// GetProfile 获取当前用户资料
// GET /api/v1/user/profile
func (h *UserHandler) GetProfile(c *gin.Context) {
	userID := getCurrentUserID(c)
	res, err := service.GetUserProfile(userID)
	if err != nil {
		utils.Fail(c, utils.CodeNotFound, err.Error())
		return
	}
	utils.Success(c, res)
}

// updateProfileRequest 更新用户资料请求
type updateProfileRequest struct {
	Nickname string `json:"nickname"`
	Avatar   string `json:"avatar"`
}

// UpdateProfile 更新用户资料(昵称/头像)
// PUT /api/v1/user/profile
func (h *UserHandler) UpdateProfile(c *gin.Context) {
	var req updateProfileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Fail(c, utils.CodeBadRequest, "参数错误: "+err.Error())
		return
	}
	userID := getCurrentUserID(c)
	u, err := service.UpdateUserProfile(userID, req.Nickname, req.Avatar)
	if err != nil {
		utils.Fail(c, utils.CodeBadRequest, err.Error())
		return
	}
	utils.Success(c, u)
}

// submitRealnameRequest 提交实名认证请求
type submitRealnameRequest struct {
	RealName string `json:"real_name" binding:"required"`
	IDCard   string `json:"id_card" binding:"required"`
}

// SubmitRealname 提交实名认证
// POST /api/v1/user/realname
func (h *UserHandler) SubmitRealname(c *gin.Context) {
	var req submitRealnameRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Fail(c, utils.CodeBadRequest, "参数错误: "+err.Error())
		return
	}
	userID := getCurrentUserID(c)
	if err := service.SubmitRealname(userID, req.RealName, req.IDCard); err != nil {
		utils.Fail(c, utils.CodeBadRequest, err.Error())
		return
	}
	utils.Success(c, gin.H{"msg": "ok"})
}

// GetRealnameStatus 获取实名认证状态
// GET /api/v1/user/realname/status
func (h *UserHandler) GetRealnameStatus(c *gin.Context) {
	userID := getCurrentUserID(c)
	res, err := service.GetRealnameStatus(userID)
	if err != nil {
		utils.Fail(c, utils.CodeBadRequest, err.Error())
		return
	}
	utils.Success(c, res)
}

// faceVerifyRequest 活体检测请求
type faceVerifyRequest struct {
	SessionID string `json:"session_id" binding:"required"`
}

// FaceVerify 活体检测校验
// POST /api/v1/user/face-verify
func (h *UserHandler) FaceVerify(c *gin.Context) {
	var req faceVerifyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Fail(c, utils.CodeBadRequest, "参数错误: "+err.Error())
		return
	}
	userID := getCurrentUserID(c)
	sessionID, err := service.FaceVerify(userID, req.SessionID)
	if err != nil {
		utils.Fail(c, utils.CodeBadRequest, err.Error())
		return
	}
	utils.Success(c, gin.H{"session_id": sessionID})
}

// ToggleFavorite 收藏/取消收藏打手
// POST /api/v1/user/favorites/:player_id
func (h *UserHandler) ToggleFavorite(c *gin.Context) {
	userID := getCurrentUserID(c)
	playerID := parseInt64Path(c, "player_id")
	if playerID == 0 {
		utils.Fail(c, utils.CodeBadRequest, "打手ID不能为空")
		return
	}
	favorited, err := service.ToggleFavorite(userID, playerID)
	if err != nil {
		utils.Fail(c, utils.CodeBadRequest, err.Error())
		return
	}
	utils.Success(c, gin.H{"favorited": favorited})
}

// ListFavorites 收藏的打手列表
// GET /api/v1/user/favorites
func (h *UserHandler) ListFavorites(c *gin.Context) {
	userID := getCurrentUserID(c)
	ids, err := service.ListFavoritePlayers(userID)
	if err != nil {
		utils.Fail(c, utils.CodeBadRequest, err.Error())
		return
	}
	utils.Success(c, gin.H{"player_ids": ids})
}
