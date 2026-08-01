package service

import (
	"errors"
	"fmt"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/jisan/e-sports-platform/internal/model"
)

// PlayerServiceItem 打手服务项目(简化:复用 Order ServiceID 字段)
// 实际项目应有独立 services 表，此处用 ClubMember + 简化字段示意

// GetGrabOrderList 可抢单列表(打手只能抢本俱乐部或公开订单)
func GetGrabOrderList(playerID, clubID int64, page, pageSize int) ([]model.Order, int64, error) {
	return orderRepo.ListGrabOrders(clubID, page, pageSize)
}

// GrabOrder 打手抢单(原子操作 + 分布式锁防并发)
// 支持普通订单（一单一打手）和车队订单（多人匹配 TeamCount）
// 安全修复:
// 1. 新增 playerClubID 参数,校验打手所属俱乐部与订单俱乐部一致(防跨俱乐部抢单)
// 2. 校验 playerID != o.UserID(防客户抢自己的单)
// 3. 校验打手身份(必须是打手角色)
// 4. 使用安全的分布式锁(原 SetNX+Del 不安全,可能误删他人锁)
// 5. 车队满员用条件更新 WHERE status=team_pending(防并发双触发)
func GrabOrder(orderID, playerID, playerClubID int64) error {
	if orderID <= 0 || playerID <= 0 {
		return errors.New("参数无效")
	}
	// 校验打手身份
	u, err := userRepo.FindByID(playerID)
	if err != nil {
		return err
	}
	if u == nil {
		return errors.New("用户不存在")
	}
	if u.Role&model.RolePlayer == 0 {
		return errors.New("仅打手可抢单")
	}
	if u.Status != 1 {
		return errors.New("账号已被禁用")
	}
	// 使用安全的分布式锁(基于项目已实现的 DistributedLock)
	if distLock != nil {
		lockKey := cacheKey("grab:" + itoa(orderID))
		ctx, cancel := contextWithTimeout()
		defer cancel()
		token, err := distLock.TryLock(ctx, lockKey, 10*time.Second)
		if err != nil {
			return errors.New("订单正在被抢,请稍后")
		}
		defer distLock.Unlock(ctx, token)
	}

	o, err := orderRepo.FindByID(orderID)
	if err != nil {
		return err
	}
	if o == nil {
		return errors.New("订单不存在")
	}
	// 校验订单状态:仅待接单/车队匹配中可抢
	if o.Status != model.OrderStatusPending && o.Status != model.OrderStatusTeamPending {
		return errors.New("订单状态不允许抢单")
	}
	// 校验打手不能抢自己的单(防套利)
	if o.UserID == playerID {
		return errors.New("不可抢自己下的订单")
	}
	// 俱乐部归属校验:订单指定了俱乐部时,打手必须属于该俱乐部
	if o.ClubID > 0 && playerClubID > 0 && o.ClubID != playerClubID {
		return errors.New("无权抢该俱乐部的订单")
	}

	// 车队订单(OrderTypeTeam 或 TeamCount>1)：加入 team_members，满员后再置已接单
	if o.Type == model.OrderTypeTeam || o.TeamCount > 1 {
		tc := o.TeamCount
		if tc < 2 {
			tc = 2
		}
		if tc > 5 {
			tc = 5
		}
		// 校验打手是否已在队中
		var exist int64
		_ = db.Model(&model.OrderTeamMember{}).
			Where("order_id = ? AND player_id = ? AND status = ?", orderID, playerID, model.TeamMemberStatusJoined).Count(&exist).Error
		if exist > 0 {
			return errors.New("您已在该车队中")
		}
		// 写入车队成员
		now := nowTimePtr()
		if err := db.Create(&model.OrderTeamMember{
			OrderID:   orderID,
			PlayerID:  playerID,
			JoinedAt:  now,
			Status:    model.TeamMemberStatusJoined,
			CreatedAt: now,
		}).Error; err != nil {
			return err
		}
		// 统计当前人数
		var current int64
		_ = db.Model(&model.OrderTeamMember{}).
			Where("order_id = ? AND status = 1", orderID).Count(&current).Error
		if current >= int64(tc) {
			// 满员 -> 置为已接单(条件更新防并发双触发)
			res := db.Model(&model.Order{}).
				Where("id = ? AND status = ?", orderID, model.OrderStatusTeamPending).
				Updates(map[string]interface{}{
					"player_id":  playerID,
					"status":     model.OrderStatusAccepted,
					"updated_at": now,
				})
			if res.Error == nil && res.RowsAffected > 0 {
				_ = orderRepo.CreateStatusLog(&model.OrderStatusLog{
					OrderID: orderID, FromStatus: model.OrderStatusTeamPending, ToStatus: model.OrderStatusAccepted,
					OperatorID: playerID, OperatorType: "player", Reason: "车队满员,自动已接单",
					CreatedAt: now,
				})
			}
		} else if o.Status != model.OrderStatusTeamPending {
			// 未满员标记车队等待状态
			_ = orderRepo.Update(orderID, map[string]interface{}{
				"status":     model.OrderStatusTeamPending,
				"updated_at": now,
			})
			_ = orderRepo.CreateStatusLog(&model.OrderStatusLog{
				OrderID: orderID, FromStatus: o.Status, ToStatus: model.OrderStatusTeamPending,
				OperatorID: playerID, OperatorType: "player", Reason: "车队匹配中(" + itoa(int64(current)) + "/" + itoa(int64(tc)) + ")",
				CreatedAt: now,
			})
		}
		return nil
	}

	// 普通抢单
	grabbed, err := orderRepo.GrabOrder(orderID, playerID)
	if err != nil {
		return err
	}
	if !grabbed {
		return errors.New("抢单失败,订单已被他人接单或状态已变更")
	}
	_ = orderRepo.CreateStatusLog(&model.OrderStatusLog{
		OrderID: orderID, FromStatus: model.OrderStatusPending, ToStatus: model.OrderStatusAccepted,
		OperatorID: playerID, OperatorType: "player", Reason: "打手抢单",
		CreatedAt: nowTimePtr(),
	})
	return nil
}

