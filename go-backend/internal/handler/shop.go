package handler

import (
	"github.com/gin-gonic/gin"
	"github.com/jisan/e-sports-platform/internal/service"
	"github.com/jisan/e-sports-platform/internal/utils"
)

// ShopHandler 内置管理端(俱乐部)处理器
type ShopHandler struct{}

// NewShopHandler 创建内置管理端处理器
func NewShopHandler() *ShopHandler { return &ShopHandler{} }

// getClubID 获取当前俱乐部ID(内置管理端强制携带)
func getClubID(c *gin.Context) int64 {
	return getClubScopeID(c)
}

// GetClubInfo 获取俱乐部信息
// GET /api/v1/shop/club
func (h *ShopHandler) GetClubInfo(c *gin.Context) {
	clubID := getClubID(c)
	cl, err := service.ShopGetClubInfo(clubID)
	if err != nil {
		utils.Fail(c, utils.CodeNotFound, err.Error())
		return
	}
	utils.Success(c, cl)
}

// updateClubInfoRequest 更新俱乐部信息请求
type updateClubInfoRequest struct {
	Name        string `json:"name"`
	Abbreviation string `json:"abbreviation"`
	Logo        string `json:"logo"`
	Intro       string `json:"intro"`
}

// UpdateClubInfo 更新俱乐部信息
// PUT /api/v1/shop/club
func (h *ShopHandler) UpdateClubInfo(c *gin.Context) {
	var req updateClubInfoRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Fail(c, utils.CodeBadRequest, "参数错误: "+err.Error())
		return
	}
	clubID := getClubID(c)
	fields := map[string]interface{}{
		"name": req.Name, "abbreviation": req.Abbreviation,
		"logo": req.Logo, "intro": req.Intro,
	}
	if err := service.ShopUpdateClubInfo(clubID, fields); err != nil {
		utils.Fail(c, utils.CodeBadRequest, err.Error())
		return
	}
	utils.Success(c, gin.H{"msg": "ok"})
}

// GetApplications 俱乐部入会申请列表
// GET /api/v1/shop/applications
func (h *ShopHandler) GetApplications(c *gin.Context) {
	clubID := getClubID(c)
	page, pageSize := getPage(c)
	status := c.Query("status")
	list, total, err := service.ShopGetApplications(clubID, page, pageSize, status)
	if err != nil {
		utils.Fail(c, utils.CodeServerError, err.Error())
		return
	}
	utils.SuccessWithTotal(c, list, total)
}

// GetApplicationDetail 入会申请详情
// GET /api/v1/shop/applications/:id
func (h *ShopHandler) GetApplicationDetail(c *gin.Context) {
	applicationID := parseInt64Path(c, "id")
	clubID := getClubID(c)
	a, err := service.ShopGetApplicationDetail(applicationID, clubID)
	if err != nil {
		utils.Fail(c, utils.CodeNotFound, err.Error())
		return
	}
	utils.Success(c, a)
}

// startExamRequest 开始考核请求
type startExamRequest struct {
	Requirement string `json:"requirement"`
}

// StartExam 开始考核
// POST /api/v1/shop/applications/:id/exam/start
// 安全修复:传入 clubID 校验申请归属
func (h *ShopHandler) StartExam(c *gin.Context) {
	applicationID := parseInt64Path(c, "id")
	var req startExamRequest
	_ = c.ShouldBindJSON(&req)
	clubID := getClubID(c)
	examinerID := getCurrentUserID(c)
	if err := service.ShopStartExam(clubID, applicationID, examinerID, req.Requirement); err != nil {
		utils.Fail(c, utils.CodeBadRequest, err.Error())
		return
	}
	utils.Success(c, gin.H{"msg": "ok"})
}

// submitExamResultRequest 提交考核结果请求
type submitExamResultRequest struct {
	Result   string `json:"result" binding:"required"`
	Remark   string `json:"remark"`
	VideoURL string `json:"video_url"`
}

