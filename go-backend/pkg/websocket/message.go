package websocket

import (
	"encoding/json"
	"time"
)

// Message WebSocket 消息结构体定义
type Message struct {
	Type      string          `json:"type"`       // 消息类型 chat/group/system/order/ping/pong/probe...
	Data      json.RawMessage `json:"data"`       // 消息数据(按 type 解释)
	SessionID int64           `json:"session_id"` // 会话ID
	SenderID  int64           `json:"sender_id"`  // 发送者ID
	Timestamp int64           `json:"timestamp"`  // 消息时间戳(毫秒)
	ID        string          `json:"id"`         // 消息唯一ID(用于幂等/ACK)
}

// NewMessage 构造一条消息
func NewMessage(msgType string, sessionID, senderID int64, data interface{}) (*Message, error) {
	raw, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}
	return &Message{
		Type:      msgType,
		Data:      raw,
		SessionID: sessionID,
		SenderID:  senderID,
		Timestamp: time.Now().UnixMilli(),
	}, nil
}

// Encode 将消息序列化为字节
func (m *Message) Encode() ([]byte, error) {
	return json.Marshal(m)
}

// DecodeMessage 从字节反序列化消息
func DecodeMessage(data []byte) (*Message, error) {
	msg := &Message{}
	if err := json.Unmarshal(data, msg); err != nil {
		return nil, err
	}
	return msg, nil
}

// ChatPayload 聊天消息载荷
type ChatPayload struct {
	MsgID     string `json:"msg_id"`     // 消息ID
	MsgType   string `json:"msg_type"`   // 消息内容类型 text/image/voice/file
	Content   string `json:"content"`    // 文本内容
	MediaURL  string `json:"media_url"`  // 媒体URL
	SenderName string `json:"sender_name"` // 发送者昵称
	SenderAvatar string `json:"sender_avatar"` // 发送者头像
}

// OrderPayload 订单状态变更载荷
type OrderPayload struct {
	OrderID    int64  `json:"order_id"`    // 订单ID
	OrderNo    string `json:"order_no"`    // 订单号
	FromStatus int8   `json:"from_status"` // 原状态
	ToStatus   int8   `json:"to_status"`   // 新状态
	Title      string `json:"title"`       // 通知标题
}

// SystemPayload 系统通知载荷
type SystemPayload struct {
	Title   string `json:"title"`   // 标题
	Content string `json:"content"` // 内容
	Category string `json:"category"` // 分类
}
