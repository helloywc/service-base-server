-- MediaCrawlerPro 完整数据库表结构
-- 字符集统一使用 utf8mb4，兼容所有表情符号

SET NAMES utf8mb4;
SET FOREIGN_KEY_CHECKS = 0;

-- ----------------------------
-- 0. 数据库初始化
-- ----------------------------
CREATE DATABASE IF NOT EXISTS `media_operator` DEFAULT CHARACTER SET UTF8MB4 COLLATE UTF8MB4_GENERAL_CI;
USE `media_operator`;

-- ----------------------------
-- 1. 媒体主体表
-- ----------------------------
DROP TABLE IF EXISTS `media_list`;
CREATE TABLE `media_list` (
    `id`                BIGINT(20) UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '自增ID',
    `media_name`        VARCHAR(128)        NOT NULL DEFAULT ''      COMMENT '媒体/账号名称',
    `platform_name`     VARCHAR(64)         NOT NULL DEFAULT ''      COMMENT '渠道 (xhs | bili | dy | wb)',
    `platform_id`       VARCHAR(64)         NOT NULL DEFAULT ''      COMMENT 'bili / weibo / dy / ks：可存平台要求的 ID',
    `media_url`         VARCHAR(512)        NOT NULL DEFAULT ''      COMMENT 'xhs / tieba / zhihu：完整主页 URL',
    `is_followed`       TINYINT(1)          NOT NULL DEFAULT '0'     COMMENT '是否关注 (1:已关注, 0:未关注)',
    `desc`              TEXT                                         COMMENT '媒体描述/简介',
    `score`             DECIMAL(3,1)        NOT NULL DEFAULT '0.0'   COMMENT '评分',
    `add_ts`            BIGINT              NOT NULL                 COMMENT '记录添加时间戳',
    `last_modify_ts`    BIGINT              NOT NULL                 COMMENT '最后修改时间戳',
    `status`            TINYINT             NOT NULL DEFAULT '1'     COMMENT '状态(1:启用, 0:禁用)',
    `remark`            VARCHAR(512)                 DEFAULT ''      COMMENT '备注',
    `locked`            TINYINT                      DEFAULT 0       COMMENT '锁定，1:锁定 / 0: 未锁',
    `priority`          BIGINT              NOT NULL DEFAULT 0        COMMENT '优先级',
    `is_deleted`        TINYINT(1)          NOT NULL DEFAULT '0'     COMMENT '是否删除 (1:已删除, 0:未删除)',
    PRIMARY KEY (`id`),
    KEY `idx_platform_name` (`platform_name`),
    KEY `idx_is_followed` (`is_followed`)
) ENGINE=INNODB DEFAULT CHARSET=UTF8MB4 COLLATE=UTF8MB4_0900_AI_CI COMMENT='媒体列表主体表';

-- ----------------------------
-- 2. 标签字典表
-- ----------------------------
DROP TABLE IF EXISTS `media_tag`;
CREATE TABLE `media_tag` (
    `id`                INT UNSIGNED        NOT NULL AUTO_INCREMENT COMMENT '标签ID',
    `tag_name`          VARCHAR(64)         NOT NULL DEFAULT ''      COMMENT '标签名称',
    `color`             VARCHAR(16)         NOT NULL DEFAULT '#cccccc' COMMENT '色值(如#FFFFFF)',
    `create_time`       DATETIME                     DEFAULT CURRENT_TIMESTAMP,
    `is_deleted`        TINYINT(1)          NOT NULL DEFAULT '0'     COMMENT '是否删除 (1:已删除, 0:未删除)',
    PRIMARY KEY (`id`),
    UNIQUE KEY `uk_tag_name` (`tag_name`)
) ENGINE=INNODB DEFAULT CHARSET=UTF8MB4 COMMENT='媒体标签字典表';

-- ----------------------------
-- 3. 类型字典表
-- ----------------------------
DROP TABLE IF EXISTS `media_category`;
CREATE TABLE `media_category` (
    `id`                INT UNSIGNED        NOT NULL AUTO_INCREMENT COMMENT '类型ID',
    `category_name`     VARCHAR(64)         NOT NULL DEFAULT ''      COMMENT '类型名称',
    `color`             VARCHAR(16)         NOT NULL DEFAULT ''      COMMENT '色值(如#FFFFFF)',
    `create_time`       DATETIME                     DEFAULT CURRENT_TIMESTAMP,
    `is_deleted`        TINYINT(1)          NOT NULL DEFAULT '0'     COMMENT '是否删除 (1:已删除, 0:未删除)',
    PRIMARY KEY (`id`),
    UNIQUE KEY `uk_category_name` (`category_name`)
) ENGINE=INNODB DEFAULT CHARSET=UTF8MB4 COMMENT='媒体类型字典表';

-- ----------------------------
-- 4. 媒体-标签关联表 (多对多)
-- ----------------------------
DROP TABLE IF EXISTS `rel_media_tag`;
CREATE TABLE `rel_media_tag` (
    `media_id`          BIGINT(20) UNSIGNED NOT NULL COMMENT '媒体ID',
    `tag_id`            INT UNSIGNED        NOT NULL COMMENT '标签ID',
    `is_deleted`        TINYINT(1)          NOT NULL DEFAULT '0'     COMMENT '是否删除 (1:已删除, 0:未删除)',
    PRIMARY KEY (`media_id`, `tag_id`),
    KEY `idx_tag_id` (`tag_id`)
) ENGINE=INNODB DEFAULT CHARSET=UTF8MB4 COMMENT='媒体与标签关联表';

-- ----------------------------
-- 5. 媒体-类型关联表 (多对多)
-- ----------------------------
DROP TABLE IF EXISTS `rel_media_category`;
CREATE TABLE `rel_media_category` (
    `media_id`          BIGINT(20) UNSIGNED NOT NULL COMMENT '媒体ID',
    `category_id`       INT UNSIGNED        NOT NULL COMMENT '类型ID',
    `is_deleted`        TINYINT(1)          NOT NULL DEFAULT '0'     COMMENT '是否删除 (1:已删除, 0:未删除)',
    PRIMARY KEY (`media_id`, `category_id`),
    KEY `idx_category_id` (`category_id`)
) ENGINE=INNODB DEFAULT CHARSET=UTF8MB4 COMMENT='媒体与类型关联表';

-- ----------------------------
-- 6. 爬虫采集账号表 (Cookies)
-- ----------------------------
DROP TABLE IF EXISTS `crawler_cookies_account`;
CREATE TABLE `crawler_cookies_account` (
    `id`                BIGINT(20) UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '自增ID',
    `account_name`      VARCHAR(64)         NOT NULL DEFAULT ''      COMMENT '账号名称',
    `platform_name`     VARCHAR(64)         NOT NULL DEFAULT ''      COMMENT '平台名称 (xhs | dy | ks | wb | bili | tieba)',
    `cookies`           TEXT                                         COMMENT '对应自媒体平台登录成功后的cookies',
    `create_time`       DATETIME                     DEFAULT CURRENT_TIMESTAMP COMMENT '该条记录的创建时间',
    `update_time`       DATETIME                     DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '该条记录的更新时间',
    `invalid_timestamp` BIGINT(20) UNSIGNED NOT NULL DEFAULT '0'     COMMENT '账号失效时间戳',
    `status`            TINYINT             NOT NULL DEFAULT '0'     COMMENT '账号状态枚举值(0：有效，-1：无效)',
    `summary`          VARCHAR(255)                  DEFAULT NULL    COMMENT '总结',
    `keywords`         VARCHAR(255)                  DEFAULT NULL    COMMENT '关键字',
    `context`           TEXT                                         COMMENT '内容',
    `remark`            VARCHAR(512)                 DEFAULT ''      COMMENT '备注',
    `locked`            TINYINT                      DEFAULT 0       COMMENT '锁定，1:锁定 / 0: 未锁',
    `priority`          BIGINT              NOT NULL DEFAULT 0        COMMENT '优先级',
    `is_deleted`        TINYINT(1)          NOT NULL DEFAULT '0'     COMMENT '是否删除 (1:已删除, 0:未删除)',
    PRIMARY KEY (`id`),
    KEY `idx_update_time` (`update_time`)
) ENGINE=INNODB DEFAULT CHARSET=UTF8MB4 COMMENT='爬虫采集账号表（cookies）';

-- ----------------------------
-- 7. B站视频表
-- ----------------------------
DROP TABLE IF EXISTS `bilibili_video`;
CREATE TABLE `bilibili_video` (
    `id`                INT                 NOT NULL AUTO_INCREMENT COMMENT '自增ID',
    `task_id`           BIGINT              NOT NULL DEFAULT 0       COMMENT '爬虫任务id(关联crawler_tasks.id)',
    `user_id`           VARCHAR(64)                  DEFAULT NULL    COMMENT '用户ID',
    `nickname`          VARCHAR(64)                  DEFAULT NULL    COMMENT '用户昵称',
    `avatar`            VARCHAR(255)                 DEFAULT NULL    COMMENT '用户头像地址',
    `add_ts`            BIGINT              NOT NULL                 COMMENT '记录添加时间戳',
    `last_modify_ts`    BIGINT              NOT NULL                 COMMENT '记录最后修改时间戳',
    `video_id`          VARCHAR(64)         NOT NULL                 COMMENT '视频ID (aid)',
    `bvid`              VARCHAR(64)                  DEFAULT NULL    COMMENT '视频ID (bvid)',
    `video_type`        VARCHAR(16)         NOT NULL                 COMMENT '视频类型',
    `title`             VARCHAR(500)                 DEFAULT NULL    COMMENT '视频标题',
    `desc`              LONGTEXT                                     COMMENT '视频描述',
    `create_time`       BIGINT              NOT NULL                 COMMENT '视频发布时间戳',
    `liked_count`       VARCHAR(16)                  DEFAULT NULL    COMMENT '视频点赞数',
    `video_play_count`  VARCHAR(16)                  DEFAULT NULL    COMMENT '视频播放数量',
    `video_danmaku`     VARCHAR(16)                  DEFAULT NULL    COMMENT '视频弹幕数量',
    `video_comment`     VARCHAR(16)                  DEFAULT NULL    COMMENT '视频评论数量',
    `video_url`         VARCHAR(512)                 DEFAULT NULL    COMMENT '视频详情URL',
    `video_cover_url`   VARCHAR(512)                 DEFAULT NULL    COMMENT '视频封面图 URL',
    `source_keyword`    VARCHAR(255)                 DEFAULT ''      COMMENT '搜索来源关键字',
    `duration`          VARCHAR(16)                  DEFAULT NULL    COMMENT '视频时长',
    `video_download_url` VARCHAR(2048)               DEFAULT NULL    COMMENT '视频下载地址',
    `audio_download_url` VARCHAR(2048)               DEFAULT NULL    COMMENT '音频下载地址',
    `summary`          VARCHAR(255)                  DEFAULT NULL    COMMENT '总结',
    `keywords`         VARCHAR(255)                  DEFAULT NULL    COMMENT '关键字',
    `context`           TEXT                                         COMMENT '内容',
    `remark`            VARCHAR(512)                 DEFAULT ''      COMMENT '备注',
    `status`            TINYINT                      DEFAULT 0       COMMENT '状态，1:成功 / 0: 默认 / -1: 失败',
    `locked`            TINYINT                      DEFAULT 0       COMMENT '锁定，1:锁定 / 0: 未锁',
    `priority`          BIGINT              NOT NULL DEFAULT 0        COMMENT '优先级',
    `is_deleted`        TINYINT(1)          NOT NULL DEFAULT '0'     COMMENT '是否删除 (1:已删除, 0:未删除)',
    PRIMARY KEY (`id`),
    UNIQUE KEY `uk_bilibili_video_id` (`video_id`),
    KEY `idx_video_id` (`video_id`),
    KEY `idx_create_time` (`create_time`)
) ENGINE=INNODB DEFAULT CHARSET=UTF8MB4 COLLATE=UTF8MB4_0900_AI_CI COMMENT='B站视频';

