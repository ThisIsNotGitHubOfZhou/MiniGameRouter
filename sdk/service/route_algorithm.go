package service

import (
	"context"
)

type RouteAlgorithm interface {
	ConsistentHashingRouting(ctx context.Context) error

	RandomRouting(ctx context.Context) error

	WeightedRouting(ctx context.Context) error

	TargetedRouting(ctx context.Context) error

	MetadataRouting(ctx context.Context) error // 根据routeinfo里面的metadata匹配清空进行路由
}
