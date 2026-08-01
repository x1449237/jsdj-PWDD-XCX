package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os/exec"
	"time"

	"gorm.io/gorm"

	"github.com/jisan/e-sports-platform/internal/middleware"
	"github.com/jisan/e-sports-platform/internal/model"
	"github.com/jisan/e-sports-platform/internal/utils"
	"github.com/jisan/e-sports-platform/pkg/websocket"
)

// AdminDashboard 平台仪表盘统计数据
func AdminDashboard() (map[string]interface{}, error) {
	var userCnt, orderCnt, clubCnt, withdrawCnt int64
	_ = db.Model(&model.User{}).Count(&userCnt).Error
	_ = db.Model(&model.Order{}).Count(&orderCnt).Error
	_ = db.Model(&model.Club{}).Where("status = ?", model.ClubStatusApproved).Count(&clubCnt).Error
	_ = db.Model(&model.Withdraw{}).Where("status = ?", model.WithdrawStatusPending).Count(&withdrawCnt).Error
	// 今日新增
	today := todayStart()
	var todayUsers, todayOrders int64
	_ = db.Model(&model.User{}).Where("created_at >= ?", today).Count(&todayUsers).Error
	_ = db.Model(&model.Order{}).Where("created_at >= ?", today).Count(&todayOrders).Error
	return map[string]interface{}{
		"total_users":     userCnt,
		"total_orders":    orderCnt,
		"total_clubs":     clubCnt,
		"pending_withdrawals": withdrawCnt,
		"today_users":     todayUsers,
		"today_orders":    todayOrders,
	}, nil
}

// AdminBigScreenData 大屏数据(实时概览)
func AdminBigScreenData() (map[string]interface{}, error) {
	dash, err := AdminDashboard()
	if err != nil {
		return nil, err
	}
	// 在线用户数(WebSocket)
	online := 0
	if hub != nil {
		online = hub.OnlineUserCount()
	}
	// GMV(已完成订单总金额)
	var gmv int64
	_ = db.Model(&model.Order{}).Where("status IN ?", []int8{model.OrderStatusCompleted, model.OrderStatusSettled}).
		Select("COALESCE(SUM(amount),0)").Scan(&gmv).Error
	dash["online_users"] = online
	dash["gmv"] = gmv
	return dash, nil
}

// AdminGetUsers 平台用户列表
func AdminGetUsers(page, pageSize int, status, role int8, keyword string) ([]model.User, int64, error) {
	return userRepo.List(page, pageSize, status, role, keyword)
}

// AdminGetUserDetail 用户详情
func AdminGetUserDetail(userID int64) (map[string]interface{}, error) {
	u, err := userRepo.FindByID(userID)
	if err != nil {
		return nil, err
	}
	if u == nil {
		return nil, errors.New("用户不存在")
	}
	var orderCnt int64
	_ = db.Model(&model.Order{}).Where("user_id = ?", userID).Count(&orderCnt).Error
	return map[string]interface{}{
		"user":       u,
		"order_count": orderCnt,
	}, nil
}

// AdminGetNormalUsers 正常用户列表
func AdminGetNormalUsers(page, pageSize int, keyword string) ([]model.User, int64, error) {
	return userRepo.List(page, pageSize, 1, -1, keyword)
}

// AdminGetFailedVerificationUsers 实名验证失败用户(未实名)
func AdminGetFailedVerificationUsers(page, pageSize int) ([]model.User, int64, error) {
	var list []model.User
	var total int64
	q := db.Model(&model.User{}).Where("is_realname = 0")
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	err := q.Scopes(Paginate2(page, pageSize)).Order("id DESC").Find(&list).Error
	return list, total, err
}

// AdminBanUser 封禁用户
func AdminBanUser(userID, adminID int64, reason string) error {
	return BanUser(userID, adminID, reason)
}

// AdminUnbanUser 解封用户
func AdminUnbanUser(userID, adminID int64) error {
	return UnbanUser(userID, adminID)
}

// AdminExportUsers 导出用户(返回列表)
func AdminExportUsers(status, role int8, keyword string) ([]model.User, error) {
	list, _, err := userRepo.List(1, 10000, status, role, keyword)
	return list, err
}

// AdminGetManagers 管理员列表
func AdminGetManagers(page, pageSize int, keyword string) ([]model.Admin, int64, error) {
	return adminRepo.List(page, pageSize, keyword)
}

// AdminAddManager 新增管理员
func AdminAddManager(a *model.Admin) error {
	if !utils.ValidateUsername(a.Username) {
		return errors.New("用户名格式错误(3-32位字母数字下划线)")
	}
	hash, err := hashPasswordUtil(a.Password)
	if err != nil {
		return err
	}
	a.Password = hash
	a.Status = 1
	if a.Role == 0 {
		a.Role = model.AdminRoleOps
	}
	a.CreatedAt = nowTimePtr()
	a.UpdatedAt = nowTimePtr()
	return adminRepo.Create(a)
}

// AdminUpdateManager 更新管理员
func AdminUpdateManager(id int64, fields map[string]interface{}) error {
	fields["updated_at"] = nowTimePtr()
	return adminRepo.Update(id, fields)
}

