package model

import (
	"time"

	"gorm.io/datatypes"
)

// GlobalGroupConfig 全域官方总群配置 (需求141~170)
type GlobalGroupConfig struct {
	ID                              int64          `gorm:"primaryKey;autoIncrement;column:id" json:"id"`
	GroupChatID                     int64          `gorm:"column:group_chat_id;uniqueIndex:uk_group_id;not null;default:0" json:"group_chat_id"`
	Enabled                         int8           `gorm:"column:enabled;not null;default:1" json:"enabled"`
	NicknameLock                    int8           `gorm:"column:nickname_lock;not null;default:1" json:"nickname_lock"`
	AtAllDailyLimit                 int            `gorm:"column:at_all_daily_limit;not null;default:3" json:"at_all_daily_limit"`
	AtAllUsedToday                  int            `gorm:"column:at_all_used_today;not null;default:0" json:"at_all_used_today"`
	AtAllResetDate                  *time.Time     `gorm:"column:at_all_reset_date" json:"at_all_reset_date"`
	MuteMode                        int8           `gorm:"column:mute_mode;not null;default:0" json:"mute_mode"`
	SinglePageLoadCount             int            `gorm:"column:single_page_load_count;not null;default:60" json:"single_page_load_count"`
	MuteModeAllowsOnlyTransferPush  int8           `gorm:"column:mute_mode_allows_only_transfer_push;not null;default:1" json:"mute_mode_allows_only_transfer_push"`
	OnlyTransferCardForwarding      int8           `gorm:"column:only_transfer_card_forwarding;not null;default:1" json:"only_transfer_card_forwarding"`
	AutoFoldSilentMinutes           int            `gorm:"column:auto_fold_silent_minutes;not null;default:30" json:"auto_fold_silent_minutes"`
	TopNoticeHTML                   string         `gorm:"column:top_notice_html;type:longtext" json:"top_notice_html"`
	AutoCreatedSystemSessionEnabled int8           `gorm:"column:auto_created_system_session_enabled;not null;default:1" json:"auto_created_system_session_enabled"`
	TransferCardPrefixName          string         `gorm:"column:transfer_card_prefix_name;size:64;not null;default:'转单'" json:"transfer_card_prefix_name"`
	OfficialBubbleBgOverride        string         `gorm:"column:official_bubble_bg_override;size:64;not null;default:''" json:"official_bubble_bg_override"`
	TransferBubbleBgOverride        string         `gorm:"column:transfer_bubble_bg_override;size:64;not null;default:'#FFF7E6'" json:"transfer_bubble_bg_override"`
	AdminShieldList                 datatypes.JSON `gorm:"column:admin_shield_list_json" json:"admin_shield_list"`
	NoticeReads                     datatypes.JSON `gorm:"column:notice_reads_json" json:"notice_reads"`
	CreatedAt                       *time.Time     `gorm:"column:created_at" json:"created_at"`
	UpdatedAt                       *time.Time     `gorm:"column:updated_at" json:"updated_at"`
}
func (GlobalGroupConfig) TableName() string { return "global_group_config" }

// AfterSaleMessageMark 售后消息举证标记 (需求171~195)
type AfterSaleMessageMark struct {
	ID                  int64      `gorm:"primaryKey;autoIncrement;column:id" json:"id"`
	AfterSaleSessionID  int64      `gorm:"column:after_sale_session_id;index:idx_session;not null;default:0" json:"after_sale_session_id"`
	MessageID           int64      `gorm:"column:message_id;index:idx_msg;not null;default:0" json:"message_id"`
	MarkType            string     `gorm:"column:mark_type;size:32;not null;default:'evidence'" json:"mark_type"`
	MarkerUID           int64      `gorm:"column:marker_uid;not null;default:0" json:"marker_uid"`
	MarkTime            *time.Time `gorm:"column:mark_time" json:"mark_time"`
	Remark              string     `gorm:"column:remark;size:255;not null;default:''" json:"remark"`
}
func (AfterSaleMessageMark) TableName() string { return "after_sale_message_marks" }

