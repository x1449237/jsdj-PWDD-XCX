package service

import (
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/datatypes"
	"gorm.io/gorm"

	"github.com/jisan/e-sports-platform/internal/model"
	"github.com/jisan/e-sports-platform/internal/utils"
)

// ============================================================
// 通用辅助函数
// ============================================================

// paginate 构造分页 scope，page 从 1 开始，size 默认 20，上限 100。
func paginate(page, size int) func(db *gorm.DB) *gorm.DB {
	if page < 1 {
		page = 1
	}
	if size < 1 || size > 100 {
		size = 20
	}
	return func(db *gorm.DB) *gorm.DB {
		return db.Offset((page - 1) * size).Limit(size)
	}
}

// ============================================================
// 一、俱乐部管理扩展 service 函数
// ============================================================

// ---------- 主页装修 ----------

// ClubGetHomeDecoration 获取俱乐部主页装修配置。
func ClubGetHomeDecoration(clubID int64) (*model.ClubHomeDecoration, error) {
	var dec model.ClubHomeDecoration
	if err := readDB.Where("club_id = ?", clubID).First(&dec).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("主页装修不存在: club_id=%d", clubID)
		}
		return nil, err
	}
	return &dec, nil
}

// ClubUpdateHomeDecoration 更新俱乐部主页装修字段。
func ClubUpdateHomeDecoration(clubID int64, fields map[string]interface{}) error {
	if len(fields) == 0 {
		return errors.New("更新字段为空")
	}
	fields["updated_at"] = time.Now()
	res := db.Model(&model.ClubHomeDecoration{}).Where("club_id = ?", clubID).Updates(fields)
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return fmt.Errorf("主页装修不存在或未变更: club_id=%d", clubID)
	}
	return nil
}

// ---------- 俱乐部技能项目 ----------

// ClubListServices 分页查询俱乐部技能项目列表。
func ClubListServices(clubID int64, page, size int) ([]model.ClubService, int64, error) {
	var list []model.ClubService
	var total int64
	db := readDB.Model(&model.ClubService{}).Where("club_id = ?", clubID)
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if err := db.Scopes(paginate(page, size)).Order("sort ASC, id DESC").Find(&list).Error; err != nil {
		return nil, 0, err
	}
	return list, total, nil
}

// ClubCreateService 创建俱乐部技能项目。
func ClubCreateService(clubID int64, svc *model.ClubService) error {
	if svc == nil {
		return errors.New("技能项目为空")
	}
	svc.ClubID = clubID
	if err := db.Create(svc).Error; err != nil {
		return fmt.Errorf("创建技能项目失败: %w", err)
	}
	return nil
}

// ClubUpdateService 更新俱乐部技能项目字段。
func ClubUpdateService(clubID, serviceID int64, fields map[string]interface{}) error {
	if len(fields) == 0 {
		return errors.New("更新字段为空")
	}
	fields["updated_at"] = time.Now()
	res := db.Model(&model.ClubService{}).
		Where("club_id = ? AND id = ?", clubID, serviceID).
		Updates(fields)
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return fmt.Errorf("技能项目不存在或未变更: club_id=%d, id=%d", clubID, serviceID)
	}
	return nil
}

// ClubDeleteService 删除俱乐部技能项目。
func ClubDeleteService(clubID, serviceID int64) error {
	res := db.Where("club_id = ? AND id = ?", clubID, serviceID).Delete(&model.ClubService{})
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return fmt.Errorf("技能项目不存在: club_id=%d, id=%d", clubID, serviceID)
	}
	return nil
}

// ---------- 成员技能名片 ----------

// ClubGetMemberCard 获取成员技能名片。
func ClubGetMemberCard(clubID, memberUID int64) (*model.ClubMemberCard, error) {
	var card model.ClubMemberCard
	if err := readDB.Where("club_id = ? AND member_uid = ?", clubID, memberUID).First(&card).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("成员名片不存在: club_id=%d, member_uid=%d", clubID, memberUID)
		}
		return nil, err
	}
	return &card, nil
}

// ClubUpdateMemberCard 更新成员技能名片字段。
func ClubUpdateMemberCard(clubID, memberUID int64, fields map[string]interface{}) error {
	if len(fields) == 0 {
		return errors.New("更新字段为空")
	}
	fields["updated_at"] = time.Now()
	res := db.Model(&model.ClubMemberCard{}).
		Where("club_id = ? AND member_uid = ?", clubID, memberUID).
		Updates(fields)
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return fmt.Errorf("成员名片不存在或未变更: club_id=%d, member_uid=%d", clubID, memberUID)
	}
	return nil
}

// ---------- 成员档案 ----------

// ClubGetMemberProfile 获取成员档案（按游戏分区）。
func ClubGetMemberProfile(clubID, memberUID, gameID int64) (*model.ClubMemberProfile, error) {
	var profile model.ClubMemberProfile
	if err := readDB.Where("club_id = ? AND member_uid = ? AND game_id = ?", clubID, memberUID, gameID).
		First(&profile).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("成员档案不存在: club_id=%d, member_uid=%d, game_id=%d", clubID, memberUID, gameID)
		}
		return nil, err
	}
	return &profile, nil
}

// ClubUpdateMemberProfile 更新成员档案字段。
func ClubUpdateMemberProfile(clubID, memberUID, gameID int64, fields map[string]interface{}) error {
	if len(fields) == 0 {
		return errors.New("更新字段为空")
	}
	fields["updated_at"] = time.Now()
	res := db.Model(&model.ClubMemberProfile{}).
		Where("club_id = ? AND member_uid = ? AND game_id = ?", clubID, memberUID, gameID).
		Updates(fields)
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return fmt.Errorf("成员档案不存在或未变更: club_id=%d, member_uid=%d, game_id=%d", clubID, memberUID, gameID)
	}
	return nil
}

// ---------- 权限操作日志 ----------

// ClubListPermissionLogs 分页查询权限操作日志。
func ClubListPermissionLogs(clubID int64, page, size int) ([]model.ClubPermissionLog, int64, error) {
	var list []model.ClubPermissionLog
	var total int64
	db := readDB.Model(&model.ClubPermissionLog{}).Where("club_id = ?", clubID)
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if err := db.Scopes(paginate(page, size)).Order("id DESC").Find(&list).Error; err != nil {
		return nil, 0, err
	}
	return list, total, nil
}

// ClubCreatePermissionLog 创建权限操作日志。
func ClubCreatePermissionLog(log *model.ClubPermissionLog) error {
	if log == nil {
		return errors.New("权限日志为空")
	}
	if err := db.Create(log).Error; err != nil {
		return fmt.Errorf("创建权限日志失败: %w", err)
	}
	return nil
}

// ---------- 退俱乐部申报 ----------

// ClubCreateResignation 创建退俱乐部申报。
func ClubCreateResignation(r *model.ClubMemberResignation) error {
	if r == nil {
		return errors.New("退会申报为空")
	}
	if err := db.Create(r).Error; err != nil {
		return fmt.Errorf("创建退会申报失败: %w", err)
	}
	return nil
}

// ClubListResignations 分页查询退会申报列表（status<0 表示全部）。
func ClubListResignations(clubID int64, status int8, page, size int) ([]model.ClubMemberResignation, int64, error) {
	var list []model.ClubMemberResignation
	var total int64
	db := readDB.Model(&model.ClubMemberResignation{}).Where("club_id = ?", clubID)
	if status >= 0 {
		db = db.Where("status = ?", status)
	}
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if err := db.Scopes(paginate(page, size)).Order("id DESC").Find(&list).Error; err != nil {
		return nil, 0, err
	}
	return list, total, nil
}

