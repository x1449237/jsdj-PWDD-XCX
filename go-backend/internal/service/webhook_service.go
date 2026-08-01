package service

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/jisan/e-sports-platform/internal/model"
	"github.com/jisan/e-sports-platform/pkg/websocket"
)

// HandleWxPayCallback 处理微信支付回调(已校验签名后调用)
// 安全修复:
// 1. 仅 SUCCESS 状态才入账(原非 SUCCESS 也入账)
// 2. 校验回调金额与支付记录金额一致(防 1 分钱攻击)
// 3. 解密失败不再回退到 mock 数据,直接返回错误
// 4. 不再从原始 body 提取 out_trade_no 作为回退(防伪造)
func HandleWxPayCallback(body []byte) error {
	if len(body) == 0 {
		return errors.New("回调内容为空")
	}

	var cb WxPayCallbackRequest
	if err := json.Unmarshal(body, &cb); err != nil {
		return fmt.Errorf("回调 JSON 解析失败: %w", err)
	}

	apiV3Key := ""
	if cfg != nil {
		apiV3Key = cfg.WeChat.ApiV3Key
	}

	decrypted, err := DecryptWxPayResource(cb.Resource, apiV3Key)
	if err != nil {
		return fmt.Errorf("解密 resource 失败: %w", err)
	}

	// 仅 SUCCESS 状态才入账(原逻辑反了,非 SUCCESS 也入账)
	if decrypted.TradeState != "SUCCESS" {
		return fmt.Errorf("交易状态非 SUCCESS(当前:%s),忽略回调", decrypted.TradeState)
	}

	outTradeNo := decrypted.OutTradeNo
	txnID := decrypted.TransactionID
	if outTradeNo == "" {
		return errors.New("回调缺少商户订单号")
	}
	// 金额校验:回调金额传入 MarkPaymentPaid 进行二次校验
	return MarkPaymentPaid(outTradeNo, txnID, decrypted.Amount.Total)
}

// HandleWxPayCallbackWithHeaders 带请求头的微信支付回调处理
// 安全修复:
// 1. 强制校验签名(原忽略签名校验结果)
// 2. 使用 ApiV3Key 解密(原用 MchKey)
// 3. 签名校验失败直接返回错误
func HandleWxPayCallbackWithHeaders(body []byte, headers map[string]string) error {
	if len(body) == 0 {
		return errors.New("回调内容为空")
	}

	timestamp := ""
	nonce := ""
	signature := ""
	serialNo := ""
	for k, v := range headers {
		switch strings.ToLower(k) {
		case "wechatpay-timestamp":
			timestamp = v
		case "wechatpay-nonce":
			nonce = v
		case "wechatpay-signature":
			signature = v
		case "wechatpay-serial":
			serialNo = v
		}
	}

	apiV3Key := ""
	if cfg != nil {
		apiV3Key = cfg.WeChat.ApiV3Key
	}

	// 强制校验签名(原忽略校验结果)
	if !VerifyWxPaySignature(timestamp, nonce, string(body), signature, serialNo, apiV3Key) {
		return errors.New("微信支付回调签名校验失败")
	}

	return HandleWxPayCallback(body)
}

func parseWxPayNotify(body []byte) (outTradeNo, txnID string) {
	s := string(body)
	outTradeNo = extractJSONField(s, "out_trade_no")
	txnID = extractJSONField(s, "transaction_id")
	return
}

func extractJSONField(s, field string) string {
	key := "\"" + field + "\":"
	idx := strings.Index(s, key)
	if idx < 0 {
		return ""
	}
	start := idx + len(key)
	for start < len(s) && (s[start] == ' ' || s[start] == '"') {
		start++
	}
	end := start
	for end < len(s) && s[end] != '"' && s[end] != ',' && s[end] != '}' {
		end++
	}
	return s[start:end]
}

func ProxyWxPayCallback(r *http.Request) error {
	if r == nil || r.Body == nil {
		return errors.New("请求体为空")
	}
	body, err := io.ReadAll(r.Body)
	if err != nil {
		return err
	}
	headers := make(map[string]string)
	for k, vs := range r.Header {
		if len(vs) > 0 {
			headers[k] = vs[0]
		}
	}
	return HandleWxPayCallbackWithHeaders(body, headers)
}

