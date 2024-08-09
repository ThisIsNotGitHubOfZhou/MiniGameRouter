package main

import (
	"context"
	"log"
	"net"

	"github.com/mwitkow/grpc-proxy/proxy"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
)

func main() {
	// 监听端口
	lis, err := net.Listen("tcp", ":10000")
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	// 创建 gRPC 代理服务器
	grpcServer := grpc.NewServer(
		grpc.CustomCodec(proxy.Codec()),
		grpc.UnknownServiceHandler(proxy.TransparentHandler(director)),
	)

	log.Println("Starting gRPC Proxy on :10000")
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}

// director 是一个函数，用于决定请求应该被代理到哪个后端服务
func director(ctx context.Context, fullMethodName string) (context.Context, *grpc.ClientConn, error) {
	// 从上下文中提取 metadata
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return nil, nil, grpc.Errorf(grpc.Code(grpc.ErrClientConnClosing), "missing metadata")
	}

	// 提取 servername 和 prefix
	servernames := md.Get("servername")
	prefixes := md.Get("prefix")

	var servername, prefix string
	if len(servernames) > 0 {
		servername = servernames[0]
	}
	if len(prefixes) > 0 {
		prefix = prefixes[0]
	}

	log.Printf("Extracted servername: %s, prefix: %s", servername, prefix)

	// 这里可以根据 fullMethodName 或者其他上下文信息来决定代理到哪个后端服务
	backendAddr := "localhost:10002"

	// 创建到后端服务的连接
	conn, err := grpc.DialContext(ctx, backendAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, nil, err
	}

	// 将原始的 metadata 传递给后端服务
	newCtx := metadata.NewOutgoingContext(ctx, md)

	return newCtx, conn, nil
}