// ClubAuditResignation 审核退会申报。
func ClubAuditResignation(clubID, id int64, status int8, auditorUID int64) error {
	res := db.Model(&model.ClubMemberResignation{}).
		Where("club_id = ? AND id = ?", clubID, id).
		Updates(map[string]interface{}{
			"status":      status,
			"auditor_uid": auditorUID,
			"audited_at":  time.Now(),
			"updated_at":  time.Now(),
		})
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return fmt.Errorf("退会申报不存在: club_id=%d, id=%d", clubID, id)
	}
	return nil
}

// ---------- 黑名单 ----------

// ClubAddBlacklist 加入黑名单。
func ClubAddBlacklist(b *model.ClubBlacklist) error {
	if b == nil {
		return errors.New("黑名单记录为空")
	}
	if err := db.Create(b).Error; err != nil {
		return fmt.Errorf("加入黑名单失败: %w", err)
	}
	return nil
}

// ClubRemoveBlacklist 移出黑名单。
func ClubRemoveBlacklist(clubID, userID int64) error {
	res := db.Where("club_id = ? AND user_id = ?", clubID, userID).Delete(&model.ClubBlacklist{})
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return fmt.Errorf("黑名单记录不存在: club_id=%d, user_id=%d", clubID, userID)
	}
	return nil
}

// ClubListBlacklists 分页查询黑名单。
func ClubListBlacklists(clubID int64, page, size int) ([]model.ClubBlacklist, int64, error) {
	var list []model.ClubBlacklist
	var total int64
	db := readDB.Model(&model.ClubBlacklist{}).Where("club_id = ?", clubID)
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if err := db.Scopes(paginate(page, size)).Order("id DESC").Find(&list).Error; err != nil {
		return nil, 0, err
	}
	return list, total, nil
}

// ClubCheckBlacklist 检查用户是否在黑名单中。
func ClubCheckBlacklist(clubID, userID int64) (bool, error) {
	var count int64
	if err := readDB.Model(&model.ClubBlacklist{}).
		Where("club_id = ? AND user_id = ?", clubID, userID).
		Count(&count).Error; err != nil {
		return false, err
	}
	return count > 0, nil
}

// ---------- 积分体系 ----------

// ClubListPointRules 查询俱乐部积分规则列表。
func ClubListPointRules(clubID int64) ([]model.ClubPointRule, error) {
	var list []model.ClubPointRule
	if err := readDB.Where("club_id = ?", clubID).Order("id ASC").Find(&list).Error; err != nil {
		return nil, err
	}
	return list, nil
}

// ClubUpdatePointRule 更新积分规则。
func ClubUpdatePointRule(clubID, id int64, points int, description string) error {
	res := db.Model(&model.ClubPointRule{}).
		Where("club_id = ? AND id = ?", clubID, id).
		Updates(map[string]interface{}{
			"points":      points,
			"description": description,
			"updated_at":  time.Now(),
		})
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return fmt.Errorf("积分规则不存在: club_id=%d, id=%d", clubID, id)
	}
	return nil
}

// ClubAddPointLog 记录积分变动日志。
func ClubAddPointLog(log *model.ClubPointLog) error {
	if log == nil {
		return errors.New("积分日志为空")
	}
	if err := db.Create(log).Error; err != nil {
		return fmt.Errorf("记录积分日志失败: %w", err)
	}
	return nil
}

// ClubListPointLogs 分页查询成员积分变动日志（memberUID<=0 表示全部成员）。
func ClubListPointLogs(clubID, memberUID int64, page, size int) ([]model.ClubPointLog, int64, error) {
	var list []model.ClubPointLog
	var total int64
	db := readDB.Model(&model.ClubPointLog{}).Where("club_id = ?", clubID)
	if memberUID > 0 {
		db = db.Where("member_uid = ?", memberUID)
	}
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if err := db.Scopes(paginate(page, size)).Order("id DESC").Find(&list).Error; err != nil {
		return nil, 0, err
	}
	return list, total, nil
}

// ---------- 团费规则 ----------

// ClubGetFeeRule 获取俱乐部团费规则。
func ClubGetFeeRule(clubID int64) (*model.ClubFeeRule, error) {
	var rule model.ClubFeeRule
	if err := readDB.Where("club_id = ?", clubID).First(&rule).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("团费规则不存在: club_id=%d", clubID)
		}
		return nil, err
	}
	return &rule, nil
}

// ClubSaveFeeRule 保存团费规则（存在则更新，否则创建）。
func ClubSaveFeeRule(clubID int64, rule *model.ClubFeeRule) error {
	if rule == nil {
		return errors.New("团费规则为空")
	}
	rule.ClubID = clubID
	var exist model.ClubFeeRule
	err := readDB.Where("club_id = ?", clubID).First(&exist).Error
	if err == nil {
		rule.ID = exist.ID
		return db.Model(&model.ClubFeeRule{}).Where("id = ?", exist.ID).Updates(rule).Error
	}
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return db.Create(rule).Error
	}
	return err
}

// ---------- 招募卡片 ----------

// ClubListRecruitCards 查询俱乐部招募卡片列表。
func ClubListRecruitCards(clubID int64) ([]model.ClubRecruitCard, error) {
	var list []model.ClubRecruitCard
	if err := readDB.Where("club_id = ?", clubID).Order("sort ASC, id DESC").Find(&list).Error; err != nil {
		return nil, err
	}
	return list, nil
}

// ClubSaveRecruitCard 保存招募卡片（存在则更新，否则创建）。
func ClubSaveRecruitCard(card *model.ClubRecruitCard) error {
	if card == nil {
		return errors.New("招募卡片为空")
	}
	if card.ID > 0 {
		res := db.Model(&model.ClubRecruitCard{}).Where("id = ?", card.ID).Updates(card)
		if res.Error != nil {
			return res.Error
		}
		if res.RowsAffected == 0 {
			return fmt.Errorf("招募卡片不存在: id=%d", card.ID)
		}
		return nil
	}
	return db.Create(card).Error
}

// ---------- 管理员待办 ----------

// ClubListAdminTodos 分页查询管理员待办（status<0 表示全部）。
func ClubListAdminTodos(clubID, adminUID int64, status int8, page, size int) ([]model.ClubAdminTodo, int64, error) {
	var list []model.ClubAdminTodo
	var total int64
	db := readDB.Model(&model.ClubAdminTodo{}).Where("club_id = ?", clubID)
	if adminUID > 0 {
		db = db.Where("admin_uid = ?", adminUID)
	}
	if status >= 0 {
		db = db.Where("status = ?", status)
	}
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if err := db.Scopes(paginate(page, size)).Order("id DESC").Find(&list).Error; err != nil {
		return nil, 0, err
	}
	return list, total, nil
}

// ClubCreateAdminTodo 创建管理员待办。
func ClubCreateAdminTodo(todo *model.ClubAdminTodo) error {
	if todo == nil {
		return errors.New("待办为空")
	}
	if err := db.Create(todo).Error; err != nil {
		return fmt.Errorf("创建待办失败: %w", err)
	}
	return nil
}

// ClubCompleteAdminTodo 完成管理员待办。
func ClubCompleteAdminTodo(id int64) error {
	res := db.Model(&model.ClubAdminTodo{}).Where("id = ?", id).
		Updates(map[string]interface{}{
			"status":     int8(1),
			"updated_at": time.Now(),
		})
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return fmt.Errorf("待办不存在: id=%d", id)
	}
	return nil
}

