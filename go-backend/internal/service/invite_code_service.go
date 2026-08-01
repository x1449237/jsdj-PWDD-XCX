package service

import (
	"errors"
	"time"

	"github.com/jisan/e-sports-platform/internal/model"
	"github.com/jisan/e-sports-platform/internal/utils"
)

// GenerateInviteCodeInput 生成邀请码入参
type GenerateInviteCodeInput struct {
	Type         string // club / platform
	ClubID       int64
	Role         string // DS / FXS / 空
	MaxUses      int
	ExpireDays   int
	CreatorID    int64
	CreatorType  string
}

// GenerateInviteCode 生成邀请码(俱乐部/平台)
// 俱乐部码前缀: 缩写_角色标识; 平台码前缀: QPT_
func GenerateInviteCode(in *GenerateInviteCodeInput) (*model.InviteCode, error) {
	if in.MaxUses <= 0 {
		in.MaxUses = 1
	}
	var codeStr string
	var err error
	if in.Type == model.InviteCodeTypePlatform {
		codeStr, err = utils.GeneratePlatformInviteCode()
	} else {
		// 俱乐部码:取俱乐部缩写
		c, cerr := clubRepo.FindByID(in.ClubID)
		if cerr != nil || c == nil {
			return nil, errors.New("俱乐部不存在")
		}
		codeStr, err = utils.GenerateClubInviteCode(c.Abbreviation, in.Role)
	}
	if err != nil {
		return nil, err
	}
	code := &model.InviteCode{
		Code:        codeStr,
		Type:        in.Type,
		ClubID:      in.ClubID,
		Role:        in.Role,
		MaxUses:     in.MaxUses,
		UsedCount:   0,
		Status:      model.InviteCodeStatusUnused,
		CreatorID:   in.CreatorID,
		CreatorType: in.CreatorType,
		CreatedAt:   nowTimePtr(),
		UpdatedAt:   nowTimePtr(),
	}
	if in.ExpireDays > 0 {
		t := time.Now().AddDate(0, 0, in.ExpireDays)
		code.ExpireAt = &t
	}
	if code.Type == "" {
		code.Type = model.InviteCodeTypePlatform
	}
	if err := inviteCodeRepo.Create(code); err != nil {
		return nil, err
	}
	return code, nil
}

// GetInviteCodeList 邀请码列表(支持类型/俱乐部过滤)
func GetInviteCodeList(page, pageSize int, codeType string, clubID int64, status string) ([]model.InviteCode, int64, error) {
	return inviteCodeRepo.List(page, pageSize, codeType, clubID, status)
}

// RevokeInviteCode 撤销邀请码
func RevokeInviteCode(id int64) error {
	return inviteCodeRepo.Revoke(id)
}

// AdminGetInviteCodes 平台邀请码列表
func AdminGetInviteCodes(page, pageSize int, codeType string, status string) ([]model.InviteCode, int64, error) {
	return inviteCodeRepo.List(page, pageSize, codeType, 0, status)
}

// AdminGenerateClubCode 平台生成俱乐部邀请码
func AdminGenerateClubCode(clubID int64, role string, maxUses, expireDays int, adminID int64) (*model.InviteCode, error) {
	return GenerateInviteCode(&GenerateInviteCodeInput{
		Type: model.InviteCodeTypeClub, ClubID: clubID, Role: role,
		MaxUses: maxUses, ExpireDays: expireDays,
		CreatorID: adminID, CreatorType: "admin",
	})
}

// AdminGeneratePlatformCode 平台生成通用邀请码
func AdminGeneratePlatformCode(maxUses, expireDays int, adminID int64) (*model.InviteCode, error) {
	return GenerateInviteCode(&GenerateInviteCodeInput{
		Type: model.InviteCodeTypePlatform,
		MaxUses: maxUses, ExpireDays: expireDays,
		CreatorID: adminID, CreatorType: "admin",
	})
}

// AdminRevokeInviteCode 平台撤销邀请码
func AdminRevokeInviteCode(id int64) error {
	return RevokeInviteCode(id)
}

// AdminExportInviteCodes 导出邀请码(返回全部列表)
func AdminExportInviteCodes(codeType, status string) ([]model.InviteCode, error) {
	list, _, err := inviteCodeRepo.List(1, 10000, codeType, 0, status)
	return list, err
}

// ShopGetInviteCodes 俱乐部邀请码列表
func ShopGetInviteCodes(clubID int64, page, pageSize int) ([]model.InviteCode, int64, error) {
	return inviteCodeRepo.List(page, pageSize, model.InviteCodeTypeClub, clubID, "")
}

// ShopGenerateInviteCode 俱乐部生成邀请码(创始人专属)
func ShopGenerateInviteCode(clubID, creatorID int64, role string, maxUses, expireDays int) (*model.InviteCode, error) {
	return GenerateInviteCode(&GenerateInviteCodeInput{
		Type: model.InviteCodeTypeClub, ClubID: clubID, Role: role,
		MaxUses: maxUses, ExpireDays: expireDays,
		CreatorID: creatorID, CreatorType: "club",
	})
}

// ShopRevokeInviteCode 俱乐部撤销邀请码
func ShopRevokeInviteCode(id, clubID int64) error {
	c, err := inviteCodeRepo.FindByID(id)
	if err != nil {
		return err
	}
	if c == nil {
		return errors.New("邀请码不存在")
	}
	if c.ClubID != clubID {
		return errors.New("无权操作")
	}
	return RevokeInviteCode(id)
}
