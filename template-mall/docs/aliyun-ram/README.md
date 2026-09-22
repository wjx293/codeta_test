# RAM Role 配置清单（template-mall-sts-role）

账号 ID：`1310247281671820`  
Bucket：`template-mall-wjx`  
Role ARN：`acs:ram::1310247281671820:role/template-mall-sts-role`

本地 `.env` 已写入 `STS_ROLE_ARN`，**阿里云控制台仍需按下列步骤确认**（无法通过代码自动创建）。

## 1. 角色是否已存在

[RAM 控制台 → 角色](https://ram.console.aliyun.com/roles) 搜索 `template-mall-sts-role`。

- **已存在** → 跳到步骤 2、3
- **不存在** → 创建角色：
  - 信任主体：**云账号**
  - 云账号 ID：`1310247281671820`
  - 角色名称：`template-mall-sts-role`
  - 创建后把信任策略改为 `template-mall-sts-role-trust.json` 内容（若默认已允许本账号可跳过）

## 2. 给角色授权（上传权限）

角色详情 → **权限管理** → **新增授权** → **自定义策略** → JSON：

粘贴 `template-mall-sts-role-policy.json` 全部内容，策略名建议：`template-mall-sts-policy`。

## 3. 给 RAM 用户授权（服务端 + AssumeRole）

找到你 `.env` 里 AK 对应的 RAM 用户 → **权限管理** → **新增授权** → 自定义策略 JSON：

粘贴 `template-mall-ram-user-policy.json` 全部内容，策略名建议：`template-mall-ram-user-policy`。

## 4. OSS CORS（浏览器上传）

[OSS 控制台](https://oss.console.aliyun.com/) → `template-mall-wjx` → 数据安全 → 跨域设置：

| 项 | 值 |
|----|-----|
| 来源 | `http://localhost:3001` |
| Methods | PUT, GET, HEAD |
| Headers | `*` |

## 5. 重启服务并验证

```bash
# 重启 B 端 BFF
cd TemplateAdminWebServer && go run ./cmd/main.go
```

B 端上传 → Network 里 `upload/credential` 响应应含非空 `security_token`。

## 常见错误

| 报错 | 处理 |
|------|------|
| `assume role: NoPermission` | 步骤 3 未给 RAM 用户 `sts:AssumeRole` |
| `EntityNotExist.Role` | 步骤 1 角色未创建或 ARN 写错 |
| HeadObject 403 | 步骤 3 缺 `oss:HeadObject` |
| 上传 XHR -1 | 步骤 4 CORS 未配 |
