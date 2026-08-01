package service

import (
	"crypto/rand"
	"encoding/binary"
	"errors"
	"fmt"
	"math/big"
	"time"

	"gorm.io/gorm"

	"github.com/jisan/e-sports-platform/internal/model"
)

// 奖品类型常量
const (
	PrizeTypeCoupon    = "coupon"    // 优惠券
	PrizeTypeRecharge  = "recharge"  // 充值到余额
	PrizeTypePoints    = "points"    // 积分
	PrizeTypeBalance   = "balance"   // 余额(同recharge，保留兼容)
	PrizeTypeThankYou  = "thankyou"  // 谢谢参与
)

// DrawLotteryInput 抽奖入参扩展(IP等)
type DrawLotteryInput struct {
	ActivityID int64  `json:"activity_id"`
	DrawIP     string `json:"draw_ip"`
}

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
	// 记录充值日志
	var balBefore int64
	if u, err := userRepo.FindByID(userID); err == nil && u != nil {
		balBefore = u.Balance
	}
	_ = db.Create(&model.UserRechargeLog{
		UserID:        userID,
		Amount:        total,
		Source:        "recharge",
		RefID:         activityID,
		BalanceBefore: balBefore,
		BalanceAfter:  balBefore + total,
		CreatedAt:     nowTimePtr(),
	}).Error
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

// cryptoRandIntn 使用 crypto/rand 生成 [0, max) 内的真随机整数
func cryptoRandIntn(max int64) (int64, error) {
	if max <= 0 {
		return 0, errors.New("max must be > 0")
	}
	n, err := rand.Int(rand.Reader, big.NewInt(max))
	if err != nil {
		// 降级:用 crypto/rand 读4字节再取模
		var b [4]byte
		if _, err2 := rand.Read(b[:]); err2 != nil {
			return 0, err
		}
		v := int64(binary.LittleEndian.Uint32(b[:]))
		if v < 0 {
			v = -v
		}
		return v % max, nil
	}
	return n.Int64(), nil
}

// DrawLottery 抽奖(真随机 + 防重 + 扣库存 + 发奖)
func DrawLottery(userID, activityID int64, drawIP string) (map[string]interface{}, error) {
	if userID <= 0 {
		return nil, errors.New("用户未登录")
	}
	// 1. 校验活动
	var act model.LotteryActivity
	if err := db.First(&act, activityID).Error; err != nil {
		return nil, errors.New("活动不存在")
	}
	if act.Status != 1 {
		return nil, errors.New("活动未开启或已结束")
	}
	now := time.Now()
	if act.StartAt != nil && act.StartAt.After(now) {
		return nil, errors.New("活动尚未开始")
	}
	if act.EndAt != nil && act.EndAt.Before(now) {
		return nil, errors.New("活动已结束")
	}

	// 2. 防重:今日已抽奖次数(按活动+用户)
	today0 := startOfToday()
	tomorrow0 := today0.Add(24 * time.Hour)
	var todayDrawCnt int64
	err := db.Model(&model.LotteryRecord{}).
		Where("uid = ? AND activity_id = ? AND created_at >= ? AND created_at < ?",
			userID, activityID, today0, tomorrow0).
		Count(&todayDrawCnt).Error
	_ = err

	// 活动每日最大次数:从活动关联的奖品 daily_max 或 system_config 读取
	dailyMax := 1
	maxStr := getSystemConfig(fmt.Sprintf("lottery_daily_max:%d", activityID))
	if maxStr != "" {
		if n := atoi(maxStr); n > 0 {
			dailyMax = int(n)
		}
	}
	// 兜底:查任意奖品的 daily_max
	var samplePrize model.LotteryPrize
	if db.Where("activity_id = ? AND status = 1", activityID).Order("id ASC").First(&samplePrize).Error == nil {
		if samplePrize.DailyMax > 0 {
			dailyMax = samplePrize.DailyMax
		}
	}
	if todayDrawCnt >= int64(dailyMax) {
		return nil, fmt.Errorf("今日已达到活动最大抽奖次数(%d次)", dailyMax)
	}

	// 3. 加载活动奖品(按 Probability 权重抽取)
	var prizes []model.LotteryPrize
	err = db.Where("activity_id = ? AND status = 1", activityID).Order("display_order ASC, id ASC").Find(&prizes).Error
	if err != nil || len(prizes) == 0 {
		return nil, errors.New("活动奖品未配置")
	}

	// 计算总权重
	totalWeight := int64(0)
	for _, p := range prizes {
		if p.Probability <= 0 {
			continue
		}
		totalWeight += int64(p.Probability)
	}
	var wonPrize *model.LotteryPrize
	isWon := int8(0)
	prizeName := "谢谢参与"
	prizeID := int64(0)
	if totalWeight > 0 {
		rn, rerr := cryptoRandIntn(totalWeight)
		if rerr != nil {
			// 真随机失败时使用降级:仍用 crypto 读字节构造随机
			var b [8]byte
			_, _ = rand.Read(b[:])
			rn = int64(binary.LittleEndian.Uint64(b[:])) % totalWeight
			if rn < 0 {
				rn = -rn
			}
		}
		acc := int64(0)
		for i := range prizes {
			if prizes[i].Probability <= 0 {
				continue
			}
			acc += int64(prizes[i].Probability)
			if rn < acc {
				wonPrize = &prizes[i]
				prizeID = prizes[i].ID
				prizeName = prizes[i].PrizeName
				break
			}
		}
	}
	if wonPrize != nil && wonPrize.PrizeType != PrizeTypeThankYou {
		isWon = 1
	}

	// 4. 写抽奖记录(无论中奖与否)
	rec := &model.LotteryRecord{
		UID:        userID,
		ActivityID: activityID,
		PrizeID:    prizeID,
		DrawIP:     drawIP,
		IsWon:      isWon,
		Delivered:  0,
		CreatedAt:  nowTimePtr(),
	}
	if err := db.Create(rec).Error; err != nil {
		return nil, fmt.Errorf("记录抽奖失败: %w", err)
	}

	// 5. 中奖时:扣库存 + 发奖(事务)
	delivered := false
	var prizeDetail map[string]interface{}
	if wonPrize != nil && wonPrize.PrizeType != PrizeTypeThankYou {
		err = db.Transaction(func(tx *gorm.DB) error {
			// 扣库存(原子)
			res := tx.Model(&model.LotteryPrize{}).
				Where("id = ? AND stock > 0", wonPrize.ID).
				UpdateColumn("stock", gorm.Expr("stock - 1"))
			if res.Error != nil {
				return res.Error
			}
			if res.RowsAffected == 0 {
				return errors.New("奖品库存不足")
			}
			// 发奖
			if derr := deliverPrize(tx, userID, wonPrize, rec.ID); derr != nil {
				return derr
			}
			delivered = true
			return nil
		})
		if err != nil {
			return nil, err
		}
		// 标记已发奖
		if delivered {
			_ = db.Model(rec).Update("delivered", 1).Error
		}
		prizeDetail = map[string]interface{}{
			"prize_id":   wonPrize.ID,
			"prize_name": wonPrize.PrizeName,
			"prize_type": wonPrize.PrizeType,
			"prize_value": wonPrize.PrizeValue,
		}
	} else {
		prizeDetail = map[string]interface{}{
			"prize_id":   prizeID,
			"prize_name": prizeName,
			"prize_type": PrizeTypeThankYou,
		}
	}

	return map[string]interface{}{
		"activity_id": activityID,
		"record_id":   rec.ID,
		"is_won":      isWon,
		"prize":       prizeDetail,
		"daily_used":  todayDrawCnt + 1,
		"daily_max":   dailyMax,
	}, nil
}

