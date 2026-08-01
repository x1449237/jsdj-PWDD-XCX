package handler

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/jisan/e-sports-platform/internal/service"
	"github.com/jisan/e-sports-platform/internal/utils"
)

// DistributorHandler 分销商处理器
type DistributorHandler struct{}

// NewDistributorHandler 创建分销商处理器
func NewDistributorHandler() *DistributorHandler { return &DistributorHandler{} }

// GetSubordinates 分销商下级列表
// GET /api/v1/distributor/subordinates
func (h *DistributorHandler) GetSubordinates(c *gin.Context) {
	distributorID := getCurrentUserID(c)
	level := int8(0)
	if s := c.Query("level"); s != "" {
		if n, err := strconv.ParseInt(s, 10, 8); err == nil {
			level = int8(n)
		}
	}
	list, err := service.GetSubordinates(distributorID, level)
	if err != nil {
		utils.Fail(c, utils.CodeServerError, err.Error())
		return
	}
	utils.Success(c, list)
}

// GetCommissionList 分销佣金记录
// GET /api/v1/distributor/commissions
func (h *DistributorHandler) GetCommissionList(c *gin.Context) {
	distributorID := getCurrentUserID(c)
	page, pageSize := getPage(c)
	list, total, err := service.GetCommissionList(distributorID, page, pageSize)
	if err != nil {
		utils.Fail(c, utils.CodeServerError, err.Error())
		return
	}
	utils.SuccessWithTotal(c, list, total)
}

// GetRanking 分销商排行榜
// GET /api/v1/distributor/ranking
func (h *DistributorHandler) GetRanking(c *gin.Context) {
	limit := parseIntQuery(c, "limit", 50)
	list, err := service.GetDistributorRanking(limit)
	if err != nil {
		utils.Fail(c, utils.CodeServerError, err.Error())
		return
	}
	utils.Success(c, list)
}
