package service

import (
	"context"
	"crypto/rand"
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/jisan/e-sports-platform/internal/model"
	"github.com/jisan/e-sports-platform/internal/utils"
	"gorm.io/gorm"
)

// LoginResult 登录成功返回结构
type LoginResult struct {
	Token     string       `json:"token"`
	UserType  string       `json:"user_type"`
	ExpiresIn int          `json:"expires_in"`
	User      *model.User  `json:"user,omitempty"`
	Admin     *model.Admin `json:"admin,omitempty"`
	ShopAdmin *model.ShopAdminAccount `json:"shop_admin,omitempty"`
	IsInit    bool         `json:"is_init,omitempty"` // 管理员是否完成首次初始化
}

// WxLoginResult 微信登录返回
type WxLoginResult struct {
	Token     string      `json:"token"`
	ExpiresIn int         `json:"expires_in"`
	User      *model.User `json:"user"`
	IsNew     bool        `json:"is_new"` // 是否新用户
}

// wxCode2SessionResp 微信 code2session 接口响应
type wxCode2SessionResp struct {
	OpenID     string `json:"openid"`
	SessionKey string `json:"session_key"`
	UnionID    string `json:"unionid"`
	ErrCode    int    `json:"errcode"`
	ErrMsg     string `json:"errmsg"`
}

// wxCode2Session 调用微信 code2session 接口换取 openid/session_key
func wxCode2Session(code string) (*wxCode2SessionResp, error) {
	if cfg.WeChat.AppID == "" || cfg.WeChat.AppSecret == "" {
		return nil, errors.New("微信小程序配置缺失")
	}
	url := fmt.Sprintf("https://api.weixin.qq.com/sns/jscode2session?appid=%s&secret=%s&js_code=%s&grant_type=authorization_code",
		cfg.WeChat.AppID, cfg.WeChat.AppSecret, code)

	client := &http.Client{Timeout: 10 * time.Second, Transport: &http.Transport{TLSClientConfig: &tls.Config{InsecureSkipVerify: false}}}
	resp, err := client.Get(url)
	if err != nil {
		return nil, fmt.Errorf("请求微信接口失败: %w", err)
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)

	var res wxCode2SessionResp
	if err := json.Unmarshal(body, &res); err != nil {
		return nil, fmt.Errorf("解析微信响应失败: %w", err)
	}
	if res.ErrCode != 0 {
		return nil, fmt.Errorf("微信登录失败: %s", res.ErrMsg)
	}
	if res.OpenID == "" {
		return nil, errors.New("微信返回 openid 为空")
	}
	return &res, nil
}

// WxLogin 微信小程序登录
// code 为 wx.login 返回的 code；encData/iv 用于解密手机号
func WxLogin(code, encData, iv, nickname, avatar string) (*WxLoginResult, error) {
	sess, err := wxCode2Session(code)
	if err != nil {
		return nil, err
	}

	// 查询或创建用户
	u, err := userRepo.FindByOpenID(sess.OpenID)
	if err != nil {
		return nil, err
	}
	isNew := false
	if u == nil {
		u = &model.User{
			OpenID:   sess.OpenID,
			UnionID:  sess.UnionID,
			Nickname: nickname,
			Avatar:   avatar,
			Role:     model.RoleCustomer,
			Status:   1,
			CreatedAt: nowTimePtr(),
			UpdatedAt: nowTimePtr(),
		}
		if u.Nickname == "" {
			u.Nickname = "用户" + sess.OpenID[len(sess.OpenID)-6:]
		}
		if err := userRepo.Create(u); err != nil {
			return nil, err
		}
		isNew = true
	}

	// 账号封禁校验
	if u.Status == 0 {
		return nil, errors.New("账号已被封禁，请联系客服")
	}

	// 解密手机号(可选)
	if encData != "" && iv != "" && sess.SessionKey != "" {
		phone, perr := decryptWxPhone(encData, iv, sess.SessionKey)
		if perr == nil && phone != "" {
			_ = userRepo.Update(u.ID, map[string]interface{}{"phone": phone})
			u.Phone = phone
		}
	}

	// 更新昵称头像(若传入)
	if (nickname != "" && u.Nickname != nickname) || (avatar != "" && u.Avatar != avatar) {
		fields := map[string]interface{}{}
		if nickname != "" {
			fields["nickname"] = nickname
		}
		if avatar != "" {
			fields["avatar"] = avatar
		}
		_ = userRepo.Update(u.ID, fields)
	}

	// 生成 token
	token, err := jwtMgr.GenerateToken(u.ID, u.Role, utils.JWTUserTypeUser, u.ClubID)
	if err != nil {
		return nil, err
	}
	return &WxLoginResult{
		Token:     token,
		ExpiresIn: cfg.JWT.ExpireHours * 3600,
		User:      u,
		IsNew:     isNew,
	}, nil
}

