# TemplateOrderServer Dockerfile - 多阶段构建
# 构建：docker build -t template-mall/template-order-server:latest .
# 运行：docker run -p 9001:9001 --env-file .env template-mall/template-order-server:latest

# ---- 构建阶段 ----
FROM golang:1.25-alpine AS builder

ENV GOPROXY=https://goproxy.cn,direct \
    GOSUMDB=sum.golang.google.cn

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /app/server ./cmd/main.go

# ---- 运行阶段 ----
FROM alpine:3.19

RUN apk add --no-cache ca-certificates tzdata netcat-openbsd
ENV TZ=Asia/Shanghai

WORKDIR /app
COPY --from=builder /app/server .

EXPOSE 9001

CMD ["./server"]