// ---------- 游戏分区 ----------

// ClubListGameZones 查询俱乐部游戏分区列表。
func ClubListGameZones(clubID int64) ([]model.ClubGameZone, error) {
	var list []model.ClubGameZone
	if err := readDB.Where("club_id = ?", clubID).Order("sort ASC, id ASC").Find(&list).Error; err != nil {
		return nil, err
	}
	return list, nil
}

// ClubSaveGameZone 保存游戏分区（存在则更新，否则创建）。
func ClubSaveGameZone(zone *model.ClubGameZone) error {
	if zone == nil {
		return errors.New("游戏分区为空")
	}
	if zone.ID > 0 {
		res := db.Model(&model.ClubGameZone{}).Where("id = ?", zone.ID).Updates(zone)
		if res.Error != nil {
			return res.Error
		}
		if res.RowsAffected == 0 {
			return fmt.Errorf("游戏分区不存在: id=%d", zone.ID)
		}
		return nil
	}
	return db.Create(zone).Error
}

// ---------- 临时抽成 ----------

// ClubListTempCommissionRules 查询俱乐部临时抽成规则列表。
func ClubListTempCommissionRules(clubID int64) ([]model.ClubTempCommissionRule, error) {
	var list []model.ClubTempCommissionRule
	if err := readDB.Where("club_id = ?", clubID).Order("id DESC").Find(&list).Error; err != nil {
		return nil, err
	}
	return list, nil
}

// ClubCreateTempCommissionRule 创建临时抽成规则。
func ClubCreateTempCommissionRule(rule *model.ClubTempCommissionRule) error {
	if rule == nil {
		return errors.New("临时抽成规则为空")
	}
	if err := db.Create(rule).Error; err != nil {
		return fmt.Errorf("创建临时抽成规则失败: %w", err)
	}
	return nil
}

// ClubGetActiveTempCommission 获取当前生效的临时抽成规则（按游戏分区）。
func ClubGetActiveTempCommission(clubID, gameID int64) (*model.ClubTempCommissionRule, error) {
	var rule model.ClubTempCommissionRule
	now := time.Now()
	if err := readDB.Where("club_id = ? AND game_id = ? AND start_time <= ? AND end_time >= ?",
		clubID, gameID, now, now).
		Order("id DESC").First(&rule).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &rule, nil
}

// ---------- 请假 ----------

// ClubCreateLeave 创建请假申请。
func ClubCreateLeave(l *model.ClubMemberLeave) error {
	if l == nil {
		return errors.New("请假申请为空")
	}
	if err := db.Create(l).Error; err != nil {
		return fmt.Errorf("创建请假申请失败: %w", err)
	}
	return nil
}

// ClubListLeaves 分页查询请假列表（status<0 表示全部）。
func ClubListLeaves(clubID int64, status int8, page, size int) ([]model.ClubMemberLeave, int64, error) {
	var list []model.ClubMemberLeave
	var total int64
	db := readDB.Model(&model.ClubMemberLeave{}).Where("club_id = ?", clubID)
	if status >= 0 {
		db = db.Where("status = ?", status)
	}
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if err := db.Scopes(paginate(page, size)).Order("id DESC").Find(&list).Error; err != nil {
		return nil, 0, err
	}
	return list, total, nil
}

// ClubAuditLeave 审核请假申请。
func ClubAuditLeave(clubID, id int64, status int8, auditorUID int64) error {
	res := db.Model(&model.ClubMemberLeave{}).
		Where("club_id = ? AND id = ?", clubID, id).
		Updates(map[string]interface{}{
			"status":      status,
			"auditor_uid": auditorUID,
			"audited_at":  time.Now(),
			"updated_at":  time.Now(),
		})
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return fmt.Errorf("请假申请不存在: club_id=%d, id=%d", clubID, id)
	}
	return nil
}

// ---------- 资料修改审核 ----------

// ClubCreateChangeRequest 创建资料修改申请。
func ClubCreateChangeRequest(r *model.ClubMemberChangeRequest) error {
	if r == nil {
		return errors.New("资料修改申请为空")
	}
	if err := db.Create(r).Error; err != nil {
		return fmt.Errorf("创建资料修改申请失败: %w", err)
	}
	return nil
}

// ClubListChangeRequests 分页查询资料修改申请列表（status<0 表示全部）。
func ClubListChangeRequests(clubID int64, status int8, page, size int) ([]model.ClubMemberChangeRequest, int64, error) {
	var list []model.ClubMemberChangeRequest
	var total int64
	db := readDB.Model(&model.ClubMemberChangeRequest{}).Where("club_id = ?", clubID)
	if status >= 0 {
		db = db.Where("status = ?", status)
	}
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if err := db.Scopes(paginate(page, size)).Order("id DESC").Find(&list).Error; err != nil {
		return nil, 0, err
	}
	return list, total, nil
}

// ClubAuditChangeRequest 审核资料修改申请。
func ClubAuditChangeRequest(clubID, id int64, status int8, auditorUID int64) error {
	res := db.Model(&model.ClubMemberChangeRequest{}).
		Where("club_id = ? AND id = ?", clubID, id).
		Updates(map[string]interface{}{
			"status":      status,
			"auditor_uid": auditorUID,
			"audited_at":  time.Now(),
			"updated_at":  time.Now(),
		})
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return fmt.Errorf("资料修改申请不存在: club_id=%d, id=%d", clubID, id)
	}
	return nil
}

// ---------- 优先派单 ----------

// ClubListPriorityDispatch 查询俱乐部优先派单配置列表。
func ClubListPriorityDispatch(clubID int64) ([]model.ClubPriorityDispatch, error) {
	var list []model.ClubPriorityDispatch
	if err := readDB.Where("club_id = ?", clubID).Order("priority DESC, id ASC").Find(&list).Error; err != nil {
		return nil, err
	}
	return list, nil
}

// ClubSetPriority 设置成员优先派单等级（存在则更新，否则创建）。
func ClubSetPriority(clubID, memberUID int64, priority int) error {
	var exist model.ClubPriorityDispatch
	err := readDB.Where("club_id = ? AND member_uid = ?", clubID, memberUID).First(&exist).Error
	if err == nil {
		return db.Model(&model.ClubPriorityDispatch{}).Where("id = ?", exist.ID).
			Updates(map[string]interface{}{
				"priority":   priority,
				"updated_at": time.Now(),
			}).Error
	}
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return db.Create(&model.ClubPriorityDispatch{
			ClubID:    clubID,
			MemberUID: memberUID,
			Priority:  priority,
		}).Error
	}
	return err
}

// ---------- 内部资源单 ----------

// ClubCreateInternalResource 创建内部资源单。
func ClubCreateInternalResource(r *model.ClubInternalResource) error {
	if r == nil {
		return errors.New("内部资源单为空")
	}
	if err := db.Create(r).Error; err != nil {
		return fmt.Errorf("创建内部资源单失败: %w", err)
	}
	return nil
}

// ClubListInternalResources 分页查询内部资源单。
func ClubListInternalResources(clubID int64, page, size int) ([]model.ClubInternalResource, int64, error) {
	var list []model.ClubInternalResource
	var total int64
	db := readDB.Model(&model.ClubInternalResource{}).Where("club_id = ?", clubID)
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if err := db.Scopes(paginate(page, size)).Order("id DESC").Find(&list).Error; err != nil {
		return nil, 0, err
	}
	return list, total, nil
}

// ---------- 客户归属 ----------