-- ----------------------------
-- 8. B站评论表
-- ----------------------------
DROP TABLE IF EXISTS `bilibili_video_comment`;
CREATE TABLE `bilibili_video_comment` (
    `id`                INT                 NOT NULL AUTO_INCREMENT COMMENT '自增ID',
    `task_id`           BIGINT              NOT NULL DEFAULT 0       COMMENT '爬虫任务id(关联crawler_tasks.id)',
    `user_id`           VARCHAR(64)                  DEFAULT NULL    COMMENT '用户ID',
    `nickname`          VARCHAR(64)                  DEFAULT NULL    COMMENT '用户昵称',
    `avatar`            VARCHAR(255)                 DEFAULT NULL    COMMENT '用户头像地址',
    `add_ts`            BIGINT              NOT NULL                 COMMENT '记录添加时间戳',
    `last_modify_ts`    BIGINT              NOT NULL                 COMMENT '记录最后修改时间戳',
    `comment_id`        VARCHAR(64)         NOT NULL                 COMMENT '评论ID',
    `video_id`          VARCHAR(64)         NOT NULL                 COMMENT '视频ID',
    `content`           LONGTEXT                                     COMMENT '评论内容',
    `create_time`       BIGINT              NOT NULL                 COMMENT '评论时间戳',
    `sub_comment_count` VARCHAR(16)         NOT NULL                 COMMENT '评论回复数',
    `parent_comment_id` VARCHAR(64)                  DEFAULT NULL    COMMENT '父评论ID',
    `like_count`        VARCHAR(255)        NOT NULL DEFAULT '0'     COMMENT '点赞数',
    `video_download_url` VARCHAR(2048)                DEFAULT NULL    COMMENT '视频下载地址',
    `audio_download_url` VARCHAR(2048)                DEFAULT NULL    COMMENT '音频下载地址',
    `summary`          VARCHAR(255)                  DEFAULT NULL    COMMENT '总结',
    `keywords`         VARCHAR(255)                  DEFAULT NULL    COMMENT '关键字',
    `context`           TEXT                                         COMMENT '内容',
    `remark`            VARCHAR(512)                 DEFAULT ''      COMMENT '备注',
    `status`            TINYINT                      DEFAULT 0       COMMENT '状态，1:成功 / 0: 默认 / -1: 失败',
    `locked`            TINYINT                      DEFAULT 0       COMMENT '锁定，1:锁定 / 0: 未锁',
    `priority`          BIGINT              NOT NULL DEFAULT 0        COMMENT '优先级',
    `is_deleted`        TINYINT(1)          NOT NULL DEFAULT '0'     COMMENT '是否删除 (1:已删除, 0:未删除)',
    PRIMARY KEY (`id`),
    UNIQUE KEY `uk_bilibili_comment_id` (`comment_id`),
    KEY `idx_comment_id` (`comment_id`),
    KEY `idx_video_id` (`video_id`)
) ENGINE=INNODB DEFAULT CHARSET=UTF8MB4 COLLATE=UTF8MB4_0900_AI_CI COMMENT='B 站视频评论';

-- ----------------------------
-- 9. B站UP主信息表
-- ----------------------------
DROP TABLE IF EXISTS `bilibili_up_info`;
CREATE TABLE `bilibili_up_info` (
    `id`                INT                 NOT NULL AUTO_INCREMENT COMMENT '自增ID',
    `task_id`           BIGINT              NOT NULL DEFAULT 0       COMMENT '爬虫任务id(关联crawler_tasks.id)',
    `user_id`           VARCHAR(64)                  DEFAULT NULL    COMMENT '用户ID',
    `nickname`          VARCHAR(64)                  DEFAULT NULL    COMMENT '用户昵称',
    `avatar`            VARCHAR(255)                 DEFAULT NULL    COMMENT '用户头像地址',
    `follower_count`    BIGINT                       DEFAULT NULL    COMMENT '粉丝数',
    `following_count`   BIGINT                       DEFAULT NULL    COMMENT '关注数',
    `content_count`     BIGINT                       DEFAULT NULL    COMMENT '作品数',
    `description`       LONGTEXT                                     COMMENT '用户描述',
    `add_ts`            BIGINT              NOT NULL                 COMMENT '记录添加时间戳',
    `last_modify_ts`    BIGINT              NOT NULL                 COMMENT '记录最后修改时间戳',
    `video_download_url` VARCHAR(2048)                DEFAULT NULL    COMMENT '视频下载地址',
    `audio_download_url` VARCHAR(2048)                DEFAULT NULL    COMMENT '音频下载地址',
    `summary`          VARCHAR(255)                  DEFAULT NULL    COMMENT '总结',
    `keywords`         VARCHAR(255)                  DEFAULT NULL    COMMENT '关键字',
    `context`           TEXT                                         COMMENT '内容',
    `remark`            VARCHAR(512)                 DEFAULT ''      COMMENT '备注',
    `status`            TINYINT                      DEFAULT 0       COMMENT '状态，1:成功 / 0: 默认 / -1: 失败',
    `locked`            TINYINT                      DEFAULT 0       COMMENT '锁定，1:锁定 / 0: 未锁',
    `priority`          BIGINT              NOT NULL DEFAULT 0        COMMENT '优先级',
    `is_deleted`        TINYINT(1)          NOT NULL DEFAULT '0'     COMMENT '是否删除 (1:已删除, 0:未删除)',
    PRIMARY KEY (`id`),
    UNIQUE KEY `uk_bilibili_up_user_id` (`user_id`),
    KEY `idx_user_id` (`user_id`)
) ENGINE=INNODB DEFAULT CHARSET=UTF8MB4 COLLATE=UTF8MB4_0900_AI_CI COMMENT='B站UP主信息';

-- ----------------------------
-- 10. 抖音视频表
-- ----------------------------
DROP TABLE IF EXISTS `douyin_aweme`;
CREATE TABLE `douyin_aweme` (
    `id`                 INT                 NOT NULL AUTO_INCREMENT COMMENT '自增ID',
    `task_id`            BIGINT              NOT NULL DEFAULT 0       COMMENT '爬虫任务id(关联crawler_tasks.id)',
    `user_id`            VARCHAR(64)                  DEFAULT NULL    COMMENT '用户ID',
    `sec_uid`            VARCHAR(128)                 DEFAULT NULL    COMMENT '用户sec_uid',
    `short_user_id`      VARCHAR(64)                  DEFAULT NULL    COMMENT '用户短ID',
    `user_unique_id`     VARCHAR(64)                  DEFAULT NULL    COMMENT '用户唯一ID',
    `nickname`           VARCHAR(64)                  DEFAULT NULL    COMMENT '用户昵称',
    `avatar`             VARCHAR(255)                 DEFAULT NULL    COMMENT '用户头像地址',
    `user_signature`     VARCHAR(500)                 DEFAULT NULL    COMMENT '用户签名',
    `ip_location`        VARCHAR(255)                 DEFAULT NULL    COMMENT '评论时的IP地址',
    `add_ts`             BIGINT              NOT NULL                 COMMENT '记录添加时间戳',
    `last_modify_ts`     BIGINT              NOT NULL                 COMMENT '记录最后修改时间戳',
    `aweme_id`           VARCHAR(64)         NOT NULL                 COMMENT '视频ID',
    `aweme_type`         VARCHAR(16)         NOT NULL                 COMMENT '视频类型',
    `title`              VARCHAR(1024)                DEFAULT NULL    COMMENT '视频标题',
    `desc`               LONGTEXT                                     COMMENT '视频描述',
    `create_time`        BIGINT              NOT NULL                 COMMENT '视频发布时间戳',
    `liked_count`        VARCHAR(16)                  DEFAULT NULL    COMMENT '视频点赞数',
    `comment_count`      VARCHAR(16)                  DEFAULT NULL    COMMENT '视频评论数',
    `share_count`        VARCHAR(16)                  DEFAULT NULL    COMMENT '视频分享数',
    `collected_count`    VARCHAR(16)                  DEFAULT NULL    COMMENT '视频收藏数',
    `aweme_url`          VARCHAR(255)                 DEFAULT NULL    COMMENT '视频详情页URL',
    `cover_url`          VARCHAR(500)                 DEFAULT NULL    COMMENT '视频封面图URL',
    `video_download_url` VARCHAR(2048)                DEFAULT NULL    COMMENT '视频下载地址',
    `source_keyword`     VARCHAR(255)                 DEFAULT ''      COMMENT '搜索来源关键字',
    `is_ai_generated`    TINYINT(1)          NOT NULL DEFAULT '0'     COMMENT '作者是否声明视频为AI生成',
    `audio_download_url` VARCHAR(2048)                 DEFAULT NULL    COMMENT '音频下载地址',
    `summary`          VARCHAR(255)                  DEFAULT NULL    COMMENT '总结',
    `keywords`         VARCHAR(255)                  DEFAULT NULL    COMMENT '关键字',
    `context`            TEXT                                         COMMENT '内容',
    `remark`             VARCHAR(512)                 DEFAULT ''      COMMENT '备注',
    `status`             TINYINT                      DEFAULT 0       COMMENT '状态，1:成功 / 0: 默认 / -1: 失败',
    `locked`             TINYINT                      DEFAULT 0       COMMENT '锁定，1:锁定 / 0: 未锁',
    `priority`          BIGINT              NOT NULL DEFAULT 0        COMMENT '优先级',
    `is_deleted`         TINYINT(1)          NOT NULL DEFAULT '0'     COMMENT '是否删除 (1:已删除, 0:未删除)',
    PRIMARY KEY (`id`),
    UNIQUE KEY `uk_douyin_aweme_id` (`aweme_id`),
    KEY `idx_aweme_id` (`aweme_id`),
    KEY `idx_create_time` (`create_time`)
) ENGINE=INNODB DEFAULT CHARSET=UTF8MB4 COLLATE=UTF8MB4_0900_AI_CI COMMENT='抖音视频';

