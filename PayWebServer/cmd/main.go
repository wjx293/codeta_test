// PayWebServer - 支付服务（HTTP only）
// 端口：8083
// 职责：创建支付单、处理支付回调、Kafka 生产支付事件
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

	"template-mall/PayWebServer/internal/config"
	"template-mall/PayWebServer/internal/handler"
	"template-mall/PayWebServer/internal/idgen"
	"template-mall/PayWebServer/internal/middleware"
	"template-mall/PayWebServer/internal/producer"
	"template-mall/PayWebServer/internal/repository"
	"template-mall/PayWebServer/internal/service"

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

	// 初始化 DB
	if err := repository.InitDB(cfg.MySQLDSN); err != nil {
		log.Fatalf("failed to init db: %v", err)
	}
	if err := repository.AutoMigrate(); err != nil {
		log.Fatalf("failed to auto migrate: %v", err)
	}
	log.Println("database migrated")

	// 仓储层
	paymentRepo := repository.NewPaymentRepo(repository.DB)
	callbackRepo := repository.NewCallbackRepo(repository.DB)
	outboxRepo := repository.NewOutboxRepo(repository.DB)

	// Kafka 生产者
	kafkaProducer := producer.NewKafkaProducer(cfg.KafkaBrokers, cfg.KafkaTopic)

	// 服务层
	paymentSvc := service.NewPaymentService(paymentRepo, callbackRepo, outboxRepo, kafkaProducer, repository.DB)

	// Outbox relay：补发 Kafka 发送失败的事件
	relayCtx, relayCancel := context.WithCancel(context.Background())
	defer relayCancel()
	go service.NewOutboxRelay(paymentSvc, 5*time.Second).Start(relayCtx)

	// 重启后从数据库同步当日最大支付单序列，避免 payment_no 冲突
	paymentIDGen := idgen.New()
	if maxNo, err := paymentRepo.MaxPaymentNoToday(); err != nil {
		log.Printf("warn: sync payment idgen failed: %v", err)
	} else if maxNo != "" {
		seq := idgen.ParseSeqFromNo("PW", maxNo)
		paymentIDGen = idgen.NewSeeded(seq)
		log.Printf("payment idgen seeded from %s (counter=%d)", maxNo, seq)
	}

	// HTTP handler
	paymentHandler := handler.NewPaymentHandler(paymentSvc, paymentIDGen)
	callbackHandler := handler.NewCallbackHandler(paymentSvc)

	// 路由
	gin.SetMode(gin.ReleaseMode)
	r := gin.New()
	r.Use(middleware.CORS(), gin.Logger(), gin.Recovery())

	r.GET("/health", callbackHandler.Health)
	r.POST("/api/payments", paymentHandler.CreatePayment)
	r.POST("/api/paycallback", callbackHandler.PayCallback)

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

		relayCancel()

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		if err := srv.Shutdown(ctx); err != nil {
			log.Printf("server forced to shutdown: %v", err)
		}

		if err := kafkaProducer.Close(); err != nil {
			log.Printf("kafka producer close: %v", err)
		}
	}()

	log.Printf("PayWebServer listening on %s", addr)
	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("server failed: %v", err)
	}
	log.Println("server exited")
}
