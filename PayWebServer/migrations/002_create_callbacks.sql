-- 002: 支付回调流水表
-- 用途：记录每次支付回调，保证回调幂等

CREATE TABLE IF NOT EXISTS pay_callbacks (
  id            BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  callback_no   VARCHAR(32) NOT NULL,            -- CB+yyyyMMdd+10位 或外部传入
  payment_no    VARCHAR(32) NOT NULL,
  amount        BIGINT NOT NULL,
  result        VARCHAR(16) NOT NULL,            -- success / fail
  created_at    DATETIME(3) NOT NULL,
  PRIMARY KEY (id),
  UNIQUE KEY uk_callback_no (callback_no),       -- 回调幂等
  KEY idx_payment_no (payment_no)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;