// GlobalClubJoinSwitch 入驻总开关 (需求216)
type GlobalClubJoinSwitch struct {
	ID        int64      `gorm:"primaryKey;autoIncrement;column:id" json:"id"`
	SwitchKey string     `gorm:"column:switch_key;uniqueIndex:uk_switch_key;size:32;not null;default:'club_join'" json:"switch_key"`
	Enabled   int8       `gorm:"column:enabled;not null;default:1" json:"enabled"`
	UpdatedBy int64      `gorm:"column:updated_by;not null;default:0" json:"updated_by"`
	UpdatedAt *time.Time `gorm:"column:updated_at" json:"updated_at"`
}
func (GlobalClubJoinSwitch) TableName() string { return "global_club_join_switch" }

// PersonalClubRegistrationFiles 个人入驻二进制档案 (需求220~235)
type PersonalClubRegistrationFiles struct {
	ID               int64      `gorm:"primaryKey;autoIncrement;column:id" json:"id"`
	PersonalRegID    int64      `gorm:"column:personal_reg_id;uniqueIndex:uk_personal_reg;not null;default:0" json:"personal_reg_id"`
	ApplicantName    string     `gorm:"column:applicant_name;size:64;not null;default:''" json:"applicant_name"`
	IDCardNo         string     `gorm:"column:id_card_no;size:32;not null;default:''" json:"id_card_no"`
	ContactAddress   string     `gorm:"column:contact_address;size:512;not null;default:''" json:"contact_address"`
	IDCardFrontBin   []byte     `gorm:"column:id_card_front_bin;type:longblob" json:"-"` // 二进制不落明文
	IDCardBackBin    []byte     `gorm:"column:id_card_back_bin;type:longblob" json:"-"`
	ContractSignedPDF []byte    `gorm:"column:contract_signed_pdf_bin;type:longblob" json:"-"`
	ApplicantAgeCheck int8      `gorm:"column:applicant_age_check;not null;default:0" json:"applicant_age_check"`
	CreatedAt        *time.Time `gorm:"column:created_at" json:"created_at"`
	UpdatedAt        *time.Time `gorm:"column:updated_at" json:"updated_at"`
}
func (PersonalClubRegistrationFiles) TableName() string { return "personal_club_registration_files" }

// EnterpriseClubRegistrationFiles 企业入驻档案 + 对公打款验证 (需求236~255)
type EnterpriseClubRegistrationFiles struct {
	ID                            int64      `gorm:"primaryKey;autoIncrement;column:id" json:"id"`
	EnterpriseRegID               int64      `gorm:"column:enterprise_reg_id;uniqueIndex:uk_ent_reg;not null;default:0" json:"enterprise_reg_id"`
	BusinessLicenseBin            []byte     `gorm:"column:business_license_bin;type:longblob" json:"-"`
	LegalPersonFaceVerified       int8       `gorm:"column:legal_person_face_verified;not null;default:0" json:"legal_person_face_verified"`
	AgentAuthContractPDFBin       []byte     `gorm:"column:agent_auth_contract_pdf_bin;type:longblob" json:"-"`
	EnterpriseContractSignedPDFBin []byte    `gorm:"column:enterprise_contract_signed_pdf_bin;type:longblob" json:"-"`
	BankAccountName               string     `gorm:"column:bank_account_name;size:128;not null;default:''" json:"bank_account_name"`
	BankCardNo                    string     `gorm:"column:bank_card_no;size:64;not null;default:''" json:"bank_card_no"`
	RandomAmount                  float64    `gorm:"column:random_amount;type:decimal(4,1);not null;default:0.0" json:"random_amount"`
	AmountVerifyStatus            int8       `gorm:"column:amount_verify_status;not null;default:0" json:"amount_verify_status"`
	VerifyRemark                  string     `gorm:"column:verify_remark;size:255;not null;default:''" json:"verify_remark"`
	VerifyExpireAt                *time.Time `gorm:"column:verify_expire_at" json:"verify_expire_at"`
	VerifyCreatedAt               *time.Time `gorm:"column:verify_created_at" json:"verify_created_at"`
	RefundedAt                    *time.Time `gorm:"column:refunded_at" json:"refunded_at"`
}
func (EnterpriseClubRegistrationFiles) TableName() string { return "enterprise_club_registration_files" }

