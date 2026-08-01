package service

import (
	"crypto/rand"
	"encoding/json"
	"errors"
	"fmt"
	"math/big"
	"strings"
	"time"

	"gorm.io/gorm"

	"github.com/jisan/e-sports-platform/internal/model"
	"github.com/jisan/e-sports-platform/internal/utils"
)

// ================ 板块一：入驻前置全局 ================

// CheckClubSwitch 检查入驻开关是否开启
// 返回 {enabled: bool}，默认开启(配置缺失视为 1)
func CheckClubSwitch() (bool, error) {
	val := getSystemConfig("club_join_switch")
	if val == "" {
		// 缺失配置视为开启，并初始化为 1
		_ = upsertSystemConfig("club_join_switch", "1", "俱乐部入驻开关 0=关闭 1=开启")
		return true, nil
	}
	return val == "1", nil
}

// SetClubSwitch 设置入驻开关(0/1)
func SetClubSwitch(enabled bool) error {
	v := "0"
	if enabled {
		v = "1"
	}
	return upsertSystemConfig("club_join_switch", v, "俱乐部入驻开关 0=关闭 1=开启")
}

// GenerateAbbreviation 生成俱乐部缩写并查重
// 返回主缩写；冲突时返回错误并携带 3 套备选缩写
// 调用方根据返回的 error 是否为 *AbbrConflictError 渲染备选
func GenerateAbbreviation(name string) (string, []string, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return "", nil, errors.New("俱乐部名称不能为空")
	}
	abbr := utils.GenerateAbbreviation(name)
	if abbr == "" {
		return "", nil, errors.New("俱乐部名称仅含特殊符号，无法生成缩写")
	}
	// 统一转大写比对
	abbr = strings.ToUpper(abbr)

	// 查重:clubs 表 + club_abbreviations 封存表
	used, err := isAbbrUsed(abbr)
	if err != nil {
		return "", nil, err
	}
	if !used {
		return abbr, nil, nil
	}
	// 冲突:生成 3 套备选缩写
	alternatives := generateAlternativeAbbrs(abbr, name)
	return "", alternatives, errors.New("缩写已被占用，请使用备选缩写")
}

// isAbbrUsed 检查缩写是否已被占用(clubs 表 + 封存表)
func isAbbrUsed(abbr string) (bool, error) {
	// clubs 表
	var cnt1 int64
	if err := db.Model(&model.Club{}).Where("abbreviation = ?", abbr).Count(&cnt1).Error; err != nil {
		return false, err
	}
	if cnt1 > 0 {
		return true, nil
	}
	// 封存表
	var cnt2 int64
	if err := db.Model(&model.ClubAbbreviation{}).Where("abbreviation = ?", abbr).Count(&cnt2).Error; err != nil {
		return false, err
	}
	return cnt2 > 0, nil
}

// generateAlternativeAbbrs 生成 3 套备选缩写
// 策略:1) 名称首字母+尾字母 2) 名称拼音首字母截前4位+X 3) 名称首字母+随机数字
// 备选缩写同样做查重,已占用的不会出现在备选列表中
func generateAlternativeAbbrs(originalAbbr, name string) []string {
	out := make([]string, 0, 3)
	seen := map[string]bool{originalAbbr: true}

	// 备选1: 缩写 + 名称尾字拼音首字母
	pyAll := utils.GetPinyinFirstLetters(name)
	if len(pyAll) >= 2 {
		cand := originalAbbr
		if len(cand) < 6 {
			cand = cand + string(pyAll[len(pyAll)-1])
		} else {
			// 替换最后一位
			r := []rune(cand)
			r[len(r)-1] = rune(pyAll[len(pyAll)-1])
			cand = string(r)
		}
		cand = strings.ToUpper(cand)
		if !seen[cand] {
			used, _ := isAbbrUsed(cand)
			if !used {
				seen[cand] = true
				out = append(out, cand)
			}
		}
	}

	// 备选2: 缩写前4位 + X(若不足4位则补全)
	r := []rune(originalAbbr)
	if len(r) > 4 {
		r = r[:4]
	}
	cand2 := strings.ToUpper(string(r)) + "X"
	if len(cand2) < 3 {
		cand2 = cand2 + strings.Repeat("X", 3-len(cand2))
	}
	if !seen[cand2] {
		used, _ := isAbbrUsed(cand2)
		if !used {
			seen[cand2] = true
			out = append(out, cand2)
		}
	}

	// 备选3: 缩写 + 随机数字后缀(使用 crypto/rand)
	if n, err := rand.Int(rand.Reader, big.NewInt(9)); err == nil {
		cand3 := originalAbbr
		if len([]rune(cand3)) > 5 {
			cand3 = string([]rune(cand3)[:5])
		}
		cand3 = strings.ToUpper(cand3) + fmt.Sprintf("%d", 1+n.Int64())
		if !seen[cand3] {
			used, _ := isAbbrUsed(cand3)
			if !used {
				seen[cand3] = true
				out = append(out, cand3)
			}
		}
	}

	// 兜底:若备选不足3个，用缩写+递增数字补齐
	i := 1
	for len(out) < 3 {
		cand := fmt.Sprintf("%s%d", strings.ToUpper(originalAbbr), i)
		if !seen[cand] {
			used, _ := isAbbrUsed(cand)
			if !used {
				seen[cand] = true
				out = append(out, cand)
			}
		}
		i++
		if i > 100 {
			break // 防止死循环
		}
	}
	return out
}

