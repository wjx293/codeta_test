# TemplateWebServer API 文档

C 端 BFF 服务，端口 `8080`。通过 gRPC 调用 TemplateOrderServer 完成业务逻辑。

## 通用约定

- 请求/响应格式：`application/json`
- 错误响应统一格式：`{"code": <int>, "message": <string>}`
- `code=0` 表示成功，非 0 表示业务错误
- 金额字段单位：分（`int64`）
- 时间字段格式：ISO 8601（`2006-01-02T15:04:05Z07:00`）
- 鉴权方式：`Authorization: Bearer <access_token>`

---

## 系统接口

### GET /health

健康检查（无鉴权）。

**响应示例**：
```json
{
  "service": "TemplateWebServer",
  "version": "1.0.0"
}
```

---

## 认证接口

### POST /auth/register

用户注册（无鉴权）。

**请求体**：
```json
{
  "username": "alice",
  "password": "pass123456"
}
```

**响应示例**：
```json
{
  "code": 0,
  "data": {
    "access_token": "eyJhbGc...",
    "refresh_token": "eyJhbGc...",
    "user_id": 1,
    "username": "alice",
    "member_status": 0
  }
}
```

### POST /auth/login

用户登录（无鉴权）。

**请求体**：
```json
{
  "username": "alice",
  "password": "pass123456"
}
```

**响应示例**：
```json
{
  "code": 0,
  "data": {
    "access_token": "eyJhbGc...",
    "refresh_token": "eyJhbGc...",
    "user_id": 1,
    "username": "alice",
    "member_status": 0
  }
}
```

### POST /auth/refresh

刷新 access token（无鉴权）。

**请求体**：
```json
{
  "refresh_token": "eyJhbGc..."
}
```

**响应示例**：
```json
{
  "code": 0,
  "data": {
    "access_token": "eyJhbGc...",
    "refresh_token": "eyJhbGc...",
    "user_id": 1,
    "username": "alice",
    "member_status": 0
  }
}
```

### POST /auth/logout

用户登出（需 JWT 鉴权）。吊销该用户所有 refresh token。

**响应示例**：
```json
{
  "code": 0,
  "message": "logout success"
}
```

---

## 模板接口

### GET /templates

模板列表（仅上架，需 JWT 鉴权）。

**查询参数**：
| 参数 | 类型 | 默认 | 说明 |
|------|------|------|------|
| page | int | 1 | 页码 |
| page_size | int | 20 | 每页数量（最大 100） |

**响应示例**：
```json
{
  "code": 0,
  "data": {
    "templates": [
      {
        "id": 1,
        "template_no": "TO202608070000000001",
        "name": "企业官网模板",
        "description": "简洁响应式",
        "is_free": 1,
        "price": 0,
        "status": 1,
        "file_type": "zip",
        "file_size": 1048576,
        "thumbnail_download_url": "https://oss.../thumb.jpg?signature=...",
        "created_at": "2026-08-07T10:00:00+08:00",
        "updated_at": "2026-08-07T10:00:00+08:00"
      }
    ],
    "total": 50
  }
}
```

### GET /templates/:id

模板详情（需 JWT 鉴权）。

**响应示例**：
```json
{
  "code": 0,
  "data": {
    "id": 1,
    "template_no": "TO202608070000000001",
    "name": "企业官网模板",
    "description": "简洁响应式",
    "is_free": 1,
    "price": 0,
    "status": 1,
    "file_type": "zip",
    "file_size": 1048576,
    "thumbnail_download_url": "https://oss.../thumb.jpg?signature=...",
    "created_at": "2026-08-07T10:00:00+08:00",
    "updated_at": "2026-08-07T10:00:00+08:00"
  }
}
```

### POST /templates/:id/download

下载模板（触发资格判断，需 JWT 鉴权）。

**响应示例（免费模板 / 已购模板）**：
```json
{
  "code": 0,
  "data": {
    "order_no": "TO202608070000000001",
    "order_type": 1,
    "order_status": 2,
    "price_snapshot": 0,
    "download_url": "https://oss.../file.zip?signature=...",
    "payment_no": ""
  }
}
```

**响应示例（付费模板未购买）**：
```json
{
  "code": 0,
  "data": {
    "order_no": "TO202608070000000002",
    "order_type": 3,
    "order_status": 1,
    "price_snapshot": 9900,
    "download_url": "",
    "payment_no": "PW202608070000000001"
  }
}
```

---

## 订单接口

### GET /orders

我的订单列表（需 JWT 鉴权）。

**查询参数**：
| 参数 | 类型 | 默认 | 说明 |
|------|------|------|------|
| status | int | -1 | 订单状态：-1=全部 1=未支付 2=已获得下载资格 3=已取消 |
| page | int | 1 | 页码 |
| page_size | int | 20 | 每页数量 |

**响应示例**：
```json
{
  "code": 0,
  "data": {
    "orders": [
      {
        "id": 1,
        "order_no": "TO202608070000000001",
        "user_id": 1,
        "username": "alice",
        "template_id": 1,
        "template_name": "企业官网模板",
        "order_type": 1,
        "price_snapshot": 0,
        "status": 2,
        "created_at": "2026-08-07T10:00:00+08:00",
        "updated_at": "2026-08-07T10:00:00+08:00"
      }
    ],
    "total": 10
  }
}
```

### POST /orders/:order_no/cancel

取消未支付零售订单（需 JWT 鉴权）。

**响应示例**：
```json
{
  "code": 0,
  "message": "order cancelled"
}
```

---

## 用户接口

### GET /users/me

当前用户信息（需 JWT 鉴权）。

**响应示例**：
```json
{
  "code": 0,
  "data": {
    "user_id": 1,
    "username": "alice",
    "member_status": 0
  }
}
```

---

## 错误码

| code | HTTP Status | 说明 |
|------|-------------|------|
| 0 | 200 | 成功 |
| 400 | 400 | 请求参数错误 |
| 401 | 401 | 未鉴权 / token 过期 |
| 404 | 404 | 资源不存在 |
| 409 | 409 | 资源已存在 |
| 500 | 500 | 服务器内部错误 |