package plugins

import (
	pb "discoversvr/proto"
	"discoversvr/service"
	"time"

	"github.com/go-kit/log"
)

// 添加中间件
type loggingMiddleware struct {
	service.Service // 如果没有重新定义service则直接会调用这个service
	logger          log.Logger
}

// LoggingMiddleware make logging middleware
func LoggingMiddleware(logger log.Logger) service.ServiceMiddleware {
	return func(next service.Service) service.Service {
		return &loggingMiddleware{next, logger}
	}
}

func (mw *loggingMiddleware) DiscoverServiceWithName(name string) (res []*pb.ServiceInfo, err error) {
	// 函数执行结束后打印日志
	defer func(begin time.Time) {
		mw.logger.Log(
			"function", "DiscoverServiceWithName",
			"name", name,
			"service_len", len(res),
			"took", time.Since(begin),
		)
	}(time.Now())

	res, err = mw.Service.DiscoverServiceWithName(name)
	return
}

func (mw loggingMiddleware) DiscoverServiceWithID(instanceID string) (res []*pb.ServiceInfo, err error) {
	defer func(begin time.Time) {
		mw.logger.Log(
			"function", "DiscoverServiceWithID",
			"instance_id", instanceID,
			"service_len", len(res),
			"took", time.Since(begin),
		)
	}(time.Now())
	res, err = mw.Service.DiscoverServiceWithID(instanceID)
	return
}
func (mw loggingMiddleware) GetRouteInfoWithName(name string) (res []*pb.RouteInfo, err error) {
	defer func(begin time.Time) {
		mw.logger.Log(
			"function", "DiscoverServiceWithID",
			"name", name,
			"route_len", len(res),
			"took", time.Since(begin),
		)
	}(time.Now())
	res, err = mw.Service.GetRouteInfoWithName(name)
	return
}

func (mw loggingMiddleware) GetRouteInfoWithPrefix(name string, prefix string) (res []*pb.RouteInfo, err error) {
	defer func(begin time.Time) {
		mw.logger.Log(
			"function", "GetRouteInfoWithPrefix",
			"name", name,
			"prefix", prefix,
			"route_len", len(res),
			"took", time.Since(begin),
		)
	}(time.Now())
	res, err = mw.Service.GetRouteInfoWithPrefix(name, prefix)
	return
}

func (mw loggingMiddleware) SetRouteRule(info *pb.RouteInfo) (err error) {
	defer func(begin time.Time) {
		mw.logger.Log(
			"function", "SetRouteRule",
			"name", info.Name,
			"prefix", info.Prefix,
			"metadata", info.Metadata,
			"took", time.Since(begin),
		)
	}(time.Now())
	err = mw.Service.SetRouteRule(info)
	return
}
