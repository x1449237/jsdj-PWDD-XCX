package service

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/jisan/e-sports-platform/internal/model"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

// ============================================================
// 平台方管理人员 IM 会话归类清单 - 业务服务层
// 全部函数使用 service 层统一的 db/readDB 全局变量
// ============================================================

// ---------- 会话列表：按优先级自动排序 + 分组 ----------

// PlatformIMSessionsQuery 会话列表查询参数
type PlatformIMSessionsQuery struct {
	PlatformUID int64  // 当前平台账号UID
	TabKey      string // 筛选 tab (11~17): all/risk/player_int/timeo_out2h/today_new/club_only/closed
	TimeRange   string // 18: yesterday/3d/7d
	ClubKeyword string // 19: 俱乐部缩写/名称
	GameID      int    // 20: 游戏品类ID
	Keyword     string // 搜索关键词 (21~24)
	TagID       int64  // 按标签筛选 (30)
	OnlyStarred bool   // 仅看星标
	Page        int
	PageSize    int
}

// PlatformIMSessionsResult 会话查询结果
type PlatformIMSessionsResult struct {
	Groups []*SessBucketGroup `json:"groups"` // 分组结果(按顺序)
	Total  int64              `json:"total"`
}

// SessBucketGroup 一个分组
type SessBucketGroup struct {
	BucketKey   string                `json:"bucket_key"`
	BucketName  string                `json:"bucket_name"`
	IsCollapsed bool                  `json:"is_collapsed"` // 默认折叠(沉寂)
	Items       []*SessionListItem    `json:"items"`
}

// SessionListItem 单条会话
type SessionListItem struct {
	SessionID         int64                `json:"session_id"`
	SessionType       string               `json:"session_type"`
	PriorityLevel     int8                 `json:"priority_level"`
	RiskFlag          int8                 `json:"risk_flag"`
	Title             string               `json:"title"`
	SubTitle          string               `json:"sub_title"` // 订单金额/下单时间/订单状态 (35)
	RemarkTags        []string             `json:"remark_tags"`
	LastMsgPreview    string               `json:"last_msg_preview"`
	LastMsgAt         *time.Time           `json:"last_msg_at"`
	UnreadCount       int                  `json:"unread_count"`
	UnreadBold        int8                 `json:"unread_bold"` // 高风险未读加粗(40)
	OfficialHasReply  int8                 `json:"official_has_reply"` // 官方是否已回复
	OfficialBadge     string               `json:"official_badge"` // 官方已介入(76)
	OfficialBadgeColor string             `json:"official_badge_color"`
	HornFlag          int8                 `json:"horn_flag"` // 金色小喇叭(75)
	GroupLabel        string               `json:"group_label"` // 群类型闲聊/福利/售后(39)
	ClubID            int64                `json:"club_id"`
	ClubName          string               `json:"club_name"`
	RefOrderID        int64                `json:"ref_order_id"`
	OrderAmount       int64                `json:"order_amount"`
	OrderStatus       int8                 `json:"order_status"`
	OrderCreatedAt    *time.Time           `json:"order_created_at"`
	EvidenceTimeout   int64                `json:"evidence_timeout"` // 举证倒计时秒(38)
	TriggerReasonText string               `json:"trigger_reason_text"` // 36/37 预警原因
	NotePreview       string               `json:"note_preview"` // 33 备注预览前几字
	IsStarred         int8                 `json:"is_starred"`
	Tags              []*TagBrief          `json:"tags"` // 29 小色块标签
	SenderStyleKey    string               `json:"-"` // 供样式匹配用
}

// TagBrief 标签摘要
type TagBrief struct {
	ID    int64  `json:"id"`
	Name  string `json:"name"`
	Color string `json:"color"`
}

