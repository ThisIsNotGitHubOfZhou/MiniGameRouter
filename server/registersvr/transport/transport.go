package transport

import (
	"context"
	"fmt"
	grpctransport "github.com/go-kit/kit/transport/grpc"
	"registersvr/endpoint"
	pb "registersvr/proto"
)

// 定义 gRPC 服务器
type GrpcServer struct {
	register                              grpctransport.Handler
	deregister                            grpctransport.Handler
	pb.UnimplementedRegisterServiceServer // 嵌入未实现的服务，新版grpc需要
}

// NewGRPCServer 创建 gRPC 服务器,注册endpoint功能
func NewGRPCServer(edp endpoint.RegisterEndpoint) *GrpcServer {

	// TODO : 添加tracer
	// serverTracer := kitzipkin.GRPCServerTrace(tracer, kitzipkin.Name("grpc-transport")) // 也可以传入zipkinTracer，在NewGRPCServer内生成serverTracer
	return &GrpcServer{

		register: grpctransport.NewServer(
			edp.Register,
			decodeGRPCRegisterRequest,
			encodeGRPCRegisterResponse,
		),
		deregister: grpctransport.NewServer(
			edp.DeRegister,
			decodeGRPCDeRegisterRequest,
			encodeGRPCDeRegisterResponse,
		),
	}
}

// 注册
func (s *GrpcServer) Register(ctx context.Context, req *pb.RegisterRequest) (*pb.RegisterResponse, error) {
	_, resp, err := s.register.ServeGRPC(ctx, req)
	if err != nil {
		return nil, err
	}
	return resp.(*pb.RegisterResponse), err
}

// 请求解码器,用于将 gRPC 请求解码为 Go Kit 端点的请求
func decodeGRPCRegisterRequest(ctx context.Context, grpcReq interface{}) (interface{}, error) {

	req, ok := grpcReq.(*pb.RegisterRequest) // 先断言成grpc请求
	if !ok {
		return nil, fmt.Errorf("decodeGRPCRegisterRequest invalid request type: %T", grpcReq)
	}

	// 再转化为edpt请求
	edptReq := endpoint.RegisterRequest{
		Name:     req.Name,
		Host:     req.Host,
		Port:     req.Port,
		Protocol: req.Protocol,
		Metadata: req.Metadata,
		Weight:   int(req.Weight),
		Timeout:  int(req.Timeout),
	}
	// req := grpcReq.(endpoint.VoteToRedisRequest)
	return edptReq, nil
}

// 响应编码器,用于将 Go Kit 端点的响应编码为 gRPC 响应
func encodeGRPCRegisterResponse(_ context.Context, grpcResp interface{}) (interface{}, error) {

	// 将结果转换为endppoint层，实际上MakeVoteToRedisEndpoint返回的函数就已经做了
	resp, ok := grpcResp.(endpoint.RegisterResponse)
	if !ok {
		return nil, fmt.Errorf("encodeGRPCRegisterResponse invalid response type: %T", grpcResp)
	}
	if resp.Error != nil {
		return &pb.RegisterResponse{
			Id:       resp.InstanceID,
			ErrorMes: resp.Error.Error(),
		}, nil
	}
	return &pb.RegisterResponse{
		Id:       resp.InstanceID,
		ErrorMes: "",
	}, nil
}

// 实现 gRPC 投票结果服务接口
func (s *GrpcServer) DeRegister(ctx context.Context, req *pb.DeRegisterRequest) (*pb.DeRegisterResponse, error) {
	_, resp, err := s.deregister.ServeGRPC(ctx, req)
	if err != nil {
		return nil, err
	}
	return resp.(*pb.DeRegisterResponse), nil
}

// 请求解码器 转换成rpc请求
func decodeGRPCDeRegisterRequest(_ context.Context, grpcReq interface{}) (interface{}, error) {
	res, ok := grpcReq.(*pb.DeRegisterRequest)

	if !ok {
		return nil, fmt.Errorf("decodeGRPCDeRegisterRequest invalid request type: %T", grpcReq)
	}

	req := endpoint.DeRegisterRequest{
		ID:   res.Id,
		Name: res.Name,
		Host: res.Host,
		Port: res.Port,
	}

	return req, nil
}

// 响应编码器
func encodeGRPCDeRegisterResponse(_ context.Context, grpcResp interface{}) (interface{}, error) {
	res, ok := grpcResp.(endpoint.DeRegisterResponse)
	if !ok {
		return nil, fmt.Errorf("encodeGRPCDeRegisterResponse invalid response type: %T", grpcResp)
	}
	if res.Error != nil {
		return &pb.DeRegisterResponse{ErrorMes: res.Error.Error()}, res.Error
	}
	return &pb.DeRegisterResponse{}, nil
}
