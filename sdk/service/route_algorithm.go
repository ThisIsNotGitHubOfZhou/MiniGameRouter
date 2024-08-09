package service

import (
	"context"
	discoverpb "github.com/ThisIsNotGitHubOfZhou/MiniGameRouter/sdk/proto/discover"
)

type RouteAlgorithmType int

const (
	ConsistentHashing RouteAlgorithmType = iota
	Random
	Weighted
	Targeted
	Metadata
)

type RouteRequest struct {
	RouteType   RouteAlgorithmType
	Name        string
	Prefix      string
	TargetedKey string
	MetaKey     string
	MetaVal     string
}

type RouteAlgorithm interface {
	Routing(ctx context.Context, routeReq RouteRequest) (*discoverpb.RouteInfo, error)

	ConsistentHashingRouting(ctx context.Context, routes []*discoverpb.RouteInfo, key string) (*discoverpb.RouteInfo, error)

	RandomRouting(ctx context.Context, routes []*discoverpb.RouteInfo) (*discoverpb.RouteInfo, error)

	WeightedRouting(ctx context.Context, routes []*discoverpb.RouteInfo) (*discoverpb.RouteInfo, error)

	TargetedRouting(ctx context.Context, routes []*discoverpb.RouteInfo, key string) (*discoverpb.RouteInfo, error)

	MetadataRouting(ctx context.Context, routes []*discoverpb.RouteInfo, key string, val string) (*discoverpb.RouteInfo, error) // 根据routeInfo里面的metadata匹配清空进行路由
}