-- ----------------------------
-- 11. 抖音评论表
-- ----------------------------
DROP TABLE IF EXISTS `douyin_aweme_comment`;
CREATE TABLE `douyin_aweme_comment` (
    `id`                INT                 NOT NULL AUTO_INCREMENT COMMENT '自增ID',
    `task_id`           BIGINT              NOT NULL DEFAULT 0       COMMENT '爬虫任务id(关联crawler_tasks.id)',
    `user_id`           VARCHAR(64)                  DEFAULT NULL    COMMENT '用户ID',
    `sec_uid`           VARCHAR(128)                 DEFAULT NULL    COMMENT '用户sec_uid',
    `short_user_id`     VARCHAR(64)                  DEFAULT NULL    COMMENT '用户短ID',
    `user_unique_id`    VARCHAR(64)                  DEFAULT NULL    COMMENT '用户唯一ID',
    `nickname`          VARCHAR(64)                  DEFAULT NULL    COMMENT '用户昵称',
    `avatar`            VARCHAR(255)                 DEFAULT NULL    COMMENT '用户头像地址',
    `user_signature`    VARCHAR(500)                 DEFAULT NULL    COMMENT '用户签名',
    `ip_location`       VARCHAR(255)                 DEFAULT NULL    COMMENT '评论时的IP地址',
    `add_ts`            BIGINT              NOT NULL                 COMMENT '记录添加时间戳',
    `last_modify_ts`    BIGINT              NOT NULL                 COMMENT '记录最后修改时间戳',
    `comment_id`        VARCHAR(64)         NOT NULL                 COMMENT '评论ID',
    `aweme_id`          VARCHAR(64)         NOT NULL                 COMMENT '视频ID',
    `content`           LONGTEXT                                     COMMENT '评论内容',
    `create_time`       BIGINT              NOT NULL                 COMMENT '评论时间戳',
    `sub_comment_count` VARCHAR(16)         NOT NULL                 COMMENT '评论回复数',
    `parent_comment_id` VARCHAR(64)                  DEFAULT NULL    COMMENT '父评论ID',
    `like_count`        VARCHAR(255)        NOT NULL DEFAULT '0'     COMMENT '点赞数',
    `pictures`          VARCHAR(500)        NOT NULL DEFAULT ''      COMMENT '评论图片列表',
    `reply_to_reply_id` VARCHAR(64)                  DEFAULT NULL    COMMENT '目标评论ID',
    `video_download_url` VARCHAR(2048)                DEFAULT NULL    COMMENT '视频下载地址',
    `audio_download_url` VARCHAR(2048)                DEFAULT NULL    COMMENT '音频下载地址',
    `summary`          VARCHAR(255)                  DEFAULT NULL    COMMENT '总结',
    `keywords`         VARCHAR(255)                  DEFAULT NULL    COMMENT '关键字',
    `context`           TEXT                                         COMMENT '内容',
    `remark`            VARCHAR(512)                 DEFAULT ''      COMMENT '备注',
    `status`            TINYINT                      DEFAULT 0       COMMENT '状态，1:成功 / 0: 默认 / -1: 失败',
    `locked`            TINYINT                      DEFAULT 0       COMMENT '锁定，1:锁定 / 0: 未锁',
    `priority`          BIGINT              NOT NULL DEFAULT 0        COMMENT '优先级',
    `is_deleted`        TINYINT(1)          NOT NULL DEFAULT '0'     COMMENT '是否删除 (1:已删除, 0:未删除)',
    PRIMARY KEY (`id`),
    UNIQUE KEY `uk_douyin_comment_id` (`comment_id`),
    KEY `idx_comment_id` (`comment_id`),
    KEY `idx_aweme_id` (`aweme_id`)
) ENGINE=INNODB DEFAULT CHARSET=UTF8MB4 COLLATE=UTF8MB4_0900_AI_CI COMMENT='抖音视频评论';

-- ----------------------------
-- 12. 抖音博主表
-- ----------------------------
DROP TABLE IF EXISTS `dy_creator`;
CREATE TABLE `dy_creator` (
    `id`                INT                 NOT NULL AUTO_INCREMENT COMMENT '自增ID',
    `task_id`           BIGINT              NOT NULL DEFAULT 0       COMMENT '爬虫任务id(关联crawler_tasks.id)',
    `user_id`           VARCHAR(128)        NOT NULL                 COMMENT '用户ID',
    `nickname`          VARCHAR(64)                  DEFAULT NULL    COMMENT '用户昵称',
    `avatar`            VARCHAR(255)                 DEFAULT NULL    COMMENT '用户头像地址',
    `ip_location`       VARCHAR(255)                 DEFAULT NULL    COMMENT '评论时的IP地址',
    `add_ts`            BIGINT              NOT NULL                 COMMENT '记录添加时间戳',
    `last_modify_ts`    BIGINT              NOT NULL                 COMMENT '记录最后修改时间戳',
    `desc`              LONGTEXT                                     COMMENT '用户描述',
    `gender`            VARCHAR(2)                   DEFAULT NULL    COMMENT '性别',
    `follows`           VARCHAR(16)                  DEFAULT NULL    COMMENT '关注数',
    `fans`              VARCHAR(16)                  DEFAULT NULL    COMMENT '粉丝数',
    `interaction`       VARCHAR(16)                  DEFAULT NULL    COMMENT '获赞数',
    `videos_count`      VARCHAR(16)                  DEFAULT NULL    COMMENT '作品数',
    `video_download_url` VARCHAR(2048)                DEFAULT NULL    COMMENT '视频下载地址',
    `audio_download_url` VARCHAR(2048)                DEFAULT NULL    COMMENT '音频下载地址',
    `summary`          VARCHAR(255)                  DEFAULT NULL    COMMENT '总结',
    `keywords`         VARCHAR(255)                  DEFAULT NULL    COMMENT '关键字',
    `context`           TEXT                                         COMMENT '内容',
    `remark`            VARCHAR(512)                 DEFAULT ''      COMMENT '备注',
    `status`            TINYINT                      DEFAULT 0       COMMENT '状态，1:成功 / 0: 默认 / -1: 失败',
    `locked`            TINYINT                      DEFAULT 0       COMMENT '锁定，1:锁定 / 0: 未锁',
    `priority`          BIGINT              NOT NULL DEFAULT 0        COMMENT '优先级',
    `is_deleted`        TINYINT(1)          NOT NULL DEFAULT '0'     COMMENT '是否删除 (1:已删除, 0:未删除)',
    PRIMARY KEY (`id`),
    UNIQUE KEY `uk_dy_creator_user_id` (`user_id`)
) ENGINE=INNODB DEFAULT CHARSET=UTF8MB4 COLLATE=UTF8MB4_0900_AI_CI COMMENT='抖音博主信息';

-- ----------------------------
-- 13. 快手视频表
-- ----------------------------
DROP TABLE IF EXISTS `kuaishou_video`;
CREATE TABLE `kuaishou_video` (
    `id`                INT                 NOT NULL AUTO_INCREMENT COMMENT '自增ID',
    `task_id`           BIGINT              NOT NULL DEFAULT 0       COMMENT '爬虫任务id(关联crawler_tasks.id)',
    `user_id`           VARCHAR(64)                  DEFAULT NULL    COMMENT '用户ID',
    `nickname`          VARCHAR(64)                  DEFAULT NULL    COMMENT '用户昵称',
    `avatar`            VARCHAR(255)                 DEFAULT NULL    COMMENT '用户头像地址',
    `add_ts`            BIGINT              NOT NULL                 COMMENT '记录添加时间戳',
    `last_modify_ts`    BIGINT              NOT NULL                 COMMENT '记录最后修改时间戳',
    `video_id`          VARCHAR(64)         NOT NULL                 COMMENT '视频ID',
    `video_type`        VARCHAR(16)         NOT NULL                 COMMENT '视频类型',
    `title`             VARCHAR(500)                 DEFAULT NULL    COMMENT '视频标题',
    `desc`              LONGTEXT                                     COMMENT '视频描述',
    `create_time`       BIGINT              NOT NULL                 COMMENT '视频发布时间戳',
    `liked_count`       VARCHAR(16)                  DEFAULT NULL    COMMENT '视频点赞数',
    `viewd_count`       VARCHAR(16)                  DEFAULT NULL    COMMENT '视频浏览数量',
    `video_url`         VARCHAR(512)                 DEFAULT NULL    COMMENT '视频详情URL',
    `video_cover_url`   TEXT                         DEFAULT NULL    COMMENT '视频封面图 URL',
    `video_play_url`    TEXT                         DEFAULT NULL    COMMENT '视频播放 URL',
    `source_keyword`    VARCHAR(255)                 DEFAULT ''      COMMENT '搜索来源关键字',
    `video_download_url` VARCHAR(2048)                DEFAULT NULL    COMMENT '视频下载地址',
    `audio_download_url` VARCHAR(2048)                DEFAULT NULL    COMMENT '音频下载地址',
    `summary`          VARCHAR(255)                  DEFAULT NULL    COMMENT '总结',
    `keywords`         VARCHAR(255)                  DEFAULT NULL    COMMENT '关键字',
    `context`           TEXT                                         COMMENT '内容',
    `remark`            VARCHAR(512)                 DEFAULT ''      COMMENT '备注',
    `status`            TINYINT                      DEFAULT 0       COMMENT '状态，1:成功 / 0: 默认 / -1: 失败',
    `locked`            TINYINT                      DEFAULT 0       COMMENT '锁定，1:锁定 / 0: 未锁',
    `priority`          BIGINT              NOT NULL DEFAULT 0        COMMENT '优先级',
    `is_deleted`        TINYINT(1)          NOT NULL DEFAULT '0'     COMMENT '是否删除 (1:已删除, 0:未删除)',
    PRIMARY KEY (`id`),
    UNIQUE KEY `uk_kuaishou_video_id` (`video_id`),
    KEY `idx_video_id` (`video_id`),
    KEY `idx_create_time` (`create_time`)
) ENGINE=INNODB DEFAULT CHARSET=UTF8MB4 COLLATE=UTF8MB4_0900_AI_CI COMMENT='快手视频';

