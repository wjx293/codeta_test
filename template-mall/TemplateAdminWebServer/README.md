# TemplateAdminWebServer

B 端 BFF 服务，端口 **8081**。基于 Gin，负责管理员认证、OSS STS 签发与 gRPC 调用 `TemplateOrderServer`。

## 职责

- B 端管理员认证（WPS OAuth2 / Mock Auth），会话 Cookie `admin_sid`
- 管理员白名单校验（`ADMIN_USER_IDS`）
- 签发 OSS STS 临时凭证，前端直传对象存储
- HeadObject 确认上传对象存在（不中转文件内容）
- 模板创建/编辑时可选自动提取缩略图
- 通过 gRPC 完成模板、订单、用户/会员管理
- **不** 直连数据库；**不** 签发 Admin JWT

## 前置依赖

- Go 1.22+
- `TemplateOrderServer` 已启动（默认 `localhost:9001`）
- 上传功能需配置阿里云 OSS（见 `docs/oss-setup.md`）
- WPS 正式登录需 WPS 开放平台应用（Mock 模式可跳过）

## 环境变量

复制 `.env.example` 为 `.env`：


| 变量                      | 默认值                                       | 说明                   |
| ----------------------- | ----------------------------------------- | -------------------- |
| `HTTP_PORT`             | `8081`                                    | HTTP 监听端口            |
| `GRPC_TARGET`           | `localhost:9001`                          | OrderServer gRPC 地址  |
| `AUTH_PROVIDER`         | `wps`                                     | 认证方式：`wps` 或 `mock`  |
| `WPS_CLIENT_ID`         | —                                         | WPS 应用 ID            |
| `WPS_CLIENT_SECRET`     | —                                         | WPS 应用密钥（仅服务端）       |
| `WPS_REDIRECT_URI`      | `http://localhost:3001/api/auth/callback` | OAuth 回调地址           |
| `WPS_SCOPE`             | `kso.user_base.read`                      | OAuth scope          |
| `WPS_AUTH_URL`          | WPS 授权页 URL                               | —                    |
| `WPS_TOKEN_URL`         | WPS token URL                             | —                    |
| `WPS_USERINFO_URL`      | WPS 用户信息 URL                              | —                    |
| `WPS_KSO_SIGN_ENABLED`  | `false`                                   | 是否启用 KSO 接口签名        |
| `ADMIN_USER_IDS`        | `10001,10002`                             | 管理员白名单（WPS user_id）  |
| `FRONTEND_HOME_URL`     | `http://localhost:3001/`                  | OAuth 成功后跳转          |
| `MOCK_ADMIN_USER_ID`    | `10001`                                   | Mock 管理员 ID          |
| `MOCK_ADMIN_NAME`       | `MockAdmin`                               | Mock 管理员名称           |
| `OSS_ENDPOINT`          | —                                         | OSS Endpoint         |
| `OSS_ACCESS_KEY_ID`     | —                                         | OSS AccessKey（服务端）   |
| `OSS_ACCESS_KEY_SECRET` | —                                         | OSS Secret（服务端）      |
| `OSS_BUCKET`            | —                                         | Bucket 名称            |
| `OSS_REGION`            | —                                         | 地域                   |
| `STS_ROLE_ARN`          | —                                         | RAM Role ARN（STS 必填） |
| `STS_SESSION_NAME`      | `template-mall-admin`                     | STS 会话名              |


## 启动

```bash
cd TemplateAdminWebServer
cp .env.example .env
# 开发演示建议：AUTH_PROVIDER=mock
go run ./cmd/main.go
```

健康检查：[http://localhost:8081/health](http://localhost:8081/health)

## 与其他服务的关系

```
TemplateAdminFrontend (3001)
    │ HTTP + Cookie (admin_sid)
    ▼
TemplateAdminWebServer (8081)     ← 本服务
    │ gRPC                    │ STS / HeadObject
    ▼                         ▼
TemplateOrderServer (9001)   阿里云 OSS
```

## 主要接口

详见 [api.md](api.md)。


| 分组  | 路径前缀                     | 说明                  |
| --- | ------------------------ | ------------------- |
| 认证  | `/auth/*`, `/api/auth/*` | OAuth 登录、Mock 登录、退出 |
| 上传  | `/api/admin/upload/*`    | STS 凭证、确认上传         |
| 模板  | `/api/admin/templates`   | CRUD、上下架            |
| 订单  | `/api/admin/orders`      | 订单列表筛选              |
| 用户  | `/api/admin/users`       | 检索用户、设置/取消会员        |


## 测试方式

### 单元测试

```bash
cd TemplateAdminWebServer
go test ./...
```

### Mock 模式联调（推荐）

1. `.env` 设置 `AUTH_PROVIDER=mock`
2. 启动 OrderServer 与本服务
3. Mock 登录：

```bash
curl -X POST http://localhost:8081/auth/mock-login -c cookies.txt
curl -b cookies.txt http://localhost:8081/auth/me
```

1. 打开 B 端前端 [http://localhost:3001](http://localhost:3001) 完成上传与模板管理

### WPS OAuth 联调

1. WPS 开放平台配置回调 `http://localhost:8081/auth/callback`
2. `.env` 设置 `AUTH_PROVIDER=wps` 及 WPS 凭证
3. 浏览器访问 `GET /auth/login` 完成授权流程

## 已知限制

- 文件内容不经本服务中转，仅签发 STS 与 HeadObject 校验
- 前端不得获得永久 AccessKey Secret
- 上传凭证短期有效；Object Key 由服务端生成

