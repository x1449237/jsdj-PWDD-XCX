package service

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/jisan/e-sports-platform/internal/model"
	"github.com/jisan/e-sports-platform/pkg/websocket"
)

// GetChatSessions 用户会话列表
func GetChatSessions(userID int64, page, pageSize int) ([]model.ChatSession, int64, error) {
	return chatRepo.ListUserSessions(userID, page, pageSize)
}

// GetChatMessages 会话消息列表(校验会话参与权)
func GetChatMessages(sessionID, userID int64, page, pageSize int) ([]model.ChatMessage, int64, error) {
	s, err := chatRepo.FindSessionByID(sessionID)
	if err != nil {
		return nil, 0, err
	}
	if s == nil {
		return nil, 0, errors.New("会话不存在")
	}
	// 订单会话:校验客户/打手归属
	if s.SessionType == model.SessionTypeOrder {
		o, _ := orderRepo.FindByID(s.RefID)
		if o == nil || (o.UserID != userID && o.PlayerID != userID) {
			return nil, 0, errors.New("无权查看该会话消息")
		}
	}
	// 标记已读
	_ = chatRepo.MarkRead(sessionID, userID)
	return chatRepo.ListMessages(sessionID, page, pageSize)
}

// SendMessageInput 发送消息入参
type SendMessageInput struct {
	SessionID  int64  `json:"session_id"`
	MsgType    string `json:"msg_type"`
	Content    string `json:"content"`
	MediaURL   string `json:"media_url"`
	Duration   int    `json:"duration"`
	AsrText    string `json:"asr_text"`
}

// SendMessage 发送聊天消息(含敏感词风控、WebSocket 推送)
func SendMessage(userID int64, senderType string, in *SendMessageInput) (*model.ChatMessage, error) {
	if in.Content == "" && in.MediaURL == "" {
		return nil, errors.New("消息内容不能为空")
	}
	if in.MsgType == "" {
		in.MsgType = model.MsgTypeText
	}
	// 敏感词与售后关键词风控扫描
	riskLevel := scanMessageRisk(in.Content)

	m := &model.ChatMessage{
		SessionID:  in.SessionID,
		SenderID:   userID,
		SenderType: senderType,
		MsgType:    in.MsgType,
		Content:    in.Content,
		MediaURL:   in.MediaURL,
		Duration:   in.Duration,
		AsrText:    in.AsrText,
		IsRead:     0,
		IsRevoked:  0,
		RiskLevel:  riskLevel,
		CreatedAt:  nowTimePtr(),
	}
	if err := chatRepo.CreateMessage(m); err != nil {
		return nil, err
	}

	// WebSocket 推送给会话其他参与者
	if hub != nil {
		s, _ := chatRepo.FindSessionByID(in.SessionID)
		if s != nil && s.SessionType == model.SessionTypeOrder {
			o, _ := orderRepo.FindByID(s.RefID)
			if o != nil {
				peerID := o.UserID
				if peerID == userID {
					peerID = o.PlayerID
				}
				_ = pushChatMessage(peerID, m)
			}
		}
	}
	return m, nil
}

// pushChatMessage 通过 WebSocket 推送聊天消息
func pushChatMessage(userID int64, m *model.ChatMessage) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	msg, err := websocket.NewMessage(websocket.MsgTypeChat, m.SessionID, m.SenderID, map[string]interface{}{
		"msg_id":   itoa(m.ID),
		"msg_type": m.MsgType,
		"content":  m.Content,
		"media_url": m.MediaURL,
	})
	if err != nil {
		return err
	}
	return hub.SendToUser(ctx, userID, msg)
}

// RevokeMessage 撤回消息(2 分钟内)
func RevokeMessage(messageID, userID int64) error {
	m, err := chatRepo.FindMessage(messageID)
	if err != nil {
		return err
	}
	if m == nil {
		return errors.New("消息不存在")
	}
	if m.SenderID != userID {
		return errors.New("无权撤回他人消息")
	}
	if m.IsRevoked == 1 {
		return errors.New("消息已撤回")
	}
	// 2 分钟时效校验
	if m.CreatedAt != nil && time.Since(*m.CreatedAt) > 2*time.Minute {
		return errors.New("超过 2 分钟不可撤回")
	}
	return chatRepo.RevokeMessage(messageID)
}