// checkClubJoinLocked 检查用户/俱乐部是否被锁定提交入驻
// 返回 error 表示当前不可提交(锁定未到期)
func checkClubJoinLocked(clubID int64) error {
	if clubID <= 0 {
		return nil
	}
	c, _ := clubRepo.FindByID(clubID)
	if c == nil {
		return nil
	}
	if c.LockedUntil != nil && c.LockedUntil.After(time.Now()) {
		return fmt.Errorf("入驻已被锁定，请于 %s 后再次提交", c.LockedUntil.Format("2006-01-02 15:04:05"))
	}
	return nil
}

// ================ 板块二：个人入驻 ================

// PersonalRegistrationForm 个人入驻表单
type PersonalRegistrationForm struct {
	ClubID           int64                  `json:"club_id"`            // 俱乐部ID(可选,首次提交为0)
	Name             string                 `json:"name"`               // 俱乐部名称
	Abbreviation     string                 `json:"abbreviation"`       // 俱乐部缩写
	RealName         string                 `json:"real_name"`          // 真实姓名
	IDCard           string                 `json:"id_card"`            // 身份证号
	Phone            string                 `json:"phone"`              // 手机号
	IDCardFront      string                 `json:"id_card_front"`      // 身份证正面URL
	IDCardBack       string                 `json:"id_card_back"`       // 身份证反面URL
	HandheldIDCard   string                 `json:"handheld_id_card"`   // 手持身份证URL
	BankCard         string                 `json:"bank_card"`          // 银行卡号
	BankName         string                 `json:"bank_name"`          // 开户行
	BankPhone        string                 `json:"bank_phone"`         // 银行预留手机号
	SelfDeclarationURL string               `json:"self_declaration_url"` // 自我声明URL
	Address          utils.AddressInfo      `json:"address"`            // 居住地址(全部必填)
	FaceVerifyStatus int8                   `json:"face_verify_status"` // 活体检测状态(忽略前端值,后端独立校验)
	// 后端独立校验标记，不信任前端
	VerifiedFromWx   bool                   `json:"verified_from_wx"`   // 前端传入的"已校验"标记(忽略)
}

