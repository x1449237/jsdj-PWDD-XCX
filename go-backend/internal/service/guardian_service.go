package service

import (
	"errors"
	"time"

	"gorm.io/gorm"

	"github.com/jisan/e-sports-platform/internal/model"
)

// BindGuardian 家长绑定未成年账户
// 通过未成年用户ID + 验证(简化:直接绑定)
func BindGuardian(parentUID, childUID int64) (*model.ParentGuardianBind, error) {
	if parentUID == childUID {
		return nil, errors.New("不可绑定自己")
	}
	child, err := userRepo.FindByID(childUID)
	if err != nil {
		return nil, err
	}
	if child == nil {
		return nil, errors.New("未成年用户不存在")
	}
	if child.IsMinor != 1 {
		return nil, errors.New("目标用户非未成年")
	}
	// 防重
	var cnt int64
	_ = db.Model(&model.ParentGuardianBind{}).Where("parent_uid = ? AND child_uid = ?", parentUID, childUID).Count(&cnt).Error
	if cnt > 0 {
		return nil, errors.New("已绑定该未成年账户")
	}
	b := &model.ParentGuardianBind{
		ParentUID:  parentUID,
		ChildUID:   childUID,
		VerifiedAt: nowTimePtr(),
		CreatedAt:  nowTimePtr(),
		UpdatedAt:  nowTimePtr(),
	}
	if err := db.Create(b).Error; err != nil {
		return nil, err
	}
	// 初始化家长设置(默认允许下单/打赏,月度限额 500 元)
	_ = db.Create(&model.ParentGuardianSetting{
		ParentUID:    parentUID,
		ChildUID:     childUID,
		MonthlyLimit: 50000,
		AllowOrder:   1,
		AllowReward:  1,
		IsFrozen:     0,
		CreatedAt:    nowTimePtr(),
		UpdatedAt:    nowTimePtr(),
	}).Error
	return b, nil
}

// GetChildReport 未成年消费报告(当月消费汇总)
func GetChildReport(parentUID, childUID int64) (map[string]interface{}, error) {
	if !isGuardian(parentUID, childUID) {
		return nil, errors.New("非该未成年人的家长")
	}
	// 当月订单消费
	now := time.Now()
	monthStart := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())
	var spent int64
	_ = db.Model(&model.Order{}).Where("user_id = ? AND created_at >= ? AND pay_status IN ?", childUID, monthStart, []int8{1, 3}).
		Select("COALESCE(SUM(amount - refund_amount),0)").Scan(&spent).Error
	// 当月打赏
	var reward int64
	_ = db.Model(&model.Reward{}).Where("user_id = ? AND created_at >= ?", childUID, monthStart).
		Select("COALESCE(SUM(amount),0)").Scan(&reward).Error
	// 家长设置
	var setting model.ParentGuardianSetting
	_ = db.Where("parent_uid = ? AND child_uid = ?", parentUID, childUID).First(&setting).Error
	return map[string]interface{}{
		"child_uid":      childUID,
		"month_spent":    spent,
		"month_reward":   reward,
		"monthly_limit":  setting.MonthlyLimit,
		"allow_order":    setting.AllowOrder,
		"allow_reward":   setting.AllowReward,
		"is_frozen":      setting.IsFrozen,
	}, nil
}

// UpdateChildSettings 家长更新未成年设置
func UpdateChildSettings(parentUID, childUID int64, monthlyLimit int64, allowOrder, allowReward, isFrozen int8) error {
	if !isGuardian(parentUID, childUID) {
		return errors.New("非该未成年人的家长")
	}
	return db.Model(&model.ParentGuardianSetting{}).
		Where("parent_uid = ? AND child_uid = ?", parentUID, childUID).
		Updates(map[string]interface{}{
			"monthly_limit": monthlyLimit,
			"allow_order":   allowOrder,
			"allow_reward":  allowReward,
			"is_frozen":     isFrozen,
			"updated_at":    nowTimePtr(),
		}).Error
}

// FreezeChild 家长冻结/解冻未成年账户
func FreezeChild(parentUID, childUID int64, freeze bool) error {
	if !isGuardian(parentUID, childUID) {
		return errors.New("非该未成年人的家长")
	}
	frozen := int8(0)
	if freeze {
		frozen = 1
	}
	// 更新家长设置中的冻结标记
	if err := db.Model(&model.ParentGuardianSetting{}).
		Where("parent_uid = ? AND child_uid = ?", parentUID, childUID).
		Update("is_frozen", frozen).Error; err != nil {
		return err
	}
	// 冻结则将用户状态置为封禁,解冻恢复
	status := int8(1)
	if freeze {
		status = 0
	}
	return db.Model(&model.User{}).Where("id = ?", childUID).Update("status", status).Error
}

// isGuardian 校验家长绑定关系
func isGuardian(parentUID, childUID int64) bool {
	var cnt int64
	_ = db.Model(&model.ParentGuardianBind{}).Where("parent_uid = ? AND child_uid = ?", parentUID, childUID).Count(&cnt).Error
	return cnt > 0
}

// CheckMinorOrderAllowed 未成年下单前置校验(供 order service 调用)
// 返回 nil 表示允许下单
func CheckMinorOrderAllowed(userID int64, amount int64) error {
	u, err := userRepo.FindByID(userID)
	if err != nil || u == nil {
		return nil
	}
	if u.IsMinor != 1 {
		return nil
	}
	// 宵禁校验
	hour := time.Now().Hour()
	if hour >= 22 || hour < 8 {
		_ = db.Create(&model.MinorCurfewLog{
			UserID: userID, Action: model.MinorActionOrder,
			BlockedAt: nowTimePtr(), CreatedAt: nowTimePtr(),
		}).Error
		return errors.New("未成年人宵禁时段禁止下单")
	}
	// 家长设置校验
	var setting model.ParentGuardianSetting
	if err := db.Where("child_uid = ?", userID).First(&setting).Error; err == nil {
		if setting.IsFrozen == 1 {
			return errors.New("账户已被家长冻结")
		}
		if setting.AllowOrder == 0 {
			return errors.New("家长已禁止下单")
		}
		// 月度消费限额校验
		if setting.MonthlyLimit > 0 {
			now := time.Now()
			monthStart := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())
			var spent int64
			_ = db.Model(&model.Order{}).Where("user_id = ? AND created_at >= ? AND pay_status IN ?", userID, monthStart, []int8{1, 3}).
				Select("COALESCE(SUM(amount - refund_amount),0)").Scan(&spent).Error
			if spent+amount > setting.MonthlyLimit {
				return errors.New("超过家长设置的月度消费限额")
			}
		}
	}
	return nil
}

// gormExpr 占位(避免直接引用 gorm.Expr 时跨包)
var _ = gorm.ErrRecordNotFound