// GetPlayerOrders 打手接单列表
func GetPlayerOrders(playerID int64, page, pageSize int, status int8) ([]model.Order, int64, error) {
	return orderRepo.ListPlayerOrders(playerID, page, pageSize, status)
}

// StartService 开始服务(已接单 -> 进行中)
func StartService(orderID, playerID int64) error {
	o, err := orderRepo.FindByID(orderID)
	if err != nil {
		return err
	}
	if o == nil {
		return errors.New("订单不存在")
	}
	if o.PlayerID != playerID {
		return errors.New("无权操作该订单")
	}
	if o.Status != model.OrderStatusAccepted {
		return errors.New("订单状态不允许开始服务")
	}
	return orderRepo.Update(orderID, map[string]interface{}{
		"status":     model.OrderStatusInProgress,
		"started_at": nowTimePtr(),
		"updated_at": nowTimePtr(),
	})
}

// CompleteService 完成服务(进行中 -> 待验收)
func CompleteService(orderID, playerID int64) error {
	o, err := orderRepo.FindByID(orderID)
	if err != nil {
		return err
	}
	if o == nil {
		return errors.New("订单不存在")
	}
	if o.PlayerID != playerID {
		return errors.New("无权操作该订单")
	}
	if o.Status != model.OrderStatusInProgress {
		return errors.New("订单状态不允许完成服务")
	}
	return orderRepo.Update(orderID, map[string]interface{}{
		"status":     model.OrderStatusToVerify,
		"ended_at":   nowTimePtr(),
		"updated_at": nowTimePtr(),
	})
}

