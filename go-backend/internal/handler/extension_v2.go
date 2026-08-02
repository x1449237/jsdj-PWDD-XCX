package handler

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/jisan/e-sports-platform/internal/model"
	"github.com/jisan/e-sports-platform/internal/service"
	"github.com/jisan/e-sports-platform/internal/utils"
)

// ExtensionHandler 俱乐部小助手APP扩展功能handler
type ExtensionHandler struct{}

// NewExtensionHandler 构造函数
func NewExtensionHandler() *ExtensionHandler {
	return &ExtensionHandler{}
}

// ============================================================
// 一、俱乐部管理扩展（内置管理端 /shop/club-ext/*）
// ============================================================

// -------------------- 主页装修 --------------------

// GetHomeDecoration GET /shop/club-ext/home-decoration
func (h *ExtensionHandler) GetHomeDecoration(c *gin.Context) {
	clubID := getClubID(c)
	data, err := service.ClubGetHomeDecoration(clubID)
	if err != nil {
		utils.Fail(c, utils.CodeBadRequest, err.Error())
		return
	}
	utils.Success(c, data)
}

// UpdateHomeDecoration PUT /shop/club-ext/home-decoration
func (h *ExtensionHandler) UpdateHomeDecoration(c *gin.Context) {
	clubID := getClubID(c)
	fields := make(map[string]interface{})
	if err := c.ShouldBindJSON(&fields); err != nil {
		utils.Fail(c, utils.CodeBadRequest, err.Error())
		return
	}
	if err := service.ClubUpdateHomeDecoration(clubID, fields); err != nil {
		utils.Fail(c, utils.CodeBadRequest, err.Error())
		return
	}
	utils.Success(c, gin.H{"club_id": clubID})
}

// -------------------- 俱乐部技能项目 --------------------

// ListClubServices GET /shop/club-ext/services
func (h *ExtensionHandler) ListClubServices(c *gin.Context) {
	clubID := getClubID(c)
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	size, _ := strconv.Atoi(c.DefaultQuery("size", "20"))
	list, total, err := service.ClubListServices(clubID, page, size)
	if err != nil {
		utils.Fail(c, utils.CodeBadRequest, err.Error())
		return
	}
	utils.Success(c, gin.H{"list": list, "total": total})
}

// CreateClubService POST /shop/club-ext/services
func (h *ExtensionHandler) CreateClubService(c *gin.Context) {
	clubID := getClubID(c)
	var req model.ClubService
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Fail(c, utils.CodeBadRequest, err.Error())
		return
	}
	if err := service.ClubCreateService(clubID, &req); err != nil {
		utils.Fail(c, utils.CodeBadRequest, err.Error())
		return
	}
	utils.Success(c, req)
}

// UpdateClubService PUT /shop/club-ext/services/:id
func (h *ExtensionHandler) UpdateClubService(c *gin.Context) {
	clubID := getClubID(c)
	serviceID := parseInt64Path(c, "id")
	fields := make(map[string]interface{})
	if err := c.ShouldBindJSON(&fields); err != nil {
		utils.Fail(c, utils.CodeBadRequest, err.Error())
		return
	}
	if err := service.ClubUpdateService(clubID, serviceID, fields); err != nil {
		utils.Fail(c, utils.CodeBadRequest, err.Error())
		return
	}
	utils.Success(c, gin.H{"id": serviceID})
}

// DeleteClubService DELETE /shop/club-ext/services/:id
func (h *ExtensionHandler) DeleteClubService(c *gin.Context) {
	clubID := getClubID(c)
	serviceID := parseInt64Path(c, "id")
	if err := service.ClubDeleteService(clubID, serviceID); err != nil {
		utils.Fail(c, utils.CodeBadRequest, err.Error())
		return
	}
	utils.Success(c, gin.H{"id": serviceID})
}

// -------------------- 成员技能名片 --------------------

// GetMemberCard GET /shop/club-ext/member-card
func (h *ExtensionHandler) GetMemberCard(c *gin.Context) {
	clubID := getClubID(c)
	memberUID, _ := strconv.ParseInt(c.Query("userId"), 10, 64)
	data, err := service.ClubGetMemberCard(clubID, memberUID)
	if err != nil {
		utils.Fail(c, utils.CodeBadRequest, err.Error())
		return
	}
	utils.Success(c, data)
}

// UpdateMemberCard PUT /shop/club-ext/member-card
func (h *ExtensionHandler) UpdateMemberCard(c *gin.Context) {
	clubID := getClubID(c)
	memberUID, _ := strconv.ParseInt(c.Query("userId"), 10, 64)
	fields := make(map[string]interface{})
	if err := c.ShouldBindJSON(&fields); err != nil {
		utils.Fail(c, utils.CodeBadRequest, err.Error())
		return
	}
	if err := service.ClubUpdateMemberCard(clubID, memberUID, fields); err != nil {
		utils.Fail(c, utils.CodeBadRequest, err.Error())
		return
	}
	utils.Success(c, gin.H{"club_id": clubID, "member_uid": memberUID})
}

// -------------------- 成员档案 --------------------

// GetMemberProfile GET /shop/club-ext/member-profile
func (h *ExtensionHandler) GetMemberProfile(c *gin.Context) {
	clubID := getClubID(c)
	memberUID, _ := strconv.ParseInt(c.Query("userId"), 10, 64)
	gameID, _ := strconv.ParseInt(c.Query("gameId"), 10, 64)
	data, err := service.ClubGetMemberProfile(clubID, memberUID, gameID)
	if err != nil {
		utils.Fail(c, utils.CodeBadRequest, err.Error())
		return
	}
	utils.Success(c, data)
}

// UpdateMemberProfile PUT /shop/club-ext/member-profile
func (h *ExtensionHandler) UpdateMemberProfile(c *gin.Context) {
	clubID := getClubID(c)
	memberUID, _ := strconv.ParseInt(c.Query("userId"), 10, 64)
	gameID, _ := strconv.ParseInt(c.Query("gameId"), 10, 64)
	fields := make(map[string]interface{})
	if err := c.ShouldBindJSON(&fields); err != nil {
		utils.Fail(c, utils.CodeBadRequest, err.Error())
		return
	}
	if err := service.ClubUpdateMemberProfile(clubID, memberUID, gameID, fields); err != nil {
		utils.Fail(c, utils.CodeBadRequest, err.Error())
		return
	}
	utils.Success(c, gin.H{"club_id": clubID, "member_uid": memberUID, "game_id": gameID})
}

// -------------------- 权限操作日志 --------------------

// ListPermissionLogs GET /shop/club-ext/permission-logs
func (h *ExtensionHandler) ListPermissionLogs(c *gin.Context) {
	clubID := getClubID(c)
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	size, _ := strconv.Atoi(c.DefaultQuery("size", "20"))
	list, total, err := service.ClubListPermissionLogs(clubID, page, size)
	if err != nil {
		utils.Fail(c, utils.CodeBadRequest, err.Error())
		return
	}
	utils.Success(c, gin.H{"list": list, "total": total})
}

// -------------------- 退俱乐部申报 --------------------

// ListResignations GET /shop/club-ext/resignations
func (h *ExtensionHandler) ListResignations(c *gin.Context) {
	clubID := getClubID(c)
	statusInt, _ := strconv.Atoi(c.DefaultQuery("status", "-1"))
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	size, _ := strconv.Atoi(c.DefaultQuery("size", "20"))
	list, total, err := service.ClubListResignations(clubID, int8(statusInt), page, size)
	if err != nil {
		utils.Fail(c, utils.CodeBadRequest, err.Error())
		return
	}
	utils.Success(c, gin.H{"list": list, "total": total})
}