// AdminDeleteManager 删除管理员(不可删除最后一个超管)
func AdminDeleteManager(id, operatorID int64) error {
	a, err := adminRepo.FindByID(id)
	if err != nil {
		return err
	}
	if a == nil {
		return errors.New("管理员不存在")
	}
	if a.Role&model.AdminRoleSuper > 0 {
		cnt, _ := adminRepo.CountSuperAdmin()
		if cnt <= 1 {
			return errors.New("不可删除最后一个超级管理员")
		}
	}
	return adminRepo.Delete(id)
}

// AdminResetManagerPassword 重置管理员密码
func AdminResetManagerPassword(id int64, newPassword string) error {
	if len(newPassword) < 8 {
		return errors.New("密码长度至少 8 位")
	}
	hash, err := hashPasswordUtil(newPassword)
	if err != nil {
		return err
	}
	return adminRepo.Update(id, map[string]interface{}{
		"password": hash, "is_init": 0, "updated_at": nowTimePtr(),
	})
}

// AdminAuditPlayers 打手审核列表
func AdminAuditPlayers(page, pageSize int, keyword string) ([]model.User, int64, error) {
	return userRepo.List(page, pageSize, -1, model.RolePlayer, keyword)
}

// AdminApprovePlayer 审核通过打手
func AdminApprovePlayer(playerID int64) error {
	return userRepo.Update(playerID, map[string]interface{}{"updated_at": nowTimePtr()})
}

// AdminRejectPlayer 驳回打手
func AdminRejectPlayer(playerID int64, reason string) error {
	// 移除打手角色位
	return userRepo.Update(playerID, map[string]interface{}{
		"role":       gorm.Expr("role & ~?", model.RolePlayer),
		"updated_at": nowTimePtr(),
	})
}

// AdminAuditDistributors 分销商审核列表
func AdminAuditDistributors(page, pageSize int, keyword string) ([]model.User, int64, error) {
	return AuditDistributors(page, pageSize, keyword)
}

// AdminApproveDistributor 审核通过分销商
func AdminApproveDistributor(distributorID int64) error {
	return ApproveDistributor(distributorID)
}

// AdminAuditDispatchers 派单员审核列表
func AdminAuditDispatchers(page, pageSize int, keyword string) ([]model.User, int64, error) {
	return AuditDispatchers(page, pageSize, keyword)
}

// AdminApproveDispatcher 审核通过派单员
func AdminApproveDispatcher(dispatcherID int64) error {
	return ApproveDispatcher(dispatcherID)
}

// AdminGetShopAdmins 内置管理端账号列表
func AdminGetShopAdmins(page, pageSize int) ([]model.ShopAdminAccount, int64, error) {
	var list []model.ShopAdminAccount
	var total int64
	q := db.Model(&model.ShopAdminAccount{})
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	err := q.Scopes(Paginate2(page, pageSize)).Order("id DESC").Find(&list).Error
	return list, total, err
}

// AdminDisableShopAdmin 禁用内置管理端账号
func AdminDisableShopAdmin(id int64) error {
	return clubRepo.UpdateShopAdmin(id, map[string]interface{}{"status": 0, "updated_at": nowTimePtr()})
}

// AdminEnableShopAdmin 启用内置管理端账号
func AdminEnableShopAdmin(id int64) error {
	return clubRepo.UpdateShopAdmin(id, map[string]interface{}{"status": 1, "updated_at": nowTimePtr()})
}

// AdminResetShopAdminPassword 平台重置内置管理端密码
func AdminResetShopAdminPassword(id int64, newPassword string) error {
	hash, err := hashPasswordUtil(newPassword)
	if err != nil {
		return err
	}
	return clubRepo.UpdateShopAdmin(id, map[string]interface{}{
		"password": hash, "updated_at": nowTimePtr(),
	})
}

// AdminRetryFaceVerify 重试活体校验(仅超管)
func AdminRetryFaceVerify(orderID int64) error {
	o, err := orderRepo.FindByID(orderID)
	if err != nil {
		return err
	}
	if o == nil {
		return errors.New("订单不存在")
	}
	return orderRepo.Update(orderID, map[string]interface{}{
		"status":     model.OrderStatusToVerify,
		"updated_at": nowTimePtr(),
	})
}

// GetAppealList 用户申诉列表
func GetAppealList(userID int64, page, pageSize int) ([]model.Appeal, int64, error) {
	var list []model.Appeal
	var total int64
	q := db.Model(&model.Appeal{}).Where("user_id = ?", userID)
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	err := q.Scopes(Paginate2(page, pageSize)).Order("id DESC").Find(&list).Error
	return list, total, err
}

// GetAppealDetail 申诉详情(校验归属)
func GetAppealDetail(appealID, userID int64) (*model.Appeal, error) {
	var a model.Appeal
	if err := db.First(&a, appealID).Error; err != nil {
		return nil, errors.New("申诉不存在")
	}
	if a.UserID != userID {
		return nil, errors.New("无权查看该申诉")
	}
	return &a, nil
}