// deliverPrize 发奖:按奖品类型写 user_coupon / user_recharge_log / 积分 / 余额
func deliverPrize(tx *gorm.DB, userID int64, prize *model.LotteryPrize, recID int64) error {
	if prize == nil {
		return nil
	}
	now := nowTimePtr()
	switch prize.PrizeType {
	case PrizeTypeCoupon:
		// 发券
		templateID := prize.PrizeValue
		if templateID <= 0 {
			// 无模板则跳过
			return nil
		}
		var validDays int
		var tmpl model.CouponTemplate
		if tx.First(&tmpl, templateID).Error == nil {
			validDays = tmpl.ValidDays
		}
		var expireAt *time.Time
		if validDays > 0 {
			t := time.Now().AddDate(0, 0, validDays)
			expireAt = &t
		}
		uc := &model.UserCoupon{
			UserID:     userID,
			TemplateID: templateID,
			Status:     model.UserCouponStatusUnused,
			ExpireAt:   expireAt,
			CreatedAt:  now,
			UpdatedAt:  now,
		}
		return tx.Create(uc).Error
	case PrizeTypeRecharge, PrizeTypeBalance:
		// 余额充值
		amount := prize.PrizeValue
		if amount <= 0 {
			return nil
		}
		var balBefore int64
		var u model.User
		if err := tx.Where("id = ?", userID).First(&u).Error; err == nil {
			balBefore = u.Balance
		}
		if err := tx.Model(&model.User{}).Where("id = ?", userID).
			UpdateColumn("balance", gorm.Expr("balance + ?", amount)).Error; err != nil {
			return err
		}
		return tx.Create(&model.UserRechargeLog{
			UserID:        userID,
			Amount:        amount,
			Source:        "lottery",
			RefID:         recID,
			BalanceBefore: balBefore,
			BalanceAfter:  balBefore + amount,
			CreatedAt:     now,
		}).Error
	case PrizeTypePoints:
		points := int(prize.PrizeValue)
		if points <= 0 {
			return nil
		}
		return tx.Model(&model.User{}).Where("id = ?", userID).
			UpdateColumn("points", gorm.Expr("points + ?", points)).Error
	default:
		// thankyou 或其他:不发奖
		return nil
	}
}

// DrawLotteryV1 兼容旧签名(不带drawIP)
func DrawLotteryV1Compat(userID, activityID int64) (map[string]interface{}, error) {
	return DrawLottery(userID, activityID, "")
}

// GetGroupBuyActivities 拼团活动列表（真实实现移至 group_buy_service.go）

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

// AdminGetGroupBuyActivities 拼团活动列表(管理端)
func AdminGetGroupBuyActivities(page, pageSize int) ([]model.GroupBuyActivity, int64, error) {
	var list []model.GroupBuyActivity
	var total int64
	q := db.Model(&model.GroupBuyActivity{})
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	err := q.Scopes(Paginate2(page, pageSize)).Order("id DESC").Find(&list).Error
	return list, total, err
}

// AdminCreateGroupBuyActivity 创建拼团活动(真实模型)
func AdminCreateGroupBuyActivity(a *model.GroupBuyActivity) error {
	a.Status = model.GroupBuyActivityStatusEnabled
	a.CreatedAt = nowTimePtr()
	a.UpdatedAt = nowTimePtr()
	return db.Create(a).Error
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