-- ----------------------------
-- 14. 快手评论表
-- ----------------------------
DROP TABLE IF EXISTS `kuaishou_video_comment`;
CREATE TABLE `kuaishou_video_comment` (
    `id`                INT                 NOT NULL AUTO_INCREMENT COMMENT '自增ID',
    `task_id`           BIGINT              NOT NULL DEFAULT 0       COMMENT '爬虫任务id(关联crawler_tasks.id)',
    `user_id`           VARCHAR(64)                  DEFAULT NULL    COMMENT '用户ID',
    `nickname`          VARCHAR(64)                  DEFAULT NULL    COMMENT '用户昵称',
    `avatar`            VARCHAR(255)                 DEFAULT NULL    COMMENT '用户头像地址',
    `add_ts`            BIGINT              NOT NULL                 COMMENT '记录添加时间戳',
    `last_modify_ts`    BIGINT              NOT NULL                 COMMENT '记录最后修改时间戳',
    `comment_id`        VARCHAR(64)         NOT NULL                 COMMENT '评论ID',
    `parent_comment_id` VARCHAR(64)                  DEFAULT NULL    COMMENT '父评论ID',
    `video_id`          VARCHAR(64)         NOT NULL                 COMMENT '视频ID',
    `content`           LONGTEXT                                     COMMENT '评论内容',
    `create_time`       BIGINT              NOT NULL                 COMMENT '评论时间戳',
    `sub_comment_count` VARCHAR(16)         NOT NULL                 COMMENT '评论回复数',
    `like_count`        VARCHAR(255)        NOT NULL DEFAULT '0'     COMMENT '点赞数',
    `video_download_url` VARCHAR(2048)                DEFAULT NULL    COMMENT '视频下载地址',
    `audio_download_url` VARCHAR(2048)                DEFAULT NULL    COMMENT '音频下载地址',
    `summary`          VARCHAR(255)                  DEFAULT NULL    COMMENT '总结',
    `keywords`         VARCHAR(255)                  DEFAULT NULL    COMMENT '关键字',
    `context`           TEXT                                         COMMENT '内容',
    `remark`            VARCHAR(512)                 DEFAULT ''      COMMENT '备注',
    `status`            TINYINT                      DEFAULT 0       COMMENT '状态，1:成功 / 0: 默认 / -1: 失败',
    `locked`            TINYINT                      DEFAULT 0       COMMENT '锁定，1:锁定 / 0: 未锁',
    `priority`          BIGINT              NOT NULL DEFAULT 0        COMMENT '优先级',
    `is_deleted`        TINYINT(1)          NOT NULL DEFAULT '0'     COMMENT '是否删除 (1:已删除, 0:未删除)',
    PRIMARY KEY (`id`),
    UNIQUE KEY `uk_kuaishou_comment_id` (`comment_id`),
    KEY `idx_comment_id` (`comment_id`),
    KEY `idx_video_id` (`video_id`)
) ENGINE=INNODB DEFAULT CHARSET=UTF8MB4 COLLATE=UTF8MB4_0900_AI_CI COMMENT='快手视频评论';

-- ----------------------------
-- 15. 快手博主表
-- ----------------------------
DROP TABLE IF EXISTS `kuaishou_creator`;
CREATE TABLE `kuaishou_creator` (
    `id`                INT                 NOT NULL AUTO_INCREMENT COMMENT '自增ID',
    `task_id`           BIGINT              NOT NULL DEFAULT 0       COMMENT '爬虫任务id(关联crawler_tasks.id)',
    `user_id`           VARCHAR(64)         NOT NULL                 COMMENT '用户ID',
    `nickname`          VARCHAR(64)                  DEFAULT NULL    COMMENT '用户昵称',
    `avatar`            VARCHAR(255)                 DEFAULT NULL    COMMENT '用户头像地址',
    `ip_location`       VARCHAR(255)                 DEFAULT NULL    COMMENT '评论时的IP地址',
    `add_ts`            BIGINT              NOT NULL                 COMMENT '记录添加时间戳',
    `last_modify_ts`    BIGINT              NOT NULL                 COMMENT '记录最后修改时间戳',
    `desc`              LONGTEXT                                     COMMENT '用户描述',
    `gender`            VARCHAR(2)                   DEFAULT NULL    COMMENT '性别',
    `follows`           VARCHAR(16)                  DEFAULT NULL    COMMENT '关注数',
    `fans`              VARCHAR(16)                  DEFAULT NULL    COMMENT '粉丝数',
    `videos_count`      VARCHAR(16)                  DEFAULT NULL    COMMENT '作品数',
    `video_download_url` VARCHAR(2048)                DEFAULT NULL    COMMENT '视频下载地址',
    `audio_download_url` VARCHAR(2048)                DEFAULT NULL    COMMENT '音频下载地址',
    `summary`          VARCHAR(255)                  DEFAULT NULL    COMMENT '总结',
    `keywords`         VARCHAR(255)                  DEFAULT NULL    COMMENT '关键字',
    `context`           TEXT                                         COMMENT '内容',
    `remark`            VARCHAR(512)                 DEFAULT ''      COMMENT '备注',
    `status`            TINYINT                      DEFAULT 0       COMMENT '状态，1:成功 / 0: 默认 / -1: 失败',
    `locked`            TINYINT                      DEFAULT 0       COMMENT '锁定，1:锁定 / 0: 未锁',
    `priority`          BIGINT              NOT NULL DEFAULT 0        COMMENT '优先级',
    `is_deleted`        TINYINT(1)          NOT NULL DEFAULT '0'     COMMENT '是否删除 (1:已删除, 0:未删除)',
    PRIMARY KEY (`id`),
    UNIQUE KEY `uk_kuaishou_creator_user_id` (`user_id`)
) ENGINE=INNODB DEFAULT CHARSET=UTF8MB4 COLLATE=UTF8MB4_0900_AI_CI COMMENT='快手博主';

-- ----------------------------
-- 16. 微博帖子表
-- ----------------------------
DROP TABLE IF EXISTS `weibo_note`;
CREATE TABLE `weibo_note` (
    `id`                INT                 NOT NULL AUTO_INCREMENT COMMENT '自增ID',
    `task_id`           BIGINT              NOT NULL DEFAULT 0       COMMENT '爬虫任务id(关联crawler_tasks.id)',
    `user_id`           VARCHAR(64)                  DEFAULT NULL    COMMENT '用户ID',
    `nickname`          VARCHAR(64)                  DEFAULT NULL    COMMENT '用户昵称',
    `avatar`            VARCHAR(255)                 DEFAULT NULL    COMMENT '用户头像地址',
    `gender`            VARCHAR(12)                  DEFAULT NULL    COMMENT '用户性别',
    `profile_url`       VARCHAR(255)                 DEFAULT NULL    COMMENT '用户主页地址',
    `ip_location`       VARCHAR(32)                  DEFAULT '发布微博的地理信息',
    `add_ts`            BIGINT              NOT NULL                 COMMENT '记录添加时间戳',
    `last_modify_ts`    BIGINT              NOT NULL                 COMMENT '记录最后修改时间戳',
    `note_id`           VARCHAR(64)         NOT NULL                 COMMENT '帖子ID',
    `content`           LONGTEXT                                     COMMENT '帖子正文内容',
    `create_time`       BIGINT              NOT NULL                 COMMENT '帖子发布时间戳',
    `create_date_time`  VARCHAR(32)         NOT NULL                 COMMENT '帖子发布日期时间',
    `liked_count`       VARCHAR(16)                  DEFAULT NULL    COMMENT '帖子点赞数',
    `comments_count`    VARCHAR(16)                  DEFAULT NULL    COMMENT '帖子评论数量',
    `shared_count`      VARCHAR(16)                  DEFAULT NULL    COMMENT '帖子转发数量',
    `note_url`          VARCHAR(512)                 DEFAULT NULL    COMMENT '帖子详情URL',
    `image_list`        LONGTEXT                                     COMMENT '封面图片列表',
    `video_url`         LONGTEXT                                     COMMENT '视频地址',
    `source_keyword`    VARCHAR(255)                 DEFAULT ''      COMMENT '搜索来源关键字',
    `video_download_url` VARCHAR(2048)                DEFAULT NULL    COMMENT '视频下载地址',
    `audio_download_url` VARCHAR(2048)                DEFAULT NULL    COMMENT '音频下载地址',
    `summary`          VARCHAR(255)                  DEFAULT NULL    COMMENT '总结',
    `keywords`         VARCHAR(255)                  DEFAULT NULL    COMMENT '关键字',
    `context`           TEXT                                         COMMENT '内容',
    `remark`            VARCHAR(512)                 DEFAULT ''      COMMENT '备注',
    `status`            TINYINT                      DEFAULT 0       COMMENT '状态，1:成功 / 0: 默认 / -1: 失败',
    `locked`            TINYINT                      DEFAULT 0       COMMENT '锁定，1:锁定 / 0: 未锁',
    `priority`          BIGINT              NOT NULL DEFAULT 0        COMMENT '优先级',
    `is_deleted`        TINYINT(1)          NOT NULL DEFAULT '0'     COMMENT '是否删除 (1:已删除, 0:未删除)',
    PRIMARY KEY (`id`),
    UNIQUE KEY `uk_weibo_note_id` (`note_id`),
    KEY `idx_note_id` (`note_id`),
    KEY `idx_create_time` (`create_time`),
    KEY `idx_create_date_time` (`create_date_time`)
) ENGINE=INNODB DEFAULT CHARSET=UTF8MB4 COLLATE=UTF8MB4_0900_AI_CI COMMENT='微博帖子';

-- ----------------------------
-- 17. 微博评论表
-- ----------------------------
DROP TABLE IF EXISTS `weibo_note_comment`;
CREATE TABLE `weibo_note_comment` (
    `id`                INT                 NOT NULL AUTO_INCREMENT COMMENT '自增ID',
    `task_id`           BIGINT              NOT NULL DEFAULT 0       COMMENT '爬虫任务id(关联crawler_tasks.id)',
    `user_id`           VARCHAR(64)                  DEFAULT NULL    COMMENT '用户ID',
    `nickname`          VARCHAR(64)                  DEFAULT NULL    COMMENT '用户昵称',
    `avatar`            VARCHAR(255)                 DEFAULT NULL    COMMENT '用户头像地址',
    `gender`            VARCHAR(12)                  DEFAULT NULL    COMMENT '用户性别',
    `profile_url`       VARCHAR(255)                 DEFAULT NULL    COMMENT '用户主页地址',
    `ip_location`       VARCHAR(32)                  DEFAULT '发布微博的地理信息',
    `add_ts`            BIGINT              NOT NULL                 COMMENT '记录添加时间戳',
    `last_modify_ts`    BIGINT              NOT NULL                 COMMENT '记录最后修改时间戳',
    `comment_id`        VARCHAR(64)         NOT NULL                 COMMENT '评论ID',
    `note_id`           VARCHAR(64)         NOT NULL                 COMMENT '帖子ID',
    `content`           LONGTEXT                                     COMMENT '评论内容',
    `create_time`       BIGINT              NOT NULL                 COMMENT '评论时间戳',
    `create_date_time`  VARCHAR(32)         NOT NULL                 COMMENT '评论日期时间',
    `sub_comment_count` VARCHAR(16)         NOT NULL                 COMMENT '评论回复数',
    `parent_comment_id` VARCHAR(64)                  DEFAULT NULL    COMMENT '父评论ID',
    `like_count`        VARCHAR(255)        NOT NULL DEFAULT '0'     COMMENT '点赞数',
    `video_download_url` VARCHAR(2048)                DEFAULT NULL    COMMENT '视频下载地址',
    `audio_download_url` VARCHAR(2048)                DEFAULT NULL    COMMENT '音频下载地址',
    `summary`          VARCHAR(255)                  DEFAULT NULL    COMMENT '总结',
    `keywords`         VARCHAR(255)                  DEFAULT NULL    COMMENT '关键字',
    `context`           TEXT                                         COMMENT '内容',
    `remark`            VARCHAR(512)                 DEFAULT ''      COMMENT '备注',
    `status`            TINYINT                      DEFAULT 0       COMMENT '状态，1:成功 / 0: 默认 / -1: 失败',
    `locked`            TINYINT                      DEFAULT 0       COMMENT '锁定，1:锁定 / 0: 未锁',
    `priority`          BIGINT              NOT NULL DEFAULT 0        COMMENT '优先级',
    `is_deleted`        TINYINT(1)          NOT NULL DEFAULT '0'     COMMENT '是否删除 (1:已删除, 0:未删除)',
    PRIMARY KEY (`id`),
    UNIQUE KEY `uk_weibo_comment_id` (`comment_id`),
    KEY `idx_comment_id` (`comment_id`),
    KEY `idx_note_id` (`note_id`),
    KEY `idx_create_date_time` (`create_date_time`)
) ENGINE=INNODB DEFAULT CHARSET=UTF8MB4 COLLATE=UTF8MB4_0900_AI_CI COMMENT='微博帖子评论';