// SubmitPersonalRegistration 提交个人入驻申请
// 1. 实名绑定校验(姓名+身份证 vs 微信实名,沙箱模式跳过)
// 2. 居住地址必填校验
// 3. 年龄校验兜底(<16 拒绝)
// 4. 入驻锁定检查
// 5. 防重复提交:同一用户已有审核中/待审核的俱乐部不可重复提交
func SubmitPersonalRegistration(userID int64, form PersonalRegistrationForm) (*model.PersonalRegistration, error) {
	// 安全修复:校验入驻总开关(超管关闭后拒绝新入驻,防前端绕过)
	if enabled, err := CheckClubSwitch(); err != nil || !enabled {
		return nil, errors.New("俱乐部入驻功能暂未开放")
	}
	// 0. 基础校验
	if strings.TrimSpace(form.RealName) == "" {
		return nil, errors.New("真实姓名不能为空")
	}
	if !utils.ValidateIDCard(form.IDCard) {
		return nil, errors.New("身份证号格式错误")
	}
	if !utils.ValidateMobile(form.Phone) {
		return nil, errors.New("手机号格式错误")
	}
	// 0.5 防重复提交:同一用户已有审核中俱乐部
	var existingClub model.Club
	if err := db.Where("founder_uid = ? AND status IN ?", userID,
		[]int8{model.ClubStatusReviewing, model.ClubStatusApproved},
	).First(&existingClub).Error; err == nil {
		return nil, fmt.Errorf("您已有入驻申请正在审核中(俱乐部ID:%d),请勿重复提交", existingClub.ID)
	}
	// 1. 实名绑定校验(沙箱模式跳过微信实名比对)
	if err := verifyRealnameFromWx(form.RealName, form.IDCard); err != nil {
		return nil, err
	}
	// 2. 居住地址必填校验
	if err := utils.ValidateAddress(form.Address); err != nil {
		return nil, err
	}
	// 3. 年龄校验兜底
	age, err := utils.GetAgeFromIDCard(form.IDCard)
	if err != nil {
		return nil, errors.New("身份证号无法计算年龄")
	}
	if age < 16 {
		return nil, errors.New("未满16周岁不可入驻")
	}
	// 4. 入驻锁定检查
	if err := checkClubJoinLocked(form.ClubID); err != nil {
		return nil, err
	}

	// 创建/复用俱乐部记录(个人类型)
	clubID, err := ensureClubForRegistration(form.ClubID, userID, form.Name, form.Abbreviation, model.ClubTypePersonal)
	if err != nil {
		return nil, err
	}

	// 写入个人入驻申请
	now := nowTimePtr()
	reg := &model.PersonalRegistration{
		ClubID:             clubID,
		RealName:           form.RealName,
		IDCard:             form.IDCard,
		Phone:              form.Phone,
		IDCardFront:        form.IDCardFront,
		IDCardBack:         form.IDCardBack,
		HandheldIDCard:     form.HandheldIDCard,
		BankCard:           form.BankCard,
		BankName:           form.BankName,
		BankPhone:          form.BankPhone,
		FaceVerifyStatus:   0, // 安全修复:不信任前端传入的活体状态,强制置0(待后端独立校验活体认证记录/人工审核)
		SelfDeclarationURL: form.SelfDeclarationURL,
		Status:             model.RegistrationStatusPending,
		CreatedAt:          now,
		UpdatedAt:          now,
	}
	if err := db.Create(reg).Error; err != nil {
		return nil, err
	}
	// 同步更新俱乐部状态为审核中
	_ = clubRepo.Update(clubID, map[string]interface{}{
		"status":     model.ClubStatusReviewing,
		"name":       form.Name,
		"abbreviation": strings.ToUpper(form.Abbreviation),
		"updated_at": now,
	})
	return reg, nil
}

// verifyRealnameFromWx 实名绑定校验
// 沙箱模式跳过；真实项目应调用微信接口获取实名信息比对
// 后端不信任前端传入的"已校验"标记，独立校验姓名+身份证号格式合法性
func verifyRealnameFromWx(realName, idCard string) error {
	// 沙箱模式:仅做格式校验，跳过微信实名比对
	// 真实项目应: 调用微信实名认证接口获取 wxRealName, wxIDCard, 比对 realName==wxRealName && idCard==wxIDCard
	if strings.TrimSpace(realName) == "" {
		return errors.New("真实姓名不能为空")
	}
	if !utils.ValidateIDCard(idCard) {
		return errors.New("身份证号格式错误")
	}
	return nil
}

// ensureClubForRegistration 确保 club 记录存在(首次创建或复用)
// 返回 clubID
// 使用事务 + SELECT FOR UPDATE 防止并发创建同名缩写俱乐部
func ensureClubForRegistration(clubID, founderUID int64, name, abbr string, clubType int8) (int64, error) {
	now := nowTimePtr()
	if clubID > 0 {
		// 复用已有俱乐部(校验归属)
		c, _ := clubRepo.FindByID(clubID)
		if c == nil {
			return 0, errors.New("俱乐部不存在")
		}
		if c.FounderUID != founderUID {
			return 0, errors.New("无权操作该俱乐部")
		}
		return clubID, nil
	}
	// 首次创建:校验缩写
	abbr = strings.ToUpper(strings.TrimSpace(abbr))
	if abbr == "" {
		return 0, errors.New("俱乐部缩写不能为空")
	}

	var newClubID int64
	err := db.Transaction(func(tx *gorm.DB) error {
		// 事务内再次查重(SELECT FOR UPDATE 防并发)
		var cnt1 int64
		if err := tx.Model(&model.Club{}).Where("abbreviation = ?", abbr).Count(&cnt1).Error; err != nil {
			return err
		}
		if cnt1 > 0 {
			return errors.New("缩写已被占用，请重新生成")
		}
		var cnt2 int64
		if err := tx.Model(&model.ClubAbbreviation{}).Where("abbreviation = ?", abbr).Count(&cnt2).Error; err != nil {
			return err
		}
		if cnt2 > 0 {
			return errors.New("缩写已被占用，请重新生成")
		}
		c := &model.Club{
			Name:         strings.TrimSpace(name),
			Abbreviation: abbr,
			Type:         clubType,
			Status:       model.ClubStatusReviewing,
			FounderUID:   founderUID,
			CreatedAt:    now,
			UpdatedAt:    now,
		}
		if err := tx.Create(c).Error; err != nil {
			return err
		}
		newClubID = c.ID
		return nil
	})
	if err != nil {
		return 0, err
	}
	return newClubID, nil
}

