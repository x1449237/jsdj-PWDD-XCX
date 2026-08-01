package repository

import (
	"errors"

	"gorm.io/gorm"

	"github.com/jisan/e-sports-platform/internal/model"
)

// GroupBuyRepo 拼团数据访问仓储
type GroupBuyRepo struct {
	db *gorm.DB
}

// NewGroupBuyRepo 创建拼团仓储
func NewGroupBuyRepo(db *gorm.DB) *GroupBuyRepo {
	return &GroupBuyRepo{db: db}
}

// ============ Activity 活动相关 ============

// CreateActivity 创建拼团活动
func (r *GroupBuyRepo) CreateActivity(a *model.GroupBuyActivity) error {
	return r.db.Create(a).Error
}

// FindActivityByID 根据ID查询活动
func (r *GroupBuyRepo) FindActivityByID(id int64) (*model.GroupBuyActivity, error) {
	var a model.GroupBuyActivity
	if err := r.db.First(&a, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &a, nil
}

// ListActiveActivities 查询进行中的活动列表
func (r *GroupBuyRepo) ListActiveActivities() ([]model.GroupBuyActivity, error) {
	var list []model.GroupBuyActivity
	err := r.db.Where("status = ?", model.GroupBuyActivityStatusEnabled).
		Order("id DESC").Find(&list).Error
	return list, err
}

// UpdateActivity 更新活动
func (r *GroupBuyRepo) UpdateActivity(id int64, fields map[string]interface{}) error {
	return r.db.Model(&model.GroupBuyActivity{}).Where("id = ?", id).Updates(fields).Error
}

// ============ Group 群组相关 ============

// CreateGroup 创建拼团团组
func (r *GroupBuyRepo) CreateGroup(g *model.GroupBuyGroup) error {
	return r.db.Create(g).Error
}

// FindGroupByID 根据ID查询群组
func (r *GroupBuyRepo) FindGroupByID(id int64) (*model.GroupBuyGroup, error) {
	var g model.GroupBuyGroup
	if err := r.db.First(&g, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &g, nil
}

// FindPendingGroupByActivity 查找活动中待成团的群组
func (r *GroupBuyRepo) FindPendingGroupByActivity(activityID int64) (*model.GroupBuyGroup, error) {
	var g model.GroupBuyGroup
	err := r.db.Where("activity_id = ? AND status = ?", activityID, model.GroupBuyGroupStatusPending).
		Order("id ASC").First(&g).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &g, nil
}

// ListGroupsByActivity 按活动查询群组列表
func (r *GroupBuyRepo) ListGroupsByActivity(activityID int64) ([]model.GroupBuyGroup, error) {
	var list []model.GroupBuyGroup
	err := r.db.Where("activity_id = ?", activityID).
		Order("id DESC").Find(&list).Error
	return list, err
}

// UpdateGroup 更新群组
func (r *GroupBuyRepo) UpdateGroup(id int64, fields map[string]interface{}) error {
	return r.db.Model(&model.GroupBuyGroup{}).Where("id = ?", id).Updates(fields).Error
}

// IncrementMemberCount 增加群组人数(原子)
func (r *GroupBuyRepo) IncrementMemberCount(groupID int64) error {
	return r.db.Model(&model.GroupBuyGroup{}).Where("id = ?", groupID).
		UpdateColumn("member_count", gorm.Expr("member_count + ?", 1)).Error
}

// ============ Member 成员相关 ============

// CreateMember 创建拼团成员
func (r *GroupBuyRepo) CreateMember(m *model.GroupBuyMember) error {
	return r.db.Create(m).Error
}

// ListMembersByGroup 按群组查询成员
func (r *GroupBuyRepo) ListMembersByGroup(groupID int64) ([]model.GroupBuyMember, error) {
	var list []model.GroupBuyMember
	err := r.db.Where("group_id = ?", groupID).Find(&list).Error
	return list, err
}

// ListMembersByActivity 按活动查询成员
func (r *GroupBuyRepo) ListMembersByActivity(activityID int64) ([]model.GroupBuyMember, error) {
	var list []model.GroupBuyMember
	err := r.db.Where("activity_id = ?", activityID).Find(&list).Error
	return list, err
}

// CheckUserJoinedActivity 检查用户是否已参加某活动
func (r *GroupBuyRepo) CheckUserJoinedActivity(uid, activityID int64) (bool, error) {
	var cnt int64
	err := r.db.Model(&model.GroupBuyMember{}).
		Where("uid = ? AND activity_id = ?", uid, activityID).
		Count(&cnt).Error
	return cnt > 0, err
}

// UpdateMemberOrderID 更新成员订单ID
func (r *GroupBuyRepo) UpdateMemberOrderID(id, orderID int64) error {
	return r.db.Model(&model.GroupBuyMember{}).Where("id = ?", id).
		Update("order_id", orderID).Error
}