// AuditResignation PUT /shop/club-ext/resignations/:id/audit
func (h *ExtensionHandler) AuditResignation(c *gin.Context) {
	clubID := getClubID(c)
	id := parseInt64Path(c, "id")
	auditorUID := getCurrentUserID(c)
	var req struct {
		Status int8 `json:"status"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Fail(c, utils.CodeBadRequest, err.Error())
		return
	}
	if err := service.ClubAuditResignation(clubID, id, req.Status, auditorUID); err != nil {
		utils.Fail(c, utils.CodeBadRequest, err.Error())
		return
	}
	utils.Success(c, gin.H{"id": id, "status": req.Status})
}

// -------------------- 黑名单 --------------------

// ListBlacklists GET /shop/club-ext/blacklists
func (h *ExtensionHandler) ListBlacklists(c *gin.Context) {
	clubID := getClubID(c)
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	size, _ := strconv.Atoi(c.DefaultQuery("size", "20"))
	list, total, err := service.ClubListBlacklists(clubID, page, size)
	if err != nil {
		utils.Fail(c, utils.CodeBadRequest, err.Error())
		return
	}
	utils.Success(c, gin.H{"list": list, "total": total})
}

// AddBlacklist POST /shop/club-ext/blacklists
func (h *ExtensionHandler) AddBlacklist(c *gin.Context) {
	clubID := getClubID(c)
	operatorUID := getCurrentUserID(c)
	var req model.ClubBlacklist
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Fail(c, utils.CodeBadRequest, err.Error())
		return
	}
	req.ClubID = clubID
	req.OperatorUID = operatorUID
	if err := service.ClubAddBlacklist(&req); err != nil {
		utils.Fail(c, utils.CodeBadRequest, err.Error())
		return
	}
	utils.Success(c, req)
}

// RemoveBlacklist DELETE /shop/club-ext/blacklists/:userId
func (h *ExtensionHandler) RemoveBlacklist(c *gin.Context) {
	clubID := getClubID(c)
	userID, _ := strconv.ParseInt(c.Param("userId"), 10, 64)
	if err := service.ClubRemoveBlacklist(clubID, userID); err != nil {
		utils.Fail(c, utils.CodeBadRequest, err.Error())
		return
	}
	utils.Success(c, gin.H{"userId": userID})
}

// -------------------- 积分体系 --------------------

// ListPointRules GET /shop/club-ext/point-rules
func (h *ExtensionHandler) ListPointRules(c *gin.Context) {
	clubID := getClubID(c)
	list, err := service.ClubListPointRules(clubID)
	if err != nil {
		utils.Fail(c, utils.CodeBadRequest, err.Error())
		return
	}
	utils.Success(c, list)
}

// UpdatePointRule PUT /shop/club-ext/point-rules/:id
func (h *ExtensionHandler) UpdatePointRule(c *gin.Context) {
	clubID := getClubID(c)
	id := parseInt64Path(c, "id")
	var req struct {
		Points      int    `json:"points"`
		Description string `json:"description"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Fail(c, utils.CodeBadRequest, err.Error())
		return
	}
	if err := service.ClubUpdatePointRule(clubID, id, req.Points, req.Description); err != nil {
		utils.Fail(c, utils.CodeBadRequest, err.Error())
		return
	}
	utils.Success(c, gin.H{"id": id})
}

// ListPointLogs GET /shop/club-ext/point-logs
func (h *ExtensionHandler) ListPointLogs(c *gin.Context) {
	clubID := getClubID(c)
	memberUID, _ := strconv.ParseInt(c.Query("memberUid"), 10, 64)
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	size, _ := strconv.Atoi(c.DefaultQuery("size", "20"))
	list, total, err := service.ClubListPointLogs(clubID, memberUID, page, size)
	if err != nil {
		utils.Fail(c, utils.CodeBadRequest, err.Error())
		return
	}
	utils.Success(c, gin.H{"list": list, "total": total})
}

// -------------------- 团费规则 --------------------

// GetFeeRule GET /shop/club-ext/fee-rule
func (h *ExtensionHandler) GetFeeRule(c *gin.Context) {
	clubID := getClubID(c)
	data, err := service.ClubGetFeeRule(clubID)
	if err != nil {
		utils.Fail(c, utils.CodeBadRequest, err.Error())
		return
	}
	utils.Success(c, data)
}

// SaveFeeRule PUT /shop/club-ext/fee-rule
func (h *ExtensionHandler) SaveFeeRule(c *gin.Context) {
	clubID := getClubID(c)
	var req model.ClubFeeRule
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Fail(c, utils.CodeBadRequest, err.Error())
		return
	}
	if err := service.ClubSaveFeeRule(clubID, &req); err != nil {
		utils.Fail(c, utils.CodeBadRequest, err.Error())
		return
	}
	utils.Success(c, req)
}

// -------------------- 招募卡片 --------------------

// ListRecruitCards GET /shop/club-ext/recruit-cards
func (h *ExtensionHandler) ListRecruitCards(c *gin.Context) {
	clubID := getClubID(c)
	list, err := service.ClubListRecruitCards(clubID)
	if err != nil {
		utils.Fail(c, utils.CodeBadRequest, err.Error())
		return
	}
	utils.Success(c, list)
}

// SaveRecruitCard POST /shop/club-ext/recruit-cards
func (h *ExtensionHandler) SaveRecruitCard(c *gin.Context) {
	clubID := getClubID(c)
	var req model.ClubRecruitCard
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Fail(c, utils.CodeBadRequest, err.Error())
		return
	}
	req.ClubID = clubID
	if err := service.ClubSaveRecruitCard(&req); err != nil {
		utils.Fail(c, utils.CodeBadRequest, err.Error())
		return
	}
	utils.Success(c, req)
}

// -------------------- 管理员待办 --------------------

// ListAdminTodos GET /shop/club-ext/todos
func (h *ExtensionHandler) ListAdminTodos(c *gin.Context) {
	clubID := getClubID(c)
	adminUID, _ := strconv.ParseInt(c.Query("adminUid"), 10, 64)
	statusInt, _ := strconv.Atoi(c.DefaultQuery("status", "-1"))
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	size, _ := strconv.Atoi(c.DefaultQuery("size", "20"))
	list, total, err := service.ClubListAdminTodos(clubID, adminUID, int8(statusInt), page, size)
	if err != nil {
		utils.Fail(c, utils.CodeBadRequest, err.Error())
		return
	}
	utils.Success(c, gin.H{"list": list, "total": total})
}

// CompleteAdminTodo PUT /shop/club-ext/todos/:id/complete
func (h *ExtensionHandler) CompleteAdminTodo(c *gin.Context) {
	id := parseInt64Path(c, "id")
	if err := service.ClubCompleteAdminTodo(id); err != nil {
		utils.Fail(c, utils.CodeBadRequest, err.Error())
		return
	}
	utils.Success(c, gin.H{"id": id})
}

// -------------------- 游戏分区 --------------------

// ListGameZones GET /shop/club-ext/game-zones
func (h *ExtensionHandler) ListGameZones(c *gin.Context) {
	clubID := getClubID(c)
	list, err := service.ClubListGameZones(clubID)
	if err != nil {
		utils.Fail(c, utils.CodeBadRequest, err.Error())
		return
	}
	utils.Success(c, list)
}

