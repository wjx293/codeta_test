-- 001: 支付单表
-- 数据库：template_order_db
-- 前缀：pay_（PayWebServer 管理）
-- 规则：一订单一支付单（uk_business_order_no）

CREATE TABLE IF NOT EXISTS pay_payments (
  id                BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  payment_no        VARCHAR(32) NOT NULL,        -- PW+yyyyMMdd+10位
  business_order_no VARCHAR(32) NOT NULL,        -- 关联 tos_orders.order_no
  amount            BIGINT NOT NULL,             -- 分
  status            TINYINT NOT NULL,            -- 1=未支付 2=已支付
  created_at        DATETIME(3) NOT NULL,
  updated_at        DATETIME(3) NOT NULL,
  PRIMARY KEY (id),
  UNIQUE KEY uk_payment_no (payment_no),
  UNIQUE KEY uk_business_order_no (business_order_no)  -- 一订单一支付单
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;