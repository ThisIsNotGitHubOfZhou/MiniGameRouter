package api

import (
	"context"
	"fmt"
	healthcheckpb "github.com/ThisIsNotGitHubOfZhou/MiniGameRouter/sdk/proto/healthcheck"
	grpctransport "github.com/go-kit/kit/transport/grpc"
	"google.golang.org/grpc"
	"net/http"
	"time"
)

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

	conn, err := grpc.Dial(HealthCheckGrpcHost+":"+HealthCheckGrpcPort, grpc.WithInsecure()) // NOTE:应该不需要从连接池里面拿，因为只发送一次
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
		// 轮询服务
		c.healthCheckLock.Lock()
		c.healthCheckFlag++
		tempFlag := c.registerFlag
		c.healthCheckLock.Unlock()

		if len(c.HealthCheckGRPCPools) == 0 {
			fmt.Println("[Error][sdk] HealthCheckGRPCPools为空")
			return
		}
		conn, err := c.HealthCheckGRPCPools[tempFlag%(int64(len(c.HealthCheckGRPCPools)))].Get() // 优化后
		defer c.HealthCheckGRPCPools[tempFlag%(int64(len(c.HealthCheckGRPCPools)))].Put(conn)

		//conn, err := grpc.Dial(HealthCheckGrpcHost+":"+HealthCheckGrpcPort, grpc.WithInsecure())
		if err != nil {
			fmt.Println("[Error][sdk] gprc连接问题：", err)
			return
		}
		//defer conn.Close()

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