// decryptWxPhone 解密微信手机号(AES-128-CBC)
func decryptWxPhone(encData, iv, sessionKey string) (string, error) {
	return DecryptWxPhoneNumber(encData, iv, sessionKey)
}

// Register 用户注册(绑定手机号/邀请码等补充信息)
func Register(userID int64, phone, inviteCode string) (*model.User, error) {
	u, err := userRepo.FindByID(userID)
	if err != nil {
		return nil, err
	}
	if u == nil {
		return nil, errors.New("用户不存在")
	}
	if phone != "" {
		if !utils.ValidateMobile(phone) {
			return nil, errors.New("手机号格式错误")
		}
		// 手机号二次放号检测:同手机号存在其他用户且非自己
		exist, err := userRepo.FindByPhone(phone)
		if err != nil {
			return nil, err
		}
		if exist != nil && exist.ID != userID {
			// 标记旧账号手机号被回收
			_ = userRepo.Update(exist.ID, map[string]interface{}{"is_phone_abandoned": 1})
		}
		_ = userRepo.Update(userID, map[string]interface{}{"phone": phone})
		u.Phone = phone
	}
	// 邀请码绑定(若提供)
	if inviteCode != "" && u.InviteCode == "" {
		code, err := inviteCodeRepo.FindByCode(inviteCode)
		if err != nil {
			return nil, err
		}
		if code == nil {
			return nil, errors.New("邀请码不存在")
		}
		if code.Status == model.InviteCodeStatusRevoked {
			return nil, errors.New("邀请码已撤销")
		}
		if code.ExpireAt != nil && code.ExpireAt.Before(time.Now()) {
			return nil, errors.New("邀请码已过期")
		}
		// 消费邀请码(原子)
		_, ok, err := inviteCodeRepo.Consume(inviteCode, userID)
		if err != nil {
			return nil, err
		}
		if !ok {
			return nil, errors.New("邀请码已用尽或不可用")
		}
		// 绑定角色与俱乐部
		updates := map[string]interface{}{"invite_code": inviteCode}
		if code.Type == model.InviteCodeTypeClub && code.ClubID > 0 {
			updates["club_id"] = code.ClubID
			// 俱乐部打手/分销商角色
			if code.Role == model.InviteCodeRoleDS {
				updates["role"] = gorm.Expr("role | ?", model.RolePlayer)
			} else if code.Role == model.InviteCodeRoleFXS {
				updates["role"] = gorm.Expr("role | ?", model.RoleDistributor)
			}
		}
		_ = userRepo.Update(userID, updates)
	}
	// 返回最新用户
	u, _ = userRepo.FindByID(userID)
	return u, nil
}

// DecodePhone 解密微信手机号(登录态下二次解密)
func DecodePhone(userID int64, encData, iv, sessionKey string) (string, error) {
	if encData == "" || iv == "" {
		return "", errors.New("加密数据/IV 不能为空")
	}
	phone, err := DecryptWxPhoneNumber(encData, iv, sessionKey)
	if err != nil {
		return "", err
	}
	if userID > 0 {
		_ = userRepo.Update(userID, map[string]interface{}{"phone": phone})
	}
	return phone, nil
}