// OrderSeqClub 俱乐部订单号序列 (需求262~263)
type OrderSeqClub struct {
	ID        int64  `gorm:"primaryKey;autoIncrement;column:id" json:"id"`
	ClubID    int64  `gorm:"column:club_id;uniqueIndex:uk_club_date;not null;default:0" json:"club_id"`
	OrderDate string `gorm:"column:order_date;uniqueIndex:uk_club_date;size:8;not null;default:''" json:"order_date"`
	DailySeq  int64  `gorm:"column:daily_seq;not null;default:0" json:"daily_seq"`
}
func (OrderSeqClub) TableName() string { return "order_seq_club" }

// OrderNoGenerateLog 订单号唯一日志 (需求479~480)
type OrderNoGenerateLog struct {
	ID             int64      `gorm:"primaryKey;autoIncrement;column:id" json:"id"`
	OrderNo        string     `gorm:"column:order_no;uniqueIndex:uk_order_no;size:64;not null;default:''" json:"order_no"`
	ClubID         int64      `gorm:"column:club_id;not null;default:0" json:"club_id"`
	ClubAbbr       string     `gorm:"column:club_abbr;size:16;not null;default:''" json:"club_abbr"`
	YYYYMMDDHHMI   string     `gorm:"column:yyyymmddhhmi;size:12;not null;default:''" json:"yyyymmddhhmi"`
	DailySeq       int64      `gorm:"column:daily_seq;not null;default:0" json:"daily_seq"`
	CreatedAt      *time.Time `gorm:"column:created_at" json:"created_at"`
}
func (OrderNoGenerateLog) TableName() string { return "order_no_generate_logs" }

// OrderTransferCard 跨俱乐部转单卡片 (需求144/145/151/152/157/164/165/170/484)
type OrderTransferCard struct {
	ID                     int64          `gorm:"primaryKey;autoIncrement;column:id" json:"id"`
	CardBatchID            string         `gorm:"column:card_batch_id;size:64;not null;default:''" json:"card_batch_id"`
	CardType               string         `gorm:"column:card_type;size:16;not null;default:'public'" json:"card_type"`
	FromClubID             int64          `gorm:"column:from_club_id;index:idx_from_club;not null;default:0" json:"from_club_id"`
	ToClubID               int64          `gorm:"column:to_club_id;index:idx_to_club;not null;default:0" json:"to_club_id"`
	OrderIDs               datatypes.JSON `gorm:"column:order_ids_json;not null" json:"order_ids"`
	ServiceType            string         `gorm:"column:service_type;size:32;not null;default:''" json:"service_type"`
	PriceTotal             float64        `gorm:"column:price_total;type:decimal(14,2);not null;default:0.00" json:"price_total"`
	PriceHideAfterTaken    int8           `gorm:"column:price_hide_after_taken;not null;default:1" json:"price_hide_after_taken"`
	CardStatus             int8           `gorm:"column:card_status;index:idx_status_expire;not null;default:0" json:"card_status"`
	ValidHours             int            `gorm:"column:valid_hours;not null;default:24" json:"valid_hours"`
	ExpireAt               *time.Time     `gorm:"column:expire_at;index:idx_status_expire" json:"expire_at"`
	CreatedAt              *time.Time     `gorm:"column:created_at" json:"created_at"`
	TakenByClubID          int64          `gorm:"column:taken_by_club_id;not null;default:0" json:"taken_by_club_id"`
	TakenAt                *time.Time     `gorm:"column:taken_at" json:"taken_at"`
	GrayOutAfterExpire     int8           `gorm:"column:gray_out_after_expire;not null;default:1" json:"gray_out_after_expire"`
	SenderReceiptSent      int8           `gorm:"column:sender_receipt_sent;not null;default:0" json:"sender_receipt_sent"`
	HandledByGroupChatID   int64          `gorm:"column:handled_by_group_chat_id;not null;default:0" json:"handled_by_group_chat_id"`
	HandledMsgID           int64          `gorm:"column:handled_msg_id;not null;default:0" json:"handled_msg_id"`
}
func (OrderTransferCard) TableName() string { return "order_transfer_cards" }

