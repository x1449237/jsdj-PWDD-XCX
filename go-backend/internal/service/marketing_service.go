package service

import (
	"errors"
	"fmt"
	"time"

	"github.com/jisan/e-sports-platform/internal/model"
)

// GetMyCoupons 用户优惠券列表
func GetMyCoupons(userID int64, status string) ([]model.UserCoupon, error) {
	var list []model.UserCoupon
	q := db.Model(&model.UserCoupon{}).Where("user_id = ?", userID)
	if status != "" {
		q = q.Where("status = ?", status)
	}
	err := q.Order("id DESC").Find(&list).Error
	return list, err
}

// GetRechargeActivities 充值活动列表
func GetRechargeActivities() ([]model.RechargeActivity, error) {
	var list []model.RechargeActivity
	now := time.Now()
	err := db.Where("status = 1 AND start_at <= ? AND end_at >= ?", now, now).
		Order("id DESC").Find(&list).Error
	return list, err
}

// Recharge 用户充值(余额入账 + 赠送)
func Recharge(userID int64, amount int64, activityID int64) (map[string]int64, error) {
	if amount <= 0 {
		return nil, errors.New("充值金额必须大于 0")
	}
	bonus := int64(0)
	if activityID > 0 {
		var act model.RechargeActivity
		if err := db.First(&act, activityID).Error; err == nil {
			if act.Status == 1 && amount >= act.ThresholdAmount {
				bonus = act.BonusAmount
			}
		}
	}
	total := amount + bonus
	if err := userRepo.UpdateBalance(userID, total); err != nil {
		return nil, err
	}
	return map[string]int64{
		"amount": amount,
		"bonus":  bonus,
		"total":  total,
	}, nil
}

// GetLotteryActivities 抽奖活动列表
func GetLotteryActivities() ([]model.LotteryActivity, error) {
	var list []model.LotteryActivity
	now := time.Now()
	err := db.Where("status = 1 AND start_at <= ? AND end_at >= ?", now, now).
		Order("id DESC").Find(&list).Error
	return list, err
}

// DrawLottery 抽奖(简化:随机返回中奖等级)
func DrawLottery(userID, activityID int64) (map[string]interface{}, error) {
	var act model.LotteryActivity
	if err := db.First(&act, activityID).Error; err != nil {
		return nil, errors.New("活动不存在")
	}
	if act.Status != 1 {
		return nil, errors.New("活动未开启")
	}
	// 随机中奖(简化:基于时间戳)
	n := time.Now().UnixNano() % 100
	prize := "谢谢参与"
	switch {
	case n < 1:
		prize = "一等奖"
	case n < 5:
		prize = "二等奖"
	case n < 20:
		prize = "三等奖"
	}
	return map[string]interface{}{
		"activity_id": activityID,
		"prize":       prize,
	}, nil
}

// GetGroupBuyActivities 拼团活动列表(占位:复用抽奖活动表结构)
func GetGroupBuyActivities() ([]model.LotteryActivity, error) {
	return GetLotteryActivities()
}

// JoinGroupBuy 参与拼团
func JoinGroupBuy(userID, activityID int64) (map[string]interface{}, error) {
	return map[string]interface{}{
		"activity_id": activityID,
		"user_id":     userID,
		"status":      "joined",
	}, nil
}

// GenerateInviteQRCode 生成用户专属邀请二维码内容
func GenerateInviteQRCode(userID int64) (string, error) {
	// 简化:返回包含用户ID的二维码内容
	return fmt.Sprintf("https://example.com/invite?uid=%d", userID), nil
}

// RedeemInviteCode 福利兑换(用户输入平台通用邀请码兑换福利)
func RedeemInviteCode(userID int64, code string) (map[string]interface{}, error) {
	if code == "" {
		return nil, errors.New("邀请码不能为空")
	}
	c, err := inviteCodeRepo.FindByCode(code)
	if err != nil {
		return nil, err
	}
	if c == nil {
		return nil, errors.New("邀请码不存在")
	}
	if c.Type != model.InviteCodeTypePlatform {
		return nil, errors.New("该邀请码不可用于福利兑换")
	}
	if c.Status == model.InviteCodeStatusRevoked {
		return nil, errors.New("邀请码已撤销")
	}
	if c.ExpireAt != nil && c.ExpireAt.Before(time.Now()) {
		return nil, errors.New("邀请码已过期")
	}
	consumed, ok, err := inviteCodeRepo.Consume(code, userID)
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, errors.New("邀请码已用尽或不可用")
	}
	// 解析福利配置(简化:固定赠送 100 积分)
	bonusPoints := 100
	_ = userRepo.UpdatePoints(userID, bonusPoints)
	return map[string]interface{}{
		"code":         consumed.Code,
		"bonus_points": bonusPoints,
	}, nil
}