// AdminLogin 平台管理员登录(用户名+密码)
func AdminLogin(username, password, ip string) (*LoginResult, error) {
	a, err := adminRepo.FindByUsername(username)
	if err != nil {
		return nil, err
	}
	if a == nil {
		return nil, errors.New("账号或密码错误")
	}
	if a.Status == 0 {
		return nil, errors.New("管理员账号已被禁用")
	}
	if !utils.CheckPassword(password, a.Password) {
		return nil, errors.New("账号或密码错误")
	}
	_ = adminRepo.UpdateLastLogin(a.ID, ip)
	token, err := jwtMgr.GenerateToken(a.ID, a.Role, utils.JWTUserTypeAdmin, 0)
	if err != nil {
		return nil, err
	}
	a.Password = ""
	return &LoginResult{
		Token:     token,
		UserType:  utils.JWTUserTypeAdmin,
		ExpiresIn: cfg.JWT.AdminExpireHours * 3600,
		Admin:     a,
		IsInit:    a.IsInit == 1,
	}, nil
}

// ShopAdminLogin 内置管理端登录
func ShopAdminLogin(username, password, ip string) (*LoginResult, error) {
	a, err := clubRepo.FindShopAdminByUsername(username)
	if err != nil {
		return nil, err
	}
	if a == nil {
		return nil, errors.New("账号或密码错误")
	}
	if a.Status == 0 {
		return nil, errors.New("管理端账号已被禁用")
	}
	if !utils.CheckPassword(password, a.Password) {
		return nil, errors.New("账号或密码错误")
	}
	_ = clubRepo.UpdateShopAdmin(a.ID, map[string]interface{}{
		"last_login_at": nowTimePtr(),
		"last_login_ip": ip,
	})
	token, err := jwtMgr.GenerateToken(a.ID, a.Role, utils.JWTUserTypeShopAdmin, a.ClubID)
	if err != nil {
		return nil, err
	}
	a.Password = ""
	return &LoginResult{
		Token:     token,
		UserType:  utils.JWTUserTypeShopAdmin,
		ExpiresIn: cfg.JWT.AdminExpireHours * 3600,
		ShopAdmin: a,
	}, nil
}

// ForgotAccount 忘记账号(通过邮箱反查用户名 + 发送验证码)
func ForgotAccount(email string) (string, error) {
	if !utils.ValidateEmail(email) {
		return "", errors.New("邮箱格式错误")
	}
	var a model.Admin
	if err := db.Where("email = ?", email).First(&a).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return "", errors.New("未找到绑定该邮箱的管理员")
		}
		return "", err
	}
	// 生成6位验证码并发送邮件
	code := genVerifyCode() // 安全:crypto/rand 生成,不可预测
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	_ = redis.Set(ctx, cacheKey("captcha:forgot:"+email), code+":"+a.Username, 5*time.Minute)
	// 真发送邮件（沙箱 fallback 到日志）
	smtpCfg := &utils.SMTPConfig{
		Host:     cfg.SMTP.Host,
		Port:     cfg.SMTP.Port,
		User:     cfg.SMTP.User,
		Password: cfg.SMTP.Password,
		From:     cfg.SMTP.From,
		Sandbox:  cfg.SMTP.Sandbox,
	}
	_ = utils.SendVerifyCode(email, code, "forgot", smtpCfg, logger)
	return a.Username, nil
}

// ForgotPassword 忘记密码(发送重置验证码到邮箱)
func ForgotPassword(username string) error {
	a, err := adminRepo.FindByUsername(username)
	if err != nil {
		return err
	}
	if a == nil || a.Email == "" {
		return errors.New("账号未绑定邮箱，无法重置")
	}
	// 生成 6 位验证码并存入 Redis(5 分钟)
	code := genVerifyCode() // 安全:crypto/rand 生成,不可预测
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	_ = redis.Set(ctx, cacheKey("captcha:reset:"+a.Email), code, 5*time.Minute)
	// 真发送邮件（沙箱 fallback 到日志）
	smtpCfg := &utils.SMTPConfig{
		Host:     cfg.SMTP.Host,
		Port:     cfg.SMTP.Port,
		User:     cfg.SMTP.User,
		Password: cfg.SMTP.Password,
		From:     cfg.SMTP.From,
		Sandbox:  cfg.SMTP.Sandbox,
	}
	_ = utils.SendVerifyCode(a.Email, code, "reset", smtpCfg, logger)
	return nil
}

