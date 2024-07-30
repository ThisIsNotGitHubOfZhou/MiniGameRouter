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
	id         string
	host       string
	port       string
	protocol   string
	metadata   string
	weight     int
	timeout    int
	healthport string
}

var _ service.RegisterService = (*MiniClient)(nil)

var _ service.DiscoverService = (*MiniClient)(nil)

var _ service.HealthCheckService = (*MiniClient)(nil)

func NewMiniClient(name, host, port, protocol, metadata string, weight, timeout int) *MiniClient {
	return &MiniClient{
		name:     name,
		host:     host,
		port:     port,
		protocol: protocol,
		metadata: metadata,
		weight:   weight,
		timeout:  timeout,
	}
}

func (c *MiniClient) Name() string {
	return c.name
}

func (c *MiniClient) ID() string {
	return c.id
}

func (c *MiniClient) Host() string {
	return c.host
}

func (c *MiniClient) Port() string {
	return c.port
}

func (c *MiniClient) Protocol() string {
	return c.protocol
}

func (c *MiniClient) Metadata() string {
	return c.metadata
}

func (c *MiniClient) Weight() int {
	return c.weight
}

func (c *MiniClient) Timeout() int {
	return c.timeout
}

func (c *MiniClient) HealthPort() string {
	return c.healthport
}

func (c *MiniClient) Register(ctx context.Context, name, host, port, protocol, metadata string, weight, timeout int) (string, error) {
	fmt.Println("[Info][sdk] Register，注册服务:", name)
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

	fmt.Println("[Info][sdk]  register 结果", r)
	c.id = r.Id
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
	fmt.Println("[Info][sdk] DeRegister，删除服务:", id, name)
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

	go func() {
		http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("OK"))
		})
		fmt.Println("[Info][sdk] 启动healthcheck 端口:", port)
		http.ListenAndServe("0.0.0.0:"+port, nil)

	}()

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
		Url:        "http://" + c.host + ":" + port + "/",
		Timeout:    int64(c.timeout),
	}
	response, err := ep(ctx, request)
	if err != nil {
		fmt.Println("[Error][sdk] healthcheckS grpc 出错：", err)
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

// TODO:优化GRPC连接问题
func (c *MiniClient) HealthCheckC(ctx context.Context, id, name, port, ip string, timeout int) error {

	go func() {
		conn, err := grpc.Dial(HealthCheckGrpcHost+":"+HealthCheckGrpcPort, grpc.WithInsecure())
		if err != nil {
			fmt.Println("[Error][sdk] gprc连接问题：", err)
			return
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
		ticker := time.NewTicker(3 * time.Duration(c.timeout) * time.Second)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				// 使用端点进行调用grpc

				request := &healthcheckpb.HealthCheckCRequest{
					Id:      id,
					Name:    name,
					Host:    c.host,
					Port:    port,
					Timeout: int64(c.timeout),
				}
				fmt.Println("[Info][sdk] healthcheckc 发送续约请求：", request)
				response, err := ep(ctx, request)
				if err != nil {
					fmt.Println("[Error][sdk] healthcheckc grpc 出错：", err)
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