// SaveGameZone POST /shop/club-ext/game-zones
func (h *ExtensionHandler) SaveGameZone(c *gin.Context) {
	clubID := getClubID(c)
	var req model.ClubGameZone
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Fail(c, utils.CodeBadRequest, err.Error())
		return
	}
	req.ClubID = clubID
	if err := service.ClubSaveGameZone(&req); err != nil {
		utils.Fail(c, utils.CodeBadRequest, err.Error())
		return
	}
	utils.Success(c, req)
}

// -------------------- 临时抽成 --------------------

// ListTempCommissionRules GET /shop/club-ext/temp-commission
func (h *ExtensionHandler) ListTempCommissionRules(c *gin.Context) {
	clubID := getClubID(c)
	list, err := service.ClubListTempCommissionRules(clubID)
	if err != nil {
		utils.Fail(c, utils.CodeBadRequest, err.Error())
		return
	}
	utils.Success(c, list)
}

// CreateTempCommissionRule POST /shop/club-ext/temp-commission
func (h *ExtensionHandler) CreateTempCommissionRule(c *gin.Context) {
	clubID := getClubID(c)
	var req model.ClubTempCommissionRule
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Fail(c, utils.CodeBadRequest, err.Error())
		return
	}
	req.ClubID = clubID
	if err := service.ClubCreateTempCommissionRule(&req); err != nil {
		utils.Fail(c, utils.CodeBadRequest, err.Error())
		return
	}
	utils.Success(c, req)
}

// -------------------- 请假管理 --------------------

// ListLeaves GET /shop/club-ext/leaves
func (h *ExtensionHandler) ListLeaves(c *gin.Context) {
	clubID := getClubID(c)
	statusInt, _ := strconv.Atoi(c.DefaultQuery("status", "-1"))
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	size, _ := strconv.Atoi(c.DefaultQuery("size", "20"))
	list, total, err := service.ClubListLeaves(clubID, int8(statusInt), page, size)
	if err != nil {
		utils.Fail(c, utils.CodeBadRequest, err.Error())
		return
	}
	utils.Success(c, gin.H{"list": list, "total": total})
}

// AuditLeave PUT /shop/club-ext/leaves/:id/audit
func (h *ExtensionHandler) AuditLeave(c *gin.Context) {
	clubID := getClubID(c)
	id := parseInt64Path(c, "id")
	auditorUID := getCurrentUserID(c)
	var req struct {
		Status int8 `json:"status"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Fail(c, utils.CodeBadRequest, err.Error())
		return
	}
	if err := service.ClubAuditLeave(clubID, id, req.Status, auditorUID); err != nil {
		utils.Fail(c, utils.CodeBadRequest, err.Error())
		return
	}
	utils.Success(c, gin.H{"id": id, "status": req.Status})
}

// -------------------- 资料修改审核 --------------------

// ListChangeRequests GET /shop/club-ext/change-requests
func (h *ExtensionHandler) ListChangeRequests(c *gin.Context) {
	clubID := getClubID(c)
	statusInt, _ := strconv.Atoi(c.DefaultQuery("status", "-1"))
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	size, _ := strconv.Atoi(c.DefaultQuery("size", "20"))
	list, total, err := service.ClubListChangeRequests(clubID, int8(statusInt), page, size)
	if err != nil {
		utils.Fail(c, utils.CodeBadRequest, err.Error())
		return
	}
	utils.Success(c, gin.H{"list": list, "total": total})
}

// AuditChangeRequest PUT /shop/club-ext/change-requests/:id/audit
func (h *ExtensionHandler) AuditChangeRequest(c *gin.Context) {
	clubID := getClubID(c)
	id := parseInt64Path(c, "id")
	auditorUID := getCurrentUserID(c)
	var req struct {
		Status int8 `json:"status"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Fail(c, utils.CodeBadRequest, err.Error())
		return
	}
	if err := service.ClubAuditChangeRequest(clubID, id, req.Status, auditorUID); err != nil {
		utils.Fail(c, utils.CodeBadRequest, err.Error())
		return
	}
	utils.Success(c, gin.H{"id": id, "status": req.Status})
}

// -------------------- 优先派单 --------------------

// ListPriorityDispatch GET /shop/club-ext/priority-dispatch
func (h *ExtensionHandler) ListPriorityDispatch(c *gin.Context) {
	clubID := getClubID(c)
	list, err := service.ClubListPriorityDispatch(clubID)
	if err != nil {
		utils.Fail(c, utils.CodeBadRequest, err.Error())
		return
	}
	utils.Success(c, list)
}