// ClubGetCustomerRelation 获取客户归属关系。
func ClubGetCustomerRelation(clubID, customerUID int64) (*model.ClubCustomerRelation, error) {
	var rel model.ClubCustomerRelation
	if err := readDB.Where("club_id = ? AND customer_uid = ?", clubID, customerUID).First(&rel).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("客户归属不存在: club_id=%d, customer_uid=%d", clubID, customerUID)
		}
		return nil, err
	}
	return &rel, nil
}

// ClubSaveCustomerRelation 保存客户归属关系（存在则更新，否则创建）。
func ClubSaveCustomerRelation(r *model.ClubCustomerRelation) error {
	if r == nil {
		return errors.New("客户归属为空")
	}
	var exist model.ClubCustomerRelation
	err := readDB.Where("club_id = ? AND customer_uid = ?", r.ClubID, r.CustomerUID).First(&exist).Error
	if err == nil {
		r.ID = exist.ID
		return db.Model(&model.ClubCustomerRelation{}).Where("id = ?", exist.ID).Updates(r).Error
	}
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return db.Create(r).Error
	}
	return err
}

// ClubListCustomerRelations 分页查询接待的客户归属列表（receptionistUID<=0 表示全部）。
func ClubListCustomerRelations(clubID, receptionistUID int64, page, size int) ([]model.ClubCustomerRelation, int64, error) {
	var list []model.ClubCustomerRelation
	var total int64
	db := readDB.Model(&model.ClubCustomerRelation{}).Where("club_id = ?", clubID)
	if receptionistUID > 0 {
		db = db.Where("receptionist_uid = ?", receptionistUID)
	}
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if err := db.Scopes(paginate(page, size)).Order("id DESC").Find(&list).Error; err != nil {
		return nil, 0, err
	}
	return list, total, nil
}

// ---------- 模板话术 ----------

// ClubListTemplatePhrases 查询俱乐部模板话术（按分类，category 为空表示全部）。
func ClubListTemplatePhrases(clubID int64, category string) ([]model.ClubTemplatePhrase, error) {
	var list []model.ClubTemplatePhrase
	db := readDB.Where("club_id = ?", clubID)
	if category != "" {
		db = db.Where("category = ?", category)
	}
	if err := db.Order("sort ASC, id DESC").Find(&list).Error; err != nil {
		return nil, err
	}
	return list, nil
}

// ClubSaveTemplatePhrase 保存模板话术（存在则更新，否则创建）。
func ClubSaveTemplatePhrase(p *model.ClubTemplatePhrase) error {
	if p == nil {
		return errors.New("模板话术为空")
	}
	if p.ID > 0 {
		res := db.Model(&model.ClubTemplatePhrase{}).Where("id = ?", p.ID).Updates(p)
		if res.Error != nil {
			return res.Error
		}
		if res.RowsAffected == 0 {
			return fmt.Errorf("模板话术不存在: id=%d", p.ID)
		}
		return nil
	}
	return db.Create(p).Error
}

// ClubDeleteTemplatePhrase 删除模板话术。
func ClubDeleteTemplatePhrase(clubID, id int64) error {
	res := db.Where("club_id = ? AND id = ?", clubID, id).Delete(&model.ClubTemplatePhrase{})
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return fmt.Errorf("模板话术不存在: club_id=%d, id=%d", clubID, id)
	}
	return nil
}

// ---------- 业绩排行 ----------

// ClubGetRanking 分页查询业绩排行（按周期类型与周期日期）。
func ClubGetRanking(clubID int64, periodType int8, periodDate string, page, size int) ([]model.ClubMemberRanking, int64, error) {
	var list []model.ClubMemberRanking
	var total int64
	db := readDB.Model(&model.ClubMemberRanking{}).
		Where("club_id = ? AND period_type = ? AND period_date = ?", clubID, periodType, periodDate)
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if err := db.Scopes(paginate(page, size)).Order("income DESC, id ASC").Find(&list).Error; err != nil {
		return nil, 0, err
	}
	return list, total, nil
}

// ---------- 俱乐部扩展设置 ----------

// ClubUpdateSettings 更新俱乐部扩展设置字段。
func ClubUpdateSettings(clubID int64, fields map[string]interface{}) error {
	if len(fields) == 0 {
		return errors.New("更新字段为空")
	}
	fields["updated_at"] = time.Now()
	res := db.Model(&model.Club{}).Where("id = ?", clubID).Updates(fields)
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return fmt.Errorf("俱乐部不存在或未变更: id=%d", clubID)
	}
	return nil
}

// ClubTransferFounder 转移俱乐部创始人权限。
func ClubTransferFounder(clubID, toUID int64) error {
	tx := db.Begin()
	if tx.Error != nil {
		return tx.Error
	}
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// 将原创始人降级为管理员
	res := tx.Model(&model.ClubMember{}).
		Where("club_id = ? AND role = ?", clubID, int8(1)).
		Updates(map[string]interface{}{
			"role":        int8(2),
			"role_detail": int8(0),
			"updated_at":  time.Now(),
		})
	if res.Error != nil {
		tx.Rollback()
		return res.Error
	}

	// 将目标成员升级为创始人
	res = tx.Model(&model.ClubMember{}).
		Where("club_id = ? AND user_id = ?", clubID, toUID).
		Updates(map[string]interface{}{
			"role":        int8(1),
			"role_detail": int8(1),
			"updated_at":  time.Now(),
		})
	if res.Error != nil {
		tx.Rollback()
		return res.Error
	}
	if res.RowsAffected == 0 {
		tx.Rollback()
		return fmt.Errorf("目标成员不存在: club_id=%d, user_id=%d", clubID, toUID)
	}

	// 更新俱乐部创始人字段
	res = tx.Model(&model.Club{}).Where("id = ?", clubID).
		Updates(map[string]interface{}{
			"founder_uid": toUID,
			"updated_at":  time.Now(),
		})
	if res.Error != nil {
		tx.Rollback()
		return res.Error
	}

	return tx.Commit().Error
}

// ClubSetMemberBan 设置成员禁单状态。
func ClubSetMemberBan(clubID, memberUID int64, isBanned int8) error {
	res := db.Model(&model.ClubMember{}).
		Where("club_id = ? AND user_id = ?", clubID, memberUID).
		Updates(map[string]interface{}{
			"is_banned":  isBanned,
			"updated_at": time.Now(),
		})
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return fmt.Errorf("成员不存在: club_id=%d, user_id=%d", clubID, memberUID)
	}
	return nil
}

// ClubSetMemberRole 设置成员角色。
func ClubSetMemberRole(clubID, memberUID int64, role, roleDetail int8) error {
	res := db.Model(&model.ClubMember{}).
		Where("club_id = ? AND user_id = ?", clubID, memberUID).
		Updates(map[string]interface{}{
			"role":        role,
			"role_detail": roleDetail,
			"updated_at":  time.Now(),
		})
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return fmt.Errorf("成员不存在: club_id=%d, user_id=%d", clubID, memberUID)
	}
	return nil
}

// ============================================================
// 二、订单扩展 service 函数
// ============================================================

// ---------- 订单模板 ----------

// OrderListTemplates 查询用户订单模板列表。
func OrderListTemplates(userID int64) ([]model.OrderTemplate, error) {
	var list []model.OrderTemplate
	if err := readDB.Where("user_id = ?", userID).Order("id DESC").Find(&list).Error; err != nil {
		return nil, err
	}
	return list, nil
}