// SubmitExamResult 提交考核结果
// POST /api/v1/shop/applications/:id/exam/submit
// 安全修复:传入 clubID 校验申请归属
func (h *ShopHandler) SubmitExamResult(c *gin.Context) {
	applicationID := parseInt64Path(c, "id")
	var req submitExamResultRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Fail(c, utils.CodeBadRequest, "参数错误: "+err.Error())
		return
	}
	clubID := getClubID(c)
	examinerID := getCurrentUserID(c)
	if err := service.ShopSubmitExamResult(clubID, applicationID, examinerID, req.Result, req.Remark, req.VideoURL); err != nil {
		utils.Fail(c, utils.CodeBadRequest, err.Error())
		return
	}
	utils.Success(c, gin.H{"msg": "ok"})
}

// ApproveApplication 通过入会申请
// POST /api/v1/shop/applications/:id/approve
// 安全修复:传入 clubID 校验申请归属
func (h *ShopHandler) ApproveApplication(c *gin.Context) {
	applicationID := parseInt64Path(c, "id")
	clubID := getClubID(c)
	if err := service.ShopApproveApplication(clubID, applicationID); err != nil {
		utils.Fail(c, utils.CodeBadRequest, err.Error())
		return
	}
	utils.Success(c, gin.H{"msg": "ok"})
}

// rejectApplicationRequest 驳回入会申请请求
type rejectApplicationRequest struct {
	Reason string `json:"reason"`
}

// RejectApplication 驳回入会申请
// POST /api/v1/shop/applications/:id/reject
// 安全修复:传入 clubID 校验申请归属
func (h *ShopHandler) RejectApplication(c *gin.Context) {
	applicationID := parseInt64Path(c, "id")
	var req rejectApplicationRequest
	_ = c.ShouldBindJSON(&req)
	clubID := getClubID(c)
	if err := service.ShopRejectApplication(clubID, applicationID, req.Reason); err != nil {
		utils.Fail(c, utils.CodeBadRequest, err.Error())
		return
	}
	utils.Success(c, gin.H{"msg": "ok"})
}

// GetGamers 俱乐部打手列表
// GET /api/v1/shop/gamers
func (h *ShopHandler) GetGamers(c *gin.Context) {
	clubID := getClubID(c)
	page, pageSize := getPage(c)
	list, total, err := service.ShopGetGamers(clubID, page, pageSize)
	if err != nil {
		utils.Fail(c, utils.CodeServerError, err.Error())
		return
	}
	utils.SuccessWithTotal(c, list, total)
}

// GetGamerDetail 打手详情
// GET /api/v1/shop/gamers/:id
func (h *ShopHandler) GetGamerDetail(c *gin.Context) {
	gamerID := parseInt64Path(c, "id")
	clubID := getClubID(c)
	res, err := service.ShopGetGamerDetail(clubID, gamerID)
	if err != nil {
		utils.Fail(c, utils.CodeNotFound, err.Error())
		return
	}
	utils.Success(c, res)
}

// ApproveGamer 审核打手通过
// POST /api/v1/shop/gamers/:id/approve
func (h *ShopHandler) ApproveGamer(c *gin.Context) {
	gamerID := parseInt64Path(c, "id")
	clubID := getClubID(c)
	if err := service.ShopApproveGamer(clubID, gamerID); err != nil {
		utils.Fail(c, utils.CodeBadRequest, err.Error())
		return
	}
	utils.Success(c, gin.H{"msg": "ok"})
}

// RemoveGamer 移除打手
// POST /api/v1/shop/gamers/:id/remove
func (h *ShopHandler) RemoveGamer(c *gin.Context) {
	gamerID := parseInt64Path(c, "id")
	clubID := getClubID(c)
	if err := service.ShopRemoveGamer(clubID, gamerID); err != nil {
		utils.Fail(c, utils.CodeBadRequest, err.Error())
		return
	}
	utils.Success(c, gin.H{"msg": "ok"})
}

