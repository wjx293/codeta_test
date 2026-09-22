// TemplateAdminWebServer - B 端 BFF (HTTP only)
// 端口：8081
// 职责：WPS OAuth2 / Mock Auth；STS 签发；HeadObject 确认；gRPC 调 TemplateOrderServer
// 不直连 DB（宪法 I）；不中转文件内容（宪法 I）；不签发 Admin JWT（宪法 I）
package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"template-mall/TemplateAdminWebServer/internal/auth"
	"template-mall/TemplateAdminWebServer/internal/client"
	"template-mall/TemplateAdminWebServer/internal/config"
	"template-mall/TemplateAdminWebServer/internal/handler"
	"template-mall/TemplateAdminWebServer/internal/middleware"
	"template-mall/TemplateAdminWebServer/internal/oss"

	"github.com/gin-gonic/gin"
)

func main() {
	// 时区初始化（宪法 CL-6）
	loc, err := time.LoadLocation("Asia/Shanghai")
	if err != nil {
		log.Fatalf("failed to load timezone Asia/Shanghai: %v", err)
	}
	time.Local = loc

	// 加载配置
	cfg := config.Load()
	log.Printf("config loaded: %s", cfg.String())

	// gRPC 客户端
	grpcClient, err := client.NewGRPCClient(cfg.GRPCTarget)
	if err != nil {
		log.Fatalf("failed to connect gRPC: %v", err)
	}
	defer grpcClient.Close()

	// Auth Provider（根据 AUTH_PROVIDER 切换）
	var authProvider auth.AuthProvider
	switch cfg.AuthProvider {
	case "wps":
		authProvider = auth.NewWPSOAuthProvider(cfg)
		log.Println("auth provider: WPS OAuth2")
	case "mock":
		authProvider = auth.NewMockAuthProvider(cfg)
		log.Println("auth provider: Mock (开发演示)")
	default:
		log.Fatalf("unknown AUTH_PROVIDER: %s (must be wps|mock)", cfg.AuthProvider)
	}

	// 会话存储（8h TTL）
	sessionStore := auth.NewSessionStore(8 * time.Hour)

	// OSS
	stsIssuer := oss.NewSTSIssuer(cfg)
	headObject := oss.NewHeadObject(cfg)
	objectKey := oss.NewObjectKey()
	thumbGen := oss.NewThumbnailGenerator(headObject, objectKey)

	// Handlers
	authHandler := handler.NewAuthHandler(authProvider, sessionStore, cfg)
	uploadHandler := handler.NewUploadHandler(stsIssuer, headObject, objectKey)
	templateHandler := handler.NewTemplateHandler(grpcClient, headObject, thumbGen)
	orderHandler := handler.NewOrderHandler(grpcClient)
	userHandler := handler.NewUserHandler(grpcClient)

	// 路由
	gin.SetMode(gin.ReleaseMode)
	r := gin.New()
	r.Use(middleware.CORS(), gin.Logger(), gin.Recovery())

	// 健康检查（无鉴权）
	r.GET("/health", handler.Health)

	// 认证接口（无鉴权）
	r.GET("/auth/login", authHandler.Login)
	r.GET("/auth/callback", authHandler.Callback)
	r.GET("/api/auth/login", authHandler.Login)
	r.GET("/api/auth/callback", authHandler.Callback)
	// Mock 模式登录端点（AUTH_PROVIDER=mock 时使用）
	if cfg.MockAuthEnabled {
		r.POST("/auth/mock-login", authHandler.MockLogin)
		r.POST("/api/auth/mock-login", authHandler.MockLogin)
	}

	// 退出登录（不需要会话鉴权，幂等清 Cookie）
	r.POST("/auth/logout", authHandler.Logout)
	r.GET("/auth/logout", authHandler.Logout)
	r.POST("/api/auth/logout", authHandler.Logout)
	r.GET("/api/auth/logout", authHandler.Logout)

	// 需要会话鉴权的接口
	authRequired := r.Group("")
	authRequired.Use(middleware.SessionAuth(sessionStore))
	{
		authRequired.GET("/auth/me", authHandler.Me)
		authRequired.GET("/api/auth/me", authHandler.Me)

		// 上传
		authRequired.POST("/api/admin/upload/credential", uploadHandler.Credential)
		authRequired.POST("/api/admin/upload/confirm", uploadHandler.Confirm)

		// 模板
		authRequired.POST("/api/admin/templates", templateHandler.Create)
		authRequired.PUT("/api/admin/templates/:id", templateHandler.Update)
		authRequired.GET("/api/admin/templates", templateHandler.List)
		authRequired.PUT("/api/admin/templates/:id/status", templateHandler.UpdateStatus)

		// 订单
		authRequired.GET("/api/admin/orders", orderHandler.List)

		// 用户
		authRequired.GET("/api/admin/users", userHandler.List)
		authRequired.PUT("/api/admin/users/:id/member", userHandler.SetMember)
	}

	// 启动 HTTP server
	addr := fmt.Sprintf(":%s", cfg.HTTPPort)
	srv := &http.Server{
		Addr:    addr,
		Handler: r,
	}

	// 优雅关闭
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-quit
		log.Println("shutting down server...")

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		if err := srv.Shutdown(ctx); err != nil {
			log.Printf("server forced to shutdown: %v", err)
		}
	}()

	log.Printf("TemplateAdminWebServer listening on %s", addr)
	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("server failed: %v", err)
	}
	log.Println("server exited")
}