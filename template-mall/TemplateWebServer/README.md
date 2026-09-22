# TemplateWebServer

C 端 BFF 服务，端口 **8080**。基于 Gin，通过 gRPC 调用 `TemplateOrderServer` 完成业务逻辑。

## 职责

- 暴露 C 端 HTTP API（注册/登录、模板浏览、下载、订单）
- JWT 鉴权（access token + refresh token）
- 将 HTTP 请求转换为 gRPC 调用，不直连数据库
- 将 gRPC 响应转换为前端 JSON（状态字段保持 0/1/2/3 整数格式）

## 前置依赖

- Go 1.22+
- `TemplateOrderServer` 已启动（默认 `localhost:9001`）
- MySQL、Kafka 由 OrderServer 使用，本服务不直连

## 环境变量

复制 `.env.example` 为 `.env`：


| 变量            | 默认值                        | 说明                          |
| ------------- | -------------------------- | --------------------------- |
| `HTTP_PORT`   | `8080`                     | HTTP 监听端口                   |
| `GRPC_TARGET` | `localhost:9001`           | TemplateOrderServer gRPC 地址 |
| `JWT_SECRET`  | `template-mall-secret-key` | JWT 签名密钥（须与 OrderServer 一致） |


## 启动

```bash
cd TemplateWebServer
cp .env.example .env
go run ./cmd/main.go
```

健康检查：[http://localhost:8080/health](http://localhost:8080/health)

## 与其他服务的关系

```
TemplateFrontend (3000)
    │ HTTP + Authorization: Bearer <token>
    ▼
TemplateWebServer (8080)          ← 本服务
    │ gRPC
    ▼
TemplateOrderServer (9001)
```

本服务 **不** 调用 PayWebServer；零售支付的模拟回调由 C 端前端直接请求 PayWebServer。

## 主要接口

详见 [api.md](api.md)。


| 分组  | 路径前缀         | 说明          |
| --- | ------------ | ----------- |
| 认证  | `/auth/*`    | 注册、登录、刷新、登出 |
| 模板  | `/templates` | 列表、详情、下载    |
| 订单  | `/orders`    | 我的订单、取消     |
| 用户  | `/users/me`  | 当前用户信息      |


## 测试方式

### 单元测试

```bash
cd TemplateWebServer
go test ./...
```

### 手工联调

1. 启动 `TemplateOrderServer` 与 MySQL、Kafka
2. 启动本服务
3. 使用 curl 或 C 端前端验证：

```bash
# 健康检查
curl http://localhost:8080/health

# 注册
curl -X POST http://localhost:8080/auth/register \
  -H 'Content-Type: application/json' \
  -d '{"username":"testuser","password":"password123"}'

# 登录（获取 access_token）
curl -X POST http://localhost:8080/auth/login \
  -H 'Content-Type: application/json' \
  -d '{"username":"testuser","password":"password123"}'
```

1. 带 token 访问模板列表：`GET /templates`

## 已知限制

- 不签发 B 端会话，不处理 OSS 上传
- 下载地址由 OrderServer 生成预签名 URL，本服务透传
- JWT 密钥须与 `TemplateOrderServer` 的 `JWT_SECRET` 保持一致

