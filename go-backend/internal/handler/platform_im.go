package handler

import (
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jisan/e-sports-platform/internal/model"
	"github.com/jisan/e-sports-platform/internal/service"
	"github.com/jisan/e-sports-platform/internal/utils"
)

// PlatformIMHandler 平台方 IM 会话归类 - HTTP 接口处理器
type PlatformIMHandler struct{}

// NewPlatformIMHandler 构造
func NewPlatformIMHandler() *PlatformIMHandler { return &PlatformIMHandler{} }

// 当前平台人员UID (修复: 直接复用 utils.GetCurrentUserID,与全局中间件一致)
func (h *PlatformIMHandler) getUID(c *gin.Context) int64 {
	return utils.GetCurrentUserID(c)
}

// ---------- 会话列表 + 排序筛选 (1~26) ----------

// ListSessions GET /platform-im/sessions  (分页+分组)
func (h *PlatformIMHandler) ListSessions(c *gin.Context) {
	q := &service.PlatformIMSessionsQuery{
		PlatformUID: h.getUID(c),
		TabKey:      c.DefaultQuery("tab", "all"),
		TimeRange:   c.Query("time_range"),
		ClubKeyword: c.Query("club_keyword"),
		Keyword:     c.Query("keyword"),
		Page:        int(parseIntQuery(c, "page", 1)),
		PageSize:    int(parseIntQuery(c, "page_size", 20)),
	}
	if v, e := strconv.Atoi(c.DefaultQuery("game_id", "0")); e == nil {q.GameID = v}
	if v, e := strconv.ParseInt(c.DefaultQuery("tag_id", "0"), 10, 64); e == nil {q.TagID = v}
	if c.Query("starred") == "1" {q.OnlyStarred = true}
	// 搜索历史写入
	if q.Keyword != "" {
		_ = service.PlatformIMSearchHistoryAdd(&model.ImSearchHistory{
			PlatformUID: q.PlatformUID, Keyword: q.Keyword, SearchType: c.DefaultQuery("search_type", "all"),
		})
	}
	res, err := service.PlatformIMListSessions(q)
	if err != nil {utils.Fail(c, utils.CodeServerError, err.Error()); return}
	utils.SuccessWithTotal(c, res.Groups, res.Total)
}

// MarkSessionClosed POST /platform-im/sessions/:id/close  标记办结 (31)
func (h *PlatformIMHandler) MarkSessionClosed(c *gin.Context) {
	id := parseInt64Path(c, "id")
	if id <= 0 {utils.Fail(c, utils.CodeBadRequest, "会话ID无效"); return}
	affected, err := service.PlatformIMSessionMarkClosed(id)
	if err != nil {utils.Fail(c, utils.CodeServerError, err.Error()); return}
	if affected == 0 {utils.Fail(c, utils.CodeNotFound, "会话不存在"); return}
	utils.Success(c, gin.H{"msg": "ok"})
}

// ---------- 会话标签 (27~30) ----------

// ListTags GET /platform-im/tags
func (h *PlatformIMHandler) ListTags(c *gin.Context) {
	list, err := service.PlatformIMTagList()
	if err != nil {utils.Fail(c, utils.CodeServerError, err.Error()); return}
	utils.Success(c, list)
}

// CreateTag POST /platform-im/tags
func (h *PlatformIMHandler) CreateTag(c *gin.Context) {
	var t model.ImTagDefinition
	if err := c.ShouldBindJSON(&t); err != nil {
		utils.Fail(c, utils.CodeBadRequest, "参数错误: "+err.Error()); return
	}
	t.CreatedBy = h.getUID(c)
	if t.Color == "" {t.Color = "#409EFF"}
	if err := service.PlatformIMTagCreate(&t); err != nil {
		utils.Fail(c, utils.CodeBadRequest, err.Error()); return
	}
	utils.Success(c, t)
}

