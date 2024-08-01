package api

import (
	"context"
	"fmt"
	registerpb "github.com/ThisIsNotGitHubOfZhou/MiniGameRouter/sdk/proto/register"
	grpctransport "github.com/go-kit/kit/transport/grpc"
)

func (c *MiniClient) Register(ctx context.Context, name, host, port, protocol, metadata string, weight, timeout int) (string, error) {
	fmt.Println("[Info][sdk] Register，注册服务:", name)
	// 轮询服务
	c.registerLock.Lock()
	c.registerFlag++
	tempFlag := c.registerFlag
	c.registerLock.Unlock()

	if len(c.RegisterGRPCPools) == 0 {
		fmt.Println("[Error][sdk] RegisterGRPCPools为空")
		return "", fmt.Errorf("RegisterGRPCPools empty")
	}
	conn, err := c.RegisterGRPCPools[tempFlag%(int64(len(c.RegisterGRPCPools)))].Get() // 优化后
	defer c.RegisterGRPCPools[tempFlag%(int64(len(c.RegisterGRPCPools)))].Put(conn)

	// conn, err := grpc.Dial(RegisteGrpcrHost+":"+RegisterGrpcPort, grpc.WithInsecure())
	if err != nil {
		return "", err
	}

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
	// 轮询服务
	c.registerLock.Lock()
	c.registerFlag++
	tempFlag := c.registerFlag
	c.registerLock.Unlock()

	if len(c.RegisterGRPCPools) == 0 {
		fmt.Println("[Error][sdk] RegisterGRPCPools为空")
		return fmt.Errorf("RegisterGRPCPools empty")
	}
	conn, err := c.RegisterGRPCPools[tempFlag%(int64(len(c.RegisterServerInfo)))].Get()
	defer c.RegisterGRPCPools[tempFlag%(int64(len(c.RegisterServerInfo)))].Put(conn)

	if err != nil {
		return err
	}

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
