package service

import (
	"encoding/json"
	"errors"
	"fmt"

	"gorm.io/gorm"

	"github.com/jisan/e-sports-platform/internal/model"
	"github.com/jisan/e-sports-platform/internal/utils"
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
// - 如修改俱乐部名称，触发缩写重新生成+查重，新缩写冲突则名称修改失败
// - 所有修改记录写入 club_info_change_logs
func ShopUpdateClubInfo(clubID int64, fields map[string]interface{}) error {
	c, err := clubRepo.FindByID(clubID)
	if err != nil {
		return err
	}
	if c == nil {
		return errors.New("俱乐部不存在")
	}
	operatorID := int64(0)
	if op, ok := fields["__operator_id"]; ok {
		if v, ok2 := op.(int64); ok2 {
			operatorID = v
		}
		delete(fields, "__operator_id")
	}

	now := nowTimePtr()
	changeLogs := make([]*model.ClubInfoChangeLog, 0, len(fields))

	// 安全防护:禁止外部直接设置 abbreviation(只能由 name 变更时由系统重新生成)
	// 防止攻击者通过 updateClubInfoRequest 直接覆盖缩写绕过查重
	delete(fields, "abbreviation")

	// 同样禁止外部直接修改 status/founder_uid/id 等敏感字段
	delete(fields, "status")
	delete(fields, "founder_uid")
	delete(fields, "id")
	delete(fields, "type")
	delete(fields, "reject_count")
	delete(fields, "locked_until")
	delete(fields, "is_archived")
	delete(fields, "v_badge_type")
	delete(fields, "v_badge_hidden")
	delete(fields, "deposit_amount")
	delete(fields, "deposit_paid")
	delete(fields, "created_at")

	// 名称变更:触发缩写重新生成 + 查重
	if newName, ok := fields["name"].(string); ok && newName != "" && newName != c.Name {
		newAbbr := utils.GenerateAbbreviation(newName)
		if newAbbr == "" {
			return errors.New("俱乐部名称无法生成有效缩写")
		}
		// 查重:排除当前俱乐部自身
		var cnt int64
		if err := db.Model(&model.Club{}).
			Where("abbreviation = ? AND id <> ?", newAbbr, clubID).
			Count(&cnt).Error; err != nil {
			return err
		}
		if cnt > 0 {
			return errors.New("新名称生成的缩写与已有俱乐部冲突，名称修改失败")
		}
		// 同步更新缩写
		fields["abbreviation"] = newAbbr
		changeLogs = append(changeLogs, &model.ClubInfoChangeLog{
			ClubID: clubID, Field: "name", OldValue: c.Name, NewValue: newName,
			OperatorID: operatorID, CreatedAt: now,
		})
		changeLogs = append(changeLogs, &model.ClubInfoChangeLog{
			ClubID: clubID, Field: "abbreviation", OldValue: c.Abbreviation, NewValue: newAbbr,
			OperatorID: operatorID, CreatedAt: now,
		})
	}

	// 其余字段变更记录
	fieldValueMap := map[string]string{
		"logo":         c.Logo,
		"intro":        c.Description,
		"description":  c.Description,
		"background":   c.Background,
		"contact_phone":   c.ContactPhone,
		"contact_wechat":  c.ContactWechat,
		"contact_qq":      c.ContactQQ,
		"business_hours":  c.BusinessHours,
	}
	for field, newVal := range fields {
		if field == "name" || field == "abbreviation" || field == "updated_at" {
			continue
		}
		oldVal, tracked := fieldValueMap[field]
		if !tracked {
			continue
		}
		newValStr, _ := newVal.(string)
		if newValStr == oldVal {
			continue
		}
		changeLogs = append(changeLogs, &model.ClubInfoChangeLog{
			ClubID: clubID, Field: field, OldValue: oldVal, NewValue: newValStr,
			OperatorID: operatorID, CreatedAt: now,
		})
	}

	fields["updated_at"] = now
	return db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Model(&model.Club{}).Where("id = ?", clubID).Updates(fields).Error; err != nil {
			return err
		}
		for _, l := range changeLogs {
			if err := tx.Create(l).Error; err != nil {
				return err
			}
		}
		return nil
	})
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
// 安全修复:新增 clubID 参数,校验申请归属(防跨俱乐部操作他人申请)
func ShopStartExam(clubID, applicationID, examinerID int64, requirement string) error {
	a, err := clubRepo.FindApplication(applicationID)
	if err != nil {
		return err
	}
	if a == nil {
		return errors.New("申请不存在")
	}
	if a.ClubID != clubID {
		return errors.New("无权操作该俱乐部的申请")
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
// 安全修复:新增 clubID 参数,校验申请归属(防跨俱乐部操作他人申请)
func ShopSubmitExamResult(clubID, applicationID, examinerID int64, result, remark, videoURL string) error {
	a, err := clubRepo.FindApplication(applicationID)
	if err != nil {
		return err
	}
	if a == nil {
		return errors.New("申请不存在")
	}
	if a.ClubID != clubID {
		return errors.New("无权操作该俱乐部的申请")
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
// 安全修复:新增 clubID 参数,校验申请归属(防跨俱乐部审批他人申请)
func ShopApproveApplication(clubID, applicationID int64) error {
	a, err := clubRepo.FindApplication(applicationID)
	if err != nil {
		return err
	}
	if a == nil {
		return errors.New("申请不存在")
	}
	if a.ClubID != clubID {
		return errors.New("无权操作该俱乐部的申请")
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
// 安全修复:新增 clubID 参数,校验申请归属(原直接更新无任何校验,任意俱乐部可驳回他人申请)
func ShopRejectApplication(clubID, applicationID int64, reason string) error {
	a, err := clubRepo.FindApplication(applicationID)
	if err != nil {
		return err
	}
	if a == nil {
		return errors.New("申请不存在")
	}
	if a.ClubID != clubID {
		return errors.New("无权操作该俱乐部的申请")
	}
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
// - 成员置为已移除(status=0)
// - 自动退出俱乐部所有群聊(group_chat_members)
// - 断开售后 IM:关闭该打手作为处理方的进行中售后会话
// - 清除用户 club_id 绑定并移除打手角色位
// - 推送 WebSocket 通知给被移除用户
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
	now := nowTimePtr()
	err = db.Transaction(func(tx *gorm.DB) error {
		// 1. 成员置为已移除
		if err := tx.Model(&model.ClubMember{}).Where("id = ?", m.ID).
			Updates(map[string]interface{}{"status": 0, "updated_at": now}).Error; err != nil {
			return err
		}
		// 2. 退出俱乐部所有群聊
		if err := tx.Where("user_id = ? AND group_id IN (?)",
			gamerID, tx.Model(&model.GroupChat{}).Select("id").Where("club_id = ?", clubID)).
			Delete(&model.GroupChatMember{}).Error; err != nil {
			return err
		}
		// 3. 断开售后 IM:关闭该打手关联的进行中售后会话
		//    售后会话通过订单关联打手,关闭 status=1 -> 2
		if err := tx.Model(&model.AfterSaleSession{}).
			Where("club_id = ? AND status = ? AND order_id IN (?)",
				clubID, 1,
				tx.Model(&model.Order{}).Select("id").Where("club_id = ? AND player_id = ?", clubID, gamerID)).
			Updates(map[string]interface{}{"status": 2, "updated_at": now}).Error; err != nil {
			return err
		}
		// 4. 清除用户 club_id 绑定并移除打手角色位
		return tx.Model(&model.User{}).Where("id = ?", gamerID).
			Updates(map[string]interface{}{
				"club_id":    0,
				"role":        gorm.Expr("role & ~?", model.RolePlayer),
				"updated_at": now,
			}).Error
	})
	if err != nil {
		return err
	}
	// 5. 推送通知给被移除用户(站内消息 + WebSocket)
	_ = AdminSendNotification(gamerID, "club", "您已被移出俱乐部",
		"您已被移出该俱乐部,相关群聊与售后会话已自动断开。", model.NotificationCategorySystem)
	return nil
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
// 安全修复:
// 1. 新增 operatorID 参数,校验操作者必须是创始人(原任意管理员可添加,导致权限提升)
// 2. 禁止创建创始人角色账号(防普通管理员通过创建创始人账号夺取俱乐部控制权)
func ShopAddAdmin(clubID, operatorID int64, username, password, realName, phone string, role int8) (*model.ShopAdminAccount, error) {
	// 校验操作者是创始人
	if err := assertShopFounder(clubID, operatorID); err != nil {
		return nil, err
	}
	// 禁止创建创始人角色(防权限提升,创始人只能由系统在俱乐部入驻时创建)
	if role == model.ShopAdminRoleFounder {
		return nil, errors.New("不可创建创始人角色账号")
	}
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

// AdminAddShopAccountByPlatform 平台超管代建内置管理端账号
// 与 ShopAddAdmin 的区别:不校验操作者是否创始人(平台超管权限更高),
// 但仍禁止创建创始人角色(防权限提升)
func AdminAddShopAccountByPlatform(clubID int64, username, password, realName, phone string, role int8) (*model.ShopAdminAccount, error) {
	// 禁止创建创始人角色(创始人只能由系统在俱乐部入驻时创建)
	if role == model.ShopAdminRoleFounder {
		return nil, errors.New("不可创建创始人角色账号")
	}
	// 校验账号唯一
	exist, _ := clubRepo.FindShopAdminByUsername(username)
	if exist != nil {
		return nil, errors.New("账号已存在")
	}
	// 校验俱乐部存在
	c, _ := clubRepo.FindByID(clubID)
	if c == nil {
		return nil, errors.New("俱乐部不存在")
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

// assertShopFounder 校验操作者是该俱乐部的创始人(且账号有效)
func assertShopFounder(clubID, operatorID int64) error {
	if operatorID <= 0 {
		return errors.New("操作者身份无效")
	}
	op, err := clubRepo.FindShopAdminByID(operatorID)
	if err != nil || op == nil {
		return errors.New("操作者账号不存在")
	}
	if op.ClubID != clubID {
		return errors.New("操作者不属于该俱乐部")
	}
	if op.Status != 1 {
		return errors.New("操作者账号已被禁用")
	}
	if op.Role != model.ShopAdminRoleFounder {
		return errors.New("仅创始人可执行该操作")
	}
	return nil
}

// ShopRemoveAdmin 移除管理员(不可移除创始人,仅创始人可操作)
// 安全修复:新增 operatorID 参数,校验操作者是创始人(原任意管理员可移除他人,导致权限滥用)
func ShopRemoveAdmin(clubID, operatorID, adminID int64) error {
	if err := assertShopFounder(clubID, operatorID); err != nil {
		return err
	}
	if operatorID == adminID {
		return errors.New("不可移除自己")
	}
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

// ShopResetAdminPassword 重置管理员密码(仅创始人可操作,不可重置创始人密码)
// 安全修复:
// 1. 新增 operatorID 参数,校验操作者是创始人
// 2. 禁止重置创始人密码(防普通管理员重置创始人密码夺取控制权)
func ShopResetAdminPassword(clubID, operatorID, adminID int64, newPassword string) error {
	if err := assertShopFounder(clubID, operatorID); err != nil {
		return err
	}
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
	if a.Role == model.ShopAdminRoleFounder && operatorID != adminID {
		return errors.New("不可重置其他创始人的密码")
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

// maxClubGroups 俱乐部群聊数量上限(防滥用)
const maxClubGroups = 10

// ShopCreateGroup 创建群聊
// - 限制单俱乐部群聊数量上限(maxClubGroups)
func ShopCreateGroup(clubID, creatorID int64, groupName, groupType, categoryType string) (*model.GroupChat, error) {
	// 群聊数量上限校验
	var cnt int64
	if err := db.Model(&model.GroupChat{}).Where("club_id = ? AND status = 1", clubID).Count(&cnt).Error; err != nil {
		return nil, err
	}
	if cnt >= maxClubGroups {
		return nil, errors.New("群聊数量已达上限，无法继续创建")
	}
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
// 安全修复:新增 clubID 参数,校验群归属(防跨俱乐部查看他人群成员)
func ShopGetGroupMembers(clubID, groupID int64) ([]model.GroupChatMember, error) {
	g, err := chatRepo.FindGroup(groupID)
	if err != nil {
		return nil, err
	}
	if g == nil {
		return nil, errors.New("群聊不存在")
	}
	if g.ClubID != clubID {
		return nil, errors.New("无权查看该俱乐部的群成员")
	}
	return chatRepo.ListGroupMembers(groupID)
}

// ShopSendGroupMessage 群发消息
// - 群聊内容防代练 + 敏感词风控扫描
// 安全修复:新增 clubID 参数,校验群归属(防跨俱乐部向他人群注入消息)
func ShopSendGroupMessage(clubID, groupID, senderID int64, msgType, content, mediaURL string) (*model.GroupChatMessage, error) {
	g, err := chatRepo.FindGroup(groupID)
	if err != nil {
		return nil, err
	}
	if g == nil {
		return nil, errors.New("群聊不存在")
	}
	if g.ClubID != clubID {
		return nil, errors.New("无权向该俱乐部的群聊发送消息")
	}
	if content != "" {
		// 防代练检测
		hit, _, abErr := CheckContentAntiBoosting(AntiBoostingContentTypeAnnouncement, senderID, content)
		if abErr == nil && hit {
			return nil, errors.New("群消息包含违规关键词，发送失败")
		}
	}
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
// 安全修复:新增 clubID 参数,校验群归属(防跨俱乐部篡改他人群公告)
func ShopPublishAnnouncement(clubID, groupID int64, announcement string) error {
	g, err := chatRepo.FindGroup(groupID)
	if err != nil {
		return err
	}
	if g == nil {
		return errors.New("群聊不存在")
	}
	if g.ClubID != clubID {
		return errors.New("无权修改该俱乐部的群公告")
	}
	// 发布公告前防代练检测
	if announcement != "" {
		hit, _, abErr := CheckContentAntiBoosting(AntiBoostingContentTypeAnnouncement, 0, announcement)
		if abErr == nil && hit {
			return errors.New("公告内容包含违规关键词，发布失败")
		}
	}
	return chatRepo.UpdateGroupAnnouncement(groupID, announcement)
}

// ShopMarkAnnouncementRead 标记公告已读(写入 announcement_read_logs,幂等)
func ShopMarkAnnouncementRead(groupID, userID int64) error {
	var cnt int64
	if err := db.Model(&model.AnnouncementReadLog{}).
		Where("announcement_id = ? AND user_id = ?", groupID, userID).Count(&cnt).Error; err != nil {
		return err
	}
	if cnt > 0 {
		return nil // 已记录,幂等
	}
	now := nowTimePtr()
	return db.Create(&model.AnnouncementReadLog{
		AnnouncementID: groupID, UserID: userID,
		ReadAt: now, CreatedAt: now,
	}).Error
}

// ShopGetAnnouncementReadStats 公告已读统计(已读数 / 群成员总数)
func ShopGetAnnouncementReadStats(groupID int64) (map[string]interface{}, error) {
	var members []model.GroupChatMember
	if err := db.Where("group_id = ?", groupID).Find(&members).Error; err != nil {
		return nil, err
	}
	total := int64(len(members))
	if total == 0 {
		return map[string]interface{}{"read_count": int64(0), "total_members": int64(0), "unread_count": int64(0)}, nil
	}
	memberIDs := make([]int64, 0, len(members))
	for _, m := range members {
		memberIDs = append(memberIDs, m.UserID)
	}
	var readCnt int64
	if err := db.Model(&model.AnnouncementReadLog{}).
		Where("announcement_id = ? AND user_id IN ?", groupID, memberIDs).Count(&readCnt).Error; err != nil {
		return nil, err
	}
	return map[string]interface{}{
		"read_count":    readCnt,
		"total_members": total,
		"unread_count":  total - readCnt,
	}, nil
}

// ShopUpdateCommissionRate 更新创始人抽成比例(0-100)
func ShopUpdateCommissionRate(clubID int64, rate int8) error {
	if rate < 0 || rate > 100 {
		return errors.New("抽成比例必须在 0-100 之间")
	}
	return clubRepo.Update(clubID, map[string]interface{}{
		"commission_rate": rate,
		"updated_at":      nowTimePtr(),
	})
}

// ShopCreateFineRule 创建俱乐部内部罚款规则(需平台备案审核)
// 规则创建后状态为 active,同时写入备案审核记录(待平台审核)
func ShopCreateFineRule(clubID, createdBy int64, name, description string, amount int64) (*model.ClubFineRule, error) {
	if name == "" {
		return nil, errors.New("规则名称不能为空")
	}
	if amount < 0 {
		return nil, errors.New("罚款金额不能为负")
	}
	now := nowTimePtr()
	rule := &model.ClubFineRule{
		ClubID:      clubID,
		Name:        name,
		Description: description,
		Amount:      amount,
		Status:      model.ClubFineRuleStatusActive,
		CreatedBy:   createdBy,
		CreatedAt:   now,
		UpdatedAt:   now,
	}
	err := db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(rule).Error; err != nil {
			return err
		}
		// 写入平台备案审核记录(待审核)
		return tx.Create(&model.ClubFineRuleReview{
			RuleID:       rule.ID,
			ClubID:       clubID,
			ReviewStatus: model.ClubFineRuleReviewPending,
			CreatedAt:    now,
		}).Error
	})
	if err != nil {
		return nil, err
	}
	return rule, nil
}

// ShopListFineRules 俱乐部罚款规则列表(支持状态过滤)
func ShopListFineRules(clubID int64, status string) ([]model.ClubFineRule, error) {
	var list []model.ClubFineRule
	q := db.Model(&model.ClubFineRule{}).Where("club_id = ?", clubID)
	if status != "" {
		q = q.Where("status = ?", status)
	}
	err := q.Order("id DESC").Find(&list).Error
	return list, err
}

// ShopRevokeFineRule 下架罚款规则(status=revoked)
func ShopRevokeFineRule(clubID, ruleID int64) error {
	res := db.Model(&model.ClubFineRule{}).
		Where("id = ? AND club_id = ?", ruleID, clubID).
		Updates(map[string]interface{}{"status": model.ClubFineRuleStatusRevoked, "updated_at": nowTimePtr()})
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return errors.New("罚款规则不存在或无权操作")
	}
	return nil
}

// AdminReviewFineRule 平台审核罚款规则备案(approved/revoked)
func AdminReviewFineRule(ruleID, reviewerID int64, approve bool, note string) error {
	status := model.ClubFineRuleReviewRevoked
	if approve {
		status = model.ClubFineRuleReviewApproved
	}
	now := nowTimePtr()
	return db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Model(&model.ClubFineRuleReview{}).
			Where("rule_id = ? AND review_status = ?", ruleID, model.ClubFineRuleReviewPending).
			Updates(map[string]interface{}{
				"review_status": status,
				"reviewer_id":   reviewerID,
				"review_note":   note,
				"reviewed_at":   now,
			}).Error; err != nil {
			return err
		}
		// 驳回(下架)则同步下架罚款规则
		if !approve {
			return tx.Model(&model.ClubFineRule{}).Where("id = ?", ruleID).
				Updates(map[string]interface{}{"status": model.ClubFineRuleStatusRevoked, "updated_at": now}).Error
		}
		return nil
	})
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

// ClubAuditFilter 俱乐部审核多条件筛选
type ClubAuditFilter struct {
	Status        int8  // -1 全部
	Type          int8  // -1 全部 1企业 2个人
	VBadgeType    int8  // -1 全部 0无 1蓝V 2绿V
	DepositStatus int8  // -1 全部
	Keyword       string
}

// AdminAuditClubsFiltered 平台俱乐部审核列表(多条件筛选)
// 支持按 状态/类型/V标/保证金状态/关键词 组合筛选
func AdminAuditClubsFiltered(page, pageSize int, f ClubAuditFilter) ([]model.Club, int64, error) {
	var list []model.Club
	var total int64
	q := db.Model(&model.Club{})
	if f.Status >= 0 {
		q = q.Where("status = ?", f.Status)
	}
	if f.Type > 0 {
		q = q.Where("type = ?", f.Type)
	}
	if f.VBadgeType >= 0 {
		q = q.Where("v_badge_type = ?", f.VBadgeType)
	}
	if f.DepositStatus >= 0 {
		q = q.Where("deposit_status = ?", f.DepositStatus)
	}
	if f.Keyword != "" {
		like := "%" + f.Keyword + "%"
		q = q.Where("name LIKE ? OR abbreviation LIKE ?", like, like)
	}
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	err := q.Scopes(Paginate2(page, pageSize)).Order("id DESC").Find(&list).Error
	return list, total, err
}

// AdminGetClubChangeLogs 俱乐部资料修改日志(入驻/资料变更审计溯源)
func AdminGetClubChangeLogs(clubID int64, page, pageSize int) ([]model.ClubInfoChangeLog, int64, error) {
	var list []model.ClubInfoChangeLog
	var total int64
	q := db.Model(&model.ClubInfoChangeLog{}).Where("club_id = ?", clubID)
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	err := q.Scopes(Paginate2(page, pageSize)).Order("id DESC").Find(&list).Error
	return list, total, err
}

// AdminApproveClub 审核通过俱乐部 + 点亮 V 标
// 安全修复:
// 1. 状态机校验:仅"审核中"可通过(原可对已驳回/冻结/注销的俱乐部执行通过)
// 2. 保证金校验:必须已缴纳保证金(deposit_status==1)才能审核通过上线
func AdminApproveClub(clubID, adminID int64) error {
	c, _ := clubRepo.FindByID(clubID)
	if c == nil {
		return errors.New("俱乐部不存在")
	}
	// 状态机校验:仅审核中可通过
	if c.Status != model.ClubStatusReviewing {
		return fmt.Errorf("俱乐部当前状态(%d)不允许审核通过,仅审核中可通过", c.Status)
	}
	// 保证金校验:必须已缴纳才能上线接单
	if c.DepositStatus != 1 {
		return errors.New("俱乐部尚未缴纳保证金,不可审核通过")
	}
	if err := clubRepo.Update(clubID, map[string]interface{}{
		"status":     model.ClubStatusApproved,
		"updated_at": nowTimePtr(),
	}); err != nil {
		return err
	}
	// 审核通过后点亮 V 标
	_ = GrantClubVBadge(clubID, int64(c.Type))
	_ = adminID
	return nil
}

// AdminRejectClub 驳回俱乐部 + 撤销 V 标
// 驳回次数 reject_count++，达到 3 次时设置 locked_until = NOW() + 7 天
// 安全修复:状态机校验,仅"审核中"可驳回(原可驳回已通过/已注销的俱乐部)
func AdminRejectClub(clubID int64, reason string) error {
	c, _ := clubRepo.FindByID(clubID)
	if c == nil {
		return errors.New("俱乐部不存在")
	}
	// 状态机校验:仅审核中可驳回
	if c.Status != model.ClubStatusReviewing {
		return fmt.Errorf("俱乐部当前状态(%d)不允许驳回,仅审核中可驳回", c.Status)
	}
	now := nowTimePtr()
	fields := map[string]interface{}{
		"status":        model.ClubStatusRejected,
		"reject_reason": reason,
		"reject_count":  c.RejectCount + 1,
		"updated_at":    now,
	}
	// 驳回次数达到 3 次，锁定 7 天禁止再次提交入驻
	if c.RejectCount+1 >= 3 {
		locked := now.AddDate(0, 0, 7)
		fields["locked_until"] = locked
	}
	if err := clubRepo.Update(clubID, fields); err != nil {
		return err
	}
	_ = RevokeClubVBadge(clubID)
	_ = reason
	return nil
}

// AdminFreezeClub 冻结俱乐部
// - 俱乐部状态置为冻结
// - 撤销 V 标
// - 级联禁用:解散所有群聊(status=0)、禁用内置管理端账号、进行中售后会话标记关闭
func AdminFreezeClub(clubID int64, reason string) error {
	c, _ := clubRepo.FindByID(clubID)
	if c == nil {
		return errors.New("俱乐部不存在")
	}
	if c.Status == model.ClubStatusCanceled {
		return errors.New("俱乐部已注销,不可冻结")
	}
	now := nowTimePtr()
	err := db.Transaction(func(tx *gorm.DB) error {
		// 1. 俱乐部置为冻结
		if err := tx.Model(&model.Club{}).Where("id = ?", clubID).
			Updates(map[string]interface{}{"status": model.ClubStatusFrozen, "updated_at": now}).Error; err != nil {
			return err
		}
		// 2. 解散所有群聊
		if err := tx.Model(&model.GroupChat{}).Where("club_id = ?", clubID).
			Updates(map[string]interface{}{"status": 0, "updated_at": now}).Error; err != nil {
			return err
		}
		// 3. 禁用内置管理端账号
		if err := tx.Model(&model.ShopAdminAccount{}).Where("club_id = ? AND status = 1", clubID).
			Updates(map[string]interface{}{"status": 0, "updated_at": now}).Error; err != nil {
			return err
		}
		// 4. 关闭进行中售后会话
		return tx.Model(&model.AfterSaleSession{}).
			Where("club_id = ? AND status = 1", clubID).
			Updates(map[string]interface{}{"status": 2, "updated_at": now}).Error
	})
	if err != nil {
		return err
	}
	_ = RevokeClubVBadge(clubID)
	_ = reason
	return nil
}

// AdminUnfreezeClub 解冻俱乐部 + 恢复 V 标
// 安全修复:状态机校验,仅"冻结"状态可解冻(原可将审核中/已注销的俱乐部直接变更为通过)
func AdminUnfreezeClub(clubID int64) error {
	c, _ := clubRepo.FindByID(clubID)
	if c == nil {
		return errors.New("俱乐部不存在")
	}
	// 状态机校验:仅冻结状态可解冻
	if c.Status != model.ClubStatusFrozen {
		return fmt.Errorf("俱乐部当前状态(%d)不允许解冻,仅冻结状态可解冻", c.Status)
	}
	if err := clubRepo.Update(clubID, map[string]interface{}{
		"status":     model.ClubStatusApproved,
		"updated_at": nowTimePtr(),
	}); err != nil {
		return err
	}
	_ = GrantClubVBadge(clubID, int64(c.Type))
	return nil
}

// AdminCancelClub 注销俱乐部(安全注销)
// - 前置条件:无进行中订单(待接单/已接单/进行中/待验收)
// - 缩写封存:写入 club_abbreviations,该缩写此后不可被其他俱乐部复用
// - 资料归档:俱乐部资料 JSON 加密(AES-256) + SHA-256 哈希 + 上链存证,写入 club_archives
// - 状态置为注销 + is_archived=1 + 撤销 V 标
func AdminCancelClub(clubID int64, reason string) error {
	c, _ := clubRepo.FindByID(clubID)
	if c == nil {
		return errors.New("俱乐部不存在")
	}
	if c.Status == model.ClubStatusCanceled {
		return errors.New("俱乐部已注销")
	}
	// 前置条件:无进行中订单
	var pendingCnt int64
	if err := db.Model(&model.Order{}).
		Where("club_id = ? AND status IN ?", clubID, []int8{
			model.OrderStatusPending, model.OrderStatusAccepted,
			model.OrderStatusInProgress, model.OrderStatusToVerify,
		}).Count(&pendingCnt).Error; err != nil {
		return err
	}
	if pendingCnt > 0 {
		return errors.New("存在进行中订单,不可注销")
	}
	now := nowTimePtr()
	// 1. 缩写封存 + 状态注销
	err := db.Transaction(func(tx *gorm.DB) error {
		// 封存缩写(若尚未封存)
		var abbrCnt int64
		_ = tx.Model(&model.ClubAbbreviation{}).Where("abbreviation = ?", c.Abbreviation).Count(&abbrCnt).Error
		if abbrCnt == 0 && c.Abbreviation != "" {
			if err := tx.Create(&model.ClubAbbreviation{
				Abbreviation: c.Abbreviation, ClubID: clubID,
				AbandonedAt: now, CreatedAt: now,
			}).Error; err != nil {
				return err
			}
		}
		// 状态置为注销 + 归档标记
		return tx.Model(&model.Club{}).Where("id = ?", clubID).
			Updates(map[string]interface{}{
				"status":      model.ClubStatusCanceled,
				"is_archived": int8(1),
				"updated_at":  now,
			}).Error
	})
	if err != nil {
		return err
	}
	// 2. 资料归档:JSON 加密 + 哈希 + 上链存证
	archiveClubData(c, reason)
	// 3. 撤销 V 标
	_ = RevokeClubVBadge(clubID)
	return nil
}

// archiveClubData 将俱乐部资料归档加密并上链存证
// 失败不阻断注销主流程(已通过事务完成状态变更)
func archiveClubData(c *model.Club, reason string) {
	if c == nil || db == nil {
		return
	}
	data, err := json.Marshal(c)
	if err != nil {
		return
	}
	encKey := utils.PadKey(cfg.JWT.Secret)
	encrypted, encErr := utils.EncryptFile(data, encKey)
	hash := utils.GenerateFileHash(data)
	rec := &model.ClubArchive{
		ClubID:      c.ID,
		ArchiveData: data,
		Encrypted:   encErr == nil,
		FileHash:    hash,
		ArchivedAt:  nowTimePtr(),
		CreatedAt:   nowTimePtr(),
	}
	if encErr == nil {
		// 加密成功则用密文覆盖归档内容,并上链存证
		rec.ArchiveData = json.RawMessage(encrypted)
		txID, bcErr := utils.UploadToBlockchain(hash, map[string]string{
			"file_type": "club_archive",
			"ref_type":  "club",
			"ref_id":    itoa(c.ID),
			"reason":    reason,
		})
		if bcErr == nil {
			rec.BlockchainTxID = txID
		}
	}
	_ = db.Create(rec).Error
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