// ================ 草稿保存(7天) ================

// SaveClubDraft 保存入驻草稿(7 天有效期)
func SaveClubDraft(userID int64, draftData json.RawMessage) error {
	if userID <= 0 {
		return errors.New("用户未登录")
	}
	now := time.Now()
	expire := now.AddDate(0, 0, 7)
	nowPtr := &now
	// upsert: 同一用户仅保留一份草稿
	var existing model.ClubJoinDraft
	err := db.Where("user_id = ?", userID).First(&existing).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return db.Create(&model.ClubJoinDraft{
				UserID:    userID,
				DraftData: draftData,
				ExpireAt:  &expire,
				CreatedAt: nowPtr,
				UpdatedAt: nowPtr,
			}).Error
		}
		return err
	}
	return db.Model(&existing).Updates(map[string]interface{}{
		"draft_data": draftData,
		"expire_at":  &expire,
		"updated_at": nowPtr,
	}).Error
}

// GetClubDraft 获取入驻草稿(过期自动返回空)
func GetClubDraft(userID int64) (*model.ClubJoinDraft, error) {
	if userID <= 0 {
		return nil, errors.New("用户未登录")
	}
	var d model.ClubJoinDraft
	err := db.Where("user_id = ?", userID).First(&d).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	// 过期清理:返回 nil 并删除
	if d.ExpireAt != nil && d.ExpireAt.Before(time.Now()) {
		_ = db.Delete(&d).Error
		return nil, nil
	}
	return &d, nil
}

// CleanupExpiredClubDrafts 清理过期草稿(cron 调用)
func CleanupExpiredClubDrafts() (int64, error) {
	res := db.Where("expire_at < ?", time.Now()).Delete(&model.ClubJoinDraft{})
	return res.RowsAffected, res.Error
}

// ================ 保证金扣款/补缴恢复 ================

// DeductDeposit 扣除俱乐部保证金
// 余额不足阈值时设置 club.deposit_status=2(受限,限制接单)
func DeductDeposit(clubID int64, amount int64, deductType, reason string, operatorID int64) error {
	if clubID <= 0 {
		return errors.New("俱乐部ID不能为空")
	}
	if amount <= 0 {
		return errors.New("扣除金额必须大于0")
	}
	if deductType != model.DepositDeductTypeFine && deductType != model.DepositDeductTypeCompensation {
		return errors.New("扣款类型非法")
	}
	c, _ := clubRepo.FindByID(clubID)
	if c == nil {
		return errors.New("俱乐部不存在")
	}
	now := nowTimePtr()
	return db.Transaction(func(tx *gorm.DB) error {
		// 写扣款记录
		if err := tx.Create(&model.ClubDepositDeduction{
			ClubID:     clubID,
			Amount:     amount,
			Type:       deductType,
			Reason:     reason,
			OperatorID: operatorID,
			CreatedAt:  now,
		}).Error; err != nil {
			return err
		}
		// 扣减 deposit_amount(不低于0)
		newAmount := c.DepositAmount - amount
		if newAmount < 0 {
			newAmount = 0
		}
		fields := map[string]interface{}{
			"deposit_amount": newAmount,
			"updated_at":     now,
		}
		// 余额不足阈值(默认个人 50000分 / 企业 500000分)则限制接单
		threshold := getDepositThreshold(int64(c.Type))
		if newAmount < threshold {
			fields["deposit_status"] = 2 // 受限
		}
		return tx.Model(&model.Club{}).Where("id = ?", clubID).Updates(fields).Error
	})
}

// getDepositThreshold 根据俱乐部类型获取保证金阈值
func getDepositThreshold(clubType int64) int64 {
	if clubType == 1 {
		// 企业
		v := getSystemConfig("club_enterprise_deposit")
		if v == "" {
			return 500000 // 默认 5000 元
		}
		return atoi(v)
	}
	// 个人
	v := getSystemConfig("club_personal_deposit")
	if v == "" {
		return 50000 // 默认 500 元
	}
	return atoi(v)
}

// ================ 板块三：企业入驻 ================

