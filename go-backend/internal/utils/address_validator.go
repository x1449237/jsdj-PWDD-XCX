package utils

import (
	"errors"
	"strings"
)

// AddressInfo 居住地址信息(个人入驻必填字段)
type AddressInfo struct {
	Province    string `json:"province"`    // 省份
	City        string `json:"city"`        // 城市
	District    string `json:"district"`    // 区/县
	Street      string `json:"street"`      // 街道/乡镇
	Community   string `json:"community"`   // 社区/村
	Building    string `json:"building"`    // 楼栋
	HouseNumber string `json:"house_number"` // 门牌号
}

// AddressMaxLength 单个地址字段最大长度
const AddressMaxLength = 128

// ValidateAddress 校验居住地址合法性(简化:非空 + 长度校验)
// 任意一项为空返回错误 "地址信息不完整"
func ValidateAddress(addr AddressInfo) error {
	if strings.TrimSpace(addr.Province) == "" ||
		strings.TrimSpace(addr.City) == "" ||
		strings.TrimSpace(addr.District) == "" ||
		strings.TrimSpace(addr.Street) == "" ||
		strings.TrimSpace(addr.Community) == "" ||
		strings.TrimSpace(addr.Building) == "" ||
		strings.TrimSpace(addr.HouseNumber) == "" {
		return errors.New("地址信息不完整")
	}
	// 长度校验
	for _, v := range []string{addr.Province, addr.City, addr.District, addr.Street, addr.Community, addr.Building, addr.HouseNumber} {
		if len([]rune(v)) > AddressMaxLength {
			return errors.New("地址字段长度超限")
		}
	}
	return nil
}
