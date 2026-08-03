package service

import (
	"crypto/sha256"
	"fmt"
	"math"
	"regexp"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"
)

// ============================================================
// 小程序前端零逻辑配套：所有业务计算/校验/格式化全部后端完成
// 前端只渲染 API 返回的 _text/_color/_display 等字段
// ============================================================

// ---------- 金额格式化（前端禁止 fenToYuan/yuanToFen） ----------

// FormatFenToYuan 分→元字符串(保留2位小数,含千分位)
// 例: 1250 -> "12.50", 100000 -> "1,000.00"
func FormatFenToYuan(fen int64) string {
	if fen < 0 {
		return "-" + FormatFenToYuan(-fen)
	}
	yuan := float64(fen) / 100.0
	str := fmt.Sprintf("%.2f", yuan)
	// 加千分位
	parts := strings.SplitN(str, ".", 2)
	intPart := parts[0]
	if len(intPart) > 3 {
		var b strings.Builder
		n := len(intPart)
		for i, c := range intPart {
			if i > 0 && (n-i)%3 == 0 {
				b.WriteString(",")
			}
			b.WriteRune(c)
		}
		intPart = b.String()
	}
	if len(parts) == 2 {
		return intPart + "." + parts[1]
	}
	return intPart
}

// ParseYuanToFen 元字符串→分(权威换算,前端提交元后端转分)
func ParseYuanToFen(yuanStr string) (int64, error) {
	yuan, err := strconv.ParseFloat(strings.TrimSpace(yuanStr), 64)
	if err != nil {
		return 0, fmt.Errorf("金额格式错误")
	}
	if yuan < 0 {
		return 0, fmt.Errorf("金额不能为负数")
	}
	fen := int64(math.Round(yuan * 100))
	return fen, nil
}

// ---------- 订单状态文案/颜色映射（前端禁止 getOrderStatusText） ----------

// OrderStatusRender 订单状态渲染信息
type OrderStatusRender struct {
	StatusText  string `json:"status_text"`
	StatusColor string `json:"status_color"`
	StatusDesc  string `json:"status_desc"`
}

var orderStatusMap = map[int8]OrderStatusRender{
	0: {"待接单", "#FF9900", "等待打手接单"},
	1: {"服务中", "#00AAFF", "打手正在服务中"},
	2: {"待确认完成", "#FF6600", "请确认服务是否完成"},
	3: {"已完结", "#00CC66", "订单已完成"},
	4: {"售后中", "#FF3333", "订单处于售后申诉中"},
	5: {"已取消", "#999999", "订单已取消"},
}

// GetOrderStatusRender 根据状态码返回渲染信息
func GetOrderStatusRender(status int8) OrderStatusRender {
	if r, ok := orderStatusMap[status]; ok {
		return r
	}
	return OrderStatusRender{"未知", "#999999", ""}
}

// ---------- 时间格式化（前端禁止 formatTime/formatRelativeTime） ----------

// FormatTime 时间→标准格式字符串
func FormatTime(t *time.Time, layout string) string {
	if t == nil || t.IsZero() {
		return ""
	}
	if layout == "" {
		layout = "2006-01-02 15:04:05"
	}
	return t.In(time.Local).Format(layout)
}

// FormatRelativeTime 时间→相对时间文案(刚刚/X分钟前/X小时前/MM-DD HH:mm)
func FormatRelativeTime(t *time.Time) string {
	if t == nil || t.IsZero() {
		return ""
	}
	now := time.Now()
	diff := now.Sub(*t)
	if diff < time.Minute {
		return "刚刚"
	}
	if diff < time.Hour {
		return fmt.Sprintf("%d分钟前", int(diff.Minutes()))
	}
	if diff < 24*time.Hour {
		return fmt.Sprintf("%d小时前", int(diff.Hours()))
	}
	if diff < 48*time.Hour {
		return "昨天"
	}
	if diff < 7*24*time.Hour {
		weekdays := []string{"周日", "周一", "周二", "周三", "周四", "周五", "周六"}
		return weekdays[t.Weekday()]
	}
	return t.In(time.Local).Format("01-02 15:04")
}

