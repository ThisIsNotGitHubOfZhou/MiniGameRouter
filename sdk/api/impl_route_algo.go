package api

import (
	"context"
	"encoding/json"
	"fmt"
	discoverpb "github.com/ThisIsNotGitHubOfZhou/MiniGameRouter/sdk/proto/discover"
	"github.com/stathat/consistent"
	"math/rand"
	"time"
)

func init() {
	rand.Seed(time.Now().UnixNano()) // 在程序启动时设置随机数种子
}

// 一致性哈希算法
// key: 请求的某个标识符
// routes的name一定相同，host+port不一定相同
// 避免每次都重新生成NeHashMap
func (mc *MiniClient) ConsistentHashingRouting(ctx context.Context, routes []*discoverpb.RouteInfo, key string) (*discoverpb.RouteInfo, error) {
	// 根据一致性哈希算法来从routes里面选一个route返回
	if len(routes) == 0 {
		return nil, fmt.Errorf("no routes available")
	}

	// 创建一致性哈希对象
	ch := consistent.New()

	// 添加路由信息到一致性哈希环
	for _, route := range routes {
		ch.Add(route.Name + route.Host + route.Port) // 用name+host+port作为key
	}

	// 获取最近的路由
	routeID, err := ch.Get(key)
	if err != nil {
		return nil, fmt.Errorf("no matching route found: %v", err)
	}

	for _, route := range routes {
		if route.Name+route.Host+route.Port == routeID {
			return route, nil
		}
	}

	return nil, fmt.Errorf("no matching route found")
}

func (mc *MiniClient) RandomRouting(ctx context.Context, routes []*discoverpb.RouteInfo) (*discoverpb.RouteInfo, error) {
	if len(routes) == 0 {
		return nil, fmt.Errorf("no routes available")
	}

	selectedIndex := rand.Intn(len(routes)) // 生成一个 0 到 len(routes)-1 之间的随机数
	return routes[selectedIndex], nil
}

type WeightedRoute struct {
	index  int // 到路由list的下标映射
	weight int
}

func (mc *MiniClient) WeightedRouting(ctx context.Context, routes []*discoverpb.RouteInfo) (*discoverpb.RouteInfo, error) {
	if len(routes) == 0 {
		return nil, fmt.Errorf("no routes available")
	}

	toRoutes := []WeightedRoute{}
	totalWeight := 0

	for i, route := range routes {
		var meta map[string]interface{}
		if err := json.Unmarshal([]byte(route.Metadata), &meta); err != nil {
			continue // 如果解析失败，跳过这个路由
		}

		weightFloat, ok := meta["weight"].(float64) // JSON 解析后的数字类型是 float64
		if !ok {
			continue // 如果没有 weight 字段，跳过这个路由
		}

		weight := int(weightFloat)
		toRoutes = append(toRoutes, WeightedRoute{index: i, weight: weight})
		totalWeight += weight
	}

	if totalWeight == 0 {
		return nil, fmt.Errorf("total weight is zero")
	}

	randomWeight := rand.Intn(totalWeight)

	for _, route := range toRoutes {
		if randomWeight < route.weight {
			return routes[route.index], nil
		}
		randomWeight -= route.weight
	}

	return nil, fmt.Errorf("failed to select a route based on weight")
}

func (mc *MiniClient) TargetedRouting(ctx context.Context, routes []*discoverpb.RouteInfo, val string) (*discoverpb.RouteInfo, error) {
	return mc.MetadataRouting(ctx, routes, "targeted", val)
}

func (mc *MiniClient) MetadataRouting(ctx context.Context, routes []*discoverpb.RouteInfo, key string, val string) (*discoverpb.RouteInfo, error) {
	tempRoutes := []*discoverpb.RouteInfo{}
	for _, route := range routes {
		var meta map[string]interface{}
		if err := json.Unmarshal([]byte(route.Metadata), &meta); err != nil {
			continue // 如果解析失败，跳过这个路由
		}

		targetVal, ok := meta[key].(string)
		if !ok {
			continue // 如果没有 key的字段 字段，跳过这个路由
		}
		if targetVal != val { // 如果不相同则也跳过
			continue
		}

		tempRoutes = append(tempRoutes, route)
	}

	if len(tempRoutes) == 0 {
		return nil, fmt.Errorf("no routes available")
	}

	selectedIndex := rand.Intn(len(tempRoutes)) // 生成一个 0 到 len(routes)-1 之间的随机数
	return tempRoutes[selectedIndex], nil
}
