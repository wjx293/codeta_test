# Build stage — context 为 template-mall 根目录（见 docker-compose build.context）
FROM golang:1.25-alpine AS builder

ENV GOPROXY=https://goproxy.cn,direct \
    GOSUMDB=sum.golang.google.cn

WORKDIR /src

# 先复制 OrderServer（满足 go.mod replace）
COPY TemplateOrderServer /src/TemplateOrderServer

WORKDIR /src/TemplateAdminWebServer
COPY TemplateAdminWebServer/go.mod TemplateAdminWebServer/go.sum ./
RUN go mod download

COPY TemplateAdminWebServer .
RUN CGO_ENABLED=0 GOOS=linux go build -o /templateadminwebserver ./cmd/main.go

# Run stage
FROM alpine:3.20

RUN apk --no-cache add ca-certificates tzdata wget
ENV TZ=Asia/Shanghai

COPY --from=builder /templateadminwebserver /usr/local/bin/templateadminwebserver

EXPOSE 8081

ENTRYPOINT ["templateadminwebserver"]
