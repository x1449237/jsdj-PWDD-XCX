package service

import (
	"context"
	"crypto/rand"
	"errors"
	"fmt"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/jisan/e-sports-platform/internal/model"
	"github.com/jisan/e-sports-platform/internal/utils"
)

// FaceVerifyFee 每次活体认证费用(单位:分,2元=200分)
// 先收费再认证:非缓存命中时先扣费,再调用第三方SDK认证(认证失败不退款,因第三方按调用次数计费)
// 缓存命中(7天内复用)/频控拦截(每日超5次)不收取
const FaceVerifyFee int64 = 200

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

// FaceVerify 活体检测校验(先收费再认证)
// 1. 7天缓存:通过的 7天内复用(realname_caches表,不收费)
// 2. 频控:每天5次(face_verify_rate_limits表 + Redis,超限不收费直接拒绝)
// 3. 先收费:扣2元(200分)再调用第三方SDK认证(认证失败不退款,第三方按调用次数计费)
// 4. 调用腾讯云/阿里云标准SDK(注释给出示例代码结构)，沙箱模拟通过率
func FaceVerify(userID int64, sessionID string) (string, error) {
	if sessionID == "" {
		return "", errors.New("活体会话ID不能为空")
	}
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
	now := time.Now()
	today0 := startOfToday()
	tomorrow0 := today0.Add(24 * time.Hour)

	// Step 1. 检查 7 天缓存(realname_caches 表:status=pass + expire > now)
	cacheKeyStr := fmt.Sprintf("face_verify:%d", userID)
	// 先查 Redis 快速命中
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	if redis != nil {
		v, rerr := redis.Get(ctx, cacheKey(cacheKeyStr))
		if rerr == nil && v == "pass" {
			_ = userRepo.Update(userID, map[string]interface{}{
				"face_session_id": sessionID,
				"face_verified_at": now,
			})
			return sessionID, nil
		}
	}
	// 查 DB realname_caches
	var cached model.RealnameCache
	found := false
	err = db.Where("uid = ? AND verify_type = ? AND status = ? AND expire_at > ?",
		userID, "face", "pass", now).Order("id DESC").First(&cached).Error
	if err == nil {
		found = true
	}
	if found {
		// 命中缓存：写 Redis 快速缓存并返回
		if redis != nil {
			_ = redis.Set(ctx, cacheKey(cacheKeyStr), "pass", 24*time.Hour)
		}
		_ = userRepo.Update(userID, map[string]interface{}{
			"face_session_id":   sessionID,
			"face_verified_at":  now,
		})
		return sessionID, nil
	}

	// Step 2. 频控:每天最多 5 次(Redis 计数 + DB 记录)
	rateLimitKey := cacheKey(fmt.Sprintf("face_verify:ratelimit:%d:%s", userID, today0.Format("2006-01-02")))
	var todayUsed int
	if redis != nil {
		v, rerr := redis.Get(ctx, rateLimitKey)
		if rerr == nil && v != "" {
			todayUsed = int(atoi(v))
		}
	}
	if todayUsed <= 0 {
		// 从 DB face_verify_rate_limits 兜底
		var cnt int64
		_ = db.Model(&model.FaceVerifyRateLimit{}).
			Where("user_id = ? AND created_at >= ? AND created_at < ?", userID, today0, tomorrow0).
			Count(&cnt).Error
		todayUsed = int(cnt)
	}
	if todayUsed >= 5 {
		return "", errors.New("今日活体检测次数已达上限(5次)，请明日再试")
	}

	// Step 3. 先收费:扣费 + 写 DB 频控记录(事务原子,防并发超额扣费)
	// 先收费再认证:在调用第三方SDK前先扣2元,认证失败不退款(第三方按调用次数计费)
	// 缓存命中/频控拦截不收取;余额不足拒绝调用,避免平台垫付
	err = db.Transaction(func(tx *gorm.DB) error {
		// 行锁用户记录,防并发超额扣费
		var locked model.User
		if err := tx.Where("id = ?", userID).
			Clauses(clause.Locking{Strength: "UPDATE"}).
			First(&locked).Error; err != nil {
			return err
		}
		if locked.Balance < FaceVerifyFee {
			return fmt.Errorf("余额不足,活体认证需%.2f元,当前余额%.2f元,请先充值",
				float64(FaceVerifyFee)/100, float64(locked.Balance)/100)
		}
		// 原子扣减余额(条件更新防超额)
		res := tx.Model(&model.User{}).
			Where("id = ? AND balance >= ?", userID, FaceVerifyFee).
			UpdateColumn("balance", gorm.Expr("balance - ?", FaceVerifyFee))
		if res.Error != nil {
			return res.Error
		}
		if res.RowsAffected == 0 {
			return fmt.Errorf("余额不足,活体认证需%.2f元", float64(FaceVerifyFee)/100)
		}
		// 写频控记录(带扣费金额,可追溯)
		return tx.Create(&model.FaceVerifyRateLimit{
			UserID:    userID,
			IP:        "face-api",
			Count:     1,
			Date:      today0.Format("2006-01-02"),
			Fee:       FaceVerifyFee,
			CreatedAt: &now,
			UpdatedAt: &now,
		}).Error
	})
	if err != nil {
		return "", err
	}
	if redis != nil {
		_ = redis.Set(ctx, rateLimitKey, fmt.Sprintf("%d", todayUsed+1),
			time.Until(tomorrow0))
	}

	// Step 4. 调用第三方SDK(腾讯云/阿里云 FaceCompare/LiveDetectFour)
	// ——腾讯云 SDK 示例结构——
	// import (
	//   "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common"
	//   faceid "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/faceid/v20180301"
	// )
	// cred := common.NewCredential(SecretId, SecretKey)
	// cpf := profile.NewClientProfile()
	// client, _ := faceid.NewClient(cred, "ap-beijing", cpf)
	// req := faceid.NewLivenessCompareRequest()
	// req.SessionId = common.StringPtr(sessionID)
	// resp, err := client.LivenessCompare(req)
	// passed := resp.Response != nil && *resp.Response.Sim >= 70
	// ——阿里云 SDK 示例结构——
	// import facebody20191230 "github.com/alibabacloud-go/facebody-20191230/v4/client"
	// client, _ = facebody20191230.NewClient(...)
	// req := &facebody20191230.ExecuteServerSideVerificationRequest{SceneId:tea.Int64(1)}
	// passed := result passed
	passed := false
	verifyMsg := "沙箱模拟通过"
	// 沙箱模式:sessionID 以 "fail_" 前缀=失败 否则 95%概率通过
	if len(sessionID) >= 5 && sessionID[:5] == "fail_" {
		passed = false
		verifyMsg = "沙箱模拟失败(命中fail_前缀)"
	} else {
		// 用 crypto/rand 模拟 95% 通过率
		var b [1]byte
		_, _ = rand.Read(b[:])
		passed = int(b[0])%100 < 95
		if !passed {
			verifyMsg = "沙箱模拟随机不通过(5%)"
		}
	}
	if !passed {
		return "", errors.New("活体检测失败: " + verifyMsg)
	}

	// Step 5. 写入 realname_caches 缓存 7 天
	expireAt := now.AddDate(0, 0, 7)
	_ = db.Create(&model.RealnameCache{
		UserID:         userID,
		LastVerifyTime: &now,
		ExpireTime:     &expireAt,
		VerifySession:  sessionID,
		CreatedAt:      &now,
		UpdatedAt:      &now,
	})
	if redis != nil {
		_ = redis.Set(ctx, cacheKey(cacheKeyStr), "pass", 24*time.Hour)
	}

	// Step 6. 更新用户表
	_ = userRepo.Update(userID, map[string]interface{}{
		"face_session_id":  sessionID,
		"face_verified_at": now,
	})
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

// UpdateCreditScore 信用分变更：写入 credit_log 并更新 users.credit_score
// 信用分范围 0-150；delta 为正时加分、为负时扣分；ref_id 关联订单/评价等引用ID
func UpdateCreditScore(uid int64, delta int, reason string, refID int64) error {
	if uid <= 0 {
		return errors.New("用户ID不能为空")
	}
	if delta == 0 {
		return nil
	}
	return db.Transaction(func(tx *gorm.DB) error {
		// 先读取当前信用分（for update 简化为 SELECT + UPDATE）
		var before int
		u := &model.User{}
		if err := tx.Select("id, credit_score").Where("id = ?", uid).First(u).Error; err != nil {
			return err
		}
		before = u.CreditScore
		// 计算更新后值，保证在 [0, 150] 内
		after := before + delta
		if after < 0 {
			after = 0
		}
		if after > 150 {
			after = 150
		}
		// 更新 users 表
		if err := tx.Model(&model.User{}).Where("id = ?", uid).
			Updates(map[string]interface{}{
				"credit_score": after,
				"updated_at":   nowTimePtr(),
			}).Error; err != nil {
			return err
		}
		// 写流水 credit_log
		if err := tx.Create(&model.CreditLog{
			UID:       uid,
			Delta:     delta,
			After:     after,
			Reason:    reason,
			RefID:     refID,
			CreatedAt: nowTimePtr(),
		}).Error; err != nil {
			return err
		}
		return nil
	})
}

// IncrPlayerRejectCount 玩家拒单计数，超过3次扣信用分 -5
func IncrPlayerRejectCount(playerID int64) (int, error) {
	ctx, cancel := contextWithTimeout()
	defer cancel()
	key := cacheKey("reject_cnt:" + itoa(playerID) + ":" + time.Now().Format("2006-01-02"))
	cntStr, _ := redis.Get(ctx, key)
	cnt := 0
	if cntStr != "" {
		fmt.Sscanf(cntStr, "%d", &cnt)
	}
	cnt++
	_ = redis.Set(ctx, key, itoa(int64(cnt)), 26*time.Hour)
	if cnt > 3 {
		_ = UpdateCreditScore(playerID, -5, "打手当日累计拒单超过3次", 0)
	}
	return cnt, nil
}