// TransferOrder 转单(打手转给同俱乐部其他打手)
// 校验：同俱乐部、目标打手无进行中同订单冲突、写 order_status_log
func TransferOrder(orderID, fromPlayer, toPlayer int64) error {
	o, err := orderRepo.FindByID(orderID)
	if err != nil {
		return err
	}
	if o == nil {
		return errors.New("订单不存在")
	}
	if o.PlayerID != fromPlayer {
		return errors.New("无权转单")
	}
	if toPlayer <= 0 || fromPlayer == toPlayer {
		return errors.New("转单目标打手无效")
	}
	fromUser, err := userRepo.FindByID(fromPlayer)
	if err != nil || fromUser == nil {
		return errors.New("来源打手不存在")
	}
	toUser, err := userRepo.FindByID(toPlayer)
	if err != nil || toUser == nil {
		return errors.New("目标打手不存在")
	}
	// 校验必须同俱乐部
	if fromUser.ClubID == 0 || toUser.ClubID == 0 || fromUser.ClubID != toUser.ClubID {
		return errors.New("转单仅限同俱乐部成员")
	}
	// 校验 toPlayer 无进行中同订单（冲突）
	var conflict int64
	_ = db.Model(&model.Order{}).
		Where("player_id = ? AND id = ? AND status IN ?", toPlayer, orderID,
			[]int8{model.OrderStatusAccepted, model.OrderStatusInProgress}).Count(&conflict).Error
	if conflict > 0 {
		return errors.New("目标打手已在该订单中")
	}
	// 校验 toPlayer 是否有 3 个以上进行中订单（容量风控）
	var busy int64
	_ = db.Model(&model.Order{}).
		Where("player_id = ? AND status IN ?", toPlayer,
			[]int8{model.OrderStatusAccepted, model.OrderStatusInProgress}).Count(&busy).Error
	if busy >= 5 {
		return errors.New("目标打手进行中订单过多，无法转单")
	}
	// 执行转单 + 写状态日志
	now := nowTimePtr()
	return db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Model(&model.Order{}).Where("id = ?", orderID).
			Updates(map[string]interface{}{
				"player_id":  toPlayer,
				"updated_at": now,
			}).Error; err != nil {
			return err
		}
		return tx.Create(&model.OrderStatusLog{
			OrderID:      orderID,
			FromStatus:   o.Status,
			ToStatus:     o.Status,
			OperatorID:   fromPlayer,
			OperatorType: "player",
			Reason:       "转单: " + itoa(fromPlayer) + "→" + itoa(toPlayer),
			CreatedAt:    now,
		}).Error
	})
}

// GetEarnings 打手收益概览(已结算/待结算)
func GetEarnings(playerID int64) (map[string]int64, error) {
	frozen, err := paymentRepo.SumFrozenEarnings(playerID)
	if err != nil {
		return nil, err
	}
	settled, err := paymentRepo.SumSettledEarnings(playerID)
	if err != nil {
		return nil, err
	}
	// 已提现金额
	var withdrawn int64
	_ = db.Model(&model.Withdraw{}).Where("user_id = ? AND status = ?", playerID, model.WithdrawStatusPaid).
		Select("COALESCE(SUM(amount),0)").Scan(&withdrawn).Error
	return map[string]int64{
		"frozen":     frozen,
		"settled":    settled,
		"withdrawn":  withdrawn,
		"balance":    settled - withdrawn,
	}, nil
}

// GetFrozenEarnings 冻结收益明细
func GetFrozenEarnings(playerID int64) ([]model.Order, error) {
	var orders []model.Order
	err := db.Where("player_id = ? AND status IN ?", playerID, []int8{
		model.OrderStatusAccepted, model.OrderStatusInProgress, model.OrderStatusToVerify,
	}).Order("id DESC").Find(&orders).Error
	return orders, err
}

