package endpoint

import (
	"context"
	pb "discoversvr/proto"
	"discoversvr/service"
	"github.com/go-kit/kit/endpoint"
)

// 定义所有的endpoint
type DiscoverEndpoint struct {
	DiscoverServiceWithName endpoint.Endpoint
	DiscoverServiceWithID   endpoint.Endpoint
	GetRouteInfoWithName    endpoint.Endpoint
	GetRouteInfoWithPrefix  endpoint.Endpoint
}

// 定义服务的请求和返回
type DiscoverServiceWithNameRequest struct {
	Name string `json:"name"`
}

type DiscoverServiceWithNameResponse struct {
	ServiceInfos []*pb.ServiceInfo `json:"service_infos"`
	Error        error             `json:"error"`
}

// 定义创建edpt
func MakeDiscoverServiceWithNameEndpoint(svc service.Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		// 请求转换成endpoint层的请求
		req := request.(DiscoverServiceWithNameRequest)

		// 调用service层服务
		infos, err := svc.DiscoverServiceWithName(req.Name)

		return DiscoverServiceWithNameResponse{ServiceInfos: infos, Error: err}, nil
	}
}

// 定义服务的请求和返回
type DiscoverServiceWithIDRequest struct {
	InstanceID string `json:"instance_id"`
}

type DiscoverServiceWithIDResponse struct {
	ServiceInfos []*pb.ServiceInfo `json:"service_infos"`
	Error        error             `json:"error"`
}

// 定义创建edpt
func MakeDiscoverServiceWithIDEndpoint(svc service.Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		// 请求转换成endpoint层的请求
		req := request.(DiscoverServiceWithIDRequest)

		// 调用service层服务
		infos, err := svc.DiscoverServiceWithID(req.InstanceID)

		return DiscoverServiceWithIDResponse{ServiceInfos: infos, Error: err}, nil
	}
}

// 定义服务的请求和返回
type GetRouteInfoWithNameRequest struct {
	Name string `json:"name"`
}

type GetRouteInfoWithNameResponse struct {
	Routes []*pb.RouteInfo `json:"routes"`
	Error  error           `json:"error"`
}

// 定义创建edpt
func MakeGetRouteInfoWithNameEndpoint(svc service.Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		// 请求转换成endpoint层的请求
		req := request.(GetRouteInfoWithNameRequest)

		// 调用service层服务
		routes, err := svc.GetRouteInfoWithName(req.Name)

		return GetRouteInfoWithNameResponse{Routes: routes, Error: err}, nil
	}
}

// 定义服务的请求和返回
type GetRouteInfoWithPrefixRequest struct {
	Name   string `json:"name"`
	Prefix string `json:"prefix"`
}

type GetRouteInfoWithPrefixResponse struct {
	Routes []*pb.RouteInfo `json:"routes"`
	Error  error           `json:"error"`
}

// 定义创建edpt
func MakeGetRouteInfoWithPrefixEndpoint(svc service.Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		// 请求转换成endpoint层的请求
		req := request.(GetRouteInfoWithPrefixRequest)

		// 调用service层服务
		routes, err := svc.GetRouteInfoWithPrefix(req.Name, req.Prefix)

		return GetRouteInfoWithPrefixResponse{Routes: routes, Error: err}, nil
	}
}
