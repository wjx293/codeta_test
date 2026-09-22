# TemplateOrderServer 数据库设计

数据库：`template_order_db`，MySQL 8.0，InnoDB，utf8mb4

## 表清单（tos_* 前缀，TemplateOrderServer 管理）

| 序号 | 表名 | 用途 | DDL 文件 |
|---|---|---|---|
| 001 | `tos_users` | C 端用户 | `001_create_users.sql` |
| 002 | `tos_refresh_tokens` | Refresh Token 记录 | `002_create_refresh_tokens.sql` |
| 003 | `tos_templates` | 模板元数据 | `003_create_templates.sql` |
| 004 | `tos_orders` | 模板订单 | `004_create_orders.sql` |
| 005 | `tos_kafka_consume_records` | Kafka 消费幂等 | `005_create_kafka_consume_records.sql` |

## 通用约定

- 金额字段：`BIGINT`，单位「分」
- 状态字段：`TINYINT`
- 时间字段：`DATETIME(3)`，毫秒精度
- 默认字符集：`utf8mb4`，排序规则：`utf8mb4_unicode_ci`
- 引擎：`InnoDB`

---

## tos_users - C 端用户

| 字段 | 类型 | 说明 |
|---|---|---|
| id | BIGINT UNSIGNED PK | 自增主键 |
| username | VARCHAR(64) | 用户名 |
| password | VARCHAR(255) | bcrypt 密码哈希 |
| member_status | TINYINT | 0=非会员 1=会员 |
| created_at | DATETIME(3) | 创建时间 |
| updated_at | DATETIME(3) | 更新时间 |

**唯一约束**：
- `uk_username`：用户名唯一

---

## tos_refresh_tokens - Refresh Token

| 字段 | 类型 | 说明 |
|---|---|---|
| id | BIGINT UNSIGNED PK | 自增主键 |
| user_id | BIGINT UNSIGNED | 关联 tos_users.id |
| token_hash | VARCHAR(64) | SHA256(refresh_token) hex |
| expires_at | DATETIME(3) | 过期时间（UTC） |
| revoked | TINYINT | 0=有效 1=已吊销 |
| created_at | DATETIME(3) | 创建时间 |

**唯一约束**：
- `uk_token_hash`：token_hash 唯一，防止重复签发

**普通索引**：
- `idx_user_id`：按用户查询 token
- `idx_expires_at`：定时清理过期 token

**吊销逻辑**（FR-006 登出）：
- `RevokeRefreshTokens` RPC 将该用户所有 `revoked=0` 的 token 置为 `revoked=1`

---

## tos_templates - 模板元数据

| 字段 | 类型 | 说明 |
|---|---|---|
| id | BIGINT UNSIGNED PK | 自增主键 |
| template_no | VARCHAR(32) | 业务编号 |
| name | VARCHAR(128) | 模板名称 |
| description | TEXT | 模板描述 |
| is_free | TINYINT | 0=付费 1=免费 |
| price | BIGINT | 价格（分），免费模板为 0 |
| status | TINYINT | 0=已下架 1=已上架 |
| oss_type | VARCHAR(16) | 对象存储类型（aliyun-oss） |
| bucket_name | VARCHAR(64) | OSS Bucket 名称 |
| original_filename | VARCHAR(255) | 原始文件名 |
| file_type | VARCHAR(8) | ppt/pptx/doc/docx（应用层校验） |
| file_size | BIGINT UNSIGNED | 文件大小（字节） |
| file_oss_key | VARCHAR(255) | 模板文件 OSS Object Key |
| thumbnail_oss_key | VARCHAR(255) | 缩略图 OSS Object Key（可空） |
| thumbnail_file_size | BIGINT UNSIGNED | 缩略图文件大小 |
| created_at | DATETIME(3) | 创建时间 |
| updated_at | DATETIME(3) | 更新时间 |
| deleted_at | DATETIME(3) | 软删除时间（NULL=未删除） |

**唯一约束**：
- `uk_template_no`：业务编号唯一

**普通索引**：
- `idx_status_created`：按状态+创建时间排序（列表查询）
- `idx_deleted_at`：软删除查询（GORM 自动过滤）

**软删除**：`deleted_at IS NOT NULL` 视为已删除，列表查询需过滤

---

## tos_orders - 模板订单（核心）

