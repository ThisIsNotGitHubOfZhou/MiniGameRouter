package service

import (
	"discoversvr/config"
	"discoversvr/database"
	pb "discoversvr/proto"
	"discoversvr/tools"
	"fmt"
	"strconv"
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

	// 同步路由
	SyncRoutes(req *pb.RouteSyncRequest, stream pb.DiscoverService_SyncRoutesServer) error
}

// 定义中间键服务
type ServiceMiddleware func(Service) Service

type DiscoverService struct {
	// TODO:用这些优化~避免每次都全量发送~
	routeInfoCache map[string]*pb.RouteInfo // 存储RoutInfo
	routeDirty     map[string]bool          // route信息是否dirty，方便后续
}

var _ Service = (*DiscoverService)(nil)

func (s *DiscoverService) DiscoverServiceWithName(name string) ([]*pb.ServiceInfo, error) {
	config.Logger.Println("[Info][discover] DiscoverServiceWithName begin", name)
	rawData, err := database.DiscoverServices(config.RedisClient, name)
	if err != nil {
		config.Logger.Println("[Error][discover] DiscoverServiceWithName error with redis:", err)

	}
	var serviceInfos []*pb.ServiceInfo
	for _, item := range rawData {
		serviceInfo, err := convertMapToServiceInfo(item)
		if err != nil {
			return nil, err
		}
		serviceInfos = append(serviceInfos, serviceInfo)
	}
	return serviceInfos, nil
}

// convertMapToServiceInfo 将 map[string]string 转换为 *pb.ServiceInfo
// 需要跟registersvr对齐~
func convertMapToServiceInfo(data map[string]string) (*pb.ServiceInfo, error) {
	weight, err := strconv.ParseInt(data["weight"], 10, 64)
	if err != nil {
		return nil, fmt.Errorf("invalid weight: %v", err)
	}

	timeout, err := strconv.ParseInt(data["timeout"], 10, 64)
	if err != nil {
		return nil, fmt.Errorf("invalid timeout: %v", err)
	}

	return &pb.ServiceInfo{
		ServiceName: data["service_name"],
		InstanceId:  data["instance_id"],
		Host:        data["host"],
		Port:        data["port"],
		Protocol:    data["protocol"],
		Weight:      weight,
		Timeout:     timeout,
		Metadata:    data["metadata"],
	}, nil
}

func (s *DiscoverService) DiscoverServiceWithID(instanceID string) ([]*pb.ServiceInfo, error) {
	config.Logger.Println("[Info][discover] DiscoverServiceWithID begin:", instanceID)
	// 直接复用
	return s.DiscoverServiceWithName(instanceID)
}

func (s *DiscoverService) GetRouteInfoWithName(name string) ([]*pb.RouteInfo, error) {
	config.Logger.Println("[Info][discover] GetRouteInfoWithName begin")
	return database.ReadFromMysqlWithName(name)
}

func (s *DiscoverService) GetRouteInfoWithPrefix(name string, prefix string) ([]*pb.RouteInfo, error) {
	config.Logger.Println("[Info][discover] GetRouteInfoWithPrefix begin")
	return database.ReadFromMysqlWithPrefix(name, prefix)
}

func (s *DiscoverService) SetRouteRule(info *pb.RouteInfo) error {
	config.Logger.Println("[Info][discover] SetRouteRule begin")
	return database.WriteToMysql(info)
}

func (s *DiscoverService) SyncRoutes(req *pb.RouteSyncRequest, stream pb.DiscoverService_SyncRoutesServer) error {
	config.Logger.Println("[Info][discover] SyncRoutes begin")
	if tools.SyncRouteUpdates == nil { // 确保通道已经初始化
		return fmt.Errorf("SyncRouteUpdates channel is not initialized")
	}

	for {
		select {
		case route, ok := <-tools.SyncRouteUpdates:
			if !ok {
				return fmt.Errorf("SyncRouteUpdates channel closed")
			}
			if err := stream.Send(route); err != nil {
				return err
			}
		case <-stream.Context().Done():
			config.Logger.Println("[Info][discover] SyncRoutes end", req)
			return stream.Context().Err()
		}
	}
}