// GetGamerEvaluations 打手评价列表
// GET /api/v1/shop/gamers/:id/evaluations
func (h *ShopHandler) GetGamerEvaluations(c *gin.Context) {
	gamerID := parseInt64Path(c, "id")
	page, pageSize := getPage(c)
	list, total, err := service.ShopGetGamerEvaluations(gamerID, page, pageSize)
	if err != nil {
		utils.Fail(c, utils.CodeServerError, err.Error())
		return
	}
	utils.SuccessWithTotal(c, list, total)
}

// GetAdmins 俱乐部管理员列表
// GET /api/v1/shop/admins
func (h *ShopHandler) GetAdmins(c *gin.Context) {
	clubID := getClubID(c)
	list, err := service.ShopGetAdmins(clubID)
	if err != nil {
		utils.Fail(c, utils.CodeServerError, err.Error())
		return
	}
	utils.Success(c, list)
}

// addShopAdminRequest 添加管理员请求
type addShopAdminRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
	RealName string `json:"real_name"`
	Phone    string `json:"phone"`
	Role     int8   `json:"role"`
}

// AddAdmin 添加内置管理端账号(创始人专属)
// POST /api/v1/shop/admins
// 安全修复:传入 operatorID 校验操作者是创始人
func (h *ShopHandler) AddAdmin(c *gin.Context) {
	var req addShopAdminRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Fail(c, utils.CodeBadRequest, "参数错误: "+err.Error())
		return
	}
	clubID := getClubID(c)
	operatorID := getCurrentUserID(c)
	a, err := service.ShopAddAdmin(clubID, operatorID, req.Username, req.Password, req.RealName, req.Phone, req.Role)
	if err != nil {
		utils.Fail(c, utils.CodeBadRequest, err.Error())
		return
	}
	utils.Success(c, a)
}

// RemoveAdmin 移除管理员
// POST /api/v1/shop/admins/:id/remove
// 安全修复:传入 operatorID 校验操作者是创始人
func (h *ShopHandler) RemoveAdmin(c *gin.Context) {
	adminID := parseInt64Path(c, "id")
	clubID := getClubID(c)
	operatorID := getCurrentUserID(c)
	if err := service.ShopRemoveAdmin(clubID, operatorID, adminID); err != nil {
		utils.Fail(c, utils.CodeBadRequest, err.Error())
		return
	}
	utils.Success(c, gin.H{"msg": "ok"})
}

// resetShopAdminPwdRequest 重置管理员密码请求
type resetShopAdminPwdRequest struct {
	Password string `json:"password" binding:"required"`
}

// ResetAdminPassword 重置管理员密码
// POST /api/v1/shop/admins/:id/reset-password
// 安全修复:传入 operatorID 校验操作者是创始人
func (h *ShopHandler) ResetAdminPassword(c *gin.Context) {
	adminID := parseInt64Path(c, "id")
	var req resetShopAdminPwdRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Fail(c, utils.CodeBadRequest, "参数错误: "+err.Error())
		return
	}
	clubID := getClubID(c)
	operatorID := getCurrentUserID(c)
	if err := service.ShopResetAdminPassword(clubID, operatorID, adminID, req.Password); err != nil {
		utils.Fail(c, utils.CodeBadRequest, err.Error())
		return
	}
	utils.Success(c, gin.H{"msg": "ok"})
}

// GetGroups 俱乐部群聊列表
// GET /api/v1/shop/groups
func (h *ShopHandler) GetGroups(c *gin.Context) {
	clubID := getClubID(c)
	list, err := service.ShopGetGroups(clubID)
	if err != nil {
		utils.Fail(c, utils.CodeServerError, err.Error())
		return
	}
	utils.Success(c, list)
}

// createGroupRequest 创建群聊请求
type createGroupRequest struct {
	GroupName    string `json:"group_name" binding:"required"`
	GroupType    string `json:"group_type"`
	CategoryType string `json:"category_type"`
}

