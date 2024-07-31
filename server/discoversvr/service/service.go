package service

import (
	"discoversvr/config"
	"discoversvr/database"
	pb "discoversvr/proto"
)

type Service interface {
	// 根据服务名发现
	DiscoverServiceWithName(name string) ([]*pb.ServiceInfo, error)

	// 根据服务InstanceID返回
	DiscoverServiceWithID(instanceID string) ([]*pb.ServiceInfo, error)

	// 根据服务名返回路由
	GetRouteInfoWithName(name string) ([]*pb.RouteInfo, error)

	// 根据服务名+前缀返回路由
	GetRouteInfoWithPrefix(name string, prefix string) ([]*pb.RouteInfo, error)

	// 前缀路由(prefix)or定向路由(metadata)
	SetRouteRule(*pb.RouteInfo) error
}

// 定义中间键服务
type ServiceMiddleware func(Service) Service

type DiscoverService struct {
	routeInfoCache map[string]*pb.RouteInfo // 存储RoutInfo
	routeDirty     map[string]bool          // route信息是否dirty，方便后续
}

var _ Service = (*DiscoverService)(nil)

func (s *DiscoverService) DiscoverServiceWithName(name string) ([]*pb.ServiceInfo, error) {
	config.Logger.Println("[Info][discover] DiscoverServiceWithName begin")

	return nil, nil
}

func (s *DiscoverService) DiscoverServiceWithID(instanceID string) ([]*pb.ServiceInfo, error) {
	config.Logger.Println("[Info][discover] DiscoverServiceWithID begin")
	return nil, nil
}

func (s *DiscoverService) GetRouteInfoWithName(name string) ([]*pb.RouteInfo, error) {
	config.Logger.Println("[Info][discover] GetRouteInfoWithName begin")
	return nil, nil
}

func (s *DiscoverService) GetRouteInfoWithPrefix(name string, prefix string) ([]*pb.RouteInfo, error) {
	config.Logger.Println("[Info][discover] GetRouteInfoWithPrefix begin")
	return nil, nil
}

func (s *DiscoverService) SetRouteRule(info *pb.RouteInfo) error {
	config.Logger.Println("[Info][discover] GetRouteInfoWithPrefix begin")
	database.ReadFromMysql(nil)
	return nil
}
