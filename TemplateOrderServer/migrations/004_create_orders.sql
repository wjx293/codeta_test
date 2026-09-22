-- 004: 模板订单表（核心：4 类唯一约束）
-- 用途：记录 C 端用户下载模板的订单，支持免费/会员/零售三种模式
-- 关键设计：
--   1. month CHAR(7) NULL：免费/会员订单填 "yyyyMM"（Asia/Shanghai），
--      零售订单填 NULL，从而 MySQL 唯一索引不对 NULL 冲突
--   2. uk_free_month / uk_member_month：仅对 month IS NOT NULL 的行生效
--   3. active_unpaid_key 生成列 + uk_active_unpaid：仅对零售未支付订单生效
--      同一 user×template 只能有一笔"未支付"零售订单

CREATE TABLE IF NOT EXISTS tos_orders (
  id              BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  order_no        VARCHAR(32) NOT NULL,          -- TO+yyyyMMdd+10位
  user_id         BIGINT UNSIGNED NOT NULL,
  template_id     BIGINT UNSIGNED NOT NULL,
  order_type      TINYINT NOT NULL,              -- 1=免费 2=会员 3=零售
  price_snapshot  BIGINT NOT NULL,               -- 分，创建时快照
  status          TINYINT NOT NULL,              -- 1=未支付 2=已获得下载资格 3=已取消
  month           CHAR(7) NULL DEFAULT NULL,     -- 免费/会员填 "yyyyMM"，零售填 NULL
  created_at      DATETIME(3) NOT NULL,
  updated_at      DATETIME(3) NOT NULL,
  PRIMARY KEY (id),
  UNIQUE KEY uk_order_no (order_no),
  UNIQUE KEY uk_free_month (user_id, template_id, month, order_type)
    COMMENT '仅对 month IS NOT NULL 的行生效（免费订单同月唯一）',
  UNIQUE KEY uk_member_month (user_id, template_id, month, order_type)
    COMMENT '仅对 month IS NOT NULL 的行生效（会员订单同月唯一）',
  KEY idx_user_status (user_id, status),
  KEY idx_template (template_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- 零售订单未支付唯一性：生成列 + 唯一索引
-- 仅当 order_type=3 AND status=1 时生成值，否则为 NULL
-- MySQL 唯一索引允许多个 NULL 共存，故已取消/已支付的零售订单不冲突
ALTER TABLE tos_orders
  ADD COLUMN active_unpaid_key VARCHAR(64) AS (
    CASE WHEN order_type = 3 AND status = 1
         THEN CONCAT(user_id, '_', template_id)
         ELSE NULL
    END
  ) STORED,
  ADD UNIQUE KEY uk_active_unpaid (active_unpaid_key);