// CreateGroup 创建群聊
// POST /api/v1/shop/groups
func (h *ShopHandler) CreateGroup(c *gin.Context) {
	var req createGroupRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Fail(c, utils.CodeBadRequest, "参数错误: "+err.Error())
		return
	}
	clubID := getClubID(c)
	creatorID := getCurrentUserID(c)
	g, err := service.ShopCreateGroup(clubID, creatorID, req.GroupName, req.GroupType, req.CategoryType)
	if err != nil {
		utils.Fail(c, utils.CodeBadRequest, err.Error())
		return
	}
	utils.Success(c, g)
}

// GetGroupMembers 群成员列表
// GET /api/v1/shop/groups/:id/members
// 安全修复:传入 clubID 校验群归属
func (h *ShopHandler) GetGroupMembers(c *gin.Context) {
	groupID := parseInt64Path(c, "id")
	clubID := getClubID(c)
	list, err := service.ShopGetGroupMembers(clubID, groupID)
	if err != nil {
		utils.Fail(c, utils.CodeServerError, err.Error())
		return
	}
	utils.Success(c, list)
}

// sendGroupMessageRequest 群发消息请求
type sendGroupMessageRequest struct {
	MsgType  string `json:"msg_type"`
	Content  string `json:"content"`
	MediaURL string `json:"media_url"`
}

// SendGroupMessage 群发消息
// POST /api/v1/shop/groups/:id/messages
// 安全修复:传入 clubID 校验群归属
func (h *ShopHandler) SendGroupMessage(c *gin.Context) {
	groupID := parseInt64Path(c, "id")
	var req sendGroupMessageRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Fail(c, utils.CodeBadRequest, "参数错误: "+err.Error())
		return
	}
	clubID := getClubID(c)
	senderID := getCurrentUserID(c)
	m, err := service.ShopSendGroupMessage(clubID, groupID, senderID, req.MsgType, req.Content, req.MediaURL)
	if err != nil {
		utils.Fail(c, utils.CodeBadRequest, err.Error())
		return
	}
	utils.Success(c, m)
}

// publishAnnouncementRequest 发布群公告请求
type publishAnnouncementRequest struct {
	Announcement string `json:"announcement" binding:"required"`
}

// PublishAnnouncement 发布群公告
// PUT /api/v1/shop/groups/:id/announcement
// 安全修复:传入 clubID 校验群归属
func (h *ShopHandler) PublishAnnouncement(c *gin.Context) {
	groupID := parseInt64Path(c, "id")
	var req publishAnnouncementRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Fail(c, utils.CodeBadRequest, "参数错误: "+err.Error())
		return
	}
	clubID := getClubID(c)
	if err := service.ShopPublishAnnouncement(clubID, groupID, req.Announcement); err != nil {
		utils.Fail(c, utils.CodeBadRequest, err.Error())
		return
	}
	utils.Success(c, gin.H{"msg": "ok"})
}

// GetRiskUsers 俱乐部风控用户
// GET /api/v1/shop/risk/users
func (h *ShopHandler) GetRiskUsers(c *gin.Context) {
	clubID := getClubID(c)
	page, pageSize := getPage(c)
	list, total, err := service.ShopGetRiskUsers(clubID, page, pageSize)
	if err != nil {
		utils.Fail(c, utils.CodeServerError, err.Error())
		return
	}
	utils.SuccessWithTotal(c, list, total)
}

// GetRiskOrders 俱乐部风险订单
// GET /api/v1/shop/risk/orders
func (h *ShopHandler) GetRiskOrders(c *gin.Context) {
	clubID := getClubID(c)
	page, pageSize := getPage(c)
	list, total, err := service.ShopGetRiskOrders(clubID, page, pageSize)
	if err != nil {
		utils.Fail(c, utils.CodeServerError, err.Error())
		return
	}
	utils.SuccessWithTotal(c, list, total)
}

// GetOrders 俱乐部订单列表
// GET /api/v1/shop/orders
func (h *ShopHandler) GetOrders(c *gin.Context) {
	clubID := getClubID(c)
	page, pageSize := getPage(c)
	status := parseInt8Query(c, "status", -1)
	keyword := c.Query("keyword")
	list, total, err := service.ShopGetOrders(clubID, page, pageSize, status, keyword)
	if err != nil {
		utils.Fail(c, utils.CodeServerError, err.Error())
		return
	}
	utils.SuccessWithTotal(c, list, total)
}