-- ----------------------------
-- 18. 小红书博主表
-- ----------------------------
DROP TABLE IF EXISTS `xhs_creator`;
CREATE TABLE `xhs_creator` (
    `id`                INT                 NOT NULL AUTO_INCREMENT COMMENT '自增ID',
    `task_id`           BIGINT              NOT NULL DEFAULT 0       COMMENT '爬虫任务id(关联crawler_tasks.id)',
    `user_id`           VARCHAR(64)         NOT NULL                 COMMENT '用户ID',
    `nickname`          VARCHAR(64)                  DEFAULT NULL    COMMENT '用户昵称',
    `avatar`            VARCHAR(255)                 DEFAULT NULL    COMMENT '用户头像地址',
    `ip_location`       VARCHAR(255)                 DEFAULT NULL    COMMENT '评论时的IP地址',
    `add_ts`            BIGINT              NOT NULL                 COMMENT '记录添加时间戳',
    `last_modify_ts`    BIGINT              NOT NULL                 COMMENT '记录最后修改时间戳',
    `desc`              LONGTEXT                                     COMMENT '用户描述',
    `gender`            VARCHAR(2)                   DEFAULT NULL    COMMENT '性别',
    `follows`           VARCHAR(16)                  DEFAULT NULL    COMMENT '关注数',
    `fans`              VARCHAR(16)                  DEFAULT NULL    COMMENT '粉丝数',
    `interaction`       VARCHAR(16)                  DEFAULT NULL    COMMENT '获赞和收藏数',
    `tag_list`          LONGTEXT                                     COMMENT '标签列表',
    `video_download_url` VARCHAR(2048)                DEFAULT NULL    COMMENT '视频下载地址',
    `audio_download_url` VARCHAR(2048)                DEFAULT NULL    COMMENT '音频下载地址',
    `summary`          VARCHAR(255)                  DEFAULT NULL    COMMENT '总结',
    `keywords`         VARCHAR(255)                  DEFAULT NULL    COMMENT '关键字',
    `context`           TEXT                                         COMMENT '内容',
    `remark`            VARCHAR(512)                 DEFAULT ''      COMMENT '备注',
    `status`            TINYINT                      DEFAULT 0       COMMENT '状态，1:成功 / 0: 默认 / -1: 失败',
    `locked`            TINYINT                      DEFAULT 0       COMMENT '锁定，1:锁定 / 0: 未锁',
    `priority`          BIGINT              NOT NULL DEFAULT 0        COMMENT '优先级',
    `is_deleted`        TINYINT(1)          NOT NULL DEFAULT '0'     COMMENT '是否删除 (1:已删除, 0:未删除)',
    PRIMARY KEY (`id`),
    UNIQUE KEY `uk_xhs_creator_user_id` (`user_id`)
) ENGINE=INNODB DEFAULT CHARSET=UTF8MB4 COLLATE=UTF8MB4_0900_AI_CI COMMENT='小红书博主';

-- ----------------------------
-- 19. 小红书笔记表
-- ----------------------------
DROP TABLE IF EXISTS `xhs_note`;
CREATE TABLE `xhs_note` (
    `id`                INT                 NOT NULL AUTO_INCREMENT COMMENT '自增ID',
    `task_id`           BIGINT              NOT NULL DEFAULT 0       COMMENT '爬虫任务id(关联crawler_tasks.id)',
    `user_id`           VARCHAR(64)         NOT NULL                 COMMENT '用户ID',
    `nickname`          VARCHAR(64)                  DEFAULT NULL    COMMENT '用户昵称',
    `avatar`            VARCHAR(255)                 DEFAULT NULL    COMMENT '用户头像地址',
    `ip_location`       VARCHAR(255)                 DEFAULT NULL    COMMENT '评论时的IP地址',
    `add_ts`            BIGINT              NOT NULL                 COMMENT '记录添加时间戳',
    `last_modify_ts`    BIGINT              NOT NULL                 COMMENT '记录最后修改时间戳',
    `note_id`           VARCHAR(64)         NOT NULL                 COMMENT '笔记ID',
    `type`              VARCHAR(16)                  DEFAULT NULL    COMMENT '笔记类型(normal | video)',
    `title`             VARCHAR(255)                 DEFAULT NULL    COMMENT '笔记标题',
    `desc`              LONGTEXT                                     COMMENT '笔记描述',
    `video_url`         LONGTEXT                                     COMMENT '视频地址',
    `time`              BIGINT              NOT NULL                 COMMENT '笔记发布时间戳',
    `last_update_time`  BIGINT              NOT NULL                 COMMENT '笔记最后更新时间戳',
    `liked_count`       VARCHAR(16)                  DEFAULT NULL    COMMENT '笔记点赞数',
    `collected_count`   VARCHAR(16)                  DEFAULT NULL    COMMENT '笔记收藏数',
    `comment_count`     VARCHAR(16)                  DEFAULT NULL    COMMENT '笔记评论数',
    `share_count`       VARCHAR(16)                  DEFAULT NULL    COMMENT '笔记分享数',
    `image_list`        LONGTEXT                                     COMMENT '笔记封面图片列表',
    `tag_list`          LONGTEXT                                     COMMENT '标签列表',
    `note_url`          VARCHAR(255)                 DEFAULT NULL    COMMENT '笔记详情页的URL',
    `source_keyword`    VARCHAR(255)                 DEFAULT ''      COMMENT '搜索来源关键字',
    `video_download_url` VARCHAR(2048)                DEFAULT NULL    COMMENT '视频下载地址',
    `audio_download_url` VARCHAR(2048)                DEFAULT NULL    COMMENT '音频下载地址',
    `summary`          VARCHAR(255)                  DEFAULT NULL    COMMENT '总结',
    `keywords`         VARCHAR(255)                  DEFAULT NULL    COMMENT '关键字',
    `context`           TEXT                                         COMMENT '内容',
    `remark`            VARCHAR(512)                 DEFAULT ''      COMMENT '备注',
    `status`            TINYINT                      DEFAULT 0       COMMENT '状态，1:成功 / 0: 默认 / -1: 失败',
    `locked`            TINYINT                      DEFAULT 0       COMMENT '锁定，1:锁定 / 0: 未锁',
    `priority`          BIGINT              NOT NULL DEFAULT 0        COMMENT '优先级',
    `is_deleted`        TINYINT(1)          NOT NULL DEFAULT '0'     COMMENT '是否删除 (1:已删除, 0:未删除)',
    PRIMARY KEY (`id`),
    UNIQUE KEY `uk_xhs_note_id` (`note_id`),
    KEY `idx_note_id` (`note_id`),
    KEY `idx_time` (`time`)
) ENGINE=INNODB DEFAULT CHARSET=UTF8MB4 COLLATE=UTF8MB4_0900_AI_CI COMMENT='小红书笔记';

-- ----------------------------
-- 20. 小红书评论表
-- ----------------------------
DROP TABLE IF EXISTS `xhs_note_comment`;
CREATE TABLE `xhs_note_comment` (
    `id`                INT                 NOT NULL AUTO_INCREMENT COMMENT '自增ID',
    `task_id`           BIGINT              NOT NULL DEFAULT 0       COMMENT '爬虫任务id(关联crawler_tasks.id)',
    `user_id`           VARCHAR(64)         NOT NULL                 COMMENT '用户ID',
    `nickname`          VARCHAR(64)                  DEFAULT NULL    COMMENT '用户昵称',
    `avatar`            VARCHAR(255)                 DEFAULT NULL    COMMENT '用户头像地址',
    `ip_location`       VARCHAR(255)                 DEFAULT NULL    COMMENT '评论时的IP地址',
    `add_ts`            BIGINT              NOT NULL                 COMMENT '记录添加时间戳',
    `last_modify_ts`    BIGINT              NOT NULL                 COMMENT '记录最后修改时间戳',
    `comment_id`        VARCHAR(64)         NOT NULL                 COMMENT '评论ID',
    `create_time`       BIGINT              NOT NULL                 COMMENT '评论时间戳',
    `note_id`           VARCHAR(64)         NOT NULL                 COMMENT '笔记ID',
    `content`           LONGTEXT            NOT NULL                 COMMENT '评论内容',
    `sub_comment_count` VARCHAR(64)         NOT NULL                 COMMENT '子评论数量',
    `pictures`          VARCHAR(512)                 DEFAULT NULL,
    `parent_comment_id` VARCHAR(64)                  DEFAULT NULL    COMMENT '父评论ID',
    `like_count`        VARCHAR(255)        NOT NULL DEFAULT '0'     COMMENT '点赞数',
    `note_url`          VARCHAR(255)        NOT NULL DEFAULT ''      COMMENT '所属的笔记链接',
    `target_comment_id` VARCHAR(64)                  DEFAULT NULL    COMMENT '目标评论ID',
    `video_download_url` VARCHAR(2048)                DEFAULT NULL    COMMENT '视频下载地址',
    `audio_download_url` VARCHAR(2048)                DEFAULT NULL    COMMENT '音频下载地址',
    `summary`          VARCHAR(255)                  DEFAULT NULL    COMMENT '总结',
    `keywords`         VARCHAR(255)                  DEFAULT NULL    COMMENT '关键字',
    `context`           TEXT                                         COMMENT '内容',
    `remark`            VARCHAR(512)                 DEFAULT ''      COMMENT '备注',
    `status`            TINYINT                      DEFAULT 0       COMMENT '状态，1:成功 / 0: 默认 / -1: 失败',
    `locked`            TINYINT                      DEFAULT 0       COMMENT '锁定，1:锁定 / 0: 未锁',
    `priority`          BIGINT              NOT NULL DEFAULT 0        COMMENT '优先级',
    `is_deleted`        TINYINT(1)          NOT NULL DEFAULT '0'     COMMENT '是否删除 (1:已删除, 0:未删除)',
    PRIMARY KEY (`id`),
    UNIQUE KEY `uk_xhs_comment_id` (`comment_id`),
    KEY `idx_comment_id` (`comment_id`),
    KEY `idx_create_time` (`create_time`)
) ENGINE=INNODB DEFAULT CHARSET=UTF8MB4 COLLATE=UTF8MB4_0900_AI_CI COMMENT='小红书笔记评论';

