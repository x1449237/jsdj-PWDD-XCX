package handler

import (
	"github.com/gin-gonic/gin"
	"github.com/jisan/e-sports-platform/internal/service"
	"github.com/jisan/e-sports-platform/internal/utils"
)

// AdminFinanceHandler 平台财务处理器(提现/退款/保证金)
type AdminFinanceHandler struct{}

// NewAdminFinanceHandler 创建平台财务处理器
func NewAdminFinanceHandler() *AdminFinanceHandler { return &AdminFinanceHandler{} }

// GetWithdrawals 平台提现记录列表
// GET /api/v1/admin/withdrawals
func (h *AdminFinanceHandler) GetWithdrawals(c *gin.Context) {
	page, pageSize := getPage(c)
	status := c.Query("status")
	list, total, err := service.AdminGetWithdrawals(page, pageSize, status)
	if err != nil {
		utils.Fail(c, utils.CodeServerError, err.Error())
		return
	}
	utils.SuccessWithTotal(c, list, total)
}

// ApproveWithdrawal 审核通过提现
// POST /api/v1/admin/withdrawals/:id/approve
func (h *AdminFinanceHandler) ApproveWithdrawal(c *gin.Context) {
	id := parseInt64Path(c, "id")
	adminID := getCurrentUserID(c)
	if err := service.AdminApproveWithdrawal(id, adminID); err != nil {
		utils.Fail(c, utils.CodeBadRequest, err.Error())
		return
	}
	utils.Success(c, gin.H{"msg": "ok"})
}

// rejectWithdrawalRequest 驳回提现请求
type rejectWithdrawalRequest struct {
	Reason string `json:"reason"`
}

// RejectWithdrawal 驳回提现
// POST /api/v1/admin/withdrawals/:id/reject
func (h *AdminFinanceHandler) RejectWithdrawal(c *gin.Context) {
	id := parseInt64Path(c, "id")
	var req rejectWithdrawalRequest
	_ = c.ShouldBindJSON(&req)
	adminID := getCurrentUserID(c)
	if err := service.AdminRejectWithdrawal(id, adminID, req.Reason); err != nil {
		utils.Fail(c, utils.CodeBadRequest, err.Error())
		return
	}
	utils.Success(c, gin.H{"msg": "ok"})
}

// batchWithdrawRequest 批量提现请求
type batchWithdrawRequest struct {
	IDs    []int64 `json:"ids" binding:"required"`
	Action string  `json:"action" binding:"required"`
}

// BatchWithdraw 批量提现处理
// POST /api/v1/admin/withdrawals/batch
func (h *AdminFinanceHandler) BatchWithdraw(c *gin.Context) {
	var req batchWithdrawRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Fail(c, utils.CodeBadRequest, "参数错误: "+err.Error())
		return
	}
	adminID := getCurrentUserID(c)
	success, err := service.AdminBatchWithdraw(adminID, req.IDs, req.Action)
	if err != nil {
		utils.Fail(c, utils.CodeBadRequest, err.Error())
		return
	}
	utils.Success(c, gin.H{"success": success})
}

// processRefundRequest 处理退款请求
type processRefundRequest struct {
	RefundAmount int64 `json:"refund_amount"`
}

// ProcessRefund 平台处理退款
// POST /api/v1/admin/orders/:id/refund
func (h *AdminFinanceHandler) ProcessRefund(c *gin.Context) {
	orderID := parseInt64Path(c, "id")
	var req processRefundRequest
	_ = c.ShouldBindJSON(&req)
	adminID := getCurrentUserID(c)
	p, err := service.AdminProcessRefund(orderID, adminID, req.RefundAmount)
	if err != nil {
		utils.Fail(c, utils.CodeBadRequest, err.Error())
		return
	}
	utils.Success(c, p)
}

// GetDeposits 保证金列表
// GET /api/v1/admin/deposits
func (h *AdminFinanceHandler) GetDeposits(c *gin.Context) {
	page, pageSize := getPage(c)
	list, total, err := service.AdminGetDeposits(page, pageSize)
	if err != nil {
		utils.Fail(c, utils.CodeServerError, err.Error())
		return
	}
	utils.SuccessWithTotal(c, list, total)
}

// ConfirmDeposit 确认保证金缴纳
// POST /api/v1/admin/deposits/:club_id/confirm
func (h *AdminFinanceHandler) ConfirmDeposit(c *gin.Context) {
	clubID := parseInt64Path(c, "club_id")
	if err := service.AdminConfirmDeposit(clubID); err != nil {
		utils.Fail(c, utils.CodeBadRequest, err.Error())
		return
	}
	utils.Success(c, gin.H{"msg": "ok"})
}

// RefundDeposit 退还保证金
// POST /api/v1/admin/deposits/:club_id/refund
func (h *AdminFinanceHandler) RefundDeposit(c *gin.Context) {
	clubID := parseInt64Path(c, "club_id")
	if err := service.AdminRefundDeposit(clubID); err != nil {
		utils.Fail(c, utils.CodeBadRequest, err.Error())
		return
	}
	utils.Success(c, gin.H{"msg": "ok"})
}