// SetPriority POST /shop/club-ext/priority-dispatch
func (h *ExtensionHandler) SetPriority(c *gin.Context) {
	clubID := getClubID(c)
	var req struct {
		MemberUID int64 `json:"member_uid"`
		Priority  int   `json:"priority"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Fail(c, utils.CodeBadRequest, err.Error())
		return
	}
	if err := service.ClubSetPriority(clubID, req.MemberUID, req.Priority); err != nil {
		utils.Fail(c, utils.CodeBadRequest, err.Error())
		return
	}
	utils.Success(c, req)
}

// -------------------- 内部资源单 --------------------

// ListInternalResources GET /shop/club-ext/internal-resources
func (h *ExtensionHandler) ListInternalResources(c *gin.Context) {
	clubID := getClubID(c)
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	size, _ := strconv.Atoi(c.DefaultQuery("size", "20"))
	list, total, err := service.ClubListInternalResources(clubID, page, size)
	if err != nil {
		utils.Fail(c, utils.CodeBadRequest, err.Error())
		return
	}
	utils.Success(c, gin.H{"list": list, "total": total})
}

// CreateInternalResource POST /shop/club-ext/internal-resources
func (h *ExtensionHandler) CreateInternalResource(c *gin.Context) {
	clubID := getClubID(c)
	var req model.ClubInternalResource
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Fail(c, utils.CodeBadRequest, err.Error())
		return
	}
	req.ClubID = clubID
	if err := service.ClubCreateInternalResource(&req); err != nil {
		utils.Fail(c, utils.CodeBadRequest, err.Error())
		return
	}
	utils.Success(c, req)
}

// -------------------- 客户归属 --------------------

// ListCustomerRelations GET /shop/club-ext/customer-relations
func (h *ExtensionHandler) ListCustomerRelations(c *gin.Context) {
	clubID := getClubID(c)
	receptionistUID, _ := strconv.ParseInt(c.Query("receptionistUid"), 10, 64)
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	size, _ := strconv.Atoi(c.DefaultQuery("size", "20"))
	list, total, err := service.ClubListCustomerRelations(clubID, receptionistUID, page, size)
	if err != nil {
		utils.Fail(c, utils.CodeBadRequest, err.Error())
		return
	}
	utils.Success(c, gin.H{"list": list, "total": total})
}

// SaveCustomerRelation POST /shop/club-ext/customer-relations
func (h *ExtensionHandler) SaveCustomerRelation(c *gin.Context) {
	clubID := getClubID(c)
	var req model.ClubCustomerRelation
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Fail(c, utils.CodeBadRequest, err.Error())
		return
	}
	req.ClubID = clubID
	if err := service.ClubSaveCustomerRelation(&req); err != nil {
		utils.Fail(c, utils.CodeBadRequest, err.Error())
		return
	}
	utils.Success(c, req)
}

// -------------------- 模板话术 --------------------

// ListTemplatePhrases GET /shop/club-ext/template-phrases
func (h *ExtensionHandler) ListTemplatePhrases(c *gin.Context) {
	clubID := getClubID(c)
	category := c.Query("category")
	list, err := service.ClubListTemplatePhrases(clubID, category)
	if err != nil {
		utils.Fail(c, utils.CodeBadRequest, err.Error())
		return
	}
	utils.Success(c, list)
}

// SaveTemplatePhrase POST /shop/club-ext/template-phrases
func (h *ExtensionHandler) SaveTemplatePhrase(c *gin.Context) {
	clubID := getClubID(c)
	var req model.ClubTemplatePhrase
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Fail(c, utils.CodeBadRequest, err.Error())
		return
	}
	req.ClubID = clubID
	if err := service.ClubSaveTemplatePhrase(&req); err != nil {
		utils.Fail(c, utils.CodeBadRequest, err.Error())
		return
	}
	utils.Success(c, req)
}

// DeleteTemplatePhrase DELETE /shop/club-ext/template-phrases/:id
func (h *ExtensionHandler) DeleteTemplatePhrase(c *gin.Context) {
	clubID := getClubID(c)
	id := parseInt64Path(c, "id")
	if err := service.ClubDeleteTemplatePhrase(clubID, id); err != nil {
		utils.Fail(c, utils.CodeBadRequest, err.Error())
		return
	}
	utils.Success(c, gin.H{"id": id})
}

// -------------------- 业绩排行 --------------------

// GetRanking GET /shop/club-ext/rankings
func (h *ExtensionHandler) GetRanking(c *gin.Context) {
	clubID := getClubID(c)
	periodTypeInt, _ := strconv.Atoi(c.DefaultQuery("periodType", "0"))
	periodDate := c.Query("periodDate")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	size, _ := strconv.Atoi(c.DefaultQuery("size", "20"))
	list, total, err := service.ClubGetRanking(clubID, int8(periodTypeInt), periodDate, page, size)
	if err != nil {
		utils.Fail(c, utils.CodeBadRequest, err.Error())
		return
	}
	utils.Success(c, gin.H{"list": list, "total": total})
}

// -------------------- 俱乐部设置 --------------------

// UpdateClubSettings PUT /shop/club-ext/settings
func (h *ExtensionHandler) UpdateClubSettings(c *gin.Context) {
	clubID := getClubID(c)
	fields := make(map[string]interface{})
	if err := c.ShouldBindJSON(&fields); err != nil {
		utils.Fail(c, utils.CodeBadRequest, err.Error())
		return
	}
	if err := service.ClubUpdateSettings(clubID, fields); err != nil {
		utils.Fail(c, utils.CodeBadRequest, err.Error())
		return
	}
	utils.Success(c, gin.H{"club_id": clubID})
}

// TransferFounder POST /shop/club-ext/transfer-founder
func (h *ExtensionHandler) TransferFounder(c *gin.Context) {
	clubID := getClubID(c)
	var req struct {
		ToUID int64 `json:"to_uid"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Fail(c, utils.CodeBadRequest, err.Error())
		return
	}
	if err := service.ClubTransferFounder(clubID, req.ToUID); err != nil {
		utils.Fail(c, utils.CodeBadRequest, err.Error())
		return
	}
	utils.Success(c, gin.H{"club_id": clubID, "to_uid": req.ToUID})
}

// SetMemberBan PUT /shop/club-ext/members/:userId/ban
func (h *ExtensionHandler) SetMemberBan(c *gin.Context) {
	clubID := getClubID(c)
	memberUID, _ := strconv.ParseInt(c.Param("userId"), 10, 64)
	var req struct {
		IsBanned int8 `json:"is_banned"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Fail(c, utils.CodeBadRequest, err.Error())
		return
	}
	if err := service.ClubSetMemberBan(clubID, memberUID, req.IsBanned); err != nil {
		utils.Fail(c, utils.CodeBadRequest, err.Error())
		return
	}
	utils.Success(c, gin.H{"member_uid": memberUID, "is_banned": req.IsBanned})
}

// SetMemberRole PUT /shop/club-ext/members/:userId/role
func (h *ExtensionHandler) SetMemberRole(c *gin.Context) {
	clubID := getClubID(c)
	memberUID, _ := strconv.ParseInt(c.Param("userId"), 10, 64)
	var req struct {
		Role       int8 `json:"role"`
		RoleDetail int8 `json:"role_detail"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Fail(c, utils.CodeBadRequest, err.Error())
		return
	}
	if err := service.ClubSetMemberRole(clubID, memberUID, req.Role, req.RoleDetail); err != nil {
		utils.Fail(c, utils.CodeBadRequest, err.Error())
		return
	}
	utils.Success(c, gin.H{"member_uid": memberUID, "role": req.Role, "role_detail": req.RoleDetail})
}

// -------------------- 用户端申报类 --------------------

// CreateResignation POST /api/v1/user/club-resignation
func (h *ExtensionHandler) CreateResignation(c *gin.Context) {
	userID := getCurrentUserID(c)
	var req model.ClubMemberResignation
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Fail(c, utils.CodeBadRequest, err.Error())
		return
	}
	req.MemberUID = userID
	if err := service.ClubCreateResignation(&req); err != nil {
		utils.Fail(c, utils.CodeBadRequest, err.Error())
		return
	}
	utils.Success(c, req)
}

// CreateLeave POST /api/v1/user/club-leave
func (h *ExtensionHandler) CreateLeave(c *gin.Context) {
	userID := getCurrentUserID(c)
	var req model.ClubMemberLeave
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Fail(c, utils.CodeBadRequest, err.Error())
		return
	}
	req.MemberUID = userID
	if err := service.ClubCreateLeave(&req); err != nil {
		utils.Fail(c, utils.CodeBadRequest, err.Error())
		return
	}
	utils.Success(c, req)
}

// CreateChangeRequest POST /api/v1/user/club-change-request
func (h *ExtensionHandler) CreateChangeRequest(c *gin.Context) {
	userID := getCurrentUserID(c)
	var req model.ClubMemberChangeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Fail(c, utils.CodeBadRequest, err.Error())
		return
	}
	req.MemberUID = userID
	if err := service.ClubCreateChangeRequest(&req); err != nil {
		utils.Fail(c, utils.CodeBadRequest, err.Error())
		return
	}
	utils.Success(c, req)
}

// ============================================================
// 二、订单扩展（用户端 /api/v1/order-ext/* 和管理端 /shop/order-ext/*）
// ============================================================

// ListOrderTemplates GET /order-ext/templates
func (h *ExtensionHandler) ListOrderTemplates(c *gin.Context) {
	userID := getCurrentUserID(c)
	list, err := service.OrderListTemplates(userID)
	if err != nil {
		utils.Fail(c, utils.CodeBadRequest, err.Error())
		return
	}
	utils.Success(c, list)
}

// CreateOrderTemplate POST /order-ext/templates
func (h *ExtensionHandler) CreateOrderTemplate(c *gin.Context) {
	userID := getCurrentUserID(c)
	var req model.OrderTemplate
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Fail(c, utils.CodeBadRequest, err.Error())
		return
	}
	req.UserID = userID
	if err := service.OrderCreateTemplate(&req); err != nil {
		utils.Fail(c, utils.CodeBadRequest, err.Error())
		return
	}
	utils.Success(c, req)
}