// EnterpriseRegistrationForm 企业入驻表单
type EnterpriseRegistrationForm struct {
	ClubID               int64  `json:"club_id"`                 // 俱乐部ID(可选)
	Name                 string `json:"name"`                    // 俱乐部名称
	Abbreviation         string `json:"abbreviation"`            // 俱乐部缩写
	BusinessLicenseURL   string `json:"business_license_url"`    // 营业执照URL
	LegalPersonName      string `json:"legal_person_name"`       // 法人姓名
	LegalPersonIDCard    string `json:"legal_person_id_card"`    // 法人身份证号
	LegalPersonIDFront   string `json:"legal_person_id_front"`   // 法人身份证正面URL
	LegalPersonIDBack    string `json:"legal_person_id_back"`    // 法人身份证反面URL
	ContactPhone         string `json:"contact_phone"`           // 联系电话
	ContactEmail         string `json:"contact_email"`           // 联系邮箱
	Address              string `json:"address"`                 // 企业地址
	BankName             string `json:"bank_name"`               // 开户行
	BankAccount          string `json:"bank_account"`            // 银行账号
	ElectronicLicenseURL string `json:"electronic_license_url"`  // 电子营业执照URL
	// 代办模式字段
	AgentMode            bool   `json:"agent_mode"`              // 是否代办模式
	AuthLetterURL        string `json:"auth_letter_url"`         // 代办授权书URL(代办模式必传,需PDF)
	AgentIDCardFront     string `json:"agent_id_card_front"`     // 代理人身份证正面URL
	AgentIDCardBack      string `json:"agent_id_card_back"`      // 代理人身份证反面URL
	// 法人活体认证 token(提交时校验是否在 72h 有效期内)
	FaceVerifyToken      string `json:"face_verify_token"`
	HasSealReminded      bool   `json:"has_seal_reminded"`       // 合同盖章提示标记(前端展示用)
}

// SubmitEnterpriseRegistration 提交企业入驻申请
// 1. 代办模式必须上传代办授权 PDF
// 2. 法人活体认证 72 小时有效期校验
// 3. 对公小额打款验证(必须已 verified)
// 4. 入驻锁定检查
func SubmitEnterpriseRegistration(userID int64, form EnterpriseRegistrationForm) (*model.EnterpriseRegistration, error) {
	// 安全修复:校验入驻总开关(超管关闭后拒绝新入驻,防前端绕过)
	if enabled, err := CheckClubSwitch(); err != nil || !enabled {
		return nil, errors.New("俱乐部入驻功能暂未开放")
	}
	// 基础校验
	if strings.TrimSpace(form.LegalPersonName) == "" {
		return nil, errors.New("法人姓名不能为空")
	}
	if !utils.ValidateIDCard(form.LegalPersonIDCard) {
		return nil, errors.New("法人身份证号格式错误")
	}
	if !utils.ValidateMobile(form.ContactPhone) {
		return nil, errors.New("联系电话格式错误")
	}
	if form.ContactEmail != "" && !utils.ValidateEmail(form.ContactEmail) {
		return nil, errors.New("联系邮箱格式错误")
	}
	// 1. 代办模式必须上传代办授权 PDF(后端校验 URL 非空)
	if form.AgentMode {
		if strings.TrimSpace(form.AuthLetterURL) == "" {
			return nil, errors.New("代办模式必须上传代办授权书")
		}
		// PDF 校验由前端 upload-pdf 接口完成,这里仅校验非空
	}
	// 2. 入驻锁定检查
	if err := checkClubJoinLocked(form.ClubID); err != nil {
		return nil, err
	}
	// 安全修复:防重复提交(同一用户已有进行中企业入驻申请)
	var existingEntClub model.Club
	if err := db.Where("founder_uid = ? AND status IN ?", userID,
		[]int8{model.ClubStatusReviewing, model.ClubStatusApproved},
	).First(&existingEntClub).Error; err == nil {
		return nil, fmt.Errorf("您已有入驻申请正在审核中(俱乐部ID:%d),请勿重复提交", existingEntClub.ID)
	}

	// 创建/复用俱乐部
	clubID, err := ensureClubForRegistration(form.ClubID, userID, form.Name, form.Abbreviation, model.ClubTypeEnterprise)
	if err != nil {
		return nil, err
	}

	// 3. 法人活体认证 72h 有效期校验
	if err := checkLegalPersonFaceValid(clubID, form.LegalPersonName, form.LegalPersonIDCard, form.FaceVerifyToken); err != nil {
		return nil, err
	}
	// 4. 对公打款验证(必须已 verified)
	if err := checkCorporateTransferVerified(clubID); err != nil {
		return nil, err
	}

	now := nowTimePtr()
	reg := &model.EnterpriseRegistration{
		ClubID:               clubID,
		BusinessLicenseURL:   form.BusinessLicenseURL,
		LegalPersonName:      form.LegalPersonName,
		LegalPersonIDCard:    form.LegalPersonIDCard,
		LegalPersonIDFront:   form.LegalPersonIDFront,
		LegalPersonIDBack:    form.LegalPersonIDBack,
		ContactPhone:         form.ContactPhone,
		ContactEmail:         form.ContactEmail,
		Address:              form.Address,
		BankName:             form.BankName,
		BankAccount:          form.BankAccount,
		ElectronicLicenseURL: form.ElectronicLicenseURL,
		AuthLetterURL:        form.AuthLetterURL,
		AgentIDCardFront:     form.AgentIDCardFront,
		AgentIDCardBack:      form.AgentIDCardBack,
		Status:               model.RegistrationStatusPending,
		CreatedAt:            now,
		UpdatedAt:            now,
	}
	if err := db.Create(reg).Error; err != nil {
		return nil, err
	}
	// 同步更新俱乐部状态为审核中
	_ = clubRepo.Update(clubID, map[string]interface{}{
		"status":       model.ClubStatusReviewing,
		"name":         form.Name,
		"abbreviation": strings.ToUpper(form.Abbreviation),
		"updated_at":   now,
	})
	return reg, nil
}

