-- 001: C 端用户表
-- 数据库：template_order_db
-- 前缀：tos_（TemplateOrderServer 管理）

CREATE TABLE IF NOT EXISTS tos_users (
  id            BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  username      VARCHAR(64) NOT NULL,
  password      VARCHAR(255) NOT NULL,          -- bcrypt
  member_status TINYINT NOT NULL DEFAULT 0,     -- 0=非会员 1=会员
  created_at    DATETIME(3) NOT NULL,
  updated_at    DATETIME(3) NOT NULL,
  PRIMARY KEY (id),
  UNIQUE KEY uk_username (username)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;