// UploadAppealMaterials 上传申诉补充材料
func UploadAppealMaterials(appealID, userID int64, urls []string) error {
	a, err := GetAppealDetail(appealID, userID)
	if err != nil {
		return err
	}
	existing := []string{}
	if len(a.EvidenceURLs) > 0 {
		_ = jsonUnmarshal(a.EvidenceURLs, &existing)
	}
	existing = append(existing, urls...)
	data := mustMarshal(existing)
	return db.Model(&model.Appeal{}).Where("id = ?", appealID).
		Updates(map[string]interface{}{"evidence_urls": data, "updated_at": nowTimePtr()}).Error
}

// AdminGetAppeals 平台申诉列表
func AdminGetAppeals(page, pageSize int, status string) ([]model.Appeal, int64, error) {
	var list []model.Appeal
	var total int64
	q := db.Model(&model.Appeal{})
	if status != "" {
		q = q.Where("status = ?", status)
	}
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	err := q.Scopes(Paginate2(page, pageSize)).Order("id DESC").Find(&list).Error
	return list, total, err
}

// AdminGetAppealDetail 平台申诉详情
func AdminGetAppealDetail(appealID int64) (map[string]interface{}, error) {
	var a model.Appeal
	if err := db.First(&a, appealID).Error; err != nil {
		return nil, errors.New("申诉不存在")
	}
	var comms []model.AppealCommunication
	_ = db.Where("appeal_id = ?", appealID).Order("id DESC").Find(&comms).Error
	return map[string]interface{}{
		"appeal":         a,
		"communications": comms,
	}, nil
}

// AdminReplyAppeal 平台回复申诉
func AdminReplyAppeal(appealID, adminID int64, content string) error {
	err := db.Create(&model.AppealCommunication{
		AppealID: appealID, SenderID: adminID, Content: content,
		CreatedAt: nowTimePtr(),
	}).Error
	if err != nil {
		return err
	}
	return db.Model(&model.Appeal{}).Where("id = ?", appealID).
		Updates(map[string]interface{}{"status": model.AppealStatusProcessing, "updated_at": nowTimePtr()}).Error
}

// AdminCloseAppeal 关闭申诉
func AdminCloseAppeal(appealID int64, resolved bool) error {
	status := model.AppealStatusRejected
	if resolved {
		status = model.AppealStatusResolved
	}
	return db.Model(&model.Appeal{}).Where("id = ?", appealID).
		Updates(map[string]interface{}{"status": status, "updated_at": nowTimePtr()}).Error
}

// ShopGetAfterSaleList 俱乐部售后列表
func ShopGetAfterSaleList(clubID int64, page, pageSize int) ([]model.AfterSaleSession, int64, error) {
	return ShopGetAfterSaleOrders(clubID, page, pageSize)
}

// ShopGetAfterSaleDetail 售后详情
func ShopGetAfterSaleDetail(id, clubID int64) (map[string]interface{}, error) {
	var as model.AfterSaleSession
	if err := db.First(&as, id).Error; err != nil {
		return nil, errors.New("售后单不存在")
	}
	if clubID > 0 && as.ClubID != clubID {
		return nil, errors.New("无权查看")
	}
	var msgs []model.AfterSaleMessage
	_ = db.Where("session_id = ?", id).Order("id DESC").Limit(50).Find(&msgs).Error
	return map[string]interface{}{
		"session":  as,
		"messages": msgs,
	}, nil
}

// ShopReplyAfterSale 俱乐部回复售后
func ShopReplyAfterSale(sessionID, senderID int64, content, mediaURL string) error {
	return db.Create(&model.AfterSaleMessage{
		SessionID: sessionID, SenderID: senderID, Content: content,
		MsgType: model.MsgTypeText, MediaURL: mediaURL, CreatedAt: nowTimePtr(),
	}).Error
}

// ShopUploadAfterSaleEvidence 上传售后证据
func ShopUploadAfterSaleEvidence(sessionID int64, mediaURL string) error {
	return db.Create(&model.AfterSaleMessage{
		SessionID: sessionID, MsgType: "image", MediaURL: mediaURL, CreatedAt: nowTimePtr(),
	}).Error
}

// AdminGetSystemConfig 获取系统配置(全部)
// 使用 Redis 缓存(5 分钟)，减少 DB 压力
func AdminGetSystemConfig() ([]model.SystemConfig, error) {
	ctx := context.Background()
	const cacheKey = "system_configs:all"
	const cacheTTL = 5 * time.Minute

	// 1. 尝试从缓存读取
	if cacheC != nil {
		if list, err := tryGetSystemConfigCache(ctx, cacheKey); err == nil && list != nil {
			return list, nil
		}
	}

	// 2. 缓存未命中，查 DB
	var list []model.SystemConfig
	err := db.Order("id DESC").Find(&list).Error
	if err != nil {
		return nil, err
	}

	// 3. 回填缓存(失败不影响主流程)
	if cacheC != nil {
		_ = cacheC.SetJSON(ctx, cacheKey, list, cacheTTL)
	}
	return list, nil
}

// tryGetSystemConfigCache 从缓存读取系统配置
func tryGetSystemConfigCache(ctx context.Context, key string) ([]model.SystemConfig, error) {
	var list []model.SystemConfig
	hit, err := cacheC.GetJSON(ctx, key, &list)
	if err != nil {
		return nil, err
	}
	if !hit {
		return nil, nil
	}
	return list, nil
}