// ApplyWithdraw 申请提现
// 安全修复:
// 1. 事务内行锁用户记录,防并发超额提现(原余额校验在事务外,存在 TOCTOU)
// 2. 冻结金额:提现申请时原子扣减用户余额并冻结,审核通过后转为已审核
// 3. 余额校验改为读取锁定后的最新余额
func ApplyWithdraw(userID int64, amount int64, channel string, bankInfo map[string]string) (*model.Withdraw, error) {
	if amount <= 0 {
		return nil, errors.New("提现金额必须大于 0")
	}
	if channel == "" {
		return nil, errors.New("提现渠道不能为空")
	}
	u, err := userRepo.FindByID(userID)
	if err != nil {
		return nil, err
	}
	if u == nil {
		return nil, errors.New("用户不存在")
	}
	// 手续费与个税(简化:手续费 0.6%,个税 20%)
	fee := amount * 6 / 1000
	tax := (amount - fee) * 20 / 100
	net := amount - fee - tax
	if net <= 0 {
		return nil, errors.New("提现金额不足抵扣手续费")
	}
	w := &model.Withdraw{
		UserID:    userID,
		Amount:    amount,
		Fee:       fee,
		Tax:       tax,
		NetAmount: net,
		Channel:   channel,
		Status:    model.WithdrawStatusPending,
		BankCard:  bankInfo["bank_card"],
		BankName:  bankInfo["bank_name"],
		BankPhone: bankInfo["bank_phone"],
		IDCard:    bankInfo["id_card"],
		RealName:  bankInfo["real_name"],
		CreatedAt: nowTimePtr(),
		UpdatedAt: nowTimePtr(),
	}
	// 事务内:行锁用户 + 校验余额 + 原子扣减余额 + 创建提现记录
	err = db.Transaction(func(tx *gorm.DB) error {
		// 行锁用户记录
		var locked model.User
		if err := tx.Where("id = ?", userID).
			Clauses(clause.Locking{Strength: "UPDATE"}).
			First(&locked).Error; err != nil {
			return err
		}
		// 余额校验(锁定后重读)
		if locked.Balance < amount {
			return fmt.Errorf("余额不足(当前:%d,需:%d)", locked.Balance, amount)
		}
		// 原子扣减余额(冻结金额,审核通过后转为已审核)
		res := tx.Model(&model.User{}).
			Where("id = ? AND balance >= ?", userID, amount).
			UpdateColumn("balance", gorm.Expr("balance - ?", amount))
		if res.Error != nil {
			return res.Error
		}
		if res.RowsAffected == 0 {
			return errors.New("余额不足,提现失败")
		}
		// 创建提现记录
		return tx.Create(w).Error
	})
	if err != nil {
		return nil, err
	}
	return w, nil
}

// GetMyServices 打手服务项目列表(通过 club_members 关联)
func GetMyServices(playerID int64) ([]map[string]interface{}, error) {
	// 简化:返回打手所属俱乐部信息
	var members []model.ClubMember
	if err := db.Where("user_id = ? AND status = 1 AND role = ?", playerID, model.ClubMemberRolePlayer).Find(&members).Error; err != nil {
		return nil, err
	}
	result := make([]map[string]interface{}, 0, len(members))
	for _, m := range members {
		c, _ := clubRepo.FindByID(m.ClubID)
		result = append(result, map[string]interface{}{
			"club_id":   m.ClubID,
			"club_name": cName(c),
			"joined_at": m.JoinedAt,
		})
	}
	return result, nil
}

func cName(c *model.Club) string {
	if c == nil {
		return ""
	}
	return c.Name
}

// CreateService 打手创建服务项目(占位:写入 club_member 标记)
func CreateService(playerID, clubID int64, name string, amount int64) error {
	if name == "" || amount <= 0 {
		return errors.New("服务名称与金额不能为空")
	}
	// 实际项目应写入 services 表;此处仅校验打手身份
	m, err := clubRepo.FindMember(clubID, playerID)
	if err != nil {
		return err
	}
	if m == nil {
		return errors.New("您不是该俱乐部打手")
	}
	return nil
}

