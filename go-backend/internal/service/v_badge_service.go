package service

import (
	"errors"

	"gorm.io/gorm"

	"github.com/jisan/e-sports-platform/internal/model"
)

// GrantClubVBadge 给俱乐部授予V标
// clubType=1(企业) → v_badge_type=1(蓝V)；=2(个人) → v_badge_type=2(绿V)
// 同步给俱乐部所有成员写入 UserVBadge 记录
func GrantClubVBadge(clubID, clubType int64) error {
	if clubID <= 0 {
		return errors.New("俱乐部ID无效")
	}
	return db.Transaction(func(tx *gorm.DB) error {
		club := &model.Club{}
		if err := tx.First(club, clubID).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return errors.New("俱乐部不存在")
			}
			return err
		}
		badgeType := int8(0)
		badgeTypeName := model.VBadgeTypeNone
		switch clubType {
		case 1:
			badgeType = 1
			badgeTypeName = model.VBadgeTypeBlue
		case 2:
			badgeType = 2
			badgeTypeName = model.VBadgeTypeGreen
		default:
			return errors.New("未知俱乐部类型")
		}
		if err := tx.Model(club).Updates(map[string]interface{}{
			"v_badge_type": badgeType,
			"updated_at":   nowTimePtr(),
		}).Error; err != nil {
			return err
		}
		// 给所有有效成员写入V标
		var members []model.ClubMember
		if err := tx.Where("club_id = ? AND status = 1", clubID).Find(&members).Error; err != nil {
			return err
		}
		now := nowTimePtr()
		for _, m := range members {
			// 先删旧的再写新的（幂等）
			_ = tx.Where("user_id = ? AND club_id = ? AND badge_type IN ?",
				m.UserID, clubID, []string{model.VBadgeTypeBlue, model.VBadgeTypeGreen}).
				Delete(&model.UserVBadge{}).Error
			if err := tx.Create(&model.UserVBadge{
				UserID:    m.UserID,
				ClubID:    clubID,
				BadgeType: badgeTypeName,
				Status:    1,
				GrantedAt: now,
				CreatedAt: now,
				UpdatedAt: now,
			}).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

// RevokeClubVBadge 撤销/驳回/冻结/注销俱乐部时 撤销V标
func RevokeClubVBadge(clubID int64) error {
	if clubID <= 0 {
		return errors.New("俱乐部ID无效")
	}
	return db.Transaction(func(tx *gorm.DB) error {
		club := &model.Club{}
		err := tx.First(club, clubID).Error
		if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
			return err
		}
		if err == nil {
			if err := tx.Model(club).Updates(map[string]interface{}{
				"v_badge_type": int8(0),
				"updated_at":   nowTimePtr(),
			}).Error; err != nil {
				return err
			}
		}
		// 失效该俱乐部下所有蓝/绿V标
		return tx.Model(&model.UserVBadge{}).
			Where("club_id = ? AND badge_type IN ?", clubID, []string{model.VBadgeTypeBlue, model.VBadgeTypeGreen}).
			Updates(map[string]interface{}{"status": 0, "updated_at": nowTimePtr()}).Error
	})
}

// GrantPlatformGoldV 授予平台金V
func GrantPlatformGoldV(userID int64) error {
	if userID <= 0 {
		return errors.New("用户ID无效")
	}
	now := nowTimePtr()
	return db.Transaction(func(tx *gorm.DB) error {
		// 先失效旧的金V
		_ = tx.Where("user_id = ? AND badge_type = ? AND club_id = 0",
			userID, model.VBadgeTypeGold).Delete(&model.UserVBadge{}).Error
		return tx.Create(&model.UserVBadge{
			UserID:    userID,
			ClubID:    0,
			BadgeType: model.VBadgeTypeGold,
			Status:    1,
			GrantedAt: now,
			CreatedAt: now,
			UpdatedAt: now,
		}).Error
	})
}

// RevokePlatformGoldV 撤销平台金V
func RevokePlatformGoldV(userID int64) error {
	if userID <= 0 {
		return errors.New("用户ID无效")
	}
	return db.Model(&model.UserVBadge{}).
		Where("user_id = ? AND badge_type = ? AND club_id = 0", userID, model.VBadgeTypeGold).
		Updates(map[string]interface{}{"status": 0, "updated_at": nowTimePtr()}).Error
}
