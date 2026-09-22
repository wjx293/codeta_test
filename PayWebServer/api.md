# PayWebServer API 文档

支付服务，端口 `8083`。仅 HTTP，无鉴权（作业范围内为内网/开发模拟支付）。

## 通用约定

- 请求/响应格式：`application/json`
- 错误响应：`{"code": <int>, "message": <string>}`
- `code=0` 表示成功
- 金额字段单位：分（`int64`）
- 支付单状态：`1`=未支付，`2`=已支付

---

## GET /health

健康检查。

**响应示例**：

```json
{
  "service": "PayWebServer",
  "version": "1.0.0"
}
```

---

## POST /api/payments

创建支付单。由 `TemplateOrderServer` 在零售下载时调用；也可手工测试。

**请求体**：

```json
{
  "business_order_no": "TO202608110000000001",
  "amount": 9900
}
```

| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| `business_order_no` | string | 是 | 业务订单号（`tos_orders.order_no`） |
| `amount` | int64 | 是 | 金额（分） |

**成功响应**（`code=0`）：

```json
{
  "code": 0,
  "data": {
    "payment_no": "PW202608110000000001",
    "business_order_no": "TO202608110000000001",
    "amount": 9900,
    "status": 1
  }
}
```

**幂等**：相同 `business_order_no` 重复请求返回同一 `payment_no`（`uk_business_order_no`）。

**错误**：

| HTTP | code | 说明 |
|------|------|------|
| 400 | 400 | 参数缺失或无效 |
| 500 | 500 | 数据库等内部错误 |

---

## POST /api/paycallback

支付回调（模拟第三方支付结果）。C 端前端在「模拟支付」对话框中调用。

**请求体**：

```json
{
  "payment_no": "PW202608110000000001",
  "business_order_no": "TO202608110000000001",
  "amount": 9900,
  "result": "success",
  "callback_no": "CB202608110000000001"
}
```

| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| `payment_no` | string | 是 | 支付单号 |
| `business_order_no` | string | 否 | 业务订单号（可选） |
| `amount` | int64 | 是 | 回调金额（分），须与支付单一致 |
| `result` | string | 是 | `success` 或 `fail` |
| `callback_no` | string | 是 | 回调流水号（幂等键） |

**成功响应**：

```json
{
  "code": 0,
  "message": "success"
}
```

### `result=success` 处理逻辑

1. 插入回调流水（`uk_callback_no` 幂等）
2. 校验支付单存在、状态为未支付（`1`）、金额一致
3. 事务：支付单 `1→2` + 写入 `pay_outbox`
4. 发送 Kafka 消息到 `template-pay-events`
5. 标记 Outbox 已发送；若 Kafka 失败返回 500，Outbox Relay 后续补发

### `result=fail` 处理逻辑

1. 仅记录回调流水
2. **不** 更新支付单状态
3. **不** 发送 Kafka

### 重复回调

相同 `callback_no` 视为重复：直接返回成功，不重复发 Kafka（`success` 且支付单仍未支付时会尝试补全状态）。

**错误**：

| HTTP | code | 说明 |
|------|------|------|
| 400 | 400 | 参数错误、`result` 非法、金额不一致、支付单状态非法 |
| 404 | 404 | 支付单不存在 |
| 500 | 500 | Kafka 发送失败等 |

---

## Kafka 消息格式

Topic：`template-pay-events`（可配置）

- **Key**：`business_order_no`（业务订单号）
- **Value**：JSON

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

| 字段 | 说明 |
|------|------|
| `msg_key` | 消费幂等键，等于 `callback_no` |
| `payment_no` | 支付单号 |
| `business_order_no` | 业务订单号 |
| `amount` | 金额（分） |
| `result` | 固定 `success`（仅成功回调会发消息） |
| `callback_no` | 回调流水号 |

消费方：`TemplateOrderServer`，消费组 `template-order-consumer-group`。

---

## 调用方说明

| 调用方 | 接口 | 场景 |
|--------|------|------|
| TemplateOrderServer | `POST /api/payments` | 零售下载创建未支付订单后建支付单 |
| TemplateFrontend | `POST /api/paycallback` | C 端模拟支付确认/取消 |

---

## 测试示例

```bash
# 1. 创建支付单
curl -s -X POST http://localhost:8083/api/payments \
  -H 'Content-Type: application/json' \
  -d '{"business_order_no":"TO202608110000000001","amount":9900}'

# 2. 支付成功
curl -s -X POST http://localhost:8083/api/paycallback \
  -H 'Content-Type: application/json' \
  -d '{
    "payment_no": "PW202608110000000001",
    "amount": 9900,
    "result": "success",
    "callback_no": "CB202608110000000001"
  }'

# 3. 重复回调（应幂等，不重复发 Kafka）
curl -s -X POST http://localhost:8083/api/paycallback \
  -H 'Content-Type: application/json' \
  -d '{
    "payment_no": "PW202608110000000001",
    "amount": 9900,
    "result": "success",
    "callback_no": "CB202608110000000001"
  }'

# 4. 支付失败
curl -s -X POST http://localhost:8083/api/paycallback \
  -H 'Content-Type: application/json' \
  -d '{
    "payment_no": "PW202608110000000001",
    "amount": 9900,
    "result": "fail",
    "callback_no": "CB202608110000000002"
  }'
```