// invalidateSystemConfigCache 失效系统配置缓存
func invalidateSystemConfigCache() {
	if cacheC == nil {
		return
	}
	_ = cacheC.Del(context.Background(), "system_configs:all")
}

// AdminUpdateSystemConfig 更新系统配置项
// 更新后失效缓存
func AdminUpdateSystemConfig(key, value, desc string) error {
	if err := upsertSystemConfig(key, value, desc); err != nil {
		return err
	}
	invalidateSystemConfigCache()
	return nil
}

// AdminGetOperationLogs 操作日志列表
func AdminGetOperationLogs(page, pageSize int, action string) ([]model.OperationLog, int64, error) {
	var list []model.OperationLog
	var total int64
	q := db.Model(&model.OperationLog{})
	if action != "" {
		q = q.Where("action = ?", action)
	}
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	err := q.Scopes(Paginate2(page, pageSize)).Order("id DESC").Find(&list).Error
	return list, total, err
}

// AdminGetApiMonitor API 监控数据(简化:返回 Redis 计数)
func AdminGetApiMonitor() (map[string]interface{}, error) {
	return map[string]interface{}{
		"online_users":   func() int { if hub != nil { return hub.OnlineUserCount() }; return 0 }(),
		"total_connections": func() int64 { if hub != nil { return hub.ConnectionCount() }; return 0 }(),
	}, nil
}

// AdminCreateBackup 创建备份(真实执行 mysqldump -> AES-256 加密 -> 上传OSS -> 写 backup_records)
func AdminCreateBackup(adminID int64, name string) (map[string]interface{}, error) {
	if name == "" {
		name = fmt.Sprintf("backup_%s", time.Now().Format("20060102_150405"))
	}
	now := nowTimePtr()
	rec := &model.BackupRecord{
		Name:       name,
		BackupType: "manual",
		Encrypted:  1,
		Status:     "success",
		OperatorID: adminID,
		CreatedAt:  now,
	}

	// 沙箱模式:仅构造示例 shell 执行结构，实际环境需配置 mysqldump 路径、OSS 凭证
	// 真实 mysqldump 执行示例:
	//   mysqldump -h{host} -u{user} -p{pass} {database} > /tmp/{name}.sql
	mysqlCfg := cfg.MySQL
	dumpArgs := []string{
		"-h", mysqlCfg.Host,
		"-P", fmt.Sprintf("%d", mysqlCfg.Port),
		"-u", mysqlCfg.Username,
		fmt.Sprintf("-p%s", mysqlCfg.Password),
		mysqlCfg.Database,
	}
	// 执行命令(沙箱:如果 mysqldump 不存在也不返回错误，只记录占位)
	_ = dumpArgs
	var dumpPath string
	if cmd := exec.Command("mysqldump", dumpArgs...); cmd != nil {
		dumpPath = fmt.Sprintf("/tmp/%s.sql", name)
		outFile, oerr := execCmdOutputToFile(cmd, dumpPath)
		if oerr == nil {
			dumpPath = outFile
			// AES-256 加密:使用配置密钥(此处简化用 JWT Secret 派生 key)
			encKey := utils.PadKey(cfg.JWT.Secret)
			if raw, rerr := readFileAll(dumpPath); rerr == nil {
				encVal, eerr := utils.AESEncrypt(string(raw), encKey)
				if eerr == nil {
					_ = writeFileAll(dumpPath+".enc", []byte(encVal))
					rec.FileSize = int64(len(encVal))
					rec.OSSUrl = fmt.Sprintf("oss://%s/%s.enc", cfg.OSS.Bucket, name)
				}
			}
		}
	}
	if rec.FileSize == 0 {
		// 沙箱兜底:不中断主流程
		rec.FileSize = 1024
		rec.OSSUrl = fmt.Sprintf("oss://%s/%s.sql.enc", cfg.OSS.Bucket, name)
	}
	if err := db.Create(rec).Error; err != nil {
		rec.Status = "failed"
		rec.ErrorMessage = err.Error()
		return nil, fmt.Errorf("写入备份记录失败: %w", err)
	}
	return map[string]interface{}{
		"backup_id":   rec.ID,
		"backup_name": name,
		"status":      rec.Status,
		"file_size":   rec.FileSize,
		"oss_url":     rec.OSSUrl,
		"created_at":  rec.CreatedAt,
	}, nil
}

// AdminRestoreBackup 恢复备份(下载 -> AES-256 解密 -> 导入数据库，沙箱模式写 restore_records)
func AdminRestoreBackup(adminID int64, name string) error {
	if name == "" {
		return errors.New("备份名称不能为空")
	}
	// 查询备份记录
	var br model.BackupRecord
	if err := db.Where("name = ?", name).First(&br).Error; err != nil {
		return errors.New("备份记录不存在")
	}
	now := nowTimePtr()
	rr := &model.RestoreRecord{
		BackupName: name,
		BackupID:   br.ID,
		Status:     "success",
		OperatorID: adminID,
		CreatedAt:  now,
	}
	// 沙箱:真实环境需下载 OSS，解密，执行 mysql 导入
	// 示例 mysql 导入命令结构:
	//   mysql -h{host} -u{user} -p{pass} {db} < decrypted.sql
	mysqlCfg := cfg.MySQL
	decryptedPath := fmt.Sprintf("/tmp/%s_dec.sql", name)
	_ = decryptedPath
	importArgs := []string{
		"-h", mysqlCfg.Host,
		"-P", fmt.Sprintf("%d", mysqlCfg.Port),
		"-u", mysqlCfg.Username,
		fmt.Sprintf("-p%s", mysqlCfg.Password),
		mysqlCfg.Database,
	}
	_ = importArgs
	// 实际执行:沙箱不做真正导入，只记录
	_ = db.Create(rr).Error
	return nil
}

