package handler

import (
	"github.com/gin-gonic/gin"
	"github.com/jisan/e-sports-platform/internal/service"
	"github.com/jisan/e-sports-platform/internal/utils"
)

// ThinClientHandler 小程序前端零逻辑配套：校验/计算/配置 API
type ThinClientHandler struct{}

func NewThinClientHandler() *ThinClientHandler { return &ThinClientHandler{} }

// ========== 身份证校验 + 年龄计算 ==========

// ValidateIDCard POST /thin/id-card/validate (前端只传 id_card, 后端返回校验结果+年龄)
func (h *ThinClientHandler) ValidateIDCard(c *gin.Context) {
	var req struct {
		IDCard string `json:"id_card"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Fail(c, utils.CodeBadRequest, "参数错误"); return
	}
	result := service.ValidateIDCard(req.IDCard)
	utils.Success(c, result)
}

// ========== 手机号校验 ==========

// ValidatePhone POST /thin/phone/validate
func (h *ThinClientHandler) ValidatePhone(c *gin.Context) {
	var req struct {
		Phone string `json:"phone"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Fail(c, utils.CodeBadRequest, "参数错误"); return
	}
	utils.Success(c, gin.H{"valid": service.ValidatePhone(req.Phone)})
}

// ========== 密码强度校验 ==========

// CalcPasswordStrength POST /thin/password/strength
func (h *ThinClientHandler) CalcPasswordStrength(c *gin.Context) {
	var req struct {
		Password string `json:"password"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Fail(c, utils.CodeBadRequest, "参数错误"); return
	}
	result := service.CalcPasswordStrength(req.Password)
	utils.Success(c, result)
}

// ========== 金额校验 ==========

// ValidateAmount POST /thin/amount/validate
func (h *ThinClientHandler) ValidateAmount(c *gin.Context) {
	var req struct {
		Amount string `json:"amount"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Fail(c, utils.CodeBadRequest, "参数错误"); return
	}
	if err := service.ValidateAmount(req.Amount); err != nil {
		utils.Success(c, gin.H{"valid": false, "message": err.Error()})
		return
	}
	utils.Success(c, gin.H{"valid": true, "message": "金额校验通过"})
}

// ========== 金额换算(前端传元字符串,后端返回分) ==========

// YuanToFen POST /thin/amount/yuan-to-fen
func (h *ThinClientHandler) YuanToFen(c *gin.Context) {
	var req struct {
		Amount string `json:"amount"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Fail(c, utils.CodeBadRequest, "参数错误"); return
	}
	fen, err := service.ParseYuanToFen(req.Amount)
	if err != nil {
		utils.Fail(c, utils.CodeBadRequest, err.Error()); return
	}
	utils.Success(c, gin.H{"fen": fen, "yuan_text": service.FormatFenToYuan(fen)})
}

// ========== 折扣计算 ==========

// CalcDiscount POST /thin/discount/calc
func (h *ThinClientHandler) CalcDiscount(c *gin.Context) {
	var req struct {
		OriginalPriceFen int64 `json:"original_price_fen"`
		GroupPriceFen    int64 `json:"group_price_fen"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Fail(c, utils.CodeBadRequest, "参数错误"); return
	}
	utils.Success(c, service.CalcDiscount(req.OriginalPriceFen, req.GroupPriceFen))
}

// ========== 未成年人宵禁校验 ==========

// MinorCurfewCheck GET /thin/minor/curfew-check?is_minor=1
func (h *ThinClientHandler) MinorCurfewCheck(c *gin.Context) {
	isMinor := c.Query("is_minor") == "1"
	blocked, msg := service.MinorCurfewCheck(isMinor)
	utils.Success(c, gin.H{"blocked": blocked, "message": msg})
}

// ========== 文件大小校验 ==========

// ValidateFileSize POST /thin/file/validate-size
func (h *ThinClientHandler) ValidateFileSize(c *gin.Context) {
	var req struct {
		SizeBytes int64 `json:"size_bytes"`
		MaxSizeMB int   `json:"max_size_mb"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Fail(c, utils.CodeBadRequest, "参数错误"); return
	}
	if err := service.ValidateFileSize(req.SizeBytes, req.MaxSizeMB); err != nil {
		utils.Success(c, gin.H{"valid": false, "message": err.Error()})
		return
	}
	utils.Success(c, gin.H{"valid": true})
}

// ========== 消息撤回时效校验 ==========

// CanRecallMessage POST /thin/message/can-recall
func (h *ThinClientHandler) CanRecallMessage(c *gin.Context) {
	var req struct {
		MsgTimeUnix int64 `json:"msg_time_unix"`
		LimitSec    int   `json:"limit_sec"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Fail(c, utils.CodeBadRequest, "参数错误"); return
	}
	msgTime := unixToTime(req.MsgTimeUnix)
	canRecall := service.CanRecallMessage(msgTime, req.LimitSec)
	utils.Success(c, gin.H{"can_recall": canRecall})
}

