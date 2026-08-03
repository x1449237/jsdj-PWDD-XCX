package service

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"math/rand"
	"strconv"
	"strings"
	"time"

	"github.com/jisan/e-sports-platform/internal/model"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

// =============== 全局参数 & 入驻总开关 ===============

// GetClubJoinSwitch 查询入驻总开关 (需求216/433/650)
func GetClubJoinSwitch() (bool, error) {
	var s model.GlobalClubJoinSwitch
	err := readDB.Where("switch_key = ?", "club_join").First(&s).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return true, nil // 默认开启
	}
	if err != nil {return false, err}
	return s.Enabled == 1, nil
}

// SetClubJoinSwitch 更新入驻总开关 (仅Web超管)
func SetClubJoinSwitch(enabled int8, byUID int64) error {
	now := time.Now()
	return db.Where("switch_key = ?", "club_join").
		Assign(model.GlobalClubJoinSwitch{Enabled: enabled, UpdatedBy: byUID, UpdatedAt: &now}).
		FirstOrCreate(&model.GlobalClubJoinSwitch{SwitchKey: "club_join"}).Error
}

// GetGlobalParam 读取平台阈值参数 (需求406~460 配套)
func GetGlobalParam(key, defVal string) (string, error) {
	var p model.PlatformGlobalParam
	err := readDB.Where("param_key = ?", key).First(&p).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {return defVal, nil}
	if err != nil {return "", err}
	if p.ParamValue == "" {return defVal, nil}
	return p.ParamValue, nil
}

// SetGlobalParam 设置平台参数 (Web超管)
func SetGlobalParam(key, val, typ, mod, desc string) error {
	now := time.Now()
	return db.Where("param_key = ?", key).Assign(model.PlatformGlobalParam{
		ParamValue: val, ParamType: typ, Module: mod, Description: desc, UpdatedAt: &now,
	}).FirstOrCreate(&model.PlatformGlobalParam{ParamKey: key}).Error
}

// =============== 订单号生成:BCYL-260721222138 (需求262~263,479~480,560) ===============

