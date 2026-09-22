# TemplateAdminWebServer API 文档

B 端 BFF 服务，端口 `8081`。

## 通用约定

- 请求/响应格式：`application/json`
- 错误响应统一格式：`{"code": <int>, "message": <string>}`
- `code=0` 表示成功，非 0 表示业务错误
- 金额字段单位：分（`int64`）
- 时间字段格式：ISO 8601（`2006-01-02T15:04:05Z07:00`）
- 鉴权方式：会话 Cookie `admin_sid`（httpOnly + SameSite=Lax + Path=/）
- **不签发 Admin JWT**（宪法 I）

---

## OAuth2 认证流程

### AUTH_PROVIDER=wps（生产环境）

```
1. 前端点击登录 → GET /auth/login
2. BFF 302 跳转 WPS 授权页：
   https://openapi.wps.cn/oauth2/auth?client_id=AK20230207WTGKYW
     &redirect_uri=http://localhost:8081/auth/callback
     &scope=kso.user_base.read
     &response_type=code
     &state=<timestamp>
3. 用户授权后 WPS 回调 → GET /auth/callback?code=xxx&state=xxx
4. BFF 校验 state → POST https://openapi.wps.cn/oauth2/token 换 access_token
5. BFF GET https://openapi.wps.cn/v7/users/current 拉用户信息
6. BFF 校验 user_id ∈ ADMIN_USER_IDS（.env 白名单）
7. BFF 写服务端会话 → Set-Cookie: admin_sid=<sid>; HttpOnly; SameSite=Lax; Path=/
8. BFF 302 跳转前端 http://localhost:3001/
```



### AUTH_PROVIDER=mock（开发演示）

```
1. 前端调用 POST /auth/mock-login
2. BFF 用 .env 固定账号（MOCK_ADMIN_USER_ID/MOCK_ADMIN_NAME）写会话
3. 返回 Set-Cookie: admin_sid=<sid>
```

### 安全约束

- `client_secret` **仅在服务端使用**，不进响应、不进日志、不进前端
- 会话 Cookie 设置 `httpOnly` + `SameSite=Lax` + `Path=/`
- 不签发 Admin JWT

---

## 系统接口

### GET /health

健康检查（无鉴权）。

**响应示例**：

```json
{
  "service": "TemplateAdminWebServer",
  "version": "1.0.0"
}
```

---

## 认证接口

### GET /auth/login

跳转到 OAuth 授权页（WPS）或 Mock 登录端点。

**响应**：302 重定向

### GET /auth/callback

OAuth 回调处理。

**查询参数**：


| 参数    | 类型     | 说明            |
| ----- | ------ | ------------- |
| code  | string | WPS 授权码       |
| state | string | CSRF 防护 state |


**响应**：302 跳转 `FRONTEND_HOME_URL`（成功）或 400/403/502（失败）

### POST /auth/mock-login

Mock 模式登录（仅 AUTH_PROVIDER=mock 时可用）。

**响应示例**：

```json
{
  "code": 0,
  "data": {
    "user_id": "10001",
    "username": "MockAdmin"
  }
}
```

Set-Cookie: `admin_sid=<sid>; HttpOnly; SameSite=Lax; Path=/`

### POST /auth/logout

### GET /auth/logout

退出登录（幂等，清 Cookie）。

**响应示例**：

```json
{
  "code": 0,
  "message": "logout success"
}
```

### GET /auth/me （会话鉴权）

当前管理员信息。

**响应示例**：

```json
{
  "code": 0,
  "data": {
    "user_id": "10001",
    "username": "MockAdmin",
    "avatar": ""
  }
}
```

---

## 上传接口

### POST /api/admin/upload/credential （会话鉴权）

签发 STS 临时凭证 + 生成 Object Key。

**请求体**：

```json
{
  "file_type": "pptx",
  "file_size": 1048576,
  "is_thumbnail": false
}
```

**校验规则**：

- 模板文件：file_type ∈ [ppt, pptx, doc, docx]，size ≤ 5MB
- 缩略图：file_type ∈ [jpg, jpeg, png]，size ≤ 1MB
- Object Key 规则：`templates/{yyyyMM}/{uuid}.{ext}` 或 `thumbnails/{yyyyMM}/{uuid}.{ext}`

**响应示例**：

```json
{
  "code": 0,
  "data": {
    "object_key": "templates/202608/abc123-def456.pptx",
    "access_key_id": "...",
    "access_key_secret": "...",
    "security_token": "...",
    "expiration": "2026-08-07T16:05:00+08:00",
    "bucket": "template-mall",
    "endpoint": "oss-cn-beijing.aliyuncs.com",
    "region": "cn-beijing"
  },
  "max_size": 5242880
}
```

### POST /api/admin/upload/confirm （会话鉴权）

确认对象已上传到 OSS（HeadObject）。

**请求体**：

```json
{
  "object_key": "templates/202608/abc123-def456.pptx",
  "file_size": 1048576
}
```

**响应示例**：