// WebauthnBegin 开始 WebAuthn 注册/认证流程
// 返回 challenge，并将 challenge->uid 映射存储到 Redis(5分钟过期)
// ——真实项目集成 go-webauthn 说明——
// 1. go get github.com/go-webauthn/webauthn/v4
// 2. 初始化 webauthn := &webauthn.WebAuthn{RPID=xxx, RPOrigin=xxx, Timeout:60000}
// 3. 注册 BeginRegistration(user webauthn.User) -> sessionData, credentialCreationOptions
//    将 sessionData 序列化存入 Redis(key = webauthn:challenge:<challenge>, value=uid+sessionData)
// 4. 前端完成 navigator.credentials.create 后回传 assertion
// 5. FinishRegistration(user, session, req) 校验 signature -> 写入 DB webauthn_credentials
// ——以下为沙箱模式:仅校验 challenge 存在+未过期+uid匹配——
func WebauthnBegin(username string, userID int64) (map[string]interface{}, error) {
	// 用 crypto/rand 生成 challenge(32字节 hex 64位)
	var b [32]byte
	_, _ = rand.Read(b[:])
	challenge := fmt.Sprintf("%x", b[:])
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	// 存储 challenge -> uid JSON(含username,过期时间)
	cacheObj := map[string]interface{}{
		"uid":       userID,
		"username":  username,
		"created":   time.Now().Unix(),
		"ttl_sec":   300,
	}
	cacheBytes, _ := json.Marshal(cacheObj)
	_ = redis.Set(ctx, cacheKey("webauthn:challenge:"+challenge), cacheBytes, 5*time.Minute)
	// 同时保留 username->challenge 兼容旧版
	_ = redis.Set(ctx, cacheKey("webauthn:"+username), challenge, 5*time.Minute)
	return map[string]interface{}{
		"challenge": challenge,
		"username":  username,
		"uid":       userID,
		"timeout":   300,
		// 真实项目：把 webauthn 生成的 publicKeyCredentialCreationOptions 返回给前端
		"rp": map[string]interface{}{
			"id":   "localhost",
			"name": "E-Sports Platform",
		},
		"user": map[string]interface{}{
			"id":   userID,
			"name": username,
		},
	}, nil
}

