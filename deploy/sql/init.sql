-- =========================================================
-- sup-iam 数据库初始化脚本
-- =========================================================

CREATE DATABASE IF NOT EXISTS sup_iam
  DEFAULT CHARACTER SET utf8mb4
  DEFAULT COLLATE utf8mb4_general_ci;

USE sup_iam;

-- =========================================================
-- 1. users 表（控制面用户）
-- =========================================================
CREATE TABLE users (
                       id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '自增主键',
                       instanceID VARCHAR(32) NOT NULL COMMENT '跨域 UUID',
                       username VARCHAR(255) NOT NULL COMMENT '用户名',
                       nickname VARCHAR(30) NOT NULL COMMENT '昵称',
                       passwordHash VARCHAR(255) NOT NULL COMMENT '密码哈希值',
                       isEnable TINYINT(1) NOT NULL COMMENT '是否启用：1-可用，0-不可用',
                       phone VARCHAR(20) DEFAULT NULL COMMENT '手机号',
                       email VARCHAR(256) DEFAULT NULL COMMENT '邮箱',
                       isAdmin TINYINT(1) DEFAULT 0 COMMENT '是否管理员',
                       extandShadow LONGTEXT DEFAULT NULL COMMENT '扩展字段',
                       loggedAt TIMESTAMP NULL DEFAULT NULL COMMENT '最近登录时间',
                       createdAt TIMESTAMP NOT NULL COMMENT '创建时间',
                       updatedAt TIMESTAMP NOT NULL COMMENT '更新时间',
                       PRIMARY KEY (id),
                       UNIQUE KEY uk_users_instanceID (instanceID),
                       UNIQUE KEY uk_users_username (username)
) ENGINE=InnoDB COMMENT='IAM 控制面用户表';

-- =========================================================
-- 2. secrets 表（AccessKey / SecretKey）
-- =========================================================
CREATE TABLE secrets (
                         id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '自增主键',
                         instanceID VARCHAR(32) NOT NULL COMMENT '跨域 UUID',
                         userID BIGINT UNSIGNED NOT NULL COMMENT '所属用户 ID',
                         username VARCHAR(255) NOT NULL COMMENT '用户名（冗余字段）',
                         accessKey VARCHAR(36) NOT NULL COMMENT 'AccessKey',
                         secretKey VARCHAR(255) NOT NULL COMMENT 'SecretKey（明文存储）',
                         expires INT UNSIGNED DEFAULT 1534308590 COMMENT '过期时间（Unix 秒）',
                         description VARCHAR(255) DEFAULT NULL COMMENT '密钥描述',
                         extendShadow LONGTEXT DEFAULT NULL COMMENT '扩展字段',
                         createdAt TIMESTAMP NOT NULL COMMENT '创建时间',
                         updatedAt TIMESTAMP NOT NULL COMMENT '更新时间',
                         PRIMARY KEY (id),
                         UNIQUE KEY uk_secrets_instanceID (instanceID),
                         UNIQUE KEY uk_secrets_accessKey (accessKey),
                         KEY idx_secrets_userID (userID),
                         KEY idx_secrets_username (username)
) ENGINE=InnoDB COMMENT='密钥表';

-- =========================================================
-- 3. policies 表（策略表）
-- =========================================================
CREATE TABLE policies (
                          id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '自增主键',
                          instanceID VARCHAR(32) NOT NULL COMMENT '跨域 UUID',
                          name VARCHAR(64) NOT NULL COMMENT '策略名称',
                          username VARCHAR(255) NOT NULL COMMENT '所属用户名',
                          description VARCHAR(255) DEFAULT NULL COMMENT '策略描述',
                          policyShadow LONGTEXT DEFAULT NULL COMMENT '策略 DSL 描述',
                          extendShadow LONGTEXT DEFAULT NULL COMMENT '扩展字段',
                          createdAt TIMESTAMP NOT NULL COMMENT '创建时间',
                          updatedAt TIMESTAMP NOT NULL COMMENT '更新时间',
                          PRIMARY KEY (id),
                          UNIQUE KEY uk_policies_instanceID (instanceID),
                          KEY idx_policies_username (username)
) ENGINE=InnoDB COMMENT='策略表';

-- =========================================================
-- 4. secret_policy_binding 表（绑定关系）
-- =========================================================
CREATE TABLE secret_policy_binding (
                                       id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '自增主键',
                                       secretID BIGINT UNSIGNED NOT NULL COMMENT '密钥 ID',
                                       policyID BIGINT UNSIGNED NOT NULL COMMENT '策略 ID',
                                       username VARCHAR(255) NOT NULL COMMENT '所属用户名',
                                       extendShadow LONGTEXT DEFAULT NULL COMMENT '扩展字段',
                                       createdAt TIMESTAMP NOT NULL COMMENT '创建时间',
                                       PRIMARY KEY (id),
                                       UNIQUE KEY uk_secret_policy (secretID, policyID),
                                       KEY idx_binding_secretID (secretID),
                                       KEY idx_binding_policyID (policyID),
                                       KEY idx_binding_username (username)
) ENGINE=InnoDB COMMENT='Secret 与 Policy 绑定关系表';

-- =========================================================
-- 5. secret_policy_binding_audit 表（绑定审计）
-- =========================================================
CREATE TABLE secret_policy_binding_audit (
                                             id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '自增主键',
                                             secretID BIGINT UNSIGNED NOT NULL COMMENT '密钥 ID',
                                             policyID BIGINT UNSIGNED NOT NULL COMMENT '策略 ID',
                                             username VARCHAR(255) NOT NULL COMMENT '所属用户名',
                                             extendShadow LONGTEXT DEFAULT NULL COMMENT '扩展字段',
                                             createdAt TIMESTAMP NOT NULL COMMENT '创建时间',
                                             PRIMARY KEY (id),
                                             KEY idx_binding_audit_secretID (secretID),
                                             KEY idx_binding_audit_policyID (policyID),
                                             KEY idx_binding_audit_username (username)
) ENGINE=InnoDB COMMENT='绑定关系审计表';

-- =========================================================
-- 6. policies_audit 表（策略审计）
-- =========================================================
CREATE TABLE policies_audit (
                                id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '自增主键',
                                instanceID VARCHAR(32) NOT NULL COMMENT '跨域 UUID',
                                name VARCHAR(64) NOT NULL COMMENT '策略名称',
                                username VARCHAR(255) NOT NULL COMMENT '所属用户名',
                                description VARCHAR(255) DEFAULT NULL COMMENT '策略描述',
                                policyShadow LONGTEXT DEFAULT NULL COMMENT '策略 DSL 描述',
                                extendShadow LONGTEXT DEFAULT NULL COMMENT '扩展字段',
                                createdAt TIMESTAMP NOT NULL COMMENT '创建时间',
                                updatedAt TIMESTAMP NOT NULL COMMENT '更新时间',
                                PRIMARY KEY (id),
                                KEY idx_policies_audit_instanceID (instanceID),
                                KEY idx_policies_audit_username (username)
) ENGINE=InnoDB COMMENT='策略审计表';