// DecryptWxPayResourceAESGCM 使用 AES-GCM 解密微信支付回调资源
// 安全修复:解密失败不再回退到 mock,直接返回错误
func DecryptWxPayResourceAESGCM(r WxPayResource, apiV3Key string) (*WxPayDecryptedResource, error) {
	if apiV3Key == "" {
		return nil, errors.New("apiV3Key 未配置")
	}
	if len(apiV3Key) < 32 {
		return nil, errors.New("apiV3Key 长度不足 32 字节")
	}
	keyBytes := []byte(apiV3Key)[:32]

	ciphertext, err := base64.StdEncoding.DecodeString(r.Ciphertext)
	if err != nil {
		return nil, fmt.Errorf("ciphertext base64 解码失败: %w", err)
	}

	nonce := []byte(r.Nonce)
	if len(nonce) < 12 {
		padded := make([]byte, 12)
		copy(padded, nonce)
		nonce = padded
	}
	if len(nonce) > 12 {
		nonce = nonce[:12]
	}

	associatedData := []byte(r.AssociatedData)

	block, err := aes.NewCipher(keyBytes)
	if err != nil {
		return nil, fmt.Errorf("创建 AES cipher 失败: %w", err)
	}

	aesGCM, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("创建 GCM 失败: %w", err)
	}

	if len(ciphertext) < aesGCM.Overhead() {
		return nil, errors.New("密文长度异常，缺少 GCM tag")
	}

	plaintext, err := aesGCM.Open(nil, nonce, ciphertext, associatedData)
	if err != nil {
		// 安全修复:解密失败直接返回错误,不再回退到 mock
		return nil, fmt.Errorf("AES-GCM 解密失败(密文可能被篡改): %w", err)
	}

	var result WxPayDecryptedResource
	if err := json.Unmarshal(plaintext, &result); err != nil {
		return nil, fmt.Errorf("解析解密后 JSON 失败: %w", err)
	}
	return &result, nil
}

// VerifyWxPaySignatureV3 微信支付 V3 签名校验
// 安全修复:原恒返回 true,现改为:
// 1. 沙箱模式跳过校验
// 2. 生产模式未配置公钥则拒绝
// TODO: 接入 wechatpay-go SDK 进行真正的 RSA 签名校验
func VerifyWxPaySignatureV3(timestamp, nonce, body, signature, serialNo, publicKeyPEM string) bool {
	if timestamp == "" || nonce == "" || signature == "" {
		return false
	}
	// 沙箱模式跳过
	if cfg != nil && cfg.WeChat.MchID != "" && strings.HasPrefix(cfg.WeChat.MchID, "sandbox") {
		return true
	}
	// 生产模式未配置公钥则拒绝(安全失败)
	if publicKeyPEM == "" {
		return false
	}
	// TODO: 接入真正 RSA 验签
	return false
}

func sha256Hex(s string) string {
	h := sha256.Sum256([]byte(s))
	return hex.EncodeToString(h[:])
}

func constantTimeEq(a, b string) bool {
	if len(a) != len(b) {
		return false
	}
	var v byte
	for i := 0; i < len(a); i++ {
		v |= a[i] ^ b[i]
	}
	return v == 0
}

func NotifyOrderStatusChanged(o *model.Order) {
	if hub != nil && o.UserID > 0 {
		_ = pushOrderNotification(o)
	}
	_ = db.Create(&model.Notification{
		UserID: o.UserID, Type: "order",
		Title:     "订单状态更新",
		Content:   "您的订单 " + o.OrderNo + " 状态已更新",
		IsRead:    0, Category: model.NotificationCategorySystem,
		CreatedAt: nowTimePtr(), UpdatedAt: nowTimePtr(),
	}).Error
}

func pushOrderNotification(o *model.Order) error {
	ctx, cancel := contextWithTimeout()
	defer cancel()
	msg, err := websocket.NewMessage(websocket.MsgTypeOrder, 0, o.UserID, map[string]interface{}{
		"order_id": o.ID, "order_no": o.OrderNo,
		"status": o.Status,
	})
	if err != nil {
		return err
	}
	return hub.SendToUser(ctx, o.UserID, msg)
}

func unpadPKCS7(data []byte, blockSize int) ([]byte, error) {
	if blockSize <= 0 {
		return nil, errors.New("invalid block size")
	}
	if len(data)%blockSize != 0 || len(data) == 0 {
		return nil, fmt.Errorf("invalid data len %d", len(data))
	}
	padLen := int(data[len(data)-1])
	if padLen == 0 || padLen > blockSize {
		return nil, fmt.Errorf("invalid padding length: %d", padLen)
	}
	pad := data[len(data)-padLen:]
	for i := 0; i < padLen; i++ {
		if pad[i] != byte(padLen) {
			return nil, errors.New("invalid padding")
		}
	}
	return data[:len(data)-padLen], nil
}

func padPKCS7(data []byte, blockSize int) []byte {
	padLen := blockSize - len(data)%blockSize
	padding := bytes.Repeat([]byte{byte(padLen)}, padLen)
	return append(data, padding...)
}