// ExportWatermarkLog 隐形水印导出 + 区块链存证 (需求461~485)
type ExportWatermarkLog struct {
	ID                    int64          `gorm:"primaryKey;autoIncrement;column:id" json:"id"`
	ExportNo              string         `gorm:"column:export_no;index:idx_export_no;size:64;not null;default:''" json:"export_no"`
	ExporterUID           int64          `gorm:"column:exporter_uid;index:idx_exporter;not null;default:0" json:"exporter_uid"`
	ExporterName          string         `gorm:"column:exporter_name;size:64;not null;default:''" json:"exporter_name"`
	ExporterRole          string         `gorm:"column:exporter_role;size:32;not null;default:''" json:"exporter_role"`
	ExporterDeviceID      string         `gorm:"column:exporter_device_id;size:128;not null;default:''" json:"exporter_device_id"`
	ExporterDeviceModel   string         `gorm:"column:exporter_device_model;size:128;not null;default:''" json:"exporter_device_model"`
	ExportMilliUTC        int64          `gorm:"column:export_milli_utc;not null;default:0" json:"export_milli_utc"`
	ExportLocation        string         `gorm:"column:export_location;size:255;not null;default:''" json:"export_location"`
	ExportFilterSummary   string         `gorm:"column:export_filter_summary;size:512;not null;default:''" json:"export_filter_summary"`
	ExportScope           string         `gorm:"column:export_scope;size:32;not null;default:'admin'" json:"export_scope"`
	EncryptedLevel        int8           `gorm:"column:encrypted_level;not null;default:0" json:"encrypted_level"`
	AuthorizedReason      string         `gorm:"column:authorized_reason;size:255;not null;default:''" json:"authorized_reason"`
	AuthorizedOTPVerified int8           `gorm:"column:authorized_otp_verified;not null;default:0" json:"authorized_otp_verified"`
	OriginHashSHA256      string         `gorm:"column:origin_hash_sha256;size:128;not null;default:''" json:"origin_hash_sha256"`
	BlockchainTxid        string         `gorm:"column:blockchain_txid;size:255;not null;default:''" json:"blockchain_txid"`
	BlockchainTimestamp   int64          `gorm:"column:blockchain_timestamp;not null;default:0" json:"blockchain_timestamp"`
	FileNameSuffixTS      string         `gorm:"column:file_name_suffix_ts;size:32;not null;default:''" json:"file_name_suffix_ts"`
	WatermarkEnabled      int8           `gorm:"column:watermark_enabled;not null;default:1" json:"watermark_enabled"`
	WatermarkTemplate     datatypes.JSON `gorm:"column:watermark_template_json" json:"watermark_template"`
	DecryptLogID          int64          `gorm:"column:decrypt_log_id;not null;default:0" json:"decrypt_log_id"`
	BatchZipGenerated     int8           `gorm:"column:batch_zip_generated;not null;default:0" json:"batch_zip_generated"`
	ExportStatus          int8           `gorm:"column:export_status;not null;default:0" json:"export_status"`
	CreatedAt             *time.Time     `gorm:"column:created_at" json:"created_at"`
}
func (ExportWatermarkLog) TableName() string { return "export_watermark_logs" }

// WatermarkDetectLog 水印检测记录
type WatermarkDetectLog struct {
	ID               int64          `gorm:"primaryKey;autoIncrement;column:id" json:"id"`
	OperatorUID      int64          `gorm:"column:operator_uid;index:idx_op;not null;default:0" json:"operator_uid"`
	SourceType       string         `gorm:"column:source_type;size:16;not null;default:'image'" json:"source_type"`
	DetectedInfoJSON datatypes.JSON `gorm:"column:detected_info_json" json:"detected_info_json"`
	CreatedAt        *time.Time     `gorm:"column:created_at" json:"created_at"`
}
func (WatermarkDetectLog) TableName() string { return "watermark_detect_logs" }

