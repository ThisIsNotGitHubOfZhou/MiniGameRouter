package service

import (
	pb "discoversvr/proto"
)

type Service interface {
	// 根据服务名发现
	// TODO:返回服务实例
	DiscoverServiceWithName(name string) ([]*pb.ServiceInfo, error)

	// 根据服务InstanceID返回
	DiscoverServiceWithID(instanceID string) ([]*pb.ServiceInfo, error)

	// 根据服务名返回路由
	GetRouteInfoWithName(name string) ([]*pb.RouteInfo, error)

	// 根据服务名+前缀返回路由
	GetRouteInfoWithPrefix(name string, prefix string) ([]*pb.RouteInfo, error)
}

// 定义中间键服务
type ServiceMiddleware func(Service) Service

type DiscoverService struct {
	routeInfoCache map[string]pb.RouteInfo // 存储RoutInfo
}

var _ Service = (*DiscoverService)(nil)

func (s *DiscoverService) DiscoverServiceWithName(name string) ([]*pb.ServiceInfo, error) {
	return nil, nil
}

func (s *DiscoverService) DiscoverServiceWithID(instanceID string) ([]*pb.ServiceInfo, error) {
	return nil, nil
}

func (s *DiscoverService) GetRouteInfoWithName(name string) ([]*pb.RouteInfo, error) {
	return nil, nil
}

func (s *DiscoverService) GetRouteInfoWithPrefix(name string, prefix string) ([]*pb.RouteInfo, error) {
	return nil, nil
}
