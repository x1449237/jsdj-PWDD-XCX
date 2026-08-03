package service

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/jisan/e-sports-platform/internal/model"
	"github.com/jisan/e-sports-platform/internal/utils"
)

// ============================================================
// 俱乐部入驻零逻辑配套服务
// 所有步骤校验、活体认证、合同模板下发、统一提交均在后端完成。
// 前端只调用 API 并渲染后端返回的 _text/_masked 字段。
// ============================================================

// ---------- 入驻开关 + 保证金配置下发 ----------

// ClubSwitchInfo 入驻开关与保证金配置(前端零逻辑直读)
type ClubSwitchInfo struct {
	ClubJoinOpen         bool   `json:"club_join_open"`
	PersonalDeposit      int64  `json:"personal_deposit"`        // 分
	EnterpriseDeposit    int64  `json:"enterprise_deposit"`      // 分
	PersonalDepositText  string `json:"personal_deposit_text"`   // 后端格式化文本
	EnterpriseDepositText string `json:"enterprise_deposit_text"`
}

// CheckClubSwitchFull 查询入驻开关 + 保证金配置(前端直读,无需换算)
func CheckClubSwitchFull() (*ClubSwitchInfo, error) {
	val := getSystemConfig("club_join_switch")
	if val == "" {
		// 缺失配置视为开启,并初始化为 1
		_ = upsertSystemConfig("club_join_switch", "1", "俱乐部入驻开关 0=关闭 1=开启")
		val = "1"
	}
	open := val == "1"

	personalFen := atoi(getSystemConfig("club_personal_deposit_fen"))
	if personalFen <= 0 {
		personalFen = 50000 // 默认 500 元
		_ = upsertSystemConfig("club_personal_deposit_fen", itoa(personalFen), "个人入驻保证金(分)")
	}
	enterpriseFen := atoi(getSystemConfig("club_enterprise_deposit_fen"))
	if enterpriseFen <= 0 {
		enterpriseFen = 200000 // 默认 2000 元
		_ = upsertSystemConfig("club_enterprise_deposit_fen", itoa(enterpriseFen), "企业入驻保证金(分)")
	}

	return &ClubSwitchInfo{
		ClubJoinOpen:          open,
		PersonalDeposit:       personalFen,
		EnterpriseDeposit:     enterpriseFen,
		PersonalDepositText:   FormatFenToYuan(personalFen) + " 元",
		EnterpriseDepositText: FormatFenToYuan(enterpriseFen) + " 元",
	}, nil
}

// ---------- 步骤校验 ----------

// ClubStepForm 俱乐部入驻步骤表单(前端原样回传,后端权威校验)
type ClubStepForm struct {
	ClubType         string `json:"club_type"`          // green_v / blue_v
	Step             int    `json:"step"`
	ClubName         string `json:"club_name"`
	Abbreviation     string `json:"abbreviation"`
	RealName         string `json:"real_name"`
	IDCard           string `json:"id_card"`
	Phone            string `json:"phone"`
	AddressProvince  string `json:"address_province"`
	AddressCity      string `json:"address_city"`
	AddressDistrict  string `json:"address_district"`
	AddressStreet    string `json:"address_street"`
	AddressCommunity string `json:"address_community"`
	AddressBuilding  string `json:"address_building"`
	AddressHouseNo   string `json:"address_house_no"`
	IDCardFront      string `json:"id_card_front"`
	IDCardBack       string `json:"id_card_back"`
	LivenessStatus   int8   `json:"liveness_status"`
	ContractFile     string `json:"contract_file"`
	ContractPdfValid bool   `json:"contract_pdf_valid"`
	// 企业专属
	EnterpriseName    string `json:"enterprise_name"`
	CreditCode        string `json:"credit_code"`
	BusinessLicense   string `json:"business_license"`
	CorporateBank     string `json:"corporate_bank"`
	CorporateAccount  string `json:"corporate_account"`
	HandleType        string `json:"handle_type"` // self / agent
	AgentName         string `json:"agent_name"`
	AgentIDCard       string `json:"agent_id_card"`
	AgentIDCardFront  string `json:"agent_id_card_front"`
	AgentIDCardBack   string `json:"agent_id_card_back"`
	AgentAuthorization string `json:"agent_authorization"`
	AgentAuthPdfValid bool   `json:"agent_auth_pdf_valid"`
}

// ClubStepValidateResult 步骤校验结果(前端只读 can_next + message)
type ClubStepValidateResult struct {
	CanNext     bool   `json:"can_next"`
	Message     string `json:"message"`
	ErrorField  string `json:"error_field"`
}

