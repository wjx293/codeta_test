-- 002: Refresh Token 记录表
-- 用途：JWT refresh token 存储与吊销，支持 C 端登出（FR-006）

CREATE TABLE IF NOT EXISTS tos_refresh_tokens (
  id           BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  user_id      BIGINT UNSIGNED NOT NULL,
  token_hash   VARCHAR(64) NOT NULL,           -- SHA256(refresh_token) hex
  expires_at   DATETIME(3) NOT NULL,           -- UTC
  revoked      TINYINT NOT NULL DEFAULT 0,     -- 0=有效 1=已吊销
  created_at   DATETIME(3) NOT NULL,
  PRIMARY KEY (id),
  UNIQUE KEY uk_token_hash (token_hash),
  KEY idx_user_id (user_id),
  KEY idx_expires_at (expires_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;