# OSS Bucket 与 STS 配置指引

> 宪法 §IV：模板文件必须保存到对象存储（OSS），Bucket 必须为私有权限。前端不得持有永久 AccessKey Secret，必须使用 STS 临时凭证或预签名 URL。

## 1. 创建私有 Bucket

### 阿里云 OSS 控制台操作

1. 登录 [阿里云 OSS 控制台](https://oss.console.aliyun.com/)
2. 创建 Bucket：
   - **Bucket 名称**：`template-mall-<env>`（如 `template-mall-dev`）
   - **地域**：`cn-hangzhou`（按实际选择）
   - **存储类型**：标准存储
   - **读写权限**：**私有**（重要：不可选公共读）
   - **服务端加密**：AES-256
3. 记录以下信息：
   - `OSS_BUCKET` = Bucket 名称
   - `OSS_ENDPOINT` = `oss-cn-hangzhou.aliyuncs.com`（按实际地域）
   - `OSS_REGION` = `cn-hangzhou`（按实际地域）

## 2. 创建 RAM 用户与 AccessKey

1. 登录 [RAM 控制台](https://ram.console.aliyun.com/)
2. 创建 RAM 用户：
   - 用户名：`template-mall-admin`
   - 访问方式：编程访问
3. 创建 AccessKey，记录：
   - `OSS_ACCESS_KEY_ID` = AccessKey ID
   - `OSS_ACCESS_KEY_SECRET` = AccessKey Secret
   > **警告**：AccessKey Secret 不得进入前端代码或 git 仓库。仅在后端 `.env` 中使用。

## 3. 创建 STS Role（用于前端直传）

### 3.1 创建自定义权限策略

在 RAM 控制台 → 权限策略管理 → 创建权限策略：

**策略名称**：`template-mall-sts-policy`

**策略内容**（JSON）：
```json
{
  "Version": "1",
  "Statement": [
    {
      "Effect": "Allow",
      "Action": [
        "oss:PutObject"
      ],
      "Resource": [
        "acs:oss:*:*:template-mall-dev/templates/*",
        "acs:oss:*:*:template-mall-dev/thumbnails/*"
      ]
    },
    {
      "Effect": "Allow",
      "Action": [
        "oss:HeadObject"
      ],
      "Resource": [
        "acs:oss:*:*:template-mall-dev/templates/*",
        "acs:oss:*:*:template-mall-dev/thumbnails/*"
      ]
    }
  ]
}
```

> 将 `template-mall-dev` 替换为实际 Bucket 名称。此策略仅允许 PUT（上传）和 HEAD（确认存在）到 `templates/` 和 `thumbnails/` 前缀下的对象，不允许读取、删除或其他操作。

### 3.2 创建 RAM 角色

1. RAM 控制台 → 角色管理 → 创建角色：
   - 角色类型：普通服务角色
   - 角色名称：`template-mall-sts-role`
   - 信任的云服务：当前账号
2. 为角色附加权限策略：`template-mall-sts-policy`
3. 记录角色 ARN：
   - `STS_ROLE_ARN` = `acs:ram::<account-id>:role/template-mall-sts-role`

### 3.3 为 RAM 用户授予 AssumeRole 权限

为步骤 2 创建的 RAM 用户 `template-mall-admin` 附加以下策略，允许其扮演 STS 角色：

**策略内容**（JSON）：
```json
{
  "Version": "1",
  "Statement": [
    {
      "Effect": "Allow",
      "Action": "sts:AssumeRole",
      "Resource": "acs:ram::<account-id>:role/template-mall-sts-role"
    }
  ]
}
```

## 4. Object Key 命名规则

> 宪法 §IV：Object Key 由服务端生成，前端不得自行指定。

### 4.1 模板文件

```
templates/{template_no}/{uuid}.{ext}
```

示例：
```
templates/T20260807000001/550e8400-e29b-41d4-a716-446655440000.pptx
```

- `template_no`：业务模板编号（`T` + `yyyyMMdd` + 8 位序列）
- `uuid`：UUID v4，防止文件名冲突
- `ext`：原始文件扩展名（ppt/pptx/doc/docx）

### 4.2 缩略图

```
thumbnails/{template_no}/{uuid}.png
```

示例：
```
thumbnails/T20260807000001/550e8400-e29b-41d4-a716-446655440000.png
```

## 5. STS 临时凭证签发流程

```
[B 端前端] 点击"上传模板"
  ↓
[B 端 BFF] POST /api/admin/upload/credential
  ↓
  1. 生成 template_no（业务编号生成器）
  2. 生成 file_oss_key = templates/{template_no}/{uuid}.{ext}
  3. 生成 thumbnail_oss_key = thumbnails/{template_no}/{uuid}.png
  4. 调用 STS AssumeRole(STS_ROLE_ARN, STS_SESSION_NAME, 策略, 300s)
  5. 返回 STS 临时凭证 + Object Key
  ↓
[B 端前端] 使用 STS 凭证直接 PUT 到 OSS
  ↓
[B 端前端] 上传完成 → POST /api/admin/upload/confirm
  ↓
[B 端 BFF] HeadObject 确认对象存在 + 获取文件大小
  ↓
[B 端 BFF] 调 gRPC CreateTemplate（含 file_oss_key, file_size 等）
```

### STS 凭证有效期

- **5 分钟（300 秒）**：宪法 §IV 强制要求
- 过期后前端需重新申请

### STS 策略限制

签发 STS 时附加 inline policy，限制：
- 仅允许 `oss:PutObject` 到指定的 `file_oss_key` 和 `thumbnail_oss_key`
- 不允许读取、删除、列举

## 6. 预签名 URL 生成

### 下载地址（C 端用户下载模板）

```
PresignDownload(file_oss_key, 5*time.Minute)
  → https://template-mall-dev.oss-cn-hangzhou.aliyuncs.com/templates/...?Expires=...&Signature=...
```

- 有效期：5 分钟
- 使用 RAM 用户的 AccessKey 签名（非 STS）
- 仅在 `order_status=2`（已获得下载资格）时生成

### 缩略图预签名（列表/详情展示）

```
PresignThumbnail(thumbnail_oss_key)
  → https://template-mall-dev.oss-cn-hangzhou.aliyuncs.com/thumbnails/...?Expires=...&Signature=...
```

- 有效期：5 分钟
- 在 ListTemplates / GetTemplate 返回时填充 `thumbnail_download_url`

## 7. .env 变量清单

`TemplateAdminWebServer/.env`：

```bash
# OSS 基础配置
OSS_ACCESS_KEY_ID=LTAI5tXXXXXXXXXXXX
OSS_ACCESS_KEY_SECRET=XXXXXXXXXXXXXXXXXXXXXXXX
OSS_BUCKET=template-mall-dev
OSS_ENDPOINT=oss-cn-hangzhou.aliyuncs.com
OSS_REGION=cn-hangzhou

# STS 配置（配置 OSS 后必填）
STS_ROLE_ARN=acs:ram::1234567890:role/template-mall-sts-role
STS_SESSION_NAME=template-mall-admin
```

## 8. 安全检查清单

- [ ] Bucket 读写权限为**私有**（非公共读）
- [ ] AccessKey Secret **不在前端代码**中
- [ ] AccessKey Secret **不在 git 仓库**中（.env 已加入 .gitignore）
- [ ] STS 凭证有效期 ≤ 5 分钟
- [ ] STS 策略仅允许 PUT + HEAD 到 `templates/*` 和 `thumbnails/*`
- [ ] Object Key 由服务端生成（前端不可指定）
- [ ] 预签名 URL 有效期 ≤ 5 分钟
- [ ] 下载地址仅在 `order_status=2` 时生成
- [ ] RAM 用户最小权限原则（仅 AssumeRole + 必要的预签名操作）