| 字段 | 类型 | 说明 |
|---|---|---|
| id | BIGINT UNSIGNED PK | 自增主键 |
| order_no | VARCHAR(32) | 订单编号 TO+yyyyMMdd+10位 |
| user_id | BIGINT UNSIGNED | 关联 tos_users.id |
| template_id | BIGINT UNSIGNED | 关联 tos_templates.id |
| order_type | TINYINT | 1=免费 2=会员 3=零售 |
| price_snapshot | BIGINT | 下单时价格快照（分） |
| status | TINYINT | 1=未支付 2=已获得下载资格 3=已取消 |
| month | CHAR(7) NULL | 免费/会员填 yyyyMM，零售填 NULL |
| active_unpaid_key | VARCHAR(64) GENERATED | 生成列，仅零售未支付时生成值 |
| created_at | DATETIME(3) | 创建时间 |
| updated_at | DATETIME(3) | 更新时间 |

**4 类唯一约束**：

| 约束名 | 列 | 用途 | 生效条件 |
|---|---|---|---|
| `uk_order_no` | order_no | 订单编号唯一 | 全局 |
| `uk_free_month` | user_id, template_id, month, order_type | 免费订单同月唯一 | month IS NOT NULL |
| `uk_member_month` | user_id, template_id, month, order_type | 会员订单同月唯一 | month IS NOT NULL |
| `uk_active_unpaid` | active_unpaid_key | 零售订单未支付唯一 | active_unpaid_key IS NOT NULL |

**唯一约束设计原理**：

1. **免费/会员月度唯一**：`uk_free_month` 和 `uk_member_month` 包含 `month` 列，该列在免费/会员订单中填 `"yyyyMM"`，在零售订单中填 `NULL`。MySQL 唯一索引**允许多个 NULL 值共存**，因此零售订单不会触发此约束冲突。

2. **零售订单未支付唯一**：使用生成列 `active_unpaid_key`，仅当 `order_type=3 AND status=1` 时生成 `CONCAT(user_id, '_', template_id)`，否则为 `NULL`。`uk_active_unpaid` 唯一索引同样允许多个 NULL，因此已取消/已支付的零售订单不冲突。

3. **零售订单可重购**：同一 user×template 的零售订单，只要前一笔已取消(status=3)或已支付(status=2)，`active_unpaid_key` 即为 NULL，允许创建新的未支付订单。

**普通索引**：
- `idx_user_status`：按用户+状态查询订单
- `idx_template`：按模板查询订单

**状态流转**：
```
1（未支付）→ 支付成功 → 2（已获得下载资格）
1（未支付）→ 取消     → 3（已取消）
```

---

## tos_kafka_consume_records - Kafka 消费幂等

| 字段 | 类型 | 说明 |
|---|---|---|
| id | BIGINT UNSIGNED PK | 自增主键 |
| msg_key | VARCHAR(64) | 消息唯一标识（payment_no 或 callback_no） |
| business_order_no | VARCHAR(32) | 关联业务订单号 |
| result | VARCHAR(16) | success / already_processed / order_canceled / amount_mismatch |
| created_at | DATETIME(3) | 消费时间 |

**唯一约束**：
- `uk_msg_key`：消息 key 唯一，防止重复消费

**幂等机制**：消费 Kafka 消息时，先 INSERT 该记录，若 `uk_msg_key` 冲突则视为已处理，跳过业务逻辑。

---

## 执行方式

```bash
# 启动 MySQL
docker-compose up -d mysql

# 等待 MySQL 就绪
docker-compose exec mysql mysqladmin ping -h localhost -u root -p<password>

# 依次执行 migration
mysql -h127.0.0.1 -uroot -p<password> template_order_db < TemplateOrderServer/migrations/001_create_users.sql
mysql -h127.0.0.1 -uroot -p<password> template_order_db < TemplateOrderServer/migrations/002_create_refresh_tokens.sql
mysql -h127.0.0.1 -uroot -p<password> template_order_db < TemplateOrderServer/migrations/003_create_templates.sql
mysql -h127.0.0.1 -uroot -p<password> template_order_db < TemplateOrderServer/migrations/004_create_orders.sql
mysql -h127.0.0.1 -uroot -p<password> template_order_db < TemplateOrderServer/migrations/005_create_kafka_consume_records.sql

# 验证表
mysql -h127.0.0.1 -uroot -p<password> -e "SHOW TABLES;" template_order_db

# 验证 tos_orders 索引
mysql -h127.0.0.1 -uroot -p<password> -e "SHOW INDEX FROM tos_orders;" template_order_db
```