// GenerateClubOrderNo 唯一订单号生成 (俱乐部缩写 + 年月日时分 + 俱乐部当日内部序号)
// 返回: BCYL-260721222138
func GenerateClubOrderNo(clubID int64, abbr string) (string, error) {
	if abbr == "" || clubID <= 0 {return "", fmt.Errorf("俱乐部缩写/ID为空")}
	// 大写缩写，仅保留A-Z0-9
	cleaned := strings.ToUpper(abbr)
	cleaned = strings.Map(func(r rune) rune {
		if (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') {return r}
		return -1
	}, cleaned)
	if cleaned == "" {return "", fmt.Errorf("俱乐部缩写非法")}

	now := time.Now()
	// 年月日: 260721 (yy mm dd)
	yy := now.Year() % 100
	mm := int(now.Month())
	dd := now.Day()
	orderDate := fmt.Sprintf("%02d%02d%02d", yy, mm, dd)
	// 时分: 2221 (24小时制)
	hh24 := now.Hour()
	mi := now.Minute()
	hhmm := fmt.Sprintf("%02d%02d", hh24, mi)

	// 通过自增序列表原子+1获取当日内部序号 (事务内)
	var dailySeq int64
	err := db.Transaction(func(tx *gorm.DB) error {
		// SELECT ... FOR UPDATE
		var seq model.OrderSeqClub
		err := tx.Set("gorm:query_option", "FOR UPDATE").
			Where("club_id = ? AND order_date = ?", clubID, orderDate).First(&seq).Error
		if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {return err}
		if errors.Is(err, gorm.ErrRecordNotFound) {
			dailySeq = 1
			return tx.Create(&model.OrderSeqClub{ClubID: clubID, OrderDate: orderDate, DailySeq: dailySeq}).Error
		}
		dailySeq = seq.DailySeq + 1
		return tx.Model(&model.OrderSeqClub{}).Where("id = ?", seq.ID).Update("daily_seq", dailySeq).Error
	})
	if err != nil {return "", err}

	// 拼最终订单号: 缩写-年月日时分+序号
	// 样例: BCYL-260721222138
	orderNo := fmt.Sprintf("%s-%s%s%02d", cleaned, orderDate, hhmm, dailySeq)
	// 记录生成日志(含唯一索引防重)
	log := &model.OrderNoGenerateLog{
		OrderNo: orderNo, ClubID: clubID, ClubAbbr: cleaned,
		YYYYMMDDHHMI: orderDate + hhmm, DailySeq: dailySeq, CreatedAt: &now,
	}
	if err := db.Create(log).Error; err != nil {return "", err}
	return orderNo, nil
}

// =============== 入驻个人/企业 档案 (资料+对公打款) ===============

// GenerateEnterpriseRandomAmount 生成 0.1~2.0 元,一位小数(需求240)
func GenerateEnterpriseRandomAmount() (float64, error) {
	a := 1 + rand.Intn(20) // 1~20 => /10 => 0.1 到 2.0
	return float64(a) / 10.0, nil
}

// UpsertPersonalRegFiles 保存/更新入驻个人二进制档案 (身份证/PDF 全存MySQL,禁用OSS 551~552)
func UpsertPersonalRegFiles(f *model.PersonalClubRegistrationFiles) error {
	now := time.Now()
	f.UpdatedAt = &now
	return db.Where("personal_reg_id = ?", f.PersonalRegID).Assign(*f).FirstOrCreate(f).Error
}

// UpsertEnterpriseRegFiles 保存/更新入驻企业二进制档案 + 对公打款验证
func UpsertEnterpriseRegFiles(f *model.EnterpriseClubRegistrationFiles) error {
	return db.Where("enterprise_reg_id = ?", f.EnterpriseRegID).Assign(*f).FirstOrCreate(f).Error
}

// =============== 隐形水印导出 + 区块链存证 (需求461~485) ===============

// BuildExportWatermarkLog 构建一条水印导出日志 + 计算文件Hash + 区块链TXID占位
func BuildExportWatermarkLog(log *model.ExportWatermarkLog) error {
	// 基础校验
	if log.ExporterUID <= 0 || log.ExportScope == "" {return fmt.Errorf("导出日志字段不完整")}
	now := time.Now()
	// 流水号
	log.ExportNo = "EXP" + strconv.FormatInt(now.UnixNano()/1e6, 10)
	log.ExportMilliUTC = now.UnixNano() / 1e6
	log.FileNameSuffixTS = now.Format("20060102-150405")
	if log.OriginHashSHA256 == "" {
		// 若无原始文件字节,至少用导出人+时间+筛选条件摘要生成伪Hash占位(实际应传文件字节)
		payload := fmt.Sprintf("%d|%s|%s|%d|%s", log.ExporterUID, log.ExporterDeviceID, log.ExportFilterSummary, log.ExportMilliUTC, log.ExportLocation)
		h := sha256.Sum256([]byte(payload))
		log.OriginHashSHA256 = hex.EncodeToString(h[:])
	}
	// 区块链存证 TXID占位(真实环境接入法院联盟链SDK,此处以"BC-"+Hash前32位+时间戳 模拟)
	log.BlockchainTxid = "BC-" + log.OriginHashSHA256[:32] + "-" + strconv.FormatInt(now.Unix(), 10)
	log.BlockchainTimestamp = now.Unix()
	log.CreatedAt = &now
	return db.Create(log).Error
}

// QueryExportWatermarkByOrderNo 按导出流水号溯源
func QueryExportWatermarkByOrderNo(exportNo string) (*model.ExportWatermarkLog, error) {
	var r model.ExportWatermarkLog
	err := readDB.Where("export_no = ?", exportNo).First(&r).Error
	return &r, err
}

// InsertDetectLog 水印检测记录入库
func InsertDetectLog(d *model.WatermarkDetectLog) error {
	now := time.Now()
	d.CreatedAt = &now
	return db.Create(d).Error
}

// =============== 用户 IM 个性化偏好 (需求99/103/105/107/110, 316~328) ===============

// GetUserIMPreference 读用户偏好,不存在返回默认值模板
func GetUserIMPreference(uid int64) (*model.IMUserPreference, error) {
	var p model.IMUserPreference
	err := readDB.Where("uid = ?", uid).First(&p).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		def := &model.IMUserPreference{UID: uid}
		return def, nil
	}
	return &p, err
}