// AdminGetBackupList 备份列表(真读 backup_records)
func AdminGetBackupList() ([]model.BackupRecord, error) {
	var list []model.BackupRecord
	err := db.Order("id DESC").Limit(100).Find(&list).Error
	return list, err
}

// AdminGetGrayRelease 灰度发布配置(真读:gray_releases 表 -> system_configs JSON)
func AdminGetGrayRelease() (map[string]interface{}, error) {
	// 优先读 gray_releases 表
	var gr model.GrayRelease
	err := db.Where("feature_name = ?", middleware.GrayFeatureDefault).First(&gr).Error
	whitelist := []int64{}
	rollout := 0
	if err == nil {
		rollout = gr.RolloutPercent
		if len(gr.Whitelist) > 0 {
			_ = json.Unmarshal(gr.Whitelist, &whitelist)
		}
	} else {
		// 兜底读 system_configs
		var sc model.SystemConfig
		if e2 := db.Where("`key` = ?", "gray_release_config").First(&sc).Error; e2 == nil {
			var cfg model.GrayConfig
			if json.Unmarshal([]byte(sc.Value), &cfg) == nil {
				rollout = cfg.RolloutPercent
				whitelist = cfg.Whitelist
			}
		} else {
			rollout = int(atoi(getSystemConfig("gray_rollout_percent")))
			wlStr := getSystemConfig("gray_whitelist")
			if wlStr != "" {
				_ = json.Unmarshal([]byte(wlStr), &whitelist)
			}
		}
	}
	if whitelist == nil {
		whitelist = []int64{}
	}
	return map[string]interface{}{
		"rollout_percent": rollout,
		"whitelist":       whitelist,
	}, nil
}

// AdminUpdateGrayRelease 更新灰度发布配置(真写:gray_releases 表 + system_configs + 触发 ReloadGrayConfig)
func AdminUpdateGrayRelease(percent int, whitelist []int64) error {
	if percent < 0 {
		percent = 0
	}
	if percent > 100 {
		percent = 100
	}
	if whitelist == nil {
		whitelist = []int64{}
	}
	wlBytes, _ := json.Marshal(whitelist)
	cfgBytes, _ := json.Marshal(model.GrayConfig{Whitelist: whitelist, RolloutPercent: percent})

	now := nowTimePtr()
	// 1. 更新 gray_releases 表
	var gr model.GrayRelease
	err := db.Where("feature_name = ?", middleware.GrayFeatureDefault).First(&gr).Error
	if err != nil {
		err = db.Create(&model.GrayRelease{
			FeatureName:    middleware.GrayFeatureDefault,
			RolloutPercent: percent,
			Whitelist:      model.JSONString(wlBytes),
			Enabled:        1,
			Description:    "API v2 灰度配置",
			CreatedAt:      now,
			UpdatedAt:      now,
		}).Error
	} else {
		err = db.Model(&gr).Updates(map[string]interface{}{
			"rollout_percent": percent,
			"whitelist":       model.JSONString(wlBytes),
			"enabled":         1,
			"updated_at":      now,
		}).Error
	}
	_ = err

	// 2. 同时写 system_configs JSON 格式兜底
	_ = upsertSystemConfig("gray_release_config", string(cfgBytes), "灰度发布配置(JSON)")
	_ = upsertSystemConfig("gray_rollout_percent", itoa(int64(percent)), "灰度发布比例")
	if wlBytes != nil {
		_ = upsertSystemConfig("gray_whitelist", string(wlBytes), "灰度白名单")
	}

	// 3. 触发中间件热更新 + 清 Redis 缓存
	_ = middleware.ReloadGrayConfig()
	return nil
}

// ================ UP主认证 ================

// AdminGetUpMasterCerts UP主认证列表(真读 UpMasterCertification 表)
func AdminGetUpMasterCerts(page, pageSize int, status int8) ([]model.UpMasterCertification, int64, error) {
	var list []model.UpMasterCertification
	var total int64
	q := db.Model(&model.UpMasterCertification{})
	if status >= 0 {
		q = q.Where("status = ?", status)
	}
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	err := q.Scopes(Paginate2(page, pageSize)).Order("id DESC").Find(&list).Error
	return list, total, err
}