// FormatChatTime 聊天时间展示规则(需求136:当天只显示时分;跨天显示月日+时分)
func FormatChatTime(t *time.Time) string {
	if t == nil || t.IsZero() {
		return ""
	}
	now := time.Now()
	if t.In(time.Local).Format("2006-01-02") == now.Format("2006-01-02") {
		return t.In(time.Local).Format("15:04")
	}
	return t.In(time.Local).Format("01-02 15:04")
}

// ---------- 聊天消息预览格式化（前端禁止 formatLastMessage） ----------

// FormatLastMessagePreview 根据消息类型生成会话列表预览文案
func FormatLastMessagePreview(msgType, content string) string {
	switch msgType {
	case "text":
		return content
	case "image":
		return "[图片]"
	case "voice":
		return "[语音]"
	case "video":
		return "[视频]"
	case "file":
		return "[文件]"
	case "system":
		return content
	case "order_card":
		return "[订单卡片]"
	case "transfer_card":
		return "[转单卡片]"
	case "recall":
		return "消息已撤回"
	default:
		return content
	}
}

// FormatFileSize 字节数→B/KB/MB 显示文案
func FormatFileSize(bytes int64) string {
	if bytes < 1024 {
		return fmt.Sprintf("%dB", bytes)
	}
	if bytes < 1024*1024 {
		return fmt.Sprintf("%.1fKB", float64(bytes)/1024)
	}
	return fmt.Sprintf("%.1fMB", float64(bytes)/(1024*1024))
}

// ---------- 身份证校验 + 年龄计算（前端禁止 validateIdCard/calcAgeFromIdCard） ----------

var idCardRegex = regexp.MustCompile(`^[1-9]\d{5}(18|19|20)\d{2}(0[1-9]|1[0-2])(0[1-9]|[12]\d|3[01])\d{3}[\dXx]$`)

// IDCardValidateResult 身份证校验结果
type IDCardValidateResult struct {
	Valid     bool   `json:"valid"`
	Age       int    `json:"age"`
	AgeText   string `json:"age_text"`    // 后端下发的年龄文案(如"24岁 / 已成年")
	IsMinor   bool   `json:"is_minor"`    // 未成年(<18)
	IsUnder16 bool   `json:"is_under_16"` // 未满16周岁(入驻门槛)
	Gender    string `json:"gender"`      // male/female
	BirthDate string `json:"birth_date"`  // YYYY-MM-DD
	Message   string `json:"message"`
}

// ValidateIDCard 身份证完整校验(格式+校验位算法+年龄计算)
func ValidateIDCard(idCard string) IDCardValidateResult {
	idCard = strings.TrimSpace(strings.ToUpper(idCard))
	if len(idCard) != 18 {
		return IDCardValidateResult{Valid: false, Message: "身份证号长度必须为18位"}
	}
	if !idCardRegex.MatchString(idCard) {
		return IDCardValidateResult{Valid: false, Message: "身份证号格式不正确"}
	}
	// 校验位算法
	weights := []int{7, 9, 10, 5, 8, 4, 2, 1, 6, 3, 7, 9, 10, 5, 8, 4, 2}
	checkCodes := []string{"1", "0", "X", "9", "8", "7", "6", "5", "4", "3", "2"}
	sum := 0
	for i := 0; i < 17; i++ {
		n, _ := strconv.Atoi(string(idCard[i]))
		sum += n * weights[i]
	}
	if checkCodes[sum%11] != string(idCard[17]) {
		return IDCardValidateResult{Valid: false, Message: "身份证校验位不正确"}
	}
	// 解析出生日期与年龄
	year, _ := strconv.Atoi(idCard[6:10])
	month, _ := strconv.Atoi(idCard[10:12])
	day, _ := strconv.Atoi(idCard[12:14])
	now := time.Now()
	age := now.Year() - year
	if int(now.Month()) < month || (int(now.Month()) == month && now.Day() < day) {
		age--
	}
	// 性别(第17位奇=男 偶=女)
	gender := "male"
	if (int(idCard[16]-'0'))%2 == 0 {
		gender = "female"
	}
	return IDCardValidateResult{
		Valid:     true,
		Age:       age,
		AgeText:   buildAgeText(age),
		IsMinor:   age < 18,
		IsUnder16: age < 16,
		Gender:    gender,
		BirthDate: fmt.Sprintf("%04d-%02d-%02d", year, month, day),
		Message:   "校验通过",
	}
}