// UpdateTag PUT /platform-im/tags/:id
func (h *PlatformIMHandler) UpdateTag(c *gin.Context) {
	var req struct {Name string `json:"name"`; Color string `json:"color"`}
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Fail(c, utils.CodeBadRequest, "参数错误"); return
	}
	id := parseInt64Path(c, "id")
	if err := service.PlatformIMTagUpdate(id, req.Name, req.Color); err != nil {
		utils.Fail(c, utils.CodeBadRequest, err.Error()); return
	}
	utils.Success(c, gin.H{"msg": "ok"})
}

// DeleteTag DELETE /platform-im/tags/:id
func (h *PlatformIMHandler) DeleteTag(c *gin.Context) {
	id := parseInt64Path(c, "id")
	if err := service.PlatformIMTagDelete(id); err != nil {
		utils.Fail(c, utils.CodeBadRequest, err.Error()); return
	}
	utils.Success(c, gin.H{"msg": "ok"})
}

// TagSession POST /platform-im/sessions/:id/tags  (支持批量 tag_ids + star_flag)
func (h *PlatformIMHandler) TagSession(c *gin.Context) {
	sid := parseInt64Path(c, "id")
	if sid <= 0 {utils.Fail(c, utils.CodeBadRequest, "会话ID无效"); return}
	// 修复: 兼容前端传 tag_id(单个) 和 tag_ids(批量)
	var req struct {
		TagID    int64   `json:"tag_id"`
		TagIDs   []int64 `json:"tag_ids"`
		StarFlag *int8   `json:"star_flag"`
	}
	if err := c.ShouldBindJSON(&req); err != nil && err.Error() != "EOF" {
		utils.Fail(c, utils.CodeBadRequest, "参数错误: "+err.Error()); return
	}
	uid := h.getUID(c)
	// 打标签
	if req.TagID > 0 {req.TagIDs = append(req.TagIDs, req.TagID)}
	if len(req.TagIDs) > 0 {
		if err := service.PlatformIMSessionTagSetBatch(sid, req.TagIDs, uid); err != nil {
			utils.Fail(c, utils.CodeBadRequest, err.Error()); return
		}
	}
	// 星标切换
	if req.StarFlag != nil {
		if err := service.PlatformIMNoteUpsert(&model.ImSessionNote{
			SessionID: sid, PlatformUID: uid, IsStarred: *req.StarFlag,
		}); err != nil {utils.Fail(c, utils.CodeBadRequest, err.Error()); return}
	}
	utils.Success(c, gin.H{"msg": "ok"})
}

// UntagSession DELETE /platform-im/sessions/:id/tags/:tag_id
func (h *PlatformIMHandler) UntagSession(c *gin.Context) {
	sid := parseInt64Path(c, "id")
	tid := parseInt64Path(c, "tag_id")
	if err := service.PlatformIMSessionTagRemove(sid, tid); err != nil {
		utils.Fail(c, utils.CodeBadRequest, err.Error()); return
	}
	utils.Success(c, gin.H{"msg": "ok"})
}

// ---------- 备注与星标 (32~34) ----------

