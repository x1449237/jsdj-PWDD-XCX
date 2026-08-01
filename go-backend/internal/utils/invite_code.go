package utils

import (
	"crypto/rand"
	"strings"
)

// 邀请码相关常量
const (
	// InviteCodePlatformPrefix 平台通用邀请码前缀
	InviteCodePlatformPrefix = "QPT"
	// InviteCodeRandLen 随机字符长度
	InviteCodeRandLen = 8
	// InviteCodeCharSet 随机字符集(去除易混淆字符)
	InviteCodeCharSet = "ABCDEFGHJKLMNPQRSTUVWXYZ23456789"
)

// 角色标识常量
const (
	InviteRolePlayer     = "DS"  // 打手
	InviteRoleDistributor = "FXS" // 分销商
)

// GeneratePlatformInviteCode 生成平台通用邀请码: QPT_ + 随机8位字符
func GeneratePlatformInviteCode() (string, error) {
	randStr, err := randomString(InviteCodeRandLen)
	if err != nil {
		return "", err
	}
	return InviteCodePlatformPrefix + "_" + randStr, nil
}

// GenerateClubInviteCode 生成俱乐部邀请码: 缩写_角色标识 + 随机8位字符
// abbreviation 俱乐部缩写，roleFlag 角色标识(DS/FXS，可为空表示不限定角色)
func GenerateClubInviteCode(abbreviation, roleFlag string) (string, error) {
	randStr, err := randomString(InviteCodeRandLen)
	if err != nil {
		return "", err
	}
	abbr := strings.ToUpper(strings.TrimSpace(abbreviation))
	prefix := abbr
	if roleFlag != "" {
		prefix = abbr + "_" + strings.ToUpper(roleFlag)
	}
	return prefix + randStr, nil
}

// GenerateInviteCode 通用邀请码生成入口
// inviteType: platform / club
func GenerateInviteCode(inviteType, abbreviation, roleFlag string) (string, error) {
	if inviteType == "platform" {
		return GeneratePlatformInviteCode()
	}
	return GenerateClubInviteCode(abbreviation, roleFlag)
}

// randomString 生成指定长度的随机字符串
func randomString(length int) (string, error) {
	bytes := make([]byte, length)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	for i := range bytes {
		bytes[i] = InviteCodeCharSet[int(bytes[i])%len(InviteCodeCharSet)]
	}
	return string(bytes), nil
}
