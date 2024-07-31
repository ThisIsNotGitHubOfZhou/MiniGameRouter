package transport

import (
	"context"
	"discoversvr/config"
	"discoversvr/endpoint"
	pb "discoversvr/proto"
	"fmt"
	grpctransport "github.com/go-kit/kit/transport/grpc"
)

// 定义 gRPC 服务器
type GrpcServer struct {
	discoverServiceWithName               grpctransport.Handler
	discoverServiceWithID                 grpctransport.Handler
	getRouteInfoWithName                  grpctransport.Handler
	getRouteInfoWithPrefix                grpctransport.Handler
	setRouteRule                          grpctransport.Handler
	pb.UnimplementedDiscoverServiceServer // 嵌入未实现的服务，新版grpc需要
}

// NewGRPCServer 创建 gRPC 服务器,注册endpoint功能
func NewGRPCServer(edp endpoint.DiscoverEndpoint) *GrpcServer {

	// TODO : 添加tracer
	// serverTracer := kitzipkin.GRPCServerTrace(tracer, kitzipkin.Name("grpc-transport")) // 也可以传入zipkinTracer，在NewGRPCServer内生成serverTracer
	return &GrpcServer{

		discoverServiceWithName: grpctransport.NewServer(
			edp.DiscoverServiceWithName,
			decodeGRPCDiscoverServiceWithNameRequest,
			encodeGRPCDiscoverServiceWithNameResponse,
		),
		discoverServiceWithID: grpctransport.NewServer(
			edp.DiscoverServiceWithID,
			decodeGRPCDiscoverServiceWithIDRequest,
			encodeGRPCDiscoverServiceWithIDResponse,
		),
		getRouteInfoWithName: grpctransport.NewServer(
			edp.GetRouteInfoWithName,
			decodeGRPCGetRouteInfoWithNameRequest,
			encodeGRPCGetRouteInfoWithNameResponse,
		),
		getRouteInfoWithPrefix: grpctransport.NewServer(
			edp.GetRouteInfoWithPrefix,
			decodeGRPCGetRouteInfoWithPrefixRequest,
			encodeGRPCGetRouteInfoWithPrefixResponse,
		),
		setRouteRule: grpctransport.NewServer(
			edp.SetRouteRule,
			decodeGRPCSetRouteRuleRequest,
			encodeGRPCSetRouteRuleResponse,
		),
	}
}

// 注册
func (s *GrpcServer) DiscoverServiceWithName(ctx context.Context, req *pb.DiscoverServiceWithNameRequest) (*pb.DiscoverServiceResponse, error) {
	_, resp, err := s.discoverServiceWithName.ServeGRPC(ctx, req)
	if err != nil {
		config.Logger.Println("[Error][discover] DiscoverServiceWithName grpc服务出错")
		return nil, err
	}
	return resp.(*pb.DiscoverServiceResponse), err
}

// 请求解码器,用于将 gRPC 请求解码为 Go Kit 端点的请求
func decodeGRPCDiscoverServiceWithNameRequest(ctx context.Context, grpcReq interface{}) (interface{}, error) {

	req, ok := grpcReq.(*pb.DiscoverServiceWithNameRequest) // 先断言成grpc请求
	if !ok {
		return nil, fmt.Errorf("decodeGRPCDiscoverServiceWithNameRequest invalid request type: %T", grpcReq)
	}

	// 再转化为edpt请求
	edptReq := endpoint.DiscoverServiceWithNameRequest{
		Name: req.Name,
	}

	return edptReq, nil
}

// 响应编码器,用于将 Go Kit 端点的响应编码为 gRPC 响应
func encodeGRPCDiscoverServiceWithNameResponse(_ context.Context, grpcResp interface{}) (interface{}, error) {

	// 将结果转换为endppoint层，实际上MakeVoteToRedisEndpoint返回的函数就已经做了
	resp, ok := grpcResp.(endpoint.DiscoverServiceWithNameResponse)
	if !ok {
		return nil, fmt.Errorf("encodeGRPCDiscoverServiceWithNameResponse invalid response type: %T", grpcResp)
	}
	if resp.Error != nil {
		return &pb.DiscoverServiceResponse{
			Services: resp.ServiceInfos,
			ErrorMes: resp.Error.Error(),
		}, nil
	}
	return &pb.DiscoverServiceResponse{
		Services: resp.ServiceInfos,
		ErrorMes: "",
	}, nil
}

// 实现 gRPC 投票结果服务接口
func (s *GrpcServer) DiscoverServiceWithID(ctx context.Context, req *pb.DiscoverServiceWithIDRequest) (*pb.DiscoverServiceResponse, error) {
	_, resp, err := s.discoverServiceWithID.ServeGRPC(ctx, req)
	if err != nil {
		return nil, err
	}
	return resp.(*pb.DiscoverServiceResponse), nil
}

// 请求解码器 转换成rpc请求
func decodeGRPCDiscoverServiceWithIDRequest(_ context.Context, grpcReq interface{}) (interface{}, error) {
	res, ok := grpcReq.(*pb.DiscoverServiceWithIDRequest)

	if !ok {
		return nil, fmt.Errorf("decodeGRPCDiscoverServiceWithIDRequest invalid request type: %T", grpcReq)
	}

	req := endpoint.DiscoverServiceWithIDRequest{
		InstanceID: res.InstanceId,
	}

	return req, nil
}

