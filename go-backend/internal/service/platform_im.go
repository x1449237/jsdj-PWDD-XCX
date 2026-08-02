package service

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
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

// localMidnight 返回本地时区当天 00:00:00 (修复 time.Truncate 按 UTC 截断的时区 bug)
func localMidnight(t time.Time) time.Time {
	y, m, d := t.In(time.Local).Date()
	return time.Date(y, m, d, 0, 0, 0, 0, time.Local)
}

// validateBucketOrder 校验工作台板块顺序,防止注入非法 bucket key
func validateBucketOrder(order []string) []string {
	valid := map[string]bool{
		model.TaskBucketEmergency: true,
		model.TaskBucketTodo:      true,
		model.TaskBucketYesterday: true,
	}
	result := make([]string, 0, len(order))
	seen := make(map[string]bool)
	for _, k := range order {
		if valid[k] && !seen[k] {
			result = append(result, k)
			seen[k] = true
		}
	}
	// 确保三个板块都在
	for _, k := range []string{model.TaskBucketEmergency, model.TaskBucketTodo, model.TaskBucketYesterday} {
		if !seen[k] {
			result = append(result, k)
		}
	}
	return result
}

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
		// 14: 2小时以内到达举证时限 (举证时限=最后消息时间+48h,剩<2h即为预警)
		// 修复: 原代码查 last_msg_at < now-2h 逻辑错误,应查 deadline 在 2h 内的
		deadlineWindow := time.Now().Add(2 * time.Hour)
		cond = cond.Where("risk_flag = ? AND last_msg_at IS NOT NULL AND DATE_ADD(last_msg_at, INTERVAL 48 HOUR) < ?", model.RiskFlagTimeout, deadlineWindow)
	case "today_new":
		start := localMidnight(time.Now())
		cond = cond.Where("created_at >= ?", start)
	case "club_only":
		cond = cond.Where("session_type IN ?", []string{model.SessionTypeGroupInternal, model.SessionTypeGroupCategory})
	case "closed":
		cond = cond.Where("risk_flag = ?", model.RiskFlagClosed)
	}
	if q.TimeRange != "" {
		switch q.TimeRange {
		case "yesterday":
			today := localMidnight(time.Now())
			yesterday := today.Add(-24 * time.Hour)
			cond = cond.Where("last_msg_at >= ? AND last_msg_at < ?", yesterday, today)
		case "3d":
			cond = cond.Where("last_msg_at >= ?", localMidnight(time.Now()).Add(-3*24*time.Hour))
		case "7d":
			cond = cond.Where("last_msg_at >= ?", localMidnight(time.Now()).Add(-7*24*time.Hour))
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
		// 修复: ref_id 是 int64,keyword 是字符串,需先判断是否纯数字再匹配,避免 MySQL 隐式转换 "abc"→0 匹配 ref_id=0
		kw := strings.TrimSpace(q.Keyword)
		escapedKw := strings.NewReplacer("\\", "\\\\", "%", "\\%", "_", "\\_").Replace(kw)
		var subCond *gorm.DB
		if n, err := strconv.ParseInt(kw, 10, 64); err == nil {
			// 纯数字: 精准匹配订单号/会话ID/消息内容关键词
			subCond = db.Model(&model.ChatSession{}).Select("id").
				Joins("LEFT JOIN chat_messages m ON m.session_id = chat_sessions.id").
				Where("chat_sessions.ref_id = ? OR chat_sessions.id = ? OR m.content LIKE ? ESCAPE '\\\\' OR EXISTS(SELECT 1 FROM clubs c WHERE c.id = chat_sessions.club_id AND (c.name LIKE ? ESCAPE '\\\\' OR c.abbreviation LIKE ? ESCAPE '\\\\'))",
					n, n, "%"+escapedKw+"%", "%"+escapedKw+"%", "%"+escapedKw+"%")
		} else {
			// 非数字: 不匹配 ref_id,只匹配文本内容
			subCond = db.Model(&model.ChatSession{}).Select("id").
				Joins("LEFT JOIN chat_messages m ON m.session_id = chat_sessions.id").
				Where("m.content LIKE ? ESCAPE '\\\\' OR EXISTS(SELECT 1 FROM clubs c WHERE c.id = chat_sessions.club_id AND (c.name LIKE ? ESCAPE '\\\\' OR c.abbreviation LIKE ? ESCAPE '\\\\'))",
					"%"+escapedKw+"%", "%"+escapedKw+"%", "%"+escapedKw+"%")
		}
		cond = cond.Where("chat_sessions.id IN (?)", subCond)
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
	// 未读计数：直接使用 ChatSession 表中的 unread_count 字段(已在模型中)
	// (原 unreadMap 未实际使用,已移除)
	// 样式徽章(官方是否已回复)
	bubbleMap := make(map[string]*model.ImBubbleStyle)
	{
		var list []*model.ImBubbleStyle
		db.Find(&list)
		// 修复: 原代码遍历 bubbleMap(空map)赋值,导致样式永远为空;应遍历 list
		for _, s := range list {bubbleMap[s.RoleKey] = s}
	}
	_ = bubbleMap // (供后续会话样式判断使用)

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

// PlatformIMSessionTagSet 给会话打标签(增量追加,幂等:重复打标不产生多条)
func PlatformIMSessionTagSet(sessionID, tagID, byUID int64) error {
	var tag model.ImTagDefinition
	if err := db.First(&tag, tagID).Error; err != nil {return err}
	// 修复: 使用 FirstOrCreate 避免重复打标产生多条记录
	return db.Where(&model.ImSessionTag{SessionID: sessionID, TagID: tagID}).
		Assign(model.ImSessionTag{TagName: tag.Name, TagColor: tag.Color, CreatedBy: byUID}).
		FirstOrCreate(&model.ImSessionTag{SessionID: sessionID, TagID: tagID}).Error
}

// PlatformIMSessionTagSetBatch 批量打标签(支持前端传 tag_ids 数组)
func PlatformIMSessionTagSetBatch(sessionID int64, tagIDs []int64, byUID int64) error {
	if len(tagIDs) == 0 {return nil}
	for _, tid := range tagIDs {
		if err := PlatformIMSessionTagSet(sessionID, tid, byUID); err != nil {return err}
	}
	return nil
}

// PlatformIMSessionTagRemove 移除会话标签
func PlatformIMSessionTagRemove(sessionID, tagID int64) error {
	return db.Where("session_id = ? AND tag_id = ?", sessionID, tagID).Delete(&model.ImSessionTag{}).Error
}

// ---------- 备注与星标 ----------

// PlatformIMNoteUpsert 备注/星标写入 (upsert: 存在则更新,不存在则创建)
func PlatformIMNoteUpsert(note *model.ImSessionNote) error {
	now := time.Now()
	note.UpdatedAt = &now
	// 修复: 使用 Where+Assign+FirstOrCreate 确保存在时也更新 Content 和 IsStarred
	return db.Where("session_id = ? AND platform_uid = ?", note.SessionID, note.PlatformUID).
		Assign(model.ImSessionNote{Content: note.Content, IsStarred: note.IsStarred, UpdatedAt: &now}).
		FirstOrCreate(note).Error
}

// PlatformIMSessionMarkClosed 手动标记办结(31),返回影响行数,0=会话不存在
func PlatformIMSessionMarkClosed(sessionID int64) (int64, error) {
	result := db.Model(&model.ChatSession{}).Where("id = ?", sessionID).Updates(map[string]interface{}{
		"risk_flag": model.RiskFlagClosed, "group_bucket": model.BucketClosed,
	})
	return result.RowsAffected, result.Error
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
	// 1) 今日0点 (修复时区 bug)
	today0 := localMidnight(time.Now())
	yesterday0 := today0.Add(-24 * time.Hour)
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
		// 修复: 即将超时 = 举证截止时间(last_msg_at+48h)在2h内
		db.Model(&model.ChatSession{}).
			Where("risk_flag = ? AND last_msg_at IS NOT NULL AND DATE_ADD(last_msg_at, INTERVAL 48 HOUR) < ?", model.RiskFlagTimeout, next2h).
			Count(&c)
		overview.CountTimeout = int(c)
	}

	// 3) 布局自定义
	var layout model.ImWorkbenchLayout
	readDB.Where("platform_uid = ?", uid).First(&layout)
	if len(layout.BucketOrder) > 0 {
		var order []string
		_ = json.Unmarshal(layout.BucketOrder, &order)
		// 修复: 校验 bucket key 合法性,防注入
		if len(order) > 0 {overview.BucketsOrder = validateBucketOrder(order)}
	}

	// 4) emergency = 敏感词预警 + 超时(2h内)
	{
		var list []*model.ChatSession
		// 修复: 超时条件改为 deadline 在 2h 内,而非 last_msg_at < now+2h
		db.Where("risk_flag = ? OR (risk_flag = ? AND last_msg_at IS NOT NULL AND DATE_ADD(last_msg_at, INTERVAL 48 HOUR) < ?)",
			model.RiskFlagSensitive, model.RiskFlagTimeout, next2h).
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

// PlatformIMWorkbenchLayoutSave 保存板块顺序(62),校验合法 bucket key
func PlatformIMWorkbenchLayoutSave(uid int64, order []string) error {
	// 修复: 校验合法 bucket key,防止注入
	safeOrder := validateBucketOrder(order)
	b, _ := json.Marshal(safeOrder)
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
			if fs.OnlySelfClub == 1 {
				// 修复: OnlySelfClub=1 且 scope 为空时也应隐藏(不能默认全显示)
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

// PlatformIMGroupSetting 更新群聊设置(白名单校验,防篡改任意字段)
func PlatformIMGroupSetting(groupID int64, fields map[string]interface{}) error {
	// 修复: 白名单校验,防止注入 status/club_id 等敏感字段
	allowed := map[string]bool{"mute_notify": true, "is_hidden": true, "aggregate_same_club": true}
	safe := make(map[string]interface{})
	for k, v := range fields {
		if allowed[k] {safe[k] = v}
	}
	if len(safe) == 0 {return fmt.Errorf("无可更新字段")}
	return db.Model(&model.GroupChat{}).Where("id = ?", groupID).Updates(safe).Error
}

// PlatformIMGroupBatchMute 批量设免打扰
func PlatformIMGroupBatchMute(groupIDs []int64, mute int8) error {
	// 修复: 空数组会导致 WHERE id IN () SQL 语法错误
	if len(groupIDs) == 0 {
		// apply_silent=1: 自动筛选所有沉寂群(>7天无消息)批量免打扰
		silentThreshold := time.Now().Add(-7 * 24 * time.Hour)
		return db.Model(&model.GroupChat{}).
			Where("last_msg_at < ? OR last_msg_at IS NULL", silentThreshold).
			Update("mute_notify", mute).Error
	}
	return db.Model(&model.GroupChat{}).Where("id IN ?", groupIDs).Update("mute_notify", mute).Error
}
