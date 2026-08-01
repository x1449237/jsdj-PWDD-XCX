package handler

import (
	"github.com/gin-gonic/gin"
	"github.com/jisan/e-sports-platform/internal/service"
	"github.com/jisan/e-sports-platform/internal/utils"
)

// AuthHandler 认证相关处理器
type AuthHandler struct{}

// NewAuthHandler 创建认证处理器
func NewAuthHandler() *AuthHandler { return &AuthHandler{} }

// wxLoginRequest 微信小程序登录请求
type wxLoginRequest struct {
	Code     string `json:"code" binding:"required"` // wx.login 返回的 code
	EncData  string `json:"enc_data"`                // 加密数据(手机号)
	IV       string `json:"iv"`                      // 初始向量
	Nickname string `json:"nickname"`                // 昵称
	Avatar   string `json:"avatar"`                  // 头像
}

// WxLogin 微信小程序登录
// POST /api/v1/auth/wx-login
func (h *AuthHandler) WxLogin(c *gin.Context) {
	var req wxLoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Fail(c, utils.CodeBadRequest, "参数错误: "+err.Error())
		return
	}
	res, err := service.WxLogin(req.Code, req.EncData, req.IV, req.Nickname, req.Avatar)
	if err != nil {
		utils.Fail(c, utils.CodeBadRequest, err.Error())
		return
	}
	utils.Success(c, res)
}

// registerRequest 注册补充信息请求(绑定手机号/邀请码)
type registerRequest struct {
	Phone      string `json:"phone"`       // 手机号
	InviteCode string `json:"invite_code"` // 邀请码
}

// Register 用户注册(补充手机号/邀请码)
// POST /api/v1/auth/register
func (h *AuthHandler) Register(c *gin.Context) {
	var req registerRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Fail(c, utils.CodeBadRequest, "参数错误: "+err.Error())
		return
	}
	userID := getCurrentUserID(c)
	if userID == 0 {
		utils.Fail(c, utils.CodeUnauthorized, "未登录")
		return
	}
	u, err := service.Register(userID, req.Phone, req.InviteCode)
	if err != nil {
		utils.Fail(c, utils.CodeBadRequest, err.Error())
		return
	}
	utils.Success(c, u)
}

// decodePhoneRequest 解密微信手机号请求
type decodePhoneRequest struct {
	EncData    string `json:"enc_data" binding:"required"`
	IV         string `json:"iv" binding:"required"`
	SessionKey string `json:"session_key" binding:"required"`
}

// DecodePhone 登录态下二次解密微信手机号
// POST /api/v1/auth/decode-phone
func (h *AuthHandler) DecodePhone(c *gin.Context) {
	var req decodePhoneRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Fail(c, utils.CodeBadRequest, "参数错误: "+err.Error())
		return
	}
	userID := getCurrentUserID(c)
	phone, err := service.DecodePhone(userID, req.EncData, req.IV, req.SessionKey)
	if err != nil {
		utils.Fail(c, utils.CodeBadRequest, err.Error())
		return
	}
	utils.Success(c, gin.H{"phone": phone})
}

// adminLoginRequest 平台管理员登录请求
type adminLoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

// AdminLogin 平台管理员登录
// POST /api/v1/admin/login
func (h *AuthHandler) AdminLogin(c *gin.Context) {
	var req adminLoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Fail(c, utils.CodeBadRequest, "参数错误: "+err.Error())
		return
	}
	res, err := service.AdminLogin(req.Username, req.Password, getClientIP(c))
	if err != nil {
		utils.Fail(c, utils.CodeBadRequest, err.Error())
		return
	}
	utils.Success(c, res)
}

// ShopAdminLogin 内置管理端登录
// POST /api/v1/shop/login
func (h *AuthHandler) ShopAdminLogin(c *gin.Context) {
	var req adminLoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Fail(c, utils.CodeBadRequest, "参数错误: "+err.Error())
		return
	}
	res, err := service.ShopAdminLogin(req.Username, req.Password, getClientIP(c))
	if err != nil {
		utils.Fail(c, utils.CodeBadRequest, err.Error())
		return
	}
	utils.Success(c, res)
}

// forgotAccountRequest 忘记账号请求
type forgotAccountRequest struct {
	Email string `json:"email" binding:"required"`
}

// ForgotAccount 忘记账号(邮箱反查用户名)
// POST /api/v1/admin/forgot-account
func (h *AuthHandler) ForgotAccount(c *gin.Context) {
	var req forgotAccountRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Fail(c, utils.CodeBadRequest, "参数错误: "+err.Error())
		return
	}
	username, err := service.ForgotAccount(req.Email)
	if err != nil {
		utils.Fail(c, utils.CodeBadRequest, err.Error())
		return
	}
	utils.Success(c, gin.H{"username": username})
}

