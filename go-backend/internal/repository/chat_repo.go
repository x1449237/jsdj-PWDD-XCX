package repository

import (
	"errors"

	"gorm.io/gorm"

	"github.com/jisan/e-sports-platform/internal/model"
)

// ChatRepo IM 聊天数据访问仓储
type ChatRepo struct {
	db *gorm.DB
}

// NewChatRepo 创建聊天仓储
func NewChatRepo(db *gorm.DB) *ChatRepo {
	return &ChatRepo{db: db}
}

// FindSessionByID 根据ID查询会话
func (r *ChatRepo) FindSessionByID(id int64) (*model.ChatSession, error) {
	var s model.ChatSession
	if err := r.db.First(&s, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &s, nil
}

// FindOrCreateOrderSession 查询或创建订单会话
func (r *ChatRepo) FindOrCreateOrderSession(orderID int64) (*model.ChatSession, error) {
	var s model.ChatSession
	err := r.db.Where("session_type = ? AND ref_id = ?", model.SessionTypeOrder, orderID).First(&s).Error
	if err == nil {
		return &s, nil
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}
	s = model.ChatSession{
		SessionType: model.SessionTypeOrder,
		RefID:       orderID,
		Status:      1,
		CreatedAt:   nowPtr(),
		UpdatedAt:   nowPtr(),
	}
	if err := r.db.Create(&s).Error; err != nil {
		return nil, err
	}
	return &s, nil
}

// ListUserSessions 查询用户参与的会话列表
// 通过消息发送者/接收者维度聚合，简化实现：返回该用户最近活跃的订单会话
func (r *ChatRepo) ListUserSessions(userID int64, page, pageSize int) ([]model.ChatSession, int64, error) {
	var sessions []model.ChatSession
	var total int64
	q := r.db.Model(&model.ChatSession{}).Where("status = 1").
		Joins("JOIN chat_messages cm ON cm.session_id = chat_sessions.id").
		Where("cm.sender_id = ? OR cm.session_id IN (SELECT session_id FROM chat_messages WHERE sender_id = ?)", userID, userID).
		Group("chat_sessions.id")
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	err := q.Scopes(Paginate(page, pageSize)).Order("chat_sessions.updated_at DESC").Find(&sessions).Error
	return sessions, total, err
}

// CreateMessage 创建聊天消息
func (r *ChatRepo) CreateMessage(m *model.ChatMessage) error {
	return r.db.Create(m).Error
}

// ListMessages 分页查询会话消息
func (r *ChatRepo) ListMessages(sessionID int64, page, pageSize int) ([]model.ChatMessage, int64, error) {
	var msgs []model.ChatMessage
	var total int64
	q := r.db.Model(&model.ChatMessage{}).Where("session_id = ?", sessionID)
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	err := q.Scopes(Paginate(page, pageSize)).Order("id DESC").Find(&msgs).Error
	return msgs, total, err
}

// FindMessage 查询单条消息
func (r *ChatRepo) FindMessage(id int64) (*model.ChatMessage, error) {
	var m model.ChatMessage
	if err := r.db.First(&m, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &m, nil
}

// RevokeMessage 撤回消息(2 分钟内)
func (r *ChatRepo) RevokeMessage(id int64) error {
	return r.db.Model(&model.ChatMessage{}).Where("id = ?", id).
		Update("is_revoked", 1).Error
}

// MarkRead 标记会话消息已读
func (r *ChatRepo) MarkRead(sessionID, userID int64) error {
	return r.db.Model(&model.ChatMessage{}).
		Where("session_id = ? AND sender_id <> ? AND is_read = 0", sessionID, userID).
		Update("is_read", 1).Error
}

// CreateFile 创建聊天文件记录
func (r *ChatRepo) CreateFile(f *model.ChatFile) error {
	return r.db.Create(f).Error
}

// ListKeywords 查询售后风控关键词
func (r *ChatRepo) ListKeywords() ([]model.AfterSaleKeyword, error) {
	var list []model.AfterSaleKeyword
	err := r.db.Where("enabled = 1").Find(&list).Error
	return list, err
}

// CreateKeyword 新增关键词
func (r *ChatRepo) CreateKeyword(k *model.AfterSaleKeyword) error {
	return r.db.Create(k).Error
}

// CreateGroupChat 创建群聊
func (r *ChatRepo) CreateGroupChat(g *model.GroupChat) error {
	return r.db.Create(g).Error
}

// ListGroups 查询俱乐部群聊
func (r *ChatRepo) ListGroups(clubID int64) ([]model.GroupChat, error) {
	var list []model.GroupChat
	q := r.db.Where("status = 1")
	if clubID > 0 {
		q = q.Where("club_id = ?", clubID)
	}
	err := q.Order("id DESC").Find(&list).Error
	return list, err
}

// FindGroup 查询群详情
func (r *ChatRepo) FindGroup(id int64) (*model.GroupChat, error) {
	var g model.GroupChat
	if err := r.db.First(&g, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &g, nil
}

// UpdateGroupAnnouncement 更新群公告
func (r *ChatRepo) UpdateGroupAnnouncement(id int64, announcement string) error {
	return r.db.Model(&model.GroupChat{}).Where("id = ?", id).
		Updates(map[string]interface{}{"announcement": announcement, "announcement_at": nowPtr()}).Error
}

// ListGroupMembers 查询群成员
func (r *ChatRepo) ListGroupMembers(groupID int64) ([]model.GroupChatMember, error) {
	var list []model.GroupChatMember
	err := r.db.Where("group_id = ?", groupID).Find(&list).Error
	return list, err
}

// CreateGroupMessage 创建群消息
func (r *ChatRepo) CreateGroupMessage(m *model.GroupChatMessage) error {
	return r.db.Create(m).Error
}

// ListRiskSessions 查询含风险消息的会话
func (r *ChatRepo) ListRiskSessions(page, pageSize int) ([]model.ChatSession, int64, error) {
	var sessions []model.ChatSession
	var total int64
	q := r.db.Model(&model.ChatSession{}).
		Joins("JOIN chat_messages cm ON cm.session_id = chat_sessions.id").
		Where("cm.risk_level >= 2").Group("chat_sessions.id")
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	err := q.Scopes(Paginate(page, pageSize)).Order("chat_sessions.updated_at DESC").Find(&sessions).Error
	return sessions, total, err
}
