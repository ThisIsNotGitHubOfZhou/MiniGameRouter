package transport

import (
	"context"
	"fmt"
	grpctransport "github.com/go-kit/kit/transport/grpc"
	"healthchecksvr/endpoint"
	pb "healthchecksvr/proto"
)

// 定义 gRPC 服务器
type GrpcServer struct {
	healthchecks                             grpctransport.Handler
	healthcheckc                             grpctransport.Handler
	pb.UnimplementedHealthCheckServiceServer // 嵌入未实现的服务，新版grpc需要
}

// NewGRPCServer 创建 gRPC 服务器,注册endpoint功能
func NewGRPCServer(edp endpoint.HealthCheckEndpoint) *GrpcServer {

	// TODO : 添加tracer
	// serverTracer := kitzipkin.GRPCServerTrace(tracer, kitzipkin.Name("grpc-transport")) // 也可以传入zipkinTracer，在NewGRPCServer内生成serverTracer
	return &GrpcServer{

		healthchecks: grpctransport.NewServer(
			edp.HealthCheckS,
			decodeGRPCHealthcheckSRequest,
			decodeGRPCHealthcheckSResponse,
		),
		healthcheckc: grpctransport.NewServer(
			edp.HealthCheckC,
			decodeGRPCHealthCheckCRequest,
			decodeGRPCHealthCheckCResponse,
		),
	}
}

// 注册
func (s *GrpcServer) HealthCheckS(ctx context.Context, req *pb.HealthCheckSRequest) (*pb.HealthCheckSResponse, error) {
	_, resp, err := s.healthchecks.ServeGRPC(ctx, req)
	if err != nil {
		return nil, err
	}
	return resp.(*pb.HealthCheckSResponse), err
}

// 请求解码器,用于将 gRPC 请求解码为 Go Kit 端点的请求
func decodeGRPCHealthcheckSRequest(ctx context.Context, grpcReq interface{}) (interface{}, error) {

	req, ok := grpcReq.(*pb.HealthCheckSRequest) // 先断言成grpc请求
	if !ok {
		return nil, fmt.Errorf("[Error] healthCheck decodeGRPCHealthcheckSRequest invalid request type: %T", grpcReq)
	}

	// 再转化为edpt请求
	edptReq := endpoint.HealthCheckSRequest{
		Name:       req.Name,
		InstanceID: req.InstanceID,
		Url:        req.Url,
		Timeout:    int(req.Timeout),
	}
	// req := grpcReq.(endpoint.VoteToRedisRequest)
	return edptReq, nil
}

// 响应编码器,用于将 Go Kit 端点的响应编码为 gRPC 响应
func decodeGRPCHealthcheckSResponse(_ context.Context, grpcResp interface{}) (interface{}, error) {

	// 将结果转换为endppoint层，实际上MakeVoteToRedisEndpoint返回的函数就已经做了
	resp, ok := grpcResp.(endpoint.HealthCheckSResponse)
	if !ok {
		return nil, fmt.Errorf("[Error] healthCheck decodeGRPCHealthcheckSResponse invalid response type: %T", grpcResp)
	}
	if resp.Error != nil {
		return &pb.HealthCheckSResponse{
			ErrorMes: resp.Error.Error(),
		}, nil
	}
	return &pb.HealthCheckSResponse{
		ErrorMes: "",
	}, nil
}

// 实现 gRPC 投票结果服务接口
func (s *GrpcServer) HealthCheckC(ctx context.Context, req *pb.HealthCheckCRequest) (*pb.HealthCheckCResponse, error) {
	_, resp, err := s.healthcheckc.ServeGRPC(ctx, req)
	if err != nil {
		return nil, err
	}
	return resp.(*pb.HealthCheckCResponse), nil
}

// 请求解码器 转换成rpc请求
func decodeGRPCHealthCheckCRequest(_ context.Context, grpcReq interface{}) (interface{}, error) {
	res, ok := grpcReq.(*pb.HealthCheckCRequest)

	if !ok {
		return nil, fmt.Errorf("[Error] healthcheck decodeGRPCHealthCheckCRequest invalid request type: %T", grpcReq)
	}

	req := endpoint.HealthCheckCRequest{
		ID:      res.Id,
		Name:    res.Name,
		Host:    res.Host,
		Port:    res.Port,
		Timeout: int(res.Timeout),
	}

	return req, nil
}

// 响应编码器
func decodeGRPCHealthCheckCResponse(_ context.Context, grpcResp interface{}) (interface{}, error) {
	res, ok := grpcResp.(endpoint.HealthCheckCResponse)
	if !ok {
		return nil, fmt.Errorf("[Error] healthcheck encodeGRPCDeRegisterResponse invalid response type: %T", grpcResp)
	}
	if res.Error != nil {
		return &pb.HealthCheckCResponse{ErrorMes: res.Error.Error()}, res.Error
	}
	return &pb.HealthCheckCResponse{ErrorMes: ""}, nil
}
