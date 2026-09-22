-- 005: Kafka 消费幂等记录表
-- 用途：记录每条已消费的 Kafka 消息，保证消费幂等
-- msg_key 来自支付回调的 payment_no 或 callback_no

CREATE TABLE IF NOT EXISTS tos_kafka_consume_records (
  id                BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  msg_key           VARCHAR(64) NOT NULL,          -- payment_no 或 callback_no
  business_order_no VARCHAR(32) NOT NULL,
  result            VARCHAR(16) NOT NULL,          -- success / already_processed / order_canceled / amount_mismatch
  created_at        DATETIME(3) NOT NULL,
  PRIMARY KEY (id),
  UNIQUE KEY uk_msg_key (msg_key)                  -- Kafka 消费幂等
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;