// BadgeRenderConfig V 标可视化渲染配置 + 禁止前端伪造控制
type BadgeRenderConfig struct {
	ID                         int64      `gorm:"primaryKey;autoIncrement;column:id" json:"id"`
	BadgeKey                   string     `gorm:"column:badge_key;uniqueIndex:uk_badge_key;size:32;not null;default:''" json:"badge_key"`
	BadgeName                  string     `gorm:"column:badge_name;size:32;not null;default:''" json:"badge_name"`
	SizeRatioVSFont            float64    `gorm:"column:size_ratio_vs_font;type:decimal(4,2);not null;default:1.00" json:"size_ratio_vs_font"`
	TooltipText                string     `gorm:"column:tooltip_text;size:128;not null;default:''" json:"tooltip_text"`
	DisplayPriority            int        `gorm:"column:display_priority;not null;default:0" json:"display_priority"`
	PlatformOnlyCreation       int8       `gorm:"column:platform_only_creation;not null;default:1" json:"platform_only_creation"`
	AllowClubAdminVisible      int8       `gorm:"column:allow_club_admin_visible;not null;default:1" json:"allow_club_admin_visible"`
	DisableFrontendForgeryProtect int8    `gorm:"column:disable_frontend_forgery_protect;not null;default:1" json:"disable_frontend_forgery_protect"`
	HideOnBlacklist            int8       `gorm:"column:hide_on_blacklist;not null;default:1" json:"hide_on_blacklist"`
	GrayOnClubSuspend          int8       `gorm:"column:gray_on_club_suspend;not null;default:1" json:"gray_on_club_suspend"`
	PreserveInHistoryOrders    int8       `gorm:"column:preserve_in_history_orders;not null;default:1" json:"preserve_in_history_orders"`
	AttachToClubEntityOnly     int8       `gorm:"column:attach_to_club_entity_only;not null;default:1" json:"attach_to_club_entity_only"`
	RenderInSession            int8       `gorm:"column:render_in_session;not null;default:1" json:"render_in_session"`
	RenderInChatWindow         int8       `gorm:"column:render_in_chat_window;not null;default:1" json:"render_in_chat_window"`
	RenderInMemberList         int8       `gorm:"column:render_in_member_list;not null;default:1" json:"render_in_member_list"`
	TransferCardDisplay       int8       `gorm:"column:transfer_card_display;not null;default:1" json:"transfer_card_display"`
	CreatedAt                  *time.Time `gorm:"column:created_at" json:"created_at"`
	UpdatedAt                  *time.Time `gorm:"column:updated_at" json:"updated_at"`
}
func (BadgeRenderConfig) TableName() string { return "badge_render_config" }

// PlatformGlobalParam 全局阈值参数
type PlatformGlobalParam struct {
	ID          int64      `gorm:"primaryKey;autoIncrement;column:id" json:"id"`
	ParamKey    string     `gorm:"column:param_key;uniqueIndex:uk_key;size:64;not null;default:''" json:"param_key"`
	ParamValue  string     `gorm:"column:param_value;type:text" json:"param_value"`
	ParamType   string     `gorm:"column:param_type;size:16;not null;default:'string'" json:"param_type"`
	Module      string     `gorm:"column:module;size:64;not null;default:''" json:"module"`
	Description string     `gorm:"column:description;size:255;not null;default:''" json:"description"`
	UpdatedAt   *time.Time `gorm:"column:updated_at" json:"updated_at"`
}
func (PlatformGlobalParam) TableName() string { return "platform_global_params" }

// ChatTypingStatus 对方正在输入状态
type ChatTypingStatus struct {
	ID         int64      `gorm:"primaryKey;autoIncrement;column:id" json:"id"`
	SessionID  int64      `gorm:"column:session_id;uniqueIndex:uk_sess_uid;not null;default:0" json:"session_id"`
	UID        int64      `gorm:"column:uid;uniqueIndex:uk_sess_uid;not null;default:0" json:"uid"`
	TypingType string     `gorm:"column:typing_type;size:16;not null;default:'text'" json:"typing_type"`
	StartedAt  *time.Time `gorm:"column:started_at" json:"started_at"`
	ExpireAt   *time.Time `gorm:"column:expire_at" json:"expire_at"`
}
func (ChatTypingStatus) TableName() string { return "chat_typing_status" }