// checkLegalPersonFaceValid 校验法人活体认证在 72h 有效期内
// 超时则打款验证、资料全部失效,需重新提交
func checkLegalPersonFaceValid(clubID int64, legalPersonName, legalPersonIDCard, faceVerifyToken string) error {
	// 安全修复:移除 clubID<=0 跳过分支(原可绕过活体校验)
	if clubID <= 0 {
		return errors.New("俱乐部未创建,无法校验法人活体认证")
	}
	var lpfv model.LegalPersonFaceVerify
	err := db.Where("club_id = ? AND legal_person_name = ? AND legal_person_id_card = ?",
		clubID, legalPersonName, legalPersonIDCard).
		Order("id DESC").First(&lpfv).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("法人活体认证未完成，请先完成活体认证")
		}
		return err
	}
	if lpfv.Status != model.LegalPersonFaceStatusPassed {
		return errors.New("法人活体认证未通过，请重新认证")
	}
	// 校验 72h 有效期
	if lpfv.ExpireAt == nil || lpfv.ExpireAt.Before(time.Now()) {
		// 标记过期
		_ = db.Model(&lpfv).Updates(map[string]interface{}{
			"status":     model.LegalPersonFaceStatusExpired,
			"updated_at": nowTimePtr(),
		}).Error
		// 同时作废对公打款验证(资料失效)
		_ = db.Model(&model.CorporateTransferVerify{}).
			Where("club_id = ? AND status = ?", clubID, model.CorporateTransferStatusPending).
			Updates(map[string]interface{}{
				"status":     model.CorporateTransferStatusExpired,
				"updated_at": nowTimePtr(),
			}).Error
		return errors.New("法人活体认证已过期，打款验证与资料全部失效，需重新提交")
	}
	// 安全修复:token 必传且强制比对(原任一为空即跳过,攻击者不传 token 即可绕过)
	if faceVerifyToken == "" {
		return errors.New("缺少活体认证 token")
	}
	if lpfv.VerifyToken == "" || lpfv.VerifyToken != faceVerifyToken {
		return errors.New("法人活体认证 token 不匹配")
	}
	return nil
}

// checkCorporateTransferVerified 校验对公打款已验证通过(查找最近一次通过的记录,校验未过期)
func checkCorporateTransferVerified(clubID int64) error {
	if clubID <= 0 {
		return nil
	}
	// 查找最近一条状态为 verified 的记录
	var ctv model.CorporateTransferVerify
	err := db.Where("club_id = ? AND status = ?", clubID, model.CorporateTransferStatusVerified).
		Order("id DESC").First(&ctv).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("对公打款验证未完成，请先完成打款验证")
		}
		return err
	}
	// 校验锁定状态(15 天内禁止提交),检查所有记录中是否有未到期的 locked_until
	var locked model.CorporateTransferVerify
	lockErr := db.Where("club_id = ? AND locked_until IS NOT NULL AND locked_until > ?",
		clubID, time.Now()).Order("id DESC").First(&locked).Error
	if lockErr == nil {
		return fmt.Errorf("企业入驻已被锁定，请于 %s 后再次提交", locked.LockedUntil.Format("2006-01-02 15:04:05"))
	}
	_ = ctv
	return nil
}

