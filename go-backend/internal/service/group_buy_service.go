package service

import (
	"errors"
	"time"

	"gorm.io/gorm"

	"github.com/jisan/e-sports-platform/internal/model"
)

// GetActiveActivities 获取进行中的拼团活动列表
func GetActiveActivities() ([]model.GroupBuyActivity, error) {
	var list []model.GroupBuyActivity
	now := time.Now()
	err := db.Model(&model.GroupBuyActivity{}).
		Where("status = ?", model.GroupBuyActivityStatusEnabled).
		Where("(start_at IS NULL OR start_at <= ?) AND (end_at IS NULL OR end_at >= ?)", now, now).
		Order("id DESC").Find(&list).Error
	return list, err
}

// GetGroups 获取某活动的群组列表
func GetGroups(activityID int64) ([]model.GroupBuyGroup, error) {
	var list []model.GroupBuyGroup
	err := db.Where("activity_id = ?", activityID).
		Order("id DESC").Find(&list).Error
	return list, err
}

// JoinGroupInput 加入拼团入参
type JoinGroupInput struct {
	ActivityID int64 `json:"activity_id"`
	GroupID    int64 `json:"group_id"`
}

// JoinGroup 参与拼团（新建团 或 加入现有团）
// 团满 min_members 人 → group 置 success + 生成每个 member 对应 order（折扣价）
func JoinGroup(uid int64, in *JoinGroupInput) (map[string]interface{}, error) {
	if in.ActivityID <= 0 && in.GroupID <= 0 {
		return nil, errors.New("活动ID或群组ID不能为空")
	}

	result := map[string]interface{}{}
	err := db.Transaction(func(tx *gorm.DB) error {
		var activity *model.GroupBuyActivity
		var group *model.GroupBuyGroup

		if in.ActivityID > 0 {
			a := &model.GroupBuyActivity{}
			if err := tx.First(a, in.ActivityID).Error; err != nil {
				if errors.Is(err, gorm.ErrRecordNotFound) {
					return errors.New("拼团活动不存在")
				}
				return err
			}
			if a.Status != model.GroupBuyActivityStatusEnabled {
				return errors.New("拼团活动未开启")
			}
			activity = a
		}

		if in.GroupID > 0 {
			g := &model.GroupBuyGroup{}
			if err := tx.First(g, in.GroupID).Error; err != nil {
				if errors.Is(err, gorm.ErrRecordNotFound) {
					return errors.New("拼团团组不存在")
				}
				return err
			}
			group = g
			if activity == nil {
				a := &model.GroupBuyActivity{}
				if err := tx.First(a, g.ActivityID).Error; err != nil {
					return err
				}
				activity = a
			}
		}

		if activity == nil {
			return errors.New("拼团活动不存在")
		}

		now := time.Now()
		if activity.StartAt != nil && activity.StartAt.After(now) {
			return errors.New("拼团活动尚未开始")
		}
		if activity.EndAt != nil && activity.EndAt.Before(now) {
			return errors.New("拼团活动已结束")
		}
		_ = activity.MinSpend

		var cnt int64
		if err := tx.Model(&model.GroupBuyMember{}).
			Where("uid = ? AND activity_id = ?", uid, activity.ID).
			Count(&cnt).Error; err != nil {
			return err
		}
		if cnt > 0 {
			return errors.New("您已参加该拼团活动")
		}

		if group == nil {
			g := &model.GroupBuyGroup{}
			err := tx.Where("activity_id = ? AND status = ? AND member_count < ?",
				activity.ID, model.GroupBuyGroupStatusPending, activity.MaxMembers).
				Order("id ASC").First(g).Error
			if err != nil {
				if !errors.Is(err, gorm.ErrRecordNotFound) {
					return err
				}
				expireAt := now.Add(24 * time.Hour)
				group = &model.GroupBuyGroup{
					ActivityID:  activity.ID,
					LeaderUID:   uid,
					MemberCount: 0,
					Status:      model.GroupBuyGroupStatusPending,
					ExpiredAt:   &expireAt,
					CreatedAt:   nowTimePtr(),
					UpdatedAt:   nowTimePtr(),
				}
				if err := tx.Create(group).Error; err != nil {
					return err
				}
			} else {
				group = g
			}
		} else {
			if group.Status != model.GroupBuyGroupStatusPending && group.Status != model.GroupBuyGroupStatusFull {
				return errors.New("该拼团团组不可加入")
			}
			if group.MemberCount >= activity.MaxMembers {
				return errors.New("该拼团团组已满员")
			}
		}

		isLeader := int8(0)
		if group.LeaderUID == 0 || group.LeaderUID == uid {
			isLeader = 1
			if group.LeaderUID == 0 {
				_ = tx.Model(group).Update("leader_uid", uid).Error
			}
		}
		member := &model.GroupBuyMember{
			GroupID:    group.ID,
			ActivityID: activity.ID,
			UID:        uid,
			JoinedAt:   nowTimePtr(),
			IsLeader:   isLeader,
			CreatedAt:  nowTimePtr(),
		}
		if err := tx.Create(member).Error; err != nil {
			return err
		}

		if err := tx.Model(group).
			UpdateColumn("member_count", gorm.Expr("member_count + ?", 1)).Error; err != nil {
			return err
		}

		if err := tx.First(group, group.ID).Error; err != nil {
			return err
		}

		if group.MemberCount >= activity.MinMembers {
			if err := tx.Model(group).Updates(map[string]interface{}{
				"status":     model.GroupBuyGroupStatusSuccess,
				"updated_at": nowTimePtr(),
			}).Error; err != nil {
				return err
			}

			var members []model.GroupBuyMember
			if err := tx.Where("group_id = ?", group.ID).Find(&members).Error; err != nil {
				return err
			}
			for _, m := range members {
				discountAmount := int64(float64(activity.MinSpend) * activity.DiscountRatio)
				if discountAmount <= 0 {
					discountAmount = activity.MinSpend
				}
				order := &model.Order{
					OrderNo:   genOrderNo(),
					Type:      model.OrderTypeInstant,
					UserID:    m.UID,
					ClubID:    0,
					ServiceID: activity.ServiceID,
					Amount:    discountAmount,
					Status:    model.OrderStatusPending,
					PayStatus: 0,
					CreatedAt: nowTimePtr(),
					UpdatedAt: nowTimePtr(),
				}
				if err := tx.Create(order).Error; err != nil {
					return err
				}
				_ = tx.Create(&model.OrderStatusLog{
					OrderID:      order.ID,
					FromStatus:   -1,
					ToStatus:     model.OrderStatusPending,
					OperatorID:   m.UID,
					OperatorType: "system",
					Reason:       "拼团成功生成订单",
					CreatedAt:    nowTimePtr(),
				}).Error
				_ = tx.Model(&model.GroupBuyMember{}).Where("id = ?", m.ID).Update("order_id", order.ID).Error
			}
		} else if group.MemberCount >= activity.MaxMembers {
			_ = tx.Model(group).Updates(map[string]interface{}{
				"status":     model.GroupBuyGroupStatusFull,
				"updated_at": nowTimePtr(),
			}).Error
		}

		result["group_id"] = group.ID
		result["activity_id"] = activity.ID
		result["member_count"] = group.MemberCount
		result["status"] = group.Status
		return nil
	})
	return result, err
}

