-- FlyAI complete MySQL schema.
-- The official MySQL container executes this file only when data/mysql is empty.
-- MYSQL_DATABASE is selected by the container entrypoint before this script runs.

SET NAMES utf8mb4;

CREATE TABLE IF NOT EXISTS `t_ai_assistant` (
  `id` varchar(32) NOT NULL,
  `created_at` longblob,
  `name` varchar(10) NOT NULL COMMENT '名称',
  `personality` varchar(999) NOT NULL COMMENT '人设',
  `model` varchar(50) NOT NULL COMMENT '使用的模型',
  `enable` tinyint(1) NOT NULL DEFAULT '1' COMMENT '是否启用',
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;

CREATE TABLE IF NOT EXISTS `t_feature_setting` (
  `feature` varchar(32) NOT NULL,
  `send_time` varchar(5) NOT NULL,
  `schedule` varchar(16) NOT NULL DEFAULT 'daily',
  `message_threshold` bigint NOT NULL DEFAULT '0',
  `last_run_date` varchar(10) NOT NULL DEFAULT '',
  `updated_at` datetime(3) DEFAULT NULL,
  PRIMARY KEY (`feature`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;

CREATE TABLE IF NOT EXISTS `t_friend` (
  `wxid` longtext,
  `custom_account` longtext,
  `nickname` longtext,
  `pinyin` longtext,
  `pinyin_all` longtext,
  `last_active` longblob,
  `enable_monitor` tinyint(1) NOT NULL DEFAULT '0',
  `enable_ai` tinyint(1) NOT NULL DEFAULT '0',
  `ai_model` longtext,
  `prompt` longtext,
  `enable_chat_rank` tinyint(1) NOT NULL DEFAULT '0',
  `enable_yesterday_rank` tinyint(1) NOT NULL DEFAULT '0',
  `enable_weekly_rank` tinyint(1) NOT NULL DEFAULT '0',
  `enable_monthly_rank` tinyint(1) NOT NULL DEFAULT '0',
  `enable_yearly_rank` tinyint(1) NOT NULL DEFAULT '0',
  `enable_welcome` tinyint(1) NOT NULL DEFAULT '0',
  `enable_summary` tinyint(1) NOT NULL DEFAULT '0',
  `enable_news` tinyint(1) NOT NULL DEFAULT '0',
  `enable_dormant` tinyint(1) NOT NULL DEFAULT '0',
  `enable_hot_top` tinyint(1) NOT NULL DEFAULT '0',
  `clear_member` bigint DEFAULT NULL,
  `is_ok` tinyint(1) NOT NULL DEFAULT '0',
  `used_tokens` bigint DEFAULT NULL,
  `ai_free_limit` bigint DEFAULT NULL,
  `ai_used_today` bigint DEFAULT NULL,
  `enable_command` tinyint(1) NOT NULL DEFAULT '0'
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;

CREATE TABLE IF NOT EXISTS `t_group_user` (
  `group_id` longtext,
  `wxid` longtext,
  `account` longtext,
  `head_image` longtext,
  `nickname` longtext,
  `wechat_nickname` longtext,
  `is_member` tinyint(1) NOT NULL DEFAULT '0',
  `is_admin` tinyint(1) NOT NULL DEFAULT '0',
  `join_time` datetime(3) DEFAULT NULL,
  `last_active` datetime(3) DEFAULT NULL,
  `leave_time` datetime(3) DEFAULT NULL,
  `skip_chat_rank` tinyint(1) NOT NULL DEFAULT '0'
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;

CREATE TABLE IF NOT EXISTS `t_hot_top` (
  `id` varchar(32) NOT NULL,
  `created_at` longblob,
  `channel` longtext,
  `title` longtext,
  `hot` longtext,
  `url` longtext,
  `mobile_url` longtext,
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;

CREATE TABLE IF NOT EXISTS `t_message` (
  `msg_id` bigint NOT NULL AUTO_INCREMENT,
  `create_time` bigint DEFAULT NULL,
  `create_at` datetime(3) DEFAULT NULL,
  `type` bigint DEFAULT NULL,
  `content` longtext,
  `display_full_content` longtext,
    `from_user` longtext,
    `group_user` longtext,
    `sender_nickname` longtext,
    `sender_group_nickname` longtext,
    `sender_wechat_account` longtext,
    `group_nickname` longtext,
    `to_user` longtext,
  `raw` longtext,
  PRIMARY KEY (`msg_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;

CREATE TABLE IF NOT EXISTS `t_owner` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `wechat_account` varchar(128) NOT NULL,
  `wxid` varchar(128) NOT NULL,
  `nickname` varchar(255) NOT NULL,
  `created_at` datetime(3) DEFAULT NULL,
  `updated_at` datetime(3) DEFAULT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `idx_t_owner_wxid` (`wxid`),
  UNIQUE KEY `idx_t_owner_wechat_account` (`wechat_account`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;

CREATE TABLE IF NOT EXISTS `t_bridge_setting` (
  `id` bigint unsigned NOT NULL,
  `wechat_account` varchar(128) NOT NULL DEFAULT '',
  `resolved_wxid` varchar(128) NOT NULL DEFAULT '',
  `account_directory` varchar(1024) NOT NULL DEFAULT '',
  `updated_at` datetime(3) DEFAULT NULL,
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;

INSERT INTO `t_bridge_setting` (`id`, `wechat_account`, `resolved_wxid`, `account_directory`, `updated_at`)
VALUES (1, '', '', '', CURRENT_TIMESTAMP(3))
ON DUPLICATE KEY UPDATE `id` = VALUES(`id`);

CREATE TABLE IF NOT EXISTS `t_plugin_data` (
  `user_id` longtext,
  `plugin_code` longtext,
  `data` longtext
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;

INSERT INTO `t_feature_setting`
  (`feature`, `send_time`, `schedule`, `message_threshold`, `last_run_date`, `updated_at`)
VALUES
  ('news', '08:00', 'daily', 0, '', CURRENT_TIMESTAMP(3)),
  ('weekly', '09:30', 'monday', 0, '', CURRENT_TIMESTAMP(3)),
  ('yearly', '00:00', 'yearly', 0, '', CURRENT_TIMESTAMP(3)),
  ('dormant', '09:00', 'monday', 1, '', CURRENT_TIMESTAMP(3))
ON DUPLICATE KEY UPDATE `feature` = VALUES(`feature`);