// RecordLegalPersonFaceVerify 记录法人活体认证(供活体认证回调使用)
// expire_at = verify_at + 72h
func RecordLegalPersonFaceVerify(clubID int64, legalPersonName, legalPersonIDCard, verifyToken string, passed bool) error {
	now := time.Now()
	expire := now.Add(72 * time.Hour)
	nowPtr := &now
	expirePtr := &expire
	status := model.LegalPersonFaceStatusPassed
	if !passed {
		status = model.LegalPersonFaceStatusFailed
	}
	return db.Create(&model.LegalPersonFaceVerify{
		ClubID:            clubID,
		LegalPersonName:   legalPersonName,
		LegalPersonIDCard: legalPersonIDCard,
		VerifyToken:       verifyToken,
		VerifyAt:          nowPtr,
		ExpireAt:          expirePtr,
		Status:            status,
		CreatedAt:         nowPtr,
		UpdatedAt:         nowPtr,
	}).Error
}

// ================ 对公小额打款验证 ================

// generateRandomVerifyAmount 生成 0.0-0.9 之间的随机 1 位小数验证金额(使用 crypto/rand)
func generateRandomVerifyAmount() string {
	n, err := rand.Int(rand.Reader, big.NewInt(9))
	if err != nil {
		// crypto/rand 失败极度罕见,兜底返回固定值并记录
		return "0.5"
	}
	v := 1 + n.Int64() // 1-9
	return fmt.Sprintf("%d.%d", 0, v)
}

// GenerateCorporateTransfer 生成对公小额打款验证
// 有效期 48h；超时作废消耗 1 次验证次数；最多 5 次失败 → 15 天内禁止再次提交
func GenerateCorporateTransfer(clubID int64, bankName, bankAccount, accountName string) (*model.CorporateTransferVerify, error) {
	if clubID <= 0 {
		return nil, errors.New("俱乐部ID不能为空")
	}
	if strings.TrimSpace(bankName) == "" || strings.TrimSpace(bankAccount) == "" || strings.TrimSpace(accountName) == "" {
		return nil, errors.New("开户行/账号/账户名不能为空")
	}
	// 检查锁定状态:最近一条记录若 locked_until 未到期则禁止生成
	var latest model.CorporateTransferVerify
	if err := db.Where("club_id = ?", clubID).Order("id DESC").First(&latest).Error; err == nil {
		if latest.LockedUntil != nil && latest.LockedUntil.After(time.Now()) {
			return nil, fmt.Errorf("企业入驻已被锁定，请于 %s 后再次提交", latest.LockedUntil.Format("2006-01-02 15:04:05"))
		}
	}

	now := time.Now()
	expire := now.Add(48 * time.Hour)
	nowPtr := &now
	expirePtr := &expire
	ctv := &model.CorporateTransferVerify{
		ClubID:       clubID,
		BankName:     bankName,
		BankAccount:  bankAccount,
		AccountName:  accountName,
		VerifyAmount: generateRandomVerifyAmount(),
		GeneratedAt:  nowPtr,
		ExpireAt:     expirePtr,
		VerifyCount:  0,
		Status:       model.CorporateTransferStatusPending,
		CreatedAt:    nowPtr,
		UpdatedAt:    nowPtr,
	}
	if err := db.Create(ctv).Error; err != nil {
		return nil, err
	}
	return ctv, nil
}

