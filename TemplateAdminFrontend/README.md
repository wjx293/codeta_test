# TemplateAdminFrontend

B 端管理员前端，端口 **3001**。基于 React + Vite + TypeScript，对接 `TemplateAdminWebServer`（8081）。

## 职责

- 管理员登录（Mock / WPS OAuth2）
- 模板上传、创建与编辑
- 设置免费/付费与价格
- 模板上架与下架
- 订单列表查询（多维度筛选）
- 用户检索与会员设置

## 前置依赖

- Node.js 18+
- `TemplateAdminWebServer` 已启动（默认 `http://localhost:8081`）
- Mock 模式：后端 `AUTH_PROVIDER=mock`

## 环境变量

复制 `.env.example` 为 `.env`：

| 变量 | 默认值 | 说明 |
|------|--------|------|
| `VITE_API_BASE_URL` | `http://localhost:8081` | B 端 BFF 地址 |
| `VITE_AUTH_MODE` | `mock` | 认证模式：`mock` 或 `wps` |

## 启动

```bash
cd TemplateAdminFrontend
npm install
npm run dev
```

浏览器访问：http://localhost:3001

生产构建：

```bash
npm run build
npm run preview
```

## 与其他服务的关系

```
TemplateAdminFrontend (3001)
    │ HTTP + Cookie (admin_sid)
    ▼
TemplateAdminWebServer (8081)
    │ gRPC + OSS STS/HeadObject
    ▼
TemplateOrderServer (9001) / 阿里云 OSS
```

## 主要页面

| 路由 | 功能 |
|------|------|
| `/login` | 管理员登录 |
| `/templates` | 模板列表、上架/下架 |
| `/templates/new` | 新建模板（上传 + 创建） |
| `/templates/:id/edit` | 编辑模板信息 |
| `/orders` | 订单列表 |
| `/users` | 用户检索、设置/取消会员 |

## 上传流程

1. `POST /api/admin/upload/credential` 获取 STS 凭证与 Object Key
2. 前端直传 OSS（开发 mock 环境可跳过实际上传）
3. `POST /api/admin/upload/confirm` 确认对象存在
4. `POST /api/admin/templates` 创建模板记录

## 测试方式

1. 启动后端（含 `TemplateAdminWebServer`，`AUTH_PROVIDER=mock`）
2. 访问 http://localhost:3001 ，点击 Mock 登录
3. 新建模板并上架，在 C 端验证可见
4. 在订单/用户页验证管理功能

## 已知限制

- 编辑模板不支持替换模板文件（作业不要求）
- 开发环境 OSS 为 mock，实际上传需配置真实 OSS 并集成 ali-oss SDK
- WPS OAuth 需配置 `.env` 中的 WPS 相关变量与白名单