// bucketDisplay 分组展示配置
var bucketDisplay = []struct {
	Key         string
	Name        string
	Collapsed   bool
}{
	{model.BucketRiskTop,    "⚠️ 风险预警会话（置顶）", false},
	{model.BucketPlayerInt,  "玩家申请官方介入",         false},
	{model.BucketTimeout,    "举证超时倒计时",          false},
	{model.BucketNormalSale, "普通售后（新消息优先）",    false},
	{model.BucketClubChat,   "俱乐部群聊",              false},
	{model.BucketStarred,    "⭐ 稍后处理（星标）",       false},
	{model.BucketSilentSale, "沉寂会话（>7天无新消息）", true},
	{model.BucketSilentClub, "长期沉寂群聊（>15天）",    true},
	{model.BucketClosed,     "已办结会话",              false},
}

// PlatformIMListSessions 按优先级返回会话分组列表
func PlatformIMListSessions(q *PlatformIMSessionsQuery) (*PlatformIMSessionsResult, error) {
	if q.Page < 1 {q.Page = 1}
	if q.PageSize < 1 {q.PageSize = 20}

	// 1) 条件构造
	cond := db.Model(&model.ChatSession{}).Where("1=1")

	switch q.TabKey {
	case "risk":
		cond = cond.Where("risk_flag = ?", model.RiskFlagSensitive)
	case "player_int":
		cond = cond.Where("risk_flag = ?", model.RiskFlagPlayerInt)
	case "timeo_out2h":
		cond = cond.Where("risk_flag = ? AND last_msg_at < ?", model.RiskFlagTimeout, time.Now().Add(-2*time.Hour))
	case "today_new":
		start := time.Now().Truncate(24*time.Hour)
		cond = cond.Where("created_at >= ?", start)
	case "club_only":
		cond = cond.Where("session_type IN ?", []string{model.SessionTypeGroupInternal, model.SessionTypeGroupCategory})
	case "closed":
		cond = cond.Where("risk_flag = ?", model.RiskFlagClosed)
	}
	if q.TimeRange != "" {
		switch q.TimeRange {
		case "yesterday":
			today := time.Now().Truncate(24*time.Hour)
			yesterday := today.Add(-24*time.Hour)
			cond = cond.Where("last_msg_at BETWEEN ? AND ?", yesterday, today)
		case "3d":
			cond = cond.Where("last_msg_at >= ?", time.Now().Add(-3*24*time.Hour))
		case "7d":
			cond = cond.Where("last_msg_at >= ?", time.Now().Add(-7*24*time.Hour))
		}
	}
	if q.ClubKeyword != "" {
		// 匹配 club id by name/abbr
		var cids []int64
		db.Model(&model.Club{}).
			Where("name LIKE ? OR abbreviation LIKE ?", "%"+q.ClubKeyword+"%", "%"+q.ClubKeyword+"%").
			Pluck("id", &cids)
		if len(cids) > 0 {
			cond = cond.Where("club_id IN ?", cids)
		} else {
			cond = cond.Where("1=0")
		}
	}
	if q.GameID > 0 {
		cond = cond.Where("game_id = ?", q.GameID)
	}
	if q.Keyword != "" {
		// 22~24: 订单号/俱乐部/昵称/聊天关键词
		sub := db.Model(&model.ChatSession{}).Select("id").
			Joins("LEFT JOIN chat_messages m ON m.session_id = chat_sessions.id").
			Where("chat_sessions.ref_id = ? OR m.content LIKE ? OR EXISTS(SELECT 1 FROM clubs c WHERE c.id = chat_sessions.club_id AND (c.name LIKE ? OR c.abbreviation LIKE ?))",
				q.Keyword, "%"+q.Keyword+"%", "%"+q.Keyword+"%", "%"+q.Keyword+"%")
		cond = cond.Where("chat_sessions.id IN (?)", sub)
	}
	if q.OnlyStarred {
		cond = cond.Where("EXISTS(SELECT 1 FROM im_session_notes sn WHERE sn.session_id = chat_sessions.id AND sn.platform_uid = ? AND sn.is_starred = 1)", q.PlatformUID)
	}

	// 2) 拉取会话(基础字段) + 总条数
	var total int64
	if err := cond.Count(&total).Error; err != nil {
		return nil, fmt.Errorf("count sessions: %w", err)
	}

	var rows []*model.ChatSession
	if err := cond.Order("priority_level DESC, last_msg_at DESC").
		Limit(q.PageSize).Offset((q.Page-1)*q.PageSize).Find(&rows).Error; err != nil {
		return nil, fmt.Errorf("load sessions: %w", err)
	}

	// 3) 收集关联数据
	sessIDs := make([]int64, 0, len(rows))
	clubIDs := make([]int64, 0)
	orderIDs := make([]int64, 0)
	for _, s := range rows {
		sessIDs = append(sessIDs, s.ID)
		if s.ClubID > 0 {clubIDs = append(clubIDs, s.ClubID)}
		if s.RefID > 0 && (s.SessionType == model.SessionTypeOrder || s.SessionType == model.SessionTypeAfterSale) {
			orderIDs = append(orderIDs, s.RefID)
		}
	}
	// 备注与星标(仅当前平台账号)
	notes := make(map[int64]*model.ImSessionNote)
	{
		var list []*model.ImSessionNote
		db.Where("platform_uid = ? AND session_id IN ?", q.PlatformUID, sessIDs).Find(&list)
		for _, n := range list {notes[n.SessionID] = n}
	}
	// 标签
	tagMap := make(map[int64][]*model.ImSessionTag)
	{
		var list []*model.ImSessionTag
		db.Where("session_id IN ?", sessIDs).Find(&list)
		for _, t := range list {tagMap[t.SessionID] = append(tagMap[t.SessionID], t)}
	}
	// 俱乐部名称
	clubMap := make(map[int64]string)
	if len(clubIDs) > 0 {
		var cl []*model.Club
		db.Select("id,name").Where("id IN ?", clubIDs).Find(&cl)
		for _, c := range cl {clubMap[c.ID] = c.Name}
	}
	// 订单基础信息(35 金额/下单时间/状态)
	type orderBrief struct {ID int64; Amount int64; Status int8; CreatedAt *time.Time}
	orderMap := make(map[int64]*orderBrief)
	if len(orderIDs) > 0 {
		var ol []*orderBrief
		db.Model(&model.Order{}).Select("id,amount,status,created_at").Where("id IN ?", orderIDs).Scan(&ol)
		for _, o := range ol {orderMap[o.ID] = o}
	}
	// 未读计数：若表里已有，就用；缺失时实时查
	unreadMap := make(map[int64]int)
	// 样式徽章(官方是否已回复)
	bubbleMap := make(map[string]*model.ImBubbleStyle)
	{
		var list []*model.ImBubbleStyle
		db.Find(&list)
		for _, s := range bubbleMap {bubbleMap[s.RoleKey] = s}
	}
	_ = unreadMap
	_ = bubbleMap

	// 4) 按 bucket 拼装
	items := make(map[string][]*SessionListItem)
	for _, s := range rows {
		bucket := s.GroupBucket
		// 星标单独抽一份到 Starred 桶
		var li SessionListItem
		li.SessionID = s.ID
		li.SessionType = s.SessionType
		li.PriorityLevel = s.PriorityLevel
		li.RiskFlag = s.RiskFlag
		li.LastMsgPreview = s.LastMsgPreview
		li.LastMsgAt = s.LastMsgAt
		li.UnreadCount = s.UnreadCount
		if s.PriorityLevel >= 3 {li.UnreadBold = 1}
		li.ClubID = s.ClubID
		li.ClubName = clubMap[s.ClubID]
		li.OfficialHasReply = s.OfficialHasReply
		if s.OfficialHasReply == 1 {
			li.OfficialBadge = "官方已介入"
			li.OfficialBadgeColor = "#409EFF"
		}
		if s.HasOfficialNotice == 1 {li.HornFlag = 1}
		li.GroupLabel = sessGroupLabel(s.SessionType)

		// 订单
		if ob, ok := orderMap[s.RefID]; ok {
			li.RefOrderID = ob.ID
			li.OrderAmount = ob.Amount
			li.OrderStatus = ob.Status
			li.OrderCreatedAt = ob.CreatedAt
		}

		// 标签
		if tags, ok := tagMap[s.ID]; ok {
			for _, t := range tags {
				li.Tags = append(li.Tags, &TagBrief{ID: t.TagID, Name: t.TagName, Color: t.TagColor})
			}
		}
		if q.TagID > 0 {
			hit := false
			for _, t := range li.Tags {if t.ID == q.TagID {hit = true}}
			if !hit {continue} // 标签筛选未命中
		}
		// 备注预览
		if n, ok := notes[s.ID]; ok {
			if len(n.Content) > 20 {li.NotePreview = string([]rune(n.Content)[:20]) + "…"} else {li.NotePreview = n.Content}
			if n.IsStarred == 1 {li.IsStarred = 1}
		}
		// 预警原因文案
		switch s.RiskFlag {
		case model.RiskFlagSensitive:
			li.TriggerReasonText = "【敏感词触发·强制介入】"
		case model.RiskFlagPlayerInt:
			li.TriggerReasonText = "【买家申请官方介入】"
		case model.RiskFlagTimeout:
			// 举证倒计时:2小时 - (now - last_msg_at)
			if s.LastMsgAt != nil {
				deadline := s.LastMsgAt.Add(2 * time.Hour)
				left := time.Until(deadline)
				if left < 0 {left = 0}
				li.EvidenceTimeout = int64(left.Seconds())
			}
		}
		// 标题：售后→订单号+俱乐部；群→群名
		switch s.SessionType {
		case model.SessionTypeOrder, model.SessionTypeAfterSale:
			if li.RefOrderID > 0 {li.Title = fmt.Sprintf("订单 #%d", li.RefOrderID)} else {li.Title = "订单售后会话"}
			if li.ClubName != "" {li.SubTitle = li.ClubName}
		default:
			li.Title = "群聊会话"
			if li.ClubName != "" {li.SubTitle = li.ClubName}
		}
		items[bucket] = append(items[bucket], &li)
		if li.IsStarred == 1 {
			items[model.BucketStarred] = append(items[model.BucketStarred], &li)
		}
	}

	groups := make([]*SessBucketGroup, 0, len(bucketDisplay))
	for _, b := range bucketDisplay {
		list := items[b.Key]
		if len(list) == 0 && b.Key != model.BucketStarred {continue}
		groups = append(groups, &SessBucketGroup{
			BucketKey: b.Key, BucketName: b.Name, IsCollapsed: b.Collapsed, Items: list,
		})
	}

	return &PlatformIMSessionsResult{Groups: groups, Total: total}, nil
}

