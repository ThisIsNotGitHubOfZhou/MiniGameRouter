package main

import (
	"context"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
	"log"
	"time"

	pb "client/proto"
	"google.golang.org/grpc"
)

func main() {
	conn, err := grpc.Dial("localhost:10000", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	defer conn.Close()

	client := pb.NewGreeterClient(conn)

	// 创建包含 metadata 的上下文
	md := metadata.Pairs(
		"servername", "zcf",
		"prefix", "hello",
	)
	ctx := metadata.NewOutgoingContext(context.Background(), md)

	// 设置超时
	ctx, cancel := context.WithTimeout(ctx, time.Second)
	defer cancel()

	// 调用 gRPC 方法
	res, err := client.SayHello(ctx, &pb.HelloRequest{})
	if err != nil {
		log.Fatalf("could not greet: %v", err)
	}

	log.Printf("Response: %s", res.Message)
}