// DeleteOrderTemplate DELETE /order-ext/templates/:id
func (h *ExtensionHandler) DeleteOrderTemplate(c *gin.Context) {
	userID := getCurrentUserID(c)
	id := parseInt64Path(c, "id")
	if err := service.OrderDeleteTemplate(id, userID); err != nil {
		utils.Fail(c, utils.CodeBadRequest, err.Error())
		return
	}
	utils.Success(c, gin.H{"id": id})
}

// CreateSupplement POST /order-ext/supplements
func (h *ExtensionHandler) CreateSupplement(c *gin.Context) {
	userID := getCurrentUserID(c)
	var req model.OrderSupplement
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Fail(c, utils.CodeBadRequest, err.Error())
		return
	}
	req.UserID = userID
	if err := service.OrderCreateSupplement(&req); err != nil {
		utils.Fail(c, utils.CodeBadRequest, err.Error())
		return
	}
	utils.Success(c, req)
}

// ListSupplements GET /order-ext/supplements
func (h *ExtensionHandler) ListSupplements(c *gin.Context) {
	orderID, _ := strconv.ParseInt(c.Query("orderId"), 10, 64)
	list, err := service.OrderListSupplements(orderID)
	if err != nil {
		utils.Fail(c, utils.CodeBadRequest, err.Error())
		return
	}
	utils.Success(c, list)
}

// CreatePartialRefund POST /order-ext/partial-refunds
func (h *ExtensionHandler) CreatePartialRefund(c *gin.Context) {
	userID := getCurrentUserID(c)
	var req model.OrderPartialRefund
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Fail(c, utils.CodeBadRequest, err.Error())
		return
	}
	req.OperatorID = userID
	req.OperatorType = "user"
	if err := service.OrderCreatePartialRefund(&req); err != nil {
		utils.Fail(c, utils.CodeBadRequest, err.Error())
		return
	}
	utils.Success(c, req)
}

// ListPartialRefunds GET /order-ext/partial-refunds
func (h *ExtensionHandler) ListPartialRefunds(c *gin.Context) {
	orderID, _ := strconv.ParseInt(c.Query("orderId"), 10, 64)
	list, err := service.OrderListPartialRefunds(orderID)
	if err != nil {
		utils.Fail(c, utils.CodeBadRequest, err.Error())
		return
	}
	utils.Success(c, list)
}

// AuditPartialRefund PUT /order-ext/partial-refunds/:id/audit (管理端)
func (h *ExtensionHandler) AuditPartialRefund(c *gin.Context) {
	id := parseInt64Path(c, "id")
	auditorUID := getCurrentUserID(c)
	var req struct {
		Status int8 `json:"status"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Fail(c, utils.CodeBadRequest, err.Error())
		return
	}
	if err := service.OrderAuditPartialRefund(id, req.Status, auditorUID); err != nil {
		utils.Fail(c, utils.CodeBadRequest, err.Error())
		return
	}
	utils.Success(c, gin.H{"id": id, "status": req.Status})
}

// AddOrderRemark POST /order-ext/remarks
func (h *ExtensionHandler) AddOrderRemark(c *gin.Context) {
	userID := getCurrentUserID(c)
	var req model.OrderRemark
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Fail(c, utils.CodeBadRequest, err.Error())
		return
	}
	req.UserID = userID
	req.UserType = "user"
	if err := service.OrderAddRemark(&req); err != nil {
		utils.Fail(c, utils.CodeBadRequest, err.Error())
		return
	}
	utils.Success(c, req)
}

// ListOrderRemarks GET /order-ext/remarks
func (h *ExtensionHandler) ListOrderRemarks(c *gin.Context) {
	orderID, _ := strconv.ParseInt(c.Query("orderId"), 10, 64)
	list, err := service.OrderListRemarks(orderID)
	if err != nil {
		utils.Fail(c, utils.CodeBadRequest, err.Error())
		return
	}
	utils.Success(c, list)
}

// CreateOrderExtension POST /order-ext/extensions
func (h *ExtensionHandler) CreateOrderExtension(c *gin.Context) {
	userID := getCurrentUserID(c)
	var req model.OrderExtension
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Fail(c, utils.CodeBadRequest, err.Error())
		return
	}
	req.PlayerID = userID
	if err := service.OrderCreateExtension(&req); err != nil {
		utils.Fail(c, utils.CodeBadRequest, err.Error())
		return
	}
	utils.Success(c, req)
}

// AuditOrderExtension PUT /order-ext/extensions/:id/audit (管理端)
func (h *ExtensionHandler) AuditOrderExtension(c *gin.Context) {
	id := parseInt64Path(c, "id")
	auditorUID := getCurrentUserID(c)
	var req struct {
		Status int8 `json:"status"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Fail(c, utils.CodeBadRequest, err.Error())
		return
	}
	if err := service.OrderAuditExtension(id, req.Status, auditorUID); err != nil {
		utils.Fail(c, utils.CodeBadRequest, err.Error())
		return
	}
	utils.Success(c, gin.H{"id": id, "status": req.Status})
}

// CreateOrderTransfer POST /order-ext/transfers
func (h *ExtensionHandler) CreateOrderTransfer(c *gin.Context) {
	userID := getCurrentUserID(c)
	var req model.OrderTransfer
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Fail(c, utils.CodeBadRequest, err.Error())
		return
	}
	req.FromPlayerID = userID
	if err := service.OrderCreateTransfer(&req); err != nil {
		utils.Fail(c, utils.CodeBadRequest, err.Error())
		return
	}
	utils.Success(c, req)
}

// AuditOrderTransfer PUT /order-ext/transfers/:id/audit
func (h *ExtensionHandler) AuditOrderTransfer(c *gin.Context) {
	id := parseInt64Path(c, "id")
	var req struct {
		Status int8 `json:"status"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Fail(c, utils.CodeBadRequest, err.Error())
		return
	}
	if err := service.OrderAuditTransfer(id, req.Status); err != nil {
		utils.Fail(c, utils.CodeBadRequest, err.Error())
		return
	}
	utils.Success(c, gin.H{"id": id, "status": req.Status})
}

// CreateOrderPriceChange POST /order-ext/price-changes
func (h *ExtensionHandler) CreateOrderPriceChange(c *gin.Context) {
	userID := getCurrentUserID(c)
	var req model.OrderPriceChange
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Fail(c, utils.CodeBadRequest, err.Error())
		return
	}
	req.ApplicantID = userID
	if err := service.OrderCreatePriceChange(&req); err != nil {
		utils.Fail(c, utils.CodeBadRequest, err.Error())
		return
	}
	utils.Success(c, req)
}

// AuditOrderPriceChange PUT /order-ext/price-changes/:id/audit (管理端)
func (h *ExtensionHandler) AuditOrderPriceChange(c *gin.Context) {
	id := parseInt64Path(c, "id")
	auditorUID := getCurrentUserID(c)
	var req struct {
		Status int8 `json:"status"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Fail(c, utils.CodeBadRequest, err.Error())
		return
	}
	if err := service.OrderAuditPriceChange(id, req.Status, auditorUID); err != nil {
		utils.Fail(c, utils.CodeBadRequest, err.Error())
		return
	}
	utils.Success(c, gin.H{"id": id, "status": req.Status})
}

