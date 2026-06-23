-- 创建数据库
CREATE DATABASE IF NOT EXISTS game_db DEFAULT CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;

-- 使用数据库
USE game_db;
-- drop table `role`;
-- 创建角色表
CREATE TABLE `role` (
                        `role_id` bigint NOT NULL COMMENT '角色唯一ID',
                        `account` varchar(32) NOT NULL COMMENT '账号ID',
                        `server_id` int NOT NULL COMMENT '服务器ID',
                        `plat_id` int NOT NULL COMMENT '平台ID',
                        `name` varchar(32) NOT NULL COMMENT '角色名',
                        `gender` int NOT NULL COMMENT '性别',
                        `career` int NOT NULL COMMENT '职业', 
                        `lv` int NOT NULL DEFAULT '1' COMMENT '等级',
                        `data` TEXT NOT NULL COMMENT '玩家业务聚合数据',
                        `created_at` datetime DEFAULT CURRENT_TIMESTAMP,
                        `update_at` datetime DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP, -- 加上自动更新时间
                        PRIMARY KEY (`role_id`),
                        UNIQUE KEY `uk_name` (`name`), -- 服内唯一
                        INDEX `idx_account` (`account`), -- 按账号查询的普通索引
                        INDEX `idx_server` (`server_id`), -- 按服务器查询的普通索引
                        INDEX `idx_plat` (`plat_id`) -- 按平台查询的普通索引
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='角色表';