package service

import (
	"errors"

	"github.com/jisan/e-sports-platform/internal/model"
)

// CreateArbitrationCase 创建仲裁案件(由售后介入升级)
func CreateArbitrationCase(orderID, sessionID int64) (*model.ArbitrationCase, error) {
	// 防重
	var cnt int64
	_ = db.Model(&model.ArbitrationCase{}).Where("order_id = ? AND status != ?", orderID, model.ArbitrationStatusClosed).Count(&cnt).Error
	if cnt > 0 {
		return nil, errors.New("该订单已有进行中的仲裁案件")
	}
	c := &model.ArbitrationCase{
		OrderID:   orderID,
		SessionID: sessionID,
		Status:    model.ArbitrationStatusPending,
		CreatedAt: nowTimePtr(),
		UpdatedAt: nowTimePtr(),
	}
	if err := db.Create(c).Error; err != nil {
		return nil, err
	}
	return c, nil
}

// AdminGetArbitrationCases 仲裁案件列表
func AdminGetArbitrationCases(page, pageSize int, status string) ([]model.ArbitrationCase, int64, error) {
	var list []model.ArbitrationCase
	var total int64
	q := db.Model(&model.ArbitrationCase{})
	if status != "" {
		q = q.Where("status = ?", status)
	}
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	err := q.Scopes(Paginate2(page, pageSize)).Order("id DESC").Find(&list).Error
	return list, total, err
}

// AdminGetArbitrationCaseDetail 仲裁案件详情(含证据)
func AdminGetArbitrationCaseDetail(caseID int64) (map[string]interface{}, error) {
	var c model.ArbitrationCase
	if err := db.First(&c, caseID).Error; err != nil {
		return nil, errors.New("案件不存在")
	}
	var evidences []model.ArbitrationEvidence
	_ = db.Where("case_id = ?", caseID).Find(&evidences).Error
	return map[string]interface{}{
		"case":      c,
		"evidences": evidences,
	}, nil
}

// AdminJudgeArbitration 仲裁判决
func AdminJudgeArbitration(caseID, arbitratorID int64, result string) error {
	return db.Model(&model.ArbitrationCase{}).Where("id = ?", caseID).
		Updates(map[string]interface{}{
			"status":        model.ArbitrationStatusClosed,
			"arbitrator_id": arbitratorID,
			"result":        result,
			"updated_at":    nowTimePtr(),
		}).Error
}

// AdminGetArbitrationRules 判责规则列表
func AdminGetArbitrationRules(page, pageSize int) ([]model.ArbitrationRule, int64, error) {
	var list []model.ArbitrationRule
	var total int64
	q := db.Model(&model.ArbitrationRule{})
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	err := q.Scopes(Paginate2(page, pageSize)).Order("id DESC").Find(&list).Error
	return list, total, err
}

// AdminAddArbitrationRule 新增判责规则
func AdminAddArbitrationRule(rule *model.ArbitrationRule) error {
	rule.CreatedAt = nowTimePtr()
	rule.UpdatedAt = nowTimePtr()
	return db.Create(rule).Error
}

// AdminGetEvidenceTemplates 证据模板列表(占位:复用 ArbitrationRule 表)
func AdminGetEvidenceTemplates(page, pageSize int) ([]model.ArbitrationRule, int64, error) {
	return AdminGetArbitrationRules(page, pageSize)
}

// AdminAddEvidenceTemplate 新增证据模板
func AdminAddEvidenceTemplate(t *model.ArbitrationRule) error {
	return AdminAddArbitrationRule(t)
}
