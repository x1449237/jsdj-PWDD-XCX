package service

import (
	"errors"

	"gorm.io/gorm"

	"github.com/jisan/e-sports-platform/internal/model"
)

// GetClubList 俱乐部列表(已审核通过)
func GetClubList(page, pageSize int, keyword string) ([]model.Club, int64, error) {
	return clubRepo.List(page, pageSize, model.ClubStatusApproved, keyword)
}

// GetClubDetail 俱乐部详情
func GetClubDetail(clubID int64) (*model.Club, error) {
	c, err := clubRepo.FindByID(clubID)
	if err != nil {
		return nil, err
	}
	if c == nil {
		return nil, errors.New("俱乐部不存在")
	}
	return c, nil
}

// ShopGetClubInfo 内置管理端获取俱乐部信息
func ShopGetClubInfo(clubID int64) (*model.Club, error) {
	return clubRepo.FindByID(clubID)
}

// ShopUpdateClubInfo 内置管理端更新俱乐部信息
func ShopUpdateClubInfo(clubID int64, fields map[string]interface{}) error {
	fields["updated_at"] = nowTimePtr()
	return clubRepo.Update(clubID, fields)
}

// ShopGetApplications 俱乐部入会申请列表
func ShopGetApplications(clubID int64, page, pageSize int, status string) ([]model.JoinApplication, int64, error) {
	return clubRepo.ListApplications(clubID, page, pageSize, status)
}

// ShopGetApplicationDetail 入会申请详情(校验俱乐部归属)
func ShopGetApplicationDetail(applicationID, clubID int64) (*model.JoinApplication, error) {
	a, err := clubRepo.FindApplication(applicationID)
	if err != nil {
		return nil, err
	}
	if a == nil {
		return nil, errors.New("申请不存在")
	}
	if a.ClubID != clubID {
		return nil, errors.New("无权查看该申请")
	}
	return a, nil
}

// ShopStartExam 开始考核(申请状态 -> examining)
func ShopStartExam(applicationID, examinerID int64, requirement string) error {
	a, err := clubRepo.FindApplication(applicationID)
	if err != nil {
		return err
	}
	if a == nil {
		return errors.New("申请不存在")
	}
	if a.Status != model.JoinStatusPending {
		return errors.New("申请状态不允许开始考核")
	}
	if err := clubRepo.UpdateApplication(applicationID, map[string]interface{}{
		"status":     model.JoinStatusExamining,
		"updated_at": nowTimePtr(),
	}); err != nil {
		return err
	}
	return clubRepo.CreateExamRecord(&model.ExamRecord{
		ApplicationID: applicationID,
		ExaminerID:    examinerID,
		Requirement:   requirement,
		CreatedAt:     nowTimePtr(),
		UpdatedAt:     nowTimePtr(),
	})
}

// ShopSubmitExamResult 提交考核结果
func ShopSubmitExamResult(applicationID, examinerID int64, result, remark, videoURL string) error {
	a, err := clubRepo.FindApplication(applicationID)
	if err != nil {
		return err
	}
	if a == nil {
		return errors.New("申请不存在")
	}
	if a.Status != model.JoinStatusExamining {
		return errors.New("申请未在考核中")
	}
	if err := clubRepo.CreateExamRecord(&model.ExamRecord{
		ApplicationID: applicationID,
		ExaminerID:    examinerID,
		Result:        result,
		Remark:        remark,
		VideoURL:      videoURL,
		CreatedAt:     nowTimePtr(),
		UpdatedAt:     nowTimePtr(),
	}); err != nil {
		return err
	}
	// 考核通过则申请状态保持 examining，等待审批
	return nil
}