// sessGroupLabel 辅助：返回群聊类型文字（不可给外部 model 包加方法，只能包内函数）
func sessGroupLabel(st string) string {
	switch st {
	case model.SessionTypeGroupInternal, model.SessionTypeGroupCategory:
		return "俱乐部群聊"
	}
	return ""
}

// ---------- 会话标签 ----------

// PlatformIMTagList 获取标签定义
func PlatformIMTagList() ([]*model.ImTagDefinition, error) {
	var list []*model.ImTagDefinition
	if err := readDB.Order("is_system DESC, sort ASC, id ASC").Find(&list).Error; err != nil {
		return nil, err
	}
	return list, nil
}

// PlatformIMTagCreate 创建标签
func PlatformIMTagCreate(t *model.ImTagDefinition) error {
	return db.Create(t).Error
}

// PlatformIMTagUpdate 更新标签
func PlatformIMTagUpdate(id int64, name, color string) error {
	return db.Model(&model.ImTagDefinition{}).Where("id = ?", id).UpdateColumns(map[string]interface{}{
		"name": name, "color": color,
	}).Error
}

// PlatformIMTagDelete 删除标签(非系统内置)
func PlatformIMTagDelete(id int64) error {
	return db.Where("id = ? AND is_system = 0", id).Delete(&model.ImTagDefinition{}).Error
}