// buildAgeText 拼接年龄文案(后端权威,前端禁止计算)
// 例: "24岁 / 已成年" / "15岁 / 未满16周岁,不可入驻"
func buildAgeText(age int) string {
	if age < 16 {
		return fmt.Sprintf("%d岁 / 未满16周岁,不可入驻", age)
	}
	if age < 18 {
		return fmt.Sprintf("%d岁 / 未成年(16-17岁)", age)
	}
	return fmt.Sprintf("%d岁 / 已成年", age)
}

// ---------- 手机号校验 ----------

var phoneRegex = regexp.MustCompile(`^1[3-9]\d{9}$`)

// ValidatePhone 手机号校验
func ValidatePhone(phone string) bool {
	return phoneRegex.MatchString(strings.TrimSpace(phone))
}

// ---------- 密码强度计算（前端禁止 validatePassword） ----------

// PasswordStrengthResult 密码强度结果
type PasswordStrengthResult struct {
	Valid    bool   `json:"valid"`
	Level    string `json:"level"`     // weak/medium/strong
	Score    int    `json:"score"`     // 0-3
	Message  string `json:"message"`
}

// CalcPasswordStrength 密码策略校验+强度计算
func CalcPasswordStrength(password string) PasswordStrengthResult {
	if utf8.RuneCountInString(password) < 6 || utf8.RuneCountInString(password) > 20 {
		return PasswordStrengthResult{Valid: false, Level: "weak", Message: "密码长度须6-20位"}
	}
	score := 0
	hasDigit, hasLetter, hasSpecial := false, false, false
	for _, c := range password {
		switch {
		case c >= '0' && c <= '9':
			hasDigit = true
		case (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z'):
			hasLetter = true
		case c == '!' || c == '@' || c == '#' || c == '$' || c == '%' || c == '^' || c == '&' || c == '*':
			hasSpecial = true
		}
	}
	if hasDigit {score++}
	if hasLetter {score++}
	if hasSpecial {score++}
	level := "weak"
	msg := "密码强度: 弱"
	if score >= 2 {level = "medium"; msg = "密码强度: 中"}
	if score >= 3 {level = "strong"; msg = "密码强度: 强"}
	return PasswordStrengthResult{Valid: true, Level: level, Score: score, Message: msg}
}

// ---------- 金额校验 ----------

// ValidateAmount 金额校验(>0, 上限99999.99, 最多2位小数)
func ValidateAmount(yuanStr string) error {
	yuan, err := strconv.ParseFloat(strings.TrimSpace(yuanStr), 64)
	if err != nil {
		return fmt.Errorf("金额格式错误")
	}
	if yuan <= 0 {
		return fmt.Errorf("金额必须大于0")
	}
	if yuan > 99999.99 {
		return fmt.Errorf("金额不能超过99999.99")
	}
	// 检查小数位
	parts := strings.SplitN(yuanStr, ".", 2)
	if len(parts) == 2 && len(parts[1]) > 2 {
		return fmt.Errorf("金额最多2位小数")
	}
	return nil
}

// ---------- 脱敏（前端禁止 maskPhone/maskIdCard/maskName） ----------

// MaskPhone 手机号脱敏: 138****1234
func MaskPhone(phone string) string {
	if len(phone) < 7 {return phone}
	return phone[:3] + "****" + phone[len(phone)-4:]
}

// MaskIDCard 身份证脱敏: 110101********1234
func MaskIDCard(idCard string) string {
	if len(idCard) < 10 {return idCard}
	return idCard[:6] + strings.Repeat("*", len(idCard)-10) + idCard[len(idCard)-4:]
}

// MaskName 姓名脱敏: 张*三
func MaskName(name string) string {
	runes := []rune(name)
	n := len(runes)
	if n <= 1 {return name}
	if n == 2 {return string(runes[0]) + "*"}
	return string(runes[0]) + strings.Repeat("*", n-2) + string(runes[n-1])
}

// MaskBankCard 银行卡/对公账户脱敏: 前4后4中间****
func MaskBankAccount(acc string) string {
	if len(acc) <= 8 {return acc}
	return acc[:4] + strings.Repeat("*", len(acc)-8) + acc[len(acc)-4:]
}

// ---------- 撤回时效校验（前端禁止判断 recallTimeLimit） ----------

// CanRecallMessage 撤回时效校验(5分钟内可撤回)
func CanRecallMessage(msgTime time.Time, limitSeconds int) bool {
	if limitSeconds <= 0 {limitSeconds = 300}
	return time.Since(msgTime) <= time.Duration(limitSeconds)*time.Second
}

// ---------- 未成年人宵禁校验（前端禁止判断时间） ----------

// MinorCurfewCheck 未成年人宵禁校验(22:00~08:00 禁止下单/打赏)
func MinorCurfewCheck(isMinor bool) (blocked bool, message string) {
	if !isMinor {return false, ""}
	hour := time.Now().In(time.Local).Hour()
	if hour >= 22 || hour < 8 {
		return true, "未成年人每日22:00至次日08:00禁止消费"
	}
	return false, ""
}

// ---------- 文件大小校验 ----------

// ValidateFileSize 文件大小校验
func ValidateFileSize(sizeBytes int64, maxSizeMB int) error {
	if maxSizeMB <= 0 {maxSizeMB = 5}
	maxBytes := int64(maxSizeMB) * 1024 * 1024
	if sizeBytes > maxBytes {
		return fmt.Errorf("文件大小不能超过%dMB", maxSizeMB)
	}
	return nil
}

// ---------- 未读数格式化（前端禁止 formatUnread） ----------

// FormatUnreadCount 未读数格式化(>99显示99+)
func FormatUnreadCount(count int) string {
	if count > 99 {return "99+"}
	return strconv.Itoa(count)
}

// ---------- 折扣/节省金额计算（前端禁止计算折扣） ----------

// DiscountCalcResult 折扣计算结果
type DiscountCalcResult struct {
	DiscountPercent int     `json:"discount_percent"` // 85 = 8.5折
	DiscountLabel   string  `json:"discount_label"`   // "8.5折"
	SaveAmount      string  `json:"save_amount"`      // "12.50"
}

// CalcDiscount 折扣/节省金额计算
func CalcDiscount(originalPriceFen, groupPriceFen int64) DiscountCalcResult {
	if originalPriceFen <= 0 {
		return DiscountCalcResult{DiscountLabel: "0折", SaveAmount: "0.00"}
	}
	ratio := float64(groupPriceFen) / float64(originalPriceFen)
	percent := int(math.Round(ratio * 100))
	label := fmt.Sprintf("%.1f折", float64(percent)/10.0)
	saveFen := originalPriceFen - groupPriceFen
	if saveFen < 0 {saveFen = 0}
	return DiscountCalcResult{
		DiscountPercent: percent,
		DiscountLabel:   label,
		SaveAmount:      FormatFenToYuan(saveFen),
	}
}

// ---------- 证据/内容签名（防篡改，前端禁止 generateId 做业务ID） ----------

// GenerateContentHash 生成内容哈希(用于业务签名)
func GenerateContentHash(content string) string {
	h := sha256.Sum256([]byte(content))
	return fmt.Sprintf("%x", h[:16])
}

// ---------- 统一错误结构（前端禁止错误码映射） ----------

// UserError 用户友好错误结构(前端只读 user_message 展示)
type UserError struct {
	Code        int    `json:"code"`         // 业务错误码
	UserMessage string `json:"user_message"` // 前端直接展示的文案
	FieldType   string `json:"field_type"`   // 关联字段(如表单字段名)
}

// NewUserError 构造用户错误
func NewUserError(code int, msg string, field string) UserError {
	return UserError{Code: code, UserMessage: msg, FieldType: field}
}
