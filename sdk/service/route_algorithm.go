package service

import (
	"context"
	discoverpb "github.com/ThisIsNotGitHubOfZhou/MiniGameRouter/sdk/proto/discover"
)

type RouteAlgorithm interface {
	ConsistentHashingRouting(ctx context.Context, routes []*discoverpb.RouteInfo, key string) (*discoverpb.RouteInfo, error)

	RandomRouting(ctx context.Context, routes []*discoverpb.RouteInfo) (*discoverpb.RouteInfo, error)

	WeightedRouting(ctx context.Context, routes []*discoverpb.RouteInfo) (*discoverpb.RouteInfo, error)

	TargetedRouting(ctx context.Context, routes []*discoverpb.RouteInfo, key string) (*discoverpb.RouteInfo, error)

	MetadataRouting(ctx context.Context, routes []*discoverpb.RouteInfo, key string, val string) (*discoverpb.RouteInfo, error) // 根据routeinfo里面的metadata匹配清空进行路由
}
