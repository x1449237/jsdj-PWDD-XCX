package handler

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/jisan/e-sports-platform/internal/service"
	"github.com/jisan/e-sports-platform/internal/utils"
)

// GuardianHandler 家长(未成年守护)处理器
type GuardianHandler struct{}

// NewGuardianHandler 创建家长处理器
func NewGuardianHandler() *GuardianHandler { return &GuardianHandler{} }

// bindGuardianRequest 绑定未成年账户请求
type bindGuardianRequest struct {
	ChildUID int64 `json:"child_uid" binding:"required"`
}

// BindGuardian 家长绑定未成年账户
// POST /api/v1/guardian/bind
func (h *GuardianHandler) BindGuardian(c *gin.Context) {
	var req bindGuardianRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Fail(c, utils.CodeBadRequest, "参数错误: "+err.Error())
		return
	}
	parentUID := getCurrentUserID(c)
	b, err := service.BindGuardian(parentUID, req.ChildUID)
	if err != nil {
		utils.Fail(c, utils.CodeBadRequest, err.Error())
		return
	}
	utils.Success(c, b)
}

// GetChildReport 未成年消费报告
// GET /api/v1/guardian/children/:id/report
func (h *GuardianHandler) GetChildReport(c *gin.Context) {
	childUID := parseInt64Path(c, "id")
	parentUID := getCurrentUserID(c)
	res, err := service.GetChildReport(parentUID, childUID)
	if err != nil {
		utils.Fail(c, utils.CodeBadRequest, err.Error())
		return
	}
	utils.Success(c, res)
}

// updateChildSettingsRequest 更新未成年设置请求
type updateChildSettingsRequest struct {
	MonthlyLimit int64 `json:"monthly_limit"`
	AllowOrder   int8  `json:"allow_order"`
	AllowReward  int8  `json:"allow_reward"`
	IsFrozen     int8  `json:"is_frozen"`
}

// UpdateChildSettings 家长更新未成年设置
// PUT /api/v1/guardian/children/:id/settings
func (h *GuardianHandler) UpdateChildSettings(c *gin.Context) {
	childUID := parseInt64Path(c, "id")
	var req updateChildSettingsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Fail(c, utils.CodeBadRequest, "参数错误: "+err.Error())
		return
	}
	parentUID := getCurrentUserID(c)
	if err := service.UpdateChildSettings(parentUID, childUID, req.MonthlyLimit, req.AllowOrder, req.AllowReward, req.IsFrozen); err != nil {
		utils.Fail(c, utils.CodeBadRequest, err.Error())
		return
	}
	utils.Success(c, gin.H{"msg": "ok"})
}

// freezeChildRequest 冻结/解冻请求
type freezeChildRequest struct {
	Freeze bool `json:"freeze"`
}

// FreezeChild 家长冻结/解冻未成年账户
// POST /api/v1/guardian/children/:id/freeze
func (h *GuardianHandler) FreezeChild(c *gin.Context) {
	childUID := parseInt64Path(c, "id")
	var req freezeChildRequest
	_ = c.ShouldBindJSON(&req)
	parentUID := getCurrentUserID(c)
	if err := service.FreezeChild(parentUID, childUID, req.Freeze); err != nil {
		utils.Fail(c, utils.CodeBadRequest, err.Error())
		return
	}
	utils.Success(c, gin.H{"msg": "ok"})
}

// parseBoolQuery 解析布尔型 query 参数
func parseBoolQuery(c *gin.Context, key string) bool {
	s := c.Query(key)
	b, _ := strconv.ParseBool(s)
	return b
}
