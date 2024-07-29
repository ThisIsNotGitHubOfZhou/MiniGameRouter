package endpoint

import (
	"context"
	"github.com/go-kit/kit/endpoint"
	"healthchecksvr/service"
)

// 定义所有的endpoint
type HealthCheckEndpoint struct {
	HealthCheckS endpoint.Endpoint
	HealthCheckC endpoint.Endpoint
}

// 定义服务的请求和返回
type HealthCheckSRequest struct {
	Name       string `json:"name"`
	InstanceID string `json:"instance_id"`
	Url        string `json:"url"`
	Timeout    int    `json:"timeout"`
}

type HealthCheckSResponse struct {
	Error error `json:"error"`
}

// 定义创建edpt
func MakeHealthCheckSEndpoint(svc service.Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		// 请求转换成endpoint层的请求
		req := request.(HealthCheckSRequest)

		// 调用service层服务
		err := svc.HealthCheckS(req.Url, req.Name, req.Timeout)

		return HealthCheckSResponse{Error: err}, nil
	}
}

type HealthCheckCRequest struct {
	ID      string `json:"id"`
	Name    string `json:"name"`
	Host    string `json:"host"`
	Port    string `json:"port"`
	Timeout int    `json:"timeout"`
	//id, instanceID, host, port string, second int
}

type HealthCheckCResponse struct {
	Error error `json:"error"`
}

func MakeHealthCheckCEndpoint(svc service.Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (response interface{}, err error) {
		req := request.(HealthCheckCRequest)

		res := svc.HealthCheckC(req.ID, req.Name, req.Host, req.Port, req.Timeout)
		return HealthCheckCResponse{res}, nil
	}
}
