# Proto 代码生成

## 前置依赖

```bash
go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
```

确保 `$GOPATH/bin` 在 `PATH` 中。

## 生成命令

在 `TemplateOrderServer/` 目录下执行：

```bash
# 方式一：使用 protoc 直接生成
protoc \
  --go_out=. \
  --go_opt=paths=source_relative \
  --go-grpc_out=. \
  --go-grpc_opt=paths=source_relative \
  api/proto/template_order.proto

# 方式二：使用 buf（推荐，需先初始化 buf.gen.yaml）
# buf generate
```

生成产物：
- `api/proto/template_order.pb.go` - 消息类型
- `api/proto/template_order_grpc.pb.go` - gRPC 服务端/客户端 stub

## 验证

```bash
cd TemplateOrderServer
go build ./...
```

## 类型约定

- 二值字段：`bool`（`is_member`、`is_free`）
- 离散状态：`enum`（`TemplateStatus`、`OrderStatus`、`OrderType`）
- HTTP JSON 仍输出 `member_status` / `is_free` / `status` 的 `0/1/2/3` 数值，BFF 层负责转换

## RPC 清单（13 个）

| 分类 | RPC | 请求 | 响应 |
|---|---|---|---|
| 用户 | Register | RegisterRequest | AuthResponse |
| 用户 | Login | LoginRequest | AuthResponse |
| 用户 | RefreshToken | RefreshTokenRequest | AuthResponse |
| 用户 | RevokeRefreshTokens | RevokeRefreshTokensRequest | BaseResponse |
| 用户 | ListUsers | ListUsersRequest | ListUsersResponse |
| 模板 | CreateTemplate | CreateTemplateRequest | BaseResponse |
| 模板 | UpdateTemplate | UpdateTemplateRequest | BaseResponse |
| 模板 | ListTemplates | ListTemplatesRequest | ListTemplatesResponse |
| 模板 | GetTemplate | GetTemplateRequest | GetTemplateResponse |
| 订单 | DownloadTemplate | DownloadTemplateRequest | DownloadResponse |
| 订单 | ListOrders | ListOrdersRequest | ListOrdersResponse |
| 订单 | CancelOrder | CancelOrderRequest | BaseResponse |
| 会员 | SetMember | SetMemberRequest | BaseResponse |