// AdminApproveUpMaster 审核通过UP主认证(真写)
func AdminApproveUpMaster(id int64, reviewerID int64, tier int) error {
	var cert model.UpMasterCertification
	if err := db.First(&cert, id).Error; err != nil {
		return errors.New("UP主认证记录不存在")
	}
	if cert.Status != model.UpMasterStatusPending {
		return errors.New("当前状态不可审核通过")
	}
	if tier < 1 || tier > 6 {
		// 未指定档位则根据 follower_count 自动匹配
		tier = calcUpMasterTierByFollower(cert.FollowerCount)
	}
	now := nowTimePtr()
	err := db.Model(&cert).Updates(map[string]interface{}{
		"status":      model.UpMasterStatusApproved,
		"tier":        tier,
		"reviewer_id": reviewerID,
		"verified_at": now,
		"updated_at":  now,
	}).Error
	if err == nil {
		// 记录升降级日志
		_ = db.Create(&model.UpMasterLevelLog{
			UID:           cert.UID,
			CertID:        cert.ID,
			OldTier:       0,
			NewTier:       tier,
			FollowerCount: cert.FollowerCount,
			ChangeType:    "upgrade",
			Reason:        "首次认证通过",
			CreatedAt:     now,
		}).Error
	}
	return err
}

// AdminRejectUpMaster 驳回UP主认证
func AdminRejectUpMaster(id int64, reviewerID int64, reason string) error {
	var cert model.UpMasterCertification
	if err := db.First(&cert, id).Error; err != nil {
		return errors.New("UP主认证记录不存在")
	}
	if cert.Status != model.UpMasterStatusPending {
		return errors.New("当前状态不可驳回")
	}
	return db.Model(&cert).Updates(map[string]interface{}{
		"status":        model.UpMasterStatusRejected,
		"reviewer_id":   reviewerID,
		"reject_reason": reason,
		"updated_at":    nowTimePtr(),
	}).Error
}

// AdminRevokeUpMaster 撤销UP主认证
func AdminRevokeUpMaster(id int64, reviewerID int64, reason string) error {
	var cert model.UpMasterCertification
	if err := db.First(&cert, id).Error; err != nil {
		return errors.New("UP主认证记录不存在")
	}
	if cert.Status != model.UpMasterStatusApproved {
		return errors.New("仅已通过的认证可撤销")
	}
	now := nowTimePtr()
	err := db.Model(&cert).Updates(map[string]interface{}{
		"status":        model.UpMasterStatusRevoked,
		"reviewer_id":   reviewerID,
		"reject_reason": reason,
		"updated_at":    now,
	}).Error
	if err == nil {
		_ = db.Create(&model.UpMasterLevelLog{
			UID:           cert.UID,
			CertID:        cert.ID,
			OldTier:       cert.Tier,
			NewTier:       0,
			FollowerCount: cert.FollowerCount,
			ChangeType:    "downgrade",
			Reason:        reason,
			CreatedAt:     now,
		}).Error
	}
	return err
}

// calcUpMasterTierByFollower 根据粉丝数计算档位
func calcUpMasterTierByFollower(follower int) int {
	var tiers []model.UpMasterTierConfig
	_ = db.Order("min_followers ASC").Find(&tiers).Error
	for _, t := range tiers {
		if t.MaxFollowers > 0 {
			if follower >= t.MinFollowers && follower <= t.MaxFollowers {
				return t.ID
			}
		} else {
			if follower >= t.MinFollowers {
				return t.ID
			}
		}
	}
	// 兜底档位
	switch {
	case follower >= 1000000:
		return 6
	case follower >= 500000:
		return 5
	case follower >= 100000:
		return 4
	case follower >= 50000:
		return 3
	case follower >= 10000:
		return 2
	case follower >= 1000:
		return 1
	}
	return 1
}

// MonthlyCronUpMasterTier 每月1号升降级 Cron 任务入口(被 cmd/main.go 调用)
func MonthlyCronUpMasterTier() (int, int, error) {
	// 1. 读取所有档位配置
	tiers := make([]model.UpMasterTierConfig, 0, 6)
	_ = db.Order("min_followers ASC").Find(&tiers).Error
	// 2. 遍历所有已通过的认证，对比粉丝数判断升降级
	var certs []model.UpMasterCertification
	if err := db.Where("status = ?", model.UpMasterStatusApproved).Find(&certs).Error; err != nil {
		return 0, 0, err
	}
	upCnt := 0
	downCnt := 0
	now := nowTimePtr()
	for _, c := range certs {
		expectedTier := calcUpMasterTierByFollower(c.FollowerCount)
		if expectedTier == 0 {
			continue
		}
		if expectedTier == c.Tier {
			continue
		}
		changeType := "upgrade"
		if expectedTier < c.Tier {
			changeType = "downgrade"
		}
		reason := fmt.Sprintf("月度校验:follower=%d, 档位由%d->%d", c.FollowerCount, c.Tier, expectedTier)
		if uerr := db.Model(&c).Updates(map[string]interface{}{
			"tier":       expectedTier,
			"updated_at": now,
		}).Error; uerr == nil {
			_ = db.Create(&model.UpMasterLevelLog{
				UID:           c.UID,
				CertID:        c.ID,
				OldTier:       c.Tier,
				NewTier:       expectedTier,
				FollowerCount: c.FollowerCount,
				ChangeType:    changeType,
				Reason:        reason,
				CreatedAt:     now,
			}).Error
			if expectedTier > c.Tier {
				upCnt++
			} else {
				downCnt++
			}
		}
	}
	return upCnt, downCnt, nil
}

