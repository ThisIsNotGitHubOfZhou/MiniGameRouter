package api

import (
	"context"
	"fmt"
	healthcheckpb "github.com/ThisIsNotGitHubOfZhou/MiniGameRouter/sdk/proto/healthcheck"
	registerpb "github.com/ThisIsNotGitHubOfZhou/MiniGameRouter/sdk/proto/register"
	"github.com/ThisIsNotGitHubOfZhou/MiniGameRouter/sdk/service"
	grpctransport "github.com/go-kit/kit/transport/grpc"
	"google.golang.org/grpc"
	"net/http"
	"time"
)

type MiniClient struct {
	name       string
	ip         string
	healthport string
	timeout    int
}

var _ service.RegisterService = (*MiniClient)(nil)

var _ service.DiscoverService = (*MiniClient)(nil)

var _ service.HealthCheckService = (*MiniClient)(nil)

func NewMiniClient(name, ip string, timeout int) *MiniClient {
	return &MiniClient{
		name:    name,
		timeout: timeout,
		ip:      ip,
	}
}

func (c *MiniClient) Register(ctx context.Context, name, host, port, protocol, metadata string, weight, timeout int) (string, error) {
	conn, err := grpc.Dial(RegisteGrpcrHost+":"+RegisterGrpcPort, grpc.WithInsecure())
	if err != nil {
		return "", err
	}
	defer conn.Close()

	//clientTracer := kitzipkin.GRPCClientTrace(config.ZipkinTracer)

	//// 使用 go-kit 的 gRPC 客户端传输层~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~
	var ep = grpctransport.NewClient(
		conn,
		"register.RegisterService", // 服务名称,注意前面要带包名！！！！！包名+在proto文件里定义的服务名
		"Register",                 // 方法名称
		encodeGRPCRegisterRequest,
		decodeGRPCRegisterResponse,
		registerpb.RegisterResponse{},
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
		Timeout:  int64(timeout),
	}
	response, err := ep(ctx, request)
	if err != nil {
		fmt.Println("sdk_api_impl_Register grpc error", err)
		return "", err
	}
	r := response.(*registerpb.RegisterResponse)

	fmt.Println(r)
	return r.Id, nil

	// 原版grpc请求~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~
	//cli := registerpb.NewRegisterServiceClient(conn)
	//ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	//defer cancel()
	//
	//request := &registerpb.RegisterRequest{
	//	Name:     name,
	//	Host:     host,
	//	Port:     port,
	//	Protocol: protocol,
	//	Metadata: metadata,
	//	Weight:   int64(weight),
	//	Timout:   int64(timeout),
	//}
	//r, err := cli.Register(ctx, request)
	//
	//if err != nil {
	//	fmt.Println("sdk_api_impl_Register error", err)
	//	return r.Id, err
	//} else {
	//	return r.Id, err
	//}
	// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

}
func encodeGRPCRegisterRequest(_ context.Context, request interface{}) (interface{}, error) {
	req := request.(*registerpb.RegisterRequest)
	return req, nil
}
func decodeGRPCRegisterResponse(_ context.Context, response interface{}) (interface{}, error) {
	resp := response.(*registerpb.RegisterResponse)
	return resp, nil
}

func (c *MiniClient) DeRegister(ctx context.Context, id, name, host, port string) error {
	conn, err := grpc.Dial(RegisteGrpcrHost+":"+RegisterGrpcPort, grpc.WithInsecure())
	if err != nil {
		return err
	}
	defer conn.Close()

	//clientTracer := kitzipkin.GRPCClientTrace(config.ZipkinTracer)

	//// 使用 go-kit 的 gRPC 客户端传输层~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~
	var ep = grpctransport.NewClient(
		conn,
		"register.RegisterService", // 服务名称,注意前面要带包名！！！！！包名+在proto文件里定义的服务名
		"DeRegister",               // 方法名称
		encodeGRPCDeRegisterRequest,
		decodeGRPCDeRegisterResponse,
		registerpb.DeRegisterResponse{},
		//clientTracer,
	).Endpoint()

	// 使用端点进行调用grpc
	request := &registerpb.DeRegisterRequest{
		Id:   id,
		Name: name,
		Host: host,
		Port: port,
	}
	response, err := ep(ctx, request)
	if err != nil {
		fmt.Println("sdk_api_impl_Register grpc error", err)
		return err
	}
	r := response.(*registerpb.DeRegisterResponse)
	if r.ErrorMes != "" {
		return fmt.Errorf("%v", r.ErrorMes)
	}

	return nil

	// 原版grpc请求~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~
	//cli := registerpb.NewRegisterServiceClient(conn)
	//ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	//defer cancel()
	//
	//request := &registerpb.DeRegisterRequest{
	//	Id:   id,
	//	Name: name,
	//	Host: host,
	//	Port: port,
	//}
	//r, err := cli.DeRegister(ctx, request)
	//
	//if err != nil {
	//	fmt.Println("sdk_api_impl_DeRegister grpc error", err)
	//	return err
	//}
	//
	//if r.ErrorMes != "" {
	//	fmt.Println("sdk_api_impl_DeRegister error", r.ErrorMes)
	//	err = fmt.Errorf("%s", r.ErrorMes)
	//}
	//
	//return err
	// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~
}

