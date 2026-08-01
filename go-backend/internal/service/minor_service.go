package service

import (
	"github.com/jisan/e-sports-platform/internal/model"
)

// GetMinorCurfewLogs 未成年宵禁拦截日志
func GetMinorCurfewLogs(userID int64, page, pageSize int) ([]model.MinorCurfewLog, int64, error) {
	var list []model.MinorCurfewLog
	var total int64
	q := db.Model(&model.MinorCurfewLog{})
	if userID > 0 {
		q = q.Where("user_id = ?", userID)
	}
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	err := q.Scopes(Paginate2(page, pageSize)).Order("id DESC").Find(&list).Error
	return list, total, err
}

// ListMinors 未成年用户列表(供平台监管)
func ListMinors(page, pageSize int) ([]model.User, int64, error) {
	var list []model.User
	var total int64
	q := db.Model(&model.User{}).Where("is_minor = 1")
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	err := q.Scopes(Paginate2(page, pageSize)).Order("id DESC").Find(&list).Error
	return list, total, err
}
