# Deployments - Docker Compose 编排

## 快速启动

### 一键启动全部服务

```bash
cd Deployments
cp .env.example .env    # 首次：按需填写 OSS / WPS / JWT
docker-compose up -d --build
docker-compose ps
```

全部服务 `healthy` 后：

- C 端：[http://localhost:3000](http://localhost:3000)
- B 端：[http://localhost:3001](http://localhost:3001)
- Kafka UI：[http://localhost:8090](http://localhost:8090)

### 仅启动基础设施（本地混合开发）

```bash
cd Deployments
docker-compose up -d mysql kafka kafka-ui
```

然后在各服务目录 `go run` / `npm run dev`。详见 [docs/quickstart.md](../docs/quickstart.md)。

### 停止与清理

```bash
docker-compose down       # 停止，保留数据卷
docker-compose down -v    # 停止并清空 MySQL/Kafka 数据
```

## 服务端口映射


| 服务                        | 容器端口 | 主机端口 | 用途           |
| ------------------------- | ---- | ---- | ------------ |
| mysql                     | 3306 | 3306 | MySQL 数据库    |
| kafka                     | 9092 | 9092 | Kafka（宿主机访问） |
| kafka-ui                  | 8090 | 8090 | Kafka 管理界面   |
| template-order-server     | 9001 | 9001 | gRPC 核心业务    |
| pay-web-server            | 8083 | 8083 | 支付服务         |
| template-web-server       | 8080 | 8080 | C 端 BFF      |
| template-admin-web-server | 8081 | 8081 | B 端 BFF      |
| template-frontend         | 3000 | 3000 | C 端前端        |
| template-admin-frontend   | 3001 | 3001 | B 端前端        |




## 构建说明

`docker-compose.yml` 为 6 个应用服务配置了 `build` 上下文，镜像标签 `template-mall/<service>:latest`。

首次或代码变更后：

```bash
docker-compose build
# 或
docker-compose up -d --build
```



## 环境变量

通过 `Deployments/.env` 覆盖（参见 `.env.example`）：


| 变量                  | 默认值                      | 用途                            |
| ------------------- | ------------------------ | ----------------------------- |
| MYSQL_ROOT_PASSWORD | root123456               | MySQL root 密码                 |
| MYSQL_DATABASE      | template_order_db        | 数据库名                          |
| JWT_SECRET          | template-mall-secret-key | JWT 签名（C 端 BFF + OrderServer） |
| AUTH_PROVIDER       | mock                     | B 端认证：mock / wps              |
| ADMIN_USER_IDS      | 10001,10002              | B 端管理员白名单                     |
| OSS_*               | (空)                      | 阿里云 OSS / STS                 |
| WPS_*               | (空)                      | WPS OAuth                     |




## 健康检查


| 服务                                  | 检查方式                      |
| ----------------------------------- | ------------------------- |
| mysql                               | mysqladmin ping           |
| kafka                               | kafka-broker-api-versions |
| template-order-server               | TCP 9001                  |
| pay/template-web/template-admin-web | GET /health               |
| frontends                           | GET /health (nginx)       |




## 依赖顺序

```
mysql, kafka
  → template-order-server, pay-web-server
    → template-web-server, template-admin-web-server
      → template-frontend, template-admin-frontend
```



## E2E 测试

服务就绪后：

```bash
cd ../tests/e2e
./run_all.sh
```



## 注意事项

1. **Kafka**：使用 `apache/kafka:3.7.0` KRaft 模式；容器内 broker 地址 `kafka:19092`
2. **AutoMigrate**：OrderServer / PayWebServer 容器启动时自动建表
3. **OSS**：未配置时 B 端上传 confirm 会失败；E2E `admin_upload.sh` 会跳过 OSS 步骤
4. **端口固定**：遵循宪法，不可修改