// ListOrderPriceLogs GET /order-ext/price-logs
func (h *ExtensionHandler) ListOrderPriceLogs(c *gin.Context) {
	orderID, _ := strconv.ParseInt(c.Query("orderId"), 10, 64)
	list, err := service.OrderListPriceLogs(orderID)
	if err != nil {
		utils.Fail(c, utils.CodeBadRequest, err.Error())
		return
	}
	utils.Success(c, list)
}

// AddOrderTag POST /order-ext/tags
func (h *ExtensionHandler) AddOrderTag(c *gin.Context) {
	var req struct {
		OrderID int64 `json:"order_id"`
		TagID   int64 `json:"tag_id"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Fail(c, utils.CodeBadRequest, err.Error())
		return
	}
	if err := service.OrderAddTag(req.OrderID, req.TagID); err != nil {
		utils.Fail(c, utils.CodeBadRequest, err.Error())
		return
	}
	utils.Success(c, req)
}

// RemoveOrderTag DELETE /order-ext/tags
func (h *ExtensionHandler) RemoveOrderTag(c *gin.Context) {
	var req struct {
		OrderID int64 `json:"order_id"`
		TagID   int64 `json:"tag_id"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Fail(c, utils.CodeBadRequest, err.Error())
		return
	}
	if err := service.OrderRemoveTag(req.OrderID, req.TagID); err != nil {
		utils.Fail(c, utils.CodeBadRequest, err.Error())
		return
	}
	utils.Success(c, req)
}

// ListOrderTags GET /order-ext/tags
func (h *ExtensionHandler) ListOrderTags(c *gin.Context) {
	list, err := service.OrderListTags()
	if err != nil {
		utils.Fail(c, utils.CodeBadRequest, err.Error())
		return
	}
	utils.Success(c, list)
}

// ListRefundLedger GET /order-ext/refund-ledger
func (h *ExtensionHandler) ListRefundLedger(c *gin.Context) {
	orderID, _ := strconv.ParseInt(c.Query("orderId"), 10, 64)
	list, err := service.OrderListRefundLedger(orderID)
	if err != nil {
		utils.Fail(c, utils.CodeBadRequest, err.Error())
		return
	}
	utils.Success(c, list)
}

// MarkOrderAbnormal PUT /order-ext/:id/abnormal (管理端)
func (h *ExtensionHandler) MarkOrderAbnormal(c *gin.Context) {
	orderID := parseInt64Path(c, "id")
	var req struct {
		IsAbnormal int8 `json:"is_abnormal"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Fail(c, utils.CodeBadRequest, err.Error())
		return
	}
	if err := service.OrderMarkAbnormal(orderID, req.IsAbnormal); err != nil {
		utils.Fail(c, utils.CodeBadRequest, err.Error())
		return
	}
	utils.Success(c, gin.H{"order_id": orderID, "is_abnormal": req.IsAbnormal})
}

// ArchiveOrder PUT /order-ext/:id/archive (管理端)
func (h *ExtensionHandler) ArchiveOrder(c *gin.Context) {
	orderID := parseInt64Path(c, "id")
	if err := service.OrderArchive(orderID); err != nil {
		utils.Fail(c, utils.CodeBadRequest, err.Error())
		return
	}
	utils.Success(c, gin.H{"id": orderID})
}

// -------------------- 俱乐部收藏（用户端） --------------------

// FavoriteClub POST /api/v1/user/favorite-clubs
func (h *ExtensionHandler) FavoriteClub(c *gin.Context) {
	userID := getCurrentUserID(c)
	var req struct {
		ClubID int64 `json:"club_id"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Fail(c, utils.CodeBadRequest, err.Error())
		return
	}
	if err := service.UserFavoriteClub(userID, req.ClubID); err != nil {
		utils.Fail(c, utils.CodeBadRequest, err.Error())
		return
	}
	utils.Success(c, req)
}

// UnfavoriteClub DELETE /api/v1/user/favorite-clubs/:clubId
func (h *ExtensionHandler) UnfavoriteClub(c *gin.Context) {
	userID := getCurrentUserID(c)
	clubID, _ := strconv.ParseInt(c.Param("clubId"), 10, 64)
	if err := service.UserUnfavoriteClub(userID, clubID); err != nil {
		utils.Fail(c, utils.CodeBadRequest, err.Error())
		return
	}
	utils.Success(c, gin.H{"clubId": clubID})
}

// ListFavoriteClubs GET /api/v1/user/favorite-clubs
func (h *ExtensionHandler) ListFavoriteClubs(c *gin.Context) {
	userID := getCurrentUserID(c)
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	size, _ := strconv.Atoi(c.DefaultQuery("size", "20"))
	list, total, err := service.UserListFavoriteClubs(userID, page, size)
	if err != nil {
		utils.Fail(c, utils.CodeBadRequest, err.Error())
		return
	}
	utils.Success(c, gin.H{"list": list, "total": total})
}

// ============================================================
// 三、IM扩展
// ============================================================

// ListGroupFiles GET /chat-ext/group-files
func (h *ExtensionHandler) ListGroupFiles(c *gin.Context) {
	groupID, _ := strconv.ParseInt(c.Query("groupId"), 10, 64)
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	size, _ := strconv.Atoi(c.DefaultQuery("size", "20"))
	list, total, err := service.GroupChatListFiles(groupID, page, size)
	if err != nil {
		utils.Fail(c, utils.CodeBadRequest, err.Error())
		return
	}
	utils.Success(c, gin.H{"list": list, "total": total})
}

// ListQuickReplies GET /chat-ext/quick-replies
func (h *ExtensionHandler) ListQuickReplies(c *gin.Context) {
	clubID := getClubID(c)
	category := c.Query("category")
	list, err := service.ChatListQuickReplies(clubID, category)
	if err != nil {
		utils.Fail(c, utils.CodeBadRequest, err.Error())
		return
	}
	utils.Success(c, list)
}

// SaveQuickReply POST /chat-ext/quick-replies (管理端)
func (h *ExtensionHandler) SaveQuickReply(c *gin.Context) {
	clubID := getClubID(c)
	var req model.ChatQuickReply
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Fail(c, utils.CodeBadRequest, err.Error())
		return
	}
	req.ClubID = clubID
	if err := service.ChatSaveQuickReply(&req); err != nil {
		utils.Fail(c, utils.CodeBadRequest, err.Error())
		return
	}
	utils.Success(c, req)
}

// MuteGroupMember POST /chat-ext/mute (管理端)
func (h *ExtensionHandler) MuteGroupMember(c *gin.Context) {
	operatorUID := getCurrentUserID(c)
	var req model.GroupChatMute
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Fail(c, utils.CodeBadRequest, err.Error())
		return
	}
	req.OperatorUID = operatorUID
	if err := service.GroupChatMuteMember(&req); err != nil {
		utils.Fail(c, utils.CodeBadRequest, err.Error())
		return
	}
	utils.Success(c, req)
}