// SaveUserIMPreference 保存偏好
func SaveUserIMPreference(p *model.IMUserPreference) error {
	now := time.Now()
	p.UpdatedAt = &now
	return db.Where("uid = ?", p.UID).Assign(*p).FirstOrCreate(p).Error
}

// =============== 正在输入状态 (需求340, 文字触发/语音不触发) ===============

// UpsertTypingStatus 更新正在输入状态 (语音消息不写入)
func UpsertTypingStatus(sessID, uid int64, typingType string) error {
	now := time.Now()
	expire := now.Add(8 * time.Second)
	return db.Where("session_id = ? AND uid = ?", sessID, uid).
		Assign(model.ChatTypingStatus{TypingType: typingType, StartedAt: &now, ExpireAt: &expire}).
		FirstOrCreate(&model.ChatTypingStatus{SessionID: sessID, UID: uid}).Error
}

// GetSessionTypingUsers 查当前会话所有正在输入的用户 (仅typing_type=text触发)
func GetSessionTypingUsers(sessID int64) ([]int64, error) {
	var rows []*model.ChatTypingStatus
	err := readDB.Where("session_id = ? AND expire_at > ? AND typing_type = 'text'", sessID, time.Now()).Find(&rows).Error
	if err != nil {return nil, err}
	uids := make([]int64, 0, len(rows))
	for _, r := range rows {uids = append(uids, r.UID)}
	return uids, nil
}

// =============== 收藏 (127/344/165/782 统一收藏表) ===============

// ToggleUserFavorite 切换收藏状态,返回是否现在收藏
func ToggleUserFavorite(uid int64, favType string, targetID int64, extraJSON []byte) (bool, error) {
	var cur model.UserFavorite
	err := db.Where("uid = ? AND fav_type = ? AND target_id = ?", uid, favType, targetID).First(&cur).Error
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {return false, err}
	if errors.Is(err, gorm.ErrRecordNotFound) {
		now := time.Now()
		n := &model.UserFavorite{UID: uid, FavType: favType, TargetID: targetID, CreatedAt: &now}
		if len(extraJSON) > 0 {
			n.ExtraDataJSON = datatypes.JSON(extraJSON)
		}
		return true, db.Create(n).Error
	}
	return false, db.Delete(&cur).Error
}

// ListUserFavorites 我的收藏 (可按类型筛选)
func ListUserFavorites(uid int64, favType string, page, pageSize int) ([]*model.UserFavorite, int64, error) {
	q := readDB.Model(&model.UserFavorite{}).Where("uid = ?", uid)
	if favType != "" && favType != "all" {q = q.Where("fav_type = ?", favType)}
	var total int64
	q.Count(&total)
	var list []*model.UserFavorite
	err := q.Order("id DESC").Limit(pageSize).Offset((page-1)*pageSize).Find(&list).Error
	return list, total, err
}

// =============== 全链路公告 (748~767, 258~261) ===============

// CreateAnnouncement 发布公告(平台=0,俱乐部=club_id, 支持对内/对外/成员+客户)
func CreateAnnouncement(a *model.Announcement) error {
	now := time.Now()
	a.CreatedAt = &now
	a.UpdatedAt = &now
	return db.Create(a).Error
}

// ListAnnouncements 用户首页看到的公告 (平台公告 + 我的俱乐部公告,按有效+置顶排序)
func ListAnnouncements(uid, clubID int64, page, pageSize int) ([]*model.Announcement, int64, error) {
	now := time.Now()
	q := readDB.Model(&model.Announcement{}).Where("status = 1").
		Where("(effective_from IS NULL OR effective_from <= ?) AND (effective_to IS NULL OR effective_to > ?)", now, now).
		Where("ann_type = 'platform' OR (club_id = ? AND ann_type IN ('club_internal','club_external'))", clubID)
	var total int64
	q.Count(&total)
	var list []*model.Announcement
	err := q.Order("pinned DESC, sort_order DESC, id DESC").Limit(pageSize).Offset((page-1)*pageSize).Find(&list).Error
	return list, total, err
}

// UpdateAnnouncementStatus 撤回/置为过期(仅发布者/超管)
func UpdateAnnouncementStatus(id int64, status int8) error {
	return db.Model(&model.Announcement{}).Where("id = ?", id).Update("status", status).Error
}