// WebauthnFinish 完成 WebAuthn 流程(校验 assertion signature)
// 沙箱模式:仅检查 challenge 存在+未过期、对应 uid 匹配通过
// ——真实项目 go-webauthn 集成点——
// 1. 用 redis 取出 sessionData(反序列化 webauthn.SessionData)
// 2. 根据用户 ID 加载 User(实现 webauthn.User 接口)
// 3. 注册场景:cred, err := webauthn.FinishRegistration(user, session, httpReq)
//    如果 err != nil 返回签名校验失败；否则存 cred.Descriptor().ID.Base64URLEncoded()
//    以及 cred.PublicKeyPem() / credential.Raw 到 DB
// 4. 认证(登录)场景:webauthn.FinishLogin(user, session, httpReq)
//    通过则颁发 JWT
// ——沙箱校验流程——
func WebauthnFinish(username, challenge, credentialID, publicKey, deviceInfo string, uid int64) error {
	if challenge == "" {
		return errors.New("缺少 challenge")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	// 1. 从 Redis 取 challenge 存储数据
	cached, err := redis.Get(ctx, cacheKey("webauthn:challenge:"+challenge))
	if err != nil || cached == "" {
		// 回退: 旧 username->challenge
		ch2, _ := redis.Get(ctx, cacheKey("webauthn:"+username))
		if ch2 != challenge {
			return errors.New("challenge 已过期或无效，请重新开始")
		}
	} else {
		var obj map[string]interface{}
		if jerr := json.Unmarshal([]byte(cached), &obj); jerr == nil {
			// 2. 校验 challenge 未过期(创建+TTL > now)
			created, _ := obj["created"].(float64)
			ttl, _ := obj["ttl_sec"].(float64)
			if created > 0 && ttl > 0 {
				if int64(created)+int64(ttl) < time.Now().Unix() {
					_ = redis.Del(ctx, cacheKey("webauthn:challenge:"+challenge))
					return errors.New("challenge 已过期，请重新开始")
				}
			}
			// 3. 校验 uid 匹配
			if uid > 0 {
				storedUID := int64(0)
				switch v := obj["uid"].(type) {
				case float64:
					storedUID = int64(v)
				case int64:
					storedUID = v
				}
				if storedUID > 0 && storedUID != uid {
					return errors.New("challenge 对应用户不匹配")
				}
			}
		}
	}
	// 通过: 存库(管理员 Webauthn / 用户 Webauthn 都走这里，兼容两种)
	if uid > 0 {
		// 用户级 Webauthn (预留，DeviceInfo 字段存 uid 关联)
		_ = db.Create(&model.AdminWebauthn{
			AdminID:      0,
			CredentialID: credentialID,
			PublicKey:    publicKey,
			DeviceInfo:   "uid:" + itoa(uid) + "|" + deviceInfo,
			CreatedAt:    nowTimePtr(),
		}).Error
	} else if username != "" {
		a, err := adminRepo.FindByUsername(username)
		if err != nil {
			return err
		}
		if a == nil {
			return errors.New("管理员不存在")
		}
		w := &model.AdminWebauthn{
			AdminID:      a.ID,
			CredentialID: credentialID,
			PublicKey:    publicKey,
			DeviceInfo:   deviceInfo,
			CreatedAt:    nowTimePtr(),
		}
		if err := adminRepo.CreateWebauthn(w); err != nil {
			return err
		}
	}
	// 清理 challenge
	_ = redis.Del(ctx, cacheKey("webauthn:challenge:"+challenge))
	_ = redis.Del(ctx, cacheKey("webauthn:"+username))
	return nil
}

// AdminChangePassword 管理员首次初始化修改密码
// 校验密码强度与历史密码不重复
func AdminChangePassword(adminID int64, oldPassword, newPassword string) error {
	a, err := adminRepo.FindByID(adminID)
	if err != nil {
		return err
	}
	if a == nil {
		return errors.New("管理员不存在")
	}
	if !utils.CheckPassword(oldPassword, a.Password) {
		return errors.New("原密码错误")
	}
	if len(newPassword) < 8 {
		return errors.New("新密码长度至少 8 位")
	}
	// 校验历史密码不重复(最近 5 条)
	hist, err := adminRepo.ListPasswordHistory(adminID, 5)
	if err != nil {
		return err
	}
	for _, h := range hist {
		if utils.CheckPassword(newPassword, h.PasswordHash) {
			return errors.New("新密码不能与最近使用过的密码重复")
		}
	}
	hash, err := utils.HashPassword(newPassword)
	if err != nil {
		return err
	}
	if err := adminRepo.Update(adminID, map[string]interface{}{
		"password": hash,
		"is_init":  1,
	}); err != nil {
		return err
	}
	_ = adminRepo.CreatePasswordHistory(&model.AdminPasswordHistory{
		AdminID:      adminID,
		PasswordHash: hash,
		CreatedAt:    nowTimePtr(),
	})
	return nil
}

// AdminBindEmail 管理员绑定邮箱(真发送验证码，沙箱模式下 fallback 写日志)
func AdminBindEmail(adminID int64, email string) error {
	if !utils.ValidateEmail(email) {
		return errors.New("邮箱格式错误")
	}
	a, err := adminRepo.FindByID(adminID)
	if err != nil {
		return err
	}
	if a == nil {
		return errors.New("管理员不存在")
	}
	code := genVerifyCode() // 安全:crypto/rand 生成,不可预测
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	_ = redis.Set(ctx, cacheKey("captcha:email:"+fmt.Sprint(adminID)), code+":"+email, 5*time.Minute)
	// 真发送邮件（沙箱 fallback 到日志）
	smtpCfg := &utils.SMTPConfig{
		Host:     cfg.SMTP.Host,
		Port:     cfg.SMTP.Port,
		User:     cfg.SMTP.User,
		Password: cfg.SMTP.Password,
		From:     cfg.SMTP.From,
		Sandbox:  cfg.SMTP.Sandbox,
	}
	if err := utils.SendVerifyCode(email, code, "bind", smtpCfg, logger); err != nil {
		// 发送失败时兜底写日志（沙箱模式已兜底，此处处理非沙箱网络异常）
		if logger != nil {
			logger.Warn("管理员绑定邮箱验证码邮件发送失败，已降级为日志输出",
				zapField("admin_id", fmt.Sprint(adminID)),
				zapField("email", email),
				zapField("code", code),
				zapField("error", err.Error()))
		}
	}
	return nil
}

// genVerifyCode 生成 6 位数字验证码(使用 crypto/rand,不可预测)
// 安全修复:原使用 time.Now().UnixNano()%1000000 基于时间可预测,
// 攻击者知道请求时间即可推算验证码,导致邮箱绑定/密码重置可被绕过
func genVerifyCode() string {
	var b [4]byte
	_, _ = rand.Read(b[:])
	n := uint32(b[0])<<24 | uint32(b[1])<<16 | uint32(b[2])<<8 | uint32(b[3])
	return fmt.Sprintf("%06d", n%1000000)
}

// captchaMaxTry 验证码最大尝试次数(防暴力破解 6 位数字验证码)
const captchaMaxTry = 5

// AdminVerifyEmail 校验邮箱验证码并完成绑定
// 安全修复:增加尝试次数限制(5 次错误后验证码失效),防暴力破解
func AdminVerifyEmail(adminID int64, code string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	cacheK := cacheKey("captcha:email:" + fmt.Sprint(adminID))
	val, err := redis.Get(ctx, cacheK)
	if err != nil || val == "" {
		return errors.New("验证码已过期，请重新获取")
	}
	parts := strings.SplitN(val, ":", 2)
	if len(parts) != 2 || parts[0] != code {
		// 安全:尝试次数限制(防暴力破解 6 位验证码)
		tryK := cacheKey("captcha:email:try:" + fmt.Sprint(adminID))
		tryStr, _ := redis.Get(ctx, tryK)
		tryCnt := int64(0)
		if tryStr != "" {
			tryCnt, _ = strconv.ParseInt(tryStr, 10, 64)
		}
		tryCnt++
		_ = redis.Set(ctx, tryK, fmt.Sprint(tryCnt), 5*time.Minute)
		if tryCnt >= captchaMaxTry {
			_ = redis.Del(ctx, cacheK, tryK)
			return errors.New("验证码错误次数过多,已失效,请重新获取")
		}
		return fmt.Errorf("验证码错误(剩余尝试 %d 次)", captchaMaxTry-tryCnt)
	}
	if err := adminRepo.Update(adminID, map[string]interface{}{"email": parts[1], "is_init": 1}); err != nil {
		return err
	}
	_ = redis.Del(ctx, cacheK)
	_ = redis.Del(ctx, cacheKey("captcha:email:try:"+fmt.Sprint(adminID)))
	return nil
}

// AdminGetInitStatus 获取管理员初始化状态
func AdminGetInitStatus(adminID int64) (map[string]interface{}, error) {
	a, err := adminRepo.FindByID(adminID)
	if err != nil {
		return nil, err
	}
	if a == nil {
		return nil, errors.New("管理员不存在")
	}
	return map[string]interface{}{
		"is_init":     a.IsInit,
		"has_email":   a.Email != "",
		"has_webauthn": len(adminWebauthnCount(adminID)) > 0,
	}, nil
}

func adminWebauthnCount(adminID int64) []model.AdminWebauthn {
	list, _ := adminRepo.ListWebauthnByAdmin(adminID)
	return list
}