// UnmuteGroupMember DELETE /chat-ext/mute (管理端)
func (h *ExtensionHandler) UnmuteGroupMember(c *gin.Context) {
	var req struct {
		GroupID   int64 `json:"group_id"`
		MemberUID int64 `json:"member_uid"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Fail(c, utils.CodeBadRequest, err.Error())
		return
	}
	if err := service.GroupChatUnmuteMember(req.GroupID, req.MemberUID); err != nil {
		utils.Fail(c, utils.CodeBadRequest, err.Error())
		return
	}
	utils.Success(c, req)
}

// CreateChatReport POST /chat-ext/reports
func (h *ExtensionHandler) CreateChatReport(c *gin.Context) {
	userID := getCurrentUserID(c)
	var req model.ChatReport
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Fail(c, utils.CodeBadRequest, err.Error())
		return
	}
	req.ReporterUID = userID
	if err := service.ChatCreateReport(&req); err != nil {
		utils.Fail(c, utils.CodeBadRequest, err.Error())
		return
	}
	utils.Success(c, req)
}

// ListChatReports GET /chat-ext/reports (管理端)
func (h *ExtensionHandler) ListChatReports(c *gin.Context) {
	statusInt, _ := strconv.Atoi(c.DefaultQuery("status", "-1"))
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	size, _ := strconv.Atoi(c.DefaultQuery("size", "20"))
	list, total, err := service.ChatListReports(int8(statusInt), page, size)
	if err != nil {
		utils.Fail(c, utils.CodeBadRequest, err.Error())
		return
	}
	utils.Success(c, gin.H{"list": list, "total": total})
}

// HandleChatReport PUT /chat-ext/reports/:id/handle (管理端)
func (h *ExtensionHandler) HandleChatReport(c *gin.Context) {
	id := parseInt64Path(c, "id")
	handlerUID := getCurrentUserID(c)
	var req struct {
		Status int8   `json:"status"`
		Result string `json:"result"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Fail(c, utils.CodeBadRequest, err.Error())
		return
	}
	if err := service.ChatHandleReport(id, req.Status, handlerUID, req.Result); err != nil {
		utils.Fail(c, utils.CodeBadRequest, err.Error())
		return
	}
	utils.Success(c, gin.H{"id": id, "status": req.Status})
}

// TogglePinSession PUT /chat-ext/sessions/:id/pin
func (h *ExtensionHandler) TogglePinSession(c *gin.Context) {
	sessionID := parseInt64Path(c, "id")
	userID := getCurrentUserID(c)
	var req struct {
		IsPinned int8 `json:"is_pinned"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Fail(c, utils.CodeBadRequest, err.Error())
		return
	}
	if err := service.ChatTogglePinSession(userID, sessionID, req.IsPinned); err != nil {
		utils.Fail(c, utils.CodeBadRequest, err.Error())
		return
	}
	utils.Success(c, gin.H{"session_id": sessionID, "is_pinned": req.IsPinned})
}

// ============================================================
// 四、财务扩展（管理端 /shop/finance-ext/* 和 /admin/finance-ext/*）
// ============================================================

// ListFinanceLedger GET /shop/finance-ext/ledger
func (h *ExtensionHandler) ListFinanceLedger(c *gin.Context) {
	clubID := getClubID(c)
	ledgerType := c.Query("type")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	size, _ := strconv.Atoi(c.DefaultQuery("size", "20"))
	list, total, err := service.ClubListFinanceLedger(clubID, ledgerType, page, size)
	if err != nil {
		utils.Fail(c, utils.CodeBadRequest, err.Error())
		return
	}
	utils.Success(c, gin.H{"list": list, "total": total})
}

// ListMonthlySettlements GET /shop/finance-ext/settlements
func (h *ExtensionHandler) ListMonthlySettlements(c *gin.Context) {
	clubID := getClubID(c)
	month := c.Query("month")
	list, err := service.ClubListMonthlySettlements(clubID, month)
	if err != nil {
		utils.Fail(c, utils.CodeBadRequest, err.Error())
		return
	}
	utils.Success(c, list)
}

// ListRebateRecords GET /shop/finance-ext/rebates
func (h *ExtensionHandler) ListRebateRecords(c *gin.Context) {
	clubID := getClubID(c)
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	size, _ := strconv.Atoi(c.DefaultQuery("size", "20"))
	list, total, err := service.ClubListRebateRecords(clubID, page, size)
	if err != nil {
		utils.Fail(c, utils.CodeBadRequest, err.Error())
		return
	}
	utils.Success(c, gin.H{"list": list, "total": total})
}

// CreateRebateRecord POST /shop/finance-ext/rebates
func (h *ExtensionHandler) CreateRebateRecord(c *gin.Context) {
	clubID := getClubID(c)
	var req model.RebateRecord
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Fail(c, utils.CodeBadRequest, err.Error())
		return
	}
	req.ClubID = clubID
	if err := service.ClubCreateRebateRecord(&req); err != nil {
		utils.Fail(c, utils.CodeBadRequest, err.Error())
		return
	}
	utils.Success(c, req)
}

// AuditRebateRecord PUT /shop/finance-ext/rebates/:id/audit
func (h *ExtensionHandler) AuditRebateRecord(c *gin.Context) {
	id := parseInt64Path(c, "id")
	var req struct {
		Status int8 `json:"status"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Fail(c, utils.CodeBadRequest, err.Error())
		return
	}
	if err := service.ClubAuditRebateRecord(id, req.Status); err != nil {
		utils.Fail(c, utils.CodeBadRequest, err.Error())
		return
	}
	utils.Success(c, gin.H{"id": id, "status": req.Status})
}

// ListWalletChangeLogs GET /api/v1/user/wallet-logs (用户端)
func (h *ExtensionHandler) ListWalletChangeLogs(c *gin.Context) {
	userID := getCurrentUserID(c)
	changeType := c.Query("type")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	size, _ := strconv.Atoi(c.DefaultQuery("size", "20"))
	list, total, err := service.UserListWalletChangeLogs(userID, changeType, page, size)
	if err != nil {
		utils.Fail(c, utils.CodeBadRequest, err.Error())
		return
	}
	utils.Success(c, gin.H{"list": list, "total": total})
}

// ListDeposits GET /api/v1/user/deposits (用户端)
func (h *ExtensionHandler) ListDeposits(c *gin.Context) {
	userID := getCurrentUserID(c)
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	size, _ := strconv.Atoi(c.DefaultQuery("size", "20"))
	list, total, err := service.UserListDeposits(userID, page, size)
	if err != nil {
		utils.Fail(c, utils.CodeBadRequest, err.Error())
		return
	}
	utils.Success(c, gin.H{"list": list, "total": total})
}

// CreateDeposit POST /api/v1/user/deposits (用户端)
func (h *ExtensionHandler) CreateDeposit(c *gin.Context) {
	userID := getCurrentUserID(c)
	var req model.UserDeposit
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Fail(c, utils.CodeBadRequest, err.Error())
		return
	}
	req.UserID = userID
	if err := service.UserCreateDeposit(&req); err != nil {
		utils.Fail(c, utils.CodeBadRequest, err.Error())
		return
	}
	utils.Success(c, req)
}

// -------------------- 管理端处罚模板 --------------------

// ListPunishmentTemplates GET /admin/punishment-templates
func (h *ExtensionHandler) ListPunishmentTemplates(c *gin.Context) {
	list, err := service.ClubListPunishmentTemplates()
	if err != nil {
		utils.Fail(c, utils.CodeBadRequest, err.Error())
		return
	}
	utils.Success(c, list)
}

// SavePunishmentTemplate POST /admin/punishment-templates
func (h *ExtensionHandler) SavePunishmentTemplate(c *gin.Context) {
	var req model.PunishmentTemplate
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Fail(c, utils.CodeBadRequest, err.Error())
		return
	}
	if err := service.ClubSavePunishmentTemplate(&req); err != nil {
		utils.Fail(c, utils.CodeBadRequest, err.Error())
		return
	}
	utils.Success(c, req)
}