// PlatformIMSessionTagSet 给会话打标签(增量追加)
func PlatformIMSessionTagSet(sessionID, tagID, byUID int64) error {
	var tag model.ImTagDefinition
	if err := db.First(&tag, tagID).Error; err != nil {return err}
	return db.Create(&model.ImSessionTag{
		SessionID: sessionID, TagID: tagID, TagName: tag.Name, TagColor: tag.Color, CreatedBy: byUID,
	}).Error
}

// PlatformIMSessionTagRemove 移除会话标签
func PlatformIMSessionTagRemove(sessionID, tagID int64) error {
	return db.Where("session_id = ? AND tag_id = ?", sessionID, tagID).Delete(&model.ImSessionTag{}).Error
}

// ---------- 备注与星标 ----------

// PlatformIMNoteUpsert 备注/星标写入 (31 标记办结就是把 status=0 + risk_flag=4)
func PlatformIMNoteUpsert(note *model.ImSessionNote) error {
	now := time.Now()
	note.UpdatedAt = &now
	return db.Where("session_id = ? AND platform_uid = ?", note.SessionID, note.PlatformUID).
		Assign(*note).FirstOrCreate(note).Error
}

// PlatformIMSessionMarkClosed 手动标记办结(31)
func PlatformIMSessionMarkClosed(sessionID int64) error {
	return db.Model(&model.ChatSession{}).Where("id = ?", sessionID).Updates(map[string]interface{}{
		"risk_flag": model.RiskFlagClosed, "group_bucket": model.BucketClosed,
	}).Error
}