// ShopApproveApplication 通过入会申请(加入俱乐部为打手)
func ShopApproveApplication(applicationID int64) error {
	a, err := clubRepo.FindApplication(applicationID)
	if err != nil {
		return err
	}
	if a == nil {
		return errors.New("申请不存在")
	}
	if a.Status == model.JoinStatusApproved {
		return nil
	}
	return db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Model(&model.JoinApplication{}).Where("id = ?", applicationID).
			Updates(map[string]interface{}{"status": model.JoinStatusApproved, "updated_at": nowTimePtr()}).Error; err != nil {
			return err
		}
		// 加入俱乐部成员表
		m := &model.ClubMember{
			ClubID:   a.ClubID,
			UserID:   a.UserID,
			Role:     model.ClubMemberRolePlayer,
			JoinedAt: nowTimePtr(),
			Status:   1,
			CreatedAt: nowTimePtr(),
			UpdatedAt: nowTimePtr(),
		}
		if err := tx.Create(m).Error; err != nil {
			return err
		}
		// 用户角色增加打手位，并绑定俱乐部
		return tx.Model(&model.User{}).Where("id = ?", a.UserID).
			Updates(map[string]interface{}{
				"role":       gorm.Expr("role | ?", model.RolePlayer),
				"club_id":    a.ClubID,
				"updated_at": nowTimePtr(),
			}).Error
	})
}

// ShopRejectApplication 驳回入会申请
func ShopRejectApplication(applicationID int64, reason string) error {
	return clubRepo.UpdateApplication(applicationID, map[string]interface{}{
		"status":     model.JoinStatusRejected,
		"updated_at": nowTimePtr(),
	})
}

// ShopGetGamers 俱乐部打手列表
func ShopGetGamers(clubID int64, page, pageSize int) ([]model.User, int64, error) {
	return clubRepo.ListPlayers(clubID, page, pageSize)
}

// ShopGetGamerDetail 打手详情
func ShopGetGamerDetail(clubID, gamerID int64) (map[string]interface{}, error) {
	m, err := clubRepo.FindMember(clubID, gamerID)
	if err != nil {
		return nil, err
	}
	if m == nil {
		return nil, errors.New("该用户非本俱乐部成员")
	}
	u, _ := userRepo.FindByID(gamerID)
	var evals []model.Evaluation
	_ = db.Where("player_id = ?", gamerID).Order("id DESC").Limit(20).Find(&evals).Error
	return map[string]interface{}{
		"member":       m,
		"user":         u,
		"evaluations":  evals,
	}, nil
}

// ShopApproveGamer 审核打手通过(占位:成员已存在则置为有效)
func ShopApproveGamer(clubID, gamerID int64) error {
	m, err := clubRepo.FindMember(clubID, gamerID)
	if err != nil {
		return err
	}
	if m == nil {
		return errors.New("该用户非本俱乐部成员")
	}
	return clubRepo.UpdateMember(m.ID, map[string]interface{}{"status": 1, "updated_at": nowTimePtr()})
}

// ShopRemoveGamer 移除打手
func ShopRemoveGamer(clubID, gamerID int64) error {
	m, err := clubRepo.FindMember(clubID, gamerID)
	if err != nil {
		return err
	}
	if m == nil {
		return errors.New("该用户非本俱乐部成员")
	}
	if m.Role == model.ClubMemberRoleFounder {
		return errors.New("不可移除创始人")
	}
	return clubRepo.UpdateMember(m.ID, map[string]interface{}{"status": 0, "updated_at": nowTimePtr()})
}

// ShopGetGamerEvaluations 打手评价列表
func ShopGetGamerEvaluations(gamerID int64, page, pageSize int) ([]model.Evaluation, int64, error) {
	var list []model.Evaluation
	var total int64
	q := db.Model(&model.Evaluation{}).Where("player_id = ?", gamerID)
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	err := q.Scopes(Paginate2(page, pageSize)).Order("id DESC").Find(&list).Error
	return list, total, err
}

// ShopGetAdmins 俱乐部管理员列表
func ShopGetAdmins(clubID int64) ([]model.ShopAdminAccount, error) {
	return clubRepo.ListShopAdmins(clubID)
}