// UpsertNote PUT /platform-im/sessions/:id/note  (兼容前端 note_text 和 content)
func (h *PlatformIMHandler) UpsertNote(c *gin.Context) {
	sid := parseInt64Path(c, "id")
	if sid <= 0 {utils.Fail(c, utils.CodeBadRequest, "会话ID无效"); return}
	// 修复: 兼容前端传 note_text 或 content
	var req struct {
		Content  string `json:"content"`
		NoteText string `json:"note_text"`
		IsStarred int8  `json:"is_starred"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Fail(c, utils.CodeBadRequest, "参数错误"); return
	}
	content := req.Content
	if content == "" {content = req.NoteText}
	note := &model.ImSessionNote{
		SessionID: sid, PlatformUID: h.getUID(c), Content: content, IsStarred: req.IsStarred,
	}
	if err := service.PlatformIMNoteUpsert(note); err != nil {
		utils.Fail(c, utils.CodeBadRequest, err.Error()); return
	}
	utils.Success(c, gin.H{"msg": "ok"})
}

// ---------- 搜索历史 (25) ----------

// ListSearchHistory GET /platform-im/search-history
func (h *PlatformIMHandler) ListSearchHistory(c *gin.Context) {
	list, err := service.PlatformIMSearchHistoryList(h.getUID(c))
	if err != nil {utils.Fail(c, utils.CodeServerError, err.Error()); return}
	utils.Success(c, list)
}

// ClearSearchHistory POST /platform-im/search-history/clear
func (h *PlatformIMHandler) ClearSearchHistory(c *gin.Context) {
	if err := service.PlatformIMSearchHistoryClear(h.getUID(c)); err != nil {
		utils.Fail(c, utils.CodeServerError, err.Error()); return
	}
	utils.Success(c, gin.H{"msg": "ok"})
}

// ---------- 平台工作台 (46~52, 62) ----------

// GetWorkbench GET /platform-im/workbench
func (h *PlatformIMHandler) GetWorkbench(c *gin.Context) {
	res, err := service.PlatformIMWorkbenchGet(h.getUID(c))
	if err != nil {utils.Fail(c, utils.CodeServerError, err.Error()); return}
	utils.Success(c, res)
}

// SaveWorkbenchLayout PUT /platform-im/workbench/layout  (兼容前端 layout_json.order 和 bucket_order)
func (h *PlatformIMHandler) SaveWorkbenchLayout(c *gin.Context) {
	// 修复: 兼容前端传 layout_json.order 或 bucket_order
	var req struct {
		BucketOrder []string                `json:"bucket_order"`
		LayoutJSON  map[string]interface{}  `json:"layout_json"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Fail(c, utils.CodeBadRequest, "参数错误"); return
	}
	order := req.BucketOrder
	if len(order) == 0 && req.LayoutJSON != nil {
		if v, ok := req.LayoutJSON["order"].([]interface{}); ok {
			for _, item := range v {
				if s, ok := item.(string); ok {order = append(order, s)}
			}
		}
	}
	if len(order) == 0 {utils.Fail(c, utils.CodeBadRequest, "板块顺序不能为空"); return}
	if err := service.PlatformIMWorkbenchLayoutSave(h.getUID(c), order); err != nil {
		utils.Fail(c, utils.CodeBadRequest, err.Error()); return
	}
	utils.Success(c, gin.H{"msg": "ok"})
}

// ---------- 快捷话术 (53) ----------

// ListQuickReplies GET /platform-im/quick-replies?category=soothe
func (h *PlatformIMHandler) ListQuickReplies(c *gin.Context) {
	list, err := service.PlatformIMQuickReplyList(c.Query("category"))
	if err != nil {utils.Fail(c, utils.CodeServerError, err.Error()); return}
	utils.Success(c, list)
}

// CreateQuickReply POST /platform-im/quick-replies
func (h *PlatformIMHandler) CreateQuickReply(c *gin.Context) {
	var r model.ImQuickReply
	if err := c.ShouldBindJSON(&r); err != nil {
		utils.Fail(c, utils.CodeBadRequest, "参数错误"); return
	}
	if err := service.PlatformIMQuickReplyCreate(&r); err != nil {
		utils.Fail(c, utils.CodeBadRequest, err.Error()); return
	}
	utils.Success(c, r)
}

// UpdateQuickReply PUT /platform-im/quick-replies/:id
func (h *PlatformIMHandler) UpdateQuickReply(c *gin.Context) {
	id := parseInt64Path(c, "id")
	var req struct {Title, Content, Category string}
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Fail(c, utils.CodeBadRequest, "参数错误"); return
	}
	if err := service.PlatformIMQuickReplyUpdate(id, req.Title, req.Content, req.Category); err != nil {
		utils.Fail(c, utils.CodeBadRequest, err.Error()); return
	}
	utils.Success(c, gin.H{"msg": "ok"})
}

// DeleteQuickReply DELETE /platform-im/quick-replies/:id
func (h *PlatformIMHandler) DeleteQuickReply(c *gin.Context) {
	id := parseInt64Path(c, "id")
	if err := service.PlatformIMQuickReplyDelete(id); err != nil {
		utils.Fail(c, utils.CodeBadRequest, err.Error()); return
	}
	utils.Success(c, gin.H{"msg": "ok"})
}

