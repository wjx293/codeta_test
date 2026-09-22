// TemplateOrderServer - 核心业务服务（gRPC only）
// 端口：9001
// 职责：用户认证、模板管理、订单处理、会员管理
package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	pb "template-mall/TemplateOrderServer/api/proto"
	"template-mall/TemplateOrderServer/internal/config"
	"template-mall/TemplateOrderServer/internal/consumer"
	"template-mall/TemplateOrderServer/internal/handler"
	"template-mall/TemplateOrderServer/internal/idgen"
	"template-mall/TemplateOrderServer/internal/oss"
	"template-mall/TemplateOrderServer/internal/payclient"
	"template-mall/TemplateOrderServer/internal/repository"
	"template-mall/TemplateOrderServer/internal/service"

	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
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
	log.Println("database connected and migrated")

	// 仓储层
	userRepo := repository.NewUserRepo(repository.DB)
	refreshTokenRepo := repository.NewRefreshTokenRepo(repository.DB)
	templateRepo := repository.NewTemplateRepo(repository.DB)
	orderRepo := repository.NewOrderRepo(repository.DB)

	// 基础设施：重启后从数据库同步当日最大序列，避免业务编号冲突
	idGen := idgen.New()
	maxSeq := uint64(0)
	if maxNo, err := orderRepo.MaxOrderNoToday(); err != nil {
		log.Printf("warn: sync order idgen failed: %v", err)
	} else if maxNo != "" {
		maxSeq = idgen.ParseSeqFromNo("TO", maxNo)
	}
	if maxNo, err := templateRepo.MaxTemplateNoToday(); err != nil {
		log.Printf("warn: sync template idgen failed: %v", err)
	} else if seq := idgen.ParseSeqFromNo("TP", maxNo); seq > maxSeq {
		maxSeq = seq
	}
	if maxSeq > 0 {
		idGen = idgen.NewSeeded(maxSeq)
		log.Printf("idgen seeded (counter=%d)", maxSeq)
	}
	presigner := oss.NewPresigner(cfg.OSSEndpoint, cfg.OSSAccessKeyID, cfg.OSSAccessKeySecret, cfg.OSSBucket)
	payClient := payclient.NewClient(cfg.PayWebServerURL)

	// 服务层
	userSvc := service.NewUserService(userRepo, refreshTokenRepo, cfg.JWTSecret, cfg.JWTAccessTTL, cfg.JWTRefreshTTL)
	templateSvc := service.NewTemplateService(templateRepo, presigner, idGen)
	orderSvc := service.NewOrderService(orderRepo, templateRepo, userRepo, presigner, payClient, idGen)
	memberSvc := service.NewMemberService(userRepo)
	consumeSvc := service.NewConsumeService(repository.DB)

	// gRPC handler
	grpcHandler := handler.NewGRPCHandler(userSvc, templateSvc, orderSvc, memberSvc)

	// 启动 gRPC server
	lis, err := net.Listen("tcp", fmt.Sprintf(":%s", cfg.GRPCPort))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	grpcServer := grpc.NewServer()
	pb.RegisterTemplateOrderServiceServer(grpcServer, grpcHandler)
	reflection.Register(grpcServer) // 方便 grpcurl 调试

	go func() {
		log.Printf("gRPC server listening on :%s", cfg.GRPCPort)
		if err := grpcServer.Serve(lis); err != nil {
			log.Fatalf("gRPC server failed: %v", err)
		}
	}()

	// 启动 Kafka consumer
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	kafkaConsumer := consumer.NewKafkaConsumer(cfg.KafkaBrokers, cfg.KafkaTopic, cfg.KafkaGroupID, consumeSvc)
	go func() {
		if err := kafkaConsumer.Start(ctx); err != nil {
			log.Printf("kafka consumer stopped: %v", err)
		}
	}()

	// 优雅关闭
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("shutting down server...")
	cancel()
	grpcServer.GracefulStop()
	lis.Close()
	log.Println("server stopped")
}
