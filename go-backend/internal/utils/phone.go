package utils

import (
	"errors"
	"strings"
)

// MaskPhone 手机号脱敏: 138****1234
func MaskPhone(phone string) string {
	if len(phone) < 7 {
		// 长度不足，仅保留首尾各1位
		if len(phone) <= 2 {
			return strings.Repeat("*", len(phone))
		}
		return string(phone[0]) + strings.Repeat("*", len(phone)-2) + string(phone[len(phone)-1])
	}
	return phone[:3] + strings.Repeat("*", 4) + phone[len(phone)-4:]
}

// MaskIDCard 身份证号脱敏: 110***********1234
func MaskIDCard(idCard string) string {
	if len(idCard) < 10 {
		return strings.Repeat("*", len(idCard))
	}
	return idCard[:3] + strings.Repeat("*", len(idCard)-7) + idCard[len(idCard)-4:]
}

// MaskName 真实姓名脱敏: 张* / 欧阳**
func MaskName(name string) string {
	runes := []rune(name)
	if len(runes) <= 1 {
		return name
	}
	if len(runes) == 2 {
		return string(runes[0]) + "*"
	}
	// 三字及以上: 保留首尾，中间脱敏
	return string(runes[0]) + strings.Repeat("*", len(runes)-2) + string(runes[len(runes)-1])
}

// MaskEmail 邮箱脱敏: z***@example.com
func MaskEmail(email string) string {
	at := strings.Index(email, "@")
	if at <= 0 {
		return email
	}
	name := email[:at]
	domain := email[at:]
	if len(name) <= 1 {
		return name + "*" + domain
	}
	return string(name[0]) + strings.Repeat("*", len(name)-1) + domain
}

// MaskBankCard 银行卡号脱敏: 6222***********1234
func MaskBankCard(card string) string {
	if len(card) < 8 {
		return strings.Repeat("*", len(card))
	}
	return card[:4] + strings.Repeat("*", len(card)-8) + card[len(card)-4:]
}

// MaskBankAccount 对公账户脱敏: 保留前4后4，中间以 **** 替换
// 长度不足 9 位时退化为 MaskBankCard 行为(保留前4后4)
func MaskBankAccount(account string) string {
	if len(account) <= 8 {
		// 长度不足，按银行卡脱敏处理
		return MaskBankCard(account)
	}
	return account[:4] + "****" + account[len(account)-4:]
}

// MaskMiddle 通用中间脱敏: 保留前后各 keep 位，中间以 * 替换
func MaskMiddle(s string, keep int) (string, error) {
	if keep < 0 {
		return "", errors.New("keep 不能为负数")
	}
	if len(s) <= keep*2 {
		return strings.Repeat("*", len(s)), nil
	}
	return s[:keep] + strings.Repeat("*", len(s)-keep*2) + s[len(s)-keep:], nil
}