// ============================================================
// 五、UX/营销扩展（用户端 /api/v1/user/* 和管理端 /admin/*）
// ============================================================

// CreateFeedback POST /api/v1/user/feedbacks
func (h *ExtensionHandler) CreateFeedback(c *gin.Context) {
	userID := getCurrentUserID(c)
	var req model.UserFeedback
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Fail(c, utils.CodeBadRequest, err.Error())
		return
	}
	req.UserID = userID
	if err := service.UserCreateFeedback(&req); err != nil {
		utils.Fail(c, utils.CodeBadRequest, err.Error())
		return
	}
	utils.Success(c, req)
}

// ListUserFeedbacks GET /api/v1/user/feedbacks
func (h *ExtensionHandler) ListUserFeedbacks(c *gin.Context) {
	userID := getCurrentUserID(c)
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	size, _ := strconv.Atoi(c.DefaultQuery("size", "20"))
	list, total, err := service.UserListFeedbacks(userID, page, size)
	if err != nil {
		utils.Fail(c, utils.CodeBadRequest, err.Error())
		return
	}
	utils.Success(c, gin.H{"list": list, "total": total})
}

// ListAllFeedbacks GET /admin/feedbacks
func (h *ExtensionHandler) ListAllFeedbacks(c *gin.Context) {
	statusInt, _ := strconv.Atoi(c.DefaultQuery("status", "-1"))
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	size, _ := strconv.Atoi(c.DefaultQuery("size", "20"))
	list, total, err := service.AdminListFeedbacks(int8(statusInt), page, size)
	if err != nil {
		utils.Fail(c, utils.CodeBadRequest, err.Error())
		return
	}
	utils.Success(c, gin.H{"list": list, "total": total})
}

// ReplyFeedback PUT /admin/feedbacks/:id/reply
func (h *ExtensionHandler) ReplyFeedback(c *gin.Context) {
	id := parseInt64Path(c, "id")
	handlerUID := getCurrentUserID(c)
	var req struct {
		Reply string `json:"reply"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Fail(c, utils.CodeBadRequest, err.Error())
		return
	}
	if err := service.AdminReplyFeedback(id, handlerUID, req.Reply); err != nil {
		utils.Fail(c, utils.CodeBadRequest, err.Error())
		return
	}
	utils.Success(c, gin.H{"id": id})
}

// BlockPlayer POST /api/v1/user/blocklist
func (h *ExtensionHandler) BlockPlayer(c *gin.Context) {
	userID := getCurrentUserID(c)
	var req struct {
		BlockedUID int64 `json:"blocked_uid"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Fail(c, utils.CodeBadRequest, err.Error())
		return
	}
	if err := service.UserBlockPlayer(userID, req.BlockedUID); err != nil {
		utils.Fail(c, utils.CodeBadRequest, err.Error())
		return
	}
	utils.Success(c, req)
}

// UnblockPlayer DELETE /api/v1/user/blocklist/:playerId
func (h *ExtensionHandler) UnblockPlayer(c *gin.Context) {
	userID := getCurrentUserID(c)
	playerID, _ := strconv.ParseInt(c.Param("playerId"), 10, 64)
	if err := service.UserUnblockPlayer(userID, playerID); err != nil {
		utils.Fail(c, utils.CodeBadRequest, err.Error())
		return
	}
	utils.Success(c, gin.H{"playerId": playerID})
}

// GetNotificationSettings GET /api/v1/user/notification-settings
func (h *ExtensionHandler) GetNotificationSettings(c *gin.Context) {
	userID := getCurrentUserID(c)
	data, err := service.UserGetNotificationSettings(userID)
	if err != nil {
		utils.Fail(c, utils.CodeBadRequest, err.Error())
		return
	}
	utils.Success(c, data)
}

// UpdateNotificationSettings PUT /api/v1/user/notification-settings
func (h *ExtensionHandler) UpdateNotificationSettings(c *gin.Context) {
	userID := getCurrentUserID(c)
	fields := make(map[string]interface{})
	if err := c.ShouldBindJSON(&fields); err != nil {
		utils.Fail(c, utils.CodeBadRequest, err.Error())
		return
	}
	if err := service.UserUpdateNotificationSettings(userID, fields); err != nil {
		utils.Fail(c, utils.CodeBadRequest, err.Error())
		return
	}
	utils.Success(c, gin.H{"user_id": userID})
}

// ListActivityPopups GET /api/v1/activity-popups (用户端)
func (h *ExtensionHandler) ListActivityPopups(c *gin.Context) {
	clubID, _ := strconv.ParseInt(c.Query("clubId"), 10, 64)
	list, err := service.UserListActivityPopups(clubID)
	if err != nil {
		utils.Fail(c, utils.CodeBadRequest, err.Error())
		return
	}
	utils.Success(c, list)
}

// AdminListActivityPopups GET /admin/activity-popups
func (h *ExtensionHandler) AdminListActivityPopups(c *gin.Context) {
	clubID := getClubID(c)
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	size, _ := strconv.Atoi(c.DefaultQuery("size", "20"))
	list, total, err := service.AdminListActivityPopups(clubID, page, size)
	if err != nil {
		utils.Fail(c, utils.CodeBadRequest, err.Error())
		return
	}
	utils.Success(c, gin.H{"list": list, "total": total})
}

// AdminSaveActivityPopup POST /admin/activity-popups
func (h *ExtensionHandler) AdminSaveActivityPopup(c *gin.Context) {
	var req model.ActivityPopup
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Fail(c, utils.CodeBadRequest, err.Error())
		return
	}
	if err := service.AdminSaveActivityPopup(&req); err != nil {
		utils.Fail(c, utils.CodeBadRequest, err.Error())
		return
	}
	utils.Success(c, req)
}

// ListFestivalTemplates GET /admin/festival-templates
func (h *ExtensionHandler) ListFestivalTemplates(c *gin.Context) {
	list, err := service.AdminListFestivalTemplates()
	if err != nil {
		utils.Fail(c, utils.CodeBadRequest, err.Error())
		return
	}
	utils.Success(c, list)
}

// SaveFestivalTemplate POST /admin/festival-templates
func (h *ExtensionHandler) SaveFestivalTemplate(c *gin.Context) {
	var req model.FestivalTemplate
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Fail(c, utils.CodeBadRequest, err.Error())
		return
	}
	if err := service.AdminSaveFestivalTemplate(&req); err != nil {
		utils.Fail(c, utils.CodeBadRequest, err.Error())
		return
	}
	utils.Success(c, req)
}

// ListPromoChannels GET /admin/promo-channels
func (h *ExtensionHandler) ListPromoChannels(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	size, _ := strconv.Atoi(c.DefaultQuery("size", "20"))
	list, total, err := service.AdminListPromoChannels(page, size)
	if err != nil {
		utils.Fail(c, utils.CodeBadRequest, err.Error())
		return
	}
	utils.Success(c, gin.H{"list": list, "total": total})
}

// CreatePromoChannel POST /admin/promo-channels
func (h *ExtensionHandler) CreatePromoChannel(c *gin.Context) {
	var req model.PromoChannel
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Fail(c, utils.CodeBadRequest, err.Error())
		return
	}
	if err := service.AdminCreatePromoChannel(&req); err != nil {
		utils.Fail(c, utils.CodeBadRequest, err.Error())
		return
	}
	utils.Success(c, req)
}
