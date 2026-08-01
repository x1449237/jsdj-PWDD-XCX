package handler

import (
	"github.com/gin-gonic/gin"
	"github.com/jisan/e-sports-platform/internal/service"
	"github.com/jisan/e-sports-platform/internal/utils"
)

// ClubHandler 俱乐部处理器(用户侧浏览)
type ClubHandler struct{}

// NewClubHandler 创建俱乐部处理器
func NewClubHandler() *ClubHandler { return &ClubHandler{} }

// GetClubList 俱乐部列表
// GET /api/v1/clubs
func (h *ClubHandler) GetClubList(c *gin.Context) {
	page, pageSize := getPage(c)
	keyword := c.Query("keyword")
	list, total, err := service.GetClubList(page, pageSize, keyword)
	if err != nil {
		utils.Fail(c, utils.CodeServerError, err.Error())
		return
	}
	utils.SuccessWithTotal(c, list, total)
}

// GetClubDetail 俱乐部详情
// GET /api/v1/clubs/:id
func (h *ClubHandler) GetClubDetail(c *gin.Context) {
	clubID := parseInt64Path(c, "id")
	cl, err := service.GetClubDetail(clubID)
	if err != nil {
		utils.Fail(c, utils.CodeNotFound, err.Error())
		return
	}
	utils.Success(c, cl)
}

// RecordClubJoinClick 记录俱乐部入驻点击(返回客服微信)
// POST /api/v1/clubs/join-click
func (h *ClubHandler) RecordClubJoinClick(c *gin.Context) {
	userID := getCurrentUserID(c)
	wechat, err := service.RecordClubJoinClick(userID)
	if err != nil {
		utils.Fail(c, utils.CodeServerError, err.Error())
		return
	}
	utils.Success(c, gin.H{"customer_wechat": wechat})
}
