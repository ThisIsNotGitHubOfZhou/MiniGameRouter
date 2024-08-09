package main

import (
	"context"
	"log"
	"net"

	"google.golang.org/grpc"
	pb "server/proto"
)

type server struct {
	pb.UnimplementedGreeterServer
}

func (s *server) SayHello(context.Context, *pb.HelloRequest) (*pb.HelloReply, error) {
	// 实现你的 gRPC 方法
	return &pb.HelloReply{Message: "Hello from backend"}, nil
}

func main() {
	lis, err := net.Listen("tcp", ":10002")
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	grpcServer := grpc.NewServer()
	pb.RegisterGreeterServer(grpcServer, &server{})

	log.Println("Starting Backend gRPC Server on :10002")
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
