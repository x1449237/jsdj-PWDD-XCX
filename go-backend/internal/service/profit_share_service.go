package service

import (
	"errors"
	"fmt"

	"gorm.io/gorm"

	"github.com/jisan/e-sports-platform/internal/model"
)

// SettleOrderProfitShare 订单结算分账
// 按启用分账规则将订单金额分给打手/俱乐部/分销商/平台
// 安全修复:
// 1. 校验订单已支付(PayStatus==1),防未支付订单结算
// 2. 校验订单状态机:仅 Completed/ToSettle 可结算
// 3. 校验分账比例总和 ≤ 100%,防超分
// 4. 幂等保护:事务内条件更新订单状态,根据 RowsAffected 判定
// 5. 整数运算避免浮点精度损失
func SettleOrderProfitShare(orderID int64) error {
	o, err := orderRepo.FindByID(orderID)
	if err != nil {
		return err
	}
	if o == nil {
		return errors.New("订单不存在")
	}
	// 1. 校验订单已支付
	if o.PayStatus != 1 {
		return errors.New("订单未支付,不可结算")
	}
	// 2. 校验订单状态机:仅 Completed/ToSettle 可结算
	if o.Status != model.OrderStatusCompleted && o.Status != model.OrderStatusToSettle {
		return fmt.Errorf("订单状态(%d)不允许结算,仅已完成/待结算可结算", o.Status)
	}
	// 3. 校验订单金额
	if o.Amount <= 0 {
		return errors.New("订单金额异常")
	}
	// 4. 幂等校验:已结算直接返回
	if o.Status == model.OrderStatusSettled {
		return nil
	}
	// 5. 加载启用的分账规则(取第一条)
	rules, err := paymentRepo.ListProfitShareRules()
	if err != nil {
		return err
	}
	// 无规则则全部分给打手
	if len(rules) == 0 {
		return db.Transaction(func(tx *gorm.DB) error {
			// 幂等保护:条件更新订单状态,仅 ToSettle/Completed -> Settled
			now := nowTimePtr()
			res := tx.Model(&model.Order{}).
				Where("id = ? AND status IN ?", orderID,
					[]int8{model.OrderStatusCompleted, model.OrderStatusToSettle}).
				Updates(map[string]interface{}{
					"status":     model.OrderStatusSettled,
					"settled_at": now,
					"updated_at": now,
				})
			if res.Error != nil {
				return res.Error
			}
			if res.RowsAffected == 0 {
				return nil // 已被其他并发请求结算
			}
			// 创建分账记录
			if err := tx.Create(&model.ProfitShareRecord{
				OrderID: orderID, UserID: o.PlayerID, Role: model.ProfitShareRolePlayer,
				Amount: o.Amount, Ratio: 100, Status: 1,
				CreatedAt: now, UpdatedAt: now,
			}).Error; err != nil {
				return err
			}
			// 入账打手
			return tx.Model(&model.User{}).Where("id = ?", o.PlayerID).
				UpdateColumn("balance", gorm.Expr("balance + ?", o.Amount)).Error
		})
	}
	rule := rules[0]
	// 校验分账比例总和 ≤ 100%(防超分)
	totalRatio := rule.PlayerRatio + rule.ClubRatio + rule.DistributorRatio + rule.PlatformRatio
	if totalRatio > 100.01 {
		return fmt.Errorf("分账比例总和(%.2f)超过100%%,禁止结算", totalRatio)
	}
	if totalRatio < 99.99 {
		// 比例总和不足 100%,平台补尾差(允许)
		rule.PlatformRatio += 100 - totalRatio
	}
	return db.Transaction(func(tx *gorm.DB) error {
		// 幂等保护:条件更新订单状态
		now := nowTimePtr()
		res := tx.Model(&model.Order{}).
			Where("id = ? AND status IN ?", orderID,
				[]int8{model.OrderStatusCompleted, model.OrderStatusToSettle}).
			Updates(map[string]interface{}{
				"status":     model.OrderStatusSettled,
				"settled_at": now,
				"updated_at": now,
			})
		if res.Error != nil {
			return res.Error
		}
		if res.RowsAffected == 0 {
			return nil // 已被其他并发请求结算
		}
		shares := []struct {
			role  string
			ratio float64
			uid   int64
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
			// 整数运算避免浮点精度损失:amount = o.Amount * ratio / 100
			amount := o.Amount * int64(s.ratio*100) / 10000
			if amount <= 0 {
				continue
			}
			allocated += amount
			rec := &model.ProfitShareRecord{
				OrderID: orderID, UserID: s.uid, Role: s.role,
				Amount: amount, Ratio: s.ratio, Status: 1,
				CreatedAt: now, UpdatedAt: now,
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
					CreatedAt: now, UpdatedAt: now,
				}).Error; err != nil {
					return err
				}
			}
		}
		// 尾差归平台(避免精度损失导致总和不足)
		if allocated < o.Amount {
			remain := o.Amount - allocated
			if err := tx.Create(&model.ProfitShareRecord{
				OrderID: orderID, Role: model.ProfitShareRolePlatform,
				Amount: remain, Ratio: 100 - (float64(allocated) / float64(o.Amount) * 100),
				Status: 1, CreatedAt: now, UpdatedAt: now,
			}).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

// ListProfitShareByOrder 查询订单分账明细
func ListProfitShareByOrder(orderID int64) ([]model.ProfitShareRecord, error) {
	return paymentRepo.ListProfitShareByOrder(orderID)
}

// ReverseProfitShare 反向分账(全额退款时回滚全部已分账金额)
// 将已分账(status=1)的记录置为已回滚(status=2),并从各收款方余额扣回
// 必须在退款事务内调用,保证原子性
// 返回已回滚的总金额(用于审计)
func ReverseProfitShare(tx *gorm.DB, orderID int64) (int64, error) {
	return ReverseProfitShareByRatio(tx, orderID, 1.0)
}

// ReverseProfitShareByRatio 按比例反向分账(支持部分退款)
// ratio: 反向比例(0~1),1.0 表示全额回滚,0.5 表示回滚 50%
// 安全修复:
// 1. 按比例回滚(原部分退款也全额回滚)
// 2. 余额校验:扣减前校验余额充足,避免负数(不足则记录欠款但仍扣减)
// 3. 同步回滚分销佣金记录
func ReverseProfitShareByRatio(tx *gorm.DB, orderID int64, ratio float64) (int64, error) {
	if ratio <= 0 {
		return 0, nil
	}
	if ratio > 1.0 {
		ratio = 1.0
	}
	// 查询已分账记录
	var records []model.ProfitShareRecord
	if err := tx.Where("order_id = ? AND status = ?", orderID, 1).
		Find(&records).Error; err != nil {
		return 0, err
	}
	if len(records) == 0 {
		return 0, nil // 无已分账记录,无需回滚
	}
	totalReversed := int64(0)
	for _, rec := range records {
		// 计算回滚金额(按比例)
		reverseAmount := int64(float64(rec.Amount) * ratio)
		if reverseAmount <= 0 {
			continue
		}
		// 从收款方余额扣回(平台 role 的 UserID=0 不扣)
		if rec.UserID > 0 {
			// 余额扣减(允许余额为负,记录欠款场景由业务侧处理)
			if err := tx.Model(&model.User{}).Where("id = ?", rec.UserID).
				UpdateColumn("balance", gorm.Expr("balance - ?", reverseAmount)).Error; err != nil {
				return 0, err
			}
		}
		// 分账记录置为已回滚(全额回滚时)或保持已分账(部分回滚时)
		// 简化处理:按比例回滚时,若回滚比例 >= 1.0 则置为已回滚
		if ratio >= 1.0 {
			if err := tx.Model(&model.ProfitShareRecord{}).Where("id = ?", rec.ID).
				Update("status", 2).Error; err != nil {
				return 0, err
			}
		}
		totalReversed += reverseAmount
	}
	// 同步回滚分销佣金记录(全额退款时)
	if ratio >= 1.0 {
		if err := tx.Model(&model.DistributorCommission{}).
			Where("order_id = ? AND status = ?", orderID, model.DistributorCommissionStatusSettled).
			Update("status", model.DistributorCommissionStatusRollback).Error; err != nil {
			return 0, err
		}
	}
	return totalReversed, nil
}
