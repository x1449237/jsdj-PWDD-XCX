package handler

import (
	"encoding/json"
	"io"
	"path/filepath"

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

// ================ 板块一：入驻前置全局 ================

// CheckClubSwitch 查询俱乐部入驻开关
// GET /api/v1/clubs/join-switch
// 公开接口：用户在前端入驻页打开时先查询开关状态
func (h *ClubHandler) CheckClubSwitch(c *gin.Context) {
	enabled, err := service.CheckClubSwitch()
	if err != nil {
		utils.Fail(c, utils.CodeServerError, err.Error())
		return
	}
	utils.Success(c, gin.H{"enabled": enabled})
}

// generateAbbrRequest 生成俱乐部缩写请求
type generateAbbrRequest struct {
	Name string `json:"name" binding:"required"`
}

// GenerateAbbr 生成俱乐部缩写(含查重)
// POST /api/v1/clubs/abbr
// 返回主缩写；冲突时返回备选缩写列表
func (h *ClubHandler) GenerateAbbr(c *gin.Context) {
	var req generateAbbrRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Fail(c, utils.CodeBadRequest, "参数错误: "+err.Error())
		return
	}
	abbr, alternatives, err := service.GenerateAbbreviation(req.Name)
	if err != nil {
		// 缩写冲突时返回备选 + 错误信息
		if len(alternatives) > 0 {
			utils.Success(c, gin.H{
				"abbreviation": "",
				"alternatives": alternatives,
				"conflict":     true,
				"msg":          err.Error(),
			})
			return
		}
		utils.Fail(c, utils.CodeBadRequest, err.Error())
		return
	}
	utils.Success(c, gin.H{
		"abbreviation": abbr,
		"alternatives": []string{},
		"conflict":     false,
	})
}

// UploadPDF 上传 PDF(校验文件大小、扩展名、文件头魔数)
// POST /api/v1/clubs/upload-pdf
// multipart/form-data 字段: file=<PDF文件>
// 用于个人入驻自我声明、企业入驻代办授权书等场景的 PDF 上传校验
func (h *ClubHandler) UploadPDF(c *gin.Context) {
	fileHeader, err := c.FormFile("file")
	if err != nil {
		utils.Fail(c, utils.CodeBadRequest, "未接收到文件: "+err.Error())
		return
	}
	// 1. 扩展名校验
	if !utils.ValidatePDFFileExtension(fileHeader.Filename) {
		utils.Fail(c, utils.CodeBadRequest, "文件扩展名必须为 .pdf")
		return
	}
	// 2. 打开文件读取字节(校验魔数 + 大小)
	f, err := fileHeader.Open()
	if err != nil {
		utils.Fail(c, utils.CodeServerError, "打开文件失败: "+err.Error())
		return
	}
	defer f.Close()
	fileBytes, err := io.ReadAll(f)
	if err != nil {
		utils.Fail(c, utils.CodeServerError, "读取文件失败: "+err.Error())
		return
	}
	// 3. 内容校验(大小 + 文件头魔数)
	if ok, vErr := utils.ValidatePDF(fileBytes); !ok {
		utils.Fail(c, utils.CodeBadRequest, vErr.Error())
		return
	}
	// 沙箱模式：校验通过即返回占位 URL，由前端再上传至 OSS
	// 真实项目应在此调用 OSS 上传并返回最终 URL
	safeName := filepath.Base(fileHeader.Filename)
	utils.Success(c, gin.H{
		"msg":      "ok",
		"filename": safeName,
		"size":     len(fileBytes),
		"url":      "", // 实际项目由 OSS 上传后回填
	})
}

// ================ 板块二：个人入驻 ================

// SubmitPersonal 提交个人入驻申请
// POST /api/v1/clubs/registrations/personal
func (h *ClubHandler) SubmitPersonal(c *gin.Context) {
	userID := getCurrentUserID(c)
	var form service.PersonalRegistrationForm
	if err := c.ShouldBindJSON(&form); err != nil {
		utils.Fail(c, utils.CodeBadRequest, "参数错误: "+err.Error())
		return
	}
	reg, err := service.SubmitPersonalRegistration(userID, form)
	if err != nil {
		utils.Fail(c, utils.CodeBadRequest, err.Error())
		return
	}
	utils.Success(c, reg)
}

// ================ 草稿保存(7天) ================

// saveDraftRequest 保存草稿请求
type saveDraftRequest struct {
	DraftData json.RawMessage `json:"draft_data" binding:"required"`
}

// SaveDraft 保存入驻草稿(7 天有效期)
// POST /api/v1/clubs/draft
func (h *ClubHandler) SaveDraft(c *gin.Context) {
	userID := getCurrentUserID(c)
	var req saveDraftRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Fail(c, utils.CodeBadRequest, "参数错误: "+err.Error())
		return
	}
	if err := service.SaveClubDraft(userID, req.DraftData); err != nil {
		utils.Fail(c, utils.CodeBadRequest, err.Error())
		return
	}
	utils.Success(c, gin.H{"msg": "ok"})
}

// GetDraft 获取入驻草稿(过期自动清理)
// GET /api/v1/clubs/draft
func (h *ClubHandler) GetDraft(c *gin.Context) {
	userID := getCurrentUserID(c)
	draft, err := service.GetClubDraft(userID)
	if err != nil {
		utils.Fail(c, utils.CodeServerError, err.Error())
		return
	}
	utils.Success(c, draft)
}

// ================ 板块三：企业入驻 ================

// SubmitEnterprise 提交企业入驻申请
// POST /api/v1/clubs/registrations/enterprise
func (h *ClubHandler) SubmitEnterprise(c *gin.Context) {
	userID := getCurrentUserID(c)
	var form service.EnterpriseRegistrationForm
	if err := c.ShouldBindJSON(&form); err != nil {
		utils.Fail(c, utils.CodeBadRequest, "参数错误: "+err.Error())
		return
	}
	reg, err := service.SubmitEnterpriseRegistration(userID, form)
	if err != nil {
		utils.Fail(c, utils.CodeBadRequest, err.Error())
		return
	}
	utils.Success(c, reg)
}

// ================ 合同盖章提醒标记 ================

// GetSealRemind 返回合同盖章提醒标记
// GET /api/v1/clubs/seal-remind
// 前端在企业入驻页展示"请确认合同已盖章"提示
func (h *ClubHandler) GetSealRemind(c *gin.Context) {
	utils.Success(c, gin.H{"reminded": service.GetSealRemindedFlag()})
}
