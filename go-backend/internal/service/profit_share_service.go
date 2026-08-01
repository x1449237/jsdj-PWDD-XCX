package service

import (
	"errors"

	"gorm.io/gorm"

	"github.com/jisan/e-sports-platform/internal/model"
)

// SettleOrderProfitShare 订单结算分账
// 按启用分账规则将订单金额分给打手/俱乐部/分销商/平台
func SettleOrderProfitShare(orderID int64) error {
	o, err := orderRepo.FindByID(orderID)
	if err != nil {
		return err
	}
	if o == nil {
		return errors.New("订单不存在")
	}
	if o.Status == model.OrderStatusSettled {
		return nil // 已结算幂等
	}
	// 加载启用的分账规则(取第一条)
	rules, err := paymentRepo.ListProfitShareRules()
	if err != nil {
		return err
	}
	if len(rules) == 0 {
		// 无规则则全部分给打手
		return db.Transaction(func(tx *gorm.DB) error {
			if err := tx.Create(&model.ProfitShareRecord{
				OrderID: orderID, UserID: o.PlayerID, Role: model.ProfitShareRolePlayer,
				Amount: o.Amount, Ratio: 100, Status: 1,
				CreatedAt: nowTimePtr(), UpdatedAt: nowTimePtr(),
			}).Error; err != nil {
				return err
			}
			if err := tx.Model(&model.User{}).Where("id = ?", o.PlayerID).
				UpdateColumn("balance", gorm.Expr("balance + ?", o.Amount)).Error; err != nil {
				return err
			}
			return tx.Model(&model.Order{}).Where("id = ?", orderID).
				Updates(map[string]interface{}{"status": model.OrderStatusSettled, "settled_at": nowTimePtr()}).Error
		})
	}
	rule := rules[0]
	return db.Transaction(func(tx *gorm.DB) error {
		shares := []struct {
			role string
			ratio float64
			uid  int64
		}{
			{model.ProfitShareRolePlayer, rule.PlayerRatio, o.PlayerID},
			{model.ProfitShareRoleClub, rule.ClubRatio, o.ClubID},
			{model.ProfitShareRolePlatform, rule.PlatformRatio, 0},
		}
		// 分销商(上级分销)
		var distRel model.DistributorRelation
		var distID int64
		if err := tx.Where("subordinate_id = ? AND is_valid = 1", o.UserID).First(&distRel).Error; err == nil {
			distID = distRel.SuperiorID
			shares = append(shares, struct {
				role  string
				ratio float64
				uid   int64
			}{model.ProfitShareRoleDistributor, rule.DistributorRatio, distID})
		}
		allocated := int64(0)
		for _, s := range shares {
			if s.ratio <= 0 || (s.uid == 0 && s.role != model.ProfitShareRolePlatform) {
				continue
			}
			amount := int64(float64(o.Amount) * s.ratio / 100)
			if amount <= 0 {
				continue
			}
			allocated += amount
			rec := &model.ProfitShareRecord{
				OrderID: orderID, UserID: s.uid, Role: s.role,
				Amount: amount, Ratio: s.ratio, Status: 1,
				CreatedAt: nowTimePtr(), UpdatedAt: nowTimePtr(),
			}
			if s.role == model.ProfitShareRolePlatform {
				rec.UserID = 0
			}
			if err := tx.Create(rec).Error; err != nil {
				return err
			}
			// 入账(平台不入用户余额)
			if s.uid > 0 && s.role != model.ProfitShareRolePlatform {
				if err := tx.Model(&model.User{}).Where("id = ?", s.uid).
					UpdateColumn("balance", gorm.Expr("balance + ?", amount)).Error; err != nil {
					return err
				}
			}
			// 分销佣金记录
			if s.role == model.ProfitShareRoleDistributor && distID > 0 {
				if err := tx.Create(&model.DistributorCommission{
					OrderID: orderID, DistributorID: distID, Amount: amount,
					Ratio: s.ratio, Level: int8(distRel.Level),
					Status: model.DistributorCommissionStatusSettled,
					CreatedAt: nowTimePtr(), UpdatedAt: nowTimePtr(),
				}).Error; err != nil {
					return err
				}
			}
		}
		// 尾差归平台(避免精度损失导致总和不足)
		if allocated < o.Amount {
			remain := o.Amount - allocated
			_ = tx.Create(&model.ProfitShareRecord{
				OrderID: orderID, Role: model.ProfitShareRolePlatform,
				Amount: remain, Ratio: 100 - (float64(allocated) / float64(o.Amount) * 100),
				Status: 1, CreatedAt: nowTimePtr(), UpdatedAt: nowTimePtr(),
			}).Error
		}
		return tx.Model(&model.Order{}).Where("id = ?", orderID).
			Updates(map[string]interface{}{"status": model.OrderStatusSettled, "settled_at": nowTimePtr()}).Error
	})
}

// ListProfitShareByOrder 查询订单分账明细
func ListProfitShareByOrder(orderID int64) ([]model.ProfitShareRecord, error) {
	return paymentRepo.ListProfitShareByOrder(orderID)
}
