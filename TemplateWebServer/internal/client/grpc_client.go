// Package client 提供 TemplateOrderServer 的 gRPC 客户端。
package client

import (
	"fmt"
	"log"
	"sync"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	pb "template-mall/TemplateOrderServer/api/proto"
)

// GRPCClient 封装 gRPC 连接和 stub。
type GRPCClient struct {
	conn   *grpc.ClientConn
	Stub   pb.TemplateOrderServiceClient
	mu     sync.Mutex
	closed bool
}

// NewGRPCClient 创建 gRPC 客户端并建立连接。
func NewGRPCClient(target string) (*GRPCClient, error) {
	conn, err := grpc.NewClient(
		target,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithDefaultCallOptions(grpc.MaxCallRecvMsgSize(10*1024*1024)),
	)
	if err != nil {
		return nil, fmt.Errorf("grpc dial %s: %w", target, err)
	}

	log.Printf("gRPC client connected to %s", target)

	return &GRPCClient{
		conn: conn,
		Stub: pb.NewTemplateOrderServiceClient(conn),
	}, nil
}

// Close 关闭连接（幂等）。
func (c *GRPCClient) Close() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.closed {
		return nil
	}
	c.closed = true

	log.Println("gRPC client closing...")
	return c.conn.Close()
}
