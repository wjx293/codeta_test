-- 003: 模板元数据表
-- 用途：存储模板基本信息与 OSS 对象关联，支持软删除

CREATE TABLE IF NOT EXISTS tos_templates (
  id                  BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  template_no         VARCHAR(32) NOT NULL,       -- 业务编号
  name                VARCHAR(128) NOT NULL,
  description         TEXT,
  is_free             TINYINT NOT NULL,           -- 0=付费 1=免费
  price               BIGINT NOT NULL,            -- 分
  status              TINYINT NOT NULL,           -- 0=已下架 1=已上架
  oss_type            VARCHAR(16) NOT NULL,       -- aliyun-oss
  bucket_name         VARCHAR(64) NOT NULL,
  original_filename   VARCHAR(255) NOT NULL,
  file_type           VARCHAR(8) NOT NULL,        -- ppt/pptx/doc/docx（应用层校验）
  file_size           BIGINT UNSIGNED NOT NULL,
  file_oss_key        VARCHAR(255) NOT NULL,
  thumbnail_oss_key   VARCHAR(255),
  thumbnail_file_size BIGINT UNSIGNED,
  created_at          DATETIME(3) NOT NULL,
  updated_at          DATETIME(3) NOT NULL,
  deleted_at          DATETIME(3),
  PRIMARY KEY (id),
  UNIQUE KEY uk_template_no (template_no),
  KEY idx_status_created (status, created_at),
  KEY idx_deleted_at (deleted_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;