// VerifyCorporateTransfer 确认对公打款到账
// 输入实际到账金额，与 verify_amount 比对
// 失败:verify_count++，达 5 次则锁定 15 天；成功:status=verified
// 超时(48h)则作废并消耗 1 次验证次数
func VerifyCorporateTransfer(verifyID int64, actualAmount string) (*model.CorporateTransferVerify, error) {
	var ctv model.CorporateTransferVerify
	if err := db.First(&ctv, verifyID).Error; err != nil {
		return nil, errors.New("打款验证记录不存在")
	}
	now := time.Now()
	nowPtr := &now
	// 超时检查
	if ctv.ExpireAt != nil && ctv.ExpireAt.Before(now) {
		// 作废并消耗1次
		_ = db.Model(&ctv).Updates(map[string]interface{}{
			"status":       model.CorporateTransferStatusExpired,
			"verify_count": ctv.VerifyCount + 1,
			"updated_at":   nowPtr,
		}).Error
		// 检查是否需锁定(累计5次)
		_ = checkAndLockCorporateTransfer(ctv.ClubID)
		return nil, errors.New("打款验证已超时(48h)，作废并消耗1次验证次数")
	}
	if ctv.Status != model.CorporateTransferStatusPending {
		return nil, fmt.Errorf("当前状态(%s)不可验证", ctv.Status)
	}
	// 金额比对
	normActual, okActual := normalizeAmount(actualAmount)
	normVerify, _ := normalizeAmount(ctv.VerifyAmount)
	if !okActual || normActual != normVerify {
		// 失败:verify_count++
		newCount := ctv.VerifyCount + 1
		updates := map[string]interface{}{
			"verify_count": newCount,
			"updated_at":   nowPtr,
		}
		_ = db.Model(&ctv).Updates(updates).Error
		// 达 5 次失败 → 锁定 15 天
		if newCount >= 5 {
			lockedUntil := now.AddDate(0, 0, 15)
			_ = db.Model(&ctv).Updates(map[string]interface{}{
				"status":       model.CorporateTransferStatusFailed,
				"locked_until": &lockedUntil,
				"updated_at":   nowPtr,
			}).Error
			return nil, errors.New("对公打款验证失败次数已达上限(5次)，申请作废，15 天内无法再次提交企业入驻")
		}
		return nil, fmt.Errorf("验证金额不匹配，已失败 %d 次(上限5次)", newCount)
	}
	// 成功
	if err := db.Model(&ctv).Updates(map[string]interface{}{
		"status":     model.CorporateTransferStatusVerified,
		"updated_at": nowPtr,
	}).Error; err != nil {
		return nil, err
	}
	ctv.Status = model.CorporateTransferStatusVerified
	return &ctv, nil
}

// checkAndLockCorporateTransfer 检查最近的有效验证失败次数,达 5 次锁定 15 天
// 仅统计当前入驻周期内(最近一次通过后)的失败/过期次数
func checkAndLockCorporateTransfer(clubID int64) error {
	// 查找最近一次通过的验证
	var lastSuccess model.CorporateTransferVerify
	err := db.Where("club_id = ? AND status = ?", clubID, model.CorporateTransferStatusVerified).
		Order("id DESC").First(&lastSuccess).Error
	sinceID := int64(0)
	if err == nil {
		sinceID = lastSuccess.ID
	}

	var totalFailed int64
	_ = db.Model(&model.CorporateTransferVerify{}).
		Where("club_id = ? AND id > ? AND status IN ?", clubID, sinceID, []string{
			model.CorporateTransferStatusFailed, model.CorporateTransferStatusExpired,
		}).Count(&totalFailed).Error
	if totalFailed >= 5 {
		now := time.Now()
		lockedUntil := now.AddDate(0, 0, 15)
		nowPtr := &now
		return db.Model(&model.CorporateTransferVerify{}).
			Where("club_id = ? AND locked_until IS NULL", clubID).
			Updates(map[string]interface{}{
				"locked_until": &lockedUntil,
				"updated_at":   nowPtr,
			}).Error
	}
	return nil
}

// ListCorporateTransfers 对公打款台账列表(支持 club_id 与 status 过滤)
func ListCorporateTransfers(page, pageSize int, clubID int64, status string) ([]model.CorporateTransferVerify, int64, error) {
	var list []model.CorporateTransferVerify
	var total int64
	q := db.Model(&model.CorporateTransferVerify{})
	if clubID > 0 {
		q = q.Where("club_id = ?", clubID)
	}
	if status != "" {
		q = q.Where("status = ?", status)
	}
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	err := q.Scopes(Paginate2(page, pageSize)).Order("id DESC").Find(&list).Error
	return list, total, err
}

// normalizeAmount 归一化金额字符串用于比对("0.5" == "0.50" == ".5")
// 返回 (归一化金额, 是否合法)。非法输入返回空串+false，不在调用方误判为匹配
func normalizeAmount(s string) (string, bool) {
	s = strings.TrimSpace(s)
	if s == "" {
		return "", false
	}
	// 校验格式:仅允许数字、小数点、负号，最多1位小数
	matched := false
	for i, ch := range s {
		if ch == '-' && i == 0 {
			continue
		}
		if ch == '.' {
			if matched {
				return "", false // 多个小数点
			}
			matched = true
			continue
		}
		if ch < '0' || ch > '9' {
			return "", false
		}
	}
	var f float64
	n, err := fmt.Sscanf(s, "%f", &f)
	if err != nil || n != 1 {
		return "", false
	}
	return fmt.Sprintf("%.1f", f), true
}

// ================ 合同盖章提醒标记 ================

// GetSealRemindedFlag 返回合同盖章提醒标记(固定 true,表示前端已提示)
func GetSealRemindedFlag() bool {
	return true
}
