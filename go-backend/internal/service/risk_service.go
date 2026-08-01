package service

import (
	"github.com/jisan/e-sports-platform/internal/model"
	"github.com/jisan/e-sports-platform/pkg/queue"
)

// MarkRiskUser 标记风险用户
func MarkRiskUser(userID int64, riskLevel, riskType string) error {
	// 防重:同用户同类型更新
	var ru model.RiskUser
	if err := db.Where("user_id = ? AND risk_type = ?", userID, riskType).First(&ru).Error; err == nil {
		return db.Model(&ru).Updates(map[string]interface{}{
			"risk_level": riskLevel, "marked_at": nowTimePtr(), "updated_at": nowTimePtr(),
		}).Error
	}
	return db.Create(&model.RiskUser{
		UserID: userID, RiskLevel: riskLevel, RiskType: riskType,
		MarkedAt: nowTimePtr(), CreatedAt: nowTimePtr(), UpdatedAt: nowTimePtr(),
	}).Error
}

// AdminGetRiskUsers 平台风险用户列表
func AdminGetRiskUsers(page, pageSize int, level string) ([]model.RiskUser, int64, error) {
	var list []model.RiskUser
	var total int64
	q := db.Model(&model.RiskUser{})
	if level != "" {
		q = q.Where("risk_level = ?", level)
	}
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	err := q.Scopes(Paginate2(page, pageSize)).Order("id DESC").Find(&list).Error
	return list, total, err
}

// AdminGetRiskAlerts AI 风险预警列表
func AdminGetRiskAlerts(page, pageSize int, alertType string) ([]model.AiRiskAlert, int64, error) {
	var list []model.AiRiskAlert
	var total int64
	q := db.Model(&model.AiRiskAlert{})
	if alertType != "" {
		q = q.Where("alert_type = ?", alertType)
	}
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	err := q.Scopes(Paginate2(page, pageSize)).Order("id DESC").Find(&list).Error
	return list, total, err
}

// AdminHandleRiskAlert 处理 AI 风险预警
// status: 1已处理 2已忽略
func AdminHandleRiskAlert(alertID, adminID int64, status int8, result string) error {
	return db.Model(&model.AiRiskAlert{}).Where("id = ?", alertID).
		Updates(map[string]interface{}{
			"status":     status,
			"updated_at": nowTimePtr(),
		}).Error
}

// GetRiskUsersBatch 批量查询风险用户ID集合
func GetRiskUsersBatch(userIDs []int64) ([]model.RiskUser, error) {
	var list []model.RiskUser
	if len(userIDs) == 0 {
		return list, nil
	}
	err := db.Where("user_id IN ?", userIDs).Find(&list).Error
	return list, err
}

// EnqueueAIScan 投递 AI 风险扫描任务
func EnqueueAIScan(targetType string, targetID int64, content string) {
	if queueC == nil {
		return
	}
	ctx, cancel := contextWithTimeout()
	defer cancel()
	_ = queueC.EnqueueAIScan(ctx, queue.AIScanPayload{
		TargetType: targetType, TargetID: targetID, Content: content,
	})
}
