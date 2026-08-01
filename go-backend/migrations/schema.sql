-- ============================================================
-- 戟三电竞平台 - MySQL 8.0 数据库设计文件
-- 技术栈：Go(Gin) + GORM + MySQL 8.0 + Redis 7.0
-- 生成时间：2026-07-23
-- 所有金额字段均为 BIGINT，单位：分
-- ============================================================
SET NAMES utf8mb4;
SET FOREIGN_KEY_CHECKS = 0;

-- ============================================================
-- 模块1: 用户体系
-- ============================================================

-- 用户表
DROP TABLE IF EXISTS `users`;
CREATE TABLE `users` (
  `id` BIGINT NOT NULL AUTO_INCREMENT COMMENT '主键ID',
  `openid` VARCHAR(64) NOT NULL DEFAULT '' COMMENT '微信openid',
  `unionid` VARCHAR(64) NOT NULL DEFAULT '' COMMENT '微信unionid',
  `phone` VARCHAR(20) NOT NULL DEFAULT '' COMMENT '手机号',
  `nickname` VARCHAR(64) NOT NULL DEFAULT '' COMMENT '昵称',
  `avatar` VARCHAR(512) NOT NULL DEFAULT '' COMMENT '头像URL',
  `role` TINYINT NOT NULL DEFAULT 1 COMMENT '角色 1=客户 2=打手 4=分销商 8=派单员(位运算)',
  `invite_code` VARCHAR(64) NOT NULL DEFAULT '' COMMENT '绑定的邀请码',
  `club_id` BIGINT NOT NULL DEFAULT 0 COMMENT '所属俱乐部ID',
  `credit_score` INT NOT NULL DEFAULT 100 COMMENT '信用分 初始100',
  `is_realname` TINYINT NOT NULL DEFAULT 0 COMMENT '是否已实名 0否 1是',
  `is_minor` TINYINT NOT NULL DEFAULT 0 COMMENT '是否未成年 0否 1是',
  `real_name` VARCHAR(32) NOT NULL DEFAULT '' COMMENT '真实姓名',
  `id_card` VARCHAR(18) NOT NULL DEFAULT '' COMMENT '身份证号',
  `is_phone_abandoned` TINYINT NOT NULL DEFAULT 0 COMMENT '手机号是否被二次放号回收 0否 1是',
  `status` TINYINT NOT NULL DEFAULT 1 COMMENT '状态 1=正常 0=封禁',
  `balance` BIGINT NOT NULL DEFAULT 0 COMMENT '余额(分)',
  `points` INT NOT NULL DEFAULT 0 COMMENT '积分',
  `created_at` DATETIME NULL DEFAULT NULL COMMENT '创建时间',
  `updated_at` DATETIME NULL DEFAULT NULL COMMENT '更新时间',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_openid` (`openid`),
  KEY `idx_phone` (`phone`),
  KEY `idx_role` (`role`),
  KEY `idx_status` (`status`),
  KEY `idx_club_id` (`club_id`),
  KEY `idx_invite_code` (`invite_code`),
  KEY `idx_created_at` (`created_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='用户表';

-- 活体校验缓存表
DROP TABLE IF EXISTS `realname_caches`;
CREATE TABLE `realname_caches` (
  `id` BIGINT NOT NULL AUTO_INCREMENT COMMENT '主键ID',
  `user_id` BIGINT NOT NULL DEFAULT 0 COMMENT '用户ID',
  `last_verify_time` DATETIME NULL DEFAULT NULL COMMENT '最后验证时间',
  `expire_time` DATETIME NULL DEFAULT NULL COMMENT '缓存过期时间',
  `created_at` DATETIME NULL DEFAULT NULL COMMENT '创建时间',
  `updated_at` DATETIME NULL DEFAULT NULL COMMENT '更新时间',
  PRIMARY KEY (`id`),
  KEY `idx_user_id` (`user_id`),
  KEY `idx_expire_time` (`expire_time`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='活体校验缓存表';

-- 活体检测频控表
DROP TABLE IF EXISTS `face_verify_rate_limits`;
CREATE TABLE `face_verify_rate_limits` (
  `id` BIGINT NOT NULL AUTO_INCREMENT COMMENT '主键ID',
  `user_id` BIGINT NOT NULL DEFAULT 0 COMMENT '用户ID',
  `ip` VARCHAR(64) NOT NULL DEFAULT '' COMMENT 'IP地址',
  `count` INT NOT NULL DEFAULT 0 COMMENT '当日次数',
  `date` DATE NOT NULL COMMENT '日期',
  `created_at` DATETIME NULL DEFAULT NULL COMMENT '创建时间',
  `updated_at` DATETIME NULL DEFAULT NULL COMMENT '更新时间',
  PRIMARY KEY (`id`),
  KEY `idx_user_date` (`user_id`, `date`),
  KEY `idx_ip` (`ip`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='活体检测频控表';

-- 监护人验证表
DROP TABLE IF EXISTS `guardian_verifies`;
CREATE TABLE `guardian_verifies` (
  `id` BIGINT NOT NULL AUTO_INCREMENT COMMENT '主键ID',
  `user_id` BIGINT NOT NULL DEFAULT 0 COMMENT '未成年用户ID',
  `guardian_name` VARCHAR(64) NOT NULL DEFAULT '' COMMENT '监护人姓名',
  `guardian_id_card` VARCHAR(18) NOT NULL DEFAULT '' COMMENT '监护人身份证号',
  `face_token` VARCHAR(255) NOT NULL DEFAULT '' COMMENT '人脸特征token',
  `face_session_id` VARCHAR(64) NOT NULL DEFAULT '' COMMENT '活体会话ID',
  `expire_time` DATETIME NULL DEFAULT NULL COMMENT '验证过期时间',
  `order_id` BIGINT NOT NULL DEFAULT 0 COMMENT '关联订单ID',
  `created_at` DATETIME NULL DEFAULT NULL COMMENT '创建时间',
  PRIMARY KEY (`id`),
  KEY `idx_user_id` (`user_id`),
  KEY `idx_order_id` (`order_id`),
  KEY `idx_created_at` (`created_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='监护人验证表';

-- 电子签名表
DROP TABLE IF EXISTS `electronic_signatures`;
CREATE TABLE `electronic_signatures` (
  `id` BIGINT NOT NULL AUTO_INCREMENT COMMENT '主键ID',
  `order_id` BIGINT NOT NULL DEFAULT 0 COMMENT '订单ID',
  `user_id` BIGINT NOT NULL DEFAULT 0 COMMENT '用户ID',
  `face_session_id` VARCHAR(64) NOT NULL DEFAULT '' COMMENT '活体会话ID',
  `signer_name` VARCHAR(64) NOT NULL DEFAULT '' COMMENT '签署人姓名',
  `signer_id_card` VARCHAR(18) NOT NULL DEFAULT '' COMMENT '签署人身份证号',
  `amount_input` BIGINT NOT NULL DEFAULT 0 COMMENT '金额确认输入(分)',
  `provider` VARCHAR(32) NOT NULL DEFAULT '' COMMENT '服务商 法大大/腾讯电子签',
  `sign_serial_no` VARCHAR(128) NOT NULL DEFAULT '' COMMENT '签署流水号',
  `contract_hash` VARCHAR(128) NOT NULL DEFAULT '' COMMENT '合同哈希',
  `status` TINYINT NOT NULL DEFAULT 0 COMMENT '签署状态 0待签 1已签 2失败',
  `created_at` DATETIME NULL DEFAULT NULL COMMENT '创建时间',
  PRIMARY KEY (`id`),
  KEY `idx_order_id` (`order_id`),
  KEY `idx_user_id` (`user_id`),
  KEY `idx_status` (`status`),
  KEY `idx_created_at` (`created_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='电子签名表';

-- 手机号二次放号申诉表
DROP TABLE IF EXISTS `phone_appeals`;
CREATE TABLE `phone_appeals` (
  `id` BIGINT NOT NULL AUTO_INCREMENT COMMENT '主键ID',
  `user_id` BIGINT NOT NULL DEFAULT 0 COMMENT '用户ID',
  `phone` VARCHAR(20) NOT NULL DEFAULT '' COMMENT '申诉手机号',
  `video_url` VARCHAR(512) NOT NULL DEFAULT '' COMMENT '申诉视频URL',
  `network_time` VARCHAR(64) NOT NULL DEFAULT '' COMMENT '入网时间',
  `first_use_date` DATE NULL DEFAULT NULL COMMENT '首次使用日期',
  `commitment_text` TEXT COMMENT '承诺书内容',
  `status` TINYINT NOT NULL DEFAULT 1 COMMENT '状态 1待审核 2通过 3驳回',
  `reject_count` INT NOT NULL DEFAULT 0 COMMENT '驳回次数',
  `created_at` DATETIME NULL DEFAULT NULL COMMENT '创建时间',
  `updated_at` DATETIME NULL DEFAULT NULL COMMENT '更新时间',
  PRIMARY KEY (`id`),
  KEY `idx_user_id` (`user_id`),
  KEY `idx_phone` (`phone`),
  KEY `idx_status` (`status`),
  KEY `idx_created_at` (`created_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='手机号二次放号申诉表';

-- ============================================================
-- 模块2: 管理员体系
-- ============================================================

-- 管理员表
DROP TABLE IF EXISTS `admins`;
CREATE TABLE `admins` (
  `id` BIGINT NOT NULL AUTO_INCREMENT COMMENT '主键ID',
  `username` VARCHAR(64) NOT NULL DEFAULT '' COMMENT '管理员账号',
  `password` VARCHAR(255) NOT NULL DEFAULT '' COMMENT '密码(bcrypt加密)',
  `nickname` VARCHAR(64) NOT NULL DEFAULT '' COMMENT '昵称',
  `role` TINYINT NOT NULL DEFAULT 2 COMMENT '角色 1=超级管理员 2=运营 4=财务 8=风控(位运算)',
  `email` VARCHAR(128) NOT NULL DEFAULT '' COMMENT '绑定邮箱',
  `phone` VARCHAR(20) NOT NULL DEFAULT '' COMMENT '手机号',
  `is_init` TINYINT NOT NULL DEFAULT 0 COMMENT '是否完成首次初始化 0否 1是',
  `status` TINYINT NOT NULL DEFAULT 1 COMMENT '状态 1正常 0封禁',
  `last_login_at` DATETIME NULL DEFAULT NULL COMMENT '最后登录时间',
  `last_login_ip` VARCHAR(64) NOT NULL DEFAULT '' COMMENT '最后登录IP',
  `created_at` DATETIME NULL DEFAULT NULL COMMENT '创建时间',
  `updated_at` DATETIME NULL DEFAULT NULL COMMENT '更新时间',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_username` (`username`),
  KEY `idx_role` (`role`),
  KEY `idx_status` (`status`),
  KEY `idx_created_at` (`created_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='管理员表';

-- 管理员密码历史表
DROP TABLE IF EXISTS `admin_password_history`;
CREATE TABLE `admin_password_history` (
  `id` BIGINT NOT NULL AUTO_INCREMENT COMMENT '主键ID',
  `admin_id` BIGINT NOT NULL DEFAULT 0 COMMENT '管理员ID',
  `password_hash` VARCHAR(255) NOT NULL DEFAULT '' COMMENT '历史密码hash',
  `created_at` DATETIME NULL DEFAULT NULL COMMENT '创建时间',
  PRIMARY KEY (`id`),
  KEY `idx_admin_id` (`admin_id`),
  KEY `idx_created_at` (`created_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='管理员密码历史表';

-- 管理员WebAuthn通行密钥表
DROP TABLE IF EXISTS `admin_webauthn`;
CREATE TABLE `admin_webauthn` (
  `id` BIGINT NOT NULL AUTO_INCREMENT COMMENT '主键ID',
  `admin_id` BIGINT NOT NULL DEFAULT 0 COMMENT '管理员ID',
  `credential_id` VARCHAR(255) NOT NULL DEFAULT '' COMMENT '凭证ID',
  `public_key` TEXT COMMENT '公钥',
  `device_info` VARCHAR(255) NOT NULL DEFAULT '' COMMENT '设备信息',
  `created_at` DATETIME NULL DEFAULT NULL COMMENT '创建时间',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_credential_id` (`credential_id`),
  KEY `idx_admin_id` (`admin_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='管理员WebAuthn通行密钥表';

-- 邮箱验证码表
DROP TABLE IF EXISTS `email_verify_codes`;
CREATE TABLE `email_verify_codes` (
  `id` BIGINT NOT NULL AUTO_INCREMENT COMMENT '主键ID',
  `email` VARCHAR(128) NOT NULL DEFAULT '' COMMENT '邮箱',
  `code` VARCHAR(16) NOT NULL DEFAULT '' COMMENT '验证码',
  `type` VARCHAR(32) NOT NULL DEFAULT '' COMMENT '类型 bind/change/reset',
  `expire_time` DATETIME NULL DEFAULT NULL COMMENT '过期时间',
  `created_at` DATETIME NULL DEFAULT NULL COMMENT '创建时间',
  PRIMARY KEY (`id`),
  KEY `idx_email` (`email`),
  KEY `idx_created_at` (`created_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='邮箱验证码表';

-- 初始化日志表
DROP TABLE IF EXISTS `init_logs`;
CREATE TABLE `init_logs` (
  `id` BIGINT NOT NULL AUTO_INCREMENT COMMENT '主键ID',
  `admin_id` BIGINT NOT NULL DEFAULT 0 COMMENT '管理员ID',
  `step` VARCHAR(64) NOT NULL DEFAULT '' COMMENT '初始化步骤',
  `created_at` DATETIME NULL DEFAULT NULL COMMENT '创建时间',
  PRIMARY KEY (`id`),
  KEY `idx_admin_id` (`admin_id`),
  KEY `idx_created_at` (`created_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='管理员初始化日志表';

-- 内置管理端账号表
DROP TABLE IF EXISTS `shop_admin_accounts`;
CREATE TABLE `shop_admin_accounts` (
  `id` BIGINT NOT NULL AUTO_INCREMENT COMMENT '主键ID',
  `username` VARCHAR(64) NOT NULL DEFAULT '' COMMENT '账号 wxadmin_缩写序号',
  `password` VARCHAR(255) NOT NULL DEFAULT '' COMMENT '密码(bcrypt加密)',
  `club_id` BIGINT NOT NULL DEFAULT 0 COMMENT '俱乐部ID',
  `role` TINYINT NOT NULL DEFAULT 2 COMMENT '角色 1=创始人 2=管理员',
  `real_name` VARCHAR(32) NOT NULL DEFAULT '' COMMENT '真实姓名',
  `phone` VARCHAR(20) NOT NULL DEFAULT '' COMMENT '手机号',
  `status` TINYINT NOT NULL DEFAULT 1 COMMENT '状态 1正常 0封禁',
  `last_login_at` DATETIME NULL DEFAULT NULL COMMENT '最后登录时间',
  `last_login_ip` VARCHAR(64) NOT NULL DEFAULT '' COMMENT '最后登录IP',
  `created_at` DATETIME NULL DEFAULT NULL COMMENT '创建时间',
  `updated_at` DATETIME NULL DEFAULT NULL COMMENT '更新时间',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_username` (`username`),
  KEY `idx_club_id` (`club_id`),
  KEY `idx_status` (`status`),
  KEY `idx_created_at` (`created_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='内置管理端账号表';

-- ============================================================
-- 模块3: 俱乐部体系
-- ============================================================

-- 俱乐部表
DROP TABLE IF EXISTS `clubs`;
CREATE TABLE `clubs` (
  `id` BIGINT NOT NULL AUTO_INCREMENT COMMENT '主键ID',
  `name` VARCHAR(128) NOT NULL DEFAULT '' COMMENT '俱乐部名称',
  `abbreviation` VARCHAR(10) NOT NULL DEFAULT '' COMMENT '缩写(唯一封存)',
  `type` TINYINT NOT NULL DEFAULT 2 COMMENT '类型 1=企业 2=个人',
  `status` TINYINT NOT NULL DEFAULT 0 COMMENT '状态 0=审核中 1=审核通过 2=驳回 3=冻结 4=注销',
  `founder_uid` BIGINT NOT NULL DEFAULT 0 COMMENT '创始人用户ID',
  `v_badge_type` TINYINT NOT NULL DEFAULT 0 COMMENT 'V标类型 0=无 1=蓝V(企业) 2=绿V(个人)',
  `deposit_status` TINYINT NOT NULL DEFAULT 0 COMMENT '保证金状态 0未缴 1已缴 2已退',
  `deposit_amount` BIGINT NOT NULL DEFAULT 0 COMMENT '保证金金额(分)',
  `description` TEXT COMMENT '俱乐部简介',
  `logo` VARCHAR(512) NOT NULL DEFAULT '' COMMENT 'Logo URL',
  `background` VARCHAR(512) NOT NULL DEFAULT '' COMMENT '背景图URL',
  `contact_phone` VARCHAR(20) NOT NULL DEFAULT '' COMMENT '联系电话',
  `contact_wechat` VARCHAR(64) NOT NULL DEFAULT '' COMMENT '联系微信',
  `contact_qq` VARCHAR(32) NOT NULL DEFAULT '' COMMENT '联系QQ',
  `business_hours` VARCHAR(64) NOT NULL DEFAULT '' COMMENT '营业时间',
  `rating` DECIMAL(2,1) NOT NULL DEFAULT 5.0 COMMENT '评分',
  `total_orders` INT NOT NULL DEFAULT 0 COMMENT '总订单数',
  `created_at` DATETIME NULL DEFAULT NULL COMMENT '创建时间',
  `updated_at` DATETIME NULL DEFAULT NULL COMMENT '更新时间',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_abbreviation` (`abbreviation`),
  KEY `idx_status` (`status`),
  KEY `idx_founder_uid` (`founder_uid`),
  KEY `idx_created_at` (`created_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='俱乐部表';

-- 俱乐部成员表
DROP TABLE IF EXISTS `club_members`;
CREATE TABLE `club_members` (
  `id` BIGINT NOT NULL AUTO_INCREMENT COMMENT '主键ID',
  `club_id` BIGINT NOT NULL DEFAULT 0 COMMENT '俱乐部ID',
  `user_id` BIGINT NOT NULL DEFAULT 0 COMMENT '用户ID',
  `role` TINYINT NOT NULL DEFAULT 3 COMMENT '角色 1=创始人 2=管理员 3=打手',
  `joined_at` DATETIME NULL DEFAULT NULL COMMENT '加入时间',
  `status` TINYINT NOT NULL DEFAULT 1 COMMENT '状态 1正常 0已移除',
  `created_at` DATETIME NULL DEFAULT NULL COMMENT '创建时间',
  `updated_at` DATETIME NULL DEFAULT NULL COMMENT '更新时间',
  PRIMARY KEY (`id`),
  KEY `idx_club_id` (`club_id`),
  KEY `idx_user_id` (`user_id`),
  KEY `idx_status` (`status`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='俱乐部成员表';

-- 俱乐部缩写封存表
DROP TABLE IF EXISTS `club_abbreviations`;
CREATE TABLE `club_abbreviations` (
  `id` BIGINT NOT NULL AUTO_INCREMENT COMMENT '主键ID',
  `abbreviation` VARCHAR(10) NOT NULL DEFAULT '' COMMENT '缩写',
  `club_id` BIGINT NOT NULL DEFAULT 0 COMMENT '原俱乐部ID',
  `abandoned_at` DATETIME NULL DEFAULT NULL COMMENT '封存时间',
  `created_at` DATETIME NULL DEFAULT NULL COMMENT '创建时间',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_abbreviation` (`abbreviation`),
  KEY `idx_club_id` (`club_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='俱乐部缩写封存表';

-- 俱乐部跨服分店表
DROP TABLE IF EXISTS `club_branches`;
CREATE TABLE `club_branches` (
  `id` BIGINT NOT NULL AUTO_INCREMENT COMMENT '主键ID',
  `club_id` BIGINT NOT NULL DEFAULT 0 COMMENT '俱乐部ID',
  `game_id` BIGINT NOT NULL DEFAULT 0 COMMENT '游戏ID',
  `name` VARCHAR(64) NOT NULL DEFAULT '' COMMENT '分店名称',
  `split_ratio` DECIMAL(5,2) NOT NULL DEFAULT 0.00 COMMENT '分账比例%',
  `created_at` DATETIME NULL DEFAULT NULL COMMENT '创建时间',
  `updated_at` DATETIME NULL DEFAULT NULL COMMENT '更新时间',
  PRIMARY KEY (`id`),
  KEY `idx_club_id` (`club_id`),
  KEY `idx_game_id` (`game_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='俱乐部跨服分店表';

-- 俱乐部动态墙表
DROP TABLE IF EXISTS `club_dynamics`;
CREATE TABLE `club_dynamics` (
  `id` BIGINT NOT NULL AUTO_INCREMENT COMMENT '主键ID',
  `club_id` BIGINT NOT NULL DEFAULT 0 COMMENT '俱乐部ID',
  `user_id` BIGINT NOT NULL DEFAULT 0 COMMENT '发布用户ID',
  `content` TEXT COMMENT '动态内容',
  `media_urls` JSON COMMENT '媒体URL数组',
  `created_at` DATETIME NULL DEFAULT NULL COMMENT '创建时间',
  `updated_at` DATETIME NULL DEFAULT NULL COMMENT '更新时间',
  PRIMARY KEY (`id`),
  KEY `idx_club_id` (`club_id`),
  KEY `idx_user_id` (`user_id`),
  KEY `idx_created_at` (`created_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='俱乐部动态墙表';

-- 俱乐部公告表
DROP TABLE IF EXISTS `club_announcements`;
CREATE TABLE `club_announcements` (
  `id` BIGINT NOT NULL AUTO_INCREMENT COMMENT '主键ID',
  `club_id` BIGINT NOT NULL DEFAULT 0 COMMENT '俱乐部ID',
  `title` VARCHAR(128) NOT NULL DEFAULT '' COMMENT '标题',
  `content` TEXT COMMENT '公告内容',
  `created_by` BIGINT NOT NULL DEFAULT 0 COMMENT '创建人ID',
  `created_at` DATETIME NULL DEFAULT NULL COMMENT '创建时间',
  `updated_at` DATETIME NULL DEFAULT NULL COMMENT '更新时间',
  PRIMARY KEY (`id`),
  KEY `idx_club_id` (`club_id`),
  KEY `idx_created_at` (`created_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='俱乐部公告表';

-- 俱乐部内部订单表
DROP TABLE IF EXISTS `club_internal_orders`;
CREATE TABLE `club_internal_orders` (
  `id` BIGINT NOT NULL AUTO_INCREMENT COMMENT '主键ID',
  `club_id` BIGINT NOT NULL DEFAULT 0 COMMENT '俱乐部ID',
  `title` VARCHAR(128) NOT NULL DEFAULT '' COMMENT '标题',
  `description` VARCHAR(500) NOT NULL DEFAULT '' COMMENT '描述',
  `amount` BIGINT NOT NULL DEFAULT 0 COMMENT '金额(分)',
  `status` TINYINT NOT NULL DEFAULT 0 COMMENT '状态 0待接 1已接 2已完成',
  `created_at` DATETIME NULL DEFAULT NULL COMMENT '创建时间',
  `updated_at` DATETIME NULL DEFAULT NULL COMMENT '更新时间',
  PRIMARY KEY (`id`),
  KEY `idx_club_id` (`club_id`),
  KEY `idx_status` (`status`),
  KEY `idx_created_at` (`created_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='俱乐部内部订单表';

-- 俱乐部日统计表
DROP TABLE IF EXISTS `club_daily_stats`;
CREATE TABLE `club_daily_stats` (
  `id` BIGINT NOT NULL AUTO_INCREMENT COMMENT '主键ID',
  `club_id` BIGINT NOT NULL DEFAULT 0 COMMENT '俱乐部ID',
  `date` DATE NOT NULL COMMENT '统计日期',
  `order_count` INT NOT NULL DEFAULT 0 COMMENT '订单数',
  `revenue` BIGINT NOT NULL DEFAULT 0 COMMENT '营收(分)',
  `created_at` DATETIME NULL DEFAULT NULL COMMENT '创建时间',
  PRIMARY KEY (`id`),
  KEY `idx_club_id` (`club_id`),
  KEY `idx_date` (`date`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='俱乐部日统计表';

-- 俱乐部保证金阶梯表
DROP TABLE IF EXISTS `club_deposit_tiers`;
CREATE TABLE `club_deposit_tiers` (
  `id` BIGINT NOT NULL AUTO_INCREMENT COMMENT '主键ID',
  `club_id` BIGINT NOT NULL DEFAULT 0 COMMENT '俱乐部ID',
  `monthly_revenue_threshold` BIGINT NOT NULL DEFAULT 0 COMMENT '月营收门槛(分)',
  `required_deposit` BIGINT NOT NULL DEFAULT 0 COMMENT '所需保证金(分)',
  `current_level` TINYINT NOT NULL DEFAULT 0 COMMENT '当前等级',
  `created_at` DATETIME NULL DEFAULT NULL COMMENT '创建时间',
  `updated_at` DATETIME NULL DEFAULT NULL COMMENT '更新时间',
  PRIMARY KEY (`id`),
  KEY `idx_club_id` (`club_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='俱乐部保证金阶梯表';

-- 俱乐部优惠券表
DROP TABLE IF EXISTS `club_coupons`;
CREATE TABLE `club_coupons` (
  `id` BIGINT NOT NULL AUTO_INCREMENT COMMENT '主键ID',
  `club_id` BIGINT NOT NULL DEFAULT 0 COMMENT '俱乐部ID',
  `template_id` BIGINT NOT NULL DEFAULT 0 COMMENT '优惠券模板ID',
  `name` VARCHAR(64) NOT NULL DEFAULT '' COMMENT '券名称',
  `type` TINYINT NOT NULL DEFAULT 1 COMMENT '类型 1满减 2新人 3折扣',
  `amount` BIGINT NOT NULL DEFAULT 0 COMMENT '优惠金额(分)',
  `min_spend` BIGINT NOT NULL DEFAULT 0 COMMENT '最低消费(分)',
  `count` INT NOT NULL DEFAULT 0 COMMENT '发放总量',
  `used_count` INT NOT NULL DEFAULT 0 COMMENT '已使用数量',
  `expire_at` DATETIME NULL DEFAULT NULL COMMENT '过期时间',
  `created_at` DATETIME NULL DEFAULT NULL COMMENT '创建时间',
  `updated_at` DATETIME NULL DEFAULT NULL COMMENT '更新时间',
  PRIMARY KEY (`id`),
  KEY `idx_club_id` (`club_id`),
  KEY `idx_template_id` (`template_id`),
  KEY `idx_type` (`type`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='俱乐部优惠券表';

-- 用户优惠券领取记录表
DROP TABLE IF EXISTS `club_coupon_users`;
CREATE TABLE `club_coupon_users` (
  `id` BIGINT NOT NULL AUTO_INCREMENT COMMENT '主键ID',
  `coupon_id` BIGINT NOT NULL DEFAULT 0 COMMENT '俱乐部优惠券ID',
  `user_id` BIGINT NOT NULL DEFAULT 0 COMMENT '用户ID',
  `status` TINYINT NOT NULL DEFAULT 1 COMMENT '状态 1未使用 2已使用 3已过期',
  `used_at` DATETIME NULL DEFAULT NULL COMMENT '使用时间',
  `created_at` DATETIME NULL DEFAULT NULL COMMENT '创建时间',
  `updated_at` DATETIME NULL DEFAULT NULL COMMENT '更新时间',
  PRIMARY KEY (`id`),
  KEY `idx_coupon_id` (`coupon_id`),
  KEY `idx_user_id` (`user_id`),
  KEY `idx_status` (`status`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='用户优惠券领取记录表';

-- 俱乐部群聊关联表
DROP TABLE IF EXISTS `club_group_chats`;
CREATE TABLE `club_group_chats` (
  `id` BIGINT NOT NULL AUTO_INCREMENT COMMENT '主键ID',
  `group_id` BIGINT NOT NULL DEFAULT 0 COMMENT '群聊ID',
  `club_id` BIGINT NOT NULL DEFAULT 0 COMMENT '俱乐部ID',
  `group_type` VARCHAR(32) NOT NULL DEFAULT '' COMMENT '群类型 internal/category',
  `category_type` VARCHAR(32) NOT NULL DEFAULT '' COMMENT '分类类型 chat/welfare/aftersale',
  `created_at` DATETIME NULL DEFAULT NULL COMMENT '创建时间',
  `updated_at` DATETIME NULL DEFAULT NULL COMMENT '更新时间',
  PRIMARY KEY (`id`),
  KEY `idx_group_id` (`group_id`),
  KEY `idx_club_id` (`club_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='俱乐部群聊关联表';

-- ============================================================
-- 模块4: 入驻审核体系
-- ============================================================

-- 企业入驻申请表
DROP TABLE IF EXISTS `enterprise_registrations`;
CREATE TABLE `enterprise_registrations` (
  `id` BIGINT NOT NULL AUTO_INCREMENT COMMENT '主键ID',
  `club_id` BIGINT NOT NULL DEFAULT 0 COMMENT '俱乐部ID',
  `business_license_url` VARCHAR(512) NOT NULL DEFAULT '' COMMENT '营业执照URL',
  `legal_person_name` VARCHAR(64) NOT NULL DEFAULT '' COMMENT '法人姓名',
  `legal_person_id_card` VARCHAR(18) NOT NULL DEFAULT '' COMMENT '法人身份证号',
  `legal_person_id_front` VARCHAR(512) NOT NULL DEFAULT '' COMMENT '法人身份证正面URL',
  `legal_person_id_back` VARCHAR(512) NOT NULL DEFAULT '' COMMENT '法人身份证反面URL',
  `contact_phone` VARCHAR(20) NOT NULL DEFAULT '' COMMENT '联系电话',
  `contact_email` VARCHAR(128) NOT NULL DEFAULT '' COMMENT '联系邮箱',
  `address` VARCHAR(255) NOT NULL DEFAULT '' COMMENT '企业地址',
  `bank_name` VARCHAR(64) NOT NULL DEFAULT '' COMMENT '开户行',
  `bank_account` VARCHAR(64) NOT NULL DEFAULT '' COMMENT '银行账号',
  `electronic_license_url` VARCHAR(512) NOT NULL DEFAULT '' COMMENT '电子营业执照URL',
  `auth_letter_url` VARCHAR(512) NOT NULL DEFAULT '' COMMENT '授权书URL',
  `agent_id_card_front` VARCHAR(512) NOT NULL DEFAULT '' COMMENT '代理人身份证正面URL',
  `agent_id_card_back` VARCHAR(512) NOT NULL DEFAULT '' COMMENT '代理人身份证反面URL',
  `status` TINYINT NOT NULL DEFAULT 0 COMMENT '状态 0待审核 1通过 2驳回',
  `reviewer_id` BIGINT NOT NULL DEFAULT 0 COMMENT '审核人ID',
  `reviewed_at` DATETIME NULL DEFAULT NULL COMMENT '审核时间',
  `reject_reason` VARCHAR(255) NOT NULL DEFAULT '' COMMENT '驳回原因',
  `created_at` DATETIME NULL DEFAULT NULL COMMENT '创建时间',
  `updated_at` DATETIME NULL DEFAULT NULL COMMENT '更新时间',
  PRIMARY KEY (`id`),
  KEY `idx_club_id` (`club_id`),
  KEY `idx_status` (`status`),
  KEY `idx_created_at` (`created_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='企业入驻申请表';

-- 个人入驻申请表
DROP TABLE IF EXISTS `personal_registrations`;
CREATE TABLE `personal_registrations` (
  `id` BIGINT NOT NULL AUTO_INCREMENT COMMENT '主键ID',
  `club_id` BIGINT NOT NULL DEFAULT 0 COMMENT '俱乐部ID',
  `real_name` VARCHAR(32) NOT NULL DEFAULT '' COMMENT '真实姓名',
  `id_card` VARCHAR(18) NOT NULL DEFAULT '' COMMENT '身份证号',
  `phone` VARCHAR(20) NOT NULL DEFAULT '' COMMENT '手机号',
  `id_card_front` VARCHAR(512) NOT NULL DEFAULT '' COMMENT '身份证正面URL',
  `id_card_back` VARCHAR(512) NOT NULL DEFAULT '' COMMENT '身份证反面URL',
  `handheld_id_card` VARCHAR(512) NOT NULL DEFAULT '' COMMENT '手持身份证URL',
  `bank_card` VARCHAR(32) NOT NULL DEFAULT '' COMMENT '银行卡号',
  `bank_name` VARCHAR(64) NOT NULL DEFAULT '' COMMENT '开户行',
  `bank_phone` VARCHAR(20) NOT NULL DEFAULT '' COMMENT '银行预留手机号',
  `face_verify_status` TINYINT NOT NULL DEFAULT 0 COMMENT '活体检测状态 0未检测 1通过 2失败',
  `self_declaration_url` VARCHAR(512) NOT NULL DEFAULT '' COMMENT '自我声明URL',
  `status` TINYINT NOT NULL DEFAULT 0 COMMENT '状态 0待审核 1通过 2驳回',
  `reviewer_id` BIGINT NOT NULL DEFAULT 0 COMMENT '审核人ID',
  `reviewed_at` DATETIME NULL DEFAULT NULL COMMENT '审核时间',
  `reject_reason` VARCHAR(255) NOT NULL DEFAULT '' COMMENT '驳回原因',
  `created_at` DATETIME NULL DEFAULT NULL COMMENT '创建时间',
  `updated_at` DATETIME NULL DEFAULT NULL COMMENT '更新时间',
  PRIMARY KEY (`id`),
  KEY `idx_club_id` (`club_id`),
  KEY `idx_status` (`status`),
  KEY `idx_created_at` (`created_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='个人入驻申请表';

-- ============================================================
-- 模块5: 订单体系
-- ============================================================

-- 订单表
DROP TABLE IF EXISTS `orders`;
CREATE TABLE `orders` (
  `id` BIGINT NOT NULL AUTO_INCREMENT COMMENT '主键ID',
  `order_no` VARCHAR(64) NOT NULL DEFAULT '' COMMENT '订单号',
  `type` TINYINT NOT NULL DEFAULT 1 COMMENT '订单类型 1=即时 2=预约 3=车队 4=教学',
  `user_id` BIGINT NOT NULL DEFAULT 0 COMMENT '客户ID',
  `player_id` BIGINT NOT NULL DEFAULT 0 COMMENT '打手ID',
  `club_id` BIGINT NOT NULL DEFAULT 0 COMMENT '俱乐部ID',
  `service_id` BIGINT NOT NULL DEFAULT 0 COMMENT '服务项目ID',
  `amount` BIGINT NOT NULL DEFAULT 0 COMMENT '订单金额(分)',
  `status` TINYINT NOT NULL DEFAULT 0 COMMENT '状态 0=待接单 1=已接单 2=进行中 3=待验收 4=已完成 5=待结算 6=已结算 10=超时取消 11=大额验证失败 12=已退款',
  `pay_status` TINYINT NOT NULL DEFAULT 0 COMMENT '支付状态 0未支付 1已支付 2已退款 3部分退款',
  `paid_at` DATETIME NULL DEFAULT NULL COMMENT '支付时间',
  `accepted_at` DATETIME NULL DEFAULT NULL COMMENT '接单时间',
  `started_at` DATETIME NULL DEFAULT NULL COMMENT '开始服务时间',
  `ended_at` DATETIME NULL DEFAULT NULL COMMENT '结束服务时间',
  `settled_at` DATETIME NULL DEFAULT NULL COMMENT '结算时间',
  `refund_amount` BIGINT NOT NULL DEFAULT 0 COMMENT '退款金额(分)',
  `is_minor_order` TINYINT NOT NULL DEFAULT 0 COMMENT '是否未成年人订单 0否 1是',
  `face_session_id` VARCHAR(64) NOT NULL DEFAULT '' COMMENT '活体会话ID',
  `appointment_time` DATETIME NULL DEFAULT NULL COMMENT '预约时间',
  `team_count` INT NOT NULL DEFAULT 1 COMMENT '车队人数',
  `created_at` DATETIME NULL DEFAULT NULL COMMENT '创建时间',
  `updated_at` DATETIME NULL DEFAULT NULL COMMENT '更新时间',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_order_no` (`order_no`),
  KEY `idx_status` (`status`),
  KEY `idx_user_id` (`user_id`),
  KEY `idx_player_id` (`player_id`),
  KEY `idx_club_id` (`club_id`),
  KEY `idx_created_at` (`created_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='订单表';

-- 订单状态流转日志表
DROP TABLE IF EXISTS `order_status_logs`;
CREATE TABLE `order_status_logs` (
  `id` BIGINT NOT NULL AUTO_INCREMENT COMMENT '主键ID',
  `order_id` BIGINT NOT NULL DEFAULT 0 COMMENT '订单ID',
  `from_status` TINYINT NOT NULL DEFAULT 0 COMMENT '原状态',
  `to_status` TINYINT NOT NULL DEFAULT 0 COMMENT '新状态',
  `operator_id` BIGINT NOT NULL DEFAULT 0 COMMENT '操作人ID',
  `operator_type` VARCHAR(32) NOT NULL DEFAULT '' COMMENT '操作人类型 user/admin/system',
  `reason` VARCHAR(255) NOT NULL DEFAULT '' COMMENT '变更原因',
  `created_at` DATETIME NULL DEFAULT NULL COMMENT '创建时间',
  PRIMARY KEY (`id`),
  KEY `idx_order_id` (`order_id`),
  KEY `idx_created_at` (`created_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='订单状态流转日志表';

-- 订单履约凭证表
DROP TABLE IF EXISTS `order_evidence`;
CREATE TABLE `order_evidence` (
  `id` BIGINT NOT NULL AUTO_INCREMENT COMMENT '主键ID',
  `order_id` BIGINT NOT NULL DEFAULT 0 COMMENT '订单ID',
  `user_id` BIGINT NOT NULL DEFAULT 0 COMMENT '上传用户ID',
  `type` VARCHAR(32) NOT NULL DEFAULT '' COMMENT '类型 video/screenshot',
  `file_url` VARCHAR(512) NOT NULL DEFAULT '' COMMENT '文件URL',
  `created_at` DATETIME NULL DEFAULT NULL COMMENT '创建时间',
  PRIMARY KEY (`id`),
  KEY `idx_order_id` (`order_id`),
  KEY `idx_user_id` (`user_id`),
  KEY `idx_created_at` (`created_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='订单履约凭证表';

-- 预约单表
DROP TABLE IF EXISTS `order_appointments`;
CREATE TABLE `order_appointments` (
  `id` BIGINT NOT NULL AUTO_INCREMENT COMMENT '主键ID',
  `order_id` BIGINT NOT NULL DEFAULT 0 COMMENT '订单ID',
  `player_id` BIGINT NOT NULL DEFAULT 0 COMMENT '打手ID',
  `appointment_time` DATETIME NULL DEFAULT NULL COMMENT '预约时间',
  `reminder_sent` TINYINT NOT NULL DEFAULT 0 COMMENT '是否已发送提醒 0否 1是',
  `created_at` DATETIME NULL DEFAULT NULL COMMENT '创建时间',
  `updated_at` DATETIME NULL DEFAULT NULL COMMENT '更新时间',
  PRIMARY KEY (`id`),
  KEY `idx_order_id` (`order_id`),
  KEY `idx_player_id` (`player_id`),
  KEY `idx_appointment_time` (`appointment_time`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='预约单表';

-- 订单竞价记录表
DROP TABLE IF EXISTS `order_bids`;
CREATE TABLE `order_bids` (
  `id` BIGINT NOT NULL AUTO_INCREMENT COMMENT '主键ID',
  `order_id` BIGINT NOT NULL DEFAULT 0 COMMENT '订单ID',
  `player_id` BIGINT NOT NULL DEFAULT 0 COMMENT '打手ID',
  `bid_amount` BIGINT NOT NULL DEFAULT 0 COMMENT '竞价金额(分)',
  `created_at` DATETIME NULL DEFAULT NULL COMMENT '创建时间',
  PRIMARY KEY (`id`),
  KEY `idx_order_id` (`order_id`),
  KEY `idx_player_id` (`player_id`),
  KEY `idx_created_at` (`created_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='订单竞价记录表';

-- 订单套餐表
DROP TABLE IF EXISTS `order_packages`;
CREATE TABLE `order_packages` (
  `id` BIGINT NOT NULL AUTO_INCREMENT COMMENT '主键ID',
  `name` VARCHAR(64) NOT NULL DEFAULT '' COMMENT '套餐名称',
  `type` VARCHAR(32) NOT NULL DEFAULT '' COMMENT '类型 hour/multi_game/teaching',
  `price` BIGINT NOT NULL DEFAULT 0 COMMENT '价格(分)',
  `content` JSON COMMENT '套餐内容',
  `status` TINYINT NOT NULL DEFAULT 1 COMMENT '状态 1上架 0下架',
  `created_at` DATETIME NULL DEFAULT NULL COMMENT '创建时间',
  `updated_at` DATETIME NULL DEFAULT NULL COMMENT '更新时间',
  PRIMARY KEY (`id`),
  KEY `idx_type` (`type`),
  KEY `idx_status` (`status`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='订单套餐表';

-- 订单服务计时表
DROP TABLE IF EXISTS `order_service_timers`;
CREATE TABLE `order_service_timers` (
  `id` BIGINT NOT NULL AUTO_INCREMENT COMMENT '主键ID',
  `order_id` BIGINT NOT NULL DEFAULT 0 COMMENT '订单ID',
  `start_time` DATETIME NULL DEFAULT NULL COMMENT '开始时间',
  `end_time` DATETIME NULL DEFAULT NULL COMMENT '结束时间',
  `duration_seconds` INT NOT NULL DEFAULT 0 COMMENT '服务时长(秒)',
  `created_at` DATETIME NULL DEFAULT NULL COMMENT '创建时间',
  `updated_at` DATETIME NULL DEFAULT NULL COMMENT '更新时间',
  PRIMARY KEY (`id`),
  KEY `idx_order_id` (`order_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='订单服务计时表';

-- 订单退款规则表
DROP TABLE IF EXISTS `order_refund_rules`;
CREATE TABLE `order_refund_rules` (
  `id` BIGINT NOT NULL AUTO_INCREMENT COMMENT '主键ID',
  `within_minutes` INT NOT NULL DEFAULT 0 COMMENT '下单后X分钟内',
  `refund_type` VARCHAR(32) NOT NULL DEFAULT '' COMMENT '退款类型 full/proportional',
  `proportion` DECIMAL(3,2) NOT NULL DEFAULT 0.00 COMMENT '退款比例',
  `created_at` DATETIME NULL DEFAULT NULL COMMENT '创建时间',
  `updated_at` DATETIME NULL DEFAULT NULL COMMENT '更新时间',
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='订单退款规则表';

-- 订单类型配置表
DROP TABLE IF EXISTS `order_type_configs`;
CREATE TABLE `order_type_configs` (
  `id` BIGINT NOT NULL AUTO_INCREMENT COMMENT '主键ID',
  `type` TINYINT NOT NULL DEFAULT 0 COMMENT '订单类型 1即时 2预约 3车队 4教学',
  `name` VARCHAR(64) NOT NULL DEFAULT '' COMMENT '类型名称',
  `enabled` TINYINT NOT NULL DEFAULT 1 COMMENT '是否启用 0否 1是',
  `sort_order` INT NOT NULL DEFAULT 0 COMMENT '排序',
  `created_at` DATETIME NULL DEFAULT NULL COMMENT '创建时间',
  `updated_at` DATETIME NULL DEFAULT NULL COMMENT '更新时间',
  PRIMARY KEY (`id`),
  KEY `idx_type` (`type`),
  KEY `idx_enabled` (`enabled`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='订单类型配置表';

-- ============================================================
-- 模块6: 支付与资金
-- ============================================================

-- 支付记录表
DROP TABLE IF EXISTS `payments`;
CREATE TABLE `payments` (
  `id` BIGINT NOT NULL AUTO_INCREMENT COMMENT '主键ID',
  `order_id` BIGINT NOT NULL DEFAULT 0 COMMENT '订单ID',
  `out_trade_no` VARCHAR(64) NOT NULL DEFAULT '' COMMENT '商户订单号',
  `transaction_id` VARCHAR(64) NOT NULL DEFAULT '' COMMENT '第三方交易号',
  `amount` BIGINT NOT NULL DEFAULT 0 COMMENT '支付金额(分)',
  `pay_method` VARCHAR(32) NOT NULL DEFAULT '' COMMENT '支付方式 wechat/ios',
  `pay_time` DATETIME NULL DEFAULT NULL COMMENT '支付时间',
  `refund_amount` BIGINT NOT NULL DEFAULT 0 COMMENT '退款金额(分)',
  `refund_time` DATETIME NULL DEFAULT NULL COMMENT '退款时间',
  `status` VARCHAR(32) NOT NULL DEFAULT 'pending' COMMENT '状态 pending/paid/refunded/partial_refund',
  `created_at` DATETIME NULL DEFAULT NULL COMMENT '创建时间',
  `updated_at` DATETIME NULL DEFAULT NULL COMMENT '更新时间',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_out_trade_no` (`out_trade_no`),
  KEY `idx_order_id` (`order_id`),
  KEY `idx_transaction_id` (`transaction_id`),
  KEY `idx_status` (`status`),
  KEY `idx_created_at` (`created_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='支付记录表';

-- 提现记录表
DROP TABLE IF EXISTS `withdrawals`;
CREATE TABLE `withdrawals` (
  `id` BIGINT NOT NULL AUTO_INCREMENT COMMENT '主键ID',
  `user_id` BIGINT NOT NULL DEFAULT 0 COMMENT '用户ID',
  `amount` BIGINT NOT NULL DEFAULT 0 COMMENT '提现金额(分)',
  `fee` BIGINT NOT NULL DEFAULT 0 COMMENT '手续费(分)',
  `tax` BIGINT NOT NULL DEFAULT 0 COMMENT '个税(分)',
  `net_amount` BIGINT NOT NULL DEFAULT 0 COMMENT '到账金额(分)',
  `bank_card` VARCHAR(32) NOT NULL DEFAULT '' COMMENT '银行卡号',
  `bank_name` VARCHAR(64) NOT NULL DEFAULT '' COMMENT '开户行',
  `bank_phone` VARCHAR(20) NOT NULL DEFAULT '' COMMENT '银行预留手机号',
  `id_card` VARCHAR(18) NOT NULL DEFAULT '' COMMENT '身份证号',
  `real_name` VARCHAR(32) NOT NULL DEFAULT '' COMMENT '真实姓名',
  `channel` VARCHAR(32) NOT NULL DEFAULT '' COMMENT '提现渠道 wechat/bank',
  `status` VARCHAR(32) NOT NULL DEFAULT 'pending' COMMENT '状态 pending/approved/rejected/paid',
  `reviewer_id` BIGINT NOT NULL DEFAULT 0 COMMENT '审核人ID',
  `reviewed_at` DATETIME NULL DEFAULT NULL COMMENT '审核时间',
  `paid_at` DATETIME NULL DEFAULT NULL COMMENT '打款时间',
  `created_at` DATETIME NULL DEFAULT NULL COMMENT '创建时间',
  `updated_at` DATETIME NULL DEFAULT NULL COMMENT '更新时间',
  PRIMARY KEY (`id`),
  KEY `idx_user_id` (`user_id`),
  KEY `idx_status` (`status`),
  KEY `idx_created_at` (`created_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='提现记录表';

-- 批量提现表
DROP TABLE IF EXISTS `withdraw_batches`;
CREATE TABLE `withdraw_batches` (
  `id` BIGINT NOT NULL AUTO_INCREMENT COMMENT '主键ID',
  `batch_no` VARCHAR(64) NOT NULL DEFAULT '' COMMENT '批次号',
  `count` INT NOT NULL DEFAULT 0 COMMENT '总笔数',
  `total_amount` BIGINT NOT NULL DEFAULT 0 COMMENT '总金额(分)',
  `status` TINYINT NOT NULL DEFAULT 0 COMMENT '状态 0待执行 1执行中 2完成 3部分失败',
  `operator_id` BIGINT NOT NULL DEFAULT 0 COMMENT '操作人ID',
  `created_at` DATETIME NULL DEFAULT NULL COMMENT '创建时间',
  `updated_at` DATETIME NULL DEFAULT NULL COMMENT '更新时间',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_batch_no` (`batch_no`),
  KEY `idx_status` (`status`),
  KEY `idx_created_at` (`created_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='批量提现表';

-- 提现配置表
DROP TABLE IF EXISTS `withdraw_configs`;
CREATE TABLE `withdraw_configs` (
  `id` BIGINT NOT NULL AUTO_INCREMENT COMMENT '主键ID',
  `key` VARCHAR(128) NOT NULL DEFAULT '' COMMENT '配置键',
  `value` TEXT COMMENT '配置值',
  `description` VARCHAR(255) NOT NULL DEFAULT '' COMMENT '配置描述',
  `created_at` DATETIME NULL DEFAULT NULL COMMENT '创建时间',
  `updated_at` DATETIME NULL DEFAULT NULL COMMENT '更新时间',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_key` (`key`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='提现配置表';

-- 分账规则表
DROP TABLE IF EXISTS `profit_share_rules`;
CREATE TABLE `profit_share_rules` (
  `id` BIGINT NOT NULL AUTO_INCREMENT COMMENT '主键ID',
  `name` VARCHAR(64) NOT NULL DEFAULT '' COMMENT '规则名称',
  `player_ratio` DECIMAL(5,2) NOT NULL DEFAULT 0.00 COMMENT '打手分账比例%',
  `club_ratio` DECIMAL(5,2) NOT NULL DEFAULT 0.00 COMMENT '俱乐部分账比例%',
  `distributor_ratio` DECIMAL(5,2) NOT NULL DEFAULT 0.00 COMMENT '分销商分账比例%',
  `platform_ratio` DECIMAL(5,2) NOT NULL DEFAULT 0.00 COMMENT '平台分账比例%',
  `conditions` JSON COMMENT '适用条件',
  `status` TINYINT NOT NULL DEFAULT 1 COMMENT '状态 1启用 0停用',
  `created_at` DATETIME NULL DEFAULT NULL COMMENT '创建时间',
  `updated_at` DATETIME NULL DEFAULT NULL COMMENT '更新时间',
  PRIMARY KEY (`id`),
  KEY `idx_status` (`status`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='分账规则表';

-- 分账记录表
DROP TABLE IF EXISTS `profit_share_records`;
CREATE TABLE `profit_share_records` (
  `id` BIGINT NOT NULL AUTO_INCREMENT COMMENT '主键ID',
  `order_id` BIGINT NOT NULL DEFAULT 0 COMMENT '订单ID',
  `user_id` BIGINT NOT NULL DEFAULT 0 COMMENT '收款方用户ID',
  `role` VARCHAR(32) NOT NULL DEFAULT '' COMMENT '角色 player/club/distributor/platform',
  `amount` BIGINT NOT NULL DEFAULT 0 COMMENT '分账金额(分)',
  `ratio` DECIMAL(5,2) NOT NULL DEFAULT 0.00 COMMENT '分账比例%',
  `status` TINYINT NOT NULL DEFAULT 0 COMMENT '状态 0待分账 1已分账 2已回滚',
  `created_at` DATETIME NULL DEFAULT NULL COMMENT '创建时间',
  `updated_at` DATETIME NULL DEFAULT NULL COMMENT '更新时间',
  PRIMARY KEY (`id`),
  KEY `idx_order_id` (`order_id`),
  KEY `idx_user_id` (`user_id`),
  KEY `idx_status` (`status`),
  KEY `idx_created_at` (`created_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='分账记录表';

-- 退款分账回滚表
DROP TABLE IF EXISTS `profit_share_refunds`;
CREATE TABLE `profit_share_refunds` (
  `id` BIGINT NOT NULL AUTO_INCREMENT COMMENT '主键ID',
  `order_id` BIGINT NOT NULL DEFAULT 0 COMMENT '订单ID',
  `original_share_id` BIGINT NOT NULL DEFAULT 0 COMMENT '原分账记录ID',
  `rollback_amount` BIGINT NOT NULL DEFAULT 0 COMMENT '回滚金额(分)',
  `created_at` DATETIME NULL DEFAULT NULL COMMENT '创建时间',
  PRIMARY KEY (`id`),
  KEY `idx_order_id` (`order_id`),
  KEY `idx_original_share_id` (`original_share_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='退款分账回滚表';

-- 子商户账户表
DROP TABLE IF EXISTS `merchant_accounts`;
CREATE TABLE `merchant_accounts` (
  `id` BIGINT NOT NULL AUTO_INCREMENT COMMENT '主键ID',
  `user_id` BIGINT NOT NULL DEFAULT 0 COMMENT '用户ID',
  `club_id` BIGINT NOT NULL DEFAULT 0 COMMENT '俱乐部ID',
  `merchant_id` VARCHAR(64) NOT NULL DEFAULT '' COMMENT '商户号',
  `provider` VARCHAR(64) NOT NULL DEFAULT '' COMMENT '服务商 MallBook等',
  `status` TINYINT NOT NULL DEFAULT 0 COMMENT '状态 0待开通 1已开通 2已冻结',
  `created_at` DATETIME NULL DEFAULT NULL COMMENT '创建时间',
  `updated_at` DATETIME NULL DEFAULT NULL COMMENT '更新时间',
  PRIMARY KEY (`id`),
  KEY `idx_user_id` (`user_id`),
  KEY `idx_club_id` (`club_id`),
  KEY `idx_status` (`status`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='子商户账户表';

-- 税率配置表
DROP TABLE IF EXISTS `tax_configs`;
CREATE TABLE `tax_configs` (
  `id` BIGINT NOT NULL AUTO_INCREMENT COMMENT '主键ID',
  `income_threshold` BIGINT NOT NULL DEFAULT 0 COMMENT '收入门槛(分)',
  `tax_rate` DECIMAL(5,2) NOT NULL DEFAULT 0.00 COMMENT '税率%',
  `deduction` BIGINT NOT NULL DEFAULT 0 COMMENT '速算扣除数(分)',
  `created_at` DATETIME NULL DEFAULT NULL COMMENT '创建时间',
  `updated_at` DATETIME NULL DEFAULT NULL COMMENT '更新时间',
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='税率配置表';

-- 完税凭证表
DROP TABLE IF EXISTS `tax_records`;
CREATE TABLE `tax_records` (
  `id` BIGINT NOT NULL AUTO_INCREMENT COMMENT '主键ID',
  `user_id` BIGINT NOT NULL DEFAULT 0 COMMENT '用户ID',
  `withdrawal_id` BIGINT NOT NULL DEFAULT 0 COMMENT '提现ID',
  `taxable_amount` BIGINT NOT NULL DEFAULT 0 COMMENT '应税金额(分)',
  `tax_amount` BIGINT NOT NULL DEFAULT 0 COMMENT '税额(分)',
  `certificate_no` VARCHAR(128) NOT NULL DEFAULT '' COMMENT '完税凭证号',
  `created_at` DATETIME NULL DEFAULT NULL COMMENT '创建时间',
  PRIMARY KEY (`id`),
  KEY `idx_user_id` (`user_id`),
  KEY `idx_withdrawal_id` (`withdrawal_id`),
  KEY `idx_created_at` (`created_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='完税凭证表';

-- ============================================================
-- 模块7: IM即时通讯
-- ============================================================

-- 聊天会话表
DROP TABLE IF EXISTS `chat_sessions`;
CREATE TABLE `chat_sessions` (
  `id` BIGINT NOT NULL AUTO_INCREMENT COMMENT '主键ID',
  `session_type` VARCHAR(32) NOT NULL DEFAULT '' COMMENT '会话类型 order/after_sale/group_internal/group_category',
  `ref_id` BIGINT NOT NULL DEFAULT 0 COMMENT '关联ID(订单ID或群ID)',
  `status` TINYINT NOT NULL DEFAULT 1 COMMENT '状态 1正常 0关闭',
  `created_at` DATETIME NULL DEFAULT NULL COMMENT '创建时间',
  `updated_at` DATETIME NULL DEFAULT NULL COMMENT '更新时间',
  PRIMARY KEY (`id`),
  KEY `idx_session_type` (`session_type`),
  KEY `idx_ref_id` (`ref_id`),
  KEY `idx_created_at` (`created_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='聊天会话表';

-- 聊天消息表
DROP TABLE IF EXISTS `chat_messages`;
CREATE TABLE `chat_messages` (
  `id` BIGINT NOT NULL AUTO_INCREMENT COMMENT '主键ID',
  `session_id` BIGINT NOT NULL DEFAULT 0 COMMENT '会话ID',
  `sender_id` BIGINT NOT NULL DEFAULT 0 COMMENT '发送者ID',
  `sender_type` VARCHAR(32) NOT NULL DEFAULT '' COMMENT '发送者类型 user/player/club_admin/platform',
  `msg_type` VARCHAR(32) NOT NULL DEFAULT 'text' COMMENT '消息类型 text/image/voice/file/card',
  `content` TEXT COMMENT '消息内容',
  `media_url` VARCHAR(512) NOT NULL DEFAULT '' COMMENT '媒体URL',
  `duration` INT NOT NULL DEFAULT 0 COMMENT '语音时长(秒)',
  `asr_text` TEXT COMMENT 'ASR转文字结果',
  `is_read` TINYINT NOT NULL DEFAULT 0 COMMENT '是否已读 0否 1是',
  `is_revoked` TINYINT NOT NULL DEFAULT 0 COMMENT '是否撤回 0否 1是',
  `risk_level` TINYINT NOT NULL DEFAULT 0 COMMENT '风险等级 0无 1低 2中 3高',
  `created_at` DATETIME NULL DEFAULT NULL COMMENT '创建时间',
  PRIMARY KEY (`id`),
  KEY `idx_session_id` (`session_id`),
  KEY `idx_created_at` (`created_at`),
  KEY `idx_sender_id` (`sender_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='聊天消息表';

-- 消息撤回记录表
DROP TABLE IF EXISTS `chat_message_revokes`;
CREATE TABLE `chat_message_revokes` (
  `id` BIGINT NOT NULL AUTO_INCREMENT COMMENT '主键ID',
  `message_id` BIGINT NOT NULL DEFAULT 0 COMMENT '消息ID',
  `revoked_by` BIGINT NOT NULL DEFAULT 0 COMMENT '撤回用户ID',
  `revoked_at` DATETIME NULL DEFAULT NULL COMMENT '撤回时间',
  `created_at` DATETIME NULL DEFAULT NULL COMMENT '创建时间',
  PRIMARY KEY (`id`),
  KEY `idx_message_id` (`message_id`),
  KEY `idx_revoked_by` (`revoked_by`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='消息撤回记录表';

-- 聊天文件传输表
DROP TABLE IF EXISTS `chat_files`;
CREATE TABLE `chat_files` (
  `id` BIGINT NOT NULL AUTO_INCREMENT COMMENT '主键ID',
  `session_id` BIGINT NOT NULL DEFAULT 0 COMMENT '会话ID',
  `uploader_id` BIGINT NOT NULL DEFAULT 0 COMMENT '上传者ID',
  `file_url` VARCHAR(512) NOT NULL DEFAULT '' COMMENT '文件URL',
  `file_name` VARCHAR(255) NOT NULL DEFAULT '' COMMENT '文件名',
  `file_size` BIGINT NOT NULL DEFAULT 0 COMMENT '文件大小(字节)',
  `file_type` VARCHAR(32) NOT NULL DEFAULT '' COMMENT '文件类型',
  `created_at` DATETIME NULL DEFAULT NULL COMMENT '创建时间',
  PRIMARY KEY (`id`),
  KEY `idx_session_id` (`session_id`),
  KEY `idx_uploader_id` (`uploader_id`),
  KEY `idx_created_at` (`created_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='聊天文件传输表';

-- ASR缓存表
DROP TABLE IF EXISTS `chat_asr_caches`;
CREATE TABLE `chat_asr_caches` (
  `id` BIGINT NOT NULL AUTO_INCREMENT COMMENT '主键ID',
  `message_id` BIGINT NOT NULL DEFAULT 0 COMMENT '消息ID',
  `asr_text` TEXT COMMENT '转文字结果',
  `provider` VARCHAR(32) NOT NULL DEFAULT '' COMMENT 'ASR服务商',
  `created_at` DATETIME NULL DEFAULT NULL COMMENT '创建时间',
  PRIMARY KEY (`id`),
  KEY `idx_message_id` (`message_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='ASR缓存表';

-- 聊天快捷卡片表
DROP TABLE IF EXISTS `chat_quick_cards`;
CREATE TABLE `chat_quick_cards` (
  `id` BIGINT NOT NULL AUTO_INCREMENT COMMENT '主键ID',
  `card_type` VARCHAR(32) NOT NULL DEFAULT '' COMMENT '卡片类型 quote/package/appointment',
  `content` JSON COMMENT '卡片内容',
  `creator_id` BIGINT NOT NULL DEFAULT 0 COMMENT '创建人ID',
  `created_at` DATETIME NULL DEFAULT NULL COMMENT '创建时间',
  `updated_at` DATETIME NULL DEFAULT NULL COMMENT '更新时间',
  PRIMARY KEY (`id`),
  KEY `idx_card_type` (`card_type`),
  KEY `idx_creator_id` (`creator_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='聊天快捷卡片表';

-- 飞单风控规则表
DROP TABLE IF EXISTS `chat_anti_fraud_rules`;
CREATE TABLE `chat_anti_fraud_rules` (
  `id` BIGINT NOT NULL AUTO_INCREMENT COMMENT '主键ID',
  `pattern` VARCHAR(255) NOT NULL DEFAULT '' COMMENT '匹配正则/关键词',
  `description` VARCHAR(255) NOT NULL DEFAULT '' COMMENT '规则描述',
  `action` VARCHAR(32) NOT NULL DEFAULT 'block' COMMENT '执行动作 block/warn/log',
  `enabled` TINYINT NOT NULL DEFAULT 1 COMMENT '是否启用 0否 1是',
  `created_at` DATETIME NULL DEFAULT NULL COMMENT '创建时间',
  `updated_at` DATETIME NULL DEFAULT NULL COMMENT '更新时间',
  PRIMARY KEY (`id`),
  KEY `idx_enabled` (`enabled`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='飞单风控规则表';

-- 飞单风控日志表
DROP TABLE IF EXISTS `chat_anti_fraud_logs`;
CREATE TABLE `chat_anti_fraud_logs` (
  `id` BIGINT NOT NULL AUTO_INCREMENT COMMENT '主键ID',
  `session_id` BIGINT NOT NULL DEFAULT 0 COMMENT '会话ID',
  `user_id` BIGINT NOT NULL DEFAULT 0 COMMENT '用户ID',
  `matched_pattern` VARCHAR(255) NOT NULL DEFAULT '' COMMENT '命中规则',
  `action` VARCHAR(32) NOT NULL DEFAULT '' COMMENT '执行动作',
  `message_content` TEXT COMMENT '消息内容',
  `created_at` DATETIME NULL DEFAULT NULL COMMENT '创建时间',
  PRIMARY KEY (`id`),
  KEY `idx_session_id` (`session_id`),
  KEY `idx_user_id` (`user_id`),
  KEY `idx_created_at` (`created_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='飞单风控日志表';

-- 聊天审计日志表
DROP TABLE IF EXISTS `chat_audit_logs`;
CREATE TABLE `chat_audit_logs` (
  `id` BIGINT NOT NULL AUTO_INCREMENT COMMENT '主键ID',
  `session_id` BIGINT NOT NULL DEFAULT 0 COMMENT '会话ID',
  `message_id` BIGINT NOT NULL DEFAULT 0 COMMENT '消息ID',
  `action` VARCHAR(64) NOT NULL DEFAULT '' COMMENT '审计动作',
  `operator_id` BIGINT NOT NULL DEFAULT 0 COMMENT '操作人ID',
  `created_at` DATETIME NULL DEFAULT NULL COMMENT '创建时间',
  PRIMARY KEY (`id`),
  KEY `idx_session_id` (`session_id`),
  KEY `idx_message_id` (`message_id`),
  KEY `idx_created_at` (`created_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='聊天审计日志表';

-- 举证上传日志表
DROP TABLE IF EXISTS `chat_upload_evidence_logs`;
CREATE TABLE `chat_upload_evidence_logs` (
  `id` BIGINT NOT NULL AUTO_INCREMENT COMMENT '主键ID',
  `session_id` BIGINT NOT NULL DEFAULT 0 COMMENT '会话ID',
  `order_id` BIGINT NOT NULL DEFAULT 0 COMMENT '订单ID',
  `uploader_id` BIGINT NOT NULL DEFAULT 0 COMMENT '上传者ID',
  `file_url` VARCHAR(512) NOT NULL DEFAULT '' COMMENT '文件URL',
  `file_type` VARCHAR(32) NOT NULL DEFAULT '' COMMENT '文件类型',
  `created_at` DATETIME NULL DEFAULT NULL COMMENT '创建时间',
  PRIMARY KEY (`id`),
  KEY `idx_session_id` (`session_id`),
  KEY `idx_order_id` (`order_id`),
  KEY `idx_created_at` (`created_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='举证上传日志表';

-- ============================================================
-- 模块8: 群聊体系
-- ============================================================

-- 群聊表
DROP TABLE IF EXISTS `group_chats`;
CREATE TABLE `group_chats` (
  `id` BIGINT NOT NULL AUTO_INCREMENT COMMENT '主键ID',
  `group_name` VARCHAR(128) NOT NULL DEFAULT '' COMMENT '群名称',
  `group_type` VARCHAR(32) NOT NULL DEFAULT '' COMMENT '群类型 internal/category',
  `club_id` BIGINT NOT NULL DEFAULT 0 COMMENT '俱乐部ID',
  `category_type` VARCHAR(32) NOT NULL DEFAULT '' COMMENT '分类类型 chat/welfare/aftersale',
  `creator_id` BIGINT NOT NULL DEFAULT 0 COMMENT '创建人ID',
  `announcement` TEXT COMMENT '群公告',
  `announcement_at` DATETIME NULL DEFAULT NULL COMMENT '公告更新时间',
  `status` TINYINT NOT NULL DEFAULT 1 COMMENT '状态 1正常 0解散',
  `created_at` DATETIME NULL DEFAULT NULL COMMENT '创建时间',
  `updated_at` DATETIME NULL DEFAULT NULL COMMENT '更新时间',
  PRIMARY KEY (`id`),
  KEY `idx_club_id` (`club_id`),
  KEY `idx_group_type` (`group_type`),
  KEY `idx_status` (`status`),
  KEY `idx_created_at` (`created_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='群聊表';

-- 群成员表
DROP TABLE IF EXISTS `group_chat_members`;
CREATE TABLE `group_chat_members` (
  `id` BIGINT NOT NULL AUTO_INCREMENT COMMENT '主键ID',
  `group_id` BIGINT NOT NULL DEFAULT 0 COMMENT '群ID',
  `user_id` BIGINT NOT NULL DEFAULT 0 COMMENT '用户ID',
  `role` VARCHAR(32) NOT NULL DEFAULT 'member' COMMENT '角色 member/admin/owner/platform',
  `is_muted` TINYINT NOT NULL DEFAULT 0 COMMENT '是否禁言 0否 1是',
  `joined_at` DATETIME NULL DEFAULT NULL COMMENT '加入时间',
  `created_at` DATETIME NULL DEFAULT NULL COMMENT '创建时间',
  `updated_at` DATETIME NULL DEFAULT NULL COMMENT '更新时间',
  PRIMARY KEY (`id`),
  KEY `idx_group_id` (`group_id`),
  KEY `idx_user_id` (`user_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='群成员表';

-- 群消息表
DROP TABLE IF EXISTS `group_chat_messages`;
CREATE TABLE `group_chat_messages` (
  `id` BIGINT NOT NULL AUTO_INCREMENT COMMENT '主键ID',
  `group_id` BIGINT NOT NULL DEFAULT 0 COMMENT '群ID',
  `sender_id` BIGINT NOT NULL DEFAULT 0 COMMENT '发送者ID',
  `msg_type` VARCHAR(32) NOT NULL DEFAULT 'text' COMMENT '消息类型',
  `content` TEXT COMMENT '消息内容',
  `media_url` VARCHAR(512) NOT NULL DEFAULT '' COMMENT '媒体URL',
  `duration` INT NOT NULL DEFAULT 0 COMMENT '时长(秒)',
  `asr_text` TEXT COMMENT '语音转文字',
  `is_revoked` TINYINT NOT NULL DEFAULT 0 COMMENT '是否撤回 0否 1是',
  `created_at` DATETIME NULL DEFAULT NULL COMMENT '创建时间',
  PRIMARY KEY (`id`),
  KEY `idx_group_id` (`group_id`),
  KEY `idx_created_at` (`created_at`),
  KEY `idx_sender_id` (`sender_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='群消息表';

-- 群黑名单表
DROP TABLE IF EXISTS `group_chat_blacklists`;
CREATE TABLE `group_chat_blacklists` (
  `id` BIGINT NOT NULL AUTO_INCREMENT COMMENT '主键ID',
  `group_id` BIGINT NOT NULL DEFAULT 0 COMMENT '群ID',
  `user_id` BIGINT NOT NULL DEFAULT 0 COMMENT '用户ID',
  `reason` VARCHAR(255) NOT NULL DEFAULT '' COMMENT '拉黑原因',
  `created_at` DATETIME NULL DEFAULT NULL COMMENT '创建时间',
  PRIMARY KEY (`id`),
  KEY `idx_group_id` (`group_id`),
  KEY `idx_user_id` (`user_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='群黑名单表';

-- 定时公告表
DROP TABLE IF EXISTS `group_announcement_schedules`;
CREATE TABLE `group_announcement_schedules` (
  `id` BIGINT NOT NULL AUTO_INCREMENT COMMENT '主键ID',
  `group_id` BIGINT NOT NULL DEFAULT 0 COMMENT '群ID',
  `content` TEXT COMMENT '公告内容',
  `scheduled_at` DATETIME NULL DEFAULT NULL COMMENT '计划发送时间',
  `status` TINYINT NOT NULL DEFAULT 0 COMMENT '状态 0待发送 1已发送 2已取消',
  `created_at` DATETIME NULL DEFAULT NULL COMMENT '创建时间',
  `updated_at` DATETIME NULL DEFAULT NULL COMMENT '更新时间',
  PRIMARY KEY (`id`),
  KEY `idx_group_id` (`group_id`),
  KEY `idx_scheduled_at` (`scheduled_at`),
  KEY `idx_status` (`status`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='定时公告表';

-- ============================================================
-- 模块9: 售后与申诉
-- ============================================================

-- 售后会话表
DROP TABLE IF EXISTS `after_sale_sessions`;
CREATE TABLE `after_sale_sessions` (
  `id` BIGINT NOT NULL AUTO_INCREMENT COMMENT '主键ID',
  `order_id` BIGINT NOT NULL DEFAULT 0 COMMENT '订单ID',
  `user_id` BIGINT NOT NULL DEFAULT 0 COMMENT '用户ID',
  `club_id` BIGINT NOT NULL DEFAULT 0 COMMENT '俱乐部ID',
  `status` TINYINT NOT NULL DEFAULT 1 COMMENT '状态 1处理中 2已关闭',
  `intervention_status` TINYINT NOT NULL DEFAULT 0 COMMENT '介入状态 0=未介入 1=待介入 2=已介入',
  `intervention_type` VARCHAR(32) NOT NULL DEFAULT '' COMMENT '介入类型 keyword/manual',
  `created_at` DATETIME NULL DEFAULT NULL COMMENT '创建时间',
  `updated_at` DATETIME NULL DEFAULT NULL COMMENT '更新时间',
  PRIMARY KEY (`id`),
  KEY `idx_order_id` (`order_id`),
  KEY `idx_user_id` (`user_id`),
  KEY `idx_status` (`status`),
  KEY `idx_intervention_status` (`intervention_status`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='售后会话表';

-- 售后消息表
DROP TABLE IF EXISTS `after_sale_messages`;
CREATE TABLE `after_sale_messages` (
  `id` BIGINT NOT NULL AUTO_INCREMENT COMMENT '主键ID',
  `session_id` BIGINT NOT NULL DEFAULT 0 COMMENT '售后会话ID',
  `sender_id` BIGINT NOT NULL DEFAULT 0 COMMENT '发送者ID',
  `content` TEXT COMMENT '消息内容',
  `msg_type` VARCHAR(32) NOT NULL DEFAULT 'text' COMMENT '消息类型',
  `media_url` VARCHAR(512) NOT NULL DEFAULT '' COMMENT '媒体URL',
  `created_at` DATETIME NULL DEFAULT NULL COMMENT '创建时间',
  PRIMARY KEY (`id`),
  KEY `idx_session_id` (`session_id`),
  KEY `idx_sender_id` (`sender_id`),
  KEY `idx_created_at` (`created_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='售后消息表';

-- 售后风控关键词表
DROP TABLE IF EXISTS `after_sale_keywords`;
CREATE TABLE `after_sale_keywords` (
  `id` BIGINT NOT NULL AUTO_INCREMENT COMMENT '主键ID',
  `keyword` VARCHAR(128) NOT NULL DEFAULT '' COMMENT '关键词',
  `match_type` VARCHAR(32) NOT NULL DEFAULT 'exact' COMMENT '匹配类型 exact/fuzzy',
  `enabled` TINYINT NOT NULL DEFAULT 1 COMMENT '是否启用 0否 1是',
  `created_at` DATETIME NULL DEFAULT NULL COMMENT '创建时间',
  `updated_at` DATETIME NULL DEFAULT NULL COMMENT '更新时间',
  PRIMARY KEY (`id`),
  KEY `idx_enabled` (`enabled`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='售后风控关键词表';

-- 售后介入日志表
DROP TABLE IF EXISTS `after_sale_intervene_logs`;
CREATE TABLE `after_sale_intervene_logs` (
  `id` BIGINT NOT NULL AUTO_INCREMENT COMMENT '主键ID',
  `session_id` BIGINT NOT NULL DEFAULT 0 COMMENT '售后会话ID',
  `trigger_type` VARCHAR(32) NOT NULL DEFAULT '' COMMENT '触发类型 keyword/manual',
  `keyword` VARCHAR(128) NOT NULL DEFAULT '' COMMENT '命中关键词',
  `operator_id` BIGINT NOT NULL DEFAULT 0 COMMENT '操作人ID',
  `created_at` DATETIME NULL DEFAULT NULL COMMENT '创建时间',
  PRIMARY KEY (`id`),
  KEY `idx_session_id` (`session_id`),
  KEY `idx_created_at` (`created_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='售后介入日志表';

-- 售后风控日志表
DROP TABLE IF EXISTS `after_sale_risk_logs`;
CREATE TABLE `after_sale_risk_logs` (
  `id` BIGINT NOT NULL AUTO_INCREMENT COMMENT '主键ID',
  `session_id` BIGINT NOT NULL DEFAULT 0 COMMENT '售后会话ID',
  `order_id` BIGINT NOT NULL DEFAULT 0 COMMENT '订单ID',
  `sender_id` BIGINT NOT NULL DEFAULT 0 COMMENT '发送者ID',
  `keyword` VARCHAR(128) NOT NULL DEFAULT '' COMMENT '命中关键词',
  `content_summary` VARCHAR(500) NOT NULL DEFAULT '' COMMENT '消息摘要',
  `created_at` DATETIME NULL DEFAULT NULL COMMENT '创建时间',
  PRIMARY KEY (`id`),
  KEY `idx_session_id` (`session_id`),
  KEY `idx_order_id` (`order_id`),
  KEY `idx_created_at` (`created_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='售后风控日志表';

-- 申诉工单表
DROP TABLE IF EXISTS `appeals`;
CREATE TABLE `appeals` (
  `id` BIGINT NOT NULL AUTO_INCREMENT COMMENT '主键ID',
  `order_id` BIGINT NOT NULL DEFAULT 0 COMMENT '订单ID',
  `user_id` BIGINT NOT NULL DEFAULT 0 COMMENT '用户ID',
  `type` VARCHAR(32) NOT NULL DEFAULT '' COMMENT '申诉类型',
  `description` TEXT COMMENT '问题描述',
  `status` VARCHAR(32) NOT NULL DEFAULT 'pending' COMMENT '状态 pending/processing/resolved/rejected',
  `evidence_urls` JSON COMMENT '证据URL数组',
  `created_at` DATETIME NULL DEFAULT NULL COMMENT '创建时间',
  `updated_at` DATETIME NULL DEFAULT NULL COMMENT '更新时间',
  PRIMARY KEY (`id`),
  KEY `idx_order_id` (`order_id`),
  KEY `idx_user_id` (`user_id`),
  KEY `idx_status` (`status`),
  KEY `idx_created_at` (`created_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='申诉工单表';

-- 申诉沟通表
DROP TABLE IF EXISTS `appeal_communications`;
CREATE TABLE `appeal_communications` (
  `id` BIGINT NOT NULL AUTO_INCREMENT COMMENT '主键ID',
  `appeal_id` BIGINT NOT NULL DEFAULT 0 COMMENT '申诉ID',
  `sender_id` BIGINT NOT NULL DEFAULT 0 COMMENT '发送者ID',
  `content` TEXT COMMENT '沟通内容',
  `created_at` DATETIME NULL DEFAULT NULL COMMENT '创建时间',
  PRIMARY KEY (`id`),
  KEY `idx_appeal_id` (`appeal_id`),
  KEY `idx_sender_id` (`sender_id`),
  KEY `idx_created_at` (`created_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='申诉沟通表';

-- 申诉催办表
DROP TABLE IF EXISTS `appeal_reminders`;
CREATE TABLE `appeal_reminders` (
  `id` BIGINT NOT NULL AUTO_INCREMENT COMMENT '主键ID',
  `appeal_id` BIGINT NOT NULL DEFAULT 0 COMMENT '申诉ID',
  `level` TINYINT NOT NULL DEFAULT 1 COMMENT '催办等级 1一级 2二级 3三级',
  `sent_at` DATETIME NULL DEFAULT NULL COMMENT '发送时间',
  `created_at` DATETIME NULL DEFAULT NULL COMMENT '创建时间',
  PRIMARY KEY (`id`),
  KEY `idx_appeal_id` (`appeal_id`),
  KEY `idx_created_at` (`created_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='申诉催办表';

-- ============================================================
-- 模块10: 平台官方账号
-- ============================================================

-- 平台官方账号表
DROP TABLE IF EXISTS `platform_official_accounts`;
CREATE TABLE `platform_official_accounts` (
  `id` BIGINT NOT NULL AUTO_INCREMENT COMMENT '主键ID',
  `username` VARCHAR(64) NOT NULL DEFAULT '' COMMENT '账号',
  `nickname` VARCHAR(64) NOT NULL DEFAULT '' COMMENT '昵称',
  `avatar` VARCHAR(512) NOT NULL DEFAULT '' COMMENT '头像URL',
  `status` TINYINT NOT NULL DEFAULT 1 COMMENT '状态 1正常 0停用',
  `created_at` DATETIME NULL DEFAULT NULL COMMENT '创建时间',
  `updated_at` DATETIME NULL DEFAULT NULL COMMENT '更新时间',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_username` (`username`),
  KEY `idx_status` (`status`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='平台官方账号表';

-- 平台介入日志表
DROP TABLE IF EXISTS `platform_intervention_logs`;
CREATE TABLE `platform_intervention_logs` (
  `id` BIGINT NOT NULL AUTO_INCREMENT COMMENT '主键ID',
  `session_id` BIGINT NOT NULL DEFAULT 0 COMMENT '会话ID',
  `order_id` BIGINT NOT NULL DEFAULT 0 COMMENT '订单ID',
  `trigger_type` VARCHAR(32) NOT NULL DEFAULT '' COMMENT '触发类型 keyword/manual',
  `handler_id` BIGINT NOT NULL DEFAULT 0 COMMENT '处理人ID',
  `result` VARCHAR(255) NOT NULL DEFAULT '' COMMENT '处理结果',
  `created_at` DATETIME NULL DEFAULT NULL COMMENT '创建时间',
  PRIMARY KEY (`id`),
  KEY `idx_session_id` (`session_id`),
  KEY `idx_order_id` (`order_id`),
  KEY `idx_created_at` (`created_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='平台介入日志表';

-- 平台处罚日志表
DROP TABLE IF EXISTS `platform_punishment_logs`;
CREATE TABLE `platform_punishment_logs` (
  `id` BIGINT NOT NULL AUTO_INCREMENT COMMENT '主键ID',
  `target_type` VARCHAR(32) NOT NULL DEFAULT '' COMMENT '目标类型 user/club',
  `target_id` BIGINT NOT NULL DEFAULT 0 COMMENT '目标ID',
  `punishment_type` VARCHAR(32) NOT NULL DEFAULT '' COMMENT '处罚类型 ban/freeze/fine',
  `reason` VARCHAR(255) NOT NULL DEFAULT '' COMMENT '处罚原因',
  `operator_id` BIGINT NOT NULL DEFAULT 0 COMMENT '操作人ID',
  `created_at` DATETIME NULL DEFAULT NULL COMMENT '创建时间',
  PRIMARY KEY (`id`),
  KEY `idx_target_type` (`target_type`),
  KEY `idx_target_id` (`target_id`),
  KEY `idx_created_at` (`created_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='平台处罚日志表';

-- ============================================================
-- 模块11: 邀请码体系
-- ============================================================

-- 邀请码表
DROP TABLE IF EXISTS `invite_codes`;
CREATE TABLE `invite_codes` (
  `id` BIGINT NOT NULL AUTO_INCREMENT COMMENT '主键ID',
  `code` VARCHAR(64) NOT NULL DEFAULT '' COMMENT '邀请码',
  `type` VARCHAR(32) NOT NULL DEFAULT 'platform' COMMENT '类型 club/platform',
  `club_id` BIGINT NOT NULL DEFAULT 0 COMMENT '指定俱乐部ID',
  `role` VARCHAR(32) NOT NULL DEFAULT '' COMMENT '角色 DS/FXS/null',
  `max_uses` INT NOT NULL DEFAULT 1 COMMENT '最大使用次数',
  `used_count` INT NOT NULL DEFAULT 0 COMMENT '已使用次数',
  `expire_at` DATETIME NULL DEFAULT NULL COMMENT '过期时间',
  `benefits` JSON COMMENT '福利配置',
  `status` VARCHAR(32) NOT NULL DEFAULT 'unused' COMMENT '状态 unused/used/exhausted/expired/revoked',
  `creator_id` BIGINT NOT NULL DEFAULT 0 COMMENT '创建人ID',
  `creator_type` VARCHAR(32) NOT NULL DEFAULT '' COMMENT '创建人类型 admin/club',
  `used_by` BIGINT NOT NULL DEFAULT 0 COMMENT '使用人ID',
  `used_at` DATETIME NULL DEFAULT NULL COMMENT '使用时间',
  `created_at` DATETIME NULL DEFAULT NULL COMMENT '创建时间',
  `updated_at` DATETIME NULL DEFAULT NULL COMMENT '更新时间',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_code` (`code`),
  KEY `idx_code` (`code`),
  KEY `idx_club_id` (`club_id`),
  KEY `idx_type` (`type`),
  KEY `idx_status` (`status`),
  KEY `idx_created_at` (`created_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='邀请码表';

-- ============================================================
-- 模块12: V标身份体系
-- ============================================================

-- 用户V标表
DROP TABLE IF EXISTS `user_v_badges`;
CREATE TABLE `user_v_badges` (
  `id` BIGINT NOT NULL AUTO_INCREMENT COMMENT '主键ID',
  `user_id` BIGINT NOT NULL DEFAULT 0 COMMENT '用户ID',
  `club_id` BIGINT NOT NULL DEFAULT 0 COMMENT '俱乐部ID',
  `badge_type` VARCHAR(32) NOT NULL DEFAULT 'none' COMMENT 'V标类型 0=none 1=blue(企业) 2=green(个人) 3=gold(平台)',
  `status` TINYINT NOT NULL DEFAULT 1 COMMENT '状态 1有效 0失效',
  `granted_at` DATETIME NULL DEFAULT NULL COMMENT '授予时间',
  `created_at` DATETIME NULL DEFAULT NULL COMMENT '创建时间',
  `updated_at` DATETIME NULL DEFAULT NULL COMMENT '更新时间',
  PRIMARY KEY (`id`),
  KEY `idx_user_id` (`user_id`),
  KEY `idx_club_id` (`club_id`),
  KEY `idx_badge_type` (`badge_type`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='用户V标表';

-- ============================================================
-- 模块13: 评价与信用
-- ============================================================

-- 评价表
DROP TABLE IF EXISTS `evaluations`;
CREATE TABLE `evaluations` (
  `id` BIGINT NOT NULL AUTO_INCREMENT COMMENT '主键ID',
  `order_id` BIGINT NOT NULL DEFAULT 0 COMMENT '订单ID',
  `user_id` BIGINT NOT NULL DEFAULT 0 COMMENT '客户ID',
  `player_id` BIGINT NOT NULL DEFAULT 0 COMMENT '打手ID',
  `score` INT NOT NULL DEFAULT 5 COMMENT '评分 1-5',
  `content` VARCHAR(500) NOT NULL DEFAULT '' COMMENT '评价内容',
  `status` VARCHAR(32) NOT NULL DEFAULT 'pending' COMMENT '状态 pending/displayed/deleted',
  `created_at` DATETIME NULL DEFAULT NULL COMMENT '创建时间',
  `display_at` DATETIME NULL DEFAULT NULL COMMENT '展示时间',
  `updated_at` DATETIME NULL DEFAULT NULL COMMENT '更新时间',
  PRIMARY KEY (`id`),
  KEY `idx_order_id` (`order_id`),
  KEY `idx_user_id` (`user_id`),
  KEY `idx_player_id` (`player_id`),
  KEY `idx_status` (`status`),
  KEY `idx_created_at` (`created_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='评价表';

-- 打赏表
DROP TABLE IF EXISTS `rewards`;
CREATE TABLE `rewards` (
  `id` BIGINT NOT NULL AUTO_INCREMENT COMMENT '主键ID',
  `order_id` BIGINT NOT NULL DEFAULT 0 COMMENT '订单ID',
  `user_id` BIGINT NOT NULL DEFAULT 0 COMMENT '客户ID',
  `player_id` BIGINT NOT NULL DEFAULT 0 COMMENT '打手ID',
  `amount` BIGINT NOT NULL DEFAULT 0 COMMENT '打赏金额(分)',
  `gift_type` VARCHAR(32) NOT NULL DEFAULT '' COMMENT '礼物类型',
  `created_at` DATETIME NULL DEFAULT NULL COMMENT '创建时间',
  PRIMARY KEY (`id`),
  KEY `idx_order_id` (`order_id`),
  KEY `idx_user_id` (`user_id`),
  KEY `idx_player_id` (`player_id`),
  KEY `idx_created_at` (`created_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='打赏表';

-- 收藏打手表
DROP TABLE IF EXISTS `player_favorites`;
CREATE TABLE `player_favorites` (
  `id` BIGINT NOT NULL AUTO_INCREMENT COMMENT '主键ID',
  `user_id` BIGINT NOT NULL DEFAULT 0 COMMENT '用户ID',
  `player_id` BIGINT NOT NULL DEFAULT 0 COMMENT '打手ID',
  `created_at` DATETIME NULL DEFAULT NULL COMMENT '创建时间',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_user_player` (`user_id`, `player_id`),
  KEY `idx_user_id` (`user_id`),
  KEY `idx_player_id` (`player_id`),
  KEY `idx_created_at` (`created_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='收藏打手表';

-- 打手标签表
DROP TABLE IF EXISTS `player_tags`;
CREATE TABLE `player_tags` (
  `id` BIGINT NOT NULL AUTO_INCREMENT COMMENT '主键ID',
  `player_id` BIGINT NOT NULL DEFAULT 0 COMMENT '打手ID',
  `tag` VARCHAR(64) NOT NULL DEFAULT '' COMMENT '标签',
  `created_at` DATETIME NULL DEFAULT NULL COMMENT '创建时间',
  PRIMARY KEY (`id`),
  KEY `idx_player_id` (`player_id`),
  KEY `idx_tag` (`tag`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='打手标签表';

-- 打手服务项目表
DROP TABLE IF EXISTS `player_services`;
CREATE TABLE `player_services` (
  `id` BIGINT NOT NULL AUTO_INCREMENT COMMENT '主键ID',
  `player_id` BIGINT NOT NULL DEFAULT 0 COMMENT '打手ID',
  `service_id` BIGINT NOT NULL DEFAULT 0 COMMENT '服务项目ID',
  `club_id` BIGINT NOT NULL DEFAULT 0 COMMENT '俱乐部ID',
  `status` TINYINT NOT NULL DEFAULT 1 COMMENT '状态 1上架 0下架',
  `created_at` DATETIME NULL DEFAULT NULL COMMENT '创建时间',
  `updated_at` DATETIME NULL DEFAULT NULL COMMENT '更新时间',
  PRIMARY KEY (`id`),
  KEY `idx_player_id` (`player_id`),
  KEY `idx_service_id` (`service_id`),
  KEY `idx_club_id` (`club_id`),
  KEY `idx_status` (`status`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='打手服务项目表';

-- 打手活跃度探针表
DROP TABLE IF EXISTS `player_probe_logs`;
CREATE TABLE `player_probe_logs` (
  `id` BIGINT NOT NULL AUTO_INCREMENT COMMENT '主键ID',
  `player_id` BIGINT NOT NULL DEFAULT 0 COMMENT '打手ID',
  `last_ping_at` DATETIME NULL DEFAULT NULL COMMENT '最后心跳时间',
  `ping_count` INT NOT NULL DEFAULT 0 COMMENT '心跳次数',
  `is_active` TINYINT NOT NULL DEFAULT 1 COMMENT '是否活跃 0否 1是',
  `created_at` DATETIME NULL DEFAULT NULL COMMENT '创建时间',
  `updated_at` DATETIME NULL DEFAULT NULL COMMENT '更新时间',
  PRIMARY KEY (`id`),
  KEY `idx_player_id` (`player_id`),
  KEY `idx_is_active` (`is_active`),
  KEY `idx_last_ping_at` (`last_ping_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='打手活跃度探针表';

-- ============================================================
-- 模块14: 服务项目
-- ============================================================

-- 服务类型表
DROP TABLE IF EXISTS `service_types`;
CREATE TABLE `service_types` (
  `id` BIGINT NOT NULL AUTO_INCREMENT COMMENT '主键ID',
  `name` VARCHAR(64) NOT NULL DEFAULT '' COMMENT '服务类型名称',
  `description` VARCHAR(255) NOT NULL DEFAULT '' COMMENT '描述',
  `icon` VARCHAR(512) NOT NULL DEFAULT '' COMMENT '图标URL',
  `sort_order` INT NOT NULL DEFAULT 0 COMMENT '排序',
  `status` TINYINT NOT NULL DEFAULT 1 COMMENT '状态 1启用 0停用',
  `created_at` DATETIME NULL DEFAULT NULL COMMENT '创建时间',
  `updated_at` DATETIME NULL DEFAULT NULL COMMENT '更新时间',
  PRIMARY KEY (`id`),
  KEY `idx_status` (`status`),
  KEY `idx_sort_order` (`sort_order`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='服务类型表';

-- 服务保证金表
DROP TABLE IF EXISTS `service_deposits`;
CREATE TABLE `service_deposits` (
  `id` BIGINT NOT NULL AUTO_INCREMENT COMMENT '主键ID',
  `player_id` BIGINT NOT NULL DEFAULT 0 COMMENT '打手ID',
  `amount` BIGINT NOT NULL DEFAULT 0 COMMENT '保证金金额(分)',
  `status` TINYINT NOT NULL DEFAULT 1 COMMENT '状态 1已缴 2已扣除 3已退还',
  `deducted_amount` BIGINT NOT NULL DEFAULT 0 COMMENT '已扣除金额(分)',
  `created_at` DATETIME NULL DEFAULT NULL COMMENT '创建时间',
  `updated_at` DATETIME NULL DEFAULT NULL COMMENT '更新时间',
  PRIMARY KEY (`id`),
  KEY `idx_player_id` (`player_id`),
  KEY `idx_status` (`status`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='服务保证金表';

-- 服务保证金日志表
DROP TABLE IF EXISTS `service_deposit_logs`;
CREATE TABLE `service_deposit_logs` (
  `id` BIGINT NOT NULL AUTO_INCREMENT COMMENT '主键ID',
  `deposit_id` BIGINT NOT NULL DEFAULT 0 COMMENT '保证金ID',
  `type` VARCHAR(32) NOT NULL DEFAULT '' COMMENT '类型 pay/deduct/refund',
  `amount` BIGINT NOT NULL DEFAULT 0 COMMENT '金额(分)',
  `reason` VARCHAR(255) NOT NULL DEFAULT '' COMMENT '原因',
  `created_at` DATETIME NULL DEFAULT NULL COMMENT '创建时间',
  PRIMARY KEY (`id`),
  KEY `idx_deposit_id` (`deposit_id`),
  KEY `idx_created_at` (`created_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='服务保证金日志表';

-- ============================================================
-- 模块15: 分销体系
-- ============================================================

-- 分销关系表
DROP TABLE IF EXISTS `distributor_relations`;
CREATE TABLE `distributor_relations` (
  `id` BIGINT NOT NULL AUTO_INCREMENT COMMENT '主键ID',
  `superior_id` BIGINT NOT NULL DEFAULT 0 COMMENT '上级分销商ID',
  `subordinate_id` BIGINT NOT NULL DEFAULT 0 COMMENT '下级用户ID',
  `level` TINYINT NOT NULL DEFAULT 1 COMMENT '级别 1=一级 2=二级',
  `is_valid` TINYINT NOT NULL DEFAULT 0 COMMENT '是否有效下级 0否 1是',
  `created_at` DATETIME NULL DEFAULT NULL COMMENT '创建时间',
  `updated_at` DATETIME NULL DEFAULT NULL COMMENT '更新时间',
  PRIMARY KEY (`id`),
  KEY `idx_superior_id` (`superior_id`),
  KEY `idx_subordinate_id` (`subordinate_id`),
  KEY `idx_level` (`level`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='分销关系表';

-- 分销佣金记录表
DROP TABLE IF EXISTS `distributor_commissions`;
CREATE TABLE `distributor_commissions` (
  `id` BIGINT NOT NULL AUTO_INCREMENT COMMENT '主键ID',
  `order_id` BIGINT NOT NULL DEFAULT 0 COMMENT '订单ID',
  `distributor_id` BIGINT NOT NULL DEFAULT 0 COMMENT '分销商ID',
  `amount` BIGINT NOT NULL DEFAULT 0 COMMENT '佣金金额(分)',
  `ratio` DECIMAL(5,2) NOT NULL DEFAULT 0.00 COMMENT '佣金比例%',
  `level` TINYINT NOT NULL DEFAULT 1 COMMENT '级别 1一级 2二级',
  `status` TINYINT NOT NULL DEFAULT 0 COMMENT '状态 0待结算 1已结算 2已回滚',
  `created_at` DATETIME NULL DEFAULT NULL COMMENT '创建时间',
  `updated_at` DATETIME NULL DEFAULT NULL COMMENT '更新时间',
  PRIMARY KEY (`id`),
  KEY `idx_order_id` (`order_id`),
  KEY `idx_distributor_id` (`distributor_id`),
  KEY `idx_status` (`status`),
  KEY `idx_created_at` (`created_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='分销佣金记录表';

-- 分销首单奖励表
DROP TABLE IF EXISTS `distributor_first_rewards`;
CREATE TABLE `distributor_first_rewards` (
  `id` BIGINT NOT NULL AUTO_INCREMENT COMMENT '主键ID',
  `distributor_id` BIGINT NOT NULL DEFAULT 0 COMMENT '分销商ID',
  `subordinate_id` BIGINT NOT NULL DEFAULT 0 COMMENT '下级ID',
  `order_id` BIGINT NOT NULL DEFAULT 0 COMMENT '订单ID',
  `reward_amount` BIGINT NOT NULL DEFAULT 0 COMMENT '奖励金额(分)',
  `created_at` DATETIME NULL DEFAULT NULL COMMENT '创建时间',
  PRIMARY KEY (`id`),
  KEY `idx_distributor_id` (`distributor_id`),
  KEY `idx_subordinate_id` (`subordinate_id`),
  KEY `idx_order_id` (`order_id`),
  KEY `idx_created_at` (`created_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='分销首单奖励表';

-- ============================================================
-- 模块16: 入会申请与考核
-- ============================================================

-- 入会申请表
DROP TABLE IF EXISTS `join_applications`;
CREATE TABLE `join_applications` (
  `id` BIGINT NOT NULL AUTO_INCREMENT COMMENT '主键ID',
  `club_id` BIGINT NOT NULL DEFAULT 0 COMMENT '俱乐部ID',
  `user_id` BIGINT NOT NULL DEFAULT 0 COMMENT '用户ID',
  `real_name` VARCHAR(32) NOT NULL DEFAULT '' COMMENT '真实姓名',
  `game_account` VARCHAR(128) NOT NULL DEFAULT '' COMMENT '游戏账号',
  `game_region` VARCHAR(64) NOT NULL DEFAULT '' COMMENT '游戏大区',
  `good_position` VARCHAR(64) NOT NULL DEFAULT '' COMMENT '擅长位置',
  `rank_level` VARCHAR(32) NOT NULL DEFAULT '' COMMENT '段位',
  `intro` VARCHAR(500) NOT NULL DEFAULT '' COMMENT '自我介绍',
  `status` VARCHAR(32) NOT NULL DEFAULT 'pending' COMMENT '状态 pending/examining/approved/rejected',
  `created_at` DATETIME NULL DEFAULT NULL COMMENT '创建时间',
  `updated_at` DATETIME NULL DEFAULT NULL COMMENT '更新时间',
  PRIMARY KEY (`id`),
  KEY `idx_club_id` (`club_id`),
  KEY `idx_user_id` (`user_id`),
  KEY `idx_status` (`status`),
  KEY `idx_created_at` (`created_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='入会申请表';

-- 考核记录表
DROP TABLE IF EXISTS `exam_records`;
CREATE TABLE `exam_records` (
  `id` BIGINT NOT NULL AUTO_INCREMENT COMMENT '主键ID',
  `application_id` BIGINT NOT NULL DEFAULT 0 COMMENT '申请ID',
  `examiner_id` BIGINT NOT NULL DEFAULT 0 COMMENT '考核人ID',
  `requirement` VARCHAR(255) NOT NULL DEFAULT '' COMMENT '考核要求',
  `result` VARCHAR(32) NOT NULL DEFAULT '' COMMENT '考核结果 pass/fail',
  `remark` VARCHAR(500) NOT NULL DEFAULT '' COMMENT '备注',
  `video_url` VARCHAR(512) NOT NULL DEFAULT '' COMMENT '考核视频URL',
  `created_at` DATETIME NULL DEFAULT NULL COMMENT '创建时间',
  `updated_at` DATETIME NULL DEFAULT NULL COMMENT '更新时间',
  PRIMARY KEY (`id`),
  KEY `idx_application_id` (`application_id`),
  KEY `idx_examiner_id` (`examiner_id`),
  KEY `idx_result` (`result`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='考核记录表';

-- 考核模板表
DROP TABLE IF EXISTS `exam_templates`;
CREATE TABLE `exam_templates` (
  `id` BIGINT NOT NULL AUTO_INCREMENT COMMENT '主键ID',
  `club_id` BIGINT NOT NULL DEFAULT 0 COMMENT '俱乐部ID',
  `game` VARCHAR(64) NOT NULL DEFAULT '' COMMENT '游戏',
  `rank_level` VARCHAR(32) NOT NULL DEFAULT '' COMMENT '段位',
  `standard` TEXT COMMENT '考核标准',
  `created_at` DATETIME NULL DEFAULT NULL COMMENT '创建时间',
  `updated_at` DATETIME NULL DEFAULT NULL COMMENT '更新时间',
  PRIMARY KEY (`id`),
  KEY `idx_club_id` (`club_id`),
  KEY `idx_game` (`game`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='考核模板表';

-- ============================================================
-- 模块17: 风控与安全
-- ============================================================

-- 敏感词库表
DROP TABLE IF EXISTS `sensitive_words`;
CREATE TABLE `sensitive_words` (
  `id` BIGINT NOT NULL AUTO_INCREMENT COMMENT '主键ID',
  `word` VARCHAR(128) NOT NULL DEFAULT '' COMMENT '敏感词',
  `category` VARCHAR(32) NOT NULL DEFAULT '' COMMENT '分类 fraud/boosting/gambling/etc',
  `enabled` TINYINT NOT NULL DEFAULT 1 COMMENT '是否启用 0否 1是',
  `created_at` DATETIME NULL DEFAULT NULL COMMENT '创建时间',
  `updated_at` DATETIME NULL DEFAULT NULL COMMENT '更新时间',
  PRIMARY KEY (`id`),
  KEY `idx_category` (`category`),
  KEY `idx_enabled` (`enabled`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='敏感词库表';

-- 风控日志表
DROP TABLE IF EXISTS `risk_control_logs`;
CREATE TABLE `risk_control_logs` (
  `id` BIGINT NOT NULL AUTO_INCREMENT COMMENT '主键ID',
  `user_id` BIGINT NOT NULL DEFAULT 0 COMMENT '用户ID',
  `risk_type` VARCHAR(32) NOT NULL DEFAULT '' COMMENT '风险类型',
  `description` VARCHAR(255) NOT NULL DEFAULT '' COMMENT '风险描述',
  `level` TINYINT NOT NULL DEFAULT 1 COMMENT '风险等级 1低 2中 3高 4最高',
  `created_at` DATETIME NULL DEFAULT NULL COMMENT '创建时间',
  PRIMARY KEY (`id`),
  KEY `idx_user_id` (`user_id`),
  KEY `idx_risk_type` (`risk_type`),
  KEY `idx_level` (`level`),
  KEY `idx_created_at` (`created_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='风控日志表';

-- 风险用户表
DROP TABLE IF EXISTS `risk_users`;
CREATE TABLE `risk_users` (
  `id` BIGINT NOT NULL AUTO_INCREMENT COMMENT '主键ID',
  `user_id` BIGINT NOT NULL DEFAULT 0 COMMENT '用户ID',
  `risk_level` VARCHAR(32) NOT NULL DEFAULT 'low' COMMENT '风险等级 low/medium/high/critical',
  `risk_type` VARCHAR(32) NOT NULL DEFAULT '' COMMENT '风险类型',
  `marked_at` DATETIME NULL DEFAULT NULL COMMENT '标记时间',
  `created_at` DATETIME NULL DEFAULT NULL COMMENT '创建时间',
  `updated_at` DATETIME NULL DEFAULT NULL COMMENT '更新时间',
  PRIMARY KEY (`id`),
  KEY `idx_user_id` (`user_id`),
  KEY `idx_risk_level` (`risk_level`),
  KEY `idx_risk_type` (`risk_type`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='风险用户表';

-- AI风险预警表
DROP TABLE IF EXISTS `ai_risk_alerts`;
CREATE TABLE `ai_risk_alerts` (
  `id` BIGINT NOT NULL AUTO_INCREMENT COMMENT '主键ID',
  `alert_type` VARCHAR(32) NOT NULL DEFAULT '' COMMENT '预警类型',
  `target_type` VARCHAR(32) NOT NULL DEFAULT '' COMMENT '目标类型 user/order/club',
  `target_id` BIGINT NOT NULL DEFAULT 0 COMMENT '目标ID',
  `description` VARCHAR(255) NOT NULL DEFAULT '' COMMENT '风险描述',
  `level` TINYINT NOT NULL DEFAULT 1 COMMENT '风险等级 1低 2中 3高 4最高',
  `status` TINYINT NOT NULL DEFAULT 0 COMMENT '状态 0待处理 1已处理 2已忽略',
  `created_at` DATETIME NULL DEFAULT NULL COMMENT '创建时间',
  `updated_at` DATETIME NULL DEFAULT NULL COMMENT '更新时间',
  PRIMARY KEY (`id`),
  KEY `idx_alert_type` (`alert_type`),
  KEY `idx_target_type` (`target_type`),
  KEY `idx_target_id` (`target_id`),
  KEY `idx_status` (`status`),
  KEY `idx_created_at` (`created_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='AI风险预警表';

-- ============================================================
-- 模块18: 未成年人保护
-- ============================================================

-- 未成年宵禁日志表
DROP TABLE IF EXISTS `minor_curfew_logs`;
CREATE TABLE `minor_curfew_logs` (
  `id` BIGINT NOT NULL AUTO_INCREMENT COMMENT '主键ID',
  `user_id` BIGINT NOT NULL DEFAULT 0 COMMENT '用户ID',
  `action` VARCHAR(32) NOT NULL DEFAULT '' COMMENT '尝试操作 order/pay/reward',
  `blocked_at` DATETIME NULL DEFAULT NULL COMMENT '拦截时间',
  `created_at` DATETIME NULL DEFAULT NULL COMMENT '创建时间',
  PRIMARY KEY (`id`),
  KEY `idx_user_id` (`user_id`),
  KEY `idx_created_at` (`created_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='未成年宵禁日志表';

-- 未成年消费预警表
DROP TABLE IF EXISTS `minor_consume_warnings`;
CREATE TABLE `minor_consume_warnings` (
  `id` BIGINT NOT NULL AUTO_INCREMENT COMMENT '主键ID',
  `user_id` BIGINT NOT NULL DEFAULT 0 COMMENT '用户ID',
  `month_amount` BIGINT NOT NULL DEFAULT 0 COMMENT '当月消费金额(分)',
  `threshold` BIGINT NOT NULL DEFAULT 0 COMMENT '预警阈值(分)',
  `warning_level` TINYINT NOT NULL DEFAULT 1 COMMENT '预警等级 1低 2中 3高',
  `notified` TINYINT NOT NULL DEFAULT 0 COMMENT '是否已通知 0否 1是',
  `created_at` DATETIME NULL DEFAULT NULL COMMENT '创建时间',
  `updated_at` DATETIME NULL DEFAULT NULL COMMENT '更新时间',
  PRIMARY KEY (`id`),
  KEY `idx_user_id` (`user_id`),
  KEY `idx_warning_level` (`warning_level`),
  KEY `idx_created_at` (`created_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='未成年消费预警表';

-- 家长绑定表
DROP TABLE IF EXISTS `parent_guardian_binds`;
CREATE TABLE `parent_guardian_binds` (
  `id` BIGINT NOT NULL AUTO_INCREMENT COMMENT '主键ID',
  `parent_uid` BIGINT NOT NULL DEFAULT 0 COMMENT '家长用户ID',
  `child_uid` BIGINT NOT NULL DEFAULT 0 COMMENT '未成年用户ID',
  `verified_at` DATETIME NULL DEFAULT NULL COMMENT '验证时间',
  `created_at` DATETIME NULL DEFAULT NULL COMMENT '创建时间',
  `updated_at` DATETIME NULL DEFAULT NULL COMMENT '更新时间',
  PRIMARY KEY (`id`),
  KEY `idx_parent_uid` (`parent_uid`),
  KEY `idx_child_uid` (`child_uid`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='家长绑定表';

-- 家长设置表
DROP TABLE IF EXISTS `parent_guardian_settings`;
CREATE TABLE `parent_guardian_settings` (
  `id` BIGINT NOT NULL AUTO_INCREMENT COMMENT '主键ID',
  `parent_uid` BIGINT NOT NULL DEFAULT 0 COMMENT '家长用户ID',
  `child_uid` BIGINT NOT NULL DEFAULT 0 COMMENT '未成年用户ID',
  `monthly_limit` BIGINT NOT NULL DEFAULT 0 COMMENT '月度消费限额(分)',
  `allow_order` TINYINT NOT NULL DEFAULT 1 COMMENT '允许下单 0否 1是',
  `allow_reward` TINYINT NOT NULL DEFAULT 1 COMMENT '允许打赏 0否 1是',
  `is_frozen` TINYINT NOT NULL DEFAULT 0 COMMENT '是否冻结账户 0否 1是',
  `created_at` DATETIME NULL DEFAULT NULL COMMENT '创建时间',
  `updated_at` DATETIME NULL DEFAULT NULL COMMENT '更新时间',
  PRIMARY KEY (`id`),
  KEY `idx_parent_uid` (`parent_uid`),
  KEY `idx_child_uid` (`child_uid`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='家长设置表';

-- 家长消费报告表
DROP TABLE IF EXISTS `parent_consume_reports`;
CREATE TABLE `parent_consume_reports` (
  `id` BIGINT NOT NULL AUTO_INCREMENT COMMENT '主键ID',
  `child_uid` BIGINT NOT NULL DEFAULT 0 COMMENT '未成年用户ID',
  `parent_uid` BIGINT NOT NULL DEFAULT 0 COMMENT '家长用户ID',
  `period` VARCHAR(16) NOT NULL DEFAULT '' COMMENT '统计周期 YYYY-MM',
  `total_amount` BIGINT NOT NULL DEFAULT 0 COMMENT '消费总额(分)',
  `order_count` INT NOT NULL DEFAULT 0 COMMENT '订单数',
  `created_at` DATETIME NULL DEFAULT NULL COMMENT '创建时间',
  PRIMARY KEY (`id`),
  KEY `idx_child_uid` (`child_uid`),
  KEY `idx_parent_uid` (`parent_uid`),
  KEY `idx_period` (`period`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='家长消费报告表';

-- ============================================================
-- 模块19: 系统配置
-- ============================================================

-- 系统配置表
DROP TABLE IF EXISTS `system_configs`;
CREATE TABLE `system_configs` (
  `id` BIGINT NOT NULL AUTO_INCREMENT COMMENT '主键ID',
  `key` VARCHAR(128) NOT NULL DEFAULT '' COMMENT '配置键',
  `value` TEXT COMMENT '配置值',
  `description` VARCHAR(255) NOT NULL DEFAULT '' COMMENT '配置描述',
  `updated_at` DATETIME NULL DEFAULT NULL COMMENT '更新时间',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_key` (`key`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='系统配置表';

-- 接口监控表
DROP TABLE IF EXISTS `system_api_monitors`;
CREATE TABLE `system_api_monitors` (
  `id` BIGINT NOT NULL AUTO_INCREMENT COMMENT '主键ID',
  `api_name` VARCHAR(128) NOT NULL DEFAULT '' COMMENT '接口名',
  `provider` VARCHAR(64) NOT NULL DEFAULT '' COMMENT '服务商',
  `success_count` BIGINT NOT NULL DEFAULT 0 COMMENT '成功次数',
  `fail_count` BIGINT NOT NULL DEFAULT 0 COMMENT '失败次数',
  `avg_latency_ms` INT NOT NULL DEFAULT 0 COMMENT '平均延迟(ms)',
  `alert_status` TINYINT NOT NULL DEFAULT 0 COMMENT '告警状态 0正常 1告警',
  `created_at` DATETIME NULL DEFAULT NULL COMMENT '创建时间',
  `updated_at` DATETIME NULL DEFAULT NULL COMMENT '更新时间',
  PRIMARY KEY (`id`),
  KEY `idx_api_name` (`api_name`),
  KEY `idx_alert_status` (`alert_status`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='接口监控表';

-- 慢查询日志表
DROP TABLE IF EXISTS `system_slow_query_logs`;
CREATE TABLE `system_slow_query_logs` (
  `id` BIGINT NOT NULL AUTO_INCREMENT COMMENT '主键ID',
  `query` TEXT COMMENT 'SQL语句',
  `duration_ms` INT NOT NULL DEFAULT 0 COMMENT '耗时(ms)',
  `created_at` DATETIME NULL DEFAULT NULL COMMENT '创建时间',
  PRIMARY KEY (`id`),
  KEY `idx_duration_ms` (`duration_ms`),
  KEY `idx_created_at` (`created_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='慢查询日志表';

-- ============================================================
-- 模块20: 运营管理
-- ============================================================

-- 操作日志表
DROP TABLE IF EXISTS `operation_logs`;
CREATE TABLE `operation_logs` (
  `id` BIGINT NOT NULL AUTO_INCREMENT COMMENT '主键ID',
  `operator_id` BIGINT NOT NULL DEFAULT 0 COMMENT '操作人ID',
  `operator_type` VARCHAR(32) NOT NULL DEFAULT '' COMMENT '操作人类型 admin/shop_admin',
  `action` VARCHAR(64) NOT NULL DEFAULT '' COMMENT '操作动作',
  `target_type` VARCHAR(64) NOT NULL DEFAULT '' COMMENT '操作对象类型',
  `target_id` BIGINT NOT NULL DEFAULT 0 COMMENT '操作对象ID',
  `content` JSON COMMENT '操作内容',
  `ip` VARCHAR(64) NOT NULL DEFAULT '' COMMENT 'IP地址',
  `device_info` VARCHAR(255) NOT NULL DEFAULT '' COMMENT '设备信息',
  `created_at` DATETIME NULL DEFAULT NULL COMMENT '创建时间',
  PRIMARY KEY (`id`),
  KEY `idx_operator_id` (`operator_id`),
  KEY `idx_created_at` (`created_at`),
  KEY `idx_action` (`action`),
  KEY `idx_target_type` (`target_type`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='操作日志表';

-- 导出日志表
DROP TABLE IF EXISTS `export_logs`;
CREATE TABLE `export_logs` (
  `id` BIGINT NOT NULL AUTO_INCREMENT COMMENT '主键ID',
  `operator_id` BIGINT NOT NULL DEFAULT 0 COMMENT '操作人ID',
  `target_type` VARCHAR(64) NOT NULL DEFAULT '' COMMENT '导出对象类型',
  `file_url` VARCHAR(512) NOT NULL DEFAULT '' COMMENT '文件URL',
  `watermark` VARCHAR(128) NOT NULL DEFAULT '' COMMENT '水印信息',
  `created_at` DATETIME NULL DEFAULT NULL COMMENT '创建时间',
  PRIMARY KEY (`id`),
  KEY `idx_operator_id` (`operator_id`),
  KEY `idx_target_type` (`target_type`),
  KEY `idx_created_at` (`created_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='导出日志表';

-- 批量操作确认表
DROP TABLE IF EXISTS `batch_operation_confirms`;
CREATE TABLE `batch_operation_confirms` (
  `id` BIGINT NOT NULL AUTO_INCREMENT COMMENT '主键ID',
  `operator_id` BIGINT NOT NULL DEFAULT 0 COMMENT '操作人ID',
  `action` VARCHAR(64) NOT NULL DEFAULT '' COMMENT '操作动作',
  `params` JSON COMMENT '操作参数',
  `confirmer_id` BIGINT NOT NULL DEFAULT 0 COMMENT '确认人ID',
  `status` TINYINT NOT NULL DEFAULT 0 COMMENT '状态 0待确认 1已确认 2已拒绝',
  `created_at` DATETIME NULL DEFAULT NULL COMMENT '创建时间',
  `updated_at` DATETIME NULL DEFAULT NULL COMMENT '更新时间',
  PRIMARY KEY (`id`),
  KEY `idx_operator_id` (`operator_id`),
  KEY `idx_confirmer_id` (`confirmer_id`),
  KEY `idx_status` (`status`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='批量操作确认表';

-- 批量操作日志表
DROP TABLE IF EXISTS `batch_operation_logs`;
CREATE TABLE `batch_operation_logs` (
  `id` BIGINT NOT NULL AUTO_INCREMENT COMMENT '主键ID',
  `batch_id` BIGINT NOT NULL DEFAULT 0 COMMENT '批次ID',
  `item_id` BIGINT NOT NULL DEFAULT 0 COMMENT '操作项ID',
  `result` VARCHAR(32) NOT NULL DEFAULT '' COMMENT '操作结果 success/fail',
  `created_at` DATETIME NULL DEFAULT NULL COMMENT '创建时间',
  PRIMARY KEY (`id`),
  KEY `idx_batch_id` (`batch_id`),
  KEY `idx_created_at` (`created_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='批量操作日志表';

-- 通知表
DROP TABLE IF EXISTS `notifications`;
CREATE TABLE `notifications` (
  `id` BIGINT NOT NULL AUTO_INCREMENT COMMENT '主键ID',
  `user_id` BIGINT NOT NULL DEFAULT 0 COMMENT '用户ID',
  `type` VARCHAR(32) NOT NULL DEFAULT '' COMMENT '通知类型',
  `title` VARCHAR(128) NOT NULL DEFAULT '' COMMENT '标题',
  `content` TEXT COMMENT '内容',
  `is_read` TINYINT NOT NULL DEFAULT 0 COMMENT '是否已读 0否 1是',
  `category` VARCHAR(32) NOT NULL DEFAULT 'system' COMMENT '分类 pending/supervision/system',
  `created_at` DATETIME NULL DEFAULT NULL COMMENT '创建时间',
  `updated_at` DATETIME NULL DEFAULT NULL COMMENT '更新时间',
  PRIMARY KEY (`id`),
  KEY `idx_user_id` (`user_id`),
  KEY `idx_category` (`category`),
  KEY `idx_is_read` (`is_read`),
  KEY `idx_created_at` (`created_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='通知表';

-- 订阅消息模板表
DROP TABLE IF EXISTS `subscribe_message_templates`;
CREATE TABLE `subscribe_message_templates` (
  `id` BIGINT NOT NULL AUTO_INCREMENT COMMENT '主键ID',
  `template_id` VARCHAR(64) NOT NULL DEFAULT '' COMMENT '微信模板ID',
  `name` VARCHAR(64) NOT NULL DEFAULT '' COMMENT '模板名称',
  `type` VARCHAR(32) NOT NULL DEFAULT '' COMMENT '模板类型',
  `content` TEXT COMMENT '模板内容',
  `created_at` DATETIME NULL DEFAULT NULL COMMENT '创建时间',
  `updated_at` DATETIME NULL DEFAULT NULL COMMENT '更新时间',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_template_id` (`template_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='订阅消息模板表';

-- 订阅消息发送日志表
DROP TABLE IF EXISTS `subscribe_message_logs`;
CREATE TABLE `subscribe_message_logs` (
  `id` BIGINT NOT NULL AUTO_INCREMENT COMMENT '主键ID',
  `template_id` VARCHAR(64) NOT NULL DEFAULT '' COMMENT '模板ID',
  `user_id` BIGINT NOT NULL DEFAULT 0 COMMENT '用户ID',
  `order_id` BIGINT NOT NULL DEFAULT 0 COMMENT '订单ID',
  `status` TINYINT NOT NULL DEFAULT 0 COMMENT '发送状态 0待发送 1成功 2失败',
  `sent_at` DATETIME NULL DEFAULT NULL COMMENT '发送时间',
  `created_at` DATETIME NULL DEFAULT NULL COMMENT '创建时间',
  PRIMARY KEY (`id`),
  KEY `idx_template_id` (`template_id`),
  KEY `idx_user_id` (`user_id`),
  KEY `idx_status` (`status`),
  KEY `idx_created_at` (`created_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='订阅消息发送日志表';

-- ============================================================
-- 模块21: 运维
-- ============================================================

-- 备份记录表
DROP TABLE IF EXISTS `backup_records`;
CREATE TABLE `backup_records` (
  `id` BIGINT NOT NULL AUTO_INCREMENT COMMENT '主键ID',
  `file_name` VARCHAR(255) NOT NULL DEFAULT '' COMMENT '文件名',
  `file_size` BIGINT NOT NULL DEFAULT 0 COMMENT '文件大小(字节)',
  `encrypted` TINYINT NOT NULL DEFAULT 0 COMMENT '是否加密 0否 1是',
  `oss_url` VARCHAR(512) NOT NULL DEFAULT '' COMMENT 'OSS存储URL',
  `created_at` DATETIME NULL DEFAULT NULL COMMENT '创建时间',
  PRIMARY KEY (`id`),
  KEY `idx_created_at` (`created_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='备份记录表';

-- 恢复记录表
DROP TABLE IF EXISTS `restore_records`;
CREATE TABLE `restore_records` (
  `id` BIGINT NOT NULL AUTO_INCREMENT COMMENT '主键ID',
  `backup_id` BIGINT NOT NULL DEFAULT 0 COMMENT '备份ID',
  `restored_at` DATETIME NULL DEFAULT NULL COMMENT '恢复时间',
  `operator_id` BIGINT NOT NULL DEFAULT 0 COMMENT '操作人ID',
  `created_at` DATETIME NULL DEFAULT NULL COMMENT '创建时间',
  PRIMARY KEY (`id`),
  KEY `idx_backup_id` (`backup_id`),
  KEY `idx_created_at` (`created_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='恢复记录表';

-- 灰度发布表
DROP TABLE IF EXISTS `gray_releases`;
CREATE TABLE `gray_releases` (
  `id` BIGINT NOT NULL AUTO_INCREMENT COMMENT '主键ID',
  `api_version` VARCHAR(32) NOT NULL DEFAULT '' COMMENT 'API版本',
  `whitelist` JSON COMMENT '白名单',
  `rollout_percent` INT NOT NULL DEFAULT 0 COMMENT '灰度比例%',
  `error_rate_threshold` DECIMAL(5,2) NOT NULL DEFAULT 0.00 COMMENT '错误率阈值%',
  `status` TINYINT NOT NULL DEFAULT 1 COMMENT '状态 1启用 0停用',
  `created_at` DATETIME NULL DEFAULT NULL COMMENT '创建时间',
  `updated_at` DATETIME NULL DEFAULT NULL COMMENT '更新时间',
  PRIMARY KEY (`id`),
  KEY `idx_status` (`status`),
  KEY `idx_created_at` (`created_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='灰度发布表';

-- 熔断记录表
DROP TABLE IF EXISTS `circuit_breakers`;
CREATE TABLE `circuit_breakers` (
  `id` BIGINT NOT NULL AUTO_INCREMENT COMMENT '主键ID',
  `service_name` VARCHAR(64) NOT NULL DEFAULT '' COMMENT '服务名',
  `state` VARCHAR(32) NOT NULL DEFAULT 'closed' COMMENT '状态 closed/open/half_open',
  `fail_count` INT NOT NULL DEFAULT 0 COMMENT '失败次数',
  `last_fail_at` DATETIME NULL DEFAULT NULL COMMENT '最后失败时间',
  `created_at` DATETIME NULL DEFAULT NULL COMMENT '创建时间',
  `updated_at` DATETIME NULL DEFAULT NULL COMMENT '更新时间',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_service_name` (`service_name`),
  KEY `idx_state` (`state`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='熔断记录表';

-- 第三方接口日志表
DROP TABLE IF EXISTS `third_party_api_logs`;
CREATE TABLE `third_party_api_logs` (
  `id` BIGINT NOT NULL AUTO_INCREMENT COMMENT '主键ID',
  `api_name` VARCHAR(128) NOT NULL DEFAULT '' COMMENT '接口名',
  `request` TEXT COMMENT '请求数据',
  `response` TEXT COMMENT '响应数据',
  `latency_ms` INT NOT NULL DEFAULT 0 COMMENT '延迟(ms)',
  `status` TINYINT NOT NULL DEFAULT 1 COMMENT '状态 1成功 2失败',
  `created_at` DATETIME NULL DEFAULT NULL COMMENT '创建时间',
  PRIMARY KEY (`id`),
  KEY `idx_api_name` (`api_name`),
  KEY `idx_status` (`status`),
  KEY `idx_created_at` (`created_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='第三方接口日志表';

-- 第三方接口重试队列表
DROP TABLE IF EXISTS `third_party_retry_queues`;
CREATE TABLE `third_party_retry_queues` (
  `id` BIGINT NOT NULL AUTO_INCREMENT COMMENT '主键ID',
  `api_name` VARCHAR(128) NOT NULL DEFAULT '' COMMENT '接口名',
  `params` JSON COMMENT '请求参数',
  `retry_count` INT NOT NULL DEFAULT 0 COMMENT '已重试次数',
  `max_retries` INT NOT NULL DEFAULT 3 COMMENT '最大重试次数',
  `next_retry_at` DATETIME NULL DEFAULT NULL COMMENT '下次重试时间',
  `status` TINYINT NOT NULL DEFAULT 0 COMMENT '状态 0待重试 1成功 2失败',
  `created_at` DATETIME NULL DEFAULT NULL COMMENT '创建时间',
  `updated_at` DATETIME NULL DEFAULT NULL COMMENT '更新时间',
  PRIMARY KEY (`id`),
  KEY `idx_status` (`status`),
  KEY `idx_next_retry_at` (`next_retry_at`),
  KEY `idx_created_at` (`created_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='第三方接口重试队列表';

-- 监控告警表
DROP TABLE IF EXISTS `monitor_alerts`;
CREATE TABLE `monitor_alerts` (
  `id` BIGINT NOT NULL AUTO_INCREMENT COMMENT '主键ID',
  `alert_type` VARCHAR(64) NOT NULL DEFAULT '' COMMENT '告警类型',
  `message` VARCHAR(500) NOT NULL DEFAULT '' COMMENT '告警信息',
  `level` TINYINT NOT NULL DEFAULT 1 COMMENT '告警等级 1低 2中 3高 4最高',
  `status` TINYINT NOT NULL DEFAULT 0 COMMENT '状态 0未处理 1已处理',
  `created_at` DATETIME NULL DEFAULT NULL COMMENT '创建时间',
  `updated_at` DATETIME NULL DEFAULT NULL COMMENT '更新时间',
  PRIMARY KEY (`id`),
  KEY `idx_alert_type` (`alert_type`),
  KEY `idx_level` (`level`),
  KEY `idx_status` (`status`),
  KEY `idx_created_at` (`created_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='监控告警表';

-- NTP同步日志表
DROP TABLE IF EXISTS `ntp_sync_logs`;
CREATE TABLE `ntp_sync_logs` (
  `id` BIGINT NOT NULL AUTO_INCREMENT COMMENT '主键ID',
  `offset_seconds` DECIMAL(10,3) NOT NULL DEFAULT 0.000 COMMENT '时间偏移(秒)',
  `synced` TINYINT NOT NULL DEFAULT 0 COMMENT '是否同步成功 0否 1是',
  `created_at` DATETIME NULL DEFAULT NULL COMMENT '创建时间',
  PRIMARY KEY (`id`),
  KEY `idx_created_at` (`created_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='NTP同步日志表';

-- 定时任务日志表
DROP TABLE IF EXISTS `cron_job_logs`;
CREATE TABLE `cron_job_logs` (
  `id` BIGINT NOT NULL AUTO_INCREMENT COMMENT '主键ID',
  `job_name` VARCHAR(64) NOT NULL DEFAULT '' COMMENT '任务名',
  `status` TINYINT NOT NULL DEFAULT 1 COMMENT '状态 1成功 2失败',
  `duration_ms` INT NOT NULL DEFAULT 0 COMMENT '耗时(ms)',
  `created_at` DATETIME NULL DEFAULT NULL COMMENT '创建时间',
  PRIMARY KEY (`id`),
  KEY `idx_job_name` (`job_name`),
  KEY `idx_status` (`status`),
  KEY `idx_created_at` (`created_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='定时任务日志表';

-- WebSocket连接表
DROP TABLE IF EXISTS `ws_connections`;
CREATE TABLE `ws_connections` (
  `id` BIGINT NOT NULL AUTO_INCREMENT COMMENT '主键ID',
  `user_id` BIGINT NOT NULL DEFAULT 0 COMMENT '用户ID',
  `user_type` VARCHAR(32) NOT NULL DEFAULT '' COMMENT '用户类型 user/admin/platform',
  `conn_id` VARCHAR(64) NOT NULL DEFAULT '' COMMENT '连接ID',
  `last_ping_at` DATETIME NULL DEFAULT NULL COMMENT '最后心跳时间',
  `is_active` TINYINT NOT NULL DEFAULT 1 COMMENT '是否活跃 0否 1是',
  `created_at` DATETIME NULL DEFAULT NULL COMMENT '创建时间',
  `updated_at` DATETIME NULL DEFAULT NULL COMMENT '更新时间',
  PRIMARY KEY (`id`),
  KEY `idx_user_id` (`user_id`),
  KEY `idx_conn_id` (`conn_id`),
  KEY `idx_is_active` (`is_active`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='WebSocket连接表';

-- 离线消息表
DROP TABLE IF EXISTS `offline_messages`;
CREATE TABLE `offline_messages` (
  `id` BIGINT NOT NULL AUTO_INCREMENT COMMENT '主键ID',
  `user_id` BIGINT NOT NULL DEFAULT 0 COMMENT '用户ID',
  `session_id` BIGINT NOT NULL DEFAULT 0 COMMENT '会话ID',
  `message_data` JSON COMMENT '消息数据',
  `created_at` DATETIME NULL DEFAULT NULL COMMENT '创建时间',
  PRIMARY KEY (`id`),
  KEY `idx_user_id` (`user_id`),
  KEY `idx_session_id` (`session_id`),
  KEY `idx_created_at` (`created_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='离线消息表';

-- ============================================================
-- 模块22: UP主认证
-- ============================================================

-- UP主认证表
DROP TABLE IF EXISTS `up_master_certifications`;
CREATE TABLE `up_master_certifications` (
  `id` BIGINT NOT NULL AUTO_INCREMENT COMMENT '主键ID',
  `user_id` BIGINT NOT NULL DEFAULT 0 COMMENT '用户ID',
  `platform` VARCHAR(64) NOT NULL DEFAULT '' COMMENT '平台',
  `follower_count` INT NOT NULL DEFAULT 0 COMMENT '粉丝数',
  `level` TINYINT NOT NULL DEFAULT 1 COMMENT '等级 1-6',
  `status` TINYINT NOT NULL DEFAULT 0 COMMENT '状态 0待审核 1通过 2驳回 3已吊销',
  `verified_at` DATETIME NULL DEFAULT NULL COMMENT '认证时间',
  `created_at` DATETIME NULL DEFAULT NULL COMMENT '创建时间',
  `updated_at` DATETIME NULL DEFAULT NULL COMMENT '更新时间',
  PRIMARY KEY (`id`),
  KEY `idx_user_id` (`user_id`),
  KEY `idx_level` (`level`),
  KEY `idx_status` (`status`),
  KEY `idx_created_at` (`created_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='UP主认证表';

-- UP主等级配置表
DROP TABLE IF EXISTS `up_master_tier_configs`;
CREATE TABLE `up_master_tier_configs` (
  `id` BIGINT NOT NULL AUTO_INCREMENT COMMENT '主键ID',
  `tier_name` VARCHAR(32) NOT NULL DEFAULT '' COMMENT '等级名称',
  `min_followers` INT NOT NULL DEFAULT 0 COMMENT '最低粉丝数',
  `max_followers` INT NOT NULL DEFAULT 0 COMMENT '最高粉丝数',
  `badge` VARCHAR(64) NOT NULL DEFAULT '' COMMENT '徽章',
  `benefits` JSON COMMENT '权益配置',
  `created_at` DATETIME NULL DEFAULT NULL COMMENT '创建时间',
  `updated_at` DATETIME NULL DEFAULT NULL COMMENT '更新时间',
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='UP主等级配置表';

-- ============================================================
-- 模块23: 营销裂变
-- ============================================================

-- 优惠券模板表
DROP TABLE IF EXISTS `coupon_templates`;
CREATE TABLE `coupon_templates` (
  `id` BIGINT NOT NULL AUTO_INCREMENT COMMENT '主键ID',
  `name` VARCHAR(64) NOT NULL DEFAULT '' COMMENT '券名称',
  `type` VARCHAR(32) NOT NULL DEFAULT '' COMMENT '类型 newuser/fullcut/discount/recharge/compensation',
  `amount` BIGINT NOT NULL DEFAULT 0 COMMENT '优惠金额(分)',
  `min_spend` BIGINT NOT NULL DEFAULT 0 COMMENT '最低消费(分)',
  `discount_ratio` DECIMAL(3,2) NOT NULL DEFAULT 0.00 COMMENT '折扣比例',
  `valid_days` INT NOT NULL DEFAULT 0 COMMENT '有效天数',
  `total_count` INT NOT NULL DEFAULT 0 COMMENT '发放总量',
  `issued_count` INT NOT NULL DEFAULT 0 COMMENT '已发放数',
  `status` TINYINT NOT NULL DEFAULT 1 COMMENT '状态 1启用 0停用',
  `created_at` DATETIME NULL DEFAULT NULL COMMENT '创建时间',
  `updated_at` DATETIME NULL DEFAULT NULL COMMENT '更新时间',
  PRIMARY KEY (`id`),
  KEY `idx_type` (`type`),
  KEY `idx_status` (`status`),
  KEY `idx_created_at` (`created_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='优惠券模板表';

-- 用户优惠券表
DROP TABLE IF EXISTS `user_coupons`;
CREATE TABLE `user_coupons` (
  `id` BIGINT NOT NULL AUTO_INCREMENT COMMENT '主键ID',
  `user_id` BIGINT NOT NULL DEFAULT 0 COMMENT '用户ID',
  `template_id` BIGINT NOT NULL DEFAULT 0 COMMENT '模板ID',
  `status` VARCHAR(32) NOT NULL DEFAULT 'unused' COMMENT '状态 unused/used/expired',
  `used_at` DATETIME NULL DEFAULT NULL COMMENT '使用时间',
  `expire_at` DATETIME NULL DEFAULT NULL COMMENT '过期时间',
  `created_at` DATETIME NULL DEFAULT NULL COMMENT '创建时间',
  `updated_at` DATETIME NULL DEFAULT NULL COMMENT '更新时间',
  PRIMARY KEY (`id`),
  KEY `idx_user_id` (`user_id`),
  KEY `idx_template_id` (`template_id`),
  KEY `idx_status` (`status`),
  KEY `idx_created_at` (`created_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='用户优惠券表';

-- 邀请奖励配置表
DROP TABLE IF EXISTS `invite_reward_configs`;
CREATE TABLE `invite_reward_configs` (
  `id` BIGINT NOT NULL AUTO_INCREMENT COMMENT '主键ID',
  `reward_type` VARCHAR(32) NOT NULL DEFAULT '' COMMENT '奖励类型',
  `amount` BIGINT NOT NULL DEFAULT 0 COMMENT '奖励金额(分)',
  `conditions` JSON COMMENT '奖励条件',
  `created_at` DATETIME NULL DEFAULT NULL COMMENT '创建时间',
  `updated_at` DATETIME NULL DEFAULT NULL COMMENT '更新时间',
  PRIMARY KEY (`id`),
  KEY `idx_reward_type` (`reward_type`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='邀请奖励配置表';

-- 邀请奖励日志表
DROP TABLE IF EXISTS `invite_reward_logs`;
CREATE TABLE `invite_reward_logs` (
  `id` BIGINT NOT NULL AUTO_INCREMENT COMMENT '主键ID',
  `inviter_id` BIGINT NOT NULL DEFAULT 0 COMMENT '邀请人ID',
  `invitee_id` BIGINT NOT NULL DEFAULT 0 COMMENT '被邀请人ID',
  `reward_type` VARCHAR(32) NOT NULL DEFAULT '' COMMENT '奖励类型',
  `reward_amount` BIGINT NOT NULL DEFAULT 0 COMMENT '奖励金额(分)',
  `created_at` DATETIME NULL DEFAULT NULL COMMENT '创建时间',
  PRIMARY KEY (`id`),
  KEY `idx_inviter_id` (`inviter_id`),
  KEY `idx_invitee_id` (`invitee_id`),
  KEY `idx_created_at` (`created_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='邀请奖励日志表';

-- 充值活动表
DROP TABLE IF EXISTS `recharge_activities`;
CREATE TABLE `recharge_activities` (
  `id` BIGINT NOT NULL AUTO_INCREMENT COMMENT '主键ID',
  `name` VARCHAR(64) NOT NULL DEFAULT '' COMMENT '活动名称',
  `threshold_amount` BIGINT NOT NULL DEFAULT 0 COMMENT '充值门槛(分)',
  `bonus_amount` BIGINT NOT NULL DEFAULT 0 COMMENT '赠送金额(分)',
  `status` TINYINT NOT NULL DEFAULT 1 COMMENT '状态 1启用 0停用',
  `start_at` DATETIME NULL DEFAULT NULL COMMENT '开始时间',
  `end_at` DATETIME NULL DEFAULT NULL COMMENT '结束时间',
  `created_at` DATETIME NULL DEFAULT NULL COMMENT '创建时间',
  `updated_at` DATETIME NULL DEFAULT NULL COMMENT '更新时间',
  PRIMARY KEY (`id`),
  KEY `idx_status` (`status`),
  KEY `idx_created_at` (`created_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='充值活动表';

-- 用户充值记录表
DROP TABLE IF EXISTS `user_recharge_logs`;
CREATE TABLE `user_recharge_logs` (
  `id` BIGINT NOT NULL AUTO_INCREMENT COMMENT '主键ID',
  `user_id` BIGINT NOT NULL DEFAULT 0 COMMENT '用户ID',
  `amount` BIGINT NOT NULL DEFAULT 0 COMMENT '充值金额(分)',
  `bonus` BIGINT NOT NULL DEFAULT 0 COMMENT '赠送金额(分)',
  `activity_id` BIGINT NOT NULL DEFAULT 0 COMMENT '活动ID',
  `payment_id` BIGINT NOT NULL DEFAULT 0 COMMENT '支付记录ID',
  `created_at` DATETIME NULL DEFAULT NULL COMMENT '创建时间',
  PRIMARY KEY (`id`),
  KEY `idx_user_id` (`user_id`),
  KEY `idx_activity_id` (`activity_id`),
  KEY `idx_payment_id` (`payment_id`),
  KEY `idx_created_at` (`created_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='用户充值记录表';

-- 抽奖活动表
DROP TABLE IF EXISTS `lottery_activities`;
CREATE TABLE `lottery_activities` (
  `id` BIGINT NOT NULL AUTO_INCREMENT COMMENT '主键ID',
  `name` VARCHAR(64) NOT NULL DEFAULT '' COMMENT '活动名称',
  `status` TINYINT NOT NULL DEFAULT 1 COMMENT '状态 1启用 0停用',
  `start_at` DATETIME NULL DEFAULT NULL COMMENT '开始时间',
  `end_at` DATETIME NULL DEFAULT NULL COMMENT '结束时间',
  `created_at` DATETIME NULL DEFAULT NULL COMMENT '创建时间',
  `updated_at` DATETIME NULL DEFAULT NULL COMMENT '更新时间',
  PRIMARY KEY (`id`),
  KEY `idx_status` (`status`),
  KEY `idx_created_at` (`created_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='抽奖活动表';

-- 抽奖奖品表
DROP TABLE IF EXISTS `lottery_prizes`;
CREATE TABLE `lottery_prizes` (
  `id` BIGINT NOT NULL AUTO_INCREMENT COMMENT '主键ID',
  `activity_id` BIGINT NOT NULL DEFAULT 0 COMMENT '活动ID',
  `name` VARCHAR(64) NOT NULL DEFAULT '' COMMENT '奖品名称',
  `type` VARCHAR(32) NOT NULL DEFAULT '' COMMENT '奖品类型 coupon/balance/gift',
  `value` BIGINT NOT NULL DEFAULT 0 COMMENT '奖品价值(分)',
  `probability` DECIMAL(5,2) NOT NULL DEFAULT 0.00 COMMENT '中奖概率%',
  `stock` INT NOT NULL DEFAULT 0 COMMENT '库存',
  `claimed` INT NOT NULL DEFAULT 0 COMMENT '已领取数',
  `created_at` DATETIME NULL DEFAULT NULL COMMENT '创建时间',
  `updated_at` DATETIME NULL DEFAULT NULL COMMENT '更新时间',
  PRIMARY KEY (`id`),
  KEY `idx_activity_id` (`activity_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='抽奖奖品表';

-- 抽奖记录表
DROP TABLE IF EXISTS `lottery_records`;
CREATE TABLE `lottery_records` (
  `id` BIGINT NOT NULL AUTO_INCREMENT COMMENT '主键ID',
  `activity_id` BIGINT NOT NULL DEFAULT 0 COMMENT '活动ID',
  `user_id` BIGINT NOT NULL DEFAULT 0 COMMENT '用户ID',
  `prize_id` BIGINT NOT NULL DEFAULT 0 COMMENT '奖品ID',
  `created_at` DATETIME NULL DEFAULT NULL COMMENT '创建时间',
  PRIMARY KEY (`id`),
  KEY `idx_activity_id` (`activity_id`),
  KEY `idx_user_id` (`user_id`),
  KEY `idx_created_at` (`created_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='抽奖记录表';

-- 拼团活动表
DROP TABLE IF EXISTS `group_buy_activities`;
CREATE TABLE `group_buy_activities` (
  `id` BIGINT NOT NULL AUTO_INCREMENT COMMENT '主键ID',
  `name` VARCHAR(64) NOT NULL DEFAULT '' COMMENT '活动名称',
  `min_members` INT NOT NULL DEFAULT 0 COMMENT '成团最低人数',
  `max_members` INT NOT NULL DEFAULT 0 COMMENT '最大人数',
  `discount_ratio` DECIMAL(3,2) NOT NULL DEFAULT 0.00 COMMENT '折扣比例',
  `status` TINYINT NOT NULL DEFAULT 1 COMMENT '状态 1启用 0停用',
  `created_at` DATETIME NULL DEFAULT NULL COMMENT '创建时间',
  `updated_at` DATETIME NULL DEFAULT NULL COMMENT '更新时间',
  PRIMARY KEY (`id`),
  KEY `idx_status` (`status`),
  KEY `idx_created_at` (`created_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='拼团活动表';

-- 拼团成员表
DROP TABLE IF EXISTS `group_buy_members`;
CREATE TABLE `group_buy_members` (
  `id` BIGINT NOT NULL AUTO_INCREMENT COMMENT '主键ID',
  `activity_id` BIGINT NOT NULL DEFAULT 0 COMMENT '活动ID',
  `group_id` BIGINT NOT NULL DEFAULT 0 COMMENT '拼团ID',
  `user_id` BIGINT NOT NULL DEFAULT 0 COMMENT '用户ID',
  `joined_at` DATETIME NULL DEFAULT NULL COMMENT '参团时间',
  `created_at` DATETIME NULL DEFAULT NULL COMMENT '创建时间',
  PRIMARY KEY (`id`),
  KEY `idx_activity_id` (`activity_id`),
  KEY `idx_group_id` (`group_id`),
  KEY `idx_user_id` (`user_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='拼团成员表';

-- 拼团订单表
DROP TABLE IF EXISTS `group_buy_orders`;
CREATE TABLE `group_buy_orders` (
  `id` BIGINT NOT NULL AUTO_INCREMENT COMMENT '主键ID',
  `activity_id` BIGINT NOT NULL DEFAULT 0 COMMENT '活动ID',
  `group_id` BIGINT NOT NULL DEFAULT 0 COMMENT '拼团ID',
  `order_id` BIGINT NOT NULL DEFAULT 0 COMMENT '订单ID',
  `status` TINYINT NOT NULL DEFAULT 0 COMMENT '状态 0拼团中 1成功 2失败',
  `created_at` DATETIME NULL DEFAULT NULL COMMENT '创建时间',
  `updated_at` DATETIME NULL DEFAULT NULL COMMENT '更新时间',
  PRIMARY KEY (`id`),
  KEY `idx_activity_id` (`activity_id`),
  KEY `idx_group_id` (`group_id`),
  KEY `idx_order_id` (`order_id`),
  KEY `idx_status` (`status`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='拼团订单表';

-- ============================================================
-- 模块24: 纠纷仲裁
-- ============================================================

-- 仲裁案件表
DROP TABLE IF EXISTS `arbitration_cases`;
CREATE TABLE `arbitration_cases` (
  `id` BIGINT NOT NULL AUTO_INCREMENT COMMENT '主键ID',
  `order_id` BIGINT NOT NULL DEFAULT 0 COMMENT '订单ID',
  `session_id` BIGINT NOT NULL DEFAULT 0 COMMENT '会话ID',
  `status` VARCHAR(32) NOT NULL DEFAULT 'pending' COMMENT '状态 pending/arbitrating/closed',
  `arbitrator_id` BIGINT NOT NULL DEFAULT 0 COMMENT '仲裁员ID',
  `result` VARCHAR(255) NOT NULL DEFAULT '' COMMENT '仲裁结果',
  `created_at` DATETIME NULL DEFAULT NULL COMMENT '创建时间',
  `updated_at` DATETIME NULL DEFAULT NULL COMMENT '更新时间',
  PRIMARY KEY (`id`),
  KEY `idx_order_id` (`order_id`),
  KEY `idx_session_id` (`session_id`),
  KEY `idx_status` (`status`),
  KEY `idx_created_at` (`created_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='仲裁案件表';

-- 仲裁证据表
DROP TABLE IF EXISTS `arbitration_evidences`;
CREATE TABLE `arbitration_evidences` (
  `id` BIGINT NOT NULL AUTO_INCREMENT COMMENT '主键ID',
  `case_id` BIGINT NOT NULL DEFAULT 0 COMMENT '案件ID',
  `uploader_id` BIGINT NOT NULL DEFAULT 0 COMMENT '上传人ID',
  `type` VARCHAR(32) NOT NULL DEFAULT '' COMMENT '证据类型 video/screenshot/chat',
  `file_url` VARCHAR(512) NOT NULL DEFAULT '' COMMENT '文件URL',
  `description` VARCHAR(255) NOT NULL DEFAULT '' COMMENT '描述',
  `created_at` DATETIME NULL DEFAULT NULL COMMENT '创建时间',
  PRIMARY KEY (`id`),
  KEY `idx_case_id` (`case_id`),
  KEY `idx_uploader_id` (`uploader_id`),
  KEY `idx_created_at` (`created_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='仲裁证据表';

-- 举证模板表
DROP TABLE IF EXISTS `arbitration_evidence_tpls`;
CREATE TABLE `arbitration_evidence_tpls` (
  `id` BIGINT NOT NULL AUTO_INCREMENT COMMENT '主键ID',
  `name` VARCHAR(64) NOT NULL DEFAULT '' COMMENT '模板名称',
  `fields` JSON COMMENT '字段定义',
  `created_at` DATETIME NULL DEFAULT NULL COMMENT '创建时间',
  `updated_at` DATETIME NULL DEFAULT NULL COMMENT '更新时间',
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='举证模板表';

-- 判责规则表
DROP TABLE IF EXISTS `arbitration_rules`;
CREATE TABLE `arbitration_rules` (
  `id` BIGINT NOT NULL AUTO_INCREMENT COMMENT '主键ID',
  `name` VARCHAR(64) NOT NULL DEFAULT '' COMMENT '规则名称',
  `condition` JSON COMMENT '触发条件',
  `responsibility` VARCHAR(32) NOT NULL DEFAULT '' COMMENT '责任方 player/user/club/both',
  `penalty` VARCHAR(255) NOT NULL DEFAULT '' COMMENT '处罚措施',
  `created_at` DATETIME NULL DEFAULT NULL COMMENT '创建时间',
  `updated_at` DATETIME NULL DEFAULT NULL COMMENT '更新时间',
  PRIMARY KEY (`id`),
  KEY `idx_responsibility` (`responsibility`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='判责规则表';

-- ============================================================
-- 模块25: 合规
-- ============================================================

-- 代练拦截规则表
DROP TABLE IF EXISTS `anti_boosting_rules`;
CREATE TABLE `anti_boosting_rules` (
  `id` BIGINT NOT NULL AUTO_INCREMENT COMMENT '主键ID',
  `pattern` VARCHAR(255) NOT NULL DEFAULT '' COMMENT '匹配规则',
  `action` VARCHAR(32) NOT NULL DEFAULT 'block' COMMENT '执行动作 block/warn',
  `enabled` TINYINT NOT NULL DEFAULT 1 COMMENT '是否启用 0否 1是',
  `created_at` DATETIME NULL DEFAULT NULL COMMENT '创建时间',
  `updated_at` DATETIME NULL DEFAULT NULL COMMENT '更新时间',
  PRIMARY KEY (`id`),
  KEY `idx_enabled` (`enabled`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='代练拦截规则表';

-- 代练拦截日志表
DROP TABLE IF EXISTS `anti_boosting_logs`;
CREATE TABLE `anti_boosting_logs` (
  `id` BIGINT NOT NULL AUTO_INCREMENT COMMENT '主键ID',
  `user_id` BIGINT NOT NULL DEFAULT 0 COMMENT '用户ID',
  `matched_pattern` VARCHAR(255) NOT NULL DEFAULT '' COMMENT '命中规则',
  `content` TEXT COMMENT '触发内容',
  `created_at` DATETIME NULL DEFAULT NULL COMMENT '创建时间',
  PRIMARY KEY (`id`),
  KEY `idx_user_id` (`user_id`),
  KEY `idx_created_at` (`created_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='代练拦截日志表';

-- 协议版本表
DROP TABLE IF EXISTS `agreement_role_versions`;
CREATE TABLE `agreement_role_versions` (
  `id` BIGINT NOT NULL AUTO_INCREMENT COMMENT '主键ID',
  `role` VARCHAR(32) NOT NULL DEFAULT '' COMMENT '角色 player/distributor/club',
  `version` VARCHAR(32) NOT NULL DEFAULT '' COMMENT '版本号',
  `file_url` VARCHAR(512) NOT NULL DEFAULT '' COMMENT '协议文件URL',
  `effective_at` DATETIME NULL DEFAULT NULL COMMENT '生效时间',
  `created_at` DATETIME NULL DEFAULT NULL COMMENT '创建时间',
  `updated_at` DATETIME NULL DEFAULT NULL COMMENT '更新时间',
  PRIMARY KEY (`id`),
  KEY `idx_role` (`role`),
  KEY `idx_version` (`version`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='协议版本表';

-- 协议签署日志表
DROP TABLE IF EXISTS `agreement_sign_logs`;
CREATE TABLE `agreement_sign_logs` (
  `id` BIGINT NOT NULL AUTO_INCREMENT COMMENT '主键ID',
  `user_id` BIGINT NOT NULL DEFAULT 0 COMMENT '用户ID',
  `agreement_id` BIGINT NOT NULL DEFAULT 0 COMMENT '协议ID',
  `role` VARCHAR(32) NOT NULL DEFAULT '' COMMENT '角色 player/distributor/club',
  `signed_at` DATETIME NULL DEFAULT NULL COMMENT '签署时间',
  `ip` VARCHAR(64) NOT NULL DEFAULT '' COMMENT 'IP地址',
  `created_at` DATETIME NULL DEFAULT NULL COMMENT '创建时间',
  PRIMARY KEY (`id`),
  KEY `idx_user_id` (`user_id`),
  KEY `idx_agreement_id` (`agreement_id`),
  KEY `idx_created_at` (`created_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='协议签署日志表';

-- ============================================================
-- 模块26: 其他
-- ============================================================

-- 游戏列表表
DROP TABLE IF EXISTS `game_lists`;
CREATE TABLE `game_lists` (
  `id` BIGINT NOT NULL AUTO_INCREMENT COMMENT '主键ID',
  `name` VARCHAR(64) NOT NULL DEFAULT '' COMMENT '游戏名称',
  `icon` VARCHAR(512) NOT NULL DEFAULT '' COMMENT '游戏图标URL',
  `sort_order` INT NOT NULL DEFAULT 0 COMMENT '排序',
  `status` TINYINT NOT NULL DEFAULT 1 COMMENT '状态 1启用 0停用',
  `created_at` DATETIME NULL DEFAULT NULL COMMENT '创建时间',
  `updated_at` DATETIME NULL DEFAULT NULL COMMENT '更新时间',
  PRIMARY KEY (`id`),
  KEY `idx_status` (`status`),
  KEY `idx_sort_order` (`sort_order`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='游戏列表表';

-- 数据快照表
DROP TABLE IF EXISTS `data_dashboard_snapshots`;
CREATE TABLE `data_dashboard_snapshots` (
  `id` BIGINT NOT NULL AUTO_INCREMENT COMMENT '主键ID',
  `snapshot_date` DATE NOT NULL COMMENT '快照日期',
  `metrics` JSON COMMENT '指标数据',
  `created_at` DATETIME NULL DEFAULT NULL COMMENT '创建时间',
  PRIMARY KEY (`id`),
  KEY `idx_snapshot_date` (`snapshot_date`),
  KEY `idx_created_at` (`created_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='数据快照表';

-- 平台文档表
DROP TABLE IF EXISTS `documents`;
CREATE TABLE `documents` (
  `id` BIGINT NOT NULL AUTO_INCREMENT COMMENT '主键ID',
  `name` VARCHAR(128) NOT NULL DEFAULT '' COMMENT '文档名称',
  `type` VARCHAR(32) NOT NULL DEFAULT '' COMMENT '类型 protocol/policy/contract',
  `file_url` VARCHAR(512) NOT NULL DEFAULT '' COMMENT '文件URL',
  `version` VARCHAR(32) NOT NULL DEFAULT '' COMMENT '版本号',
  `is_deleted` TINYINT NOT NULL DEFAULT 0 COMMENT '是否删除 0否 1是',
  `created_at` DATETIME NULL DEFAULT NULL COMMENT '创建时间',
  `updated_at` DATETIME NULL DEFAULT NULL COMMENT '更新时间',
  PRIMARY KEY (`id`),
  KEY `idx_type` (`type`),
  KEY `idx_is_deleted` (`is_deleted`),
  KEY `idx_created_at` (`created_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='平台文档表';

-- 文档版本表
DROP TABLE IF EXISTS `document_versions`;
CREATE TABLE `document_versions` (
  `id` BIGINT NOT NULL AUTO_INCREMENT COMMENT '主键ID',
  `document_id` BIGINT NOT NULL DEFAULT 0 COMMENT '文档ID',
  `file_url` VARCHAR(512) NOT NULL DEFAULT '' COMMENT '文件URL',
  `version` VARCHAR(32) NOT NULL DEFAULT '' COMMENT '版本号',
  `created_by` BIGINT NOT NULL DEFAULT 0 COMMENT '创建人ID',
  `created_at` DATETIME NULL DEFAULT NULL COMMENT '创建时间',
  PRIMARY KEY (`id`),
  KEY `idx_document_id` (`document_id`),
  KEY `idx_created_at` (`created_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='文档版本表';

-- 派单记录表
DROP TABLE IF EXISTS `dispatch_records`;
CREATE TABLE `dispatch_records` (
  `id` BIGINT NOT NULL AUTO_INCREMENT COMMENT '主键ID',
  `dispatcher_id` BIGINT NOT NULL DEFAULT 0 COMMENT '派单员ID',
  `order_id` BIGINT NOT NULL DEFAULT 0 COMMENT '订单ID',
  `player_id` BIGINT NOT NULL DEFAULT 0 COMMENT '指派打手ID',
  `status` TINYINT NOT NULL DEFAULT 0 COMMENT '状态 0待接单 1已接单 2已拒绝',
  `created_at` DATETIME NULL DEFAULT NULL COMMENT '创建时间',
  `updated_at` DATETIME NULL DEFAULT NULL COMMENT '更新时间',
  PRIMARY KEY (`id`),
  KEY `idx_dispatcher_id` (`dispatcher_id`),
  KEY `idx_order_id` (`order_id`),
  KEY `idx_player_id` (`player_id`),
  KEY `idx_status` (`status`),
  KEY `idx_created_at` (`created_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='派单记录表';

-- ============================================================
-- 初始化数据
-- ============================================================

-- 初始超级管理员账号（密码: 1234567，bcrypt加密）
INSERT INTO admins (username, password, nickname, role, is_init, status, created_at, updated_at)
VALUES ('admin', '$2a$10$N9qo8uLOickgx2ZMRZoMy.MrqK3a8W5kPQnFQ5QJYJQZqM3j5YK8e', '超级管理员', 1, 0, 1, NOW(), NOW()),
       ('admin2', '$2a$10$N9qo8uLOickgx2ZMRZoMy.MrqK3a8W5kPQnFQ5QJYJQZqM3j5YK8e', '超级管理员2', 1, 0, 1, NOW(), NOW());

-- 默认系统配置
INSERT INTO system_configs (`key`, `value`, `description`, `updated_at`) VALUES
('platform_fee_rate', '20', '平台抽成比例(%)', NOW()),
('withdraw_freeze_days', '3', '提现冻结天数', NOW()),
('order_timeout_minutes', '10', '订单超时时间(分钟)', NOW()),
('minor_single_limit', '5000', '未成年人单笔消费限制(分)', NOW()),
('minor_monthly_limit', '20000', '未成年人月消费限制(分)', NOW()),
('realname_required', '1', '实名认证开关', NOW()),
('ai_scan_enabled', '1', 'AI风控扫描开关', NOW()),
('ios_payment_enabled', '1', 'iOS支付开关', NOW()),
('esign_provider', 'fdd', '电子签服务商(fdd=法天天/tx=腾讯电子签)', NOW()),
('customer_wechat', '', '客服微信号', NOW()),
('ws_heartbeat_seconds', '25', 'WebSocket心跳(秒)', NOW()),
('ws_timeout_seconds', '70', 'WebSocket超时(秒)', NOW()),
('log_retention_days', '180', '日志保留天数', NOW()),
('curfew_start', '22', '宵禁开始时间(小时)', NOW()),
('curfew_end', '8', '宵禁结束时间(小时)', NOW()),
('face_verify_daily_limit', '5', '活体检测每日限制', NOW()),
('face_verify_cache_days', '7', '活体检测缓存天数', NOW()),
('guardian_monthly_exempt_limit', '3', '监护人每月豁免次数', NOW()),
('after_sale_keyword_intervention', '1', '售后关键词自动介入开关', NOW());

-- ============================================================
-- 俱乐部入驻流程相关迁移(板块一/二/三)
-- ============================================================

-- 1) clubs 表新增字段：入驻驳回次数 / 入驻锁定截止 / 创始人抽成 / 注销归档标记
ALTER TABLE `clubs`
  ADD COLUMN `reject_count` INT NOT NULL DEFAULT 0 COMMENT '入驻驳回次数' AFTER `total_orders`,
  ADD COLUMN `locked_until` DATETIME NULL DEFAULT NULL COMMENT '入驻锁定截止时间(驳回3次后锁定7天)' AFTER `reject_count`,
  ADD COLUMN `commission_rate` TINYINT NOT NULL DEFAULT 0 COMMENT '创始人抽成比例(0-100)' AFTER `locked_until`,
  ADD COLUMN `is_archived` TINYINT NOT NULL DEFAULT 0 COMMENT '是否已注销归档 0否 1是' AFTER `commission_rate`,
  ADD KEY `idx_is_archived` (`is_archived`);

-- 2) 入驻相关系统配置
INSERT INTO system_configs (`key`, `value`, `description`, `updated_at`) VALUES
('club_join_switch', '1', '俱乐部入驻开关 0=关闭 1=开启', NOW()),
('club_personal_deposit', '50000', '个人俱乐部保证金阈值(分)', NOW()),
('club_enterprise_deposit', '500000', '企业俱乐部保证金阈值(分)', NOW()),
('club_join_reject_lock_count', '3', '入驻驳回锁定阈值(次)', NOW()),
('club_join_lock_days', '7', '入驻锁定天数(驳回阈值后)', NOW()),
('legal_person_face_valid_hours', '72', '法人活体认证有效期(小时)', NOW()),
('corporate_transfer_expire_hours', '48', '对公打款验证有效期(小时)', NOW()),
('corporate_transfer_max_fail_count', '5', '对公打款验证最大失败次数', NOW()),
('corporate_transfer_lock_days', '15', '对公打款失败锁定天数', NOW()),
('club_draft_expire_days', '7', '俱乐部入驻草稿有效期(天)', NOW());

-- 3) 俱乐部保证金扣款记录表(板块一)
DROP TABLE IF EXISTS `club_deposit_deductions`;
CREATE TABLE `club_deposit_deductions` (
  `id` BIGINT NOT NULL AUTO_INCREMENT COMMENT '主键ID',
  `club_id` BIGINT NOT NULL DEFAULT 0 COMMENT '俱乐部ID',
  `amount` BIGINT NOT NULL DEFAULT 0 COMMENT '扣除金额(分)',
  `type` VARCHAR(32) NOT NULL DEFAULT 'fine' COMMENT '类型 fine=罚款 compensation=赔偿',
  `reason` VARCHAR(255) NOT NULL DEFAULT '' COMMENT '扣款原因',
  `operator_id` BIGINT NOT NULL DEFAULT 0 COMMENT '操作人ID',
  `created_at` DATETIME NULL DEFAULT NULL COMMENT '创建时间',
  PRIMARY KEY (`id`),
  KEY `idx_club_id` (`club_id`),
  KEY `idx_type` (`type`),
  KEY `idx_created_at` (`created_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='俱乐部保证金扣款记录表';

-- 4) 俱乐部入驻草稿表(板块二,7天有效期)
DROP TABLE IF EXISTS `club_join_drafts`;
CREATE TABLE `club_join_drafts` (
  `id` BIGINT NOT NULL AUTO_INCREMENT COMMENT '主键ID',
  `user_id` BIGINT NOT NULL DEFAULT 0 COMMENT '用户ID',
  `draft_data` JSON COMMENT '草稿数据(JSON)',
  `expire_at` DATETIME NULL DEFAULT NULL COMMENT '过期时间',
  `created_at` DATETIME NULL DEFAULT NULL COMMENT '创建时间',
  `updated_at` DATETIME NULL DEFAULT NULL COMMENT '更新时间',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_user_id` (`user_id`),
  KEY `idx_user_id` (`user_id`),
  KEY `idx_expire_at` (`expire_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='俱乐部入驻草稿表';

-- 5) 法人活体认证表(板块三,72小时有效期)
DROP TABLE IF EXISTS `legal_person_face_verifies`;
CREATE TABLE `legal_person_face_verifies` (
  `id` BIGINT NOT NULL AUTO_INCREMENT COMMENT '主键ID',
  `club_id` BIGINT NOT NULL DEFAULT 0 COMMENT '俱乐部ID',
  `legal_person_name` VARCHAR(64) NOT NULL DEFAULT '' COMMENT '法人姓名',
  `legal_person_id_card` VARCHAR(18) NOT NULL DEFAULT '' COMMENT '法人身份证号',
  `verify_token` VARCHAR(255) NOT NULL DEFAULT '' COMMENT '活体认证 token',
  `verify_at` DATETIME NULL DEFAULT NULL COMMENT '验证时间',
  `expire_at` DATETIME NULL DEFAULT NULL COMMENT '过期时间(verify_at + 72h)',
  `status` VARCHAR(32) NOT NULL DEFAULT 'pending' COMMENT '状态 pending/passed/failed/expired',
  `created_at` DATETIME NULL DEFAULT NULL COMMENT '创建时间',
  `updated_at` DATETIME NULL DEFAULT NULL COMMENT '更新时间',
  PRIMARY KEY (`id`),
  KEY `idx_club_id` (`club_id`),
  KEY `idx_expire_at` (`expire_at`),
  KEY `idx_status` (`status`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='法人活体认证表';

-- 6) 对公小额打款验证表(板块三,48小时有效期,5次失败锁定15天)
DROP TABLE IF EXISTS `corporate_transfer_verifies`;
CREATE TABLE `corporate_transfer_verifies` (
  `id` BIGINT NOT NULL AUTO_INCREMENT COMMENT '主键ID',
  `club_id` BIGINT NOT NULL DEFAULT 0 COMMENT '俱乐部ID',
  `bank_name` VARCHAR(64) NOT NULL DEFAULT '' COMMENT '开户行',
  `bank_account` VARCHAR(64) NOT NULL DEFAULT '' COMMENT '银行账号',
  `account_name` VARCHAR(64) NOT NULL DEFAULT '' COMMENT '账户名',
  `verify_amount` DECIMAL(4,1) NOT NULL DEFAULT 0.0 COMMENT '验证金额(0.0-0.9，1位小数)',
  `generated_at` DATETIME NULL DEFAULT NULL COMMENT '生成打款时间',
  `expire_at` DATETIME NULL DEFAULT NULL COMMENT '过期时间(generated_at + 48h)',
  `verify_count` INT NOT NULL DEFAULT 0 COMMENT '已用验证次数',
  `status` VARCHAR(32) NOT NULL DEFAULT 'pending' COMMENT '状态 pending/verified/failed/expired',
  `locked_until` DATETIME NULL DEFAULT NULL COMMENT '锁定时间(失败次数达上限后 15 天内禁止提交)',
  `created_at` DATETIME NULL DEFAULT NULL COMMENT '创建时间',
  `updated_at` DATETIME NULL DEFAULT NULL COMMENT '更新时间',
  PRIMARY KEY (`id`),
  KEY `idx_club_id` (`club_id`),
  KEY `idx_expire_at` (`expire_at`),
  KEY `idx_status` (`status`),
  KEY `idx_locked_until` (`locked_until`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='对公小额打款验证表';

-- ============================================================
-- 俱乐部入驻后管理 / IM联动 / 安全注销相关迁移(板块四/五/六/七)
-- ============================================================

-- 7) operation_logs 表新增字段：操作结果 + 业务模块
ALTER TABLE `operation_logs`
  ADD COLUMN `result` VARCHAR(16) NOT NULL DEFAULT 'success' COMMENT '操作结果 success/fail' AFTER `action`,
  ADD COLUMN `module` VARCHAR(32) NOT NULL DEFAULT '' COMMENT '业务模块 club_join/deposit/vbadge 等' AFTER `result`,
  ADD KEY `idx_result` (`result`),
  ADD KEY `idx_module` (`module`);

-- 8) 俱乐部资料修改日志表(板块四,资料变更审计溯源)
DROP TABLE IF EXISTS `club_info_change_logs`;
CREATE TABLE `club_info_change_logs` (
  `id` BIGINT NOT NULL AUTO_INCREMENT COMMENT '主键ID',
  `club_id` BIGINT NOT NULL DEFAULT 0 COMMENT '俱乐部ID',
  `field` VARCHAR(64) NOT NULL DEFAULT '' COMMENT '修改字段名',
  `old_value` TEXT COMMENT '旧值',
  `new_value` TEXT COMMENT '新值',
  `operator_id` BIGINT NOT NULL DEFAULT 0 COMMENT '操作人ID',
  `created_at` DATETIME NULL DEFAULT NULL COMMENT '创建时间',
  PRIMARY KEY (`id`),
  KEY `idx_club_id` (`club_id`),
  KEY `idx_created_at` (`created_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='俱乐部资料修改日志表';

-- 9) 俱乐部内部罚款规则表(板块四)
DROP TABLE IF EXISTS `club_fine_rules`;
CREATE TABLE `club_fine_rules` (
  `id` BIGINT NOT NULL AUTO_INCREMENT COMMENT '主键ID',
  `club_id` BIGINT NOT NULL DEFAULT 0 COMMENT '俱乐部ID',
  `name` VARCHAR(64) NOT NULL DEFAULT '' COMMENT '规则名称',
  `description` TEXT COMMENT '规则描述',
  `amount` BIGINT NOT NULL DEFAULT 0 COMMENT '罚款金额(分)',
  `status` VARCHAR(32) NOT NULL DEFAULT 'active' COMMENT '状态 active/revoked',
  `has_unpaid` TINYINT NOT NULL DEFAULT 0 COMMENT '是否存在未赔付罚款 0否 1是',
  `created_by` BIGINT NOT NULL DEFAULT 0 COMMENT '创建人ID',
  `created_at` DATETIME NULL DEFAULT NULL COMMENT '创建时间',
  `updated_at` DATETIME NULL DEFAULT NULL COMMENT '更新时间',
  PRIMARY KEY (`id`),
  KEY `idx_club_id` (`club_id`),
  KEY `idx_status` (`status`),
  KEY `idx_created_at` (`created_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='俱乐部内部罚款规则表';

-- 10) 罚款规则平台备案审核表(板块四)
DROP TABLE IF EXISTS `club_fine_rule_reviews`;
CREATE TABLE `club_fine_rule_reviews` (
  `id` BIGINT NOT NULL AUTO_INCREMENT COMMENT '主键ID',
  `rule_id` BIGINT NOT NULL DEFAULT 0 COMMENT '罚款规则ID',
  `club_id` BIGINT NOT NULL DEFAULT 0 COMMENT '俱乐部ID',
  `review_status` VARCHAR(32) NOT NULL DEFAULT 'pending' COMMENT '审核状态 pending/approved/revoked',
  `reviewer_id` BIGINT NOT NULL DEFAULT 0 COMMENT '审核人ID',
  `review_note` VARCHAR(255) NOT NULL DEFAULT '' COMMENT '审核备注',
  `reviewed_at` DATETIME NULL DEFAULT NULL COMMENT '审核时间',
  `created_at` DATETIME NULL DEFAULT NULL COMMENT '创建时间',
  PRIMARY KEY (`id`),
  KEY `idx_rule_id` (`rule_id`),
  KEY `idx_club_id` (`club_id`),
  KEY `idx_review_status` (`review_status`),
  KEY `idx_created_at` (`created_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='罚款规则平台备案审核表';

-- 11) 群公告已读日志表(板块四)
DROP TABLE IF EXISTS `announcement_read_logs`;
CREATE TABLE `announcement_read_logs` (
  `id` BIGINT NOT NULL AUTO_INCREMENT COMMENT '主键ID',
  `announcement_id` BIGINT NOT NULL DEFAULT 0 COMMENT '公告ID(对应 group_chats.id)',
  `user_id` BIGINT NOT NULL DEFAULT 0 COMMENT '用户ID',
  `read_at` DATETIME NULL DEFAULT NULL COMMENT '阅读时间',
  `created_at` DATETIME NULL DEFAULT NULL COMMENT '创建时间',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_announcement_user` (`announcement_id`, `user_id`),
  KEY `idx_announcement_id` (`announcement_id`),
  KEY `idx_user_id` (`user_id`),
  KEY `idx_created_at` (`created_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='群公告已读日志表';

-- 12) 俱乐部注销资料归档表(板块七,加密+上链存证)
DROP TABLE IF EXISTS `club_archives`;
CREATE TABLE `club_archives` (
  `id` BIGINT NOT NULL AUTO_INCREMENT COMMENT '主键ID',
  `club_id` BIGINT NOT NULL DEFAULT 0 COMMENT '俱乐部ID',
  `archive_data` JSON COMMENT '归档资料 JSON(加密后密文)',
  `encrypted` TINYINT NOT NULL DEFAULT 0 COMMENT '是否已加密 0否 1是',
  `file_hash` VARCHAR(64) NOT NULL DEFAULT '' COMMENT '加密后哈希(SHA-256)',
  `blockchain_tx_id` VARCHAR(128) NOT NULL DEFAULT '' COMMENT '上链交易ID',
  `archived_at` DATETIME NULL DEFAULT NULL COMMENT '归档时间',
  `created_at` DATETIME NULL DEFAULT NULL COMMENT '创建时间',
  PRIMARY KEY (`id`),
  KEY `idx_club_id` (`club_id`),
  KEY `idx_file_hash` (`file_hash`),
  KEY `idx_created_at` (`created_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='俱乐部注销资料归档表';

-- 13) 文件上链存证记录表(板块五,水印/加密/上链)
DROP TABLE IF EXISTS `file_blockchain_records`;
CREATE TABLE `file_blockchain_records` (
  `id` BIGINT NOT NULL AUTO_INCREMENT COMMENT '主键ID',
  `file_hash` VARCHAR(64) NOT NULL DEFAULT '' COMMENT '文件 SHA-256 哈希',
  `file_type` VARCHAR(32) NOT NULL DEFAULT '' COMMENT '文件类型',
  `ref_type` VARCHAR(32) NOT NULL DEFAULT '' COMMENT '关联类型 club_join/club_archive',
  `ref_id` BIGINT NOT NULL DEFAULT 0 COMMENT '关联ID',
  `oss_url` VARCHAR(512) NOT NULL DEFAULT '' COMMENT 'OSS 存储地址',
  `watermark_text` VARCHAR(128) NOT NULL DEFAULT '' COMMENT '水印文本',
  `blockchain_tx_id` VARCHAR(128) NOT NULL DEFAULT '' COMMENT '上链交易ID',
  `created_at` DATETIME NULL DEFAULT NULL COMMENT '创建时间',
  PRIMARY KEY (`id`),
  KEY `idx_file_hash` (`file_hash`),
  KEY `idx_ref_type` (`ref_type`),
  KEY `idx_ref_id` (`ref_id`),
  KEY `idx_blockchain_tx_id` (`blockchain_tx_id`),
  KEY `idx_created_at` (`created_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='文件上链存证记录表';

-- 14) 企业对公小额打款记录台账表(板块五)
DROP TABLE IF EXISTS `corporate_transfer_records`;
CREATE TABLE `corporate_transfer_records` (
  `id` BIGINT NOT NULL AUTO_INCREMENT COMMENT '主键ID',
  `club_id` BIGINT NOT NULL DEFAULT 0 COMMENT '俱乐部ID',
  `verify_id` BIGINT NOT NULL DEFAULT 0 COMMENT '验证流程ID',
  `amount` BIGINT NOT NULL DEFAULT 0 COMMENT '打款金额(分)',
  `direction` VARCHAR(16) NOT NULL DEFAULT 'out' COMMENT '方向 out/refund',
  `bank_name` VARCHAR(64) NOT NULL DEFAULT '' COMMENT '开户行',
  `bank_account` VARCHAR(64) NOT NULL DEFAULT '' COMMENT '银行账号',
  `transfer_at` DATETIME NULL DEFAULT NULL COMMENT '打款时间',
  `status` VARCHAR(32) NOT NULL DEFAULT 'pending' COMMENT '状态 pending/success/failed',
  `created_at` DATETIME NULL DEFAULT NULL COMMENT '创建时间',
  PRIMARY KEY (`id`),
  KEY `idx_club_id` (`club_id`),
  KEY `idx_verify_id` (`verify_id`),
  KEY `idx_status` (`status`),
  KEY `idx_created_at` (`created_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='企业对公小额打款记录台账表';

SET FOREIGN_KEY_CHECKS = 1;
