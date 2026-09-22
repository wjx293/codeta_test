# 模板商城微服务系统

> 基于 Speckit 规范驱动开发（SDD）实现。代码是规范的产物。

## 项目简介

模板商城是一个微服务架构的全栈项目，提供模板上传、浏览、下载、付费购买、会员管理等能力。系统由 6 个独立服务组成，遵循严格的微服务边界与数据库级幂等设计。

## 技术栈


| 层      | 技术                                      |
| ------ | --------------------------------------- |
| 后端语言   | Go 1.22                                 |
| Web 框架 | Gin                                     |
| RPC    | gRPC + protobuf3                        |
| ORM    | GORM（软删除使用 `gorm.DeletedAt`）            |
| 消息队列   | Kafka（segmentio/kafka-go）               |
| 数据库    | MySQL 8.0（单库 `template_order_db`，表归属划分） |
| 对象存储   | 阿里云 OSS（私有 Bucket + STS 临时凭证）           |
| C 端认证  | JWT（access 2h + refresh 7d）             |
| B 端认证  | OAuth2（WPS 开放平台 / Mock Auth）            |
| 前端     | React + Vite + TypeScript               |
| 容器编排   | Docker Compose                          |


## 服务拓扑与端口


| 服务                     | 端口   | 职责                               |
| ---------------------- | ---- | -------------------------------- |
| TemplateFrontend       | 3000 | C 端前端（React）                     |
| TemplateAdminFrontend  | 3001 | B 端前端（React）                     |
| TemplateWebServer      | 8080 | C 端 BFF（Gin + JWT + gRPC 客户端）    |
| TemplateAdminWebServer | 8081 | B 端 BFF（Gin + OAuth2 + gRPC 客户端） |
| PayWebServer           | 8083 | 支付服务（Gin，HTTP only，无 gRPC）       |
| TemplateOrderServer    | 9001 | 核心业务服务（gRPC only，无 HTTP）         |


## 通信矩阵

- C 端 BFF → TemplateOrderServer（gRPC）
- B 端 BFF → TemplateOrderServer（gRPC）
- TemplateOrderServer → PayWebServer（HTTP）
- PayWebServer → TemplateOrderServer（Kafka）

## 项目结构

```
template-mall/
├── TemplateFrontend/         (3000, C 端前端)
├── TemplateWebServer/        (8080, C 端 BFF, README + api.md)
├── TemplateAdminFrontend/    (3001, B 端前端)
├── TemplateAdminWebServer/   (8081, B 端 BFF, README + api.md)
├── TemplateOrderServer/      (9001, gRPC 核心, README + db.md)
├── PayWebServer/             (8083, 支付服务, README + api.md + db.md)
├── Deployments/              (docker-compose.yml)
├── docs/                     (quickstart、OSS、RAM 配置指引)
├── 全栈考核：模板商城微服务项目作业.md  (需求与验收说明)
└── README.md
```

## 启动指引

### 前置依赖

- Docker 24+ & Docker Compose v2
- Go 1.22+（本地开发）
- Node.js 18+ & pnpm（前端开发）
- protoc + protoc-gen-go + protoc-gen-go-grpc（生成 gRPC 代码）

### 一键启动（Docker Compose）

```bash
cd Deployments
cp .env.example .env
docker-compose up -d --build
```

详细步骤见 [docs/quickstart.md](docs/quickstart.md) 与 [Deployments/README.md](Deployments/README.md)。

### 各服务独立启动

详见各服务目录下的 `README.md` 与 `.env.example`。

### 测试

**单元测试**（各 Go 服务目录下执行）：

```bash
cd PayWebServer && go test ./...
cd TemplateOrderServer && go test ./...
cd TemplateWebServer && go test ./...
cd TemplateAdminWebServer && go test ./...
```

**联调与验收**：按 [docs/quickstart.md](docs/quickstart.md) 第 7 节「演示流程」手工验证；完整场景清单见 [全栈考核：模板商城微服务项目作业.md](全栈考核：模板商城微服务项目作业.md) 第二十五章。

## 文档索引


| 文档                                                           | 说明                  |
| ------------------------------------------------------------ | ------------------- |
| [全栈考核：模板商城微服务项目作业.md](全栈考核：模板商城微服务项目作业.md)                   | 项目需求、架构约束与验收场景      |
| [docs/quickstart.md](docs/quickstart.md)                     | 从零启动与演示流程           |
| [docs/oss-setup.md](docs/oss-setup.md)                       | 阿里云 OSS 与 CORS 配置   |
| [Deployments/README.md](Deployments/README.md)               | Docker Compose 部署说明 |
| 各服务 `README.md`                                              | 单服务职责、环境变量与联调方式     |
| `TemplateWebServer/api.md` / `TemplateAdminWebServer/api.md` | BFF HTTP API        |
| `TemplateOrderServer/db.md` / `PayWebServer/db.md`           | 数据库表设计              |