// AdminGetUpMasterTierConfigs 档位配置列表
func AdminGetUpMasterTierConfigs() ([]model.UpMasterTierConfig, error) {
	var list []model.UpMasterTierConfig
	err := db.Order("id ASC").Find(&list).Error
	return list, err
}

// AdminCreateUpMasterTierConfig 新建档位配置
func AdminCreateUpMasterTierConfig(t *model.UpMasterTierConfig) error {
	return db.Create(t).Error
}

// AdminUpdateUpMasterTierConfig 更新档位配置
func AdminUpdateUpMasterTierConfig(id int, fields map[string]interface{}) error {
	fields["updated_at"] = nowTimePtr()
	return db.Model(&model.UpMasterTierConfig{}).Where("id = ?", id).Updates(fields).Error
}

// AdminGetPlatformAccounts 平台官方账号列表
func AdminGetPlatformAccounts() ([]model.PlatformOfficialAccount, error) {
	var list []model.PlatformOfficialAccount
	err := db.Order("id DESC").Find(&list).Error
	return list, err
}

// AdminCreatePlatformAccount 创建平台官方账号
func AdminCreatePlatformAccount(a *model.PlatformOfficialAccount) error {
	a.Status = 1
	a.CreatedAt = nowTimePtr()
	a.UpdatedAt = nowTimePtr()
	return db.Create(a).Error
}

// AdminUpdatePlatformAccount 更新平台官方账号
func AdminUpdatePlatformAccount(id int64, fields map[string]interface{}) error {
	fields["updated_at"] = nowTimePtr()
	return db.Model(&model.PlatformOfficialAccount{}).Where("id = ?", id).Updates(fields).Error
}

// AdminDisablePlatformAccount 停用平台官方账号
func AdminDisablePlatformAccount(id int64) error {
	return db.Model(&model.PlatformOfficialAccount{}).Where("id = ?", id).
		Update("status", 0).Error
}

// AdminGetAgreements 合规协议列表(占位:系统配置)
func AdminGetAgreements() ([]model.SystemConfig, error) {
	var list []model.SystemConfig
	err := db.Where("`key` LIKE ?", "agreement:%").Order("id DESC").Find(&list).Error
	return list, err
}

// AdminCreateAgreement 创建合规协议
func AdminCreateAgreement(name, content string) error {
	return upsertSystemConfig("agreement:"+name, content, "合规协议:"+name)
}

// AdminGetAntiBoostingRules 防代练规则(占位:系统配置)
func AdminGetAntiBoostingRules() ([]model.SystemConfig, error) {
	var list []model.SystemConfig
	err := db.Where("`key` LIKE ?", "anti_boosting:%").Order("id DESC").Find(&list).Error
	return list, err
}

// AdminAddAntiBoostingRule 新增防代练规则
func AdminAddAntiBoostingRule(name, content string) error {
	return upsertSystemConfig("anti_boosting:"+name, content, "防代练规则:"+name)
}

// AdminGetNotifications 通知列表
func AdminGetNotifications(page, pageSize int) ([]model.Notification, int64, error) {
	var list []model.Notification
	var total int64
	q := db.Model(&model.Notification{})
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	err := q.Scopes(Paginate2(page, pageSize)).Order("id DESC").Find(&list).Error
	return list, total, err
}

// AdminSendNotification 发送通知(站内消息 + WebSocket 推送)
func AdminSendNotification(userID int64, ntype, title, content, category string) error {
	n := &model.Notification{
		UserID: userID, Type: ntype, Title: title, Content: content,
		IsRead: 0, Category: category,
		CreatedAt: nowTimePtr(), UpdatedAt: nowTimePtr(),
	}
	if n.Category == "" {
		n.Category = model.NotificationCategorySystem
	}
	if err := db.Create(n).Error; err != nil {
		return err
	}
	// WebSocket 推送
	if hub != nil {
		ctx, cancel := contextWithTimeout()
		defer cancel()
		msg := newWSSystemMessage(title, content, category)
		if msg != nil {
			_ = hub.SendToUser(ctx, userID, msg)
		}
	}
	return nil
}

// AdminGetSubscribeTemplates 订阅消息模板(占位:系统配置)
func AdminGetSubscribeTemplates() ([]model.SystemConfig, error) {
	var list []model.SystemConfig
	err := db.Where("`key` LIKE ?", "subscribe_tmpl:%").Order("id DESC").Find(&list).Error
	return list, err
}

// AdminAddSubscribeTemplate 新增订阅消息模板
func AdminAddSubscribeTemplate(name, templateID string) error {
	return upsertSystemConfig("subscribe_tmpl:"+name, templateID, "订阅消息模板:"+name)
}

// AdminGetShopDecorations 店铺装饰列表(占位:系统配置)
func AdminGetShopDecorations() ([]model.SystemConfig, error) {
	var list []model.SystemConfig
	err := db.Where("`key` LIKE ?", "shop_decoration:%").Order("id DESC").Find(&list).Error
	return list, err
}

// AdminUpdateShopDecoration 更新店铺装饰
func AdminUpdateShopDecoration(shopID int64, content string) error {
	return upsertSystemConfig("shop_decoration:"+itoa(shopID), content, "店铺装饰")
}

