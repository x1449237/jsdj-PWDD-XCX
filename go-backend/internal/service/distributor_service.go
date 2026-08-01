package service

import (
	"github.com/jisan/e-sports-platform/internal/model"
)

// GetSubordinates 分销商下级列表(一级/二级)
func GetSubordinates(distributorID int64, level int8) ([]model.DistributorRelation, error) {
	var list []model.DistributorRelation
	q := db.Model(&model.DistributorRelation{}).Where("superior_id = ? AND is_valid = 1", distributorID)
	if level > 0 {
		q = q.Where("level = ?", level)
	}
	err := q.Order("id DESC").Find(&list).Error
	return list, err
}

// GetCommissionList 分销佣金记录
func GetCommissionList(distributorID int64, page, pageSize int) ([]model.DistributorCommission, int64, error) {
	var list []model.DistributorCommission
	var total int64
	q := db.Model(&model.DistributorCommission{}).Where("distributor_id = ?", distributorID)
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	err := q.Scopes(Paginate2(page, pageSize)).Order("id DESC").Find(&list).Error
	return list, total, err
}

// GetDistributorRanking 分销商排行榜(按已结算佣金)
func GetDistributorRanking(limit int) ([]map[string]interface{}, error) {
	if limit <= 0 || limit > 100 {
		limit = 50
	}
	var rows []struct {
		DistributorID int64 `gorm:"column:distributor_id"`
		Total         int64 `gorm:"column:total"`
	}
	err := db.Model(&model.DistributorCommission{}).
		Select("distributor_id, COALESCE(SUM(amount),0) AS total").
		Where("status = ?", model.DistributorCommissionStatusSettled).
		Group("distributor_id").Order("total DESC").Limit(limit).Scan(&rows).Error
	if err != nil {
		return nil, err
	}
	result := make([]map[string]interface{}, 0, len(rows))
	for _, r := range rows {
		u, _ := userRepo.FindByID(r.DistributorID)
		name := ""
		if u != nil {
			name = u.Nickname
		}
		result = append(result, map[string]interface{}{
			"distributor_id": r.DistributorID,
			"nickname":       name,
			"total":          r.Total,
		})
	}
	return result, nil
}

// BindDistributorRelation 建立分销关系(供邀请码绑定调用)
// superiorID 上级分销商，subordinateID 下级用户，level 级别
func BindDistributorRelation(superiorID, subordinateID int64, level int8) error {
	rel := &model.DistributorRelation{
		SuperiorID:    superiorID,
		SubordinateID: subordinateID,
		Level:         level,
		IsValid:       1,
		CreatedAt:     nowTimePtr(),
		UpdatedAt:     nowTimePtr(),
	}
	return db.Create(rel).Error
}

// AuditDistributors 平台分销商审核列表(角色位含分销商的用户)
func AuditDistributors(page, pageSize int, keyword string) ([]model.User, int64, error) {
	return userRepo.List(page, pageSize, -1, model.RoleDistributor, keyword)
}

// ApproveDistributor 审核通过分销商(默认已具备角色,占位)
func ApproveDistributor(distributorID int64) error {
	return userRepo.Update(distributorID, map[string]interface{}{
		"updated_at": nowTimePtr(),
	})
}
