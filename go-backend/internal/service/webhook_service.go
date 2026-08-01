package service

import (
	"errors"
	"io"
	"net/http"
	"strings"

	"github.com/jisan/e-sports-platform/internal/model"
	"github.com/jisan/e-sports-platform/pkg/websocket"
)

// HandleWxPayCallback 处理微信支付回调通知
// 解析回调内容并标记支付成功
func HandleWxPayCallback(body []byte) error {
	if len(body) == 0 {
		return errors.New("回调内容为空")
	}
	// 简化解析:实际项目应校验签名、解密 resource
	outTradeNo, txnID := parseWxPayNotify(body)
	if outTradeNo == "" {
		return errors.New("回调缺少商户订单号")
	}
	return MarkPaymentPaid(outTradeNo, txnID)
}

// parseWxPayNotify 简化解析微信支付回调(实际应解密)
// 仅做关键字段提取占位
func parseWxPayNotify(body []byte) (outTradeNo, txnID string) {
	s := string(body)
	outTradeNo = extractJSONField(s, "out_trade_no")
	txnID = extractJSONField(s, "transaction_id")
	return
}

// extractJSONField 简易 JSON 字段提取(避免引入额外依赖)
func extractJSONField(s, field string) string {
	key := "\"" + field + "\":"
	idx := strings.Index(s, key)
	if idx < 0 {
		return ""
	}
	start := idx + len(key)
	// 跳过空白与引号
	for start < len(s) && (s[start] == ' ' || s[start] == '"') {
		start++
	}
	end := start
	for end < len(s) && s[end] != '"' && s[end] != ',' && s[end] != '}' {
		end++
	}
	return s[start:end]
}

// ProxyWxPayCallback HTTP 层代理:读取请求体并交给 service 处理
func ProxyWxPayCallback(r *http.Request) error {
	if r == nil || r.Body == nil {
		return errors.New("请求体为空")
	}
	body, err := io.ReadAll(r.Body)
	if err != nil {
		return err
	}
	return HandleWxPayCallback(body)
}

// NotifyOrderStatusChanged 订单状态变更通知(WebSocket 推送 + 站内消息)
func NotifyOrderStatusChanged(o *model.Order) {
	// 站内通知用户
	if hub != nil && o.UserID > 0 {
		_ = pushOrderNotification(o)
	}
	// 写入通知表
	_ = db.Create(&model.Notification{
		UserID: o.UserID, Type: "order",
		Title:    "订单状态更新",
		Content:  "您的订单 " + o.OrderNo + " 状态已更新",
		IsRead:   0, Category: model.NotificationCategorySystem,
		CreatedAt: nowTimePtr(), UpdatedAt: nowTimePtr(),
	}).Error
}

// pushOrderNotification 通过 WebSocket 推送订单状态变更
func pushOrderNotification(o *model.Order) error {
	ctx, cancel := contextWithTimeout()
	defer cancel()
	msg, err := websocket.NewMessage(websocket.MsgTypeOrder, 0, o.UserID, map[string]interface{}{
		"order_id": o.ID, "order_no": o.OrderNo,
		"status":   o.Status,
	})
	if err != nil {
		return err
	}
	return hub.SendToUser(ctx, o.UserID, msg)
}
