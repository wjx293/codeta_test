-- 003: 支付事件 Outbox 表
-- 用途：支付单状态更新与 Kafka 事件同事务写入，后台 relay 补发

CREATE TABLE IF NOT EXISTS pay_outbox (
  id         BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  event_key  VARCHAR(64) NOT NULL,              -- callback_no，幂等键
  payload    TEXT NOT NULL,                     -- PaymentEvent JSON
  status     TINYINT NOT NULL DEFAULT 0,        -- 0=待发送 1=已发送
  created_at DATETIME(3) NOT NULL,
  updated_at DATETIME(3) NOT NULL,
  PRIMARY KEY (id),
  UNIQUE KEY uk_event_key (event_key),
  KEY idx_status_created (status, created_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
