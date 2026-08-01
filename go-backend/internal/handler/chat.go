package handler

import (
	"github.com/gin-gonic/gin"
	"github.com/jisan/e-sports-platform/internal/service"
	"github.com/jisan/e-sports-platform/internal/utils"
)

// ChatHandler 聊天处理器
type ChatHandler struct{}

// NewChatHandler 创建聊天处理器
func NewChatHandler() *ChatHandler { return &ChatHandler{} }

// GetSessions 用户会话列表
// GET /api/v1/chat/sessions
func (h *ChatHandler) GetSessions(c *gin.Context) {
	userID := getCurrentUserID(c)
	page, pageSize := getPage(c)
	list, total, err := service.GetChatSessions(userID, page, pageSize)
	if err != nil {
		utils.Fail(c, utils.CodeServerError, err.Error())
		return
	}
	utils.SuccessWithTotal(c, list, total)
}

// GetMessages 会话消息列表
// GET /api/v1/chat/sessions/:id/messages
func (h *ChatHandler) GetMessages(c *gin.Context) {
	sessionID := parseInt64Path(c, "id")
	userID := getCurrentUserID(c)
	page, pageSize := getPage(c)
	list, total, err := service.GetChatMessages(sessionID, userID, page, pageSize)
	if err != nil {
		utils.Fail(c, utils.CodeBadRequest, err.Error())
		return
	}
	utils.SuccessWithTotal(c, list, total)
}

// sendMessageRequest 发送消息请求
type sendMessageRequest struct {
	SessionID int64  `json:"session_id" binding:"required"`
	MsgType   string `json:"msg_type"`
	Content   string `json:"content"`
	MediaURL  string `json:"media_url"`
	Duration  int    `json:"duration"`
	AsrText   string `json:"asr_text"`
}

// SendMessage 发送聊天消息
// POST /api/v1/chat/messages
func (h *ChatHandler) SendMessage(c *gin.Context) {
	var req sendMessageRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Fail(c, utils.CodeBadRequest, "参数错误: "+err.Error())
		return
	}
	userID := getCurrentUserID(c)
	claims, _ := utils.GetClaimsFromContext(c)
	senderType := "user"
	if claims != nil {
		senderType = claims.UserType
	}
	in := &service.SendMessageInput{
		SessionID: req.SessionID,
		MsgType:   req.MsgType,
		Content:   req.Content,
		MediaURL:  req.MediaURL,
		Duration:  req.Duration,
		AsrText:   req.AsrText,
	}
	m, err := service.SendMessage(userID, senderType, in)
	if err != nil {
		utils.Fail(c, utils.CodeBadRequest, err.Error())
		return
	}
	utils.Success(c, m)
}

// RevokeMessage 撤回消息
// POST /api/v1/chat/messages/:id/revoke
func (h *ChatHandler) RevokeMessage(c *gin.Context) {
	messageID := parseInt64Path(c, "id")
	userID := getCurrentUserID(c)
	if err := service.RevokeMessage(messageID, userID); err != nil {
		utils.Fail(c, utils.CodeBadRequest, err.Error())
		return
	}
	utils.Success(c, gin.H{"msg": "ok"})
}

// uploadChatFileRequest 上传聊天文件请求
type uploadChatFileRequest struct {
	FileURL  string `json:"file_url" binding:"required"`
	FileName string `json:"file_name"`
	FileSize int64  `json:"file_size"`
	FileType string `json:"file_type"`
}

// UploadChatFile 上传聊天文件
// POST /api/v1/chat/files
func (h *ChatHandler) UploadChatFile(c *gin.Context) {
	var req uploadChatFileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Fail(c, utils.CodeBadRequest, "参数错误: "+err.Error())
		return
	}
	sessionID := parseInt64Query(c, "session_id", 0)
	userID := getCurrentUserID(c)
	f, err := service.UploadChatFile(sessionID, userID, req.FileURL, req.FileName, req.FileSize, req.FileType)
	if err != nil {
		utils.Fail(c, utils.CodeBadRequest, err.Error())
		return
	}
	utils.Success(c, f)
}

// requestInterventionRequest 请求平台介入请求
type requestInterventionRequest struct {
	Reason string `json:"reason" binding:"required"`
}

// RequestIntervention 请求平台介入(售后会话)
// POST /api/v1/chat/sessions/:id/intervention
func (h *ChatHandler) RequestIntervention(c *gin.Context) {
	sessionID := parseInt64Path(c, "id")
	var req requestInterventionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Fail(c, utils.CodeBadRequest, "参数错误: "+err.Error())
		return
	}
	userID := getCurrentUserID(c)
	if err := service.RequestIntervention(sessionID, userID, req.Reason); err != nil {
		utils.Fail(c, utils.CodeBadRequest, err.Error())
		return
	}
	utils.Success(c, gin.H{"msg": "ok"})
}