// AdminGetCouponTemplates 优惠券模板列表
func AdminGetCouponTemplates(page, pageSize int) ([]model.CouponTemplate, int64, error) {
	var list []model.CouponTemplate
	var total int64
	q := db.Model(&model.CouponTemplate{})
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	err := q.Scopes(Paginate2(page, pageSize)).Order("id DESC").Find(&list).Error
	return list, total, err
}

// AdminCreateCouponTemplate 创建优惠券模板
func AdminCreateCouponTemplate(t *model.CouponTemplate) error {
	t.Status = 1
	t.CreatedAt = nowTimePtr()
	t.UpdatedAt = nowTimePtr()
	return db.Create(t).Error
}

// AdminGetRechargeActivities 充值活动列表(管理端)
func AdminGetRechargeActivities(page, pageSize int) ([]model.RechargeActivity, int64, error) {
	var list []model.RechargeActivity
	var total int64
	q := db.Model(&model.RechargeActivity{})
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	err := q.Scopes(Paginate2(page, pageSize)).Order("id DESC").Find(&list).Error
	return list, total, err
}

// AdminCreateRechargeActivity 创建充值活动
func AdminCreateRechargeActivity(a *model.RechargeActivity) error {
	a.Status = 1
	a.CreatedAt = nowTimePtr()
	a.UpdatedAt = nowTimePtr()
	return db.Create(a).Error
}

// AdminGetLotteryActivities 抽奖活动列表(管理端)
func AdminGetLotteryActivities(page, pageSize int) ([]model.LotteryActivity, int64, error) {
	var list []model.LotteryActivity
	var total int64
	q := db.Model(&model.LotteryActivity{})
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	err := q.Scopes(Paginate2(page, pageSize)).Order("id DESC").Find(&list).Error
	return list, total, err
}

// AdminCreateLotteryActivity 创建抽奖活动
func AdminCreateLotteryActivity(a *model.LotteryActivity) error {
	a.Status = 1
	a.CreatedAt = nowTimePtr()
	a.UpdatedAt = nowTimePtr()
	return db.Create(a).Error
}

// AdminGetGroupBuyActivities 拼团活动列表(管理端,复用抽奖表)
func AdminGetGroupBuyActivities(page, pageSize int) ([]model.LotteryActivity, int64, error) {
	return AdminGetLotteryActivities(page, pageSize)
}

// AdminCreateGroupBuyActivity 创建拼团活动(复用)
func AdminCreateGroupBuyActivity(a *model.LotteryActivity) error {
	return AdminCreateLotteryActivity(a)
}

// AdminGetInviteRewardConfig 邀请奖励配置(系统配置项)
func AdminGetInviteRewardConfig() (map[string]interface{}, error) {
	return map[string]interface{}{
		"invite_reward_points":   atoi(getSystemConfig("invite_reward_points")),
		"invite_reward_balance":  atoi(getSystemConfig("invite_reward_balance")),
	}, nil
}

// AdminUpdateInviteRewardConfig 更新邀请奖励配置
func AdminUpdateInviteRewardConfig(points, balance int64) error {
	if err := upsertSystemConfig("invite_reward_points", itoa(points), "邀请奖励积分"); err != nil {
		return err
	}
	return upsertSystemConfig("invite_reward_balance", itoa(balance), "邀请奖励余额")
}

// upsertSystemConfig 新增/更新系统配置项
func upsertSystemConfig(key, value, desc string) error {
	var sc model.SystemConfig
	res := db.Where("`key` = ?", key).First(&sc)
	if res.Error != nil {
		// 不存在则创建
		return db.Create(&model.SystemConfig{
			Key: key, Value: value, Description: desc, UpdatedAt: nowTimePtr(),
		}).Error
	}
	return db.Model(&sc).Updates(map[string]interface{}{
		"value": value, "updated_at": nowTimePtr(),
	}).Error
}