// OrderCreateTemplate 创建订单模板。
func OrderCreateTemplate(t *model.OrderTemplate) error {
	if t == nil {
		return errors.New("订单模板为空")
	}
	if err := db.Create(t).Error; err != nil {
		return fmt.Errorf("创建订单模板失败: %w", err)
	}
	return nil
}

// OrderDeleteTemplate 删除订单模板（校验归属）。
func OrderDeleteTemplate(id, userID int64) error {
	res := db.Where("id = ? AND user_id = ?", id, userID).Delete(&model.OrderTemplate{})
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return fmt.Errorf("订单模板不存在或无权删除: id=%d, user_id=%d", id, userID)
	}
	return nil
}

// ---------- 订单补单 ----------

// OrderCreateSupplement 创建订单补单。
func OrderCreateSupplement(s *model.OrderSupplement) error {
	if s == nil {
		return errors.New("订单补单为空")
	}
	if err := db.Create(s).Error; err != nil {
		return fmt.Errorf("创建订单补单失败: %w", err)
	}
	return nil
}

// OrderListSupplements 查询订单补单列表。
func OrderListSupplements(orderID int64) ([]model.OrderSupplement, error) {
	var list []model.OrderSupplement
	if err := readDB.Where("order_id = ?", orderID).Order("id DESC").Find(&list).Error; err != nil {
		return nil, err
	}
	return list, nil
}

// ---------- 部分退款 ----------

// OrderCreatePartialRefund 创建部分退款申请。
func OrderCreatePartialRefund(r *model.OrderPartialRefund) error {
	if r == nil {
		return errors.New("部分退款申请为空")
	}
	if err := db.Create(r).Error; err != nil {
		return fmt.Errorf("创建部分退款申请失败: %w", err)
	}
	return nil
}

// OrderListPartialRefunds 查询订单部分退款列表。
func OrderListPartialRefunds(orderID int64) ([]model.OrderPartialRefund, error) {
	var list []model.OrderPartialRefund
	if err := readDB.Where("order_id = ?", orderID).Order("id DESC").Find(&list).Error; err != nil {
		return nil, err
	}
	return list, nil
}

// OrderAuditPartialRefund 审核部分退款申请。
func OrderAuditPartialRefund(id int64, status int8, auditorUID int64) error {
	res := db.Model(&model.OrderPartialRefund{}).Where("id = ?", id).
		Updates(map[string]interface{}{
			"status":      status,
			"auditor_uid": auditorUID,
			"audited_at":  time.Now(),
			"updated_at":  time.Now(),
		})
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return fmt.Errorf("部分退款申请不存在: id=%d", id)
	}
	return nil
}

// ---------- 订单备注 ----------

// OrderAddRemark 添加订单备注。
func OrderAddRemark(r *model.OrderRemark) error {
	if r == nil {
		return errors.New("订单备注为空")
	}
	if err := db.Create(r).Error; err != nil {
		return fmt.Errorf("添加订单备注失败: %w", err)
	}
	return nil
}

// OrderListRemarks 查询订单备注列表。
func OrderListRemarks(orderID int64) ([]model.OrderRemark, error) {
	var list []model.OrderRemark
	if err := readDB.Where("order_id = ?", orderID).Order("id DESC").Find(&list).Error; err != nil {
		return nil, err
	}
	return list, nil
}

// ---------- 订单延期 ----------

// OrderCreateExtension 创建订单延期申请。
func OrderCreateExtension(e *model.OrderExtension) error {
	if e == nil {
		return errors.New("订单延期申请为空")
	}
	if err := db.Create(e).Error; err != nil {
		return fmt.Errorf("创建订单延期申请失败: %w", err)
	}
	return nil
}

// OrderAuditExtension 审核订单延期申请。
func OrderAuditExtension(id int64, status int8, auditorUID int64) error {
	res := db.Model(&model.OrderExtension{}).Where("id = ?", id).
		Updates(map[string]interface{}{
			"status":      status,
			"auditor_uid": auditorUID,
			"audited_at":  time.Now(),
			"updated_at":  time.Now(),
		})
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return fmt.Errorf("订单延期申请不存在: id=%d", id)
	}
	return nil
}

// ---------- 订单转单 ----------

// OrderCreateTransfer 创建订单转单申请。
func OrderCreateTransfer(t *model.OrderTransfer) error {
	if t == nil {
		return errors.New("订单转单申请为空")
	}
	if err := db.Create(t).Error; err != nil {
		return fmt.Errorf("创建订单转单申请失败: %w", err)
	}
	return nil
}

// OrderAuditTransfer 审核订单转单申请。
func OrderAuditTransfer(id int64, status int8) error {
	res := db.Model(&model.OrderTransfer{}).Where("id = ?", id).
		Updates(map[string]interface{}{
			"status":     status,
			"updated_at": time.Now(),
		})
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return fmt.Errorf("订单转单申请不存在: id=%d", id)
	}
	return nil
}

// ---------- 订单改价 ----------

// OrderCreatePriceChange 创建订单改价申请。
func OrderCreatePriceChange(p *model.OrderPriceChange) error {
	if p == nil {
		return errors.New("订单改价申请为空")
	}
	if err := db.Create(p).Error; err != nil {
		return fmt.Errorf("创建订单改价申请失败: %w", err)
	}
	return nil
}

// OrderAuditPriceChange 审核订单改价申请。
func OrderAuditPriceChange(id int64, status int8, auditorUID int64) error {
	res := db.Model(&model.OrderPriceChange{}).Where("id = ?", id).
		Updates(map[string]interface{}{
			"status":      status,
			"auditor_uid": auditorUID,
			"audited_at":  time.Now(),
			"updated_at":  time.Now(),
		})
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return fmt.Errorf("订单改价申请不存在: id=%d", id)
	}
	return nil
}

// ---------- 价格变动日志 ----------

// OrderListPriceLogs 查询订单价格变动日志。
func OrderListPriceLogs(orderID int64) ([]model.OrderPriceLog, error) {
	var list []model.OrderPriceLog
	if err := readDB.Where("order_id = ?", orderID).Order("id DESC").Find(&list).Error; err != nil {
		return nil, err
	}
	return list, nil
}

// ---------- 订单标签 ----------

// OrderAddTag 为订单添加标签（幂等）。
func OrderAddTag(orderID, tagID int64) error {
	ot := model.OrderTagRelation{OrderID: orderID, TagID: tagID}
	if err := db.Where("order_id = ? AND tag_id = ?", orderID, tagID).
		FirstOrCreate(&ot).Error; err != nil {
		return fmt.Errorf("添加订单标签失败: %w", err)
	}
	return nil
}

// OrderRemoveTag 移除订单标签。
func OrderRemoveTag(orderID, tagID int64) error {
	res := db.Where("order_id = ? AND tag_id = ?", orderID, tagID).
		Delete(&model.OrderTagRelation{})
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return fmt.Errorf("订单标签关联不存在: order_id=%d, tag_id=%d", orderID, tagID)
	}
	return nil
}

// OrderListTags 查询全部可用标签。
func OrderListTags() ([]model.OrderTag, error) {
	var list []model.OrderTag
	if err := readDB.Order("id ASC").Find(&list).Error; err != nil {
		return nil, err
	}
	return list, nil
}

// ---------- 退款台账 ----------

// OrderListRefundLedger 查询订单退款台账。
func OrderListRefundLedger(orderID int64) ([]model.OrderRefundLedger, error) {
	var list []model.OrderRefundLedger
	if err := readDB.Where("order_id = ?", orderID).Order("id DESC").Find(&list).Error; err != nil {
		return nil, err
	}
	return list, nil
}

