-- ============================================================
-- 戟三电竞 (JiSan Esports) - 游戏陪玩平台完整数据库设计
-- Target: MySQL 8.0+
-- Backend: Go (Gin + GORM)
-- 字符集: utf8mb4 / 排序规则: utf8mb4_unicode_ci
-- 金额单位: BIGINT 存储「分」
-- GORM 约定: id BIGINT UNSIGNED AUTO_INCREMENT; created_at/updated_at/deleted_at TIMESTAMP NULL DEFAULT NULL
-- 引擎: InnoDB
-- ============================================================

CREATE DATABASE IF NOT EXISTS `jisan_esports` DEFAULT CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;
USE `jisan_esports`;

SET NAMES utf8mb4;
SET FOREIGN_KEY_CHECKS = 0;


-- ============================================================
-- 模块一: 用户与认证 (User & Auth)
-- ============================================================

-- 1. 用户表
DROP TABLE IF EXISTS `users`;
CREATE TABLE `users` (
  `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  `openid` VARCHAR(128) NOT NULL DEFAULT '' COMMENT '微信openid',
  `unionid` VARCHAR(128) NOT NULL DEFAULT '' COMMENT '微信unionid',
  `phone` VARCHAR(20) NOT NULL DEFAULT '' COMMENT '手机号',
  `nickname` VARCHAR(64) NOT NULL DEFAULT '' COMMENT '昵称',
  `avatar` VARCHAR(512) NOT NULL DEFAULT '' COMMENT '头像URL',
  `gender` TINYINT NOT NULL DEFAULT 0 COMMENT '性别 0未知 1男 2女',
  `role` TINYINT NOT NULL DEFAULT 1 COMMENT '角色 1客户 2打手 3分销商 4派单员',
  `invite_code` VARCHAR(32) NOT NULL DEFAULT '' COMMENT '我的邀请码',
  `club_id` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '所属俱乐部ID',
  `is_realname` TINYINT NOT NULL DEFAULT 0 COMMENT '是否已实名 0否 1是',
  `is_minor` TINYINT NOT NULL DEFAULT 0 COMMENT '是否未成年 0否 1是',
  `real_name` VARCHAR(64) NOT NULL DEFAULT '' COMMENT '真实姓名',
  `id_card` VARCHAR(255) NOT NULL DEFAULT '' COMMENT '身份证号(加密存储)',
  `credit_score` INT NOT NULL DEFAULT 100 COMMENT '信用分 初始100',
  `player_level` TINYINT NOT NULL DEFAULT 0 COMMENT '打手段位 1青铜 2白银 3黄金 4钻石',
  `status` TINYINT NOT NULL DEFAULT 1 COMMENT '状态 1正常 2禁用',
  `balance` BIGINT NOT NULL DEFAULT 0 COMMENT '余额(分)',
  `frozen_balance` BIGINT NOT NULL DEFAULT 0 COMMENT '冻结金额(分)',
  `points` INT NOT NULL DEFAULT 0 COMMENT '积分',
  `last_login_at` TIMESTAMP NULL DEFAULT NULL COMMENT '最后登录时间',
  `last_login_ip` VARCHAR(64) NOT NULL DEFAULT '' COMMENT '最后登录IP',
  `is_phone_abandoned` TINYINT NOT NULL DEFAULT 0 COMMENT '手机号是否被二次放号回收 0否 1是',
  `created_at` TIMESTAMP NULL DEFAULT NULL,
  `updated_at` TIMESTAMP NULL DEFAULT NULL,
  `deleted_at` TIMESTAMP NULL DEFAULT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_openid` (`openid`),
  KEY `idx_phone` (`phone`),
  KEY `idx_unionid` (`unionid`),
  KEY `idx_role` (`role`),
  KEY `idx_status` (`status`),
  KEY `idx_club_id` (`club_id`),
  KEY `idx_invite_code` (`invite_code`),
  KEY `idx_is_realname` (`is_realname`),
  KEY `idx_created_at` (`created_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='用户表(玩家/打手/分销商/派单员)';

-- 2. 实名认证记录
DROP TABLE IF EXISTS `realname_verify_logs`;
CREATE TABLE `realname_verify_logs` (
  `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  `user_id` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '用户ID',
  `real_name` VARCHAR(64) NOT NULL DEFAULT '' COMMENT '姓名',
  `id_card` VARCHAR(255) NOT NULL DEFAULT '' COMMENT '身份证号(加密)',
  `face_token` VARCHAR(255) NOT NULL DEFAULT '' COMMENT '人脸特征token',
  `verify_result` TINYINT NOT NULL DEFAULT 0 COMMENT '认证结果 0待处理 1通过 2失败',
  `age` INT NOT NULL DEFAULT 0 COMMENT '年龄',
  `is_minor` TINYINT NOT NULL DEFAULT 0 COMMENT '是否未成年 0否 1是',
  `error_msg` VARCHAR(255) NOT NULL DEFAULT '' COMMENT '错误信息',
  `ip` VARCHAR(64) NOT NULL DEFAULT '' COMMENT '请求IP',
  `created_at` TIMESTAMP NULL DEFAULT NULL,
  PRIMARY KEY (`id`),
  KEY `idx_user_id` (`user_id`),
  KEY `idx_verify_result` (`verify_result`),
  KEY `idx_created_at` (`created_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='实名认证记录表';

-- 3. 活体校验缓存
DROP TABLE IF EXISTS `realname_cache`;
CREATE TABLE `realname_cache` (
  `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  `user_id` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '用户ID',
  `cache_key` VARCHAR(128) NOT NULL DEFAULT '' COMMENT '缓存键',
  `expire_at` TIMESTAMP NULL DEFAULT NULL COMMENT '过期时间',
  `created_at` TIMESTAMP NULL DEFAULT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_cache_key` (`cache_key`),
  KEY `idx_user_id` (`user_id`),
  KEY `idx_expire_at` (`expire_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='活体校验缓存表';

-- 4. 活体频控
DROP TABLE IF EXISTS `face_verify_rate_limits`;
CREATE TABLE `face_verify_rate_limits` (
  `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  `user_id` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '用户ID',
  `ip` VARCHAR(64) NOT NULL DEFAULT '' COMMENT 'IP地址',
  `daily_count` INT NOT NULL DEFAULT 0 COMMENT '当日次数',
  `date` DATE NOT NULL COMMENT '日期',
  `created_at` TIMESTAMP NULL DEFAULT NULL,
  `updated_at` TIMESTAMP NULL DEFAULT NULL,
  PRIMARY KEY (`id`),
  KEY `idx_user_date` (`user_id`, `date`),
  KEY `idx_ip` (`ip`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='活体检测频控表';

-- 5. 手机号二次放号申诉
DROP TABLE IF EXISTS `phone_appeals`;
CREATE TABLE `phone_appeals` (
  `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  `user_id` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '用户ID',
  `phone` VARCHAR(20) NOT NULL DEFAULT '' COMMENT '申诉手机号',
  `old_openid` VARCHAR(128) NOT NULL DEFAULT '' COMMENT '原openid',
  `new_openid` VARCHAR(128) NOT NULL DEFAULT '' COMMENT '新openid',
  `video_url` VARCHAR(512) NOT NULL DEFAULT '' COMMENT '申诉视频URL',
  `network_time` VARCHAR(64) NOT NULL DEFAULT '' COMMENT '入网时间',
  `first_use_date` DATE NULL DEFAULT NULL COMMENT '首次使用日期',
  `commitment_letter_url` VARCHAR(512) NOT NULL DEFAULT '' COMMENT '承诺书URL',
  `status` TINYINT NOT NULL DEFAULT 1 COMMENT '状态 1待审核 2通过 3驳回',
  `reject_reason` VARCHAR(255) NOT NULL DEFAULT '' COMMENT '驳回原因',
  `admin_id` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '审核管理员ID',
  `created_at` TIMESTAMP NULL DEFAULT NULL,
  `updated_at` TIMESTAMP NULL DEFAULT NULL,
  PRIMARY KEY (`id`),
  KEY `idx_user_id` (`user_id`),
  KEY `idx_phone` (`phone`),
  KEY `idx_status` (`status`),
  KEY `idx_created_at` (`created_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='手机号二次放号申诉表';

-- 6. 监护人验证
DROP TABLE IF EXISTS `guardian_verifies`;
CREATE TABLE `guardian_verifies` (
  `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  `user_id` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '用户ID',
  `order_id` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '订单ID',
  `guardian_name` VARCHAR(64) NOT NULL DEFAULT '' COMMENT '监护人姓名',
  `guardian_id_card` VARCHAR(255) NOT NULL DEFAULT '' COMMENT '监护人身份证(加密)',
  `face_token` VARCHAR(255) NOT NULL DEFAULT '' COMMENT '人脸token',
  `face_session_id` VARCHAR(128) NOT NULL DEFAULT '' COMMENT '活体会话ID',
  `expire_at` TIMESTAMP NULL DEFAULT NULL COMMENT '过期时间',
  `created_at` TIMESTAMP NULL DEFAULT NULL,
  PRIMARY KEY (`id`),
  KEY `idx_user_id` (`user_id`),
  KEY `idx_order_id` (`order_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='监护人验证表';

-- 7. 电子签名
DROP TABLE IF EXISTS `electronic_signatures`;
CREATE TABLE `electronic_signatures` (
  `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  `user_id` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '用户ID',
  `order_id` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '订单ID',
  `guardian_verify_id` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '监护人验证ID',
  `sign_url` VARCHAR(512) NOT NULL DEFAULT '' COMMENT '签名文件URL',
  `face_session_id` VARCHAR(128) NOT NULL DEFAULT '' COMMENT '活体会话ID',
  `amount_input` BIGINT NOT NULL DEFAULT 0 COMMENT '金额确认输入(分)',
  `provider` VARCHAR(32) NOT NULL DEFAULT '' COMMENT '服务商 法大大/腾讯电子签',
  `sign_status` TINYINT NOT NULL DEFAULT 0 COMMENT '签署状态 0待签 1已签 2失败',
  `ext_data` TEXT COMMENT '扩展数据(JSON)',
  `created_at` TIMESTAMP NULL DEFAULT NULL,
  PRIMARY KEY (`id`),
  KEY `idx_user_id` (`user_id`),
  KEY `idx_order_id` (`order_id`),
  KEY `idx_sign_status` (`sign_status`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='电子签名表';

-- 8. 家长监护绑定
DROP TABLE IF EXISTS `parent_guardian_binds`;
CREATE TABLE `parent_guardian_binds` (
  `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  `minor_user_id` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '未成年用户ID',
  `guardian_user_id` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '监护人用户ID',
  `guardian_name` VARCHAR(64) NOT NULL DEFAULT '' COMMENT '监护人姓名',
  `guardian_id_card` VARCHAR(255) NOT NULL DEFAULT '' COMMENT '监护人身份证(加密)',
  `relation` VARCHAR(32) NOT NULL DEFAULT '' COMMENT '关系 父子/母子等',
  `verify_status` TINYINT NOT NULL DEFAULT 0 COMMENT '验证状态 0待验证 1已验证 2已拒绝',
  `created_at` TIMESTAMP NULL DEFAULT NULL,
  PRIMARY KEY (`id`),
  KEY `idx_minor_user_id` (`minor_user_id`),
  KEY `idx_guardian_user_id` (`guardian_user_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='家长监护绑定表';

-- 9. 家长监护设置
DROP TABLE IF EXISTS `parent_guardian_settings`;
CREATE TABLE `parent_guardian_settings` (
  `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  `bind_id` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '绑定关系ID',
  `monthly_limit` BIGINT NOT NULL DEFAULT 0 COMMENT '月度消费限额(分)',
  `allow_order` TINYINT NOT NULL DEFAULT 1 COMMENT '允许下单 0否 1是',
  `allow_reward` TINYINT NOT NULL DEFAULT 1 COMMENT '允许打赏 0否 1是',
  `is_frozen` TINYINT NOT NULL DEFAULT 0 COMMENT '是否冻结账户 0否 1是',
  `created_at` TIMESTAMP NULL DEFAULT NULL,
  `updated_at` TIMESTAMP NULL DEFAULT NULL,
  PRIMARY KEY (`id`),
  KEY `idx_bind_id` (`bind_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='家长监护设置表';

-- 10. 家长消费报告
DROP TABLE IF EXISTS `parent_consume_reports`;
CREATE TABLE `parent_consume_reports` (
  `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  `minor_user_id` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '未成年用户ID',
  `month` VARCHAR(7) NOT NULL DEFAULT '' COMMENT '月份 YYYY-MM',
  `total_consume` BIGINT NOT NULL DEFAULT 0 COMMENT '消费总额(分)',
  `order_count` INT NOT NULL DEFAULT 0 COMMENT '订单数',
  `created_at` TIMESTAMP NULL DEFAULT NULL,
  PRIMARY KEY (`id`),
  KEY `idx_minor_user_id` (`minor_user_id`),
  KEY `idx_month` (`month`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='家长消费报告表';


-- ============================================================
-- 模块二: 管理员 (Admin)
-- ============================================================

-- 11. 管理员表
DROP TABLE IF EXISTS `admins`;
CREATE TABLE `admins` (
  `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  `username` VARCHAR(64) NOT NULL DEFAULT '' COMMENT '管理员账号',
  `nickname` VARCHAR(64) NOT NULL DEFAULT '' COMMENT '昵称',
  `password` VARCHAR(255) NOT NULL DEFAULT '' COMMENT '密码(bcrypt)',
  `email` VARCHAR(128) NOT NULL DEFAULT '' COMMENT '绑定邮箱',
  `phone` VARCHAR(20) NOT NULL DEFAULT '' COMMENT '手机号',
  `role` TINYINT NOT NULL DEFAULT 2 COMMENT '角色 1超级管理员 2运营 3财务 4风控',
  `status` TINYINT NOT NULL DEFAULT 1 COMMENT '状态 1正常 2禁用',
  `is_init` TINYINT NOT NULL DEFAULT 0 COMMENT '是否已初始化(改密+绑邮箱) 0否 1是',
  `last_login_at` TIMESTAMP NULL DEFAULT NULL COMMENT '最后登录时间',
  `last_login_ip` VARCHAR(64) NOT NULL DEFAULT '' COMMENT '最后登录IP',
  `created_at` TIMESTAMP NULL DEFAULT NULL,
  `updated_at` TIMESTAMP NULL DEFAULT NULL,
  `deleted_at` TIMESTAMP NULL DEFAULT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_username` (`username`),
  KEY `idx_email` (`email`),
  KEY `idx_phone` (`phone`),
  KEY `idx_role` (`role`),
  KEY `idx_status` (`status`),
  KEY `idx_created_at` (`created_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='管理员表';

-- 12. 密码历史
DROP TABLE IF EXISTS `admin_password_histories`;
CREATE TABLE `admin_password_histories` (
  `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  `admin_id` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '管理员ID',
  `password_hash` VARCHAR(255) NOT NULL DEFAULT '' COMMENT '历史密码hash',
  `created_at` TIMESTAMP NULL DEFAULT NULL,
  PRIMARY KEY (`id`),
  KEY `idx_admin_id` (`admin_id`),
  KEY `idx_created_at` (`created_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='管理员密码历史表(保留最近5次)';

-- 13. 通行密钥(WebAuthn)
DROP TABLE IF EXISTS `admin_webauthn_credentials`;
CREATE TABLE `admin_webauthn_credentials` (
  `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  `admin_id` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '管理员ID',
  `credential_id` VARCHAR(255) NOT NULL DEFAULT '' COMMENT '凭证ID',
  `public_key` TEXT COMMENT '公钥',
  `device_info` VARCHAR(255) NOT NULL DEFAULT '' COMMENT '设备信息',
  `created_at` TIMESTAMP NULL DEFAULT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_credential_id` (`credential_id`),
  KEY `idx_admin_id` (`admin_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='管理员WebAuthn通行密钥表';

-- 14. 操作日志
DROP TABLE IF EXISTS `admin_operation_logs`;
CREATE TABLE `admin_operation_logs` (
  `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  `admin_id` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '管理员ID',
  `admin_name` VARCHAR(64) NOT NULL DEFAULT '' COMMENT '管理员名称',
  `action` VARCHAR(64) NOT NULL DEFAULT '' COMMENT '操作动作',
  `target_type` VARCHAR(64) NOT NULL DEFAULT '' COMMENT '操作对象类型',
  `target_id` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '操作对象ID',
  `content` TEXT COMMENT '操作内容',
  `ip` VARCHAR(64) NOT NULL DEFAULT '' COMMENT 'IP地址',
  `user_agent` VARCHAR(255) NOT NULL DEFAULT '' COMMENT 'User-Agent',
  `created_at` TIMESTAMP NULL DEFAULT NULL,
  PRIMARY KEY (`id`),
  KEY `idx_admin_id` (`admin_id`),
  KEY `idx_action` (`action`),
  KEY `idx_created_at` (`created_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='管理员操作日志表';

-- 15. 初始化日志
DROP TABLE IF EXISTS `init_logs`;
CREATE TABLE `init_logs` (
  `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  `admin_id` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '管理员ID',
  `init_type` TINYINT NOT NULL DEFAULT 0 COMMENT '初始化类型 1改密 2绑邮箱',
  `status` TINYINT NOT NULL DEFAULT 0 COMMENT '状态 0进行中 1完成 2失败',
  `created_at` TIMESTAMP NULL DEFAULT NULL,
  PRIMARY KEY (`id`),
  KEY `idx_admin_id` (`admin_id`),
  KEY `idx_created_at` (`created_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='管理员初始化日志表';

-- 16. 邮箱验证码
DROP TABLE IF EXISTS `email_verify_codes`;
CREATE TABLE `email_verify_codes` (
  `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  `admin_id` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '管理员ID',
  `email` VARCHAR(128) NOT NULL DEFAULT '' COMMENT '邮箱',
  `code` VARCHAR(16) NOT NULL DEFAULT '' COMMENT '验证码',
  `purpose` VARCHAR(32) NOT NULL DEFAULT '' COMMENT '用途 bind/change',
  `expire_at` TIMESTAMP NULL DEFAULT NULL COMMENT '过期时间',
  `used` TINYINT NOT NULL DEFAULT 0 COMMENT '是否已使用 0否 1是',
  `created_at` TIMESTAMP NULL DEFAULT NULL,
  PRIMARY KEY (`id`),
  KEY `idx_admin_id` (`admin_id`),
  KEY `idx_email` (`email`),
  KEY `idx_created_at` (`created_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='邮箱验证码表';

-- 17. 导出日志
DROP TABLE IF EXISTS `export_logs`;
CREATE TABLE `export_logs` (
  `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  `admin_id` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '管理员ID',
  `export_type` VARCHAR(64) NOT NULL DEFAULT '' COMMENT '导出类型',
  `file_url` VARCHAR(512) NOT NULL DEFAULT '' COMMENT '文件URL',
  `watermark` VARCHAR(128) NOT NULL DEFAULT '' COMMENT '水印信息',
  `record_count` INT NOT NULL DEFAULT 0 COMMENT '记录数',
  `created_at` TIMESTAMP NULL DEFAULT NULL,
  PRIMARY KEY (`id`),
  KEY `idx_admin_id` (`admin_id`),
  KEY `idx_created_at` (`created_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='数据导出日志表';


-- ============================================================
-- 模块三: 俱乐部 (Club)
-- ============================================================

-- 18. 俱乐部表
DROP TABLE IF EXISTS `clubs`;
CREATE TABLE `clubs` (
  `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  `name` VARCHAR(128) NOT NULL DEFAULT '' COMMENT '俱乐部名称',
  `abbreviation` VARCHAR(32) NOT NULL DEFAULT '' COMMENT '缩写(唯一封存)',
  `type` TINYINT NOT NULL DEFAULT 2 COMMENT '类型 1企业 2个人',
  `status` TINYINT NOT NULL DEFAULT 1 COMMENT '状态 1审核中 2审核通过 3驳回 4冻结 5停业 6注销',
  `founder_user_id` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '创始人用户ID',
  `logo` VARCHAR(512) NOT NULL DEFAULT '' COMMENT 'Logo URL',
  `background` VARCHAR(512) NOT NULL DEFAULT '' COMMENT '背景图URL',
  `intro` VARCHAR(500) NOT NULL DEFAULT '' COMMENT '简介',
  `contact_phone` VARCHAR(20) NOT NULL DEFAULT '' COMMENT '联系电话',
  `contact_wechat` VARCHAR(64) NOT NULL DEFAULT '' COMMENT '联系微信',
  `address` VARCHAR(255) NOT NULL DEFAULT '' COMMENT '地址',
  `business_hours` VARCHAR(64) NOT NULL DEFAULT '' COMMENT '营业时间',
  `rating` DECIMAL(2,1) NOT NULL DEFAULT 0.0 COMMENT '评分',
  `monthly_revenue` BIGINT NOT NULL DEFAULT 0 COMMENT '月营收(分)',
  `deposit_amount` BIGINT NOT NULL DEFAULT 0 COMMENT '保证金金额(分)',
  `deposit_status` TINYINT NOT NULL DEFAULT 0 COMMENT '保证金状态 0未缴 1已缴 2已退',
  `v_badge_type` TINYINT NOT NULL DEFAULT 0 COMMENT 'V标 0无 1蓝V 2绿V',
  `created_at` TIMESTAMP NULL DEFAULT NULL,
  `updated_at` TIMESTAMP NULL DEFAULT NULL,
  `deleted_at` TIMESTAMP NULL DEFAULT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_abbreviation` (`abbreviation`),
  KEY `idx_status` (`status`),
  KEY `idx_founder_user_id` (`founder_user_id`),
  KEY `idx_created_at` (`created_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='俱乐部表';

-- 19. 俱乐部成员
DROP TABLE IF EXISTS `club_members`;
CREATE TABLE `club_members` (
  `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  `club_id` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '俱乐部ID',
  `user_id` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '用户ID',
  `role` TINYINT NOT NULL DEFAULT 3 COMMENT '角色 1创始人 2管理员 3打手',
  `join_time` TIMESTAMP NULL DEFAULT NULL COMMENT '加入时间',
  `status` TINYINT NOT NULL DEFAULT 1 COMMENT '状态 1正常 2已移除',
  `created_at` TIMESTAMP NULL DEFAULT NULL,
  PRIMARY KEY (`id`),
  KEY `idx_club_id` (`club_id`),
  KEY `idx_user_id` (`user_id`),
  KEY `idx_status` (`status`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='俱乐部成员表';

-- 20. 俱乐部缩写封存
DROP TABLE IF EXISTS `club_abbreviations`;
CREATE TABLE `club_abbreviations` (
  `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  `abbreviation` VARCHAR(32) NOT NULL DEFAULT '' COMMENT '缩写',
  `club_id` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '俱乐部ID',
  `club_name` VARCHAR(128) NOT NULL DEFAULT '' COMMENT '俱乐部名称',
  `created_at` TIMESTAMP NULL DEFAULT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_abbreviation` (`abbreviation`),
  KEY `idx_club_id` (`club_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='俱乐部缩写封存表';

-- 21. 保证金阶梯
DROP TABLE IF EXISTS `club_deposit_tiers`;
CREATE TABLE `club_deposit_tiers` (
  `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  `club_id` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '俱乐部ID',
  `min_revenue` BIGINT NOT NULL DEFAULT 0 COMMENT '最低月营收(分)',
  `deposit_amount` BIGINT NOT NULL DEFAULT 0 COMMENT '保证金金额(分)',
  `violation_count` INT NOT NULL DEFAULT 0 COMMENT '违规次数',
  `created_at` TIMESTAMP NULL DEFAULT NULL,
  PRIMARY KEY (`id`),
  KEY `idx_club_id` (`club_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='俱乐部保证金阶梯表';

-- 22. 跨服分店
DROP TABLE IF EXISTS `club_branches`;
CREATE TABLE `club_branches` (
  `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  `club_id` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '俱乐部ID',
  `branch_name` VARCHAR(64) NOT NULL DEFAULT '' COMMENT '分店名称',
  `game_type` VARCHAR(64) NOT NULL DEFAULT '' COMMENT '游戏类型',
  `revenue_ratio` DECIMAL(5,2) NOT NULL DEFAULT 0.00 COMMENT '营收占比%',
  `created_at` TIMESTAMP NULL DEFAULT NULL,
  PRIMARY KEY (`id`),
  KEY `idx_club_id` (`club_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='俱乐部跨服分店表';

-- 23. 俱乐部动态墙
DROP TABLE IF EXISTS `club_dynamics`;
CREATE TABLE `club_dynamics` (
  `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  `club_id` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '俱乐部ID',
  `user_id` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '发布用户ID',
  `type` TINYINT NOT NULL DEFAULT 1 COMMENT '类型 1战绩 2视频',
  `content` VARCHAR(500) NOT NULL DEFAULT '' COMMENT '内容',
  `media_url` VARCHAR(512) NOT NULL DEFAULT '' COMMENT '媒体URL',
  `like_count` INT NOT NULL DEFAULT 0 COMMENT '点赞数',
  `created_at` TIMESTAMP NULL DEFAULT NULL,
  PRIMARY KEY (`id`),
  KEY `idx_club_id` (`club_id`),
  KEY `idx_user_id` (`user_id`),
  KEY `idx_created_at` (`created_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='俱乐部动态墙表';

-- 24. 内部抢单
DROP TABLE IF EXISTS `club_internal_orders`;
CREATE TABLE `club_internal_orders` (
  `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  `club_id` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '俱乐部ID',
  `title` VARCHAR(128) NOT NULL DEFAULT '' COMMENT '标题',
  `desc` VARCHAR(500) NOT NULL DEFAULT '' COMMENT '描述',
  `amount` BIGINT NOT NULL DEFAULT 0 COMMENT '金额(分)',
  `status` TINYINT NOT NULL DEFAULT 0 COMMENT '状态 0待接 1已接 2已完成',
  `assigned_user_id` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '接单用户ID',
  `created_at` TIMESTAMP NULL DEFAULT NULL,
  PRIMARY KEY (`id`),
  KEY `idx_club_id` (`club_id`),
  KEY `idx_status` (`status`),
  KEY `idx_assigned_user_id` (`assigned_user_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='俱乐部内部抢单表';

-- 25. 俱乐部优惠券
DROP TABLE IF EXISTS `club_coupons`;
CREATE TABLE `club_coupons` (
  `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  `club_id` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '俱乐部ID',
  `name` VARCHAR(64) NOT NULL DEFAULT '' COMMENT '券名称',
  `type` TINYINT NOT NULL DEFAULT 1 COMMENT '类型 1满减 2新人',
  `amount` BIGINT NOT NULL DEFAULT 0 COMMENT '优惠金额(分)',
  `min_spend` BIGINT NOT NULL DEFAULT 0 COMMENT '最低消费(分)',
  `total_count` INT NOT NULL DEFAULT 0 COMMENT '发放总量',
  `used_count` INT NOT NULL DEFAULT 0 COMMENT '已使用数量',
  `expire_at` TIMESTAMP NULL DEFAULT NULL COMMENT '过期时间',
  `created_at` TIMESTAMP NULL DEFAULT NULL,
  PRIMARY KEY (`id`),
  KEY `idx_club_id` (`club_id`),
  KEY `idx_type` (`type`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='俱乐部优惠券表';

-- 26. 俱乐部公告
DROP TABLE IF EXISTS `club_announcements`;
CREATE TABLE `club_announcements` (
  `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  `club_id` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '俱乐部ID',
  `title` VARCHAR(128) NOT NULL DEFAULT '' COMMENT '标题',
  `content` TEXT COMMENT '公告内容',
  `created_by` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '创建人ID',
  `created_at` TIMESTAMP NULL DEFAULT NULL,
  PRIMARY KEY (`id`),
  KEY `idx_club_id` (`club_id`),
  KEY `idx_created_at` (`created_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='俱乐部公告表';

-- 27. 俱乐部日报
DROP TABLE IF EXISTS `club_daily_stats`;
CREATE TABLE `club_daily_stats` (
  `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  `club_id` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '俱乐部ID',
  `stat_date` DATE NOT NULL COMMENT '统计日期',
  `order_count` INT NOT NULL DEFAULT 0 COMMENT '订单数',
  `revenue` BIGINT NOT NULL DEFAULT 0 COMMENT '营收(分)',
  `new_members` INT NOT NULL DEFAULT 0 COMMENT '新增成员',
  `disputes` INT NOT NULL DEFAULT 0 COMMENT '纠纷数',
  `created_at` TIMESTAMP NULL DEFAULT NULL,
  PRIMARY KEY (`id`),
  KEY `idx_club_id` (`club_id`),
  KEY `idx_stat_date` (`stat_date`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='俱乐部日报表';


-- ============================================================
-- 模块四: 邀请 (Invite)
-- ============================================================

-- 28. 邀请码
DROP TABLE IF EXISTS `invite_codes`;
CREATE TABLE `invite_codes` (
  `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  `code` VARCHAR(32) NOT NULL DEFAULT '' COMMENT '邀请码',
  `type` TINYINT NOT NULL DEFAULT 2 COMMENT '类型 1指定俱乐部 2全平台通用',
  `club_id` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '指定俱乐部ID',
  `role_type` TINYINT NOT NULL DEFAULT 0 COMMENT '角色类型 0不限 1打手DS 2分销商FXS',
  `max_uses` INT NOT NULL DEFAULT 0 COMMENT '最大使用次数 0不限',
  `used_count` INT NOT NULL DEFAULT 0 COMMENT '已使用次数',
  `expire_at` TIMESTAMP NULL DEFAULT NULL COMMENT '过期时间',
  `status` TINYINT NOT NULL DEFAULT 1 COMMENT '状态 1未使用 2已使用 3已作废 4已用完 5已过期',
  `benefits_config` TEXT COMMENT '权益配置(JSON)',
  `created_by` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '创建人ID',
  `created_at` TIMESTAMP NULL DEFAULT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_code` (`code`),
  KEY `idx_club_id` (`club_id`),
  KEY `idx_status` (`status`),
  KEY `idx_created_at` (`created_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='邀请码表';

-- 29. 邀请码使用记录
DROP TABLE IF EXISTS `invite_code_usage_logs`;
CREATE TABLE `invite_code_usage_logs` (
  `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  `invite_code_id` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '邀请码ID',
  `user_id` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '使用用户ID',
  `benefits_detail` TEXT COMMENT '权益明细(JSON)',
  `created_at` TIMESTAMP NULL DEFAULT NULL,
  PRIMARY KEY (`id`),
  KEY `idx_invite_code_id` (`invite_code_id`),
  KEY `idx_user_id` (`user_id`),
  KEY `idx_created_at` (`created_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='邀请码使用记录表';

-- 30. 邀请奖励配置
DROP TABLE IF EXISTS `invite_reward_configs`;
CREATE TABLE `invite_reward_configs` (
  `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  `reward_type` VARCHAR(32) NOT NULL DEFAULT '' COMMENT '奖励类型',
  `reward_amount` BIGINT NOT NULL DEFAULT 0 COMMENT '奖励金额(分)',
  `reward_coupon_id` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '奖励优惠券ID',
  `description` VARCHAR(255) NOT NULL DEFAULT '' COMMENT '描述',
  `created_at` TIMESTAMP NULL DEFAULT NULL,
  `updated_at` TIMESTAMP NULL DEFAULT NULL,
  PRIMARY KEY (`id`),
  KEY `idx_reward_type` (`reward_type`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='邀请奖励配置表';

-- 31. 邀请奖励发放
DROP TABLE IF EXISTS `invite_reward_logs`;
CREATE TABLE `invite_reward_logs` (
  `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  `inviter_user_id` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '邀请人ID',
  `invitee_user_id` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '被邀请人ID',
  `order_id` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '关联订单ID',
  `reward_type` VARCHAR(32) NOT NULL DEFAULT '' COMMENT '奖励类型',
  `reward_amount` BIGINT NOT NULL DEFAULT 0 COMMENT '奖励金额(分)',
  `created_at` TIMESTAMP NULL DEFAULT NULL,
  PRIMARY KEY (`id`),
  KEY `idx_inviter_user_id` (`inviter_user_id`),
  KEY `idx_invitee_user_id` (`invitee_user_id`),
  KEY `idx_order_id` (`order_id`),
  KEY `idx_created_at` (`created_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='邀请奖励发放记录表';

-- 32. 用户权益
DROP TABLE IF EXISTS `user_benefits`;
CREATE TABLE `user_benefits` (
  `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  `user_id` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '用户ID',
  `benefit_type` TINYINT NOT NULL DEFAULT 0 COMMENT '权益类型 1余额 2积分 3免费下单 4服务时长 5满减券 6折扣券 7打手体验卡 8流量曝光 9优先推荐 10技能标签 11分销权益卡 12下级扩容 13身份标识 14头像框 15抽奖次数 16实名免验 17专属客服',
  `benefit_value` TEXT COMMENT '权益值(JSON)',
  `expire_at` TIMESTAMP NULL DEFAULT NULL COMMENT '过期时间',
  `status` TINYINT NOT NULL DEFAULT 1 COMMENT '状态 1有效 2已过期 3已使用',
  `created_at` TIMESTAMP NULL DEFAULT NULL,
  PRIMARY KEY (`id`),
  KEY `idx_user_id` (`user_id`),
  KEY `idx_benefit_type` (`benefit_type`),
  KEY `idx_status` (`status`),
  KEY `idx_created_at` (`created_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='用户权益表';


-- ============================================================
-- 模块五: 订单 (Order)
-- ============================================================

-- 33. 订单表
DROP TABLE IF EXISTS `orders`;
CREATE TABLE `orders` (
  `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  `order_no` VARCHAR(64) NOT NULL DEFAULT '' COMMENT '订单号',
  `user_id` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '客户ID',
  `player_id` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '打手ID',
  `club_id` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '俱乐部ID',
  `service_id` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '服务项目ID',
  `order_type` TINYINT NOT NULL DEFAULT 1 COMMENT '订单类型 1即时单 2预约单 3车队单 4教学单',
  `status` TINYINT NOT NULL DEFAULT 1 COMMENT '状态 1待接单 2已接单 3进行中 4待验收 5已完成 6待结算 7已结算 8已取消 9大额验证失败 10售后中',
  `amount` BIGINT NOT NULL DEFAULT 0 COMMENT '订单金额(分)',
  `paid_amount` BIGINT NOT NULL DEFAULT 0 COMMENT '已付金额(分)',
  `refunded_amount` BIGINT NOT NULL DEFAULT 0 COMMENT '已退金额(分)',
  `deposit_amount` BIGINT NOT NULL DEFAULT 0 COMMENT '保证金(分)',
  `duration_minutes` INT NOT NULL DEFAULT 0 COMMENT '服务时长(分钟)',
  `service_start_at` TIMESTAMP NULL DEFAULT NULL COMMENT '服务开始时间',
  `service_end_at` TIMESTAMP NULL DEFAULT NULL COMMENT '服务结束时间',
  `settle_at` TIMESTAMP NULL DEFAULT NULL COMMENT '结算时间',
  `expire_at` TIMESTAMP NULL DEFAULT NULL COMMENT '订单过期时间',
  `is_minor_order` TINYINT NOT NULL DEFAULT 0 COMMENT '是否未成年人订单 0否 1是',
  `guardian_verify_id` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '监护人验证ID',
  `remark` VARCHAR(500) NOT NULL DEFAULT '' COMMENT '备注',
  `created_at` TIMESTAMP NULL DEFAULT NULL,
  `updated_at` TIMESTAMP NULL DEFAULT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_order_no` (`order_no`),
  KEY `idx_user_id` (`user_id`),
  KEY `idx_player_id` (`player_id`),
  KEY `idx_club_id` (`club_id`),
  KEY `idx_status` (`status`),
  KEY `idx_order_type` (`order_type`),
  KEY `idx_created_at` (`created_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='订单表';

-- 34. 订单状态流转
DROP TABLE IF EXISTS `order_status_logs`;
CREATE TABLE `order_status_logs` (
  `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  `order_id` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '订单ID',
  `from_status` TINYINT NOT NULL DEFAULT 0 COMMENT '原状态',
  `to_status` TINYINT NOT NULL DEFAULT 0 COMMENT '新状态',
  `operator_type` TINYINT NOT NULL DEFAULT 1 COMMENT '操作人类型 1系统 2管理员 3创始人',
  `operator_id` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '操作人ID',
  `reason` VARCHAR(255) NOT NULL DEFAULT '' COMMENT '原因',
  `created_at` TIMESTAMP NULL DEFAULT NULL,
  PRIMARY KEY (`id`),
  KEY `idx_order_id` (`order_id`),
  KEY `idx_created_at` (`created_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='订单状态流转日志表';

-- 35. 履约凭证
DROP TABLE IF EXISTS `order_evidence`;
CREATE TABLE `order_evidence` (
  `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  `order_id` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '订单ID',
  `user_id` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '上传用户ID',
  `type` TINYINT NOT NULL DEFAULT 1 COMMENT '类型 1录屏 2截图 3战绩',
  `file_url` VARCHAR(512) NOT NULL DEFAULT '' COMMENT '文件URL',
  `created_at` TIMESTAMP NULL DEFAULT NULL,
  PRIMARY KEY (`id`),
  KEY `idx_order_id` (`order_id`),
  KEY `idx_user_id` (`user_id`),
  KEY `idx_created_at` (`created_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='订单履约凭证表';

-- 36. 订单套餐
DROP TABLE IF EXISTS `order_packages`;
CREATE TABLE `order_packages` (
  `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  `name` VARCHAR(64) NOT NULL DEFAULT '' COMMENT '套餐名称',
  `type` TINYINT NOT NULL DEFAULT 1 COMMENT '类型 1小时套餐 2多局套餐 3教学套餐',
  `price` BIGINT NOT NULL DEFAULT 0 COMMENT '价格(分)',
  `content` TEXT COMMENT '套餐内容',
  `status` TINYINT NOT NULL DEFAULT 1 COMMENT '状态 1上架 2下架',
  `created_at` TIMESTAMP NULL DEFAULT NULL,
  PRIMARY KEY (`id`),
  KEY `idx_type` (`type`),
  KEY `idx_status` (`status`),
  KEY `idx_created_at` (`created_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='订单套餐表';

-- 37. 竞价抢单
DROP TABLE IF EXISTS `order_bids`;
CREATE TABLE `order_bids` (
  `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  `order_id` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '订单ID',
  `player_id` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '打手ID',
  `bid_amount` BIGINT NOT NULL DEFAULT 0 COMMENT '竞价金额(分)',
  `bid_time` TIMESTAMP NULL DEFAULT NULL COMMENT '竞价时间',
  `is_winner` TINYINT NOT NULL DEFAULT 0 COMMENT '是否中标 0否 1是',
  `ip` VARCHAR(64) NOT NULL DEFAULT '' COMMENT 'IP',
  `created_at` TIMESTAMP NULL DEFAULT NULL,
  PRIMARY KEY (`id`),
  KEY `idx_order_id` (`order_id`),
  KEY `idx_player_id` (`player_id`),
  KEY `idx_created_at` (`created_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='订单竞价抢单表';

-- 38. 预约下单
DROP TABLE IF EXISTS `order_appointments`;
CREATE TABLE `order_appointments` (
  `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  `order_id` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '订单ID',
  `player_id` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '打手ID',
  `appointment_time` TIMESTAMP NULL DEFAULT NULL COMMENT '预约时间',
  `notify_sent` TINYINT NOT NULL DEFAULT 0 COMMENT '是否已通知 0否 1是',
  `created_at` TIMESTAMP NULL DEFAULT NULL,
  PRIMARY KEY (`id`),
  KEY `idx_order_id` (`order_id`),
  KEY `idx_player_id` (`player_id`),
  KEY `idx_created_at` (`created_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='预约下单表';

-- 39. 退款规则
DROP TABLE IF EXISTS `order_refund_rules`;
CREATE TABLE `order_refund_rules` (
  `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  `rule_name` VARCHAR(64) NOT NULL DEFAULT '' COMMENT '规则名称',
  `within_minutes` INT NOT NULL DEFAULT 0 COMMENT '下单后X分钟内',
  `refund_ratio` DECIMAL(5,2) NOT NULL DEFAULT 0.00 COMMENT '退款比例%',
  `is_full_refund` TINYINT NOT NULL DEFAULT 0 COMMENT '是否全额退款 0否 1是',
  `created_at` TIMESTAMP NULL DEFAULT NULL,
  `updated_at` TIMESTAMP NULL DEFAULT NULL,
  PRIMARY KEY (`id`),
  KEY `idx_created_at` (`created_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='订单退款规则表';

-- 40. 服务计时
DROP TABLE IF EXISTS `order_service_timers`;
CREATE TABLE `order_service_timers` (
  `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  `order_id` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '订单ID',
  `start_at` TIMESTAMP NULL DEFAULT NULL COMMENT '开始时间',
  `end_at` TIMESTAMP NULL DEFAULT NULL COMMENT '结束时间',
  `total_seconds` INT NOT NULL DEFAULT 0 COMMENT '总时长(秒)',
  `grace_expire_at` TIMESTAMP NULL DEFAULT NULL COMMENT '宽限期截止',
  `created_at` TIMESTAMP NULL DEFAULT NULL,
  PRIMARY KEY (`id`),
  KEY `idx_order_id` (`order_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='订单服务计时表';

-- 41. 订单类型配置
DROP TABLE IF EXISTS `order_type_configs`;
CREATE TABLE `order_type_configs` (
  `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  `type_name` VARCHAR(64) NOT NULL DEFAULT '' COMMENT '类型名称',
  `type_code` VARCHAR(32) NOT NULL DEFAULT '' COMMENT '类型编码',
  `description` VARCHAR(255) NOT NULL DEFAULT '' COMMENT '描述',
  `status` TINYINT NOT NULL DEFAULT 1 COMMENT '状态 1启用 2停用',
  `created_at` TIMESTAMP NULL DEFAULT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_type_code` (`type_code`),
  KEY `idx_status` (`status`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='订单类型配置表';

-- 42. 评价
DROP TABLE IF EXISTS `evaluations`;
CREATE TABLE `evaluations` (
  `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  `order_id` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '订单ID',
  `user_id` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '客户ID',
  `player_id` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '打手ID',
  `rating` TINYINT NOT NULL DEFAULT 5 COMMENT '评分 1-5',
  `content` VARCHAR(500) NOT NULL DEFAULT '' COMMENT '评价内容',
  `is_displayed` TINYINT NOT NULL DEFAULT 1 COMMENT '是否展示 0否 1是',
  `display_at` TIMESTAMP NULL DEFAULT NULL COMMENT '展示时间',
  `appeal_status` TINYINT NOT NULL DEFAULT 0 COMMENT '申诉状态 0无 1申诉中 2已处理',
  `created_at` TIMESTAMP NULL DEFAULT NULL,
  PRIMARY KEY (`id`),
  KEY `idx_order_id` (`order_id`),
  KEY `idx_user_id` (`user_id`),
  KEY `idx_player_id` (`player_id`),
  KEY `idx_created_at` (`created_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='订单评价表';


-- ============================================================
-- 模块六: 即时通讯 (IM)
-- ============================================================

-- 43. 聊天会话
DROP TABLE IF EXISTS `chat_sessions`;
CREATE TABLE `chat_sessions` (
  `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  `session_type` TINYINT NOT NULL DEFAULT 1 COMMENT '会话类型 1订单会话 2售后会话 3俱乐部内部群 4分类群',
  `session_type_code` VARCHAR(32) NOT NULL DEFAULT '' COMMENT '会话类型编码 after_sale/pending_intervention/risk_alert/order_supervision/group_internal/group_category/system_notice',
  `order_id` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '关联订单ID',
  `club_id` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '关联俱乐部ID',
  `group_id` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '关联群ID',
  `title` VARCHAR(128) NOT NULL DEFAULT '' COMMENT '会话标题',
  `last_message` VARCHAR(500) NOT NULL DEFAULT '' COMMENT '最后消息',
  `last_message_at` TIMESTAMP NULL DEFAULT NULL COMMENT '最后消息时间',
  `is_intervened` TINYINT NOT NULL DEFAULT 0 COMMENT '是否已介入 0否 1是',
  `intervene_reason` VARCHAR(255) NOT NULL DEFAULT '' COMMENT '介入原因',
  `created_at` TIMESTAMP NULL DEFAULT NULL,
  `updated_at` TIMESTAMP NULL DEFAULT NULL,
  PRIMARY KEY (`id`),
  KEY `idx_session_type` (`session_type`),
  KEY `idx_order_id` (`order_id`),
  KEY `idx_club_id` (`club_id`),
  KEY `idx_group_id` (`group_id`),
  KEY `idx_created_at` (`created_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='聊天会话表';

-- 44. 会话成员
DROP TABLE IF EXISTS `chat_session_members`;
CREATE TABLE `chat_session_members` (
  `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  `session_id` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '会话ID',
  `user_id` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '用户ID',
  `role` TINYINT NOT NULL DEFAULT 0 COMMENT '角色',
  `unread_count` INT NOT NULL DEFAULT 0 COMMENT '未读消息数',
  `last_read_at` TIMESTAMP NULL DEFAULT NULL COMMENT '最后已读时间',
  `created_at` TIMESTAMP NULL DEFAULT NULL,
  PRIMARY KEY (`id`),
  KEY `idx_session_id` (`session_id`),
  KEY `idx_user_id` (`user_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='会话成员表';

-- 45. 聊天消息
DROP TABLE IF EXISTS `chat_messages`;
CREATE TABLE `chat_messages` (
  `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  `session_id` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '会话ID',
  `sender_id` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '发送者ID',
  `sender_type` TINYINT NOT NULL DEFAULT 1 COMMENT '发送者类型 1用户 2打手 3俱乐部管理员 4平台官方',
  `msg_type` TINYINT NOT NULL DEFAULT 1 COMMENT '消息类型 1文字 2图片 3语音 4文件 5系统 6快捷卡片',
  `content` TEXT COMMENT '消息内容',
  `media_url` VARCHAR(512) NOT NULL DEFAULT '' COMMENT '媒体URL',
  `duration` INT NOT NULL DEFAULT 0 COMMENT '时长(秒,语音/视频)',
  `is_read` TINYINT NOT NULL DEFAULT 0 COMMENT '是否已读 0否 1是',
  `is_revoked` TINYINT NOT NULL DEFAULT 0 COMMENT '是否撤回 0否 1是',
  `revoke_at` TIMESTAMP NULL DEFAULT NULL COMMENT '撤回时间',
  `asr_text` TEXT COMMENT '语音转文字结果',
  `is_filtered` TINYINT NOT NULL DEFAULT 0 COMMENT '是否被风控过滤 0否 1是',
  `risk_level` TINYINT NOT NULL DEFAULT 0 COMMENT '风险等级 0无 1低 2中 3高',
  `created_at` TIMESTAMP NULL DEFAULT NULL,
  PRIMARY KEY (`id`),
  KEY `idx_session_id` (`session_id`),
  KEY `idx_sender_id` (`sender_id`),
  KEY `idx_created_at` (`created_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='聊天消息表';

-- 46. 消息撤回
DROP TABLE IF EXISTS `chat_message_revokes`;
CREATE TABLE `chat_message_revokes` (
  `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  `message_id` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '消息ID',
  `session_id` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '会话ID',
  `user_id` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '撤回用户ID',
  `created_at` TIMESTAMP NULL DEFAULT NULL,
  PRIMARY KEY (`id`),
  KEY `idx_message_id` (`message_id`),
  KEY `idx_session_id` (`session_id`),
  KEY `idx_user_id` (`user_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='消息撤回记录表';

-- 47. 群聊
DROP TABLE IF EXISTS `group_chats`;
CREATE TABLE `group_chats` (
  `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  `club_id` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '俱乐部ID',
  `name` VARCHAR(128) NOT NULL DEFAULT '' COMMENT '群名称',
  `type` TINYINT NOT NULL DEFAULT 1 COMMENT '类型 1内部全员群 2闲聊群 3福利群 4售后群',
  `avatar` VARCHAR(512) NOT NULL DEFAULT '' COMMENT '群头像',
  `announcement` VARCHAR(500) NOT NULL DEFAULT '' COMMENT '群公告',
  `member_count` INT NOT NULL DEFAULT 0 COMMENT '成员数',
  `status` TINYINT NOT NULL DEFAULT 1 COMMENT '状态 1正常 2解散',
  `created_by` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '创建人ID',
  `created_at` TIMESTAMP NULL DEFAULT NULL,
  PRIMARY KEY (`id`),
  KEY `idx_club_id` (`club_id`),
  KEY `idx_type` (`type`),
  KEY `idx_status` (`status`),
  KEY `idx_created_at` (`created_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='群聊表';

-- 48. 群成员
DROP TABLE IF EXISTS `group_chat_members`;
CREATE TABLE `group_chat_members` (
  `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  `group_id` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '群ID',
  `user_id` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '用户ID',
  `role` TINYINT NOT NULL DEFAULT 1 COMMENT '角色 1普通 2管理员 3创始人 4平台官方',
  `is_muted` TINYINT NOT NULL DEFAULT 0 COMMENT '是否禁言 0否 1是',
  `mute_expire_at` TIMESTAMP NULL DEFAULT NULL COMMENT '禁言截止时间',
  `joined_at` TIMESTAMP NULL DEFAULT NULL COMMENT '加入时间',
  `created_at` TIMESTAMP NULL DEFAULT NULL,
  PRIMARY KEY (`id`),
  KEY `idx_group_id` (`group_id`),
  KEY `idx_user_id` (`user_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='群成员表';

-- 49. 群消息
DROP TABLE IF EXISTS `group_chat_messages`;
CREATE TABLE `group_chat_messages` (
  `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  `group_id` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '群ID',
  `sender_id` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '发送者ID',
  `msg_type` TINYINT NOT NULL DEFAULT 1 COMMENT '消息类型',
  `content` TEXT COMMENT '内容',
  `media_url` VARCHAR(512) NOT NULL DEFAULT '' COMMENT '媒体URL',
  `duration` INT NOT NULL DEFAULT 0 COMMENT '时长(秒)',
  `is_revoked` TINYINT NOT NULL DEFAULT 0 COMMENT '是否撤回 0否 1是',
  `asr_text` TEXT COMMENT '语音转文字',
  `is_filtered` TINYINT NOT NULL DEFAULT 0 COMMENT '是否风控过滤 0否 1是',
  `created_at` TIMESTAMP NULL DEFAULT NULL,
  PRIMARY KEY (`id`),
  KEY `idx_group_id` (`group_id`),
  KEY `idx_sender_id` (`sender_id`),
  KEY `idx_created_at` (`created_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='群聊消息表';

-- 50. 定时公告
DROP TABLE IF EXISTS `group_announcement_schedules`;
CREATE TABLE `group_announcement_schedules` (
  `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  `group_id` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '群ID',
  `content` VARCHAR(500) NOT NULL DEFAULT '' COMMENT '公告内容',
  `scheduled_at` TIMESTAMP NULL DEFAULT NULL COMMENT '计划发送时间',
  `is_sent` TINYINT NOT NULL DEFAULT 0 COMMENT '是否已发送 0否 1是',
  `created_by` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '创建人ID',
  `created_at` TIMESTAMP NULL DEFAULT NULL,
  PRIMARY KEY (`id`),
  KEY `idx_group_id` (`group_id`),
  KEY `idx_scheduled_at` (`scheduled_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='群定时公告表';

-- 51. ASR缓存
DROP TABLE IF EXISTS `chat_asr_caches`;
CREATE TABLE `chat_asr_caches` (
  `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  `media_url` VARCHAR(512) NOT NULL DEFAULT '' COMMENT '媒体URL',
  `asr_text` TEXT COMMENT '转文字结果',
  `provider` VARCHAR(32) NOT NULL DEFAULT '' COMMENT 'ASR服务商',
  `created_at` TIMESTAMP NULL DEFAULT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_media_url` (`media_url`(255))
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='语音ASR缓存表';

-- 52. 飞单风控规则
DROP TABLE IF EXISTS `chat_anti_fraud_rules`;
CREATE TABLE `chat_anti_fraud_rules` (
  `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  `pattern` VARCHAR(255) NOT NULL DEFAULT '' COMMENT '匹配正则/关键词',
  `replacement` VARCHAR(255) NOT NULL DEFAULT '' COMMENT '替换文本',
  `action` TINYINT NOT NULL DEFAULT 1 COMMENT '动作 1屏蔽 2替换',
  `is_enabled` TINYINT NOT NULL DEFAULT 1 COMMENT '是否启用 0否 1是',
  `created_at` TIMESTAMP NULL DEFAULT NULL,
  `updated_at` TIMESTAMP NULL DEFAULT NULL,
  PRIMARY KEY (`id`),
  KEY `idx_is_enabled` (`is_enabled`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='飞单风控规则表';

-- 53. 飞单风控日志
DROP TABLE IF EXISTS `chat_anti_fraud_logs`;
CREATE TABLE `chat_anti_fraud_logs` (
  `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  `session_id` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '会话ID',
  `user_id` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '用户ID',
  `message_content` TEXT COMMENT '消息内容',
  `matched_pattern` VARCHAR(255) NOT NULL DEFAULT '' COMMENT '匹配规则',
  `action_taken` VARCHAR(32) NOT NULL DEFAULT '' COMMENT '执行动作',
  `created_at` TIMESTAMP NULL DEFAULT NULL,
  PRIMARY KEY (`id`),
  KEY `idx_session_id` (`session_id`),
  KEY `idx_user_id` (`user_id`),
  KEY `idx_created_at` (`created_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='飞单风控日志表';

-- 54. 快捷卡片
DROP TABLE IF EXISTS `chat_quick_cards`;
CREATE TABLE `chat_quick_cards` (
  `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  `card_type` TINYINT NOT NULL DEFAULT 1 COMMENT '卡片类型 1报价 2套餐 3预约',
  `title` VARCHAR(128) NOT NULL DEFAULT '' COMMENT '标题',
  `content` TEXT COMMENT '卡片内容(JSON)',
  `created_at` TIMESTAMP NULL DEFAULT NULL,
  PRIMARY KEY (`id`),
  KEY `idx_card_type` (`card_type`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='聊天快捷卡片表';

-- 55. 文件消息
DROP TABLE IF EXISTS `chat_file_messages`;
CREATE TABLE `chat_file_messages` (
  `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  `message_id` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '消息ID',
  `file_name` VARCHAR(255) NOT NULL DEFAULT '' COMMENT '文件名',
  `file_url` VARCHAR(512) NOT NULL DEFAULT '' COMMENT '文件URL',
  `file_size` BIGINT NOT NULL DEFAULT 0 COMMENT '文件大小(字节)',
  `file_type` VARCHAR(32) NOT NULL DEFAULT '' COMMENT '文件类型',
  `created_at` TIMESTAMP NULL DEFAULT NULL,
  PRIMARY KEY (`id`),
  KEY `idx_message_id` (`message_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='聊天文件消息表';

-- 56. 离线消息
DROP TABLE IF EXISTS `offline_messages`;
CREATE TABLE `offline_messages` (
  `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  `user_id` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '用户ID',
  `session_id` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '会话ID',
  `message_id` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '消息ID',
  `created_at` TIMESTAMP NULL DEFAULT NULL,
  PRIMARY KEY (`id`),
  KEY `idx_user_id` (`user_id`),
  KEY `idx_session_id` (`session_id`),
  KEY `idx_created_at` (`created_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='离线消息表';

-- 57. 聊天审计日志
DROP TABLE IF EXISTS `chat_audit_logs`;
CREATE TABLE `chat_audit_logs` (
  `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  `session_id` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '会话ID',
  `message_id` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '消息ID',
  `auditor_id` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '审计人ID',
  `action` VARCHAR(64) NOT NULL DEFAULT '' COMMENT '审计动作',
  `created_at` TIMESTAMP NULL DEFAULT NULL,
  PRIMARY KEY (`id`),
  KEY `idx_session_id` (`session_id`),
  KEY `idx_message_id` (`message_id`),
  KEY `idx_created_at` (`created_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='聊天审计日志表';


-- ============================================================
-- 模块七: 售后与仲裁 (After Sale & Arbitration)
-- ============================================================

-- 58. 售后风控关键词
DROP TABLE IF EXISTS `after_sale_keywords`;
CREATE TABLE `after_sale_keywords` (
  `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  `keyword` VARCHAR(128) NOT NULL DEFAULT '' COMMENT '关键词',
  `match_type` TINYINT NOT NULL DEFAULT 1 COMMENT '匹配类型 1精确 2模糊',
  `is_enabled` TINYINT NOT NULL DEFAULT 1 COMMENT '是否启用 0否 1是',
  `created_by` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '创建人ID',
  `created_at` TIMESTAMP NULL DEFAULT NULL,
  PRIMARY KEY (`id`),
  KEY `idx_is_enabled` (`is_enabled`),
  KEY `idx_created_at` (`created_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='售后风控关键词表';

-- 59. 售后风控日志
DROP TABLE IF EXISTS `after_sale_risk_logs`;
CREATE TABLE `after_sale_risk_logs` (
  `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  `session_id` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '会话ID',
  `order_id` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '订单ID',
  `sender_id` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '发送者ID',
  `keyword` VARCHAR(128) NOT NULL DEFAULT '' COMMENT '命中关键词',
  `message_summary` VARCHAR(500) NOT NULL DEFAULT '' COMMENT '消息摘要',
  `created_at` TIMESTAMP NULL DEFAULT NULL,
  PRIMARY KEY (`id`),
  KEY `idx_session_id` (`session_id`),
  KEY `idx_order_id` (`order_id`),
  KEY `idx_created_at` (`created_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='售后风控日志表';

-- 60. 介入日志
DROP TABLE IF EXISTS `after_sale_intervene_logs`;
CREATE TABLE `after_sale_intervene_logs` (
  `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  `session_id` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '会话ID',
  `order_id` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '订单ID',
  `trigger_type` TINYINT NOT NULL DEFAULT 1 COMMENT '触发类型 1关键词自动 2人工申请',
  `trigger_user_id` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '触发用户ID',
  `reason` VARCHAR(255) NOT NULL DEFAULT '' COMMENT '介入原因',
  `status` TINYINT NOT NULL DEFAULT 1 COMMENT '状态 1介入中 2已解除',
  `resolved_at` TIMESTAMP NULL DEFAULT NULL COMMENT '解除时间',
  `resolved_by` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '解除人ID',
  `created_at` TIMESTAMP NULL DEFAULT NULL,
  PRIMARY KEY (`id`),
  KEY `idx_session_id` (`session_id`),
  KEY `idx_order_id` (`order_id`),
  KEY `idx_status` (`status`),
  KEY `idx_created_at` (`created_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='售后介入日志表';

-- 61. 平台介入记录
DROP TABLE IF EXISTS `platform_intervention_logs`;
CREATE TABLE `platform_intervention_logs` (
  `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  `session_id` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '会话ID',
  `order_id` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '订单ID',
  `action` VARCHAR(64) NOT NULL DEFAULT '' COMMENT '介入动作',
  `operator_id` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '操作人ID',
  `content` TEXT COMMENT '介入内容',
  `created_at` TIMESTAMP NULL DEFAULT NULL,
  PRIMARY KEY (`id`),
  KEY `idx_session_id` (`session_id`),
  KEY `idx_order_id` (`order_id`),
  KEY `idx_created_at` (`created_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='平台介入记录表';

-- 62. 申诉工单
DROP TABLE IF EXISTS `appeal_tickets`;
CREATE TABLE `appeal_tickets` (
  `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  `ticket_no` VARCHAR(64) NOT NULL DEFAULT '' COMMENT '工单号',
  `user_id` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '用户ID',
  `order_id` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '订单ID',
  `club_id` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '俱乐部ID',
  `type` TINYINT NOT NULL DEFAULT 0 COMMENT '申诉类型',
  `description` VARCHAR(500) NOT NULL DEFAULT '' COMMENT '问题描述',
  `status` TINYINT NOT NULL DEFAULT 1 COMMENT '状态 1待审核 2处理中 3已通过 4已驳回 5需补充',
  `evidence_urls` TEXT COMMENT '证据URL(JSON数组)',
  `created_at` TIMESTAMP NULL DEFAULT NULL,
  `updated_at` TIMESTAMP NULL DEFAULT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_ticket_no` (`ticket_no`),
  KEY `idx_user_id` (`user_id`),
  KEY `idx_order_id` (`order_id`),
  KEY `idx_club_id` (`club_id`),
  KEY `idx_status` (`status`),
  KEY `idx_created_at` (`created_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='申诉工单表';

-- 63. 申诉沟通
DROP TABLE IF EXISTS `appeal_communications`;
CREATE TABLE `appeal_communications` (
  `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  `ticket_id` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '工单ID',
  `sender_id` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '发送者ID',
  `sender_type` TINYINT NOT NULL DEFAULT 1 COMMENT '发送者类型',
  `content` TEXT COMMENT '沟通内容',
  `attachments` TEXT COMMENT '附件(JSON数组)',
  `created_at` TIMESTAMP NULL DEFAULT NULL,
  PRIMARY KEY (`id`),
  KEY `idx_ticket_id` (`ticket_id`),
  KEY `idx_created_at` (`created_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='申诉沟通记录表';

-- 64. 仲裁案件
DROP TABLE IF EXISTS `arbitration_cases`;
CREATE TABLE `arbitration_cases` (
  `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  `appeal_id` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '申诉ID',
  `order_id` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '订单ID',
  `status` TINYINT NOT NULL DEFAULT 1 COMMENT '状态 1待仲裁 2仲裁中 3已判责',
  `responsible_party` VARCHAR(32) NOT NULL DEFAULT '' COMMENT '责任方',
  `judgment` TEXT COMMENT '判决结果',
  `judge_id` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '仲裁员ID',
  `created_at` TIMESTAMP NULL DEFAULT NULL,
  `updated_at` TIMESTAMP NULL DEFAULT NULL,
  PRIMARY KEY (`id`),
  KEY `idx_appeal_id` (`appeal_id`),
  KEY `idx_order_id` (`order_id`),
  KEY `idx_status` (`status`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='仲裁案件表';

-- 65. 仲裁判责规则
DROP TABLE IF EXISTS `arbitration_rules`;
CREATE TABLE `arbitration_rules` (
  `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  `rule_name` VARCHAR(64) NOT NULL DEFAULT '' COMMENT '规则名称',
  `scenario` VARCHAR(255) NOT NULL DEFAULT '' COMMENT '适用场景',
  `judgment_template` TEXT COMMENT '判责模板',
  `created_at` TIMESTAMP NULL DEFAULT NULL,
  `updated_at` TIMESTAMP NULL DEFAULT NULL,
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='仲裁判责规则表';

-- 66. 仲裁证据
DROP TABLE IF EXISTS `arbitration_evidence`;
CREATE TABLE `arbitration_evidence` (
  `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  `case_id` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '案件ID',
  `uploader_id` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '上传人ID',
  `type` TINYINT NOT NULL DEFAULT 1 COMMENT '类型 1录屏 2截图 3聊天记录 4订单截图',
  `file_url` VARCHAR(512) NOT NULL DEFAULT '' COMMENT '文件URL',
  `description` VARCHAR(255) NOT NULL DEFAULT '' COMMENT '描述',
  `created_at` TIMESTAMP NULL DEFAULT NULL,
  PRIMARY KEY (`id`),
  KEY `idx_case_id` (`case_id`),
  KEY `idx_created_at` (`created_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='仲裁证据表';

-- 67. 举证模板
DROP TABLE IF EXISTS `arbitration_evidence_templates`;
CREATE TABLE `arbitration_evidence_templates` (
  `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  `template_name` VARCHAR(64) NOT NULL DEFAULT '' COMMENT '模板名称',
  `required_items` TEXT COMMENT '必填项(JSON数组)',
  `created_at` TIMESTAMP NULL DEFAULT NULL,
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='仲裁举证模板表';


-- ============================================================
-- 模块八: 支付与财务 (Payment & Finance)
-- ============================================================

-- 68. 支付记录
DROP TABLE IF EXISTS `payments`;
CREATE TABLE `payments` (
  `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  `order_id` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '订单ID',
  `user_id` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '用户ID',
  `out_trade_no` VARCHAR(64) NOT NULL DEFAULT '' COMMENT '商户订单号',
  `transaction_id` VARCHAR(64) NOT NULL DEFAULT '' COMMENT '第三方交易号',
  `amount` BIGINT NOT NULL DEFAULT 0 COMMENT '支付金额(分)',
  `pay_type` TINYINT NOT NULL DEFAULT 1 COMMENT '支付方式 1微信 2余额 3iOS',
  `status` TINYINT NOT NULL DEFAULT 1 COMMENT '状态 1待支付 2已支付 3已退款 4部分退款',
  `paid_at` TIMESTAMP NULL DEFAULT NULL COMMENT '支付时间',
  `refund_amount` BIGINT NOT NULL DEFAULT 0 COMMENT '退款金额(分)',
  `created_at` TIMESTAMP NULL DEFAULT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_out_trade_no` (`out_trade_no`),
  KEY `idx_order_id` (`order_id`),
  KEY `idx_user_id` (`user_id`),
  KEY `idx_status` (`status`),
  KEY `idx_created_at` (`created_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='支付记录表';

-- 69. 提现
DROP TABLE IF EXISTS `withdrawals`;
CREATE TABLE `withdrawals` (
  `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  `user_id` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '用户ID',
  `user_type` TINYINT NOT NULL DEFAULT 1 COMMENT '用户类型 1打手 2俱乐部 3分销商',
  `amount` BIGINT NOT NULL DEFAULT 0 COMMENT '提现金额(分)',
  `fee` BIGINT NOT NULL DEFAULT 0 COMMENT '手续费(分)',
  `tax` BIGINT NOT NULL DEFAULT 0 COMMENT '个税(分)',
  `actual_amount` BIGINT NOT NULL DEFAULT 0 COMMENT '到账金额(分)',
  `bank_name` VARCHAR(64) NOT NULL DEFAULT '' COMMENT '开户行',
  `bank_account` VARCHAR(64) NOT NULL DEFAULT '' COMMENT '银行账号',
  `account_name` VARCHAR(64) NOT NULL DEFAULT '' COMMENT '账户名',
  `id_card` VARCHAR(255) NOT NULL DEFAULT '' COMMENT '身份证(加密)',
  `status` TINYINT NOT NULL DEFAULT 1 COMMENT '状态 1待审核 2审核通过 3已打款 4驳回',
  `batch_id` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '批次ID',
  `reject_reason` VARCHAR(255) NOT NULL DEFAULT '' COMMENT '驳回原因',
  `created_at` TIMESTAMP NULL DEFAULT NULL,
  `updated_at` TIMESTAMP NULL DEFAULT NULL,
  PRIMARY KEY (`id`),
  KEY `idx_user_id` (`user_id`),
  KEY `idx_status` (`status`),
  KEY `idx_batch_id` (`batch_id`),
  KEY `idx_created_at` (`created_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='提现记录表';

-- 70. 批量提现
DROP TABLE IF EXISTS `withdrawal_batches`;
CREATE TABLE `withdrawal_batches` (
  `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  `batch_no` VARCHAR(64) NOT NULL DEFAULT '' COMMENT '批次号',
  `total_count` INT NOT NULL DEFAULT 0 COMMENT '总笔数',
  `total_amount` BIGINT NOT NULL DEFAULT 0 COMMENT '总金额(分)',
  `status` TINYINT NOT NULL DEFAULT 1 COMMENT '状态 1待执行 2执行中 3完成 4部分失败',
  `created_by` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '创建人ID',
  `created_at` TIMESTAMP NULL DEFAULT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_batch_no` (`batch_no`),
  KEY `idx_status` (`status`),
  KEY `idx_created_at` (`created_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='批量提现批次表';

-- 71. 分账规则
DROP TABLE IF EXISTS `profit_share_rules`;
CREATE TABLE `profit_share_rules` (
  `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  `rule_name` VARCHAR(64) NOT NULL DEFAULT '' COMMENT '规则名称',
  `role_type` TINYINT NOT NULL DEFAULT 1 COMMENT '角色类型 1打手 2俱乐部 3分销商 4平台',
  `ratio` DECIMAL(5,2) NOT NULL DEFAULT 0.00 COMMENT '分账比例%',
  `min_amount` BIGINT NOT NULL DEFAULT 0 COMMENT '最低金额(分)',
  `max_amount` BIGINT NOT NULL DEFAULT 0 COMMENT '最高金额(分)',
  `created_at` TIMESTAMP NULL DEFAULT NULL,
  `updated_at` TIMESTAMP NULL DEFAULT NULL,
  PRIMARY KEY (`id`),
  KEY `idx_role_type` (`role_type`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='分账规则表';

-- 72. 分账记录
DROP TABLE IF EXISTS `profit_share_records`;
CREATE TABLE `profit_share_records` (
  `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  `order_id` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '订单ID',
  `payee_type` TINYINT NOT NULL DEFAULT 0 COMMENT '收款方类型',
  `payee_id` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '收款方ID',
  `amount` BIGINT NOT NULL DEFAULT 0 COMMENT '分账金额(分)',
  `ratio` DECIMAL(5,2) NOT NULL DEFAULT 0.00 COMMENT '分账比例%',
  `status` TINYINT NOT NULL DEFAULT 1 COMMENT '状态 1待分账 2已分账 3已回滚',
  `voucher_no` VARCHAR(64) NOT NULL DEFAULT '' COMMENT '凭证号',
  `created_at` TIMESTAMP NULL DEFAULT NULL,
  PRIMARY KEY (`id`),
  KEY `idx_order_id` (`order_id`),
  KEY `idx_payee_id` (`payee_id`),
  KEY `idx_status` (`status`),
  KEY `idx_created_at` (`created_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='分账记录表';

-- 73. 退款反向分账
DROP TABLE IF EXISTS `profit_share_refunds`;
CREATE TABLE `profit_share_refunds` (
  `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  `order_id` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '订单ID',
  `original_share_id` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '原分账记录ID',
  `recovered_amount` BIGINT NOT NULL DEFAULT 0 COMMENT '回收金额(分)',
  `status` TINYINT NOT NULL DEFAULT 1 COMMENT '状态 1待回收 2已回收',
  `created_at` TIMESTAMP NULL DEFAULT NULL,
  PRIMARY KEY (`id`),
  KEY `idx_order_id` (`order_id`),
  KEY `idx_original_share_id` (`original_share_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='退款反向分账表';

-- 74. 子商户账户
DROP TABLE IF EXISTS `merchant_accounts`;
CREATE TABLE `merchant_accounts` (
  `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  `user_id` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '用户ID',
  `user_type` TINYINT NOT NULL DEFAULT 1 COMMENT '用户类型',
  `merchant_id` VARCHAR(64) NOT NULL DEFAULT '' COMMENT '商户号',
  `merchant_name` VARCHAR(128) NOT NULL DEFAULT '' COMMENT '商户名称',
  `bank_account` VARCHAR(64) NOT NULL DEFAULT '' COMMENT '银行账号',
  `status` TINYINT NOT NULL DEFAULT 1 COMMENT '状态 1待开通 2已开通 3已冻结',
  `created_at` TIMESTAMP NULL DEFAULT NULL,
  PRIMARY KEY (`id`),
  KEY `idx_user_id` (`user_id`),
  KEY `idx_status` (`status`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='子商户账户表';

-- 75. 个税配置
DROP TABLE IF EXISTS `tax_configs`;
CREATE TABLE `tax_configs` (
  `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  `tax_type` VARCHAR(32) NOT NULL DEFAULT '' COMMENT '税种',
  `rate` DECIMAL(5,2) NOT NULL DEFAULT 0.00 COMMENT '税率%',
  `threshold` BIGINT NOT NULL DEFAULT 0 COMMENT '起征点(分)',
  `created_at` TIMESTAMP NULL DEFAULT NULL,
  `updated_at` TIMESTAMP NULL DEFAULT NULL,
  PRIMARY KEY (`id`),
  KEY `idx_tax_type` (`tax_type`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='个税配置表';

-- 76. 完税记录
DROP TABLE IF EXISTS `tax_records`;
CREATE TABLE `tax_records` (
  `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  `user_id` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '用户ID',
  `withdrawal_id` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '提现ID',
  `tax_amount` BIGINT NOT NULL DEFAULT 0 COMMENT '税额(分)',
  `tax_rate` DECIMAL(5,2) NOT NULL DEFAULT 0.00 COMMENT '税率%',
  `certificate_url` VARCHAR(512) NOT NULL DEFAULT '' COMMENT '完税凭证URL',
  `created_at` TIMESTAMP NULL DEFAULT NULL,
  PRIMARY KEY (`id`),
  KEY `idx_user_id` (`user_id`),
  KEY `idx_withdrawal_id` (`withdrawal_id`),
  KEY `idx_created_at` (`created_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='完税记录表';

-- 77. 充值记录
DROP TABLE IF EXISTS `user_recharge_logs`;
CREATE TABLE `user_recharge_logs` (
  `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  `user_id` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '用户ID',
  `amount` BIGINT NOT NULL DEFAULT 0 COMMENT '充值金额(分)',
  `bonus_amount` BIGINT NOT NULL DEFAULT 0 COMMENT '赠送金额(分)',
  `activity_id` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '活动ID',
  `pay_type` TINYINT NOT NULL DEFAULT 1 COMMENT '支付方式 1微信 2余额 3iOS',
  `status` TINYINT NOT NULL DEFAULT 1 COMMENT '状态 1待支付 2已支付 3已失败',
  `created_at` TIMESTAMP NULL DEFAULT NULL,
  PRIMARY KEY (`id`),
  KEY `idx_user_id` (`user_id`),
  KEY `idx_status` (`status`),
  KEY `idx_created_at` (`created_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='用户充值记录表';


-- ============================================================
-- 模块九: 营销 (Marketing)
-- ============================================================

-- 78. 优惠券模板
DROP TABLE IF EXISTS `coupon_templates`;
CREATE TABLE `coupon_templates` (
  `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  `name` VARCHAR(64) NOT NULL DEFAULT '' COMMENT '券名称',
  `type` TINYINT NOT NULL DEFAULT 1 COMMENT '类型 1新人券 2满减券 3俱乐部专属 4充值赠送 5售后补偿 6折扣券',
  `amount` BIGINT NOT NULL DEFAULT 0 COMMENT '优惠金额(分)',
  `min_spend` BIGINT NOT NULL DEFAULT 0 COMMENT '最低消费(分)',
  `discount_ratio` DECIMAL(3,2) NOT NULL DEFAULT 0.00 COMMENT '折扣比例',
  `total_count` INT NOT NULL DEFAULT 0 COMMENT '发放总量',
  `issued_count` INT NOT NULL DEFAULT 0 COMMENT '已发放数',
  `used_count` INT NOT NULL DEFAULT 0 COMMENT '已使用数',
  `expire_at` TIMESTAMP NULL DEFAULT NULL COMMENT '过期时间',
  `status` TINYINT NOT NULL DEFAULT 1 COMMENT '状态 1启用 2停用',
  `created_at` TIMESTAMP NULL DEFAULT NULL,
  PRIMARY KEY (`id`),
  KEY `idx_type` (`type`),
  KEY `idx_status` (`status`),
  KEY `idx_created_at` (`created_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='优惠券模板表';

-- 79. 用户优惠券
DROP TABLE IF EXISTS `user_coupons`;
CREATE TABLE `user_coupons` (
  `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  `user_id` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '用户ID',
  `template_id` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '模板ID',
  `club_id` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '俱乐部ID',
  `status` TINYINT NOT NULL DEFAULT 1 COMMENT '状态 1未使用 2已使用 3已过期',
  `used_order_id` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '使用订单ID',
  `expire_at` TIMESTAMP NULL DEFAULT NULL COMMENT '过期时间',
  `created_at` TIMESTAMP NULL DEFAULT NULL,
  PRIMARY KEY (`id`),
  KEY `idx_user_id` (`user_id`),
  KEY `idx_template_id` (`template_id`),
  KEY `idx_club_id` (`club_id`),
  KEY `idx_status` (`status`),
  KEY `idx_created_at` (`created_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='用户优惠券表';

-- 80. 充值活动
DROP TABLE IF EXISTS `recharge_activities`;
CREATE TABLE `recharge_activities` (
  `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  `name` VARCHAR(64) NOT NULL DEFAULT '' COMMENT '活动名称',
  `min_recharge` BIGINT NOT NULL DEFAULT 0 COMMENT '最低充值(分)',
  `bonus_amount` BIGINT NOT NULL DEFAULT 0 COMMENT '赠送金额(分)',
  `bonus_ratio` DECIMAL(5,2) NOT NULL DEFAULT 0.00 COMMENT '赠送比例%',
  `status` TINYINT NOT NULL DEFAULT 1 COMMENT '状态 1启用 2停用',
  `start_at` TIMESTAMP NULL DEFAULT NULL COMMENT '开始时间',
  `end_at` TIMESTAMP NULL DEFAULT NULL COMMENT '结束时间',
  `created_at` TIMESTAMP NULL DEFAULT NULL,
  PRIMARY KEY (`id`),
  KEY `idx_status` (`status`),
  KEY `idx_created_at` (`created_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='充值活动表';

-- 81. 抽奖活动
DROP TABLE IF EXISTS `lottery_activities`;
CREATE TABLE `lottery_activities` (
  `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  `name` VARCHAR(64) NOT NULL DEFAULT '' COMMENT '活动名称',
  `type` TINYINT NOT NULL DEFAULT 1 COMMENT '类型 1转盘',
  `start_at` TIMESTAMP NULL DEFAULT NULL COMMENT '开始时间',
  `end_at` TIMESTAMP NULL DEFAULT NULL COMMENT '结束时间',
  `daily_limit` INT NOT NULL DEFAULT 0 COMMENT '每日抽奖次数',
  `status` TINYINT NOT NULL DEFAULT 1 COMMENT '状态 1启用 2停用',
  `created_at` TIMESTAMP NULL DEFAULT NULL,
  PRIMARY KEY (`id`),
  KEY `idx_status` (`status`),
  KEY `idx_created_at` (`created_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='抽奖活动表';

-- 82. 奖品
DROP TABLE IF EXISTS `lottery_prizes`;
CREATE TABLE `lottery_prizes` (
  `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  `activity_id` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '活动ID',
  `name` VARCHAR(64) NOT NULL DEFAULT '' COMMENT '奖品名称',
  `type` TINYINT NOT NULL DEFAULT 1 COMMENT '类型 1优惠券 2免费时长 3平台余额',
  `value` TEXT COMMENT '奖品值(JSON)',
  `probability` DECIMAL(5,2) NOT NULL DEFAULT 0.00 COMMENT '中奖概率%',
  `stock` INT NOT NULL DEFAULT 0 COMMENT '库存',
  `issued` INT NOT NULL DEFAULT 0 COMMENT '已发放',
  `created_at` TIMESTAMP NULL DEFAULT NULL,
  PRIMARY KEY (`id`),
  KEY `idx_activity_id` (`activity_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='抽奖奖品表';

-- 83. 抽奖记录
DROP TABLE IF EXISTS `lottery_records`;
CREATE TABLE `lottery_records` (
  `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  `activity_id` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '活动ID',
  `user_id` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '用户ID',
  `prize_id` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '奖品ID',
  `created_at` TIMESTAMP NULL DEFAULT NULL,
  PRIMARY KEY (`id`),
  KEY `idx_activity_id` (`activity_id`),
  KEY `idx_user_id` (`user_id`),
  KEY `idx_created_at` (`created_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='抽奖记录表';

-- 84. 拼团活动
DROP TABLE IF EXISTS `group_buy_activities`;
CREATE TABLE `group_buy_activities` (
  `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  `name` VARCHAR(64) NOT NULL DEFAULT '' COMMENT '活动名称',
  `service_id` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '服务项目ID',
  `original_price` BIGINT NOT NULL DEFAULT 0 COMMENT '原价(分)',
  `group_price` BIGINT NOT NULL DEFAULT 0 COMMENT '拼团价(分)',
  `min_members` INT NOT NULL DEFAULT 0 COMMENT '成团人数',
  `max_members` INT NOT NULL DEFAULT 0 COMMENT '最大人数',
  `expire_hours` INT NOT NULL DEFAULT 0 COMMENT '拼团有效时长(小时)',
  `status` TINYINT NOT NULL DEFAULT 1 COMMENT '状态 1启用 2停用',
  `created_at` TIMESTAMP NULL DEFAULT NULL,
  PRIMARY KEY (`id`),
  KEY `idx_status` (`status`),
  KEY `idx_created_at` (`created_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='拼团活动表';

-- 85. 拼团成员
DROP TABLE IF EXISTS `group_buy_members`;
CREATE TABLE `group_buy_members` (
  `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  `activity_id` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '活动ID',
  `group_order_id` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '拼团订单ID',
  `user_id` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '用户ID',
  `joined_at` TIMESTAMP NULL DEFAULT NULL COMMENT '参团时间',
  `created_at` TIMESTAMP NULL DEFAULT NULL,
  PRIMARY KEY (`id`),
  KEY `idx_activity_id` (`activity_id`),
  KEY `idx_group_order_id` (`group_order_id`),
  KEY `idx_user_id` (`user_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='拼团成员表';

-- 86. 拼团订单
DROP TABLE IF EXISTS `group_buy_orders`;
CREATE TABLE `group_buy_orders` (
  `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  `activity_id` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '活动ID',
  `leader_user_id` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '团长用户ID',
  `status` TINYINT NOT NULL DEFAULT 1 COMMENT '状态 1拼团中 2成功 3失败',
  `expire_at` TIMESTAMP NULL DEFAULT NULL COMMENT '过期时间',
  `created_at` TIMESTAMP NULL DEFAULT NULL,
  PRIMARY KEY (`id`),
  KEY `idx_activity_id` (`activity_id`),
  KEY `idx_leader_user_id` (`leader_user_id`),
  KEY `idx_status` (`status`),
  KEY `idx_created_at` (`created_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='拼团订单表';


-- ============================================================
-- 模块十: 风控 (Risk Control)
-- ============================================================

-- 87. 风险用户
DROP TABLE IF EXISTS `risk_users`;
CREATE TABLE `risk_users` (
  `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  `user_id` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '用户ID',
  `risk_level` TINYINT NOT NULL DEFAULT 1 COMMENT '风险等级 1低 2中 3高 4最高',
  `risk_reason` VARCHAR(255) NOT NULL DEFAULT '' COMMENT '风险原因',
  `risk_tags` TEXT COMMENT '风险标签(JSON数组)',
  `created_at` TIMESTAMP NULL DEFAULT NULL,
  `updated_at` TIMESTAMP NULL DEFAULT NULL,
  PRIMARY KEY (`id`),
  KEY `idx_user_id` (`user_id`),
  KEY `idx_risk_level` (`risk_level`),
  KEY `idx_created_at` (`created_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='风险用户表';

-- 88. 风控日志
DROP TABLE IF EXISTS `risk_control_logs`;
CREATE TABLE `risk_control_logs` (
  `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  `user_id` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '用户ID',
  `action` VARCHAR(64) NOT NULL DEFAULT '' COMMENT '风控动作',
  `reason` VARCHAR(255) NOT NULL DEFAULT '' COMMENT '原因',
  `operator_id` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '操作人ID',
  `created_at` TIMESTAMP NULL DEFAULT NULL,
  PRIMARY KEY (`id`),
  KEY `idx_user_id` (`user_id`),
  KEY `idx_created_at` (`created_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='风控日志表';

-- 89. 敏感词库
DROP TABLE IF EXISTS `sensitive_words`;
CREATE TABLE `sensitive_words` (
  `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  `word` VARCHAR(128) NOT NULL DEFAULT '' COMMENT '敏感词',
  `category` TINYINT NOT NULL DEFAULT 1 COMMENT '分类 1违禁 2飞单 3代练 4赌博 5其他',
  `is_enabled` TINYINT NOT NULL DEFAULT 1 COMMENT '是否启用 0否 1是',
  `created_at` TIMESTAMP NULL DEFAULT NULL,
  PRIMARY KEY (`id`),
  KEY `idx_category` (`category`),
  KEY `idx_is_enabled` (`is_enabled`),
  KEY `idx_created_at` (`created_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='敏感词库表';

-- 90. 代练拦截规则
DROP TABLE IF EXISTS `anti_boosting_rules`;
CREATE TABLE `anti_boosting_rules` (
  `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  `pattern` VARCHAR(255) NOT NULL DEFAULT '' COMMENT '匹配规则',
  `action` TINYINT NOT NULL DEFAULT 1 COMMENT '动作 1拦截 2警告',
  `is_enabled` TINYINT NOT NULL DEFAULT 1 COMMENT '是否启用 0否 1是',
  `created_at` TIMESTAMP NULL DEFAULT NULL,
  PRIMARY KEY (`id`),
  KEY `idx_is_enabled` (`is_enabled`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='代练拦截规则表';

-- 91. 代练拦截日志
DROP TABLE IF EXISTS `anti_boosting_logs`;
CREATE TABLE `anti_boosting_logs` (
  `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  `user_id` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '用户ID',
  `source` TINYINT NOT NULL DEFAULT 1 COMMENT '来源 1订单 2聊天',
  `content` TEXT COMMENT '触发内容',
  `matched_pattern` VARCHAR(255) NOT NULL DEFAULT '' COMMENT '匹配规则',
  `created_at` TIMESTAMP NULL DEFAULT NULL,
  PRIMARY KEY (`id`),
  KEY `idx_user_id` (`user_id`),
  KEY `idx_created_at` (`created_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='代练拦截日志表';

-- 92. AI风险预警
DROP TABLE IF EXISTS `ai_risk_alerts`;
CREATE TABLE `ai_risk_alerts` (
  `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  `alert_type` TINYINT NOT NULL DEFAULT 1 COMMENT '预警类型 1高频退款 2批量注册 3集中提现 4深夜异常',
  `target_user_id` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '目标用户ID',
  `target_ip` VARCHAR(64) NOT NULL DEFAULT '' COMMENT '目标IP',
  `risk_data` TEXT COMMENT '风险数据(JSON)',
  `status` TINYINT NOT NULL DEFAULT 1 COMMENT '状态 1待处理 2已处理 3已忽略',
  `handled_by` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '处理人ID',
  `created_at` TIMESTAMP NULL DEFAULT NULL,
  PRIMARY KEY (`id`),
  KEY `idx_alert_type` (`alert_type`),
  KEY `idx_target_user_id` (`target_user_id`),
  KEY `idx_status` (`status`),
  KEY `idx_created_at` (`created_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='AI风险预警表';

-- 93. 宵禁日志
DROP TABLE IF EXISTS `minor_curfew_logs`;
CREATE TABLE `minor_curfew_logs` (
  `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  `user_id` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '用户ID',
  `action_attempted` VARCHAR(64) NOT NULL DEFAULT '' COMMENT '尝试操作',
  `blocked_at` TIMESTAMP NULL DEFAULT NULL COMMENT '拦截时间',
  `created_at` TIMESTAMP NULL DEFAULT NULL,
  PRIMARY KEY (`id`),
  KEY `idx_user_id` (`user_id`),
  KEY `idx_created_at` (`created_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='未成年宵禁日志表';

-- 94. 消费预警
DROP TABLE IF EXISTS `minor_consume_warnings`;
CREATE TABLE `minor_consume_warnings` (
  `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  `user_id` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '用户ID',
  `month` VARCHAR(7) NOT NULL DEFAULT '' COMMENT '月份 YYYY-MM',
  `total_amount` BIGINT NOT NULL DEFAULT 0 COMMENT '消费总额(分)',
  `warning_level` TINYINT NOT NULL DEFAULT 1 COMMENT '预警等级',
  `guardian_notified` TINYINT NOT NULL DEFAULT 0 COMMENT '是否已通知监护人 0否 1是',
  `created_at` TIMESTAMP NULL DEFAULT NULL,
  PRIMARY KEY (`id`),
  KEY `idx_user_id` (`user_id`),
  KEY `idx_month` (`month`),
  KEY `idx_created_at` (`created_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='未成年消费预警表';

-- 95. 服务保证金
DROP TABLE IF EXISTS `service_deposits`;
CREATE TABLE `service_deposits` (
  `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  `player_id` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '打手ID',
  `amount` BIGINT NOT NULL DEFAULT 0 COMMENT '保证金金额(分)',
  `status` TINYINT NOT NULL DEFAULT 1 COMMENT '状态 1已缴 2已扣除 3已退还',
  `created_at` TIMESTAMP NULL DEFAULT NULL,
  PRIMARY KEY (`id`),
  KEY `idx_player_id` (`player_id`),
  KEY `idx_status` (`status`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='服务保证金表';

-- 96. 保证金日志
DROP TABLE IF EXISTS `service_deposit_logs`;
CREATE TABLE `service_deposit_logs` (
  `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  `deposit_id` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '保证金ID',
  `action` VARCHAR(64) NOT NULL DEFAULT '' COMMENT '动作',
  `amount` BIGINT NOT NULL DEFAULT 0 COMMENT '金额(分)',
  `reason` VARCHAR(255) NOT NULL DEFAULT '' COMMENT '原因',
  `created_at` TIMESTAMP NULL DEFAULT NULL,
  PRIMARY KEY (`id`),
  KEY `idx_deposit_id` (`deposit_id`),
  KEY `idx_created_at` (`created_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='服务保证金日志表';

-- 97. 处罚日志
DROP TABLE IF EXISTS `punishment_logs`;
CREATE TABLE `punishment_logs` (
  `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  `user_id` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '用户ID',
  `club_id` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '俱乐部ID',
  `type` TINYINT NOT NULL DEFAULT 1 COMMENT '处罚类型 1禁言 2封禁 3冻结资金 4飞单永久封禁 5扣除保证金',
  `reason` VARCHAR(255) NOT NULL DEFAULT '' COMMENT '处罚原因',
  `duration` VARCHAR(32) NOT NULL DEFAULT '' COMMENT '处罚时长',
  `operator_id` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '操作人ID',
  `is_public` TINYINT NOT NULL DEFAULT 0 COMMENT '是否公示 0否 1是',
  `created_at` TIMESTAMP NULL DEFAULT NULL,
  PRIMARY KEY (`id`),
  KEY `idx_user_id` (`user_id`),
  KEY `idx_club_id` (`club_id`),
  KEY `idx_type` (`type`),
  KEY `idx_created_at` (`created_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='处罚日志表';


-- ============================================================
-- 模块十一: 平台 (Platform)
-- ============================================================

-- 98. 平台官方账号
DROP TABLE IF EXISTS `platform_official_accounts`;
CREATE TABLE `platform_official_accounts` (
  `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  `username` VARCHAR(64) NOT NULL DEFAULT '' COMMENT '账号',
  `nickname` VARCHAR(64) NOT NULL DEFAULT '' COMMENT '昵称',
  `password` VARCHAR(255) NOT NULL DEFAULT '' COMMENT '密码(bcrypt)',
  `status` TINYINT NOT NULL DEFAULT 1 COMMENT '状态 1正常 2停用',
  `created_by` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '创建人ID',
  `created_at` TIMESTAMP NULL DEFAULT NULL,
  `updated_at` TIMESTAMP NULL DEFAULT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_username` (`username`),
  KEY `idx_status` (`status`),
  KEY `idx_created_at` (`created_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='平台官方账号表';

-- 99. 平台文档
DROP TABLE IF EXISTS `platform_documents`;
CREATE TABLE `platform_documents` (
  `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  `name` VARCHAR(128) NOT NULL DEFAULT '' COMMENT '文档名称',
  `type` TINYINT NOT NULL DEFAULT 1 COMMENT '类型 1协议 2政策 3合同',
  `file_url` VARCHAR(512) NOT NULL DEFAULT '' COMMENT '文件URL',
  `version` VARCHAR(32) NOT NULL DEFAULT '' COMMENT '版本号',
  `is_deleted` TINYINT NOT NULL DEFAULT 0 COMMENT '是否删除 0否 1是',
  `created_by` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '创建人ID',
  `created_at` TIMESTAMP NULL DEFAULT NULL,
  `updated_at` TIMESTAMP NULL DEFAULT NULL,
  PRIMARY KEY (`id`),
  KEY `idx_type` (`type`),
  KEY `idx_created_at` (`created_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='平台文档表';

-- 100. 文档版本
DROP TABLE IF EXISTS `document_versions`;
CREATE TABLE `document_versions` (
  `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  `document_id` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '文档ID',
  `file_url` VARCHAR(512) NOT NULL DEFAULT '' COMMENT '文件URL',
  `version` VARCHAR(32) NOT NULL DEFAULT '' COMMENT '版本号',
  `uploaded_by` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '上传人ID',
  `created_at` TIMESTAMP NULL DEFAULT NULL,
  PRIMARY KEY (`id`),
  KEY `idx_document_id` (`document_id`),
  KEY `idx_created_at` (`created_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='文档版本表';

-- 101. 分角色协议版本
DROP TABLE IF EXISTS `agreement_role_versions`;
CREATE TABLE `agreement_role_versions` (
  `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  `role_type` TINYINT NOT NULL DEFAULT 1 COMMENT '角色 1玩家 2打手 3分销商 4俱乐部',
  `document_id` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '文档ID',
  `version` VARCHAR(32) NOT NULL DEFAULT '' COMMENT '版本号',
  `is_active` TINYINT NOT NULL DEFAULT 1 COMMENT '是否生效 0否 1是',
  `created_at` TIMESTAMP NULL DEFAULT NULL,
  PRIMARY KEY (`id`),
  KEY `idx_role_type` (`role_type`),
  KEY `idx_document_id` (`document_id`),
  KEY `idx_created_at` (`created_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='分角色协议版本表';

-- 102. 协议签署日志
DROP TABLE IF EXISTS `agreement_sign_logs`;
CREATE TABLE `agreement_sign_logs` (
  `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  `user_id` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '用户ID',
  `agreement_id` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '协议ID',
  `role_type` TINYINT NOT NULL DEFAULT 1 COMMENT '角色类型',
  `ip` VARCHAR(64) NOT NULL DEFAULT '' COMMENT 'IP',
  `user_agent` VARCHAR(255) NOT NULL DEFAULT '' COMMENT 'User-Agent',
  `created_at` TIMESTAMP NULL DEFAULT NULL,
  PRIMARY KEY (`id`),
  KEY `idx_user_id` (`user_id`),
  KEY `idx_agreement_id` (`agreement_id`),
  KEY `idx_created_at` (`created_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='协议签署日志表';

-- 103. 系统配置
DROP TABLE IF EXISTS `system_configs`;
CREATE TABLE `system_configs` (
  `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  `config_key` VARCHAR(128) NOT NULL DEFAULT '' COMMENT '配置键',
  `config_value` TEXT COMMENT '配置值',
  `config_desc` VARCHAR(255) NOT NULL DEFAULT '' COMMENT '配置描述',
  `is_editable` TINYINT NOT NULL DEFAULT 1 COMMENT '是否可编辑 0否 1是',
  `updated_by` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '更新人ID',
  `updated_at` TIMESTAMP NULL DEFAULT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_config_key` (`config_key`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='系统配置表';

-- 104. 通知
DROP TABLE IF EXISTS `notifications`;
CREATE TABLE `notifications` (
  `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  `user_id` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '用户ID',
  `title` VARCHAR(128) NOT NULL DEFAULT '' COMMENT '标题',
  `content` TEXT COMMENT '内容',
  `type` TINYINT NOT NULL DEFAULT 0 COMMENT '通知类型',
  `is_read` TINYINT NOT NULL DEFAULT 0 COMMENT '是否已读 0否 1是',
  `created_at` TIMESTAMP NULL DEFAULT NULL,
  PRIMARY KEY (`id`),
  KEY `idx_user_id` (`user_id`),
  KEY `idx_is_read` (`is_read`),
  KEY `idx_created_at` (`created_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='通知表';


-- ============================================================
-- 模块十二: 派单与打手 (Dispatch & Player)
-- ============================================================

-- 105. 派单记录
DROP TABLE IF EXISTS `dispatch_records`;
CREATE TABLE `dispatch_records` (
  `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  `dispatcher_id` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '派单员ID',
  `order_id` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '订单ID',
  `assigned_player_id` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '指派打手ID',
  `dispatch_type` TINYINT NOT NULL DEFAULT 1 COMMENT '派单类型 1手动 2转单',
  `created_at` TIMESTAMP NULL DEFAULT NULL,
  PRIMARY KEY (`id`),
  KEY `idx_dispatcher_id` (`dispatcher_id`),
  KEY `idx_order_id` (`order_id`),
  KEY `idx_assigned_player_id` (`assigned_player_id`),
  KEY `idx_created_at` (`created_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='派单记录表';

-- 106. 打手标签
DROP TABLE IF EXISTS `player_tags`;
CREATE TABLE `player_tags` (
  `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  `player_id` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '打手ID',
  `game` VARCHAR(64) NOT NULL DEFAULT '' COMMENT '游戏',
  `position` VARCHAR(64) NOT NULL DEFAULT '' COMMENT '位置',
  `voice_type` VARCHAR(32) NOT NULL DEFAULT '' COMMENT '语音类型',
  `rank` VARCHAR(32) NOT NULL DEFAULT '' COMMENT '段位',
  `skills` TEXT COMMENT '技能标签(JSON数组)',
  `created_at` TIMESTAMP NULL DEFAULT NULL,
  PRIMARY KEY (`id`),
  KEY `idx_player_id` (`player_id`),
  KEY `idx_game` (`game`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='打手标签表';

-- 107. 收藏打手
DROP TABLE IF EXISTS `player_favorites`;
CREATE TABLE `player_favorites` (
  `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  `user_id` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '用户ID',
  `player_id` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '打手ID',
  `created_at` TIMESTAMP NULL DEFAULT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_user_player` (`user_id`, `player_id`),
  KEY `idx_user_id` (`user_id`),
  KEY `idx_player_id` (`player_id`),
  KEY `idx_created_at` (`created_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='收藏打手表';

-- 108. 打手服务项目
DROP TABLE IF EXISTS `player_services`;
CREATE TABLE `player_services` (
  `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  `player_id` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '打手ID',
  `club_id` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '俱乐部ID',
  `service_name` VARCHAR(128) NOT NULL DEFAULT '' COMMENT '服务名称',
  `price` BIGINT NOT NULL DEFAULT 0 COMMENT '价格(分)',
  `desc` VARCHAR(500) NOT NULL DEFAULT '' COMMENT '描述',
  `cover_url` VARCHAR(512) NOT NULL DEFAULT '' COMMENT '封面URL',
  `status` TINYINT NOT NULL DEFAULT 1 COMMENT '状态 1上架 2下架',
  `created_at` TIMESTAMP NULL DEFAULT NULL,
  PRIMARY KEY (`id`),
  KEY `idx_player_id` (`player_id`),
  KEY `idx_club_id` (`club_id`),
  KEY `idx_status` (`status`),
  KEY `idx_created_at` (`created_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='打手服务项目表';

-- 109. 活跃度探针
DROP TABLE IF EXISTS `player_probe_logs`;
CREATE TABLE `player_probe_logs` (
  `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  `player_id` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '打手ID',
  `last_ping_at` TIMESTAMP NULL DEFAULT NULL COMMENT '最后心跳时间',
  `ping_count` INT NOT NULL DEFAULT 0 COMMENT '心跳次数',
  `is_active` TINYINT NOT NULL DEFAULT 1 COMMENT '是否活跃 0否 1是',
  `created_at` TIMESTAMP NULL DEFAULT NULL,
  PRIMARY KEY (`id`),
  KEY `idx_player_id` (`player_id`),
  KEY `idx_is_active` (`is_active`),
  KEY `idx_created_at` (`created_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='打手活跃度探针表';

-- 110. 打手等级配置
DROP TABLE IF EXISTS `player_level_configs`;
CREATE TABLE `player_level_configs` (
  `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  `level` TINYINT NOT NULL DEFAULT 1 COMMENT '等级 1青铜 2白银 3黄金 4钻石',
  `level_name` VARCHAR(32) NOT NULL DEFAULT '' COMMENT '等级名称',
  `min_credit_score` INT NOT NULL DEFAULT 0 COMMENT '最低信用分',
  `order_weight` INT NOT NULL DEFAULT 0 COMMENT '接单权重',
  `fee_discount` DECIMAL(5,2) NOT NULL DEFAULT 0.00 COMMENT '手续费折扣%',
  `created_at` TIMESTAMP NULL DEFAULT NULL,
  PRIMARY KEY (`id`),
  KEY `idx_level` (`level`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='打手等级配置表';


-- ============================================================
-- 模块十三: UP主 (UP Master)
-- ============================================================

-- 111. UP主认证
DROP TABLE IF EXISTS `up_master_certifications`;
CREATE TABLE `up_master_certifications` (
  `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  `user_id` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '用户ID',
  `platform` VARCHAR(64) NOT NULL DEFAULT '' COMMENT '平台',
  `follower_count` INT NOT NULL DEFAULT 0 COMMENT '粉丝数',
  `tier` TINYINT NOT NULL DEFAULT 1 COMMENT '等级 1-6',
  `status` TINYINT NOT NULL DEFAULT 1 COMMENT '状态 1待审核 2通过 3驳回 4已吊销',
  `evidence_urls` TEXT COMMENT '证据URL(JSON数组)',
  `created_at` TIMESTAMP NULL DEFAULT NULL,
  `updated_at` TIMESTAMP NULL DEFAULT NULL,
  PRIMARY KEY (`id`),
  KEY `idx_user_id` (`user_id`),
  KEY `idx_tier` (`tier`),
  KEY `idx_status` (`status`),
  KEY `idx_created_at` (`created_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='UP主认证表';

-- 112. UP主等级配置
DROP TABLE IF EXISTS `up_master_tier_configs`;
CREATE TABLE `up_master_tier_configs` (
  `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  `tier` TINYINT NOT NULL DEFAULT 1 COMMENT '等级 1-6',
  `min_followers` INT NOT NULL DEFAULT 0 COMMENT '最低粉丝数',
  `max_followers` INT NOT NULL DEFAULT 0 COMMENT '最高粉丝数',
  `badge_name` VARCHAR(32) NOT NULL DEFAULT '' COMMENT '徽章名称',
  `created_at` TIMESTAMP NULL DEFAULT NULL,
  PRIMARY KEY (`id`),
  KEY `idx_tier` (`tier`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='UP主等级配置表';


-- ============================================================
-- 模块十四: 分销商 (Distributor)
-- ============================================================

-- 113. 分销关系
DROP TABLE IF EXISTS `distributor_relations`;
CREATE TABLE `distributor_relations` (
  `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  `distributor_id` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '分销商ID',
  `subordinate_id` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '下级ID',
  `level` TINYINT NOT NULL DEFAULT 1 COMMENT '级别 1一级 2二级',
  `is_effective` TINYINT NOT NULL DEFAULT 1 COMMENT '是否有效 0否 1是',
  `effective_at` TIMESTAMP NULL DEFAULT NULL COMMENT '生效时间',
  `created_at` TIMESTAMP NULL DEFAULT NULL,
  PRIMARY KEY (`id`),
  KEY `idx_distributor_id` (`distributor_id`),
  KEY `idx_subordinate_id` (`subordinate_id`),
  KEY `idx_created_at` (`created_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='分销关系表';

-- 114. 分销佣金
DROP TABLE IF EXISTS `distributor_commissions`;
CREATE TABLE `distributor_commissions` (
  `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  `distributor_id` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '分销商ID',
  `order_id` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '订单ID',
  `subordinate_id` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '下级ID',
  `commission_amount` BIGINT NOT NULL DEFAULT 0 COMMENT '佣金金额(分)',
  `ratio` DECIMAL(5,2) NOT NULL DEFAULT 0.00 COMMENT '佣金比例%',
  `status` TINYINT NOT NULL DEFAULT 1 COMMENT '状态 1待结算 2已结算 3已回滚',
  `created_at` TIMESTAMP NULL DEFAULT NULL,
  PRIMARY KEY (`id`),
  KEY `idx_distributor_id` (`distributor_id`),
  KEY `idx_order_id` (`order_id`),
  KEY `idx_status` (`status`),
  KEY `idx_created_at` (`created_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='分销佣金表';

-- 115. 首单奖励
DROP TABLE IF EXISTS `distributor_first_rewards`;
CREATE TABLE `distributor_first_rewards` (
  `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  `distributor_id` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '分销商ID',
  `subordinate_id` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '下级ID',
  `order_id` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '订单ID',
  `reward_amount` BIGINT NOT NULL DEFAULT 0 COMMENT '奖励金额(分)',
  `created_at` TIMESTAMP NULL DEFAULT NULL,
  PRIMARY KEY (`id`),
  KEY `idx_distributor_id` (`distributor_id`),
  KEY `idx_order_id` (`order_id`),
  KEY `idx_created_at` (`created_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='分销首单奖励表';


-- ============================================================
-- 模块十五: 店铺管理端 (Shop Admin)
-- ============================================================

-- 116. 内置管理端账号
DROP TABLE IF EXISTS `shop_admin_accounts`;
CREATE TABLE `shop_admin_accounts` (
  `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  `username` VARCHAR(64) NOT NULL DEFAULT '' COMMENT '账号',
  `club_id` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '俱乐部ID',
  `password` VARCHAR(255) NOT NULL DEFAULT '' COMMENT '密码(bcrypt)',
  `identity` TINYINT NOT NULL DEFAULT 2 COMMENT '身份 1创始人 2管理员',
  `real_name` VARCHAR(64) NOT NULL DEFAULT '' COMMENT '真实姓名',
  `phone` VARCHAR(20) NOT NULL DEFAULT '' COMMENT '手机号',
  `status` TINYINT NOT NULL DEFAULT 1 COMMENT '状态 1正常 2禁用',
  `last_login_at` TIMESTAMP NULL DEFAULT NULL COMMENT '最后登录时间',
  `last_login_ip` VARCHAR(64) NOT NULL DEFAULT '' COMMENT '最后登录IP',
  `seq_no` INT NOT NULL DEFAULT 0 COMMENT '序号',
  `created_at` TIMESTAMP NULL DEFAULT NULL,
  `updated_at` TIMESTAMP NULL DEFAULT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_club_username` (`club_id`, `username`),
  KEY `idx_club_id` (`club_id`),
  KEY `idx_phone` (`phone`),
  KEY `idx_status` (`status`),
  KEY `idx_created_at` (`created_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='店铺管理端账号表';


-- ============================================================
-- 模块十六: 入会与考核 (Join & Assessment)
-- ============================================================

-- 117. 入会申请
DROP TABLE IF EXISTS `join_applications`;
CREATE TABLE `join_applications` (
  `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  `club_id` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '俱乐部ID',
  `user_id` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '用户ID',
  `real_name` VARCHAR(64) NOT NULL DEFAULT '' COMMENT '真实姓名',
  `game_account` VARCHAR(128) NOT NULL DEFAULT '' COMMENT '游戏账号',
  `game_zone` VARCHAR(64) NOT NULL DEFAULT '' COMMENT '游戏大区',
  `position` VARCHAR(64) NOT NULL DEFAULT '' COMMENT '位置',
  `heroes` VARCHAR(500) NOT NULL DEFAULT '' COMMENT '擅长英雄',
  `rank` VARCHAR(32) NOT NULL DEFAULT '' COMMENT '当前段位',
  `max_rank` VARCHAR(32) NOT NULL DEFAULT '' COMMENT '历史最高段位',
  `intro` VARCHAR(500) NOT NULL DEFAULT '' COMMENT '自我介绍',
  `status` TINYINT NOT NULL DEFAULT 1 COMMENT '状态 1待审核 2考核中 3已通过 4已驳回',
  `reject_reason` VARCHAR(255) NOT NULL DEFAULT '' COMMENT '驳回原因',
  `approved_at` TIMESTAMP NULL DEFAULT NULL COMMENT '通过时间',
  `created_at` TIMESTAMP NULL DEFAULT NULL,
  PRIMARY KEY (`id`),
  KEY `idx_club_id` (`club_id`),
  KEY `idx_user_id` (`user_id`),
  KEY `idx_status` (`status`),
  KEY `idx_created_at` (`created_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='入会申请表';

-- 118. 考核记录
DROP TABLE IF EXISTS `assessment_records`;
CREATE TABLE `assessment_records` (
  `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  `application_id` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '申请ID',
  `club_id` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '俱乐部ID',
  `applicant_id` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '申请人ID',
  `assessor_id` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '考核人ID',
  `result` TINYINT NOT NULL DEFAULT 0 COMMENT '结果 1通过 2不通过',
  `requirement` VARCHAR(255) NOT NULL DEFAULT '' COMMENT '考核要求',
  `remark` VARCHAR(500) NOT NULL DEFAULT '' COMMENT '备注',
  `video_url` VARCHAR(512) NOT NULL DEFAULT '' COMMENT '考核视频URL',
  `created_at` TIMESTAMP NULL DEFAULT NULL,
  PRIMARY KEY (`id`),
  KEY `idx_application_id` (`application_id`),
  KEY `idx_club_id` (`club_id`),
  KEY `idx_created_at` (`created_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='入会考核记录表';

-- 119. 考核模板
DROP TABLE IF EXISTS `assessment_templates`;
CREATE TABLE `assessment_templates` (
  `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  `club_id` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '俱乐部ID',
  `game` VARCHAR(64) NOT NULL DEFAULT '' COMMENT '游戏',
  `level` VARCHAR(32) NOT NULL DEFAULT '' COMMENT '等级',
  `standard` TEXT COMMENT '考核标准',
  `created_at` TIMESTAMP NULL DEFAULT NULL,
  PRIMARY KEY (`id`),
  KEY `idx_club_id` (`club_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='入会考核模板表';


-- ============================================================
-- 模块十七: 系统运维 (System)
-- ============================================================

-- 120. 批量操作日志
DROP TABLE IF EXISTS `batch_operation_logs`;
CREATE TABLE `batch_operation_logs` (
  `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  `admin_id` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '管理员ID',
  `operation_type` VARCHAR(64) NOT NULL DEFAULT '' COMMENT '操作类型',
  `total_count` INT NOT NULL DEFAULT 0 COMMENT '总数',
  `success_count` INT NOT NULL DEFAULT 0 COMMENT '成功数',
  `fail_count` INT NOT NULL DEFAULT 0 COMMENT '失败数',
  `status` TINYINT NOT NULL DEFAULT 1 COMMENT '状态 1进行中 2完成 3失败',
  `created_at` TIMESTAMP NULL DEFAULT NULL,
  PRIMARY KEY (`id`),
  KEY `idx_admin_id` (`admin_id`),
  KEY `idx_status` (`status`),
  KEY `idx_created_at` (`created_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='批量操作日志表';

-- 121. 批量操作确认
DROP TABLE IF EXISTS `batch_operation_confirms`;
CREATE TABLE `batch_operation_confirms` (
  `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  `batch_log_id` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '批量日志ID',
  `confirmer_id` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '确认人ID',
  `confirm_method` TINYINT NOT NULL DEFAULT 1 COMMENT '确认方式 1扫码',
  `confirmed_at` TIMESTAMP NULL DEFAULT NULL COMMENT '确认时间',
  `created_at` TIMESTAMP NULL DEFAULT NULL,
  PRIMARY KEY (`id`),
  KEY `idx_batch_log_id` (`batch_log_id`),
  KEY `idx_created_at` (`created_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='批量操作确认表';

-- 122. 备份记录
DROP TABLE IF EXISTS `backup_records`;
CREATE TABLE `backup_records` (
  `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  `file_name` VARCHAR(255) NOT NULL DEFAULT '' COMMENT '文件名',
  `file_url` VARCHAR(512) NOT NULL DEFAULT '' COMMENT '文件URL',
  `file_size` BIGINT NOT NULL DEFAULT 0 COMMENT '文件大小(字节)',
  `encrypt_method` VARCHAR(32) NOT NULL DEFAULT '' COMMENT '加密方式',
  `status` TINYINT NOT NULL DEFAULT 1 COMMENT '状态 1进行中 2成功 3失败',
  `created_at` TIMESTAMP NULL DEFAULT NULL,
  PRIMARY KEY (`id`),
  KEY `idx_status` (`status`),
  KEY `idx_created_at` (`created_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='数据库备份记录表';

-- 123. 恢复记录
DROP TABLE IF EXISTS `restore_records`;
CREATE TABLE `restore_records` (
  `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  `backup_id` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '备份ID',
  `restore_to` TIMESTAMP NULL DEFAULT NULL COMMENT '恢复至时间点',
  `status` TINYINT NOT NULL DEFAULT 1 COMMENT '状态 1进行中 2成功 3失败',
  `created_at` TIMESTAMP NULL DEFAULT NULL,
  PRIMARY KEY (`id`),
  KEY `idx_backup_id` (`backup_id`),
  KEY `idx_status` (`status`),
  KEY `idx_created_at` (`created_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='数据库恢复记录表';

-- 124. 灰度发布
DROP TABLE IF EXISTS `gray_releases`;
CREATE TABLE `gray_releases` (
  `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  `api_version` VARCHAR(32) NOT NULL DEFAULT '' COMMENT 'API版本',
  `whitelist` TEXT COMMENT '白名单(JSON数组)',
  `rollout_ratio` INT NOT NULL DEFAULT 0 COMMENT '灰度比例%',
  `error_threshold` DECIMAL(5,2) NOT NULL DEFAULT 0.00 COMMENT '错误率阈值%',
  `is_rollback` TINYINT NOT NULL DEFAULT 0 COMMENT '是否已回滚 0否 1是',
  `status` TINYINT NOT NULL DEFAULT 1 COMMENT '状态 1启用 2停用',
  `created_at` TIMESTAMP NULL DEFAULT NULL,
  PRIMARY KEY (`id`),
  KEY `idx_status` (`status`),
  KEY `idx_created_at` (`created_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='灰度发布配置表';

-- 125. 接口监控
DROP TABLE IF EXISTS `system_api_monitors`;
CREATE TABLE `system_api_monitors` (
  `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  `api_name` VARCHAR(128) NOT NULL DEFAULT '' COMMENT '接口名',
  `total_calls` BIGINT NOT NULL DEFAULT 0 COMMENT '总调用数',
  `success_count` BIGINT NOT NULL DEFAULT 0 COMMENT '成功数',
  `avg_latency_ms` INT NOT NULL DEFAULT 0 COMMENT '平均延迟(ms)',
  `error_rate` DECIMAL(5,2) NOT NULL DEFAULT 0.00 COMMENT '错误率%',
  `created_at` TIMESTAMP NULL DEFAULT NULL,
  PRIMARY KEY (`id`),
  KEY `idx_api_name` (`api_name`),
  KEY `idx_created_at` (`created_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='接口监控表';

-- 126. 第三方接口日志
DROP TABLE IF EXISTS `third_party_api_logs`;
CREATE TABLE `third_party_api_logs` (
  `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  `api_name` VARCHAR(128) NOT NULL DEFAULT '' COMMENT '接口名',
  `request_data` TEXT COMMENT '请求数据',
  `response_data` TEXT COMMENT '响应数据',
  `latency_ms` INT NOT NULL DEFAULT 0 COMMENT '延迟(ms)',
  `status` TINYINT NOT NULL DEFAULT 1 COMMENT '状态 1成功 2失败',
  `created_at` TIMESTAMP NULL DEFAULT NULL,
  PRIMARY KEY (`id`),
  KEY `idx_api_name` (`api_name`),
  KEY `idx_status` (`status`),
  KEY `idx_created_at` (`created_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='第三方接口日志表';

-- 127. 重试队列
DROP TABLE IF EXISTS `third_party_retry_queues`;
CREATE TABLE `third_party_retry_queues` (
  `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  `api_name` VARCHAR(128) NOT NULL DEFAULT '' COMMENT '接口名',
  `request_data` TEXT COMMENT '请求数据',
  `retry_count` INT NOT NULL DEFAULT 0 COMMENT '已重试次数',
  `max_retry` INT NOT NULL DEFAULT 0 COMMENT '最大重试次数',
  `status` TINYINT NOT NULL DEFAULT 1 COMMENT '状态 1待重试 2成功 3失败',
  `next_retry_at` TIMESTAMP NULL DEFAULT NULL COMMENT '下次重试时间',
  `created_at` TIMESTAMP NULL DEFAULT NULL,
  PRIMARY KEY (`id`),
  KEY `idx_status` (`status`),
  KEY `idx_next_retry_at` (`next_retry_at`),
  KEY `idx_created_at` (`created_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='第三方接口重试队列表';

-- 128. 定时任务日志
DROP TABLE IF EXISTS `cron_job_logs`;
CREATE TABLE `cron_job_logs` (
  `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  `job_name` VARCHAR(64) NOT NULL DEFAULT '' COMMENT '任务名',
  `status` TINYINT NOT NULL DEFAULT 1 COMMENT '状态 1成功 2失败',
  `duration_ms` INT NOT NULL DEFAULT 0 COMMENT '耗时(ms)',
  `error_msg` TEXT COMMENT '错误信息',
  `created_at` TIMESTAMP NULL DEFAULT NULL,
  PRIMARY KEY (`id`),
  KEY `idx_job_name` (`job_name`),
  KEY `idx_status` (`status`),
  KEY `idx_created_at` (`created_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='定时任务日志表';

-- 129. 慢查询日志
DROP TABLE IF EXISTS `slow_query_logs`;
CREATE TABLE `slow_query_logs` (
  `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  `query_sql` TEXT COMMENT 'SQL语句',
  `duration_ms` INT NOT NULL DEFAULT 0 COMMENT '耗时(ms)',
  `created_at` TIMESTAMP NULL DEFAULT NULL,
  PRIMARY KEY (`id`),
  KEY `idx_duration_ms` (`duration_ms`),
  KEY `idx_created_at` (`created_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='慢查询日志表';

-- 130. 监控告警
DROP TABLE IF EXISTS `monitor_alerts`;
CREATE TABLE `monitor_alerts` (
  `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  `alert_type` VARCHAR(64) NOT NULL DEFAULT '' COMMENT '告警类型',
  `alert_level` TINYINT NOT NULL DEFAULT 1 COMMENT '告警等级',
  `message` VARCHAR(500) NOT NULL DEFAULT '' COMMENT '告警信息',
  `status` TINYINT NOT NULL DEFAULT 1 COMMENT '状态 1未处理 2已处理',
  `created_at` TIMESTAMP NULL DEFAULT NULL,
  PRIMARY KEY (`id`),
  KEY `idx_alert_type` (`alert_type`),
  KEY `idx_status` (`status`),
  KEY `idx_created_at` (`created_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='监控告警表';

-- 131. NTP同步日志
DROP TABLE IF EXISTS `ntp_sync_logs`;
CREATE TABLE `ntp_sync_logs` (
  `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  `offset_ms` INT NOT NULL DEFAULT 0 COMMENT '时间偏移(ms)',
  `status` TINYINT NOT NULL DEFAULT 1 COMMENT '状态 1成功 2失败',
  `created_at` TIMESTAMP NULL DEFAULT NULL,
  PRIMARY KEY (`id`),
  KEY `idx_created_at` (`created_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='NTP时间同步日志表';

-- 132. 熔断器
DROP TABLE IF EXISTS `circuit_breakers`;
CREATE TABLE `circuit_breakers` (
  `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  `service_name` VARCHAR(64) NOT NULL DEFAULT '' COMMENT '服务名',
  `state` TINYINT NOT NULL DEFAULT 1 COMMENT '状态 1关闭 2打开 3半开',
  `failure_count` INT NOT NULL DEFAULT 0 COMMENT '失败次数',
  `last_failure_at` TIMESTAMP NULL DEFAULT NULL COMMENT '最后失败时间',
  `created_at` TIMESTAMP NULL DEFAULT NULL,
  `updated_at` TIMESTAMP NULL DEFAULT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_service_name` (`service_name`),
  KEY `idx_state` (`state`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='服务熔断器表';

-- 133. 订阅消息模板
DROP TABLE IF EXISTS `subscribe_message_templates`;
CREATE TABLE `subscribe_message_templates` (
  `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  `template_id` VARCHAR(64) NOT NULL DEFAULT '' COMMENT '微信模板ID',
  `name` VARCHAR(64) NOT NULL DEFAULT '' COMMENT '模板名称',
  `type` VARCHAR(32) NOT NULL DEFAULT '' COMMENT '模板类型',
  `content` TEXT COMMENT '模板内容',
  `created_at` TIMESTAMP NULL DEFAULT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_template_id` (`template_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='订阅消息模板表';

-- 134. 订阅消息日志
DROP TABLE IF EXISTS `subscribe_message_logs`;
CREATE TABLE `subscribe_message_logs` (
  `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  `template_id` VARCHAR(64) NOT NULL DEFAULT '' COMMENT '模板ID',
  `user_id` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '用户ID',
  `send_status` TINYINT NOT NULL DEFAULT 1 COMMENT '发送状态 1成功 2失败',
  `error_msg` VARCHAR(255) NOT NULL DEFAULT '' COMMENT '错误信息',
  `created_at` TIMESTAMP NULL DEFAULT NULL,
  PRIMARY KEY (`id`),
  KEY `idx_template_id` (`template_id`),
  KEY `idx_user_id` (`user_id`),
  KEY `idx_created_at` (`created_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='订阅消息发送日志表';

-- 135. 游戏列表
DROP TABLE IF EXISTS `game_lists`;
CREATE TABLE `game_lists` (
  `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  `name` VARCHAR(64) NOT NULL DEFAULT '' COMMENT '游戏名称',
  `icon` VARCHAR(512) NOT NULL DEFAULT '' COMMENT '游戏图标URL',
  `sort_order` INT NOT NULL DEFAULT 0 COMMENT '排序',
  `status` TINYINT NOT NULL DEFAULT 1 COMMENT '状态 1启用 2停用',
  `created_at` TIMESTAMP NULL DEFAULT NULL,
  PRIMARY KEY (`id`),
  KEY `idx_status` (`status`),
  KEY `idx_sort_order` (`sort_order`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='游戏列表表';

-- 136. 店铺配置
DROP TABLE IF EXISTS `shop_configs`;
CREATE TABLE `shop_configs` (
  `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  `club_id` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '俱乐部ID',
  `config_key` VARCHAR(128) NOT NULL DEFAULT '' COMMENT '配置键',
  `config_value` TEXT COMMENT '配置值',
  `created_at` TIMESTAMP NULL DEFAULT NULL,
  `updated_at` TIMESTAMP NULL DEFAULT NULL,
  PRIMARY KEY (`id`),
  KEY `idx_club_id` (`club_id`),
  KEY `idx_config_key` (`config_key`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='店铺配置表';

-- 137. 服务类型
DROP TABLE IF EXISTS `service_types`;
CREATE TABLE `service_types` (
  `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  `name` VARCHAR(64) NOT NULL DEFAULT '' COMMENT '服务类型名称',
  `game_id` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '游戏ID',
  `description` VARCHAR(255) NOT NULL DEFAULT '' COMMENT '描述',
  `base_price` BIGINT NOT NULL DEFAULT 0 COMMENT '基础价格(分)',
  `created_at` TIMESTAMP NULL DEFAULT NULL,
  PRIMARY KEY (`id`),
  KEY `idx_game_id` (`game_id`),
  KEY `idx_created_at` (`created_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='服务类型表';


-- ============================================================
-- 初始化数据 (Initial Data)
-- ============================================================

-- 默认管理员账号 (密码: 1234567, bcrypt hash, Go兼容 $2a$ 前缀)
INSERT INTO `admins` (`username`, `nickname`, `password`, `email`, `phone`, `role`, `status`, `is_init`, `created_at`, `updated_at`) VALUES
('admin',  '超级管理员',  '$2a$10$t3GL1egpW5vMkW7JvY