// ValidateClubStep 入驻步骤权威校验(所有业务规则在后端执行)
// step=1 须知勾选
// step=2 基础资料(名称/缩写/姓名/身份证/年龄≥16/手机号/地址7字段)
// step=3 活体认证(身份证正反面上传 + 活体通过)
// step=4 合同(PDF 校验通过;企业代办模式需代办合同 PDF)
// step=5 保证金支付提示(仅展示,无校验)
// step=6 提交确认(企业额外校验营业执照/对公账户/信用代码)
// step=7 完成
func ValidateClubStep(form ClubStepForm) (*ClubStepValidateResult, error) {
	if form.ClubType != "green_v" && form.ClubType != "blue_v" {
		return &ClubStepValidateResult{CanNext: false, Message: "入驻类型不正确", ErrorField: "club_type"}, nil
	}
	isEnterprise := form.ClubType == "blue_v"

	switch form.Step {
	case 1:
		// 须知勾选由前端 agreed 字段控制,这里只做类型校验
		return &ClubStepValidateResult{CanNext: true}, nil

	case 2:
		// 俱乐部名称
		if strings.TrimSpace(form.ClubName) == "" {
			return &ClubStepValidateResult{CanNext: false, Message: "请填写俱乐部名称", ErrorField: "club_name"}, nil
		}
		if len([]rune(form.ClubName)) < 2 {
			return &ClubStepValidateResult{CanNext: false, Message: "俱乐部名称至少2个字符", ErrorField: "club_name"}, nil
		}
		// 缩写(查重)
		if strings.TrimSpace(form.Abbreviation) == "" {
			return &ClubStepValidateResult{CanNext: false, Message: "请生成俱乐部缩写", ErrorField: "abbreviation"}, nil
		}
		used, err := isAbbrUsed(strings.ToUpper(form.Abbreviation))
		if err != nil {
			return nil, err
		}
		if used {
			return &ClubStepValidateResult{CanNext: false, Message: "缩写被占用,请更换名称或选择备选缩写", ErrorField: "abbreviation"}, nil
		}
		// 真实姓名
		if strings.TrimSpace(form.RealName) == "" {
			return &ClubStepValidateResult{CanNext: false, Message: "请填写真实姓名", ErrorField: "real_name"}, nil
		}
		// 身份证 + 年龄校验(后端权威)
		idResult := ValidateIDCard(form.IDCard)
		if !idResult.Valid {
			return &ClubStepValidateResult{CanNext: false, Message: idResult.Message, ErrorField: "id_card"}, nil
		}
		if idResult.IsUnder16 {
			return &ClubStepValidateResult{CanNext: false, Message: "须年满16周岁", ErrorField: "id_card"}, nil
		}
		// 手机号
		if !ValidatePhone(form.Phone) {
			return &ClubStepValidateResult{CanNext: false, Message: "请填写正确的11位手机号", ErrorField: "phone"}, nil
		}
		// 地址7字段完整性(后端权威判定)
		if !isAddressComplete(form) {
			return &ClubStepValidateResult{CanNext: false, Message: "请完整填写所有地址字段", ErrorField: "address"}, nil
		}
		return &ClubStepValidateResult{CanNext: true}, nil

	case 3:
		// 活体认证步骤:身份证正反面 + 活体通过
		if form.IDCardFront == "" {
			return &ClubStepValidateResult{CanNext: false, Message: "请上传身份证正面", ErrorField: "id_card_front"}, nil
		}
		if form.IDCardBack == "" {
			return &ClubStepValidateResult{CanNext: false, Message: "请上传身份证反面", ErrorField: "id_card_back"}, nil
		}
		if form.LivenessStatus != 1 {
			return &ClubStepValidateResult{CanNext: false, Message: "请完成活体认证", ErrorField: "liveness"}, nil
		}
		return &ClubStepValidateResult{CanNext: true}, nil

	case 4:
		// 合同 PDF 校验
		if form.ContractFile == "" {
			return &ClubStepValidateResult{CanNext: false, Message: "请上传已签署合同", ErrorField: "contract_file"}, nil
		}
		if !form.ContractPdfValid {
			return &ClubStepValidateResult{CanNext: false, Message: "合同PDF校验未通过", ErrorField: "contract_file"}, nil
		}
		// 企业代办模式:必须上传代办合同 PDF 且校验通过
		if isEnterprise && form.HandleType == "agent" {
			if form.AgentAuthorization == "" {
				return &ClubStepValidateResult{CanNext: false, Message: "请上传代办合同PDF", ErrorField: "agent_authorization"}, nil
			}
			if !form.AgentAuthPdfValid {
				return &ClubStepValidateResult{CanNext: false, Message: "代办合同PDF校验未通过", ErrorField: "agent_authorization"}, nil
			}
		}
		return &ClubStepValidateResult{CanNext: true}, nil

	case 5:
		// 保证金支付提示步骤(无校验,前端展示后端下发的金额)
		return &ClubStepValidateResult{CanNext: true}, nil

	case 6:
		// 企业额外字段校验
		if isEnterprise {
			if strings.TrimSpace(form.EnterpriseName) == "" {
				return &ClubStepValidateResult{CanNext: false, Message: "请填写企业名称", ErrorField: "enterprise_name"}, nil
			}
			if strings.TrimSpace(form.CreditCode) == "" {
				return &ClubStepValidateResult{CanNext: false, Message: "请填写统一社会信用代码", ErrorField: "credit_code"}, nil
			}
			if form.BusinessLicense == "" {
				return &ClubStepValidateResult{CanNext: false, Message: "请上传营业执照", ErrorField: "business_license"}, nil
			}
			if strings.TrimSpace(form.CorporateBank) == "" {
				return &ClubStepValidateResult{CanNext: false, Message: "请填写开户行", ErrorField: "corporate_bank"}, nil
			}
			if strings.TrimSpace(form.CorporateAccount) == "" {
				return &ClubStepValidateResult{CanNext: false, Message: "请填写对公账户", ErrorField: "corporate_account"}, nil
			}
			if form.HandleType == "agent" {
				if strings.TrimSpace(form.AgentName) == "" {
					return &ClubStepValidateResult{CanNext: false, Message: "请填写代办人姓名", ErrorField: "agent_name"}, nil
				}
				agentID := ValidateIDCard(form.AgentIDCard)
				if !agentID.Valid {
					return &ClubStepValidateResult{CanNext: false, Message: "代办人身份证号不正确", ErrorField: "agent_id_card"}, nil
				}
			}
		}
		return &ClubStepValidateResult{CanNext: true}, nil

	case 7:
		return &ClubStepValidateResult{CanNext: true}, nil
	}

	return &ClubStepValidateResult{CanNext: false, Message: "未知步骤", ErrorField: "step"}, nil
}

