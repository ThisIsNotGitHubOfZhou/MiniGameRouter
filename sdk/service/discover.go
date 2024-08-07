package service

import (
	"context"
	discoverpb "github.com/ThisIsNotGitHubOfZhou/MiniGameRouter/sdk/proto/discover"
)

type DiscoverService interface {
	// 根据服务名发现
	DiscoverServiceWithName(ctx context.Context, name string) ([]*discoverpb.ServiceInfo, error)

	// 根据服务InstanceID返回
	DiscoverServiceWithID(ctx context.Context, instanceID string) ([]*discoverpb.ServiceInfo, error)

	// 根据服务名返回路由
	GetRouteInfoWithName(ctx context.Context, name string) ([]*discoverpb.RouteInfo, error)

	// 根据服务名+前缀返回路由
	GetRouteInfoWithPrefix(ctx context.Context, name string, prefix string) ([]*discoverpb.RouteInfo, error)

	// 前缀路由(prefix)or定向路由(metadata)
	SetRouteRule(ctx context.Context, info *discoverpb.RouteInfo) error

	// 前缀路由(prefix)or定向路由(metadata)
	UpdateRouteRule(ctx context.Context, name, host, port, prefix string, info *discoverpb.RouteInfo) error
}