// ExpireGroup 处理拼团超时（未达到 min_members → 置 failed + 退款）
func ExpireGroup(groupID int64) error {
	return db.Transaction(func(tx *gorm.DB) error {
		g := &model.GroupBuyGroup{}
		if err := tx.First(g, groupID).Error; err != nil {
			return err
		}
		if g.Status != model.GroupBuyGroupStatusPending && g.Status != model.GroupBuyGroupStatusFull {
			return nil
		}
		if err := tx.Model(g).Updates(map[string]interface{}{
			"status":     model.GroupBuyGroupStatusFailed,
			"updated_at": nowTimePtr(),
		}).Error; err != nil {
			return err
		}
		var members []model.GroupBuyMember
		if err := tx.Where("group_id = ?", groupID).Find(&members).Error; err != nil {
			return err
		}
		for _, m := range members {
			if m.OrderID > 0 {
				_ = tx.Model(&model.Order{}).Where("id = ?", m.OrderID).
					Updates(map[string]interface{}{
						"status":        model.OrderStatusRefunded,
						"pay_status":    2,
						"refund_amount": gorm.Expr("amount"),
						"updated_at":    nowTimePtr(),
					}).Error
				_ = tx.Create(&model.OrderStatusLog{
					OrderID:      m.OrderID,
					FromStatus:   model.OrderStatusPending,
					ToStatus:     model.OrderStatusRefunded,
					OperatorID:   0,
					OperatorType: "system",
					Reason:       "拼团超时失败，自动退款",
					CreatedAt:    nowTimePtr(),
				}).Error
			}
		}
		return nil
	})
}

// GetGroupBuyActivities 拼团活动列表（对 marketing_service 兼容接口）
func GetGroupBuyActivities() ([]model.GroupBuyActivity, error) {
	return GetActiveActivities()
}

// JoinGroupBuy 参与拼团（兼容旧接口，默认按 activityID 加入）
func JoinGroupBuy(userID, activityID int64) (map[string]interface{}, error) {
	return JoinGroup(userID, &JoinGroupInput{ActivityID: activityID})
}