// ---------- 搜索历史 ----------

// PlatformIMSearchHistoryList 获取最近10条
func PlatformIMSearchHistoryList(uid int64) ([]*model.ImSearchHistory, error) {
	var list []*model.ImSearchHistory
	err := readDB.Where("platform_uid = ?", uid).Order("created_at DESC").Limit(10).Find(&list).Error
	return list, err
}

// PlatformIMSearchHistoryAdd 追加(超出10条自动清理老的)
func PlatformIMSearchHistoryAdd(h *model.ImSearchHistory) error {
	now := time.Now()
	h.CreatedAt = &now
	if err := db.Create(h).Error; err != nil {return err}
	// 清理超出10条老记录
	sub := db.Model(&model.ImSearchHistory{}).Select("id").
		Where("platform_uid = ?", h.PlatformUID).
		Order("created_at DESC").Limit(10)
	db.Where("platform_uid = ? AND id NOT IN (?)", h.PlatformUID, sub).Delete(&model.ImSearchHistory{})
	return nil
}

// PlatformIMSearchHistoryClear 清空
func PlatformIMSearchHistoryClear(uid int64) error {
	return db.Where("platform_uid = ?", uid).Delete(&model.ImSearchHistory{}).Error
}

// ---------- 工作台 ----------

// PlatformIMWorkbenchOverview 工作台统计+三大板块(47~52)
type PlatformIMWorkbenchOverview struct {
	CountEmergency int                     `json:"count_emergency"`
	CountNewToday  int                     `json:"count_new_today"`
	CountTimeout   int                     `json:"count_timeout"`
	BucketsOrder   []string                `json:"buckets_order"` // 自定义板块顺序
	Buckets        map[string][]*WorkbenchItem `json:"buckets"`
}

type WorkbenchItem struct {
	TaskID    int64      `json:"task_id"`
	SessionID int64      `json:"session_id"`
	Title     string     `json:"title"`
	TaskType  string     `json:"task_type"`
	Deadline  *time.Time `json:"deadline"`
}

