# Build stage — context 为 template-mall 根目录（见 docker-compose build.context）
FROM golang:1.25-alpine AS builder

ENV GOPROXY=https://goproxy.cn,direct \
    GOSUMDB=sum.golang.google.cn

WORKDIR /src

COPY TemplateOrderServer /src/TemplateOrderServer

WORKDIR /src/TemplateWebServer
COPY TemplateWebServer/go.mod TemplateWebServer/go.sum ./
RUN go mod download

COPY TemplateWebServer .
RUN CGO_ENABLED=0 GOOS=linux go build -o /templatewebserver ./cmd/main.go

# Run stage
FROM alpine:3.20

RUN apk --no-cache add ca-certificates tzdata wget
ENV TZ=Asia/Shanghai

COPY --from=builder /templatewebserver /usr/local/bin/templatewebserver

EXPOSE 8080

ENTRYPOINT ["templatewebserver"]