// MarkAnnouncementRead 标记公告已读 (写入reads_json set)
func MarkAnnouncementRead(id, uid int64) error {
	var a model.Announcement
	if err := db.First(&a, id).Error; err != nil {return err}
	var reads []int64
	if len(a.ReadsJSON) > 0 {_ = json.Unmarshal(a.ReadsJSON, &reads)}
	for _, x := range reads {if x == uid {return nil}}
	reads = append(reads, uid)
	b, _ := json.Marshal(reads)
	return db.Model(&a).Update("reads_json", b).Error
}

// =============== 全域官方总群 快捷接口 (141~170 配套) ===============

// GetGlobalGroupConfig 查总群配置(不存在返回默认)
func GetGlobalGroupConfig(groupChatID int64) (*model.GlobalGroupConfig, error) {
	var r model.GlobalGroupConfig
	err := readDB.Where("group_chat_id = ?", groupChatID).First(&r).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return &model.GlobalGroupConfig{GroupChatID: groupChatID, Enabled: 1, NicknameLock: 1, AtAllDailyLimit: 3, MuteMode: 0, SinglePageLoadCount: 60}, nil
	}
	return &r, err
}

// UpsertGlobalGroupConfig 更新总群配置(仅平台官方)
func UpsertGlobalGroupConfig(cfg *model.GlobalGroupConfig) error {
	now := time.Now()
	cfg.UpdatedAt = &now
	return db.Where("group_chat_id = ?", cfg.GroupChatID).Assign(*cfg).FirstOrCreate(cfg).Error
}

// IncrementAtAllUsed @全体消耗配额(每日重置)
func IncrementAtAllUsed(groupChatID int64) error {
	return db.Transaction(func(tx *gorm.DB) error {
		cfg, err := GetGlobalGroupConfig(groupChatID)
		if err != nil {return err}
		today := time.Now().Format("2006-01-02")
		var resetDate string
		if cfg.AtAllResetDate != nil {resetDate = cfg.AtAllResetDate.Format("2006-01-02")}
		var used int = cfg.AtAllUsedToday
		if resetDate != today {used = 0}
		if used+1 > cfg.AtAllDailyLimit {return fmt.Errorf("今日@全体次数已达上限 %d 次", cfg.AtAllDailyLimit)}
		today0 := time.Now()
		return tx.Model(&model.GlobalGroupConfig{}).Where("group_chat_id = ?", groupChatID).
			Updates(map[string]interface{}{"at_all_used_today": used+1, "at_all_reset_date": today0}).Error
	})
}

// =============== V标渲染配置 (196~215 413~432 配套) ===============

// GetAllBadgeRenderConfigs 拉取所有V标配置(前端下发渲染,禁止前端伪造 212)
func GetAllBadgeRenderConfigs() ([]*model.BadgeRenderConfig, error) {
	var list []*model.BadgeRenderConfig
	err := readDB.Order("display_priority DESC").Find(&list).Error
	return list, err
}

// UpdateBadgeRenderConfig 更新V标渲染参数(仅Web超管)
func UpdateBadgeRenderConfig(c *model.BadgeRenderConfig) error {
	now := time.Now()
	c.UpdatedAt = &now
	return db.Where("badge_key = ?", c.BadgeKey).Assign(*c).FirstOrCreate(c).Error
}

// =============== 售后举证消息标记 + 介入提示 (391~412 配套) ===============

// MarkAfterSaleMessage 标记举证消息(双方举证/平台有效凭证)
func MarkAfterSaleMessage(sessionID, messageID, markerUID int64, markType, remark string) error {
	now := time.Now()
	return db.Create(&model.AfterSaleMessageMark{
		AfterSaleSessionID: sessionID, MessageID: messageID,
		MarkType: markType, MarkerUID: markerUID, MarkTime: &now, Remark: remark,
	}).Error
}

// ListAfterSaleMarks 按售后会话+消息类型 拉出举证材料
func ListAfterSaleMarks(sessionID int64, markType string) ([]*model.AfterSaleMessageMark, error) {
	q := readDB.Where("after_sale_session_id = ?", sessionID)
	if markType != "" {q = q.Where("mark_type = ?", markType)}
	var list []*model.AfterSaleMessageMark
	err := q.Order("id ASC").Find(&list).Error
	return list, err
}