// PlatformIMWorkbenchGet 获取工作台数据
func PlatformIMWorkbenchGet(uid int64) (*PlatformIMWorkbenchOverview, error) {
	// 1) 今日0点
	today0 := time.Now().Truncate(24*time.Hour)
	yesterday0 := today0.Add(-24*time.Hour)
	next2h := time.Now().Add(2 * time.Hour)

	// 2) 统计
	overview := &PlatformIMWorkbenchOverview{
		BucketsOrder: []string{model.TaskBucketEmergency, model.TaskBucketTodo, model.TaskBucketYesterday},
		Buckets:      make(map[string][]*WorkbenchItem),
	}
	var em, nt int64
	db.Model(&model.ChatSession{}).Where("risk_flag = ? AND last_msg_at >= ?", model.RiskFlagSensitive, today0).Count(&em)
	overview.CountEmergency = int(em)
	db.Model(&model.ChatSession{}).Where("created_at >= ?", today0).Count(&nt)
	overview.CountNewToday = int(nt)
	{
		var c int64
		db.Model(&model.ChatSession{}).Where("risk_flag = ? AND last_msg_at < ?", model.RiskFlagTimeout, next2h).Count(&c)
		overview.CountTimeout = int(c)
	}

	// 3) 布局自定义
	var layout model.ImWorkbenchLayout
	readDB.Where("platform_uid = ?", uid).First(&layout)
	if len(layout.BucketOrder) > 0 {
		var order []string
		_ = json.Unmarshal(layout.BucketOrder, &order)
		if len(order) > 0 {overview.BucketsOrder = order}
	}

	// 4) emergency = 敏感词预警 + 超时(2h内)
	{
		var list []*model.ChatSession
		db.Where("(risk_flag = ?) OR (risk_flag = ? AND last_msg_at < ?)", model.RiskFlagSensitive, model.RiskFlagTimeout, next2h).
			Order("priority_level DESC, last_msg_at DESC").Limit(50).Find(&list)
		for _, s := range list {
			overview.Buckets[model.TaskBucketEmergency] = append(overview.Buckets[model.TaskBucketEmergency],
				&WorkbenchItem{SessionID: s.ID, Title: fmt.Sprintf("会话#%d", s.ID), TaskType: "emergency"})
		}
	}
	// 5) todo = 玩家介入未办结
	{
		var list []*model.ChatSession
		db.Where("risk_flag = ?", model.RiskFlagPlayerInt).Order("last_msg_at DESC").Limit(50).Find(&list)
		for _, s := range list {
			overview.Buckets[model.TaskBucketTodo] = append(overview.Buckets[model.TaskBucketTodo],
				&WorkbenchItem{SessionID: s.ID, Title: fmt.Sprintf("会话#%d", s.ID), TaskType: "player_int"})
		}
	}
	// 6) yesterday 遗留
	{
		var list []*model.ChatSession
		db.Where("last_msg_at BETWEEN ? AND ? AND risk_flag IN ?", yesterday0, today0, []int8{
			model.RiskFlagNone, model.RiskFlagPlayerInt, model.RiskFlagTimeout,
		}).Order("last_msg_at DESC").Limit(50).Find(&list)
		for _, s := range list {
			overview.Buckets[model.TaskBucketYesterday] = append(overview.Buckets[model.TaskBucketYesterday],
				&WorkbenchItem{SessionID: s.ID, Title: fmt.Sprintf("会话#%d", s.ID), TaskType: "yesterday"})
		}
	}
	return overview, nil
}

// PlatformIMWorkbenchLayoutSave 保存板块顺序(62)
func PlatformIMWorkbenchLayoutSave(uid int64, order []string) error {
	b, _ := json.Marshal(order)
	now := time.Now()
	return db.Where(model.ImWorkbenchLayout{PlatformUID: uid}).Assign(model.ImWorkbenchLayout{
		BucketOrder: datatypes.JSON(b), UpdatedAt: &now,
	}).FirstOrCreate(&model.ImWorkbenchLayout{}).Error
}

// ---------- 快捷话术(53) ----------

// PlatformIMQuickReplyList 按分类列出
func PlatformIMQuickReplyList(category string) ([]*model.ImQuickReply, error) {
	q := readDB.Model(&model.ImQuickReply{})
	if category != "" {q = q.Where("category = ?", category)}
	var list []*model.ImQuickReply
	err := q.Order("is_system DESC, sort ASC, id ASC").Find(&list).Error
	return list, err
}

// PlatformIMQuickReplyCreate 创建
func PlatformIMQuickReplyCreate(r *model.ImQuickReply) error {
	return db.Create(r).Error
}

// PlatformIMQuickReplyUpdate 更新
func PlatformIMQuickReplyUpdate(id int64, title, content, category string) error {
	return db.Model(&model.ImQuickReply{}).Where("id = ?", id).Updates(map[string]interface{}{
		"title": title, "content": content, "category": category,
	}).Error
}

// PlatformIMQuickReplyDelete 删除
func PlatformIMQuickReplyDelete(id int64) error {
	return db.Where("id = ? AND is_system = 0", id).Delete(&model.ImQuickReply{}).Error
}