// ---------- 订单异常/归档 ----------

// OrderMarkAbnormal 标记订单异常状态。
func OrderMarkAbnormal(orderID int64, isAbnormal int8) error {
	res := db.Model(&model.Order{}).Where("id = ?", orderID).
		Updates(map[string]interface{}{
			"is_abnormal": isAbnormal,
			"updated_at":  time.Now(),
		})
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return fmt.Errorf("订单不存在: id=%d", orderID)
	}
	return nil
}

// OrderArchive 归档订单。
func OrderArchive(orderID int64) error {
	res := db.Model(&model.Order{}).Where("id = ?", orderID).
		Updates(map[string]interface{}{
			"is_archived": int8(1),
			"archived_at": time.Now(),
			"updated_at":  time.Now(),
		})
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return fmt.Errorf("订单不存在: id=%d", orderID)
	}
	return nil
}

// ---------- 俱乐部收藏 ----------

// UserFavoriteClub 用户收藏俱乐部（幂等）。
func UserFavoriteClub(userID, clubID int64) error {
	fav := model.ClubFavorite{UserID: userID, ClubID: clubID}
	if err := db.Where("user_id = ? AND club_id = ?", userID, clubID).
		FirstOrCreate(&fav).Error; err != nil {
		return fmt.Errorf("收藏俱乐部失败: %w", err)
	}
	return nil
}

// UserUnfavoriteClub 用户取消收藏俱乐部。
func UserUnfavoriteClub(userID, clubID int64) error {
	res := db.Where("user_id = ? AND club_id = ?", userID, clubID).
		Delete(&model.ClubFavorite{})
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return fmt.Errorf("收藏记录不存在: user_id=%d, club_id=%d", userID, clubID)
	}
	return nil
}

// UserListFavoriteClubs 分页查询用户收藏的俱乐部列表。
func UserListFavoriteClubs(userID int64, page, size int) ([]model.ClubFavorite, int64, error) {
	var list []model.ClubFavorite
	var total int64
	db := readDB.Model(&model.ClubFavorite{}).Where("user_id = ?", userID)
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if err := db.Scopes(paginate(page, size)).Order("id DESC").Find(&list).Error; err != nil {
		return nil, 0, err
	}
	return list, total, nil
}

// ============================================================
// 三、IM扩展 service 函数
// ============================================================

// ---------- 群聊文件 ----------

// GroupChatUploadFile 记录群聊文件上传。
func GroupChatUploadFile(f *model.GroupChatFile) error {
	if f == nil {
		return errors.New("群聊文件为空")
	}
	if err := db.Create(f).Error; err != nil {
		return fmt.Errorf("记录群聊文件失败: %w", err)
	}
	return nil
}

// GroupChatListFiles 分页查询群聊文件列表。
func GroupChatListFiles(groupID int64, page, size int) ([]model.GroupChatFile, int64, error) {
	var list []model.GroupChatFile
	var total int64
	db := readDB.Model(&model.GroupChatFile{}).Where("group_id = ?", groupID)
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if err := db.Scopes(paginate(page, size)).Order("id DESC").Find(&list).Error; err != nil {
		return nil, 0, err
	}
	return list, total, nil
}

// ---------- 快捷回复 ----------

// ChatListQuickReplies 查询快捷回复列表（按俱乐部与分类，category 为空表示全部）。
func ChatListQuickReplies(clubID int64, category string) ([]model.ChatQuickReply, error) {
	var list []model.ChatQuickReply
	db := readDB.Where("club_id = ?", clubID)
	if category != "" {
		db = db.Where("category = ?", category)
	}
	if err := db.Order("sort ASC, id ASC").Find(&list).Error; err != nil {
		return nil, err
	}
	return list, nil
}

// ChatSaveQuickReply 保存快捷回复（存在则更新，否则创建）。
func ChatSaveQuickReply(r *model.ChatQuickReply) error {
	if r == nil {
		return errors.New("快捷回复为空")
	}
	if r.ID > 0 {
		res := db.Model(&model.ChatQuickReply{}).Where("id = ?", r.ID).Updates(r)
		if res.Error != nil {
			return res.Error
		}
		if res.RowsAffected == 0 {
			return fmt.Errorf("快捷回复不存在: id=%d", r.ID)
		}
		return nil
	}
	return db.Create(r).Error
}

// ---------- 已读 ----------

// ChatMarkMessageRead 标记消息已读（幂等）。
func ChatMarkMessageRead(sessionID, messageID, userID int64) error {
	sid := fmt.Sprintf("%d", sessionID)
	read := model.ChatMessageRead{
		SessionID: sid,
		MessageID: messageID,
		UserID:    userID,
	}
	if err := db.Where("session_id = ? AND message_id = ? AND user_id = ?", sid, messageID, userID).
		FirstOrCreate(&read).Error; err != nil {
		return fmt.Errorf("标记已读失败: %w", err)
	}
	return nil
}

// ---------- 群聊禁言 ----------

// GroupChatMuteMember 群聊禁言成员（存在则更新，否则创建）。
func GroupChatMuteMember(m *model.GroupChatMute) error {
	if m == nil {
		return errors.New("禁言记录为空")
	}
	var exist model.GroupChatMute
	err := readDB.Where("group_id = ? AND member_uid = ?", m.GroupID, m.MemberUID).First(&exist).Error
	if err == nil {
		m.ID = exist.ID
		return db.Model(&model.GroupChatMute{}).Where("id = ?", exist.ID).Updates(m).Error
	}
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return db.Create(m).Error
	}
	return err
}

// GroupChatUnmuteMember 解除群聊禁言。
func GroupChatUnmuteMember(groupID, memberUID int64) error {
	res := db.Where("group_id = ? AND member_uid = ?", groupID, memberUID).
		Delete(&model.GroupChatMute{})
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return fmt.Errorf("禁言记录不存在: group_id=%d, member_uid=%d", groupID, memberUID)
	}
	return nil
}

// ---------- 举报 ----------

// ChatCreateReport 创建举报。
func ChatCreateReport(r *model.ChatReport) error {
	if r == nil {
		return errors.New("举报为空")
	}
	if err := db.Create(r).Error; err != nil {
		return fmt.Errorf("创建举报失败: %w", err)
	}
	return nil
}

// ChatListReports 分页查询举报列表（status<0 表示全部）。
func ChatListReports(status int8, page, size int) ([]model.ChatReport, int64, error) {
	var list []model.ChatReport
	var total int64
	db := readDB.Model(&model.ChatReport{})
	if status >= 0 {
		db = db.Where("status = ?", status)
	}
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if err := db.Scopes(paginate(page, size)).Order("id DESC").Find(&list).Error; err != nil {
		return nil, 0, err
	}
	return list, total, nil
}

// ChatHandleReport 处理举报。
func ChatHandleReport(id int64, status int8, handlerUID int64, result string) error {
	res := db.Model(&model.ChatReport{}).Where("id = ?", id).
		Updates(map[string]interface{}{
			"status":        status,
			"handler_uid":   handlerUID,
			"handle_result": result,
			"updated_at":    time.Now(),
		})
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return fmt.Errorf("举报不存在: id=%d", id)
	}
	return nil
}

// ---------- 会话置顶 ----------

// ChatTogglePinSession 切换会话置顶状态（按用户维度，记录于 chat_session_pins 表）。
func ChatTogglePinSession(userID, sessionID int64, isPinned int8) error {
	res := db.Table("chat_session_pins").
		Where("user_id = ? AND session_id = ?", userID, sessionID).
		Update("is_pinned", isPinned)
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected > 0 {
		return nil
	}
	return db.Table("chat_session_pins").Create(map[string]interface{}{
		"user_id":    userID,
		"session_id": sessionID,
		"is_pinned":  isPinned,
	}).Error
}