// forgotPasswordRequest 忘记密码请求
type forgotPasswordRequest struct {
	Username string `json:"username" binding:"required"`
}

// ForgotPassword 忘记密码(发送重置验证码)
// POST /api/v1/admin/forgot-password
func (h *AuthHandler) ForgotPassword(c *gin.Context) {
	var req forgotPasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Fail(c, utils.CodeBadRequest, "参数错误: "+err.Error())
		return
	}
	if err := service.ForgotPassword(req.Username); err != nil {
		utils.Fail(c, utils.CodeBadRequest, err.Error())
		return
	}
	utils.Success(c, gin.H{"msg": "验证码已发送"})
}

// webauthnBeginRequest WebAuthn 开始请求
type webauthnBeginRequest struct {
	Username string `json:"username" binding:"required"`
}

// WebauthnBegin 开始 WebAuthn 流程
// POST /api/v1/admin/webauthn/begin
func (h *AuthHandler) WebauthnBegin(c *gin.Context) {
	var req webauthnBeginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Fail(c, utils.CodeBadRequest, "参数错误: "+err.Error())
		return
	}
	res, err := service.WebauthnBegin(req.Username)
	if err != nil {
		utils.Fail(c, utils.CodeBadRequest, err.Error())
		return
	}
	utils.Success(c, res)
}

// webauthnFinishRequest WebAuthn 完成请求
type webauthnFinishRequest struct {
	Username     string `json:"username" binding:"required"`
	CredentialID string `json:"credential_id" binding:"required"`
	PublicKey    string `json:"public_key" binding:"required"`
	DeviceInfo   string `json:"device_info"`
}

// WebauthnFinish 完成 WebAuthn 流程
// POST /api/v1/admin/webauthn/finish
func (h *AuthHandler) WebauthnFinish(c *gin.Context) {
	var req webauthnFinishRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Fail(c, utils.CodeBadRequest, "参数错误: "+err.Error())
		return
	}
	if err := service.WebauthnFinish(req.Username, req.CredentialID, req.PublicKey, req.DeviceInfo); err != nil {
		utils.Fail(c, utils.CodeBadRequest, err.Error())
		return
	}
	utils.Success(c, gin.H{"msg": "ok"})
}

// changePasswordRequest 修改密码请求
type changePasswordRequest struct {
	OldPassword string `json:"old_password" binding:"required"`
	NewPassword string `json:"new_password" binding:"required"`
}

// AdminChangePassword 管理员修改密码(首次初始化)
// POST /api/v1/admin/change-password
func (h *AuthHandler) AdminChangePassword(c *gin.Context) {
	var req changePasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Fail(c, utils.CodeBadRequest, "参数错误: "+err.Error())
		return
	}
	adminID := getCurrentUserID(c)
	if err := service.AdminChangePassword(adminID, req.OldPassword, req.NewPassword); err != nil {
		utils.Fail(c, utils.CodeBadRequest, err.Error())
		return
	}
	utils.Success(c, gin.H{"msg": "ok"})
}

// bindEmailRequest 绑定邮箱请求
type bindEmailRequest struct {
	Email string `json:"email" binding:"required"`
}

// AdminBindEmail 管理员绑定邮箱(发送验证码)
// POST /api/v1/admin/bind-email
func (h *AuthHandler) AdminBindEmail(c *gin.Context) {
	var req bindEmailRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Fail(c, utils.CodeBadRequest, "参数错误: "+err.Error())
		return
	}
	adminID := getCurrentUserID(c)
	if err := service.AdminBindEmail(adminID, req.Email); err != nil {
		utils.Fail(c, utils.CodeBadRequest, err.Error())
		return
	}
	utils.Success(c, gin.H{"msg": "验证码已发送"})
}

// verifyEmailRequest 校验邮箱验证码请求
type verifyEmailRequest struct {
	Code string `json:"code" binding:"required"`
}

// AdminVerifyEmail 校验邮箱验证码并完成绑定
// POST /api/v1/admin/verify-email
func (h *AuthHandler) AdminVerifyEmail(c *gin.Context) {
	var req verifyEmailRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Fail(c, utils.CodeBadRequest, "参数错误: "+err.Error())
		return
	}
	adminID := getCurrentUserID(c)
	if err := service.AdminVerifyEmail(adminID, req.Code); err != nil {
		utils.Fail(c, utils.CodeBadRequest, err.Error())
		return
	}
	utils.Success(c, gin.H{"msg": "ok"})
}

// AdminGetInitStatus 获取管理员初始化状态
// GET /api/v1/admin/init-status
func (h *AuthHandler) AdminGetInitStatus(c *gin.Context) {
	adminID := getCurrentUserID(c)
	res, err := service.AdminGetInitStatus(adminID)
	if err != nil {
		utils.Fail(c, utils.CodeBadRequest, err.Error())
		return
	}
	utils.Success(c, res)
}