-- ----------------------------
-- 21. 贴吧帖子表
-- ----------------------------
DROP TABLE IF EXISTS `tieba_note`;
CREATE TABLE `tieba_note` (
    `id`                BIGINT              NOT NULL AUTO_INCREMENT PRIMARY KEY,
    `task_id`           BIGINT              NOT NULL DEFAULT 0       COMMENT '爬虫任务id(关联crawler_tasks.id)',
    `note_id`           VARCHAR(644)        NOT NULL                 COMMENT '帖子ID',
    `title`             VARCHAR(255)        NOT NULL                 COMMENT '帖子标题',
    `desc`              TEXT                                         COMMENT '帖子描述',
    `note_url`          VARCHAR(255)        NOT NULL                 COMMENT '帖子链接',
    `publish_time`      VARCHAR(255)        NOT NULL                 COMMENT '发布时间',
    `user_link`         VARCHAR(255)                 DEFAULT ''      COMMENT '用户主页链接',
    `user_nickname`     VARCHAR(255)                 DEFAULT ''      COMMENT '用户昵称',
    `user_avatar`       VARCHAR(255)                 DEFAULT ''      COMMENT '用户头像地址',
    `tieba_id`          VARCHAR(255)                 DEFAULT ''      COMMENT '贴吧ID',
    `tieba_name`        VARCHAR(255)        NOT NULL                 COMMENT '贴吧名称',
    `tieba_link`        VARCHAR(255)        NOT NULL                 COMMENT '贴吧链接',
    `total_replay_num`  INT                          DEFAULT 0       COMMENT '帖子回复总数',
    `total_replay_page` INT                          DEFAULT 0       COMMENT '帖子回复总页数',
    `ip_location`       VARCHAR(255)                 DEFAULT ''      COMMENT 'IP地理位置',
    `add_ts`            BIGINT              NOT NULL                 COMMENT '添加时间戳',
    `last_modify_ts`    BIGINT              NOT NULL                 COMMENT '最后修改时间戳',
    `source_keyword`    VARCHAR(255)                 DEFAULT ''      COMMENT '搜索来源关键字',
    `video_download_url` VARCHAR(2048)                DEFAULT NULL    COMMENT '视频下载地址',
    `audio_download_url` VARCHAR(2048)                DEFAULT NULL    COMMENT '音频下载地址',
    `summary`          VARCHAR(255)                  DEFAULT NULL    COMMENT '总结',
    `keywords`         VARCHAR(255)                  DEFAULT NULL    COMMENT '关键字',
    `context`           TEXT                                         COMMENT '内容',
    `remark`            VARCHAR(512)                 DEFAULT ''      COMMENT '备注',
    `status`            TINYINT                      DEFAULT 0       COMMENT '状态，1:成功 / 0: 默认 / -1: 失败',
    `locked`            TINYINT                      DEFAULT 0       COMMENT '锁定，1:锁定 / 0: 未锁',
    `priority`          BIGINT              NOT NULL DEFAULT 0        COMMENT '优先级',
    `is_deleted`        TINYINT(1)          NOT NULL DEFAULT '0'     COMMENT '是否删除 (1:已删除, 0:未删除)',
    UNIQUE KEY `uk_tieba_note_id` (`note_id`),
    KEY `idx_note_id` (`note_id`),
    KEY `idx_publish_time` (`publish_time`)
) ENGINE=INNODB DEFAULT CHARSET=UTF8MB4 COLLATE=UTF8MB4_0900_AI_CI COMMENT='贴吧帖子表';

-- ----------------------------
-- 22. 贴吧评论表
-- ----------------------------
DROP TABLE IF EXISTS `tieba_comment`;
CREATE TABLE `tieba_comment` (
    `id`                BIGINT              NOT NULL AUTO_INCREMENT PRIMARY KEY,
    `task_id`           BIGINT              NOT NULL DEFAULT 0       COMMENT '爬虫任务id(关联crawler_tasks.id)',
    `comment_id`        VARCHAR(255)        NOT NULL                 COMMENT '评论ID',
    `parent_comment_id` VARCHAR(255)                 DEFAULT ''      COMMENT '父评论ID',
    `content`           TEXT                NOT NULL                 COMMENT '评论内容',
    `user_link`         VARCHAR(255)                 DEFAULT ''      COMMENT '用户主页链接',
    `user_nickname`     VARCHAR(255)                 DEFAULT ''      COMMENT '用户昵称',
    `user_avatar`       VARCHAR(255)                 DEFAULT ''      COMMENT '用户头像地址',
    `tieba_id`          VARCHAR(255)                 DEFAULT ''      COMMENT '贴吧ID',
    `tieba_name`        VARCHAR(255)        NOT NULL                 COMMENT '贴吧名称',
    `tieba_link`        VARCHAR(255)        NOT NULL                 COMMENT '贴吧链接',
    `publish_time`      VARCHAR(255)                 DEFAULT ''      COMMENT '发布时间',
    `ip_location`       VARCHAR(255)                 DEFAULT ''      COMMENT 'IP地理位置',
    `sub_comment_count` INT                          DEFAULT 0       COMMENT '子评论数',
    `note_id`           VARCHAR(255)        NOT NULL                 COMMENT '帖子ID',
    `note_url`          VARCHAR(255)        NOT NULL                 COMMENT '帖子链接',
    `add_ts`            BIGINT              NOT NULL                 COMMENT '添加时间戳',
    `last_modify_ts`    BIGINT              NOT NULL                 COMMENT '最后修改时间戳',
    `video_download_url` VARCHAR(2048)                DEFAULT NULL    COMMENT '视频下载地址',
    `audio_download_url` VARCHAR(2048)                DEFAULT NULL    COMMENT '音频下载地址',
    `summary`          VARCHAR(255)                  DEFAULT NULL    COMMENT '总结',
    `keywords`         VARCHAR(255)                  DEFAULT NULL    COMMENT '关键字',
    `context`           TEXT                                         COMMENT '内容',
    `remark`            VARCHAR(512)                 DEFAULT ''      COMMENT '备注',
    `status`            TINYINT                      DEFAULT 0       COMMENT '状态，1:成功 / 0: 默认 / -1: 失败',
    `locked`            TINYINT                      DEFAULT 0       COMMENT '锁定，1:锁定 / 0: 未锁',
    `priority`          BIGINT              NOT NULL DEFAULT 0        COMMENT '优先级',
    `is_deleted`        TINYINT(1)          NOT NULL DEFAULT '0'     COMMENT '是否删除 (1:已删除, 0:未删除)',
    UNIQUE KEY `uk_tieba_comment_id` (`comment_id`),
    KEY `idx_comment_id` (`comment_id`),
    KEY `idx_note_id` (`note_id`),
    KEY `idx_publish_time` (`publish_time`)
) ENGINE=INNODB DEFAULT CHARSET=UTF8MB4 COLLATE=UTF8MB4_0900_AI_CI COMMENT='贴吧评论表';

-- ----------------------------
-- 23. 微博博主表
-- ----------------------------
DROP TABLE IF EXISTS `weibo_creator`;
CREATE TABLE `weibo_creator` (
    `id`                INT                 NOT NULL AUTO_INCREMENT COMMENT '自增ID',
    `task_id`           BIGINT              NOT NULL DEFAULT 0       COMMENT '爬虫任务id(关联crawler_tasks.id)',
    `user_id`           VARCHAR(64)         NOT NULL                 COMMENT '用户ID',
    `nickname`          VARCHAR(64)                  DEFAULT NULL    COMMENT '用户昵称',
    `avatar`            VARCHAR(255)                 DEFAULT NULL    COMMENT '用户头像地址',
    `ip_location`       VARCHAR(255)                 DEFAULT NULL    COMMENT '评论时的IP地址',
    `add_ts`            BIGINT              NOT NULL                 COMMENT '记录添加时间戳',
    `last_modify_ts`    BIGINT              NOT NULL                 COMMENT '记录最后修改时间戳',
    `desc`              LONGTEXT                                     COMMENT '用户描述',
    `gender`            VARCHAR(2)                   DEFAULT NULL    COMMENT '性别',
    `follows`           VARCHAR(16)                  DEFAULT NULL    COMMENT '关注数',
    `fans`              VARCHAR(16)                  DEFAULT NULL    COMMENT '粉丝数',
    `tag_list`          LONGTEXT                                     COMMENT '标签列表',
    `video_download_url` VARCHAR(2048)                DEFAULT NULL    COMMENT '视频下载地址',
    `audio_download_url` VARCHAR(2048)                DEFAULT NULL    COMMENT '音频下载地址',
    `summary`          VARCHAR(255)                  DEFAULT NULL    COMMENT '总结',
    `keywords`         VARCHAR(255)                  DEFAULT NULL    COMMENT '关键字',
    `context`           TEXT                                         COMMENT '内容',
    `remark`            VARCHAR(512)                 DEFAULT ''      COMMENT '备注',
    `status`            TINYINT                      DEFAULT 0       COMMENT '状态，1:成功 / 0: 默认 / -1: 失败',
    `locked`            TINYINT                      DEFAULT 0       COMMENT '锁定，1:锁定 / 0: 未锁',
    `priority`          BIGINT              NOT NULL DEFAULT 0        COMMENT '优先级',
    `is_deleted`        TINYINT(1)          NOT NULL DEFAULT '0'     COMMENT '是否删除 (1:已删除, 0:未删除)',
    PRIMARY KEY (`id`),
    UNIQUE KEY `uk_weibo_creator_user_id` (`user_id`)
) ENGINE=INNODB DEFAULT CHARSET=UTF8MB4 COLLATE=UTF8MB4_0900_AI_CI COMMENT='微博博主';

-- ----------------------------
-- 24. 贴吧创作者表
-- ----------------------------
DROP TABLE IF EXISTS `tieba_creator`;
CREATE TABLE `tieba_creator` (
    `id`                    INT                 NOT NULL AUTO_INCREMENT COMMENT '自增ID',
    `task_id`               BIGINT              NOT NULL DEFAULT 0       COMMENT '爬虫任务id(关联crawler_tasks.id)',
    `user_id`               VARCHAR(64)         NOT NULL                 COMMENT '用户ID',
    `user_name`             VARCHAR(64)         NOT NULL                 COMMENT '用户名',
    `nickname`              VARCHAR(64)                  DEFAULT NULL    COMMENT '用户昵称',
    `avatar`                VARCHAR(255)                 DEFAULT NULL    COMMENT '用户头像地址',
    `ip_location`           VARCHAR(255)                 DEFAULT NULL    COMMENT '评论时的IP地址',
    `add_ts`                BIGINT              NOT NULL                 COMMENT '记录添加时间戳',
    `last_modify_ts`        BIGINT              NOT NULL                 COMMENT '记录最后修改时间戳',
    `gender`                VARCHAR(2)                   DEFAULT NULL    COMMENT '性别',
    `follows`               VARCHAR(16)                  DEFAULT NULL    COMMENT '关注数',
    `fans`                  VARCHAR(16)                  DEFAULT NULL    COMMENT '粉丝数',
    `registration_duration` VARCHAR(16)                  DEFAULT NULL    COMMENT '吧龄',
    `video_download_url` VARCHAR(2048)                    DEFAULT NULL    COMMENT '视频下载地址',
    `audio_download_url` VARCHAR(2048)                    DEFAULT NULL    COMMENT '音频下载地址',
    `summary`          VARCHAR(255)                  DEFAULT NULL    COMMENT '总结',
    `keywords`         VARCHAR(255)                  DEFAULT NULL    COMMENT '关键字',
    `context`               TEXT                                         COMMENT '内容',
    `remark`                VARCHAR(512)                 DEFAULT ''      COMMENT '备注',
    `status`                TINYINT                      DEFAULT 0       COMMENT '状态，1:成功 / 0: 默认 / -1: 失败',
    `locked`                TINYINT                      DEFAULT 0       COMMENT '锁定，1:锁定 / 0: 未锁',
    `priority`          BIGINT              NOT NULL DEFAULT 0        COMMENT '优先级',
    `is_deleted`            TINYINT(1)          NOT NULL DEFAULT '0'     COMMENT '是否删除 (1:已删除, 0:未删除)',
    PRIMARY KEY (`id`),
    UNIQUE KEY `uk_tieba_creator_user_id` (`user_id`)
) ENGINE=INNODB DEFAULT CHARSET=UTF8MB4 COLLATE=UTF8MB4_0900_AI_CI COMMENT='贴吧创作者';

