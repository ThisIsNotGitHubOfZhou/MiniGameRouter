package endpoint

import (
	"context"
	"github.com/go-kit/kit/endpoint"
	"registersvr/service"
)

// 定义所有的endpoint
type RegisterEndpoint struct {
	Register   endpoint.Endpoint
	DeRegister endpoint.Endpoint
}

// 定义服务的请求和返回
type RegisterRequest struct {
	Name     string `json:"name"`
	Host     string `json:"host"`
	Port     string `json:"port"`
	Protocol string `json:"protocol"`
	Metadata string `json:"metadata"`
	Weight   int    `json:"weight"`
	Timeout  int    `json:"timeout"`
}

type RegisterResponse struct {
	InstanceID string `json:"instance_id"`
	Error      error  `json:"error"`
}

// 定义创建edpt
func MakeRegisterEndpoint(svc service.Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		// 请求转换成endpoint层的请求
		req := request.(RegisterRequest)

		// 调用service层服务
		id, err := svc.Register(req.Name, req.Host, req.Port, req.Protocol, req.Metadata, req.Weight, req.Timeout)

		return RegisterResponse{InstanceID: id, Error: err}, nil
	}
}

type DeRegisterRequest struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	Host string `json:"host"`
	Port string `json:"port"`
	//id, name, host, port string
}

type DeRegisterResponse struct {
	Error error `json:"error"`
}

func MakeDeRegisterEndpoint(svc service.Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (response interface{}, err error) {
		req := request.(DeRegisterRequest)

		res := svc.Deregister(req.ID, req.Name, req.Host, req.Port)
		return DeRegisterResponse{res}, nil
	}
}

// // HealthRequest 健康检查请求结构
// type HealthCheckRequest struct{}

// // HealthResponse 健康检查响应结构
// type HealthCheckResponse struct {
// 	Status bool `json:"status"`
// }

// // MakeHealthCheckEndpoint 创建健康检查Endpoint
// func MakeHealthCheckEndpoint(svc service.Service) endpoint.Endpoint {
// 	return func(ctx context.Context, request interface{}) (response interface{}, err error) {
// 		status := svc.HealthCheck()
// 		return HealthCheckResponse{status}, nil
// 	}
// }
