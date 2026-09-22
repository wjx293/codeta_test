# PayWebServer

支付服务，端口 **8083**。基于 Gin，**仅 HTTP**，无 gRPC。负责支付单创建、支付回调与 Kafka 事件生产。

## 职责

- 创建支付单（一业务订单对应一支付单，`uk_business_order_no` 幂等）
- 处理支付回调（`success` / `fail`），回调流水幂等（`uk_callback_no`）
- 支付成功时：事务更新支付单状态 + 写入 Outbox + 发送 Kafka 事件
- Outbox Relay 后台任务（每 5 秒补发 Kafka 发送失败的事件）
- **不** 更新模板订单状态（由 TemplateOrderServer 消费 Kafka 完成）

## 前置依赖

- Go 1.22+
- MySQL 8.0（库名 `template_order_db`，表前缀 `pay_`）
- Kafka（topic `template-pay-events`）
- 通常由 `TemplateOrderServer` 通过 HTTP 调用创建支付单；C 端前端直接调用本服务模拟支付回调

## 环境变量

复制 `.env.example` 为 `.env`：


| 变量              | 默认值                                                         | 说明         |
| --------------- | ----------------------------------------------------------- | ---------- |
| `HTTP_PORT`     | `8083`                                                      | HTTP 监听端口  |
| `MYSQL_DSN`     | `root:root123456@tcp(127.0.0.1:3306)/template_order_db?...` | MySQL 连接串  |
| `KAFKA_BROKERS` | `127.0.0.1:9092`                                            | Kafka 地址   |
| `KAFKA_TOPIC`   | `template-pay-events`                                       | 支付事件 Topic |


## 启动

```bash
cd PayWebServer
cp .env.example .env
go run ./cmd/main.go
```

启动时会 **AutoMigrate** 建表。也可手动执行 `migrations/*.sql`。

## 与其他服务的关系

```
TemplateOrderServer (9001)
    │ HTTP POST /api/payments
    ▼
PayWebServer (8083)              ← 本服务
    │ Kafka produce (key=business_order_no)
    ▼
Kafka (template-pay-events)
    │ consume
    ▼
TemplateOrderServer (9001)

C 端模拟支付：
TemplateFrontend ──POST /api/paycallback──▶ PayWebServer (8083)
```

## 主要接口

详见 [api.md](api.md)。


| 方法     | 路径                 | 说明    |
| ------ | ------------------ | ----- |
| `GET`  | `/health`          | 健康检查  |
| `POST` | `/api/payments`    | 创建支付单 |
| `POST` | `/api/paycallback` | 支付回调  |


## 数据库

本服务管理 `pay_*` 表，详见 [db.md](db.md)。


| 表               | 用途          |
| --------------- | ----------- |
| `pay_payments`  | 支付单         |
| `pay_callbacks` | 回调流水（幂等）    |
| `pay_outbox`    | 支付事件 Outbox |


## 业务编号格式


| 类型   | 前缀   | 示例                     |
| ---- | ---- | ---------------------- |
| 支付单号 | `PW` | `PW202608110000000001` |
| 回调编号 | `CB` | `CB202608110000000001` |


格式：`{前缀} + yyyyMMdd + 10 位序列号`。重启后从数据库同步当日最大序列，避免冲突。

## 已知限制

- 无真实第三方支付对接，仅模拟回调
- 金额单位：分（`int64`）
- `fail` 回调不更新支付单状态、不发送 Kafka

