package utils

import (
	"crypto/tls"
	"errors"
	"fmt"
	"net/smtp"
	"strings"

	"go.uber.org/zap"
)

// smtpLogger 内部日志引用（由调用方初始化）
var smtpLogger *zap.Logger

// SetSMTPLogger 设置邮件服务日志
func SetSMTPLogger(l *zap.Logger) {
	smtpLogger = l
}

// SMTPConfig 邮件发送配置
type SMTPConfig struct {
	Host     string
	Port     int
	User     string
	Password string
	From     string
	Sandbox  bool
}

// globalSMTPConf 全局SMTP配置（由InitSMTP设置）
var globalSMTPConf *SMTPConfig

// InitSMTP 初始化SMTP配置
func InitSMTP(cfg *SMTPConfig) {
	globalSMTPConf = cfg
}

// buildHTMLTemplate 构造验证码HTML邮件内容
func buildHTMLTemplate(code, kind string) string {
	kindText := "邮箱验证码"
	switch kind {
	case "bind":
		kindText = "邮箱绑定验证码"
	case "reset":
		kindText = "密码重置验证码"
	case "forgot":
		kindText = "找回账号验证码"
	}
	return fmt.Sprintf(`<!DOCTYPE html>
<html>
<head>
<meta charset="UTF-8">
<title>%s</title>
</head>
<body style="font-family: 'Helvetica Neue', Arial, sans-serif; background: #f5f7fa; padding: 30px;">
<div style="max-width: 560px; margin: 0 auto; background: #ffffff; border-radius: 12px; overflow: hidden; box-shadow: 0 4px 20px rgba(0,0,0,0.06);">
  <div style="background: linear-gradient(135deg, #667eea 0%%, #764ba2 100%%); padding: 32px 30px; color: #fff;">
    <h2 style="margin: 0; font-size: 22px;">戟三电竞平台</h2>
    <p style="margin: 8px 0 0; opacity: 0.9;">%s</p>
  </div>
  <div style="padding: 36px 30px;">
    <p style="margin: 0 0 16px; font-size: 15px; color: #333;">您好！</p>
    <p style="margin: 0 0 20px; font-size: 15px; color: #555; line-height: 1.7;">
      您正在使用 <b>戟三电竞平台</b> 的 %s 功能，以下是您的验证码：
    </p>
    <div style="background: #f0f3ff; border: 1px solid #dce4ff; border-radius: 8px; padding: 18px 24px; text-align: center; margin-bottom: 20px;">
      <span style="font-size: 32px; font-weight: 700; letter-spacing: 6px; color: #4a5cf7;">%s</span>
    </div>
    <p style="margin: 0 0 10px; font-size: 14px; color: #888;">验证码为 <b>6 位</b>数字，请在 <b>5 分钟</b>内完成验证。</p>
    <p style="margin: 0; font-size: 14px; color: #e74c3c;">⚠️ 请勿将验证码泄露给任何人，平台官方不会主动索要验证码。</p>
  </div>
  <div style="padding: 16px 30px; background: #fafbfc; border-top: 1px solid #eef0f4; color: #9aa0a6; font-size: 12px;">
    <p style="margin: 0;">© %d 戟三电竞平台. 保留所有权利.</p>
  </div>
</div>
</body>
</html>`, kindText, kindText, kindText, code, 2025)
}

// SendVerifyCode 发送验证码邮件
// kind: bind=绑定邮箱, reset=重置密码, forgot=找回账号
func SendVerifyCode(to, code, kind string, cfg *SMTPConfig, log *zap.Logger) error {
	if to == "" {
		return errors.New("收件人不能为空")
	}
	if code == "" {
		return errors.New("验证码不能为空")
	}
	if !ValidateEmail(to) {
		return errors.New("邮箱格式错误")
	}

	actualCfg := cfg
	if actualCfg == nil {
		actualCfg = globalSMTPConf
	}
	actualLog := log
	if actualLog == nil {
		actualLog = smtpLogger
	}

	// 沙箱模式：只写日志不实际发送
	if actualCfg == nil || actualCfg.Sandbox || actualCfg.Host == "" {
		if actualLog != nil {
			actualLog.Info("[SMTP:SANDBOX] 发送验证码邮件（沙箱模式，未真实发送）",
				zap.String("to", to),
				zap.String("kind", kind),
				zap.String("code", code),
			)
		}
		return nil
	}

	subject := fmt.Sprintf("【戟三电竞】您的%s", func() string {
		switch kind {
		case "bind":
			return "邮箱绑定验证码"
		case "reset":
			return "密码重置验证码"
		case "forgot":
			return "找回账号验证码"
		default:
			return "邮箱验证码"
		}
	}())

	htmlBody := buildHTMLTemplate(code, kind)

	from := actualCfg.From
	if from == "" {
		from = actualCfg.User
	}

	return sendEmailTLS(actualCfg.Host, actualCfg.Port, actualCfg.User, actualCfg.Password, from, to, subject, htmlBody)
}

// sendEmailTLS 使用 SMTPS (TLS on connect) 发送邮件（端口465）
func sendEmailTLS(host string, port int, user, password, from, to, subject, htmlBody string) error {
	if host == "" || port <= 0 {
		return errors.New("SMTP host/port 未配置")
	}

	addr := fmt.Sprintf("%s:%d", host, port)

	tlsCfg := &tls.Config{
		ServerName:         host,
		InsecureSkipVerify: false,
	}

	conn, err := tls.Dial("tcp", addr, tlsCfg)
	if err != nil {
		return fmt.Errorf("连接SMTP服务器失败: %w", err)
	}
	defer conn.Close()

	c, err := smtp.NewClient(conn, host)
	if err != nil {
		return fmt.Errorf("创建SMTP客户端失败: %w", err)
	}
	defer c.Quit()

	auth := smtp.PlainAuth("", user, password, host)
	if err := c.Auth(auth); err != nil {
		return fmt.Errorf("SMTP认证失败: %w", err)
	}

	if err := c.Mail(from); err != nil {
		return fmt.Errorf("设置发件人失败: %w", err)
	}
	for _, rcpt := range strings.Split(to, ",") {
		rcpt = strings.TrimSpace(rcpt)
		if rcpt == "" {
			continue
		}
		if err := c.Rcpt(rcpt); err != nil {
			return fmt.Errorf("设置收件人 %s 失败: %w", rcpt, err)
		}
	}

	wc, err := c.Data()
	if err != nil {
		return fmt.Errorf("打开数据写入失败: %w", err)
	}
	defer wc.Close()

	headers := make(map[string]string)
	headers["From"] = from
	headers["To"] = to
	headers["Subject"] = subject
	headers["MIME-Version"] = "1.0"
	headers["Content-Type"] = "text/html; charset=UTF-8"

	var msg strings.Builder
	for k, v := range headers {
		msg.WriteString(k)
		msg.WriteString(": ")
		msg.WriteString(v)
		msg.WriteString("\r\n")
	}
	msg.WriteString("\r\n")
	msg.WriteString(htmlBody)

	if _, err := wc.Write([]byte(msg.String())); err != nil {
		return fmt.Errorf("写入邮件内容失败: %w", err)
	}
	return nil
}
