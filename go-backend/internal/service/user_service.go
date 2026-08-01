package service

import (
	"errors"
	"time"

	"gorm.io/gorm"

	"github.com/jisan/e-sports-platform/internal/model"
	"github.com/jisan/e-sports-platform/internal/utils"
)

// UserProfile 用户资料(脱敏后)
type UserProfile struct {
	*model.User
	IsMinor      bool   `json:"is_minor"`
	PhoneMasked  string `json:"phone_masked"`
	IDCardMasked string `json:"id_card_masked,omitempty"`
	RealNameMask string `json:"real_name_mask,omitempty"`
}

// GetUserProfile 获取用户资料
func GetUserProfile(userID int64) (*UserProfile, error) {
	u, err := userRepo.FindByID(userID)
	if err != nil {
		return nil, err
	}
	if u == nil {
		return nil, errors.New("用户不存在")
	}
	return &UserProfile{
		User:         u,
		IsMinor:      u.IsMinor == 1,
		PhoneMasked:  utils.MaskPhone(u.Phone),
		IDCardMasked: utils.MaskIDCard(u.IDCard),
		RealNameMask: utils.MaskName(u.RealName),
	}, nil
}

// UpdateUserProfile 更新用户资料(昵称/头像)
func UpdateUserProfile(userID int64, nickname, avatar string) (*model.User, error) {
	u, err := userRepo.FindByID(userID)
	if err != nil {
		return nil, err
	}
	if u == nil {
		return nil, errors.New("用户不存在")
	}
	fields := map[string]interface{}{}
	if nickname != "" {
		fields["nickname"] = nickname
	}
	if avatar != "" {
		fields["avatar"] = avatar
	}
	if len(fields) == 0 {
		return u, nil
	}
	fields["updated_at"] = nowTimePtr()
	if err := userRepo.Update(userID, fields); err != nil {
		return nil, err
	}
	return userRepo.FindByID(userID)
}

// SubmitRealname 提交实名认证(姓名+身份证号)
// 通过身份证号校验格式、提取年龄并标记是否未成年
func SubmitRealname(userID int64, realName, idCard string) error {
	if realName == "" {
		return errors.New("真实姓名不能为空")
	}
	if !utils.ValidateIDCard(idCard) {
		return errors.New("身份证号格式错误")
	}
	age, err := utils.GetAgeFromIDCard(idCard)
	if err != nil {
		return err
	}
	isMinor := utils.IsMinorByAge(age)
	fields := map[string]interface{}{
		"real_name":  realName,
		"id_card":    idCard,
		"is_minor":   0,
		"is_realname": 1,
		"updated_at": nowTimePtr(),
	}
	if isMinor {
		fields["is_minor"] = 1
	}
	if err := userRepo.Update(userID, fields); err != nil {
		return err
	}
	// 未成年则记录宵禁日志占位
	if isMinor {
		_ = db.Create(&model.MinorCurfewLog{
			UserID:    userID,
			Action:    model.MinorActionOrder,
			BlockedAt: nowTimePtr(),
			CreatedAt: nowTimePtr(),
		}).Error
	}
	return nil
}

// FaceVerify 活体检测校验(模拟:返回成功并写入会话ID)
func FaceVerify(userID int64, sessionID string) (string, error) {
	if sessionID == "" {
		return "", errors.New("活体会话ID不能为空")
	}
	// 实际项目应调用腾讯云/阿里云活体检测接口
	// 此处校验身份证已实名后通过
	u, err := userRepo.FindByID(userID)
	if err != nil {
		return "", err
	}
	if u == nil {
		return "", errors.New("用户不存在")
	}
	if u.IsRealname == 0 {
		return "", errors.New("请先完成实名认证")
	}
	// 返回会话ID供订单大额验证使用
	return sessionID, nil
}

// GetRealnameStatus 获取实名认证状态
func GetRealnameStatus(userID int64) (map[string]interface{}, error) {
	u, err := userRepo.FindByID(userID)
	if err != nil {
		return nil, err
	}
	if u == nil {
		return nil, errors.New("用户不存在")
	}
	return map[string]interface{}{
		"is_realname": u.IsRealname,
		"is_minor":    u.IsMinor,
		"real_name":   utils.MaskName(u.RealName),
	}, nil
}

// ToggleFavorite 收藏/取消收藏打手
// 简化实现:用 Redis Set 维护用户收藏集合
func ToggleFavorite(userID, playerID int64) (bool, error) {
	ctx, cancel := contextWithTimeout()
	defer cancel()
	key := cacheKey("favorite:" + itoa(userID))
	// 判断是否已收藏
	exists, err := redis.Client().SIsMember(ctx, key, playerID).Result()
	if err != nil {
		return false, err
	}
	if exists {
		_ = redis.Client().SRem(ctx, key, playerID).Err()
		return false, nil
	}
	_ = redis.Client().SAdd(ctx, key, playerID).Err()
	return true, nil
}

// ListFavoritePlayers 收藏的打手列表
func ListFavoritePlayers(userID int64) ([]int64, error) {
	ctx, cancel := contextWithTimeout()
	defer cancel()
	key := cacheKey("favorite:" + itoa(userID))
	ids, err := redis.Client().SMembers(ctx, key).Result()
	if err != nil {
		return nil, err
	}
	result := make([]int64, 0, len(ids))
	for _, s := range ids {
		var id int64
		if _, err := fmtSscanf(s, &id); err == nil {
			result = append(result, id)
		}
	}
	return result, nil
}

// GetUserByID 根据ID查询用户(供其他 service 调用)
func GetUserByID(userID int64) (*model.User, error) {
	return userRepo.FindByID(userID)
}

// BanUser 封禁用户
func BanUser(userID, operatorID int64, reason string) error {
	return db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Model(&model.User{}).Where("id = ?", userID).
			Updates(map[string]interface{}{"status": 0, "updated_at": nowTimePtr()}).Error; err != nil {
			return err
		}
		return tx.Create(&model.OperationLog{
			OperatorID:   operatorID,
			OperatorType: "admin",
			Action:       "ban_user",
			TargetType:   "user",
			TargetID:     userID,
			IP:           "",
			DeviceInfo:   reason,
			CreatedAt:    nowTimePtr(),
		}).Error
	})
}

// UnbanUser 解封用户
func UnbanUser(userID, operatorID int64) error {
	return userRepo.Update(userID, map[string]interface{}{
		"status":     1,
		"updated_at": nowTimePtr(),
	})
}

// ListUsers 平台用户列表
func ListUsers(page, pageSize int, status, role int8, keyword string) ([]model.User, int64, error) {
	return userRepo.List(page, pageSize, status, role, keyword)
}

// countUserOrdersStat 简化统计:用户订单数(供资料聚合)
func countUserOrdersStat(userID int64) int64 {
	var n int64
	_ = db.Model(&model.Order{}).Where("user_id = ?", userID).Count(&n).Error
	return n
}

// userLastLoginTime 用户最近登录(简化:用户最后更新时间)
func userLastLoginTime(userID int64) time.Time {
	u, _ := userRepo.FindByID(userID)
	if u != nil && u.UpdatedAt != nil {
		return *u.UpdatedAt
	}
	return time.Time{}
}
