package utils

import (
	"errors"
	"strconv"
	"time"
)

// 身份证号权重与校验码映射(18位身份证校验算法)
var (
	idCardWeight  = [17]int{7, 9, 10, 5, 8, 4, 2, 1, 6, 3, 7, 9, 10, 5, 8, 4, 2}
	idCardCheckMap = map[int]string{
		0: "1", 1: "0", 2: "X", 3: "9", 4: "8",
		5: "7", 6: "6", 7: "5", 8: "4", 9: "3", 10: "2",
	}
)

// ValidateIDCard 严格校验 18 位身份证号(含校验位校验)
func ValidateIDCard(idCard string) bool {
	if len(idCard) != 18 {
		return false
	}
	// 前 17 位必须为数字
	for i := 0; i < 17; i++ {
		if idCard[i] < '0' || idCard[i] > '9' {
			return false
		}
	}
	// 第 18 位允许 X
	last := idCard[17]
	if !((last >= '0' && last <= '9') || last == 'X' || last == 'x') {
		return false
	}

	// 计算校验位
	sum := 0
	for i := 0; i < 17; i++ {
		n, err := strconv.Atoi(string(idCard[i]))
		if err != nil {
			return false
		}
		sum += n * idCardWeight[i]
	}
	expected := idCardCheckMap[sum%11]

	// 将输入的末位统一转大写比较
	lastStr := string(last)
	if last == 'x' {
		lastStr = "X"
	}
	return lastStr == expected
}

// GetBirthdayFromIDCard 从 18 位身份证号提取生日
func GetBirthdayFromIDCard(idCard string) (time.Time, error) {
	if len(idCard) != 18 {
		return time.Time{}, errors.New("身份证号长度必须为 18 位")
	}
	birthdayStr := idCard[6:14] // YYYYMMDD
	birthday, err := time.Parse("20060102", birthdayStr)
	if err != nil {
		return time.Time{}, errors.New("身份证号中生日格式无效")
	}
	return birthday, nil
}

// GetGenderFromIDCard 从 18 位身份证号获取性别 1=男 2=女
func GetGenderFromIDCard(idCard string) (int, error) {
	if len(idCard) != 18 {
		return 0, errors.New("身份证号长度必须为 18 位")
	}
	genderDigit, err := strconv.Atoi(string(idCard[16]))
	if err != nil {
		return 0, errors.New("身份证号格式无效")
	}
	if genderDigit%2 == 1 {
		return 1, nil // 男
	}
	return 2, nil // 女
}

// GetAgeFromIDCard 根据 18 位身份证号计算周岁年龄
func GetAgeFromIDCard(idCard string) (int, error) {
	birthday, err := GetBirthdayFromIDCard(idCard)
	if err != nil {
		return 0, err
	}
	return CalcAge(birthday), nil
}

// CalcAge 根据生日计算周岁年龄
func CalcAge(birthday time.Time) int {
	now := time.Now()
	age := now.Year() - birthday.Year()
	// 未到生日则减一
	if now.Month() < birthday.Month() || (now.Month() == birthday.Month() && now.Day() < birthday.Day()) {
		age--
	}
	if age < 0 {
		age = 0
	}
	return age
}

// IsMinorByAge 判断是否未成年(年龄 < 18)
func IsMinorByAge(age int) bool {
	return age < 18
}

// IsMinorByIDCard 根据身份证号判断是否未成年
func IsMinorByIDCard(idCard string) (bool, error) {
	age, err := GetAgeFromIDCard(idCard)
	if err != nil {
		return false, err
	}
	return IsMinorByAge(age), nil
}