// isAddressComplete 地址7字段完整性判定(后端权威)
func isAddressComplete(form ClubStepForm) bool {
	return form.AddressProvince != "" && form.AddressCity != "" && form.AddressDistrict != "" &&
		form.AddressStreet != "" && form.AddressCommunity != "" && form.AddressBuilding != "" && form.AddressHouseNo != ""
}

// ---------- 活体认证(沙箱) ----------

// LivenessCheckResult 活体认证结果
type LivenessCheckResult struct {
	Status  int8   `json:"status"` // 0=未认证 1=通过 2=失败
	Message string `json:"message"`
}

// ClubLivenessCheck 活体认证(沙箱模式直接返回通过)
// 真实环境应调用微信活体认证 SDK,前端只触发不实现算法
func ClubLivenessCheck(userID int64, clubType string) (*LivenessCheckResult, error) {
	if clubType != "green_v" && clubType != "blue_v" {
		return nil, errors.New("入驻类型不正确")
	}
	if userID <= 0 {
		return nil, errors.New("用户未登录")
	}
	// 沙箱模式:直接返回通过;生产环境应调微信 faceid
	return &LivenessCheckResult{
		Status:  1,
		Message: "认证通过",
	}, nil
}

// ---------- 合同模板下发 ----------

// ContractTemplateInfo 合同模板信息
type ContractTemplateInfo struct {
	URL          string `json:"url"`
	Filename     string `json:"filename"`
	NeedSeal     bool   `json:"need_seal"`     // 是否需要盖章
	SealRemindText string `json:"seal_remind_text"` // 盖章提醒文案(后端下发)
}

// GetContractTemplate 获取合同模板下载地址
// 根据 club_type 返回对应的合同模板;盖章提醒文案后端统一管理
func GetContractTemplate(clubType string) (*ContractTemplateInfo, error) {
	if clubType != "green_v" && clubType != "blue_v" {
		return nil, errors.New("入驻类型不正确")
	}
	filename := "个人俱乐部入驻合同模板.pdf"
	if clubType == "blue_v" {
		filename = "企业俱乐部入驻合同模板.pdf"
	}
	// 沙箱模式:返回占位 URL;生产环境应从 OSS 下发
	return &ContractTemplateInfo{
		URL:            "/static/templates/" + filename,
		Filename:       filename,
		NeedSeal:       clubType == "blue_v",
		SealRemindText: "请确认合同已加盖企业公章",
	}, nil
}

// ---------- 统一提交 ----------

