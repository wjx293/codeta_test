# Build stage
FROM golang:1.25-alpine AS builder

ENV GOPROXY=https://goproxy.cn,direct \
    GOSUMDB=sum.golang.google.cn

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /paywebserver ./cmd/main.go

# Run stage
FROM alpine:3.20

RUN apk --no-cache add ca-certificates tzdata wget
ENV TZ=Asia/Shanghai

COPY --from=builder /paywebserver /usr/local/bin/paywebserver

EXPOSE 8083

ENTRYPOINT ["paywebserver"]