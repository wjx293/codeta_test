// TemplateWebServer - C 端 BFF（HTTP only）
// 端口：8080
// 职责：JWT 鉴权、gRPC 调 TemplateOrderServer、C 端 API
// 不直连 DB（宪法 I）
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

	"template-mall/TemplateWebServer/internal/client"
	"template-mall/TemplateWebServer/internal/config"
	"template-mall/TemplateWebServer/internal/handler"
	"template-mall/TemplateWebServer/internal/middleware"

	"github.com/gin-gonic/gin"
)

func main() {
	// 时区初始化
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

	// HTTP handlers
	authHandler := handler.NewAuthHandler(grpcClient)
	templateHandler := handler.NewTemplateHandler(grpcClient)
	orderHandler := handler.NewOrderHandler(grpcClient)
	userHandler := handler.NewUserHandler(grpcClient)

	// 路由
	gin.SetMode(gin.ReleaseMode)
	r := gin.New()
	r.Use(middleware.CORS(), gin.Logger(), gin.Recovery())

	// 健康检查（无鉴权）
	r.GET("/health", handler.Health)

	// 认证接口（无鉴权）
	auth := r.Group("/auth")
	{
		auth.POST("/register", authHandler.Register)
		auth.POST("/login", authHandler.Login)
		auth.POST("/refresh", authHandler.Refresh)
	}

	// 需要鉴权的接口
	authRequired := r.Group("")
	authRequired.Use(middleware.JWTAuth(cfg.JWTSecret))
	{
		authRequired.POST("/auth/logout", authHandler.Logout)

		authRequired.GET("/templates", templateHandler.ListTemplates)
		authRequired.GET("/templates/:id", templateHandler.GetTemplate)
		authRequired.POST("/templates/:id/download", templateHandler.DownloadTemplate)

		authRequired.GET("/orders", orderHandler.ListOrders)
		authRequired.POST("/orders/:order_no/cancel", orderHandler.CancelOrder)

		authRequired.GET("/users/me", userHandler.GetCurrentUser)
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

	log.Printf("TemplateWebServer listening on %s", addr)
	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("server failed: %v", err)
	}
	log.Println("server exited")
}