func encodeGRPCDeRegisterRequest(_ context.Context, request interface{}) (interface{}, error) {
	req := request.(*registerpb.DeRegisterRequest)
	return req, nil
}
func decodeGRPCDeRegisterResponse(_ context.Context, response interface{}) (interface{}, error) {
	resp := response.(*registerpb.DeRegisterResponse)
	return resp, nil
}

func (c *MiniClient) HealthCheckS(ctx context.Context, port string) error {
	// 启动一个http服务一直返回200
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	http.ListenAndServe(":"+port, nil)

	conn, err := grpc.Dial(HealthCheckGrpcHost+":"+HealthCheckGrpcPort, grpc.WithInsecure())
	if err != nil {
		return err
	}
	defer conn.Close()

	//clientTracer := kitzipkin.GRPCClientTrace(config.ZipkinTracer)

	//// 使用 go-kit 的 gRPC 客户端传输层~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~
	var ep = grpctransport.NewClient(
		conn,
		"healthcheck.HealthCheckService", // 服务名称,注意前面要带包名！！！！！包名+在proto文件里定义的服务名
		"HealthCheckS",                   // 方法名称
		encodeGRPCHealthCheckSRequest,
		decodeGRPCHealthCheckSResponse,
		healthcheckpb.HealthCheckSResponse{},
		//clientTracer,
	).Endpoint()

	// 使用端点进行调用grpc
	request := &healthcheckpb.HealthCheckSRequest{
		Name:       c.name,
		InstanceID: "",
		Url:        c.ip + ":" + port + "/",
		Timeout:    int64(c.timeout),
	}
	response, err := ep(ctx, request)
	if err != nil {
		fmt.Println("[Error][sdk] sdk_api_impl_healthcheckS grpc error", err)
		return err
	}
	r := response.(*healthcheckpb.HealthCheckSResponse)
	if r.ErrorMes != "" {
		return fmt.Errorf("%v", r.ErrorMes)
	}

	return nil

}

func encodeGRPCHealthCheckSRequest(_ context.Context, request interface{}) (interface{}, error) {
	req := request.(*healthcheckpb.HealthCheckSRequest)
	return req, nil
}
func decodeGRPCHealthCheckSResponse(_ context.Context, response interface{}) (interface{}, error) {
	resp := response.(*healthcheckpb.HealthCheckSResponse)
	return resp, nil
}

func (c *MiniClient) HealthCheckC(ctx context.Context, id, name, port, ip string, timeout int) error {
	conn, err := grpc.Dial(HealthCheckGrpcHost+":"+HealthCheckGrpcPort, grpc.WithInsecure())
	if err != nil {
		return err
	}
	defer conn.Close()

	//clientTracer := kitzipkin.GRPCClientTrace(config.ZipkinTracer)

	//// 使用 go-kit 的 gRPC 客户端传输层~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~
	var ep = grpctransport.NewClient(
		conn,
		"healthcheck.HealthCheckService", // 服务名称,注意前面要带包名！！！！！包名+在proto文件里定义的服务名
		"HealthCheckC",                   // 方法名称
		encodeGRPCHealthCheckCRequest,
		decodeGRPCHealthCheckCResponse,
		healthcheckpb.HealthCheckCResponse{},
		//clientTracer,
	).Endpoint()

	go func() {
		ticker := time.NewTicker(3 * time.Duration(c.timeout) * time.Second)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				// 使用端点进行调用grpc
				request := &healthcheckpb.HealthCheckCRequest{
					Id:      id,
					Name:    name,
					Host:    c.ip,
					Port:    port,
					Timeout: int64(c.timeout),
				}
				response, err := ep(ctx, request)
				if err != nil {
					fmt.Println("[Error][sdk] sdk_api_impl_healthcheckc grpc error", err)
					return
				}
				r := response.(*healthcheckpb.HealthCheckCResponse)
				if r.ErrorMes != "" {
					return
				}
			case <-ctx.Done():
				fmt.Println("[Info][sdk] HealthCheckC loop stopped")
				return
			}
		}
	}()

	return nil

}

func encodeGRPCHealthCheckCRequest(_ context.Context, request interface{}) (interface{}, error) {
	req := request.(*healthcheckpb.HealthCheckCRequest)
	return req, nil
}
func decodeGRPCHealthCheckCResponse(_ context.Context, response interface{}) (interface{}, error) {
	resp := response.(*healthcheckpb.HealthCheckCResponse)
	return resp, nil
}
