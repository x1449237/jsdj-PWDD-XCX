package handler

import (
	"encoding/json"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/jisan/e-sports-platform/internal/model"
	"github.com/jisan/e-sports-platform/internal/service"
	"github.com/jisan/e-sports-platform/internal/utils"
)

// Platform99582Handler 处理 99~582 需求扩展的全部接口
type Platform99582Handler struct{}

func NewPlatform99582Handler() *Platform99582Handler { return &Platform99582Handler{} }

// ========== 入驻总开关 & 缩写说明页 ==========

// GetClubJoinSwitch GET /platform/join-switch (前端查询, 公开 216/433/650)
func (h *Platform99582Handler) GetClubJoinSwitch(c *gin.Context) {
	enabled, err := service.GetClubJoinSwitch()
	if err != nil {utils.Fail(c, utils.CodeServerError, err.Error()); return}
	utils.Success(c, gin.H{"enabled": enabled})
}

// SetClubJoinSwitch POST /admin/join-switch (仅Web超管)
func (h *Platform99582Handler) SetClubJoinSwitch(c *gin.Context) {
	if !isPlatformAdmin(c) {utils.Fail(c, utils.CodeForbidden, "仅平台管理员"); return}
	var req struct {
		Enabled int8 `json:"enabled"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {utils.Fail(c, utils.CodeBadRequest, "参数错误"); return}
	if err := service.SetClubJoinSwitch(req.Enabled, getCurrentUserID(c)); err != nil {
		utils.Fail(c, utils.CodeServerError, err.Error()); return
	}
	utils.Success(c, gin.H{"msg": "ok"})
}

// GetAbbrHelpPage GET /public/abbr-help 缩写说明帮助页(需求226)
func (h *Platform99582Handler) GetAbbrHelpPage(c *gin.Context) {
	utils.Success(c, gin.H{
		"sections": []map[string]interface{}{
			{"title": "一、订单号生成规则说明", "body": "订单号格式：俱乐部缩写 - 年月日时分（24小时制） - 俱乐部内部当日自增序号。例如 BCYL-260721222138，yy=年份后两位，mm=月份，dd=日期，hh=小时，mi=分钟。"},
			{"title": "二、俱乐部缩写唯一性约束", "body": "平台所有俱乐部缩写全局唯一，相同中文俱乐部名称生成的缩写会产生冲突。"},
			{"title": "三、为什么缩写无法重复占用", "body": "为保证订单号全局唯一、结算台账可追溯，缩写一经成功创建即不可被第二家俱乐部重复占用。"},
			{"title": "小提示", "body": "修改俱乐部中文名称后再提交，系统将重新生成缩写可规避重名；也可在原名前后加上简短前缀或后缀（如地名、行业词等）。", "small": true},
		},
	})
}

// ========== 订单号生成接口 ==========

// GenerateClubOrderNo POST /orders/gen-club-order-no
// 前端在建单或转单时调用，返回唯一订单号 (需求262/263/479/480)
func (h *Platform99582Handler) GenerateClubOrderNo(c *gin.Context) {
	var req struct {
		ClubID int64  `json:"club_id"`
		Abbr   string `json:"abbr"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {utils.Fail(c, utils.CodeBadRequest, "参数错误"); return}
	no, err := service.GenerateClubOrderNo(req.ClubID, req.Abbr)
	if err != nil {utils.Fail(c, utils.CodeServerError, err.Error()); return}
	utils.Success(c, gin.H{"order_no": no})
}

// ========== 个人/企业入驻资料接口(二进制存MySQL 233/249) ==========

// UpsertPersonalRegFiles POST /registrations/personal/files (需求220~235)
func (h *Platform99582Handler) UpsertPersonalRegFiles(c *gin.Context) {
	uid := getCurrentUserID(c)
	var req struct {
		RegID        int64  `json:"reg_id"`
		Name         string `json:"applicant_name"`
		IDCard       string `json:"id_card_no"`
		Address      string `json:"contact_address"`
		IDFrontBin   []byte `json:"id_card_front_base64"`
		IDBackBin    []byte `json:"id_card_back_base64"`
		ContractPDF  []byte `json:"contract_signed_pdf_base64"`
		AgeCheck     int8   `json:"applicant_age_check"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {utils.Fail(c, utils.CodeBadRequest, "参数错误"); return}
	_ = uid
	if req.AgeCheck != 1 {utils.Fail(c, utils.CodeBadRequest, "申请人必须年满16周岁"); return}
	f := &model.PersonalClubRegistrationFiles{
		PersonalRegID: req.RegID, ApplicantName: req.Name, IDCardNo: req.IDCard,
		ContactAddress: req.Address, IDCardFrontBin: req.IDFrontBin, IDCardBackBin: req.IDBackBin,
		ContractSignedPDF: req.ContractPDF, ApplicantAgeCheck: req.AgeCheck,
	}
	if err := service.UpsertPersonalRegFiles(f); err != nil {
		utils.Fail(c, utils.CodeServerError, err.Error()); return
	}
	utils.Success(c, gin.H{"msg": "ok"})
}

// UpsertEnterpriseRegFiles POST /registrations/enterprise/files (236~255)
func (h *Platform99582Handler) UpsertEnterpriseRegFiles(c *gin.Context) {
	var req struct {
		RegID           int64  `json:"reg_id"`
		LicenseBin      []byte `json:"business_license_bin"`
		FaceVerified    int8   `json:"legal_person_face_verified"`
		AgentAuthPDF    []byte `json:"agent_auth_contract_pdf_bin"`
		ContractSigned  []byte `json:"enterprise_contract_signed_pdf_bin"`
		BankName        string `json:"bank_account_name"`
		BankCard        string `json:"bank_card_no"`
		RandomAmount    float64 `json:"random_amount"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {utils.Fail(c, utils.CodeBadRequest, "参数错误"); return}
	f := &model.EnterpriseClubRegistrationFiles{
		EnterpriseRegID: req.RegID, BusinessLicenseBin: req.LicenseBin,
		LegalPersonFaceVerified: req.FaceVerified,
		AgentAuthContractPDFBin: req.AgentAuthPDF, EnterpriseContractSignedPDFBin: req.ContractSigned,
		BankAccountName: req.BankName, BankCardNo: req.BankCard, RandomAmount: req.RandomAmount,
	}
	if err := service.UpsertEnterpriseRegFiles(f); err != nil {
		utils.Fail(c, utils.CodeServerError, err.Error()); return
	}
	utils.Success(c, gin.H{"msg": "ok"})
}

// GenerateRandomVerifyAmount GET /admin/enterprise/verify-amount/random (Web超管或前端拉取随机金额 240)
func (h *Platform99582Handler) GenerateRandomVerifyAmount(c *gin.Context) {
	amt, err := service.GenerateEnterpriseRandomAmount()
	if err != nil {utils.Fail(c, utils.CodeServerError, err.Error()); return}
	utils.Success(c, gin.H{"random_amount": amt})
}

// ========== 全局参数设置(Web超管 406~460) ==========

// GetGlobalParam GET /admin/global-param
func (h *Platform99582Handler) GetGlobalParam(c *gin.Context) {
	key := c.Query("param_key")
	def := c.Query("default_value")
	v, err := service.GetGlobalParam(key, def)
	if err != nil {utils.Fail(c, utils.CodeServerError, err.Error()); return}
	utils.Success(c, gin.H{"value": v})
}

// SetGlobalParam POST /admin/global-param
func (h *Platform99582Handler) SetGlobalParam(c *gin.Context) {
	if !isPlatformAdmin(c) {utils.Fail(c, utils.CodeForbidden, "仅平台管理员"); return}
	var req struct {
		Key         string `json:"param_key"`
		Value       string `json:"param_value"`
		Type        string `json:"param_type"`
		Module      string `json:"module"`
		Description string `json:"description"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {utils.Fail(c, utils.CodeBadRequest, "参数错误"); return}
	if err := service.SetGlobalParam(req.Key, req.Value, req.Type, req.Module, req.Description); err != nil {
		utils.Fail(c, utils.CodeServerError, err.Error()); return
	}
	utils.Success(c, gin.H{"msg": "ok"})
}

// ========== 隐形水印导出 & 区块链存证 ==========

// CreateExportWatermarkLog POST /export/watermark-log  (需求461~474)
func (h *Platform99582Handler) CreateExportWatermarkLog(c *gin.Context) {
	var log model.ExportWatermarkLog
	if err := c.ShouldBindJSON(&log); err != nil {utils.Fail(c, utils.CodeBadRequest, "参数错误: "+err.Error()); return}
	// 默认写当前请求人(如未传)
	if log.ExporterUID <= 0 {log.ExporterUID = getCurrentUserID(c)}
	if err := service.BuildExportWatermarkLog(&log); err != nil {
		utils.Fail(c, utils.CodeServerError, err.Error()); return
	}
	utils.Success(c, gin.H{
		"export_no": log.ExportNo, "origin_hash": log.OriginHashSHA256,
		"blockchain_txid": log.BlockchainTxid, "blockchain_ts": log.BlockchainTimestamp,
	})
}

// QueryExportWatermark GET /export/watermark-log/:export_no
func (h *Platform99582Handler) QueryExportWatermark(c *gin.Context) {
	no := c.Param("export_no")
	log, err := service.QueryExportWatermarkByOrderNo(no)
	if err != nil {utils.Fail(c, utils.CodeServerError, err.Error()); return}
	utils.Success(c, log)
}

// InsertDetectLog POST /watermark/detect-log  (需求468~485)
func (h *Platform99582Handler) InsertDetectLog(c *gin.Context) {
	var d model.WatermarkDetectLog
	if err := c.ShouldBindJSON(&d); err != nil {utils.Fail(c, utils.CodeBadRequest, "参数错误"); return}
	if d.OperatorUID <= 0 {d.OperatorUID = getCurrentUserID(c)}
	if err := service.InsertDetectLog(&d); err != nil {
		utils.Fail(c, utils.CodeServerError, err.Error()); return
	}
	utils.Success(c, gin.H{"msg": "ok", "id": d.ID})
}

// ========== 个性化偏好(需求99/103/105/107/110, 322~331) ==========

// GetUserIMPreference GET /im/preference
func (h *Platform99582Handler) GetUserIMPreference(c *gin.Context) {
	p, err := service.GetUserIMPreference(getCurrentUserID(c))
	if err != nil {utils.Fail(c, utils.CodeServerError, err.Error()); return}
	utils.Success(c, p)
}

// SaveUserIMPreference POST /im/preference
func (h *Platform99582Handler) SaveUserIMPreference(c *gin.Context) {
	var p model.IMUserPreference
	if err := c.ShouldBindJSON(&p); err != nil {utils.Fail(c, utils.CodeBadRequest, "参数错误"); return}
	p.UID = getCurrentUserID(c)
	if err := service.SaveUserIMPreference(&p); err != nil {
		utils.Fail(c, utils.CodeServerError, err.Error()); return
	}
	utils.Success(c, gin.H{"msg": "ok"})
}

// ========== 正在输入状态(340 仅文字触发/语音不触发) ==========

// UpsertTypingStatus POST /im/typing (typing_type: text/voice)
func (h *Platform99582Handler) UpsertTypingStatus(c *gin.Context) {
	var req struct {
		SessionID  int64  `json:"session_id"`
		TypingType string `json:"typing_type"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {utils.Fail(c, utils.CodeBadRequest, "参数错误"); return}
	if err := service.UpsertTypingStatus(req.SessionID, getCurrentUserID(c), req.TypingType); err != nil {
		utils.Fail(c, utils.CodeServerError, err.Error()); return
	}
	utils.Success(c, gin.H{"msg": "ok"})
}

// GetTypingUsers GET /im/:session_id/typing
func (h *Platform99582Handler) GetTypingUsers(c *gin.Context) {
	sid := parseInt64Path(c, "session_id")
	list, err := service.GetSessionTypingUsers(sid)
	if err != nil {utils.Fail(c, utils.CodeServerError, err.Error()); return}
	utils.Success(c, gin.H{"uids": list})
}

// ========== 统一收藏(127/344/165/782) ==========

// ToggleFavorite POST /favorites
func (h *Platform99582Handler) ToggleFavorite(c *gin.Context) {
	var req struct {
		FavType    string          `json:"fav_type"` // club/player/message/transfer_card
		TargetID   int64           `json:"target_id"`
		ExtraJSON  json.RawMessage `json:"extra_data_json"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {utils.Fail(c, utils.CodeBadRequest, "参数错误"); return}
	isFav, err := service.ToggleUserFavorite(getCurrentUserID(c), req.FavType, req.TargetID, []byte(req.ExtraJSON))
	if err != nil {utils.Fail(c, utils.CodeServerError, err.Error()); return}
	utils.Success(c, gin.H{"favorited": isFav})
}

// ListFavorites GET /favorites?fav_type=&page=&page_size=
func (h *Platform99582Handler) ListFavorites(c *gin.Context) {
	p, ps := getPage(c)
	favType := c.Query("fav_type")
	list, total, err := service.ListUserFavorites(getCurrentUserID(c), favType, p, ps)
	if err != nil {utils.Fail(c, utils.CodeServerError, err.Error()); return}
	utils.Success(c, gin.H{"list": list, "total": total, "page": p, "page_size": ps})
}

// ========== 公告(748~767) ==========

// CreateAnnouncement POST /announcements
func (h *Platform99582Handler) CreateAnnouncement(c *gin.Context) {
	var a model.Announcement
	if err := c.ShouldBindJSON(&a); err != nil {utils.Fail(c, utils.CodeBadRequest, "参数错误"); return}
	if a.AnnType == "platform" && !isPlatformAdmin(c) {
		utils.Fail(c, utils.CodeForbidden, "平台公告仅平台管理员可发布"); return
	}
	if a.PublisherUID <= 0 {a.PublisherUID = getCurrentUserID(c)}
	if err := service.CreateAnnouncement(&a); err != nil {
		utils.Fail(c, utils.CodeServerError, err.Error()); return
	}
	utils.Success(c, gin.H{"id": a.ID})
}

// ListAnnouncements GET /announcements?club_id=
func (h *Platform99582Handler) ListAnnouncements(c *gin.Context) {
	p, ps := getPage(c)
	clubID := parseInt64Query(c, "club_id", 0)
	list, total, err := service.ListAnnouncements(getCurrentUserID(c), clubID, p, ps)
	if err != nil {utils.Fail(c, utils.CodeServerError, err.Error()); return}
	utils.Success(c, gin.H{"list": list, "total": total, "page": p, "page_size": ps})
}

// MarkAnnouncementRead POST /announcements/:id/read
func (h *Platform99582Handler) MarkAnnouncementRead(c *gin.Context) {
	id := parseInt64Path(c, "id")
	if err := service.MarkAnnouncementRead(id, getCurrentUserID(c)); err != nil {
		utils.Fail(c, utils.CodeServerError, err.Error()); return
	}
	utils.Success(c, gin.H{"msg": "ok"})
}

// UpdateAnnouncementStatus POST /admin/announcements/:id/status
func (h *Platform99582Handler) UpdateAnnouncementStatus(c *gin.Context) {
	if !isPlatformAdmin(c) {utils.Fail(c, utils.CodeForbidden, "仅平台管理员"); return}
	id := parseInt64Path(c, "id")
	var req struct {Status int8 `json:"status"`}
	if err := c.ShouldBindJSON(&req); err != nil {utils.Fail(c, utils.CodeBadRequest, "参数错误"); return}
	if err := service.UpdateAnnouncementStatus(id, req.Status); err != nil {
		utils.Fail(c, utils.CodeServerError, err.Error()); return
	}
	utils.Success(c, gin.H{"msg": "ok"})
}

// ========== 全域官方总群配置(141~170, 415, 432, 453~455) ==========

// GetGlobalGroupConfig GET /global-group/:group_id/config
func (h *Platform99582Handler) GetGlobalGroupConfig(c *gin.Context) {
	gid := parseInt64Path(c, "group_id")
	cfg, err := service.GetGlobalGroupConfig(gid)
	if err != nil {utils.Fail(c, utils.CodeServerError, err.Error()); return}
	utils.Success(c, cfg)
}

// UpsertGlobalGroupConfig POST /admin/global-group/config
func (h *Platform99582Handler) UpsertGlobalGroupConfig(c *gin.Context) {
	if !isPlatformAdmin(c) {utils.Fail(c, utils.CodeForbidden, "仅平台管理员"); return}
	var cfg model.GlobalGroupConfig
	if err := c.ShouldBindJSON(&cfg); err != nil {utils.Fail(c, utils.CodeBadRequest, "参数错误"); return}
	if err := service.UpsertGlobalGroupConfig(&cfg); err != nil {
		utils.Fail(c, utils.CodeServerError, err.Error()); return
	}
	utils.Success(c, gin.H{"msg": "ok"})
}

// IncrementAtAllUsed POST /global-group/:group_id/at-all-consume
func (h *Platform99582Handler) IncrementAtAllUsed(c *gin.Context) {
	gid := parseInt64Path(c, "group_id")
	if err := service.IncrementAtAllUsed(gid); err != nil {
		utils.Fail(c, utils.CodeServerError, err.Error()); return
	}
	utils.Success(c, gin.H{"msg": "ok"})
}

// ========== V 标渲染配置(196~215, 212) ==========

// GetAllBadgeRenderConfigs GET /badge-configs (公开下发,前端强制按后端状态渲染,禁止本地伪造 212)
func (h *Platform99582Handler) GetAllBadgeRenderConfigs(c *gin.Context) {
	list, err := service.GetAllBadgeRenderConfigs()
	if err != nil {utils.Fail(c, utils.CodeServerError, err.Error()); return}
	utils.Success(c, gin.H{"list": list})
}

// UpdateBadgeRenderConfig POST /admin/badge-config
func (h *Platform99582Handler) UpdateBadgeRenderConfig(c *gin.Context) {
	if !isPlatformAdmin(c) {utils.Fail(c, utils.CodeForbidden, "仅平台管理员"); return}
	var c_ model.BadgeRenderConfig
	if err := c.ShouldBindJSON(&c_); err != nil {utils.Fail(c, utils.CodeBadRequest, "参数错误"); return}
	if err := service.UpdateBadgeRenderConfig(&c_); err != nil {
		utils.Fail(c, utils.CodeServerError, err.Error()); return
	}
	utils.Success(c, gin.H{"msg": "ok"})
}

// ========== 售后消息举证标记 ==========

// MarkAfterSaleMessage POST /after-sale/:session_id/marks
func (h *Platform99582Handler) MarkAfterSaleMessage(c *gin.Context) {
	sid := parseInt64Path(c, "session_id")
	var req struct {
		MessageID int64  `json:"message_id"`
		MarkType  string `json:"mark_type"`
		Remark    string `json:"remark"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {utils.Fail(c, utils.CodeBadRequest, "参数错误"); return}
	if err := service.MarkAfterSaleMessage(sid, req.MessageID, getCurrentUserID(c), req.MarkType, req.Remark); err != nil {
		utils.Fail(c, utils.CodeServerError, err.Error()); return
	}
	utils.Success(c, gin.H{"msg": "ok"})
}

// ListAfterSaleMarks GET /after-sale/:session_id/marks?mark_type=
func (h *Platform99582Handler) ListAfterSaleMarks(c *gin.Context) {
	sid := parseInt64Path(c, "session_id")
	mt := c.Query("mark_type")
	list, err := service.ListAfterSaleMarks(sid, mt)
	if err != nil {utils.Fail(c, utils.CodeServerError, err.Error()); return}
	utils.Success(c, gin.H{"list": list})
}

// helper: 复用 parse
func parseInt64Path2(c *gin.Context, key string) int64 {
	n, _ := strconv.ParseInt(c.Param(key), 10, 64)
	return n
}
// 防止未使用
var _ = parseInt64Path2