// ShopAddAdmin 添加内置管理端账号(创始人专属)
func ShopAddAdmin(clubID int64, username, password, realName, phone string, role int8) (*model.ShopAdminAccount, error) {
	// 校验账号唯一
	exist, _ := clubRepo.FindShopAdminByUsername(username)
	if exist != nil {
		return nil, errors.New("账号已存在")
	}
	hash, err := hashPwd(password)
	if err != nil {
		return nil, err
	}
	a := &model.ShopAdminAccount{
		Username:  username,
		Password:  hash,
		ClubID:    clubID,
		Role:      role,
		RealName:  realName,
		Phone:     phone,
		Status:    1,
		CreatedAt: nowTimePtr(),
		UpdatedAt: nowTimePtr(),
	}
	if a.Role == 0 {
		a.Role = model.ShopAdminRoleAdmin
	}
	if err := clubRepo.CreateShopAdmin(a); err != nil {
		return nil, err
	}
	a.Password = ""
	return a, nil
}

// ShopRemoveAdmin 移除管理员(不可移除创始人)
func ShopRemoveAdmin(clubID, adminID int64) error {
	a, err := clubRepo.FindShopAdminByID(adminID)
	if err != nil {
		return err
	}
	if a == nil {
		return errors.New("管理员不存在")
	}
	if a.ClubID != clubID {
		return errors.New("无权操作")
	}
	if a.Role == model.ShopAdminRoleFounder {
		return errors.New("不可移除创始人")
	}
	return clubRepo.DeleteShopAdmin(adminID)
}

// ShopResetAdminPassword 重置管理员密码
func ShopResetAdminPassword(clubID, adminID int64, newPassword string) error {
	a, err := clubRepo.FindShopAdminByID(adminID)
	if err != nil {
		return err
	}
	if a == nil {
		return errors.New("管理员不存在")
	}
	if a.ClubID != clubID {
		return errors.New("无权操作")
	}
	hash, err := hashPwd(newPassword)
	if err != nil {
		return err
	}
	return clubRepo.UpdateShopAdmin(adminID, map[string]interface{}{
		"password":   hash,
		"updated_at": nowTimePtr(),
	})
}

// ShopGetGroups 俱乐部群聊列表
func ShopGetGroups(clubID int64) ([]model.GroupChat, error) {
	return chatRepo.ListGroups(clubID)
}

// ShopCreateGroup 创建群聊
func ShopCreateGroup(clubID, creatorID int64, groupName, groupType, categoryType string) (*model.GroupChat, error) {
	g := &model.GroupChat{
		GroupName:    groupName,
		GroupType:    groupType,
		ClubID:       clubID,
		CategoryType: categoryType,
		CreatorID:    creatorID,
		Status:       1,
		CreatedAt:    nowTimePtr(),
		UpdatedAt:    nowTimePtr(),
	}
	if g.GroupType == "" {
		g.GroupType = model.GroupTypeInternal
	}
	if err := chatRepo.CreateGroupChat(g); err != nil {
		return nil, err
	}
	return g, nil
}

// ShopGetGroupMembers 群成员列表
func ShopGetGroupMembers(groupID int64) ([]model.GroupChatMember, error) {
	return chatRepo.ListGroupMembers(groupID)
}

// ShopSendGroupMessage 群发消息
func ShopSendGroupMessage(groupID, senderID int64, msgType, content, mediaURL string) (*model.GroupChatMessage, error) {
	m := &model.GroupChatMessage{
		GroupID:  groupID,
		SenderID: senderID,
		MsgType:  msgType,
		Content:  content,
		MediaURL: mediaURL,
		CreatedAt: nowTimePtr(),
	}
	if m.MsgType == "" {
		m.MsgType = model.MsgTypeText
	}
	if err := chatRepo.CreateGroupMessage(m); err != nil {
		return nil, err
	}
	return m, nil
}

// ShopPublishAnnouncement 发布群公告
func ShopPublishAnnouncement(groupID int64, announcement string) error {
	// 发布公告前防代练检测
	if announcement != "" {
		hit, _, abErr := CheckContentAntiBoosting(AntiBoostingContentTypeAnnouncement, 0, announcement)
		if abErr == nil && hit {
			return errors.New("公告内容包含违规关键词，发布失败")
		}
	}
	return chatRepo.UpdateGroupAnnouncement(groupID, announcement)
}