// GetOrderDetail 俱乐部订单详情
// GET /api/v1/shop/orders/:id
func (h *ShopHandler) GetOrderDetail(c *gin.Context) {
	orderID := parseInt64Path(c, "id")
	clubID := getClubID(c)
	o, err := service.ShopGetOrderDetail(orderID, clubID)
	if err != nil {
		utils.Fail(c, utils.CodeNotFound, err.Error())
		return
	}
	utils.Success(c, o)
}

// GetFailedOrders 俱乐部大额验证失败订单
// GET /api/v1/shop/orders/failed
func (h *ShopHandler) GetFailedOrders(c *gin.Context) {
	clubID := getClubID(c)
	page, pageSize := getPage(c)
	list, total, err := service.ShopGetFailedOrders(clubID, page, pageSize)
	if err != nil {
		utils.Fail(c, utils.CodeServerError, err.Error())
		return
	}
	utils.SuccessWithTotal(c, list, total)
}

// shopUpdateOrderStatusRequest 更新订单状态请求
type shopUpdateOrderStatusRequest struct {
	Status int8   `json:"status" binding:"required"`
	Reason string `json:"reason"`
}

// UpdateOrderStatus 内置管理端更新订单状态
// POST /api/v1/shop/orders/:id/status
func (h *ShopHandler) UpdateOrderStatus(c *gin.Context) {
	orderID := parseInt64Path(c, "id")
	var req shopUpdateOrderStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Fail(c, utils.CodeBadRequest, "参数错误: "+err.Error())
		return
	}
	adminID := getCurrentUserID(c)
	clubID := getClubID(c)
	if err := service.ShopUpdateOrderStatus(orderID, clubID, adminID, req.Status, req.Reason); err != nil {
		utils.Fail(c, utils.CodeBadRequest, err.Error())
		return
	}
	utils.Success(c, gin.H{"msg": "ok"})
}

// shopProcessRefundRequest 退款请求
type shopProcessRefundRequest struct {
	RefundAmount int64 `json:"refund_amount"`
}

// ProcessRefund 内置管理端处理退款
// POST /api/v1/shop/orders/:id/refund
// 安全修复:传入 clubID 校验订单归属,防跨俱乐部退款
func (h *ShopHandler) ProcessRefund(c *gin.Context) {
	orderID := parseInt64Path(c, "id")
	var req shopProcessRefundRequest
	_ = c.ShouldBindJSON(&req)
	adminID := getCurrentUserID(c)
	clubID := getClubID(c)
	p, err := service.ShopProcessRefundWithClub(orderID, adminID, req.RefundAmount, clubID)
	if err != nil {
		utils.Fail(c, utils.CodeBadRequest, err.Error())
		return
	}
	utils.Success(c, p)
}

// GetAfterSaleList 俱乐部售后列表
// GET /api/v1/shop/after-sales
func (h *ShopHandler) GetAfterSaleList(c *gin.Context) {
	clubID := getClubID(c)
	page, pageSize := getPage(c)
	list, total, err := service.ShopGetAfterSaleList(clubID, page, pageSize)
	if err != nil {
		utils.Fail(c, utils.CodeServerError, err.Error())
		return
	}
	utils.SuccessWithTotal(c, list, total)
}

// GetAfterSaleDetail 售后详情
// GET /api/v1/shop/after-sales/:id
func (h *ShopHandler) GetAfterSaleDetail(c *gin.Context) {
	id := parseInt64Path(c, "id")
	clubID := getClubID(c)
	res, err := service.ShopGetAfterSaleDetail(id, clubID)
	if err != nil {
		utils.Fail(c, utils.CodeNotFound, err.Error())
		return
	}
	utils.Success(c, res)
}

// replyAfterSaleRequest 回复售后请求
type replyAfterSaleRequest struct {
	Content  string `json:"content" binding:"required"`
	MediaURL string `json:"media_url"`
}