// ============================================================
// 四、财务扩展 service 函数
// ============================================================

// ---------- 财务台账 ----------

// ClubListFinanceLedger 分页查询俱乐部财务台账（按类型，ledgerType 为空表示全部）。
func ClubListFinanceLedger(clubID int64, ledgerType string, page, size int) ([]model.ClubFinanceLedger, int64, error) {
	var list []model.ClubFinanceLedger
	var total int64
	db := readDB.Model(&model.ClubFinanceLedger{}).Where("club_id = ?", clubID)
	if ledgerType != "" {
		db = db.Where("type = ?", ledgerType)
	}
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if err := db.Scopes(paginate(page, size)).Order("id DESC").Find(&list).Error; err != nil {
		return nil, 0, err
	}
	return list, total, nil
}

// ClubCreateFinanceLedger 创建财务台账记录。
func ClubCreateFinanceLedger(l *model.ClubFinanceLedger) error {
	if l == nil {
		return errors.New("财务台账为空")
	}
	if err := db.Create(l).Error; err != nil {
		return fmt.Errorf("创建财务台账失败: %w", err)
	}
	return nil
}

// ---------- 押金 ----------

// UserCreateDeposit 创建用户押金记录。
func UserCreateDeposit(d *model.UserDeposit) error {
	if d == nil {
		return errors.New("押金记录为空")
	}
	if err := db.Create(d).Error; err != nil {
		return fmt.Errorf("创建押金记录失败: %w", err)
	}
	return nil
}

// UserListDeposits 分页查询用户押金记录。
func UserListDeposits(userID int64, page, size int) ([]model.UserDeposit, int64, error) {
	var list []model.UserDeposit
	var total int64
	db := readDB.Model(&model.UserDeposit{}).Where("user_id = ?", userID)
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if err := db.Scopes(paginate(page, size)).Order("id DESC").Find(&list).Error; err != nil {
		return nil, 0, err
	}
	return list, total, nil
}

// UserRefundDeposit 退还押金（标记为已退款）。
func UserRefundDeposit(id int64) error {
	var dep model.UserDeposit
	if err := db.Where("id = ?", id).First(&dep).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return fmt.Errorf("押金记录不存在: id=%d", id)
		}
		return err
	}
	if dep.Status == 2 {
		return fmt.Errorf("押金已退款: id=%d", id)
	}
	res := db.Model(&model.UserDeposit{}).Where("id = ?", id).
		Updates(map[string]interface{}{
			"status":     int8(2),
			"updated_at": time.Now(),
		})
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return fmt.Errorf("押金退款失败: id=%d", id)
	}
	return nil
}

// ---------- 月结 ----------

// ClubListMonthlySettlements 查询俱乐部月结记录（按月份）。
func ClubListMonthlySettlements(clubID int64, month string) ([]model.MonthlySettlement, error) {
	var list []model.MonthlySettlement
	db := readDB.Where("club_id = ?", clubID)
	if month != "" {
		db = db.Where("settle_month = ?", month)
	}
	if err := db.Order("id DESC").Find(&list).Error; err != nil {
		return nil, err
	}
	return list, nil
}

// ClubCreateMonthlySettlement 创建月结记录。
func ClubCreateMonthlySettlement(s *model.MonthlySettlement) error {
	if s == nil {
		return errors.New("月结记录为空")
	}
	if err := db.Create(s).Error; err != nil {
		return fmt.Errorf("创建月结记录失败: %w", err)
	}
	return nil
}

// ---------- 返利 ----------

// ClubListRebateRecords 分页查询返利记录。
func ClubListRebateRecords(clubID int64, page, size int) ([]model.RebateRecord, int64, error) {
	var list []model.RebateRecord
	var total int64
	db := readDB.Model(&model.RebateRecord{}).Where("club_id = ?", clubID)
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if err := db.Scopes(paginate(page, size)).Order("id DESC").Find(&list).Error; err != nil {
		return nil, 0, err
	}
	return list, total, nil
}

// ClubCreateRebateRecord 创建返利记录。
func ClubCreateRebateRecord(r *model.RebateRecord) error {
	if r == nil {
		return errors.New("返利记录为空")
	}
	if err := db.Create(r).Error; err != nil {
		return fmt.Errorf("创建返利记录失败: %w", err)
	}
	return nil
}

// ClubAuditRebateRecord 审核返利记录。
func ClubAuditRebateRecord(id int64, status int8) error {
	res := db.Model(&model.RebateRecord{}).Where("id = ?", id).
		Updates(map[string]interface{}{
			"status":     status,
			"updated_at": time.Now(),
		})
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return fmt.Errorf("返利记录不存在: id=%d", id)
	}
	return nil
}

// ---------- 钱包变动日志 ----------

// UserListWalletChangeLogs 分页查询用户钱包变动日志（changeType 为空表示全部）。
func UserListWalletChangeLogs(userID int64, changeType string, page, size int) ([]model.WalletChangeLog, int64, error) {
	var list []model.WalletChangeLog
	var total int64
	db := readDB.Model(&model.WalletChangeLog{}).Where("user_id = ?", userID)
	if changeType != "" {
		db = db.Where("change_type = ?", changeType)
	}
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if err := db.Scopes(paginate(page, size)).Order("id DESC").Find(&list).Error; err != nil {
		return nil, 0, err
	}
	return list, total, nil
}

// UserCreateWalletChangeLog 创建钱包变动日志。
func UserCreateWalletChangeLog(l *model.WalletChangeLog) error {
	if l == nil {
		return errors.New("钱包变动日志为空")
	}
	if err := db.Create(l).Error; err != nil {
		return fmt.Errorf("创建钱包变动日志失败: %w", err)
	}
	return nil
}

// ---------- 处罚模板 ----------

// ClubListPunishmentTemplates 查询处罚模板列表。
func ClubListPunishmentTemplates() ([]model.PunishmentTemplate, error) {
	var list []model.PunishmentTemplate
	if err := readDB.Order("id ASC").Find(&list).Error; err != nil {
		return nil, err
	}
	return list, nil
}

// ClubSavePunishmentTemplate 保存处罚模板（存在则更新，否则创建）。
func ClubSavePunishmentTemplate(t *model.PunishmentTemplate) error {
	if t == nil {
		return errors.New("处罚模板为空")
	}
	if t.ID > 0 {
		res := db.Model(&model.PunishmentTemplate{}).Where("id = ?", t.ID).Updates(t)
		if res.Error != nil {
			return res.Error
		}
		if res.RowsAffected == 0 {
			return fmt.Errorf("处罚模板不存在: id=%d", t.ID)
		}
		return nil
	}
	return db.Create(t).Error
}

// ============================================================
// 五、UX/营销扩展 service 函数
// ============================================================

// ---------- 用户反馈 ----------

// UserCreateFeedback 创建用户反馈。
func UserCreateFeedback(f *model.UserFeedback) error {
	if f == nil {
		return errors.New("用户反馈为空")
	}
	if err := db.Create(f).Error; err != nil {
		return fmt.Errorf("创建用户反馈失败: %w", err)
	}
	return nil
}