// ShopGetRiskUsers 俱乐部风控用户(简化:本俱乐部低信用分用户)
func ShopGetRiskUsers(clubID int64, page, pageSize int) ([]model.User, int64, error) {
	var list []model.User
	var total int64
	q := db.Model(&model.User{}).
		Where("club_id = ? AND (credit_score < 60 OR status = 0)", clubID)
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	err := q.Scopes(Paginate2(page, pageSize)).Order("id DESC").Find(&list).Error
	return list, total, err
}

// ShopGetRiskOrders 俱乐部风险订单(异常状态)
func ShopGetRiskOrders(clubID int64, page, pageSize int) ([]model.Order, int64, error) {
	var list []model.Order
	var total int64
	q := db.Model(&model.Order{}).Where("club_id = ? AND status IN ?", clubID, []int8{
		model.OrderStatusVerifyFail, model.OrderStatusTimeout,
	})
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	err := q.Scopes(Paginate2(page, pageSize)).Order("id DESC").Find(&list).Error
	return list, total, err
}

// AdminAuditClubs 平台俱乐部审核列表
func AdminAuditClubs(page, pageSize int, status int8, keyword string) ([]model.Club, int64, error) {
	return clubRepo.List(page, pageSize, status, keyword)
}

// AdminApproveClub 审核通过俱乐部 + 点亮 V 标
func AdminApproveClub(clubID, adminID int64) error {
	if err := clubRepo.Update(clubID, map[string]interface{}{
		"status":     model.ClubStatusApproved,
		"updated_at": nowTimePtr(),
	}); err != nil {
		return err
	}
	// 审核通过后点亮 V 标
	c, _ := clubRepo.FindByID(clubID)
	if c != nil {
		_ = GrantClubVBadge(clubID, int64(c.Type))
	}
	_ = adminID
	return nil
}

// AdminRejectClub 驳回俱乐部 + 撤销 V 标
func AdminRejectClub(clubID int64, reason string) error {
	if err := clubRepo.Update(clubID, map[string]interface{}{
		"status":     model.ClubStatusRejected,
		"updated_at": nowTimePtr(),
	}); err != nil {
		return err
	}
	_ = RevokeClubVBadge(clubID)
	_ = reason
	return nil
}

// AdminFreezeClub 冻结俱乐部 + 撤销 V 标
func AdminFreezeClub(clubID int64, reason string) error {
	if err := clubRepo.Update(clubID, map[string]interface{}{
		"status":     model.ClubStatusFrozen,
		"updated_at": nowTimePtr(),
	}); err != nil {
		return err
	}
	_ = RevokeClubVBadge(clubID)
	_ = reason
	return nil
}

// AdminUnfreezeClub 解冻俱乐部 + 恢复 V 标
func AdminUnfreezeClub(clubID int64) error {
	if err := clubRepo.Update(clubID, map[string]interface{}{
		"status":     model.ClubStatusApproved,
		"updated_at": nowTimePtr(),
	}); err != nil {
		return err
	}
	c, _ := clubRepo.FindByID(clubID)
	if c != nil {
		_ = GrantClubVBadge(clubID, int64(c.Type))
	}
	return nil
}

// AdminCancelClub 注销俱乐部 + 撤销 V 标
func AdminCancelClub(clubID int64, reason string) error {
	if err := clubRepo.Update(clubID, map[string]interface{}{
		"status":     model.ClubStatusCanceled,
		"updated_at": nowTimePtr(),
	}); err != nil {
		return err
	}
	_ = RevokeClubVBadge(clubID)
	_ = reason
	return nil
}

// hashPwd 哈希密码(封装 utils.HashPassword)
func hashPwd(pwd string) (string, error) {
	return hashPasswordUtil(pwd)
}

// RecordClubJoinClick 记录俱乐部入驻点击(跳转客服微信)
func RecordClubJoinClick(userID int64) (string, error) {
	// 返回客服微信(从系统配置读取)
	val := getSystemConfig("customer_wechat")
	if val == "" {
		val = "jisan-service"
	}
	return val, nil
}

// getSystemConfig 读取系统配置项(简化:实时查表)
func getSystemConfig(key string) string {
	var sc model.SystemConfig
	if err := db.Where("`key` = ?", key).First(&sc).Error; err != nil {
		return ""
	}
	return sc.Value
}