// ReplyAfterSale 俱乐部回复售后
// POST /api/v1/shop/after-sales/:id/reply
func (h *ShopHandler) ReplyAfterSale(c *gin.Context) {
	id := parseInt64Path(c, "id")
	var req replyAfterSaleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Fail(c, utils.CodeBadRequest, "参数错误: "+err.Error())
		return
	}
	senderID := getCurrentUserID(c)
	if err := service.ShopReplyAfterSale(id, senderID, req.Content, req.MediaURL); err != nil {
		utils.Fail(c, utils.CodeBadRequest, err.Error())
		return
	}
	utils.Success(c, gin.H{"msg": "ok"})
}

// uploadAfterSaleEvidenceRequest 上传售后证据请求
type uploadAfterSaleEvidenceRequest struct {
	MediaURL string `json:"media_url" binding:"required"`
}

// UploadAfterSaleEvidence 上传售后证据
// POST /api/v1/shop/after-sales/:id/evidence
func (h *ShopHandler) UploadAfterSaleEvidence(c *gin.Context) {
	id := parseInt64Path(c, "id")
	var req uploadAfterSaleEvidenceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Fail(c, utils.CodeBadRequest, "参数错误: "+err.Error())
		return
	}
	if err := service.ShopUploadAfterSaleEvidence(id, req.MediaURL); err != nil {
		utils.Fail(c, utils.CodeBadRequest, err.Error())
		return
	}
	utils.Success(c, gin.H{"msg": "ok"})
}

// GetWithdrawals 俱乐部提现记录
// GET /api/v1/shop/withdrawals
func (h *ShopHandler) GetWithdrawals(c *gin.Context) {
	clubID := getClubID(c)
	page, pageSize := getPage(c)
	list, total, err := service.ShopGetWithdrawals(clubID, page, pageSize)
	if err != nil {
		utils.Fail(c, utils.CodeServerError, err.Error())
		return
	}
	utils.SuccessWithTotal(c, list, total)
}

// GetFinanceOverview 俱乐部财务概览
// GET /api/v1/shop/finance/overview
func (h *ShopHandler) GetFinanceOverview(c *gin.Context) {
	clubID := getClubID(c)
	res, err := service.ShopGetFinanceOverview(clubID)
	if err != nil {
		utils.Fail(c, utils.CodeServerError, err.Error())
		return
	}
	utils.Success(c, res)
}

// GetFinanceDetails 俱乐部财务明细
// GET /api/v1/shop/finance/details
func (h *ShopHandler) GetFinanceDetails(c *gin.Context) {
	clubID := getClubID(c)
	page, pageSize := getPage(c)
	list, total, err := service.ShopGetFinanceDetails(clubID, page, pageSize)
	if err != nil {
		utils.Fail(c, utils.CodeServerError, err.Error())
		return
	}
	utils.SuccessWithTotal(c, list, total)
}

// GetInviteCodes 俱乐部邀请码列表
// GET /api/v1/shop/invite-codes
func (h *ShopHandler) GetInviteCodes(c *gin.Context) {
	clubID := getClubID(c)
	page, pageSize := getPage(c)
	list, total, err := service.ShopGetInviteCodes(clubID, page, pageSize)
	if err != nil {
		utils.Fail(c, utils.CodeServerError, err.Error())
		return
	}
	utils.SuccessWithTotal(c, list, total)
}

// shopGenerateInviteCodeRequest 生成邀请码请求
type shopGenerateInviteCodeRequest struct {
	Role       string `json:"role"`
	MaxUses    int    `json:"max_uses"`
	ExpireDays int    `json:"expire_days"`
}

// GenerateInviteCode 俱乐部生成邀请码(创始人专属)
// POST /api/v1/shop/invite-codes
func (h *ShopHandler) GenerateInviteCode(c *gin.Context) {
	var req shopGenerateInviteCodeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Fail(c, utils.CodeBadRequest, "参数错误: "+err.Error())
		return
	}
	clubID := getClubID(c)
	creatorID := getCurrentUserID(c)
	code, err := service.ShopGenerateInviteCode(clubID, creatorID, req.Role, req.MaxUses, req.ExpireDays)
	if err != nil {
		utils.Fail(c, utils.CodeBadRequest, err.Error())
		return
	}
	utils.Success(c, code)
}

