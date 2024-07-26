package api

import (
	"context"
	"flag"
	"fmt"
	grpctransport "github.com/go-kit/kit/transport/grpc"
	"google.golang.org/grpc"
	registerpb "sdk/proto"
	"sdk/service"
)

var (
	RegisteGrpcrHost string
	RegisterGrpcPort string
	IsK8s            bool
)

func init() {
	// 定义命令行标志
	flag.StringVar(&RegisteGrpcrHost, "host", "10.76.143.1", "The host to register grpc")
	flag.StringVar(&RegisterGrpcPort, "port", "20001", "The port to register grpc")
	flag.BoolVar(&IsK8s, "k8s", false, "Is running in Kubernetes")
}

type MiniClient struct {
}

var _ service.RegisterService = (*MiniClient)(nil)

var _ service.DiscoverService = (*MiniClient)(nil)

var _ service.HealthCheckService = (*MiniClient)(nil)

func NewMiniClient() *MiniClient {
	return &MiniClient{}
}

func (c *MiniClient) Register(ctx context.Context, name, host, port, protocol, metadata string, weight, timeout int) (string, error) {
	conn, err := grpc.Dial(RegisteGrpcrHost+":"+RegisterGrpcPort, grpc.WithInsecure())
	//conn, err := grpc.Dial("9.135.95.71:50051", grpc.WithInsecure())
	if err != nil {
		return "", err
	}
	defer conn.Close()

	//clientTracer := kitzipkin.GRPCClientTrace(config.ZipkinTracer)

	// 使用 go-kit 的 gRPC 客户端传输层
	var ep = grpctransport.NewClient(
		conn,
		"register.RegisterServiceServer", // 服务名称,注意前面要带包名！！！！！
		"Register",                       // 方法名称
		encodeGRPCRegisterRequest,
		decodeGRPCRegisterResponse,
		registerpb.RegisterRequest{},
		//clientTracer,
	).Endpoint()

	// 使用端点进行调用grpc
	request := &registerpb.RegisterRequest{
		Name:     name,
		Host:     host,
		Port:     port,
		Protocol: protocol,
		Metadata: metadata,
		Weight:   int64(weight),
		Timout:   int64(timeout),
	}
	response, err := ep(ctx, request)
	r := response.(*registerpb.RegisterResponse)

	fmt.Println(r)
	return "", nil
}
func encodeGRPCRegisterRequest(_ context.Context, request interface{}) (interface{}, error) {
	req := request.(*registerpb.RegisterRequest)
	return req, nil
}
func decodeGRPCRegisterResponse(_ context.Context, response interface{}) (interface{}, error) {
	resp := response.(*registerpb.RegisterRequest)
	return resp, nil
}

func (c *MiniClient) Deregister(username, password string) error {
	return nil
}