// 响应编码器
func encodeGRPCDiscoverServiceWithIDResponse(_ context.Context, grpcResp interface{}) (interface{}, error) {
	res, ok := grpcResp.(endpoint.DiscoverServiceWithIDResponse)
	if !ok {
		return nil, fmt.Errorf("encodeGRPCDiscoverServiceWithIDResponse invalid response type: %T", grpcResp)
	}
	if res.Error != nil {
		return &pb.DiscoverServiceResponse{Services: res.ServiceInfos, ErrorMes: res.Error.Error()}, res.Error
	}
	return &pb.DiscoverServiceResponse{Services: res.ServiceInfos}, nil
}

// 实现 gRPC 投票结果服务接口
func (s *GrpcServer) GetRouteInfoWithName(ctx context.Context, req *pb.GetRouteInfoWithNameRequest) (*pb.RouteInfosResponse, error) {
	_, resp, err := s.getRouteInfoWithName.ServeGRPC(ctx, req)
	if err != nil {
		return nil, err
	}
	return resp.(*pb.RouteInfosResponse), nil
}

// 请求解码器 转换成rpc请求
func decodeGRPCGetRouteInfoWithNameRequest(_ context.Context, grpcReq interface{}) (interface{}, error) {
	res, ok := grpcReq.(*pb.GetRouteInfoWithNameRequest)

	if !ok {
		return nil, fmt.Errorf("decodeGRPCGetRouteInfoWithNameRequest invalid request type: %T", grpcReq)
	}

	req := endpoint.GetRouteInfoWithNameRequest{
		Name: res.Name,
	}

	return req, nil
}

// 响应编码器
func encodeGRPCGetRouteInfoWithNameResponse(_ context.Context, grpcResp interface{}) (interface{}, error) {
	res, ok := grpcResp.(endpoint.GetRouteInfoWithNameResponse)
	if !ok {
		return nil, fmt.Errorf("encodeGRPCGetRouteInfoWithNameResponse invalid response type: %T", grpcResp)
	}
	if res.Error != nil {
		return &pb.RouteInfosResponse{Routes: res.Routes, ErrorMes: res.Error.Error()}, res.Error
	}
	return &pb.RouteInfosResponse{Routes: res.Routes}, nil
}

// 实现 gRPC 投票结果服务接口
func (s *GrpcServer) GetRouteInfoWithPrefix(ctx context.Context, req *pb.GetRouteInfoWithPrefixRequest) (*pb.RouteInfosResponse, error) {
	_, resp, err := s.getRouteInfoWithPrefix.ServeGRPC(ctx, req)
	if err != nil {
		return nil, err
	}
	return resp.(*pb.RouteInfosResponse), nil
}

// 请求解码器 转换成rpc请求
func decodeGRPCGetRouteInfoWithPrefixRequest(_ context.Context, grpcReq interface{}) (interface{}, error) {
	res, ok := grpcReq.(*pb.GetRouteInfoWithPrefixRequest)

	if !ok {
		return nil, fmt.Errorf("decodeGRPCGetRouteInfoWithPrefixRequest invalid request type: %T", grpcReq)
	}

	req := endpoint.GetRouteInfoWithPrefixRequest{
		Name:   res.Name,
		Prefix: res.Prefix,
	}

	return req, nil
}

// 响应编码器
func encodeGRPCGetRouteInfoWithPrefixResponse(_ context.Context, grpcResp interface{}) (interface{}, error) {
	res, ok := grpcResp.(endpoint.GetRouteInfoWithPrefixResponse)
	if !ok {
		return nil, fmt.Errorf("encodeGRPCGetRouteInfoWithPrefixResponse invalid response type: %T", grpcResp)
	}
	if res.Error != nil {
		return &pb.RouteInfosResponse{Routes: res.Routes, ErrorMes: res.Error.Error()}, res.Error
	}
	return &pb.RouteInfosResponse{Routes: res.Routes}, nil
}

// 实现 gRPC 投票结果服务接口
func (s *GrpcServer) SetRouteRule(ctx context.Context, req *pb.RouteInfo) (*pb.SetRouteRuleResponse, error) {
	_, resp, err := s.setRouteRule.ServeGRPC(ctx, req)
	if err != nil {
		return nil, err
	}
	return resp.(*pb.SetRouteRuleResponse), nil
}

// 请求解码器 转换成rpc请求
func decodeGRPCSetRouteRuleRequest(_ context.Context, grpcReq interface{}) (interface{}, error) {
	res, ok := grpcReq.(*pb.RouteInfo)

	if !ok {
		return nil, fmt.Errorf("decodeGRPCSetRouteRuleRequest invalid request type: %T", grpcReq)
	}

	req := endpoint.SetRouteRuleRequest{
		Name:     res.Name,
		Host:     res.Host,
		Port:     res.Port,
		Prefix:   res.Prefix,
		Metadata: res.Metadata,
	}

	return req, nil
}

// 响应编码器
func encodeGRPCSetRouteRuleResponse(_ context.Context, grpcResp interface{}) (interface{}, error) {
	res, ok := grpcResp.(endpoint.SetRouteRuleResponse)
	if !ok {
		return nil, fmt.Errorf("encodeGRPCSetRouteRuleResponse invalid response type: %T", grpcResp)
	}
	if res.Error != nil {
		return &pb.SetRouteRuleResponse{ErrorMes: res.Error.Error()}, res.Error
	}
	return &pb.RouteInfosResponse{}, nil
}