// ---------- 证据打包预览(55) ----------

// PlatformIMEvidenceMessages 获取会话内某段时间的证据消息
func PlatformIMEvidenceMessages(sessionID int64, from, to *time.Time) ([]*model.ChatMessage, error) {
	q := db.Where("session_id = ?", sessionID)
	if from != nil {q = q.Where("created_at >= ?", from)}
	if to != nil {q = q.Where("created_at <= ?", to)}
	var list []*model.ChatMessage
	err := q.Order("created_at ASC").Find(&list).Error
	return list, err
}

// ---------- 样式下发 (93~98)：平台端IM打开时，按账号权限拉专属样式 ----------

// PlatformIMStyleForUser 获取指定用户会话页应使用的样式(按授权表实时取)
type PlatformIMStyleResult struct {
	BubbleStyle      *model.ImBubbleStyle      `json:"bubble_style"`
	AvatarFrameStyle *model.ImAvatarFrameStyle `json:"avatar_frame_style"`
	ScopeClubIDs     []int64                   `json:"scope_club_ids"` // 头像框生效俱乐部
}

// PlatformIMPullMyStyle 当前平台账号拉取自己专属样式(用于渲染自己发送的消息)
func PlatformIMPullMyStyle(uid int64, currentClubID int64) (*PlatformIMStyleResult, error) {
	var grant model.ImStyleGrant
	err := db.Where("user_id = ? AND (expires_at IS NULL OR expires_at > ?)", uid, time.Now()).First(&grant).Error
	if err != nil && err != gorm.ErrRecordNotFound {return nil, err}
	res := &PlatformIMStyleResult{}
	if grant.BubbleRoleKey != "" {
		var bs model.ImBubbleStyle
		if er2 := db.Where("role_key = ?", grant.BubbleRoleKey).First(&bs).Error; er2 == nil {
			res.BubbleStyle = &bs
		}
	}
	if grant.AvatarFrameKey != "" {
		var fs model.ImAvatarFrameStyle
		if er2 := db.Where("frame_key = ?", grant.AvatarFrameKey).First(&fs).Error; er2 == nil {
			// scope 校验
			var clubs []int64
			_ = json.Unmarshal(grant.ScopeClubIDs, &clubs)
			if fs.OnlySelfClub == 1 && len(clubs) > 0 {
				hit := false
				for _, c := range clubs {if c == currentClubID {hit = true}}
				if !hit {fs = model.ImAvatarFrameStyle{} /* 跨俱乐部不显示 */}
			}
			if fs.ID > 0 {res.AvatarFrameStyle = &fs}
			res.ScopeClubIDs = clubs
		}
	}
	return res, nil
}

// PlatformIMPullAllStyles 一次拉取所有角色的样式(聊天页多角色渲染)
func PlatformIMPullAllStyles() (map[string]*model.ImBubbleStyle, map[string]*model.ImAvatarFrameStyle, error) {
	var bs []*model.ImBubbleStyle
	var fs []*model.ImAvatarFrameStyle
	if err := readDB.Find(&bs).Error; err != nil {return nil, nil, err}
	if err := readDB.Find(&fs).Error; err != nil {return nil, nil, err}
	bm := make(map[string]*model.ImBubbleStyle)
	for _, s := range bs {bm[s.RoleKey] = s}
	fm := make(map[string]*model.ImAvatarFrameStyle)
	for _, s := range fs {fm[s.FrameKey] = s}
	return bm, fm, nil
}

// ---------- 群聊免打扰/隐藏/同俱乐部聚合 (41~45) ----------

// PlatformIMGroupSetting 更新群聊设置
func PlatformIMGroupSetting(groupID int64, fields map[string]interface{}) error {
	return db.Model(&model.GroupChat{}).Where("id = ?", groupID).Updates(fields).Error
}

// PlatformIMGroupBatchMute 批量设免打扰
func PlatformIMGroupBatchMute(groupIDs []int64, mute int8) error {
	return db.Model(&model.GroupChat{}).Where("id IN ?", groupIDs).Update("mute_notify", mute).Error
}