// UserListFeedbacks 分页查询用户反馈列表。
func UserListFeedbacks(userID int64, page, size int) ([]model.UserFeedback, int64, error) {
	var list []model.UserFeedback
	var total int64
	db := readDB.Model(&model.UserFeedback{}).Where("user_id = ?", userID)
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if err := db.Scopes(paginate(page, size)).Order("id DESC").Find(&list).Error; err != nil {
		return nil, 0, err
	}
	return list, total, nil
}

// AdminListFeedbacks 分页查询全部用户反馈（status<0 表示全部）。
func AdminListFeedbacks(status int8, page, size int) ([]model.UserFeedback, int64, error) {
	var list []model.UserFeedback
	var total int64
	db := readDB.Model(&model.UserFeedback{})
	if status >= 0 {
		db = db.Where("status = ?", status)
	}
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if err := db.Scopes(paginate(page, size)).Order("id DESC").Find(&list).Error; err != nil {
		return nil, 0, err
	}
	return list, total, nil
}

// AdminReplyFeedback 管理员回复用户反馈。
func AdminReplyFeedback(id, handlerUID int64, reply string) error {
	res := db.Model(&model.UserFeedback{}).Where("id = ?", id).
		Updates(map[string]interface{}{
			"reply":       reply,
			"handler_uid": handlerUID,
			"status":      int8(2),
			"updated_at":  time.Now(),
		})
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return fmt.Errorf("用户反馈不存在: id=%d", id)
	}
	return nil
}

// ---------- 屏蔽玩家 ----------

// UserBlockPlayer 屏蔽玩家（幂等）。
func UserBlockPlayer(userID, blockedUID int64) error {
	rel := model.UserBlocklist{UserID: userID, BlockedUID: blockedUID}
	if err := db.Where("user_id = ? AND blocked_uid = ?", userID, blockedUID).
		FirstOrCreate(&rel).Error; err != nil {
		return fmt.Errorf("屏蔽玩家失败: %w", err)
	}
	return nil
}

// UserUnblockPlayer 取消屏蔽玩家。
func UserUnblockPlayer(userID, blockedUID int64) error {
	res := db.Where("user_id = ? AND blocked_uid = ?", userID, blockedUID).
		Delete(&model.UserBlocklist{})
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return fmt.Errorf("屏蔽记录不存在: user_id=%d, blocked_uid=%d", userID, blockedUID)
	}
	return nil
}

// UserCheckBlock 检查是否已屏蔽某玩家。
func UserCheckBlock(userID, blockedUID int64) (bool, error) {
	var count int64
	if err := readDB.Model(&model.UserBlocklist{}).
		Where("user_id = ? AND blocked_uid = ?", userID, blockedUID).
		Count(&count).Error; err != nil {
		return false, err
	}
	return count > 0, nil
}

// ---------- 通知设置 ----------

// UserGetNotificationSettings 获取用户通知设置。
func UserGetNotificationSettings(userID int64) (*model.UserNotificationSetting, error) {
	var s model.UserNotificationSetting
	if err := readDB.Where("user_id = ?", userID).First(&s).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("通知设置不存在: user_id=%d", userID)
		}
		return nil, err
	}
	return &s, nil
}

// UserUpdateNotificationSettings 更新用户通知设置（不存在则创建）。
func UserUpdateNotificationSettings(userID int64, fields map[string]interface{}) error {
	if len(fields) == 0 {
		return errors.New("更新字段为空")
	}
	var exist model.UserNotificationSetting
	err := readDB.Where("user_id = ?", userID).First(&exist).Error
	if err == nil {
		fields["updated_at"] = time.Now()
		res := db.Model(&model.UserNotificationSetting{}).Where("user_id = ?", userID).Updates(fields)
		if res.Error != nil {
			return res.Error
		}
		if res.RowsAffected == 0 {
			return fmt.Errorf("通知设置更新失败: user_id=%d", userID)
		}
		return nil
	}
	if errors.Is(err, gorm.ErrRecordNotFound) {
		fields["user_id"] = userID
		return db.Model(&model.UserNotificationSetting{}).Create(fields).Error
	}
	return err
}

// ---------- 活动弹窗 ----------

// UserListActivityPopups 查询用户可见的活动弹窗列表。
func UserListActivityPopups(clubID int64) ([]model.ActivityPopup, error) {
	var list []model.ActivityPopup
	now := time.Now()
	if err := readDB.Where("club_id = ? AND status = ? AND start_time <= ? AND end_time >= ?",
		clubID, int8(1), now, now).
		Order("id DESC").Find(&list).Error; err != nil {
		return nil, err
	}
	return list, nil
}

// AdminListActivityPopups 分页查询活动弹窗列表。
func AdminListActivityPopups(clubID int64, page, size int) ([]model.ActivityPopup, int64, error) {
	var list []model.ActivityPopup
	var total int64
	db := readDB.Model(&model.ActivityPopup{}).Where("club_id = ?", clubID)
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if err := db.Scopes(paginate(page, size)).Order("id DESC").Find(&list).Error; err != nil {
		return nil, 0, err
	}
	return list, total, nil
}

// AdminSaveActivityPopup 保存活动弹窗（存在则更新，否则创建）。
func AdminSaveActivityPopup(p *model.ActivityPopup) error {
	if p == nil {
		return errors.New("活动弹窗为空")
	}
	if p.ID > 0 {
		res := db.Model(&model.ActivityPopup{}).Where("id = ?", p.ID).Updates(p)
		if res.Error != nil {
			return res.Error
		}
		if res.RowsAffected == 0 {
			return fmt.Errorf("活动弹窗不存在: id=%d", p.ID)
		}
		return nil
	}
	return db.Create(p).Error
}

// ---------- 节日模板 ----------

// AdminListFestivalTemplates 查询节日模板列表。
func AdminListFestivalTemplates() ([]model.FestivalTemplate, error) {
	var list []model.FestivalTemplate
	if err := readDB.Order("id ASC").Find(&list).Error; err != nil {
		return nil, err
	}
	return list, nil
}

// AdminSaveFestivalTemplate 保存节日模板（存在则更新，否则创建）。
func AdminSaveFestivalTemplate(t *model.FestivalTemplate) error {
	if t == nil {
		return errors.New("节日模板为空")
	}
	if t.ID > 0 {
		res := db.Model(&model.FestivalTemplate{}).Where("id = ?", t.ID).Updates(t)
		if res.Error != nil {
			return res.Error
		}
		if res.RowsAffected == 0 {
			return fmt.Errorf("节日模板不存在: id=%d", t.ID)
		}
		return nil
	}
	return db.Create(t).Error
}

// ---------- 推广渠道 ----------

// AdminListPromoChannels 分页查询推广渠道列表。
func AdminListPromoChannels(page, size int) ([]model.PromoChannel, int64, error) {
	var list []model.PromoChannel
	var total int64
	db := readDB.Model(&model.PromoChannel{})
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if err := db.Scopes(paginate(page, size)).Order("id DESC").Find(&list).Error; err != nil {
		return nil, 0, err
	}
	return list, total, nil
}

// AdminCreatePromoChannel 创建推广渠道。
func AdminCreatePromoChannel(c *model.PromoChannel) error {
	if c == nil {
		return errors.New("推广渠道为空")
	}
	if err := db.Create(c).Error; err != nil {
		return fmt.Errorf("创建推广渠道失败: %w", err)
	}
	return nil
}

// ============================================================
// 引用占位（保证 import 不被裁剪，便于后续接入 gin 上下文与缓存）
// ============================================================

var (
	_ = gin.Context{}
	_ = json.RawMessage(nil)
	_ = datatypes.JSON{}
	_ = utils.CodeSuccess
)
