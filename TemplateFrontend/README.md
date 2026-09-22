# TemplateFrontend

C 端用户前端，端口 **3000**。基于 React + Vite + TypeScript，对接 `TemplateWebServer`（8080）与 `PayWebServer`（8083，模拟支付回调）。

## 职责

- 用户注册与登录（JWT）
- 浏览上架模板列表
- 下载模板（免费 / 会员 / 零售）
- 零售付费模板的模拟支付流程
- 我的订单查询与取消未支付零售订单

## 前置依赖

- Node.js 18+
- `TemplateWebServer` 已启动（默认 `http://localhost:8080`）
- 零售购买场景需 `PayWebServer`（默认 `http://localhost:8083`）

## 环境变量

复制 `.env.example` 为 `.env`：

| 变量 | 默认值 | 说明 |
|------|--------|------|
| `VITE_API_BASE_URL` | `http://localhost:8080` | C 端 BFF 地址 |
| `VITE_PAY_API_BASE_URL` | `http://localhost:8083` | 支付服务地址（模拟支付回调） |

## 启动

```bash
cd TemplateFrontend
npm install
npm run dev
```

浏览器访问：http://localhost:3000

生产构建：

```bash
npm run build
npm run preview
```

## 与其他服务的关系

```
TemplateFrontend (3000)
    │ HTTP + JWT
    ▼
TemplateWebServer (8080)
    │ gRPC
    ▼
TemplateOrderServer (9001)

零售支付时：
TemplateFrontend ──POST /api/paycallback──▶ PayWebServer (8083)
                                              │ Kafka
                                              ▼
                                    TemplateOrderServer
```

## 主要页面

| 路由 | 功能 |
|------|------|
| `/login` | 用户登录 |
| `/register` | 用户注册 |
| `/` | 模板列表与下载 |
| `/orders` | 我的订单（筛选、取消、重新下载） |

## 下载与支付流程

1. 点击「下载」→ 调用 `POST /templates/:id/download`
2. 若返回 `download_url` → 直接打开下载链接
3. 若返回 `payment_no`（零售未支付）→ 弹出模拟支付对话框
4. 确认支付 → 调用 `PayWebServer` 的 `POST /api/paycallback`
5. 等待 Kafka 消费完成后自动重试下载

## 测试方式

1. 启动后端服务（MySQL、Kafka、OrderServer、PayWebServer、WebServer）
2. 注册新用户并登录
3. 在 B 端上架至少一个免费模板和一个付费模板
4. 验证免费下载、零售购买支付、订单取消等流程

## 已知限制

- 缩略图依赖后端返回的预签名 URL，开发环境可能为占位链接
- 支付成功后依赖 Kafka 消费，可能有 1–5 秒延迟
- 未实现独立模板详情页（作业不强制要求）