// UpdateService 更新服务项目(占位)
func UpdateService(serviceID, playerID int64, fields map[string]interface{}) error {
	_ = playerID
	_ = fields
	_ = serviceID
	return nil
}

// GetMyEvaluations 打手收到的评价列表
func GetMyEvaluations(playerID int64, page, pageSize int) ([]model.Evaluation, int64, error) {
	var list []model.Evaluation
	var total int64
	q := db.Model(&model.Evaluation{}).Where("player_id = ?", playerID)
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	err := q.Scopes(Paginate2(page, pageSize)).Order("id DESC").Find(&list).Error
	return list, total, err
}

// AppealEvaluation 打手对评价申诉
func AppealEvaluation(evaluationID, playerID int64, reason string) (*model.Appeal, error) {
	var e model.Evaluation
	if err := db.First(&e, evaluationID).Error; err != nil {
		return nil, errors.New("评价不存在")
	}
	if e.PlayerID != playerID {
		return nil, errors.New("无权对该评价申诉")
	}
	a := &model.Appeal{
		OrderID:      e.OrderID,
		UserID:       playerID,
		Type:         "evaluation",
		Description:  reason,
		Status:       model.AppealStatusPending,
		CreatedAt:    nowTimePtr(),
		UpdatedAt:    nowTimePtr(),
	}
	if err := db.Create(a).Error; err != nil {
		return nil, err
	}
	return a, nil
}

// GetPlayerDetail 打手详情(含评价统计)
func GetPlayerDetail(playerID int64) (map[string]interface{}, error) {
	u, err := userRepo.FindByID(playerID)
	if err != nil {
		return nil, err
	}
	if u == nil {
		return nil, errors.New("打手不存在")
	}
	// 评分与订单数统计
	var avgScore float64
	var orderCnt int64
	_ = db.Model(&model.Evaluation{}).Where("player_id = ?", playerID).
		Select("COALESCE(AVG(score),5)").Scan(&avgScore).Error
	_ = db.Model(&model.Order{}).Where("player_id = ? AND status IN ?", playerID, []int8{
		model.OrderStatusCompleted, model.OrderStatusSettled,
	}).Count(&orderCnt).Error
	return map[string]interface{}{
		"player":      u,
		"avg_score":   avgScore,
		"order_count": orderCnt,
	}, nil
}

// GetPlayerList 打手列表(分页,可按俱乐部过滤)
func GetPlayerList(clubID int64, page, pageSize int) ([]model.User, int64, error) {
	return clubRepo.ListPlayers(clubID, page, pageSize)
}

// SubmitJoinApplication 提交入会申请
func SubmitJoinApplication(userID, clubID int64, realName, gameAccount, gameRegion, goodPosition, rankLevel, intro string) (*model.JoinApplication, error) {
	if realName == "" || gameAccount == "" {
		return nil, errors.New("真实姓名与游戏账号不能为空")
	}
	a := &model.JoinApplication{
		ClubID:       clubID,
		UserID:       userID,
		RealName:     realName,
		GameAccount:  gameAccount,
		GameRegion:   gameRegion,
		GoodPosition: goodPosition,
		RankLevel:    rankLevel,
		Intro:        intro,
		Status:       model.JoinStatusPending,
		CreatedAt:    nowTimePtr(),
		UpdatedAt:    nowTimePtr(),
	}
	if err := clubRepo.CreateApplication(a); err != nil {
		return nil, err
	}
	return a, nil
}

// GetMyApplications 用户提交的入会申请列表
func GetMyApplications(userID int64, page, pageSize int) ([]model.JoinApplication, int64, error) {
	return clubRepo.ListUserApplications(userID, page, pageSize)
}

// freezePlayerAccountTx 事务冻结打手账户(供风控调用)
func freezePlayerAccountTx(tx *gorm.DB, playerID int64) error {
	return tx.Model(&model.User{}).Where("id = ?", playerID).
		UpdateColumn("status", 0).Error
}