// updateDepositConfigRequest 更新保证金配置请求
type updateDepositConfigRequest struct {
	Amount int64 `json:"amount" binding:"required"`
}

// UpdateDepositConfig 更新保证金配置
// PUT /api/v1/admin/deposits/config
func (h *AdminFinanceHandler) UpdateDepositConfig(c *gin.Context) {
	var req updateDepositConfigRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Fail(c, utils.CodeBadRequest, "参数错误: "+err.Error())
		return
	}
	if err := service.AdminUpdateDepositConfig(req.Amount); err != nil {
		utils.Fail(c, utils.CodeBadRequest, err.Error())
		return
	}
	utils.Success(c, gin.H{"msg": "ok"})
}

// GetProfitShare 订单分账明细
// GET /api/v1/admin/orders/:id/profit-share
func (h *AdminFinanceHandler) GetProfitShare(c *gin.Context) {
	orderID := parseInt64Path(c, "id")
	list, err := service.ListProfitShareByOrder(orderID)
	if err != nil {
		utils.Fail(c, utils.CodeServerError, err.Error())
		return
	}
	utils.Success(c, list)
}

// SettleOrder 结算订单分账
// POST /api/v1/admin/orders/:id/settle
func (h *AdminFinanceHandler) SettleOrder(c *gin.Context) {
	orderID := parseInt64Path(c, "id")
	if err := service.SettleOrderProfitShare(orderID); err != nil {
		utils.Fail(c, utils.CodeBadRequest, err.Error())
		return
	}
	utils.Success(c, gin.H{"msg": "ok"})
}

// ================ 对公小额打款验证(板块三) ================

// generateTransferRequest 生成对公打款验证请求
type generateTransferRequest struct {
	ClubID      int64  `json:"club_id" binding:"required"`
	BankName    string `json:"bank_name" binding:"required"`
	BankAccount string `json:"bank_account" binding:"required"`
	AccountName string `json:"account_name" binding:"required"`
}

// GenerateTransfer 生成对公小额打款验证
// POST /api/v1/admin/corporate-transfers
// 生成 0.0-0.9 之间的随机验证金额，有效期 48h
func (h *AdminFinanceHandler) GenerateTransfer(c *gin.Context) {
	var req generateTransferRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Fail(c, utils.CodeBadRequest, "参数错误: "+err.Error())
		return
	}
	ctv, err := service.GenerateCorporateTransfer(req.ClubID, req.BankName, req.BankAccount, req.AccountName)
	if err != nil {
		utils.Fail(c, utils.CodeBadRequest, err.Error())
		return
	}
	// 对公账户脱敏后返回(避免明文回显)
	resp := gin.H{
		"id":            ctv.ID,
		"club_id":       ctv.ClubID,
		"bank_name":     ctv.BankName,
		"bank_account":  utils.MaskBankAccount(ctv.BankAccount),
		"account_name":  ctv.AccountName,
		"verify_amount": ctv.VerifyAmount,
		"generated_at":  ctv.GeneratedAt,
		"expire_at":     ctv.ExpireAt,
		"verify_count":  ctv.VerifyCount,
		"status":        ctv.Status,
		"locked_until":  ctv.LockedUntil,
	}
	utils.Success(c, resp)
}

// verifyTransferRequest 确认对公打款到账请求
type verifyTransferRequest struct {
	ActualAmount string `json:"actual_amount" binding:"required"`
}

// VerifyTransfer 确认对公打款到账(金额比对)
// POST /api/v1/admin/corporate-transfers/:id/verify
// 失败累计 verify_count++，达 5 次锁定 15 天
func (h *AdminFinanceHandler) VerifyTransfer(c *gin.Context) {
	id := parseInt64Path(c, "id")
	var req verifyTransferRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Fail(c, utils.CodeBadRequest, "参数错误: "+err.Error())
		return
	}
	ctv, err := service.VerifyCorporateTransfer(id, req.ActualAmount)
	if err != nil {
		utils.Fail(c, utils.CodeBadRequest, err.Error())
		return
	}
	utils.Success(c, gin.H{
		"id":            ctv.ID,
		"club_id":       ctv.ClubID,
		"status":        ctv.Status,
		"verify_count":  ctv.VerifyCount,
		"locked_until":  ctv.LockedUntil,
	})
}

// TransferList 对公打款台账列表
// GET /api/v1/admin/corporate-transfers
// 支持 club_id 与 status 过滤
func (h *AdminFinanceHandler) TransferList(c *gin.Context) {
	page, pageSize := getPage(c)
	clubID := parseInt64Query(c, "club_id", 0)
	status := c.Query("status")
	list, total, err := service.ListCorporateTransfers(page, pageSize, clubID, status)
	if err != nil {
		utils.Fail(c, utils.CodeServerError, err.Error())
		return
	}
	// 脱敏对公账户后返回
	for i := range list {
		if list[i].BankAccount != "" {
			list[i].BankAccount = utils.MaskBankAccount(list[i].BankAccount)
		}
	}
	utils.SuccessWithTotal(c, list, total)
}
