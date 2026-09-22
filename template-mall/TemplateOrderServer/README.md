# TemplateOrderServer

核心业务服务，gRPC 端口 **9001**。无 HTTP 接口，负责用户、模板、订单、会员及 Kafka 消费。

## 职责

- C 端用户注册/登录、JWT 签发与 Refresh Token 管理
- 模板元数据 CRUD、上下架、软删除
- 下载资格校验与订单创建（免费 / 会员 / 零售）
- 订单列表、取消未支付零售订单
- B 端用户检索、会员状态设置
- 调用 PayWebServer 创建支付单（HTTP）
- 消费 Kafka 支付成功事件，更新订单状态
- OSS 预签名下载 URL 生成（5 分钟有效）
- 数据库级幂等（唯一约束 + 条件更新 + Kafka msg_key）

## 前置依赖

- Go 1.22+
- MySQL 8.0（库名 `template_order_db`）
- Kafka（topic `template-pay-events`）
- PayWebServer（默认 `http://127.0.0.1:8083`）
- OSS 配置（下载与缩略图预签名；不配置时相关功能受限）

## 环境变量

复制 `.env.example` 为 `.env`：


| 变量                      | 默认值                                                         | 说明                     |
| ----------------------- | ----------------------------------------------------------- | ---------------------- |
| `GRPC_PORT`             | `9001`                                                      | gRPC 监听端口              |
| `MYSQL_DSN`             | `root:root123456@tcp(127.0.0.1:3306)/template_order_db?...` | MySQL 连接串              |
| `JWT_SECRET`            | `change-me-in-production`                                   | JWT 签名密钥               |
| `JWT_ACCESS_TTL`        | `2h`                                                        | Access Token 有效期       |
| `JWT_REFRESH_TTL`       | `168h`                                                      | Refresh Token 有效期（7 天） |
| `KAFKA_BROKERS`         | `127.0.0.1:9092`                                            | Kafka 地址               |
| `KAFKA_TOPIC`           | `template-pay-events`                                       | 支付事件 Topic             |
| `KAFKA_GROUP_ID`        | `template-order-consumer-group`                             | 消费组 ID                 |
| `PAY_WEB_SERVER_URL`    | `http://127.0.0.1:8083`                                     | PayWebServer 地址        |
| `OSS_ENDPOINT`          | —                                                           | OSS Endpoint           |
| `OSS_ACCESS_KEY_ID`     | —                                                           | OSS AccessKey          |
| `OSS_ACCESS_KEY_SECRET` | —                                                           | OSS Secret             |
| `OSS_BUCKET`            | —                                                           | Bucket 名称              |
| `OSS_REGION`            | —                                                           | 地域                     |




## 启动

```bash
cd TemplateOrderServer
cp .env.example .env
go run ./cmd/main.go
```

启动时会 **AutoMigrate** 建表（作业允许）。也可手动执行 `migrations/*.sql`。

gRPC 反射已开启，可用 grpcurl 调试：

```bash
grpcurl -plaintext localhost:9001 list
grpcurl -plaintext localhost:9001 list template_order.TemplateOrderService
```

## 与其他服务的关系

```
TemplateWebServer (8080)  ──gRPC──▶ TemplateOrderServer (9001)  ← 本服务
TemplateAdminWebServer (8081) ──gRPC──▶
                                          │
                    HTTP POST /api/payments
                                          ▼
                                    PayWebServer (8083)
                                          │ Kafka produce
                                          ▼
                                    Kafka (template-pay-events)
                                          │ consume
                                          ▼
                                    TemplateOrderServer (本服务)
```

## gRPC 接口

定义见 `api/proto/template_order.proto`。主要 RPC：


| 分组  | RPC                                                                                        | 说明      |
| --- | ------------------------------------------------------------------------------------------ | ------- |
| 认证  | `Register`, `Login`, `RefreshToken`, `RevokeRefreshTokens`                                 | C 端用户认证 |
| 用户  | `ListUsers`                                                                                | B 端检索用户 |
| 模板  | `CreateTemplate`, `UpdateTemplate`, `ListTemplates`, `GetTemplate`, `UpdateTemplateStatus` | 模板管理    |
| 订单  | `DownloadTemplate`, `ListOrders`, `CancelOrder`                                            | 下载与订单   |
| 会员  | `SetMember`                                                                                | 设置/取消会员 |


生成代码：`buf generate`（需安装 buf / protoc 工具链）。

## 数据库

本服务管理 `tos_*` 表，详见 [db.md](db.md)。


| 表                           | 用途            |
| --------------------------- | ------------- |
| `tos_users`                 | C 端用户         |
| `tos_refresh_tokens`        | Refresh Token |
| `tos_templates`             | 模板元数据         |
| `tos_orders`                | 订单            |
| `tos_kafka_consume_records` | Kafka 消费幂等    |


PayWebServer 管理 `pay_*` 表，见 `PayWebServer/db.md`。

## 测试方式

### 单元测试

```bash
cd TemplateOrderServer
go test ./...
```

### 手工联调

1. 启动 MySQL、Kafka、PayWebServer
2. 启动本服务，确认日志 `database connected and migrated`
3. 通过 C/B BFF 或 grpcurl 调用 RPC
4. 零售支付：C 端下载付费模板 → PayWeb 回调 → 观察 Kafka 消费日志 → 再次下载获 URL

### 支付事件消费验证

```bash
# 模拟支付成功（需先有零售未支付订单与 payment_no）
curl -X POST http://localhost:8083/api/paycallback \
  -H 'Content-Type: application/json' \
  -d '{"payment_no":"PW...","amount":9900,"result":"success","callback_no":"CB202608110000000001"}'
```

观察 OrderServer 日志中 Kafka 消费与订单状态更新。

## 已知限制

- 仅 gRPC，无 HTTP API
- 会员无购买/续费流程，仅管理员设置当前状态
- 时区使用 `Asia/Shanghai`（月度免费/会员订单按 yyyyMM 计）