// RevokeInviteCode 俱乐部撤销邀请码
// POST /api/v1/shop/invite-codes/:id/revoke
func (h *ShopHandler) RevokeInviteCode(c *gin.Context) {
	id := parseInt64Path(c, "id")
	clubID := getClubID(c)
	if err := service.ShopRevokeInviteCode(id, clubID); err != nil {
		utils.Fail(c, utils.CodeBadRequest, err.Error())
		return
	}
	utils.Success(c, gin.H{"msg": "ok"})
}

// UpdateCommissionRate 更新创始人抽成比例
// PUT /api/v1/shop/club/commission
func (h *ShopHandler) UpdateCommissionRate(c *gin.Context) {
	clubID := getClubID(c)
	var req struct {
		Rate int8 `json:"rate"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Fail(c, utils.CodeBadRequest, "参数错误: "+err.Error())
		return
	}
	if err := service.ShopUpdateCommissionRate(clubID, req.Rate); err != nil {
		utils.Fail(c, utils.CodeBadRequest, err.Error())
		return
	}
	utils.Success(c, gin.H{"msg": "ok"})
}

// CreateFineRule 创建罚款规则
// POST /api/v1/shop/fine-rules
func (h *ShopHandler) CreateFineRule(c *gin.Context) {
	clubID := getClubID(c)
	creatorID := getCurrentUserID(c)
	var req struct {
		Name        string `json:"name"`
		Description string `json:"description"`
		Amount      int64  `json:"amount"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Fail(c, utils.CodeBadRequest, "参数错误: "+err.Error())
		return
	}
	rule, err := service.ShopCreateFineRule(clubID, creatorID, req.Name, req.Description, req.Amount)
	if err != nil {
		utils.Fail(c, utils.CodeBadRequest, err.Error())
		return
	}
	utils.Success(c, rule)
}

// ListFineRules 罚款规则列表
// GET /api/v1/shop/fine-rules
func (h *ShopHandler) ListFineRules(c *gin.Context) {
	clubID := getClubID(c)
	status := c.Query("status")
	list, err := service.ShopListFineRules(clubID, status)
	if err != nil {
		utils.Fail(c, utils.CodeServerError, err.Error())
		return
	}
	utils.Success(c, list)
}

// RevokeFineRule 下架罚款规则
// POST /api/v1/shop/fine-rules/:id/revoke
func (h *ShopHandler) RevokeFineRule(c *gin.Context) {
	ruleID := parseInt64Path(c, "id")
	clubID := getClubID(c)
	if err := service.ShopRevokeFineRule(clubID, ruleID); err != nil {
		utils.Fail(c, utils.CodeBadRequest, err.Error())
		return
	}
	utils.Success(c, gin.H{"msg": "ok"})
}

// GetAnnouncementReadStats 公告已读统计
// GET /api/v1/shop/groups/:id/announcement/stats
func (h *ShopHandler) GetAnnouncementReadStats(c *gin.Context) {
	groupID := parseInt64Path(c, "id")
	stats, err := service.ShopGetAnnouncementReadStats(groupID)
	if err != nil {
		utils.Fail(c, utils.CodeServerError, err.Error())
		return
	}
	utils.Success(c, stats)
}

// MarkAnnouncementRead 标记公告已读
// POST /api/v1/shop/groups/:id/announcement/read
func (h *ShopHandler) MarkAnnouncementRead(c *gin.Context) {
	groupID := parseInt64Path(c, "id")
	userID := getCurrentUserID(c)
	if err := service.ShopMarkAnnouncementRead(groupID, userID); err != nil {
		utils.Fail(c, utils.CodeServerError, err.Error())
		return
	}
	utils.Success(c, gin.H{"msg": "ok"})
}