// IMUserPreference 用户 IM 个性化偏好
type IMUserPreference struct {
	ID                 int64          `gorm:"primaryKey;autoIncrement;column:id" json:"id"`
	UID                int64          `gorm:"column:uid;uniqueIndex:uk_uid;not null;default:0" json:"uid"`
	ListPreviewRows    int8           `gorm:"column:list_preview_rows;not null;default:1" json:"list_preview_rows"`
	SortMode           int8           `gorm:"column:sort_mode;not null;default:0" json:"sort_mode"`
	MuteLevelAllSession int8          `gorm:"column:mute_level_all_session;not null;default:0" json:"mute_level_all_session"`
	IdleSessionDays    int            `gorm:"column:idle_session_days;not null;default:30" json:"idle_session_days"`
	CustomGroups       datatypes.JSON `gorm:"column:custom_groups_json" json:"custom_groups"`
	ScrollPosMap       datatypes.JSON `gorm:"column:scroll_pos_map_json" json:"scroll_pos_map"`
	AutoContVoice      int8           `gorm:"column:auto_cont_voice;not null;default:0" json:"auto_cont_voice"`
	ShortPhrase        datatypes.JSON `gorm:"column:short_phrase_json" json:"short_phrase"`
	LastReadTipStyle   int8           `gorm:"column:last_read_tip_style;not null;default:0" json:"last_read_tip_style"`
	StarredForeverTop  int8           `gorm:"column:starred_forever_top;not null;default:1" json:"starred_forever_top"`
	UpdatedAt          *time.Time     `gorm:"column:updated_at" json:"updated_at"`
}
func (IMUserPreference) TableName() string { return "im_user_preferences" }

// UserFavorite 统一收藏表
type UserFavorite struct {
	ID            int64          `gorm:"primaryKey;autoIncrement;column:id" json:"id"`
	UID           int64          `gorm:"column:uid;uniqueIndex:uk_uid_type_target;index:idx_uid_type;not null;default:0" json:"uid"`
	FavType       string         `gorm:"column:fav_type;uniqueIndex:uk_uid_type_target;index:idx_uid_type;size:32;not null;default:''" json:"fav_type"`
	TargetID      int64          `gorm:"column:target_id;uniqueIndex:uk_uid_type_target;not null;default:0" json:"target_id"`
	ExtraDataJSON datatypes.JSON `gorm:"column:extra_data_json" json:"extra_data_json"`
	CreatedAt     *time.Time     `gorm:"column:created_at" json:"created_at"`
}
func (UserFavorite) TableName() string { return "user_favorites" }

// Announcement 全链路公告
type Announcement struct {
	ID            int64          `gorm:"primaryKey;autoIncrement;column:id" json:"id"`
	AnnType       string         `gorm:"column:ann_type;index:idx_type_club;size:16;not null;default:'platform'" json:"ann_type"`
	PublisherUID  int64          `gorm:"column:publisher_uid;not null;default:0" json:"publisher_uid"`
	ClubID        int64          `gorm:"column:club_id;index:idx_type_club;not null;default:0" json:"club_id"`
	PushScope     string         `gorm:"column:push_scope;size:32;not null;default:'all'" json:"push_scope"`
	Title         string         `gorm:"column:title;size:255;not null;default:''" json:"title"`
	ContentHTML   string         `gorm:"column:content_html;type:longtext" json:"content_html"`
	CoverImages   datatypes.JSON `gorm:"column:cover_images_json" json:"cover_images"`
	EffectiveFrom *time.Time     `gorm:"column:effective_from;index:idx_status_time" json:"effective_from"`
	EffectiveTo   *time.Time     `gorm:"column:effective_to" json:"effective_to"`
	Pinned        int8           `gorm:"column:pinned;not null;default:0" json:"pinned"`
	SortOrder     int            `gorm:"column:sort_order;not null;default:0" json:"sort_order"`
	Status        int8           `gorm:"column:status;index:idx_status_time;not null;default:1" json:"status"`
	MaxCharFold   int            `gorm:"column:max_char_fold;not null;default:300" json:"max_char_fold"`
	ReadsJSON     datatypes.JSON `gorm:"column:reads_json" json:"reads_json"`
	CreatedAt     *time.Time     `gorm:"column:created_at" json:"created_at"`
	UpdatedAt     *time.Time     `gorm:"column:updated_at" json:"updated_at"`
}
func (Announcement) TableName() string { return "announcements" }
