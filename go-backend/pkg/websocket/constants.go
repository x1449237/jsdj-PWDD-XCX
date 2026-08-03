package websocket

import "time"

// WebSocket 常量定义
const (
	// HeartbeatInterval 心跳间隔 25 秒(Ping-Pong 周期)
	HeartbeatInterval = 25 * time.Second

	// TimeoutDuration 连接超时时间 70 秒(超过该时长未收到 Pong 则断开)
	TimeoutDuration = 70 * time.Second

	// WriteWait 写操作超时，超过该时长未发送成功则断开
	WriteWait = 10 * time.Second

	// MaxMessageSize 单条消息最大字节数 64KB
	MaxMessageSize = 64 * 1024

	// AlertThresholdPercent 连接数告警阈值 80%(达到最大连接数的 80% 告警)
	AlertThresholdPercent = 80

	// ProbeInterval 活跃度探针间隔 5 分钟
	ProbeInterval = 5 * time.Minute

	// ProbeMaxMiss 连续 3 次探针无响应则标记为非活跃
	ProbeMaxMiss = 3

	// MaxConnPerUser 单用户最大并发连接数
	MaxConnPerUser = 3

	// MaxConnections 系统最大连接数(用于告警阈值计算)
	MaxConnections = 100000

	// OfflineMsgTTL 离线消息保留时长 7 天
	OfflineMsgTTL = 7 * 24 * time.Hour

	// MaxOfflineMsgPerUser 单用户离线消息最大保留数量(超过则淘汰最早消息)
	MaxOfflineMsgPerUser = 500
)

// 消息类型常量(对应 message.go 中 Message.Type)
// 架构规则:WebSocket 消息的 type 字段由后端统一定义,小程序前端仅按 type 原样分发,
// 不做任何映射/归一化。下方业务类型常量即为前端 websocket.on(type) 监听的契约字符串。
const (
	// MsgTypeChatMessage 一对一聊天消息(前端监听 chat_message)
	MsgTypeChatMessage = "chat_message"
	// MsgTypeGroupChat 群聊消息(前端监听 group_chat)
	MsgTypeGroupChat = "group_chat"
	// MsgTypeAfterSale 售后会话消息(前端监听 after_sale)
	MsgTypeAfterSale = "after_sale"
	// MsgTypePlatformIntervene 平台介入通知(前端监听 platform_intervene)
	MsgTypePlatformIntervene = "platform_intervene"
	// MsgTypeMessageRecall 消息撤回(前端监听 message_recall)
	MsgTypeMessageRecall = "message_recall"
	// MsgTypeNewMessage 会话列表新消息(前端监听 new_message)
	MsgTypeNewMessage = "new_message"
	// MsgTypeMessageRead 已读回执(前端监听 message_read)
	MsgTypeMessageRead = "message_read"
	// MsgTypeNewOrder 新订单通知(陪玩端监听 new_order)
	MsgTypeNewOrder = "new_order"
	// MsgTypeOrderTaken 订单被接单通知(order_taken)
	MsgTypeOrderTaken = "order_taken"
	// MsgTypeSystem 系统通知
	MsgTypeSystem = "system"
	// MsgTypeOrder 订单状态变更(通用)
	MsgTypeOrder = "order"
	// MsgTypePing 心跳 Ping
	MsgTypePing = "ping"
	// MsgTypePong 心跳 Pong
	MsgTypePong = "pong"
	// MsgTypeProbe 活跃度探针
	MsgTypeProbe = "probe"
	// MsgTypeProbeAck 探针响应
	MsgTypeProbeAck = "probe_ack"
)

// 用户类型常量
const (
	UserTypeUser     = "user"
	UserTypeAdmin    = "admin"
	UserTypePlatform = "platform"
)