// UploadChatFile 上传聊天文件(记录元数据)
func UploadChatFile(sessionID, uploaderID int64, fileURL, fileName string, fileSize int64, fileType string) (*model.ChatFile, error) {
	f := &model.ChatFile{
		SessionID:  sessionID,
		UploaderID: uploaderID,
		FileURL:    fileURL,
		FileName:   fileName,
		FileSize:   fileSize,
		FileType:   fileType,
		CreatedAt:  nowTimePtr(),
	}
	if err := chatRepo.CreateFile(f); err != nil {
		return nil, err
	}
	return f, nil
}

// RequestIntervention 用户请求平台介入(售后会话)
func RequestIntervention(sessionID, userID int64, reason string) error {
	s, err := chatRepo.FindSessionByID(sessionID)
	if err != nil {
		return err
	}
	if s == nil {
		return errors.New("会话不存在")
	}
	// 升级为售后会话(若未关联)
	var as model.AfterSaleSession
	if err := db.Where("order_id = ?", s.RefID).First(&as).Error; err == nil {
		_ = db.Model(&as).Updates(map[string]interface{}{
			"intervention_status": model.AfterSaleInterventionPending,
			"intervention_type":   "manual",
		}).Error
	} else {
		o, _ := orderRepo.FindByID(s.RefID)
		clubID := int64(0)
		if o != nil {
			clubID = o.ClubID
		}
		_ = db.Create(&model.AfterSaleSession{
			OrderID: s.RefID, UserID: userID, ClubID: clubID,
			Status: 1, InterventionStatus: model.AfterSaleInterventionPending,
			InterventionType: "manual", CreatedAt: nowTimePtr(), UpdatedAt: nowTimePtr(),
		}).Error
	}
	// 记录平台介入日志
	_ = db.Create(&model.PlatformInterventionLog{
		SessionID: sessionID, OrderID: s.RefID,
		TriggerType: model.InterventionTriggerManual, HandlerID: userID,
		Result: reason, CreatedAt: nowTimePtr(),
	}).Error
	return nil
}

// scanMessageRisk 消息风险扫描(敏感词 + 售后关键词)
// 返回风险等级 0无 1低 2中 3高
func scanMessageRisk(content string) int8 {
	if content == "" {
		return 0
	}
	// 加载关键词(简化:每条消息实时查表，实际应缓存)
	keywords, _ := chatRepo.ListKeywords()
	risk := int8(0)
	for _, kw := range keywords {
		if kw.MatchType == model.KeywordMatchExact {
			if content == kw.Keyword {
				risk = 2
				break
			}
		} else {
			if strings.Contains(content, kw.Keyword) {
				if risk < 2 {
					risk = 2
				}
			}
		}
	}
	// 敏感词扫描
	var sws []model.SensitiveWord
	_ = db.Where("enabled = 1").Find(&sws).Error
	for _, sw := range sws {
		if strings.Contains(content, sw.Word) {
			if risk < 3 {
				risk = 3
			}
			break
		}
	}
	return risk
}

// AdminGetChatAuditList 聊天审计列表(按会话维度)
func AdminGetChatAuditList(page, pageSize int) ([]model.ChatSession, int64, error) {
	var list []model.ChatSession
	var total int64
	q := db.Model(&model.ChatSession{})
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	err := q.Scopes(Paginate2(page, pageSize)).Order("id DESC").Find(&list).Error
	return list, total, err
}

// AdminGetChatMessages 管理端查看会话消息(审计)
func AdminGetChatMessages(sessionID int64, page, pageSize int) ([]model.ChatMessage, int64, error) {
	return chatRepo.ListMessages(sessionID, page, pageSize)
}

// AdminGetRiskSessions 风险会话列表
func AdminGetRiskSessions(page, pageSize int) ([]model.ChatSession, int64, error) {
	return chatRepo.ListRiskSessions(page, pageSize)
}

// AdminAddChatKeyword 新增聊天关键词
func AdminAddChatKeyword(keyword, matchType string) error {
	return chatRepo.CreateKeyword(&model.AfterSaleKeyword{
		Keyword:   keyword,
		MatchType: matchType,
		Enabled:   1,
		CreatedAt: nowTimePtr(),
		UpdatedAt: nowTimePtr(),
	})
}

// AdminGetChatKeywords 关键词列表
func AdminGetChatKeywords(page, pageSize int) ([]model.AfterSaleKeyword, int64, error) {
	var list []model.AfterSaleKeyword
	var total int64
	q := db.Model(&model.AfterSaleKeyword{})
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	err := q.Scopes(Paginate2(page, pageSize)).Order("id DESC").Find(&list).Error
	return list, total, err
}
