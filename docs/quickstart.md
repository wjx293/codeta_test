# 模板商城 - 快速启动指南

从零启动模板商城微服务系统（本地 Docker Compose 或混合开发模式）。

---

## 1. 环境准备

| 工具 | 版本建议 |
|------|----------|
| Docker Desktop | 24+ |
| Docker Compose | v2 |
| Go | 1.22+（本地开发后端） |
| Node.js | 18+（本地开发前端） |
| curl + jq | E2E 测试 |

可选：阿里云 OSS 私有 Bucket、WPS 开放平台应用（B 端 OAuth）。

---

## 2. 克隆与配置

```bash
cd themproject/template-mall
```

### 2.1 Docker Compose 环境变量

```bash
cd Deployments
cp .env.example .env
# 编辑 .env：JWT_SECRET、OSS_*、WPS_* 等
```

### 2.2 各服务 .env（本地 `go run` 开发）

```bash
cp TemplateOrderServer/.env.example TemplateOrderServer/.env
cp TemplateAdminWebServer/.env.example TemplateAdminWebServer/.env
cp PayWebServer/.env.example PayWebServer/.env
# 填写 MYSQL_DSN、OSS、WPS 等
```

> `TemplateOrderServer` 与 `PayWebServer` 启动时会 **AutoMigrate** 建表（作业允许）。也可手动执行 `migrations/*.sql`。

---

## 3. 一键启动（Docker Compose）

```bash
cd Deployments
docker-compose up -d --build
docker-compose ps
```

等待全部服务 `healthy` 后访问：

| 服务 | URL |
|------|-----|
| C 端前端 | http://localhost:3000 |
| B 端前端 | http://localhost:3001 |
| C 端 API | http://localhost:8080/health |
| B 端 API | http://localhost:8081/health |
| 支付服务 | http://localhost:8083/health |
| Kafka UI | http://localhost:8090 |

停止：

```bash
docker-compose down        # 保留数据卷
docker-compose down -v     # 清空 MySQL/Kafka 数据
```

---

## 4. 混合开发模式（推荐日常调试）

仅启动基础设施：

```bash
cd Deployments
docker-compose up -d mysql kafka kafka-ui
```

各服务终端启动：

```bash
# 终端 1 - 订单服务
cd TemplateOrderServer && go run ./cmd/main.go

# 终端 2 - 支付服务
cd PayWebServer && go run ./cmd/main.go

# 终端 3 - C 端 BFF
cd TemplateWebServer && go run ./cmd/main.go

# 终端 4 - B 端 BFF
cd TemplateAdminWebServer && go run ./cmd/main.go

# 终端 5 - C 端前端
cd TemplateFrontend && npm install && npm run dev

# 终端 6 - B 端前端
cd TemplateAdminFrontend && npm install && npm run dev
```

---

## 5. WPS OAuth 与 Mock Auth

### Mock Auth（默认，无需 WPS）

`TemplateAdminWebServer` 设置 `AUTH_PROVIDER=mock`，B 端前端点击登录调用 `POST /auth/mock-login`。

默认管理员：`MOCK_ADMIN_USER_ID=10001`，白名单 `ADMIN_USER_IDS=10001,10002`。

### WPS OAuth

1. [WPS 开放平台](https://365.kdocs.cn/3rd/open/developer/) 配置回调：`http://localhost:8081/auth/callback`
2. 开通权限 `kso.user_base.read`
3. `.env` 设置：
   ```env
   AUTH_PROVIDER=wps
   WPS_CLIENT_ID=你的应用ID
   WPS_CLIENT_SECRET=你的密钥
   WPS_REDIRECT_URI=http://localhost:8081/auth/callback
   ADMIN_USER_IDS=你的WPS用户ID
   ```

---

## 6. 阿里云 OSS

1. 创建**私有** Bucket
2. 配置 CORS（允许 `http://localhost:3001`，方法 PUT/GET/HEAD）
3. 在两个后端 `.env` 填写 `OSS_*`
4. 详见 [docs/oss-setup.md](oss-setup.md)

---

## 7. 演示流程

### C 端

1. 打开 http://localhost:3000 → 注册/登录
2. 浏览模板列表 → 下载免费模板
3. 下载付费模板 → 获得 `payment_no`
4. 模拟支付回调：
   ```bash
   curl -X POST http://localhost:8083/api/paycallback \
     -H 'Content-Type: application/json' \
     -d '{"payment_no":"PW...","amount":9900,"result":"success","callback_no":"CB001"}'
   ```
5. 再次下载 → 获得 5 分钟预签名 URL

### B 端

1. 打开 http://localhost:3001 → Mock 登录
2. 上传模板（STS 直传 OSS）→ 创建 → 上架
3. 订单列表筛选 / 用户会员管理

---

## 8. E2E 自动化测试

```bash
cd tests/e2e
chmod +x *.sh   # Linux/macOS/Git Bash
./run_all.sh
```

详见 [tests/e2e/README.md](../tests/e2e/README.md)。

---

## 9. 常见问题

| 问题 | 处理 |
|------|------|
| MySQL 连接拒绝 | `docker-compose up -d mysql`，确认 3306 端口 |
| Kafka 镜像拉取失败 | 使用 `apache/kafka:3.7.0`（已配置）；或临时关闭镜像加速 |
| OSS 上传 XHR -1 | Bucket 配置 CORS；前端 region 与 Bucket 地域一致 |
| 表不存在 | 重启 OrderServer/PayWebServer 触发 AutoMigrate，或手动跑 migrations |