-- ----------------------------
-- 25. 知乎内容表
-- ----------------------------
DROP TABLE IF EXISTS `zhihu_content`;
CREATE TABLE `zhihu_content` (
    `id`                INT                 NOT NULL AUTO_INCREMENT COMMENT '自增ID',
    `task_id`           BIGINT              NOT NULL DEFAULT 0       COMMENT '爬虫任务id(关联crawler_tasks.id)',
    `content_id`        VARCHAR(64)         NOT NULL                 COMMENT '内容ID',
    `content_type`      VARCHAR(16)         NOT NULL                 COMMENT '内容类型(article | answer | zvideo)',
    `content_text`      LONGTEXT                                     COMMENT '内容文本, 如果是视频类型这里为空',
    `content_url`       VARCHAR(255)        NOT NULL                 COMMENT '内容落地链接',
    `question_id`       VARCHAR(64)                  DEFAULT NULL    COMMENT '问题ID, type为answer时有值',
    `title`             VARCHAR(255)        NOT NULL                 COMMENT '内容标题',
    `desc`              LONGTEXT                                     COMMENT '内容描述',
    `created_time`      VARCHAR(32)         NOT NULL                 COMMENT '创建时间',
    `updated_time`      VARCHAR(32)         NOT NULL                 COMMENT '更新时间',
    `voteup_count`      INT                 NOT NULL DEFAULT '0'     COMMENT '赞同人数',
    `comment_count`     INT                 NOT NULL DEFAULT '0'     COMMENT '评论数量',
    `source_keyword`    VARCHAR(64)                  DEFAULT NULL    COMMENT '来源关键词',
    `user_id`           VARCHAR(64)         NOT NULL                 COMMENT '用户ID',
    `user_link`         VARCHAR(255)        NOT NULL                 COMMENT '用户主页链接',
    `user_nickname`     VARCHAR(64)         NOT NULL                 COMMENT '用户昵称',
    `user_avatar`       VARCHAR(255)        NOT NULL                 COMMENT '用户头像地址',
    `user_url_token`    VARCHAR(255)        NOT NULL                 COMMENT '用户url_token',
    `add_ts`            BIGINT              NOT NULL                 COMMENT '记录添加时间戳',
    `last_modify_ts`    BIGINT              NOT NULL                 COMMENT '记录最后修改时间戳',
    `video_download_url` VARCHAR(2048)                DEFAULT NULL    COMMENT '视频下载地址',
    `audio_download_url` VARCHAR(2048)                DEFAULT NULL    COMMENT '音频下载地址',
    `summary`          VARCHAR(255)                  DEFAULT NULL    COMMENT '总结',
    `keywords`         VARCHAR(255)                  DEFAULT NULL    COMMENT '关键字',
    `context`           TEXT                                         COMMENT '内容',
    `remark`            VARCHAR(512)                 DEFAULT ''      COMMENT '备注',
    `status`            TINYINT                      DEFAULT 0       COMMENT '状态，1:成功 / 0: 默认 / -1: 失败',
    `locked`            TINYINT                      DEFAULT 0       COMMENT '锁定，1:锁定 / 0: 未锁',
    `priority`          BIGINT              NOT NULL DEFAULT 0        COMMENT '优先级',
    `is_deleted`        TINYINT(1)          NOT NULL DEFAULT '0'     COMMENT '是否删除 (1:已删除, 0:未删除)',
    PRIMARY KEY (`id`),
    UNIQUE KEY `uk_zhihu_content_id` (`content_id`),
    KEY `idx_content_id` (`content_id`),
    KEY `idx_created_time` (`created_time`)
) ENGINE=INNODB DEFAULT CHARSET=UTF8MB4 COLLATE=UTF8MB4_0900_AI_CI COMMENT='知乎内容（回答、文章、视频）';

-- ----------------------------
-- 26. 知乎评论表
-- ----------------------------
DROP TABLE IF EXISTS `zhihu_comment`;
CREATE TABLE `zhihu_comment` (
    `id`                INT                 NOT NULL AUTO_INCREMENT COMMENT '自增ID',
    `task_id`           BIGINT              NOT NULL DEFAULT 0       COMMENT '爬虫任务id(关联crawler_tasks.id)',
    `comment_id`        VARCHAR(64)         NOT NULL                 COMMENT '评论ID',
    `parent_comment_id` VARCHAR(64)                  DEFAULT NULL    COMMENT '父评论ID',
    `content`           TEXT                NOT NULL                 COMMENT '评论内容',
    `publish_time`      VARCHAR(32)         NOT NULL                 COMMENT '发布时间',
    `ip_location`       VARCHAR(64)                  DEFAULT NULL    COMMENT 'IP地理位置',
    `sub_comment_count` INT                 NOT NULL DEFAULT '0'     COMMENT '子评论数',
    `like_count`        INT                 NOT NULL DEFAULT '0'     COMMENT '点赞数',
    `dislike_count`     INT                 NOT NULL DEFAULT '0'     COMMENT '踩数',
    `content_id`        VARCHAR(64)         NOT NULL                 COMMENT '内容ID',
    `content_type`      VARCHAR(16)         NOT NULL                 COMMENT '内容类型(article | answer | zvideo)',
    `user_id`           VARCHAR(64)         NOT NULL                 COMMENT '用户ID',
    `user_link`         VARCHAR(255)        NOT NULL                 COMMENT '用户主页链接',
    `user_nickname`     VARCHAR(64)         NOT NULL                 COMMENT '用户昵称',
    `user_avatar`       VARCHAR(255)        NOT NULL                 COMMENT '用户头像地址',
    `add_ts`            BIGINT              NOT NULL                 COMMENT '记录添加时间戳',
    `last_modify_ts`    BIGINT              NOT NULL                 COMMENT '记录最后修改时间戳',
    `video_download_url` VARCHAR(2048)                DEFAULT NULL    COMMENT '视频下载地址',
    `audio_download_url` VARCHAR(2048)                DEFAULT NULL    COMMENT '音频下载地址',
    `summary`          VARCHAR(255)                  DEFAULT NULL    COMMENT '总结',
    `keywords`         VARCHAR(255)                  DEFAULT NULL    COMMENT '关键字',
    `context`           TEXT                                         COMMENT '内容',
    `remark`            VARCHAR(512)                 DEFAULT ''      COMMENT '备注',
    `status`            TINYINT                      DEFAULT 0       COMMENT '状态，1:成功 / 0: 默认 / -1: 失败',
    `locked`            TINYINT                      DEFAULT 0       COMMENT '锁定，1:锁定 / 0: 未锁',
    `priority`          BIGINT              NOT NULL DEFAULT 0        COMMENT '优先级',
    `is_deleted`        TINYINT(1)          NOT NULL DEFAULT '0'     COMMENT '是否删除 (1:已删除, 0:未删除)',
    PRIMARY KEY (`id`),
    UNIQUE KEY `uk_zhihu_comment_id` (`comment_id`),
    KEY `idx_comment_id` (`comment_id`),
    KEY `idx_content_id` (`content_id`),
    KEY `idx_publish_time` (`publish_time`)
) ENGINE=INNODB DEFAULT CHARSET=UTF8MB4 COLLATE=UTF8MB4_0900_AI_CI COMMENT='知乎评论';

-- ----------------------------
-- 27. 知乎创作者表
-- ----------------------------
DROP TABLE IF EXISTS `zhihu_creator`;
CREATE TABLE `zhihu_creator` (
    `id`                INT                 NOT NULL AUTO_INCREMENT COMMENT '自增ID',
    `task_id`           BIGINT              NOT NULL DEFAULT 0       COMMENT '爬虫任务id(关联crawler_tasks.id)',
    `user_id`           VARCHAR(64)         NOT NULL                 COMMENT '用户ID',
    `user_link`         VARCHAR(255)        NOT NULL                 COMMENT '用户主页链接',
    `user_nickname`     VARCHAR(64)         NOT NULL                 COMMENT '用户昵称',
    `user_avatar`       VARCHAR(255)        NOT NULL                 COMMENT '用户头像地址',
    `url_token`         VARCHAR(64)         NOT NULL                 COMMENT '用户URL Token',
    `gender`            VARCHAR(16)                  DEFAULT NULL    COMMENT '用户性别',
    `ip_location`       VARCHAR(64)                  DEFAULT NULL    COMMENT 'IP地理位置',
    `follows`           INT                 NOT NULL DEFAULT 0       COMMENT '关注数',
    `fans`              INT                 NOT NULL DEFAULT 0       COMMENT '粉丝数',
    `anwser_count`      INT                 NOT NULL DEFAULT 0       COMMENT '回答数',
    `video_count`       INT                 NOT NULL DEFAULT 0       COMMENT '视频数',
    `question_count`    INT                 NOT NULL DEFAULT 0       COMMENT '问题数',
    `article_count`     INT                 NOT NULL DEFAULT 0       COMMENT '文章数',
    `column_count`      INT                 NOT NULL DEFAULT 0       COMMENT '专栏数',
    `get_voteup_count`  INT                 NOT NULL DEFAULT 0       COMMENT '获得的赞同数',
    `add_ts`            BIGINT              NOT NULL                 COMMENT '记录添加时间戳',
    `last_modify_ts`    BIGINT              NOT NULL                 COMMENT '记录最后修改时间戳',
    `video_download_url` VARCHAR(2048)                DEFAULT NULL    COMMENT '视频下载地址',
    `audio_download_url` VARCHAR(2048)                DEFAULT NULL    COMMENT '音频下载地址',
    `summary`          VARCHAR(255)                  DEFAULT NULL    COMMENT '总结',
    `keywords`         VARCHAR(255)                  DEFAULT NULL    COMMENT '关键字',
    `context`           TEXT                                         COMMENT '内容',
    `remark`            VARCHAR(512)                 DEFAULT ''      COMMENT '备注',
    `status`            TINYINT                      DEFAULT 0       COMMENT '状态，1:成功 / 0: 默认 / -1: 失败',
    `locked`            TINYINT                      DEFAULT 0       COMMENT '锁定，1:锁定 / 0: 未锁',
    `priority`          BIGINT              NOT NULL DEFAULT 0        COMMENT '优先级',
    `is_deleted`        TINYINT(1)          NOT NULL DEFAULT '0'     COMMENT '是否删除 (1:已删除, 0:未删除)',
    PRIMARY KEY (`id`),
    UNIQUE KEY `uk_user_id` (`user_id`)
) ENGINE=INNODB DEFAULT CHARSET=UTF8MB4 COLLATE=UTF8MB4_0900_AI_CI COMMENT='知乎创作者';