```json
{
  "code": 0,
  "data": {
    "object_key": "templates/202608/abc123-def456.pptx",
    "file_size": 1024,
    "confirmed": true
  }
}
```

---

## 模板管理接口

### POST /api/admin/templates （会话鉴权）

创建模板。

**请求体**：

```json
{
  "name": "企业官网模板",
  "description": "简洁响应式",
  "is_free": 0,
  "price": 9900,
  "file_type": "pptx",
  "file_size": 1048576,
  "file_oss_key": "templates/202608/abc123-def456.pptx",
  "thumbnail_oss_key": "thumbnails/202608/thumb-uuid.jpg",
  "thumbnail_file_size": 51200
}
```

**响应示例**：

```json
{
  "code": 0,
  "message": "template created"
}
```

### PUT /api/admin/templates/:id （会话鉴权）

修改模板基本信息。

**请求体**：

```json
{
  "name": "企业官网模板 v2",
  "description": "更新版本",
  "is_free": 0,
  "price": 12900,
  "thumbnail_oss_key": "thumbnails/202608/new-thumb.jpg",
  "thumbnail_file_size": 61440
}
```

### GET /api/admin/templates （会话鉴权）

模板列表（分页）。

**查询参数**：


| 参数        | 类型  | 默认  | 说明                |
| --------- | --- | --- | ----------------- |
| status    | int | -1  | -1=全部 0=已下架 1=已上架 |
| page      | int | 1   | 页码                |
| page_size | int | 20  | 每页数量              |


### PUT /api/admin/templates/:id/status （会话鉴权）

上架/下架模板。

**请求体**：

```json
{
  "status": 1
}
```

**响应示例**：

```json
{
  "code": 0,
  "message": "status updated"
}
```

---

## 订单接口

### GET /api/admin/orders （会话鉴权）

订单列表（CL-12 筛选 + 分页）。

**查询参数**：


| 参数         | 类型      | 默认  | 说明                          |
| ---------- | ------- | --- | --------------------------- |
| status     | int     | -1  | -1=全部 1=未支付 2=已获得下载资格 3=已取消 |
| user_id    | int64   | -   | 用户 ID 精确筛选                  |
| username   | string  | -   | 用户名模糊筛选                     |
| start_time | RFC3339 | -   | created_at 起始时间             |
| end_time   | RFC3339 | -   | created_at 结束时间             |
| page       | int     | 1   | 页码                          |
| page_size  | int     | 20  | 每页数量                        |


---

## 用户接口

### GET /api/admin/users （会话鉴权）

用户列表（模糊搜索 + 分页，供设置会员时检索）。

**查询参数**：


| 参数        | 类型     | 默认  | 说明   |
| --------- | ------ | --- | ---- |
| username  | string | -   | 模糊搜索 |
| page      | int    | 1   | 页码   |
| page_size | int    | 20  | 每页数量 |


### PUT /api/admin/users/:id/member （会话鉴权）

设置/取消会员。

**请求体**：

```json
{
  "member_status": 1
}
```

**响应示例**：

```json
{
  "code": 0,
  "message": "member status updated"
}
```

---

## 接口汇总（13 个业务接口 + 1 个 /health）


| #   | 方法       | 路径                              | 鉴权    | 用途                  |
| --- | -------- | ------------------------------- | ----- | ------------------- |
| 1   | GET      | /auth/login                     | 无     | 跳转 WPS OAuth        |
| 2   | GET      | /auth/callback                  | 无     | OAuth 回调            |
| 3   | GET/POST | /auth/logout                    | 无（幂等） | 退出                  |
| 4   | GET      | /auth/me                        | 会话    | 当前管理员               |
| 5   | POST     | /api/admin/upload/credential    | 会话    | 签发 STS + Object Key |
| 6   | POST     | /api/admin/upload/confirm       | 会话    | 确认对象存在              |
| 7   | POST     | /api/admin/templates            | 会话    | 创建模板                |
| 8   | PUT      | /api/admin/templates/:id        | 会话    | 修改模板                |
| 9   | GET      | /api/admin/templates            | 会话    | 模板列表                |
| 10  | PUT      | /api/admin/templates/:id/status | 会话    | 上架/下架               |
| 11  | GET      | /api/admin/orders               | 会话    | 订单列表                |
| 12  | GET      | /api/admin/users                | 会话    | 用户列表                |
| 13  | PUT      | /api/admin/users/:id/member     | 会话    | 设置会员                |
| -   | GET      | /health                         | 无     | 健康检查                |


---

## 错误码


| code | HTTP Status | 说明                     |
| ---- | ----------- | ---------------------- |
| 0    | 200         | 成功                     |
| 400  | 400         | 请求参数错误                 |
| 401  | 401         | 无会话 / 会话过期             |
| 403  | 403         | 非管理员白名单                |
| 404  | 404         | 资源不存在                  |
| 500  | 500         | 服务器内部错误                |
| 502  | 502         | 上游服务（gRPC/OAuth/OSS）失败 |