// ========== 脱敏接口(后端返回已脱敏数据) ==========

// MaskData POST /thin/mask
func (h *ThinClientHandler) MaskData(c *gin.Context) {
	var req struct {
		Type  string `json:"type"`  // phone/id_card/name/bank_account
		Value string `json:"value"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Fail(c, utils.CodeBadRequest, "参数错误"); return
	}
	masked := req.Value
	switch req.Type {
	case "phone":
		masked = service.MaskPhone(req.Value)
	case "id_card":
		masked = service.MaskIDCard(req.Value)
	case "name":
		masked = service.MaskName(req.Value)
	case "bank_account":
		masked = service.MaskBankAccount(req.Value)
	}
	utils.Success(c, gin.H{"masked": masked})
}

// ========== 通用配置下发(标签/金额选项/状态列表等) ==========

// GetServiceTags GET /thin/config/service-tags (评价标签等)
func (h *ThinClientHandler) GetServiceTags(c *gin.Context) {
	tags := []string{"技术过硬", "耐心指导", "服务周到", "响应迅速", "性价比高", "沟通顺畅", "守时靠谱", "态度友好", "专业精湛"}
	utils.Success(c, gin.H{"tags": tags, "max_select": 5})
}

// GetRewardPresets GET /thin/config/reward-presets (打赏金额选项)
func (h *ThinClientHandler) GetRewardPresets(c *gin.Context) {
	utils.Success(c, gin.H{
		"preset_amounts": []int{5, 10, 18, 28, 50, 88},
		"min_amount":     1,
		"max_amount":     200,
	})
}

// GetOrderTabs GET /thin/config/order-tabs (订单Tab配置)
func (h *ThinClientHandler) GetOrderTabs(c *gin.Context) {
	tabs := []gin.H{
		{"label": "全部", "status": -1},
		{"label": "待接单", "status": 0},
		{"label": "服务中", "status": 1},
		{"label": "待确认", "status": 2},
		{"label": "已完结", "status": 3},
	}
	utils.Success(c, gin.H{"tabs": tabs})
}

// GetAppealConfig GET /thin/config/appeal (申诉类型+状态配置)
func (h *ThinClientHandler) GetAppealConfig(c *gin.Context) {
	utils.Success(c, gin.H{
		"types": []gin.H{
			{"key": "phone", "label": "手机号申诉"},
			{"key": "order", "label": "订单申诉"},
		},
		"statuses": []gin.H{
			{"key": 0, "label": "待处理", "color": "#FF9900"},
			{"key": 1, "label": "处理中", "color": "#00AAFF"},
			{"key": 2, "label": "已处理", "color": "#00CC66"},
			{"key": 3, "label": "已驳回", "color": "#999999"},
		},
	})
}

// GetGroupTypes GET /thin/config/group-types (群类型配置)
func (h *ThinClientHandler) GetGroupTypes(c *gin.Context) {
	utils.Success(c, gin.H{
		"types": []gin.H{
			{"key": "chat", "label": "闲聊群"},
			{"key": "welfare", "label": "福利群"},
			{"key": "after_sale", "label": "售后群"},
		},
	})
}

// GetServiceTypes GET /thin/config/service-types (服务类型配置)
func (h *ThinClientHandler) GetServiceTypes(c *gin.Context) {
	utils.Success(c, gin.H{
		"types": []gin.H{
			{"key": 1, "label": "排位"},
			{"key": 2, "label": "匹配"},
			{"key": 3, "label": "陪玩"},
			{"key": 4, "label": "上分"},
			{"key": 5, "label": "教学"},
			{"key": 6, "label": "代练"},
		},
	})
}

// GetInterveneConfig GET /thin/config/intervene (售后介入状态配置)
func (h *ThinClientHandler) GetInterveneConfig(c *gin.Context) {
	utils.Success(c, gin.H{
		"statuses": []gin.H{
			{"key": 0, "label": "未介入", "class": "status-normal"},
			{"key": 1, "label": "介入中", "class": "status-active"},
			{"key": 2, "label": "已解除", "class": "status-resolved"},
		},
	})
}

// GetDisputeTypes GET /thin/config/dispute-types (纠纷类型配置)
func (h *ThinClientHandler) GetDisputeTypes(c *gin.Context) {
	utils.Success(c, gin.H{
		"types": []gin.H{
			{"key": "quality", "label": "服务质量纠纷"},
			{"key": "payment", "label": "费用纠纷"},
			{"key": "attitude", "label": "服务态度纠纷"},
			{"key": "other", "label": "其他纠纷"},
		},
	})
}