-- ----------------------------
-- 28. 任务组表（用于串联多个单体任务）
-- ----------------------------
DROP TABLE IF EXISTS `crawler_task_groups`;
CREATE TABLE `crawler_task_groups` (
    `id`                BIGINT              NOT NULL AUTO_INCREMENT  COMMENT '自增ID',
    `crawler_count`     INT                 NOT NULL DEFAULT 5       COMMENT '任务爬取次数',
    `group_name`        VARCHAR(100)        NOT NULL                 COMMENT '任务组名称',
    `total_steps`       INT                 NOT NULL DEFAULT 0       COMMENT '包含的任务总数',
    `current_step`      INT                 NOT NULL DEFAULT 0       COMMENT '当前执行到第几个步骤',
    `add_ts`            BIGINT              NOT NULL                 COMMENT '记录添加时间戳',
    `last_modify_ts`    BIGINT              NOT NULL                 COMMENT '记录最后修改时间戳',
    `status`            TINYINT                      DEFAULT 0       COMMENT '执行状态: 0:等待, 3:入队, 2:运行中, 1:成功, -1:错误, -2:跳过, -3:超时, -4:取消',
    `locked`            TINYINT                      DEFAULT 0       COMMENT '锁定，1:锁定 / 0: 未锁',
    `priority`          BIGINT              NOT NULL DEFAULT 0        COMMENT '优先级',
    `is_deleted`        TINYINT(1)          NOT NULL DEFAULT '0'     COMMENT '是否删除 (1:已删除, 0:未删除)',
    PRIMARY KEY (`id`),
    INDEX `idx_status` (`status`)
) ENGINE=INNODB DEFAULT CHARSET=UTF8MB4 COLLATE=UTF8MB4_0900_AI_CI COMMENT='爬虫任务组表';

-- ----------------------------
-- 29. 单体任务表（实际执行单元）
-- ----------------------------
DROP TABLE IF EXISTS `crawler_tasks`;
CREATE TABLE `crawler_tasks` (
    `id`                BIGINT              NOT NULL AUTO_INCREMENT  COMMENT '自增ID',
    `task_name`         VARCHAR(32)         NOT NULL                 COMMENT '任务名',
    `group_id`          BIGINT                       DEFAULT NULL    COMMENT '所属任务组ID，单体任务可为NULL',
    `dependency_id`     BIGINT                       DEFAULT NULL    COMMENT '前置依赖任务ID',
    `sort_order`        INT                 NOT NULL DEFAULT 0       COMMENT '在组中的执行顺序',
    `task_type`         VARCHAR(32)         NOT NULL                 COMMENT '类型(creator/search/detail)',
    `crawler_count`     INT                 NOT NULL DEFAULT 9999    COMMENT '任务爬取次数',
    `media_list_id`     INT                 NOT NULL                 COMMENT '媒体管理中的媒体id',
    `pid`               INT                          DEFAULT NULL    COMMENT '进程PID',
    `retry_count`       INT                 NOT NULL DEFAULT 0       COMMENT '当前重试次数',
    `max_retries`       INT                 NOT NULL DEFAULT 3       COMMENT '最大重试次数',
    `log_path`          VARCHAR(255)                 DEFAULT NULL    COMMENT '日志路径',
    `error_msg`         TEXT                                         COMMENT '异常堆栈信息',
    `add_ts`            BIGINT              NOT NULL                 COMMENT '记录添加时间戳',
    `start_ts`          BIGINT                       DEFAULT NULL    COMMENT '任务实际开始时间戳',
    `last_modify_ts`    BIGINT              NOT NULL                 COMMENT '最后修改时间戳',
    `status`            TINYINT                      DEFAULT 0       COMMENT '状态: 0:等待, 3:入队, 2:运行中, 1:成功, -1:错误, -2:跳过, -3:超时, -4:取消', 
    `locked`            TINYINT                      DEFAULT 0       COMMENT '锁定，1:锁定 / 0: 未锁',
    `priority`          BIGINT              NOT NULL DEFAULT 0        COMMENT '优先级',
    `is_deleted`        TINYINT(1)          NOT NULL DEFAULT '0'     COMMENT '是否删除 (1:已删除, 0:未删除)',
    PRIMARY KEY (`id`),
    INDEX `idx_task_name` (`task_name`),
    INDEX `idx_group_id` (`group_id`),
    INDEX `idx_status` (`status`),
    INDEX `idx_media_list_id` (`media_list_id`)
) ENGINE=INNODB DEFAULT CHARSET=UTF8MB4 COLLATE=UTF8MB4_0900_AI_CI COMMENT='爬虫单体任务表';

-- ----------------------------
-- 30. 任务组与单体任务关联表（一对多：一个任务组对应多个单体任务）
-- ----------------------------
DROP TABLE IF EXISTS `rel_task_group_task`;
CREATE TABLE `rel_task_group_task` (
    `group_id`          BIGINT              NOT NULL COMMENT '任务组ID(关联crawler_task_groups.id)',
    `task_id`           BIGINT              NOT NULL COMMENT '单体任务ID(关联crawler_tasks.id)',
    `is_deleted`        TINYINT(1)          NOT NULL DEFAULT '0'     COMMENT '是否删除 (1:已删除, 0:未删除)',
    PRIMARY KEY (`group_id`, `task_id`),
    KEY `idx_task_id` (`task_id`)
) ENGINE=INNODB DEFAULT CHARSET=UTF8MB4 COLLATE=UTF8MB4_0900_AI_CI COMMENT='任务组与单体任务关联表';

-- ----------------------------
-- 31. 单体任务执行记录表（执行快照与历史回溯）
-- ----------------------------
DROP TABLE IF EXISTS `crawler_task_records`;
CREATE TABLE `crawler_task_records` (
    `id`                BIGINT              NOT NULL AUTO_INCREMENT  COMMENT '记录自增ID',
    `task_type`         TINYINT             NOT NULL DEFAULT 0       COMMENT '任务类型，0:单体任务 / 1: 任务组',
    `task_id`           BIGINT                       DEFAULT NULL    COMMENT '单体任务虫id(关联crawler_tasks.id)',
    `task_group_id`     BIGINT                       DEFAULT NULL    COMMENT '任务组id(crawler_task_groups.id)',
    `crawler_count`     INT                 NOT NULL DEFAULT 9999    COMMENT '任务爬取次数',
    `execution_id`      VARCHAR(64)         NOT NULL                 COMMENT '单次执行唯一标识(UUID/时序ID)',
    
    -- 快照字段 (Snapshot Fields - 带有 snap_ 前缀)
    `snap_task_name`    VARCHAR(32)         NOT NULL                 COMMENT '快照: 任务名',
    `snap_task_type`    VARCHAR(32)         NOT NULL                 COMMENT '快照: 类型(creator/search/detail)',
    `snap_table_name`   VARCHAR(32)         NOT NULL                 COMMENT '快照: 存储表名。可选: bilibili_video, bilibili_video_comment, bilibili_up_info, douyin_aweme, douyin_aweme_comment, dy_creator, kuaishou_video, kuaishou_video_comment, kuaishou_creator, weibo_note, weibo_note_comment, xhs_creator, xhs_note, xhs_note_comment, tieba_note, tieba_comment, weibo_creator, tieba_creator, zhihu_content, zhihu_comment, zhihu_creator, crawler_task_groups',
    `snap_media_id`     INT                 NOT NULL                 COMMENT '快照: 媒体管理ID',
    `snap_retry_count`  INT                 NOT NULL DEFAULT 0       COMMENT '快照: 当前是第几次重试',
    `snap_max_retries`  INT                 NOT NULL DEFAULT 3       COMMENT '快照: 当时的最大重试限制',
    
    -- 运行时动态数据 (Runtime Data)
    `pid`               INT                          DEFAULT NULL    COMMENT '实际执行进程PID',
    `worker_ip`         VARCHAR(64)                  DEFAULT NULL    COMMENT '执行节点IP/容器ID',
    
    -- 时间维度 (Time Metrics)
    `start_ts`          BIGINT                       DEFAULT NULL    COMMENT '任务真正开始执行时间戳',
    `end_ts`            BIGINT                       DEFAULT NULL    COMMENT '任务执行结束/中断时间戳',
    `duration_ms`       BIGINT                       DEFAULT NULL    COMMENT '实际运行耗时(毫秒)',
    
    -- 结果与留存 (Results & Logs)
    `log_path`          VARCHAR(255)                 DEFAULT NULL    COMMENT '本次运行对应的独立日志路径',
    `error_msg`         TEXT                                         COMMENT '执行失败时的异常堆栈',
    `exit_code`         INT                          DEFAULT NULL    COMMENT '进程退出状态码',
    `data_count`        INT                          DEFAULT 0       COMMENT '本次执行抓取到的条数',
    
    -- 拓展可能性 (Extensibility)
    `env_info`          JSON                         DEFAULT NULL    COMMENT '执行环境快照(如代理IP、UA、版本号)',
    `config_snapshot`   JSON                         DEFAULT NULL    COMMENT '执行时的具体爬取配置参数',
    
    `add_ts`            BIGINT              NOT NULL                 COMMENT '记录创建时间戳',


    `status`            TINYINT                      DEFAULT 0       COMMENT '执行状态: 0:等待, 3:入队, 2:运行中, 1:成功, -1:错误, -2:跳过, -3:超时, -4:取消', 
    `locked`            TINYINT                      DEFAULT 0       COMMENT '锁定，1:锁定 / 0: 未锁',
    `priority`          BIGINT              NOT NULL DEFAULT 0        COMMENT '优先级',
    `is_deleted`        TINYINT(1)          NOT NULL DEFAULT '0'     COMMENT '是否删除 (1:已删除, 0:未删除)',
    PRIMARY KEY (`id`),
    UNIQUE INDEX `uk_exec_id` (`execution_id`),
    INDEX `idx_task_id` (`task_id`),
    INDEX `idx_status` (`status`),
    INDEX `idx_add_ts` (`add_ts`)
) ENGINE=INNODB DEFAULT CHARSET=UTF8MB4 COLLATE=UTF8MB4_0900_AI_CI COMMENT='爬虫任务执行记录表(快照)';

-- ----------------------------
-- 32. AI功能
-- ----------------------------
DROP TABLE IF EXISTS `ai_agent_text`;
CREATE TABLE `ai_agent_text` (
    `id`       BIGINT       NOT NULL AUTO_INCREMENT COMMENT 'AI功能ID',
    `key`      VARCHAR(32)  NOT NULL                COMMENT 'AI功能KEY',
    `content`  TEXT         NOT NULL                COMMENT '内容',
    PRIMARY KEY (`id`)
) ENGINE=INNODB DEFAULT CHARSET=UTF8MB4 COLLATE=UTF8MB4_0900_AI_CI COMMENT='AI文字功能';

SET FOREIGN_KEY_CHECKS = 1;