// ClubSubmitResult 入驻提交结果
type ClubSubmitResult struct {
	RegistrationID          int64  `json:"registration_id"`
	ClubType                string `json:"club_type"`
	CorporateAccountMasked  string `json:"corporate_account_masked"` // 后端脱敏后的对公账户(展示用)
	StatusText              string `json:"status_text"`              // 状态文案
	Message                 string `json:"message"`
}

// SubmitClubRegistrationUnified 统一入驻提交
// 内部根据 club_type 分发到 SubmitPersonalRegistration / SubmitEnterpriseRegistration
// 提交前再次执行全步骤校验(防止前端绕过 validate-step)
func SubmitClubRegistrationUnified(userID int64, form ClubStepForm) (*ClubSubmitResult, error) {
	if userID <= 0 {
		return nil, errors.New("用户未登录")
	}
	// 服务端兜底校验:逐步校验所有必填步骤
	for step := 2; step <= 6; step++ {
		formCopy := form
		formCopy.Step = step
		res, err := ValidateClubStep(formCopy)
		if err != nil {
			return nil, err
		}
		if !res.CanNext {
			return nil, fmt.Errorf("步骤%d校验未通过:%s", step, res.Message)
		}
	}

	// 根据类型分发
	if form.ClubType == "green_v" {
		personalForm := PersonalRegistrationForm{
			Name:             form.ClubName,
			Abbreviation:     form.Abbreviation,
			RealName:         form.RealName,
			IDCard:           form.IDCard,
			Phone:            form.Phone,
			IDCardFront:      form.IDCardFront,
			IDCardBack:       form.IDCardBack,
			SelfDeclarationURL: form.ContractFile,
			Address:          buildAddressInfo(form),
			FaceVerifyStatus: form.LivenessStatus,
		}
		reg, err := SubmitPersonalRegistration(userID, personalForm)
		if err != nil {
			return nil, err
		}
		return &ClubSubmitResult{
			RegistrationID: reg.ID,
			ClubType:       "green_v",
			StatusText:     "待审核",
			Message:        "入驻申请已提交",
		}, nil
	}

	// 企业入驻
	enterpriseForm := EnterpriseRegistrationForm{
		Name:               form.ClubName,
		Abbreviation:       form.Abbreviation,
		BusinessLicenseURL: form.BusinessLicense,
		LegalPersonName:    form.RealName,
		LegalPersonIDCard:  form.IDCard,
		LegalPersonIDFront: form.IDCardFront,
		LegalPersonIDBack:  form.IDCardBack,
		ContactPhone:       form.Phone,
		Address:            buildEnterpriseAddress(form),
		BankName:           form.CorporateBank,
		BankAccount:        form.CorporateAccount,
		AgentMode:          form.HandleType == "agent",
		AuthLetterURL:      form.AgentAuthorization,
		AgentIDCardFront:   form.AgentIDCardFront,
		AgentIDCardBack:    form.AgentIDCardBack,
		HasSealReminded:    true,
	}
	reg, err := SubmitEnterpriseRegistration(userID, enterpriseForm)
	if err != nil {
		return nil, err
	}
	return &ClubSubmitResult{
		RegistrationID:         reg.ID,
		ClubType:               "blue_v",
		CorporateAccountMasked: MaskBankAccount(form.CorporateAccount),
		StatusText:             "待审核",
		Message:                "入驻申请已提交",
	}, nil
}

// buildAddressInfo 由 ClubStepForm 构造 utils.AddressInfo(后端组装,前端不参与)
func buildAddressInfo(form ClubStepForm) utils.AddressInfo {
	return utils.AddressInfo{
		Province:    form.AddressProvince,
		City:        form.AddressCity,
		District:    form.AddressDistrict,
		Street:      form.AddressStreet,
		Community:   form.AddressCommunity,
		Building:    form.AddressBuilding,
		HouseNumber: form.AddressHouseNo,
	}
}

// buildEnterpriseAddress 拼接企业完整地址字符串(后端组装)
func buildEnterpriseAddress(form ClubStepForm) string {
	return fmt.Sprintf("%s%s%s%s%s%s%s",
		form.AddressProvince, form.AddressCity, form.AddressDistrict,
		form.AddressStreet, form.AddressCommunity, form.AddressBuilding, form.AddressHouseNo)
}

// ---------- 草稿过期清理(辅助) ----------

// CleanupExpiredDrafts 清理过期入驻草稿(定时任务调用)
func CleanupExpiredDrafts() error {
	now := time.Now()
	return db.Where("expire_at < ?", now).Delete(&model.ClubJoinDraft{}).Error
}