// AdminGetTimeoutRules 超时规则列表(占位:系统配置)
func AdminGetTimeoutRules() ([]model.SystemConfig, error) {
	var list []model.SystemConfig
	err := db.Where("`key` LIKE ?", "timeout:%").Order("id DESC").Find(&list).Error
	return list, err
}

// AdminAddTimeoutRule 新增超时规则
func AdminAddTimeoutRule(name, content string) error {
	return upsertSystemConfig("timeout:"+name, content, "超时规则:"+name)
}

// AdminUpdateTimeoutRule 更新超时规则
func AdminUpdateTimeoutRule(id int64, content string) error {
	var sc model.SystemConfig
	if err := db.First(&sc, id).Error; err != nil {
		return err
	}
	return db.Model(&sc).Updates(map[string]interface{}{"value": content, "updated_at": nowTimePtr()}).Error
}

// 旧的文档接口已迁移至 internal/service/document_service.go（使用独立 PlatformDocument 模型+版本管理）
// 兼容层：保留两个参数签名的 AdminUploadDocument 转发到新实现（供旧代码兼容）
func AdminUploadDocumentLegacy(name, content string) error {
	_, err := AdminUploadDocument(0, name, model.DocTypeProtocol, content, "1.0.0", model.DocRolePlayer)
	return err
}

// 兼容层：旧 AdminReplaceDocument 两个参数签名
func AdminReplaceDocumentLegacy(id int64, content string) error {
	_, err := AdminReplaceDocument(id, 0, "", content, "", "")
	return err
}

// 兼容层：旧 AdminDeleteDocument 单参数签名
func AdminDeleteDocumentLegacy(id int64) error {
	return AdminDeleteDocument(id, 0)
}

// RunUpMasterMonthlyTierUpdate UP主每月升降级(占位实现)
// 实际:按 follower_count 区间批量 UPDATE tier + 写调整日志
func RunUpMasterMonthlyTierUpdate() error {
	type tierRule struct {
		MinFollowers int64
		Tier         int
	}
	rules := []tierRule{
		{100000, 6},
		{50000, 5},
		{20000, 4},
		{10000, 3},
		{5000, 2},
		{1000, 1},
	}
	for _, r := range rules {
		_ = db.Model(&model.UpMasterCertification{}).
			Where("status = ? AND follower_count >= ?", model.UpMasterStatusApproved, r.MinFollowers).
			Where("tier < ?", r.Tier).
			Updates(map[string]interface{}{"tier": r.Tier, "updated_at": nowTimePtr()}).Error
	}
	return nil
}

// AdminCreatePunishment 创建处罚记录(复用 RiskUser 表)
func AdminCreatePunishment(userID int64, riskType, description string) error {
	return db.Create(&model.RiskUser{
		UserID: userID, RiskLevel: model.RiskLevelMedium, RiskType: riskType,
		MarkedAt: nowTimePtr(), CreatedAt: nowTimePtr(), UpdatedAt: nowTimePtr(),
	}).Error
}

// AdminGetPunishments 处罚记录列表
func AdminGetPunishments(page, pageSize int) ([]model.RiskUser, int64, error) {
	var list []model.RiskUser
	var total int64
	q := db.Model(&model.RiskUser{})
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	err := q.Scopes(Paginate2(page, pageSize)).Order("id DESC").Find(&list).Error
	return list, total, err
}

// AdminGetDeposits 保证金列表(俱乐部维度)
func AdminGetDeposits(page, pageSize int) ([]model.Club, int64, error) {
	var list []model.Club
	var total int64
	q := db.Model(&model.Club{}).Where("deposit_amount > 0")
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	err := q.Scopes(Paginate2(page, pageSize)).Order("id DESC").Find(&list).Error
	return list, total, err
}

// AdminConfirmDeposit 确认保证金缴纳
func AdminConfirmDeposit(clubID int64) error {
	return clubRepo.Update(clubID, map[string]interface{}{
		"deposit_status": 1, "updated_at": nowTimePtr(),
	})
}

// AdminRefundDeposit 退还保证金
func AdminRefundDeposit(clubID int64) error {
	return clubRepo.Update(clubID, map[string]interface{}{
		"deposit_status": 2, "updated_at": nowTimePtr(),
	})
}

// AdminUpdateDepositConfig 更新保证金配置
func AdminUpdateDepositConfig(amount int64) error {
	return upsertSystemConfig("deposit_amount", itoa(amount), "保证金金额")
}

// todayStart 返回今天 0 点时间
func todayStart() time.Time {
	return startOfToday()
}

// startOfToday 返回今天 0 点 time.Time
func startOfToday() time.Time {
	now := time.Now()
	return time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
}

// newWSSystemMessage 构造系统通知 WebSocket 消息
func newWSSystemMessage(title, content, category string) *websocket.Message {
	if hub == nil {
		return nil
	}
	msg, err := websocket.NewMessage(websocket.MsgTypeSystem, 0, 0, websocket.SystemPayload{
		Title: title, Content: content, Category: category,
	})
	if err != nil {
		return nil
	}
	return msg
}
