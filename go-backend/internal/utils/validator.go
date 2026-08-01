package utils

import (
	"regexp"
	"strconv"
	"strings"
)

// 预编译正则
var (
	mobileRegex   = regexp.MustCompile(`^1[3-9]\d{9}$`)
	emailRegex    = regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
	orderNoRegex  = regexp.MustCompile(`^[A-Za-z0-9]{6,64}$`)
	usernameRegex = regexp.MustCompile(`^[a-zA-Z0-9_]{3,32}$`)
	urlRegex      = regexp.MustCompile(`^https?://[^\s]+$`)
)

// ValidateMobile 校验中国大陆手机号
func ValidateMobile(mobile string) bool {
	return mobileRegex.MatchString(mobile)
}

// ValidateEmail 校验邮箱格式
func ValidateEmail(email string) bool {
	return emailRegex.MatchString(email)
}

// ValidateOrderNo 校验订单号格式(6-64位字母数字)
func ValidateOrderNo(orderNo string) bool {
	return orderNoRegex.MatchString(orderNo)
}

// ValidateUsername 校验用户名(3-32位字母数字下划线)
func ValidateUsername(username string) bool {
	return usernameRegex.MatchString(username)
}

// ValidateURL 校验 URL 格式
func ValidateURL(u string) bool {
	return urlRegex.MatchString(u)
}

// ValidateAmount 校验金额(分)，必须大于 0
func ValidateAmount(amount int64) bool {
	return amount > 0
}

// ValidateIDCardFormat 校验身份证号基础格式(15或18位)
func ValidateIDCardFormat(idCard string) bool {
	idCard = strings.ToUpper(strings.TrimSpace(idCard))
	if len(idCard) == 18 {
		match, _ := regexp.MatchString(`^\d{17}[\dX]$`, idCard)
		return match
	}
	if len(idCard) == 15 {
		match, _ := regexp.MatchString(`^\d{15}$`, idCard)
		return match
	}
	return false
}

// ValidateScore 校验评分范围(1-5)
func ValidateScore(score int) bool {
	return score >= 1 && score <= 5
}

// ValidatePage 校验并修正分页参数
// page 从1开始，pageSize 默认10，最大100
func ValidatePage(page, pageSize int) (int, int) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 10
	}
	if pageSize > 100 {
		pageSize = 100
	}
	return page, pageSize
}

// Offset 根据页码与每页条数计算偏移量
func Offset(page, pageSize int) int {
	return (page - 1) * pageSize
}

// ParsePageInt 从字符串解析分页参数
func ParsePageInt(s string, def int) int {
	if s == "" {
		return def
	}
	n, err := strconv.Atoi(s)
	if err != nil || n < 1 {
		return def
	}
	return n
}