// ---------- 证据打包预览 (55) ----------

// ListEvidenceMessages GET /platform-im/sessions/:id/evidence
func (h *PlatformIMHandler) ListEvidenceMessages(c *gin.Context) {
	sid := parseInt64Path(c, "id")
	var from, to *time.Time
	if s := c.Query("from"); s != "" {
		if t, e := time.ParseInLocation("2006-01-02 15:04:05", s, time.Local); e == nil {from = &t}
	}
	if s := c.Query("to"); s != "" {
		if t, e := time.ParseInLocation("2006-01-02 15:04:05", s, time.Local); e == nil {to = &t}
	}
	list, err := service.PlatformIMEvidenceMessages(sid, from, to)
	if err != nil {utils.Fail(c, utils.CodeServerError, err.Error()); return}
	utils.Success(c, list)
}

// ---------- 样式下发 (93~98: 所有样式从后端拉取,前端仅渲染) ----------

// PullMyStyle GET /platform-im/styles/me?club_id=xxx
func (h *PlatformIMHandler) PullMyStyle(c *gin.Context) {
	cid, _ := strconv.ParseInt(c.DefaultQuery("club_id", "0"), 10, 64)
	res, err := service.PlatformIMPullMyStyle(h.getUID(c), cid)
	if err != nil {utils.Fail(c, utils.CodeServerError, err.Error()); return}
	utils.Success(c, res)
}

// PullAllStyles GET /platform-im/styles/all
// 修复: 前端 message-item.js 按 list.forEach 遍历,需返回气泡样式数组(而非 map)
func (h *PlatformIMHandler) PullAllStyles(c *gin.Context) {
	bubbles, frames, err := service.PlatformIMPullAllStyles()
	if err != nil {utils.Fail(c, utils.CodeServerError, err.Error()); return}
	// 返回气泡数组(前端按 role_key 建 map)
	bubbleList := make([]*model.ImBubbleStyle, 0, len(bubbles))
	for _, b := range bubbles {bubbleList = append(bubbleList, b)}
	utils.Success(c, gin.H{"bubbles": bubbleList, "avatar_frames": frames})
}

// ---------- 群聊批量免打扰 / 隐藏 / 聚合 (41~45) ----------

// UpdateGroupSetting PUT /platform-im/groups/:id/setting
func (h *PlatformIMHandler) UpdateGroupSetting(c *gin.Context) {
	gid := parseInt64Path(c, "id")
	var req struct {
		MuteNotify        *int8 `json:"mute_notify"`
		IsHidden          *int8 `json:"is_hidden"`
		AggregateSameClub *int8 `json:"aggregate_same_club"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Fail(c, utils.CodeBadRequest, "参数错误"); return
	}
	fields := make(map[string]interface{})
	if req.MuteNotify != nil {fields["mute_notify"] = *req.MuteNotify}
	if req.IsHidden != nil {fields["is_hidden"] = *req.IsHidden}
	if req.AggregateSameClub != nil {fields["aggregate_same_club"] = *req.AggregateSameClub}
	if err := service.PlatformIMGroupSetting(gid, fields); err != nil {
		utils.Fail(c, utils.CodeBadRequest, err.Error()); return
	}
	utils.Success(c, gin.H{"msg": "ok"})
}

// BatchGroupMute POST /platform-im/groups/batch-mute
func (h *PlatformIMHandler) BatchGroupMute(c *gin.Context) {
	var req struct {GroupIDs []int64 `json:"group_ids"`; Mute int8 `json:"mute"`}
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Fail(c, utils.CodeBadRequest, "参数错误"); return
	}
	if err := service.PlatformIMGroupBatchMute(req.GroupIDs, req.Mute); err != nil {
		utils.Fail(c, utils.CodeBadRequest, err.Error()); return
	}
	utils.Success(c, gin.H{"msg": "ok"})
}

// (parseIntQuery / parseInt64Path 已在 handler/common.go 定义，此处删除重复声明)
