package service

import (
	"context"
	"errors"
	"strings"
	"sync"
	"time"

	"github.com/jisan/e-sports-platform/internal/model"
	"github.com/jisan/e-sports-platform/pkg/websocket"
)

// GetChatSessions 用户会话列表
func GetChatSessions(userID int64, page, pageSize int) ([]model.ChatSession, int64, error) {
	return chatRepo.ListUserSessions(userID, page, pageSize)
}

// GetChatMessages 会话消息列表(校验会话参与权)
// 安全修复:补齐所有会话类型的归属校验(原仅校验订单会话,售后/群聊会话任意用户可读)
func GetChatMessages(sessionID, userID int64, page, pageSize int) ([]model.ChatMessage, int64, error) {
	s, err := chatRepo.FindSessionByID(sessionID)
	if err != nil {
		return nil, 0, err
	}
	if s == nil {
		return nil, 0, errors.New("会话不存在")
	}
	// 校验会话参与权(所有会话类型)
	if err := verifySessionAccess(s, userID); err != nil {
		return nil, 0, err
	}
	// 标记已读
	_ = chatRepo.MarkRead(sessionID, userID)
	return chatRepo.ListMessages(sessionID, page, pageSize)
}

// verifySessionAccess 校验用户是否有权访问该会话
// 订单会话:校验客户/打手归属
// 售后会话:校验订单参与方 + 俱乐部客服身份
// 群聊会话:校验群成员身份
func verifySessionAccess(s *model.ChatSession, userID int64) error {
	switch s.SessionType {
	case model.SessionTypeOrder:
		o, _ := orderRepo.FindByID(s.RefID)
		if o == nil || (o.UserID != userID && o.PlayerID != userID) {
			return errors.New("无权查看该会话消息")
		}
	case model.SessionTypeAfterSale:
		// 售后会话:校验订单参与方
		o, _ := orderRepo.FindByID(s.RefID)
		if o != nil && (o.UserID == userID || o.PlayerID == userID) {
			return nil // 订单参与方放行
		}
		// 俱乐部客服:校验是否为该俱乐部 shop_admin
		if s.ClubID > 0 {
			var cnt int64
			_ = db.Model(&model.ShopAdminAccount{}).
				Where("club_id = ? AND user_id = ? AND status = 1", s.ClubID, userID).Count(&cnt).Error
			if cnt > 0 {
				return nil
			}
		}
		return errors.New("无权查看该售后会话")
	case model.SessionTypeGroupInternal, model.SessionTypeGroupCategory:
		// 群聊:校验群成员身份
		var cnt int64
		_ = db.Model(&model.GroupChatMember{}).
			Where("group_id = ? AND user_id = ? AND status = 1", s.RefID, userID).Count(&cnt).Error
		if cnt == 0 {
			return errors.New("非群成员,无权查看群聊消息")
		}
	default:
		// 安全修复:未知会话类型默认拒绝(原直接 return nil 放行,存在越权读取风险)
		return errors.New("不支持的会话类型,无权访问")
	}
	return nil
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
// 安全修复:校验会话归属,防止向任意会话注入消息
func SendMessage(userID int64, senderType string, in *SendMessageInput) (*model.ChatMessage, error) {
	if in.Content == "" && in.MediaURL == "" {
		return nil, errors.New("消息内容不能为空")
	}
	if in.MsgType == "" {
		in.MsgType = model.MsgTypeText
	}
	// 语音消息时长校验:MP3 最大 60 秒
	if in.MsgType == model.MsgTypeVoice {
		if in.Duration <= 0 {
			return nil, errors.New("语音消息时长不能为空")
		}
		if in.Duration > 60 {
			return nil, errors.New("语音消息时长不能超过 60 秒")
		}
	}
	// 会话归属校验:防止向任意会话注入消息
	s, err := chatRepo.FindSessionByID(in.SessionID)
	if err != nil {
		return nil, err
	}
	if s == nil {
		return nil, errors.New("会话不存在")
	}
	if err := verifySessionAccess(s, userID); err != nil {
		return nil, err
	}
	// 防代练检测(聊天消息)
	if in.Content != "" {
		hit, patterns, abErr := CheckContentAntiBoosting(AntiBoostingContentTypeChat, userID, in.Content)
		if abErr == nil && hit {
			// 命中拦截消息 + 双方提示
			_ = patterns
			return nil, errors.New("消息包含违规关键词，已被拦截")
		}
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

	// WebSocket 推送给会话所有参与者（所有 session_type 全量覆盖）
	if hub != nil {
		if s != nil {
			switch s.SessionType {
			case model.SessionTypeOrder:
				o, _ := orderRepo.FindByID(s.RefID)
				if o != nil {
					peerID := o.UserID
					if peerID == userID {
						peerID = o.PlayerID
					}
					if peerID > 0 {
						_ = pushChatMessage(peerID, m)
					}
				}
			case model.SessionTypeAfterSale:
				// 售后会话：推送给订单相关成员(user/player) + 客服(若有)
				o, _ := orderRepo.FindByID(s.RefID)
				if o != nil {
					if o.UserID > 0 && o.UserID != userID {
						_ = pushChatMessage(o.UserID, m)
					}
					if o.PlayerID > 0 && o.PlayerID != userID {
						_ = pushChatMessage(o.PlayerID, m)
					}
				}
				// 客服/平台处理人（简化：所有管理员在线用户也可配置，此处仅推 session 明确成员）
			case model.SessionTypeGroupInternal, model.SessionTypeGroupCategory:
				// 群聊：遍历群成员，对在线成员调用 hub.SendToUser
				members, _ := chatRepo.ListGroupMembers(s.RefID)
				for _, mb := range members {
					if mb.UserID > 0 && mb.UserID != userID {
						_ = pushChatMessage(mb.UserID, m)
					}
				}
			default:
				// 未知会话类型兜底:通过 refID 关联订单找对端（无 owner 字段直接跳过）
				o, _ := orderRepo.FindByID(s.RefID)
				if o != nil {
					if o.UserID > 0 && o.UserID != userID {
						_ = pushChatMessage(o.UserID, m)
					}
					if o.PlayerID > 0 && o.PlayerID != userID {
						_ = pushChatMessage(o.PlayerID, m)
					}
				}
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
// - 高风险消息(risk_level>=3,命中敏感词)禁止撤回,以保留审计证据
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
	// 命中敏感词的高风险消息禁止撤回(留存审计)
	if m.RiskLevel >= 3 {
		return errors.New("该消息含敏感内容,已被风控留存,不可撤回")
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
// - 首次创建售后会话时推送通知给俱乐部创始人/内置管理端
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
	newlyCreated := false
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
		as = model.AfterSaleSession{
			OrderID: s.RefID, UserID: userID, ClubID: clubID,
			Status: 1, InterventionStatus: model.AfterSaleInterventionPending,
			InterventionType: "manual", CreatedAt: nowTimePtr(), UpdatedAt: nowTimePtr(),
		}
		if err := db.Create(&as).Error; err == nil {
			newlyCreated = true
		}
	}
	// 记录平台介入日志
	_ = db.Create(&model.PlatformInterventionLog{
		SessionID: sessionID, OrderID: s.RefID,
		TriggerType: model.InterventionTriggerManual, HandlerID: userID,
		Result: reason, CreatedAt: nowTimePtr(),
	}).Error
	// 首次创建售后会话:推送通知给俱乐部创始人/管理端
	if newlyCreated && as.ClubID > 0 {
		c, _ := clubRepo.FindByID(as.ClubID)
		if c != nil && c.FounderUID > 0 {
			_ = AdminSendNotification(c.FounderUID, "after_sale", "新售后申请介入",
				"您有新的售后会话申请平台介入,请及时处理。", model.NotificationCategoryPending)
		}
	}
	return nil
}

// scanMessageRisk 消息风险扫描(敏感词 + 售后关键词)
// 返回风险等级 0无 1低 2中 3高
// 使用缓存避免每次发送消息全表扫描关键词
func scanMessageRisk(content string) int8 {
	if content == "" {
		return 0
	}
	// 从缓存加载关键词(每 5 分钟刷新一次)
	keywords := loadCachedKeywords()
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
	// 敏感词扫描(同样走缓存)
	sws := loadCachedSensitiveWords()
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

// 关键词缓存(避免每次消息全表扫描)
var (
	cachedKeywords           []model.AfterSaleKeyword
	cachedSensitiveWords     []model.SensitiveWord
	cachedKeywordsExpireAt   time.Time
	cachedSensitiveExpireAt  time.Time
	keywordsCacheMu          sync.Mutex
)

const cacheRefreshInterval = 5 * time.Minute

// loadCachedKeywords 加载缓存的聊天关键词
func loadCachedKeywords() []model.AfterSaleKeyword {
	keywordsCacheMu.Lock()
	defer keywordsCacheMu.Unlock()
	if time.Now().Before(cachedKeywordsExpireAt) && cachedKeywords != nil {
		return cachedKeywords
	}
	ks, _ := chatRepo.ListKeywords()
	if ks != nil {
		cachedKeywords = ks
		cachedKeywordsExpireAt = time.Now().Add(cacheRefreshInterval)
	}
	return cachedKeywords
}

// loadCachedSensitiveWords 加载缓存的敏感词
func loadCachedSensitiveWords() []model.SensitiveWord {
	keywordsCacheMu.Lock()
	defer keywordsCacheMu.Unlock()
	if time.Now().Before(cachedSensitiveExpireAt) && cachedSensitiveWords != nil {
		return cachedSensitiveWords
	}
	var sws []model.SensitiveWord
	_ = db.Where("enabled = 1").Find(&sws).Error
	if sws != nil {
		cachedSensitiveWords = sws
		cachedSensitiveExpireAt = time.Now().Add(cacheRefreshInterval)
	}
	return cachedSensitiveWords
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
