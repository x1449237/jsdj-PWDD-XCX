package repository

import (
	"errors"

	"gorm.io/gorm"

	"github.com/jisan/e-sports-platform/internal/model"
)

// ClubRepo 俱乐部数据访问仓储
type ClubRepo struct {
	db *gorm.DB
}

// NewClubRepo 创建俱乐部仓储
func NewClubRepo(db *gorm.DB) *ClubRepo {
	return &ClubRepo{db: db}
}

// FindByID 根据ID查询俱乐部
func (r *ClubRepo) FindByID(id int64) (*model.Club, error) {
	var c model.Club
	if err := r.db.First(&c, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &c, nil
}

// FindByFounder 根据创始人查询俱乐部
func (r *ClubRepo) FindByFounder(uid int64) (*model.Club, error) {
	var c model.Club
	if err := r.db.Where("founder_uid = ?", uid).First(&c).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &c, nil
}

// Create 创建俱乐部
func (r *ClubRepo) Create(c *model.Club) error {
	return r.db.Create(c).Error
}

// Update 更新俱乐部字段
func (r *ClubRepo) Update(id int64, fields map[string]interface{}) error {
	return r.db.Model(&model.Club{}).Where("id = ?", id).Updates(fields).Error
}

// List 分页查询俱乐部
func (r *ClubRepo) List(page, pageSize int, status int8, keyword string) ([]model.Club, int64, error) {
	var clubs []model.Club
	var total int64
	q := r.db.Model(&model.Club{})
	if status >= 0 {
		q = q.Where("status = ?", status)
	}
	if keyword != "" {
		like := "%" + keyword + "%"
		q = q.Where("name LIKE ? OR abbreviation LIKE ?", like, like)
	}
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	err := q.Scopes(Paginate(page, pageSize)).Order("id DESC").Find(&clubs).Error
	return clubs, total, err
}

// CountByAbbr 统计缩写占用数(用于封存校验)
func (r *ClubRepo) CountByAbbr(abbr string) (int64, error) {
	var count int64
	err := r.db.Model(&model.Club{}).Where("abbreviation = ?", abbr).Count(&count).Error
	return count, err
}

// CreateAbbr 封存缩写
func (r *ClubRepo) CreateAbbr(a *model.ClubAbbreviation) error {
	return r.db.Create(a).Error
}

// CreateMember 添加俱乐部成员
func (r *ClubRepo) CreateMember(m *model.ClubMember) error {
	return r.db.Create(m).Error
}

// FindMember 查询俱乐部成员
func (r *ClubRepo) FindMember(clubID, userID int64) (*model.ClubMember, error) {
	var m model.ClubMember
	if err := r.db.Where("club_id = ? AND user_id = ?", clubID, userID).First(&m).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &m, nil
}

// UpdateMember 更新成员
func (r *ClubRepo) UpdateMember(id int64, fields map[string]interface{}) error {
	return r.db.Model(&model.ClubMember{}).Where("id = ?", id).Updates(fields).Error
}

// ListMembers 查询俱乐部成员
func (r *ClubRepo) ListMembers(clubID int64) ([]model.ClubMember, error) {
	var list []model.ClubMember
	err := r.db.Where("club_id = ? AND status = 1", clubID).Find(&list).Error
	return list, err
}

// CreateApplication 创建入会申请
func (r *ClubRepo) CreateApplication(a *model.JoinApplication) error {
	return r.db.Create(a).Error
}

// FindApplication 查询入会申请
func (r *ClubRepo) FindApplication(id int64) (*model.JoinApplication, error) {
	var a model.JoinApplication
	if err := r.db.First(&a, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &a, nil
}

// UpdateApplication 更新入会申请
func (r *ClubRepo) UpdateApplication(id int64, fields map[string]interface{}) error {
	return r.db.Model(&model.JoinApplication{}).Where("id = ?", id).Updates(fields).Error
}

// ListApplications 俱乐部入会申请列表
func (r *ClubRepo) ListApplications(clubID int64, page, pageSize int, status string) ([]model.JoinApplication, int64, error) {
	var list []model.JoinApplication
	var total int64
	q := r.db.Model(&model.JoinApplication{}).Where("club_id = ?", clubID)
	if status != "" {
		q = q.Where("status = ?", status)
	}
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	err := q.Scopes(Paginate(page, pageSize)).Order("id DESC").Find(&list).Error
	return list, total, err
}

// ListUserApplications 用户提交的入会申请
func (r *ClubRepo) ListUserApplications(userID int64, page, pageSize int) ([]model.JoinApplication, int64, error) {
	var list []model.JoinApplication
	var total int64
	q := r.db.Model(&model.JoinApplication{}).Where("user_id = ?", userID)
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	err := q.Scopes(Paginate(page, pageSize)).Order("id DESC").Find(&list).Error
	return list, total, err
}

// CreateExamRecord 创建考核记录
func (r *ClubRepo) CreateExamRecord(e *model.ExamRecord) error {
	return r.db.Create(e).Error
}

// ListPlayers 俱乐部打手列表(通过成员表关联用户)
func (r *ClubRepo) ListPlayers(clubID int64, page, pageSize int) ([]model.User, int64, error) {
	var players []model.User
	var total int64
	q := r.db.Model(&model.User{}).
		Joins("JOIN club_members cm ON cm.user_id = users.id").
		Where("cm.club_id = ? AND cm.status = 1 AND cm.role = ?", clubID, model.ClubMemberRolePlayer)
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	err := q.Scopes(Paginate(page, pageSize)).Order("users.id DESC").Find(&players).Error
	return players, total, err
}

// CreateShopAdmin 创建内置管理端账号
func (r *ClubRepo) CreateShopAdmin(a *model.ShopAdminAccount) error {
	return r.db.Create(a).Error
}

// FindShopAdminByUsername 根据账号查询内置管理端
func (r *ClubRepo) FindShopAdminByUsername(username string) (*model.ShopAdminAccount, error) {
	var a model.ShopAdminAccount
	if err := r.db.Where("username = ?", username).First(&a).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &a, nil
}

// FindShopAdminByID 根据ID查询内置管理端
func (r *ClubRepo) FindShopAdminByID(id int64) (*model.ShopAdminAccount, error) {
	var a model.ShopAdminAccount
	if err := r.db.First(&a, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &a, nil
}

// ListShopAdmins 俱乐部内置管理员列表
func (r *ClubRepo) ListShopAdmins(clubID int64) ([]model.ShopAdminAccount, error) {
	var list []model.ShopAdminAccount
	err := r.db.Where("club_id = ?", clubID).Order("id DESC").Find(&list).Error
	return list, err
}

// UpdateShopAdmin 更新内置管理端账号
func (r *ClubRepo) UpdateShopAdmin(id int64, fields map[string]interface{}) error {
	return r.db.Model(&model.ShopAdminAccount{}).Where("id = ?", id).Updates(fields).Error
}

// DeleteShopAdmin 删除内置管理端账号
func (r *ClubRepo) DeleteShopAdmin(id int64) error {
	return r.db.Delete(&model.ShopAdminAccount{}, id).Error
}

// CreateRegistration 创建企业入驻申请
func (r *ClubRepo) CreateRegistration(reg interface{}) error {
	return r.db.Create(reg).Error
}

// FindRegistrationByClub 根据俱乐部查询入驻申请(企业/个人通用)
func (r *ClubRepo) FindRegistrationByClub(table string, clubID int64) (interface{}, error) {
	switch table {
	case "enterprise":
		var reg model.EnterpriseRegistration
		if err := r.db.Where("club_id = ?", clubID).Order("id DESC").First(&reg).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return nil, nil
			}
			return nil, err
		}
		return &reg, nil
	default:
		var reg model.PersonalRegistration
		if err := r.db.Where("club_id = ?", clubID).Order("id DESC").First(&reg).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return nil, nil
			}
			return nil, err
		}
		return &reg, nil
	}
}
