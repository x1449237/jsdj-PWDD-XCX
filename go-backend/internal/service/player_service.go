package service

import (
	"errors"
	"time"

	"gorm.io/gorm"

	"github.com/jisan/e-sports-platform/internal/model"
)

// PlayerServiceItem 打手服务项目(简化:复用 Order ServiceID 字段)
// 实际项目应有独立 services 表，此处用 ClubMember + 简化字段示意

// GetGrabOrderList 可抢单列表(打手只能抢本俱乐部或公开订单)
func GetGrabOrderList(playerID, clubID int64, page, pageSize int) ([]model.Order, int64, error) {
	return orderRepo.ListGrabOrders(clubID, page, pageSize)
}

// GrabOrder 打手抢单(原子操作 + 分布式锁防并发)
func GrabOrder(orderID, playerID int64) error {
	// 分布式锁防并发抢单
	lockKey := cacheKey("grab:" + itoa(orderID))
	ctx, cancel := contextWithTimeout()
	defer cancel()
	ok, err := redis.SetNX(ctx, lockKey, playerID, 5*time.Second)
	if err != nil {
		return err
	}
	if !ok {
		return errors.New("订单正在被抢，请稍后")
	}
	defer redis.Del(ctx, lockKey)

	grabbed, err := orderRepo.GrabOrder(orderID, playerID)
	if err != nil {
		return err
	}
	if !grabbed {
		return errors.New("抢单失败，订单已被他人接单或状态已变更")
	}
	o, _ := orderRepo.FindByID(orderID)
	if o != nil {
		_ = orderRepo.CreateStatusLog(&model.OrderStatusLog{
			OrderID: orderID, FromStatus: model.OrderStatusPending, ToStatus: model.OrderStatusAccepted,
			OperatorID: playerID, OperatorType: "player", Reason: "打手抢单",
			CreatedAt: nowTimePtr(),
		})
	}
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
	if toPlayer <= 0 {
		return errors.New("转单目标打手不能为空")
	}
	return orderRepo.Update(orderID, map[string]interface{}{
		"player_id":  toPlayer,
		"updated_at": nowTimePtr(),
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
func ApplyWithdraw(userID int64, amount int64, channel string, bankInfo map[string]string) (*model.Withdraw, error) {
	if amount <= 0 {
		return nil, errors.New("提现金额必须大于 0")
	}
	u, err := userRepo.FindByID(userID)
	if err != nil {
		return nil, err
	}
	if u == nil {
		return nil, errors.New("用户不存在")
	}
	// 校验可提现余额
	settled, _ := paymentRepo.SumSettledEarnings(userID)
	var withdrawn int64
	_ = db.Model(&model.Withdraw{}).Where("user_id = ? AND status IN ?", userID, []string{model.WithdrawStatusPending, model.WithdrawStatusApproved, model.WithdrawStatusPaid}).
		Select("COALESCE(SUM(amount),0)").Scan(&withdrawn).Error
	available := settled - withdrawn
	if amount > available {
		return nil, errors.New("可提现余额不足")
	}
	// 手续费与个税(简化:手续费 0.6%，个税 20%)
	fee := amount * 6 / 1000
	tax := (amount - fee) * 20 / 100
	net := amount - fee - tax
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
	if err := db.Create(w).Error; err != nil {
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
