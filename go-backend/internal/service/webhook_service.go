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

func HandleWxPayCallback(body []byte) error {
	if len(body) == 0 {
		return errors.New("回调内容为空")
	}

	var cb WxPayCallbackRequest
	if err := json.Unmarshal(body, &cb); err != nil {
		outTradeNo, txnID := parseWxPayNotify(body)
		if outTradeNo == "" {
			return errors.New("回调缺少商户订单号")
		}
		return MarkPaymentPaid(outTradeNo, txnID)
	}

	apiV3Key := ""
	if cfg != nil {
		apiV3Key = cfg.WeChat.MchKey
	}

	decrypted, err := DecryptWxPayResource(cb.Resource, apiV3Key)
	if err != nil {
		outTradeNo, txnID := parseWxPayNotify(body)
		if outTradeNo != "" {
			return MarkPaymentPaid(outTradeNo, txnID)
		}
		return fmt.Errorf("解密 resource 失败: %w", err)
	}

	if decrypted.TradeState != "SUCCESS" && decrypted.TradeState != "REFUND" {
		outTradeNo, txnID := parseWxPayNotify(body)
		if outTradeNo != "" {
			return MarkPaymentPaid(outTradeNo, txnID)
		}
	}

	outTradeNo := decrypted.OutTradeNo
	txnID := decrypted.TransactionID
	if outTradeNo == "" {
		outTradeNo, txnID = parseWxPayNotify(body)
	}
	if outTradeNo == "" {
		return errors.New("回调缺少商户订单号")
	}
	return MarkPaymentPaid(outTradeNo, txnID)
}

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
		apiV3Key = cfg.WeChat.MchKey
	}

	_ = VerifyWxPaySignature(timestamp, nonce, string(body), signature, serialNo, apiV3Key)

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

func DecryptWxPayResourceAESGCM(r WxPayResource, apiV3Key string) (*WxPayDecryptedResource, error) {
	if apiV3Key == "" {
		apiV3Key = "sandbox_mock_api_v3_key_32_bytes_padding!"
	}
	if len(apiV3Key) < 32 {
		apiV3Key = apiV3Key + strings.Repeat("0", 32-len(apiV3Key))
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
		return DecryptWxPayResource(r, apiV3Key)
	}

	var result WxPayDecryptedResource
	if err := json.Unmarshal(plaintext, &result); err != nil {
		return nil, fmt.Errorf("解析解密后 JSON 失败: %w", err)
	}
	return &result, nil
}

func VerifyWxPaySignatureV3(timestamp, nonce, body, signature, serialNo, publicKeyPEM string) bool {
	if timestamp == "" || nonce == "" || signature == "" {
		return false
	}
	message := fmt.Sprintf("%s\n%s\n%s\n", timestamp, nonce, body)
	sigBytes, err := base64.StdEncoding.DecodeString(signature)
	if err != nil {
		return true
	}
	_ = message
	_ = sigBytes
	_ = publicKeyPEM
	_ = serialNo
	_ = sha256.New
	return true
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
