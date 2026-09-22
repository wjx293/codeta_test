# PayWebServer 数据库设计

数据库：`template_order_db`，MySQL 8.0，InnoDB，utf8mb4

本服务管理 `pay_*` 前缀表。与 `TemplateOrderServer` 的 `tos_*` 表共用同一库，按表前缀划分归属。

## 表清单


| 序号  | 表名              | 用途          | DDL 文件                     |
| --- | --------------- | ----------- | -------------------------- |
| 001 | `pay_payments`  | 支付单         | `001_create_payments.sql`  |
| 002 | `pay_callbacks` | 支付回调流水      | `002_create_callbacks.sql` |
| 003 | `pay_outbox`    | 支付事件 Outbox | `003_create_outbox.sql`    |


## 通用约定

- 金额字段：`BIGINT`，单位「分」
- 状态字段：`TINYINT`
- 时间字段：`DATETIME(3)`，毫秒精度
- 字符集：`utf8mb4`，排序规则：`utf8mb4_unicode_ci`
- 引擎：`InnoDB`

启动时 `AutoMigrate` 可自动建表；也可执行 `migrations/*.sql`。

---

## pay_payments - 支付单


| 字段                | 类型                 | 说明                             |
| ----------------- | ------------------ | ------------------------------ |
| id                | BIGINT UNSIGNED PK | 自增主键                           |
| payment_no        | VARCHAR(32)        | 支付单号，`PW` + yyyyMMdd + 10 位序列  |
| business_order_no | VARCHAR(32)        | 业务订单号，关联 `tos_orders.order_no` |
| amount            | BIGINT             | 金额（分）                          |
| status            | TINYINT            | `1`=未支付，`2`=已支付                |
| created_at        | DATETIME(3)        | 创建时间                           |
| updated_at        | DATETIME(3)        | 更新时间                           |


**唯一约束**：


| 索引                     | 字段                | 说明          |
| ---------------------- | ----------------- | ----------- |
| `uk_payment_no`        | payment_no        | 支付单号唯一      |
| `uk_business_order_no` | business_order_no | **一订单一支付单** |


**业务规则**：

- 创建支付单时 `status=1`
- 仅 `success` 回调且金额一致时，条件更新 `1→2`
- 相同 `business_order_no` 重复创建返回已有记录（幂等）

---

## pay_callbacks - 支付回调流水


| 字段          | 类型                 | 说明                 |
| ----------- | ------------------ | ------------------ |
| id          | BIGINT UNSIGNED PK | 自增主键               |
| callback_no | VARCHAR(32)        | 回调流水号（幂等键）         |
| payment_no  | VARCHAR(32)        | 关联支付单号             |
| amount      | BIGINT             | 回调金额（分）            |
| result      | VARCHAR(16)        | `success` / `fail` |
| created_at  | DATETIME(3)        | 创建时间               |


**唯一约束**：


| 索引               | 字段          | 说明             |
| ---------------- | ----------- | -------------- |
| `uk_callback_no` | callback_no | 回调幂等，重复回调不重复处理 |


**普通索引**：


| 索引               | 字段         | 说明         |
| ---------------- | ---------- | ---------- |
| `idx_payment_no` | payment_no | 按支付单查询回调记录 |


**业务规则**：

- 每次回调先插入流水；`uk_callback_no` 冲突视为重复回调
- `fail` 回调仅记流水，不更新支付单、不发 Kafka

---

## pay_outbox - 支付事件 Outbox


| 字段         | 类型                 | 说明                      |
| ---------- | ------------------ | ----------------------- |
| id         | BIGINT UNSIGNED PK | 自增主键                    |
| event_key  | VARCHAR(64)        | 事件幂等键（等于 `callback_no`） |
| payload    | TEXT               | `PaymentEvent` JSON     |
| status     | TINYINT            | `0`=待发送，`1`=已发送         |
| created_at | DATETIME(3)        | 创建时间                    |
| updated_at | DATETIME(3)        | 更新时间                    |


**唯一约束**：


| 索引             | 字段        | 说明         |
| -------------- | --------- | ---------- |
| `uk_event_key` | event_key | 防止重复写入同一事件 |


**普通索引**：


| 索引                   | 字段                 | 说明            |
| -------------------- | ------------------ | ------------- |
| `idx_status_created` | status, created_at | Relay 扫描待发送记录 |


**业务规则**：

- 与支付单状态更新在同一事务中写入
- Kafka 发送成功后 `status→1`
- 发送失败时保持 `0`，后台 Relay（每 5 秒）补发

**payload 示例**：

```json
{
  "msg_key": "CB202608110000000001",
  "payment_no": "PW202608110000000001",
  "business_order_no": "TO202608110000000001",
  "amount": 9900,
  "result": "success",
  "callback_no": "CB202608110000000001"
}
```

---

## 与 tos_orders 的关系

```
tos_orders.order_no  ←── business_order_no ──→  pay_payments
        │                                              │
        │         Kafka PaymentEvent                   │
        └──────── business_order_no ──────────────────┘
                          │
                          ▼
              TemplateOrderServer 消费后
              条件更新 tos_orders.status 1→2
```

- PayWebServer **不** 直接修改 `tos_orders`
- 订单状态变更由 OrderServer 消费 Kafka 后 `UpdateStatusConditionally(order_no, 1, 2)` 完成
- 已取消订单（`status=3`）不会被更新

---

## 迁移与维护

```bash
# 手动执行（可选，启动时 AutoMigrate 已覆盖）
mysql -u root -p template_order_db < migrations/001_create_payments.sql
mysql -u root -p template_order_db < migrations/002_create_callbacks.sql
mysql -u root -p template_order_db < migrations/003_create_outbox.sql
```

或使用项目脚本：

```bash
cd PayWebServer
go run ./scripts/migrate.go
```

