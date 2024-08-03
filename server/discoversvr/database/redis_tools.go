package database

import (
	"context"
	"discoversvr/config"
	pb "discoversvr/proto"
	"github.com/go-redis/redis/v8"
	"strings"
	"time"
)

var ctx = context.Background()

// 发现服务实例
func DiscoverServices(client *redis.Client, pattern string) ([]map[string]string, error) {
	pattern = "*" + pattern + "*" // 因为Key是由服务名+IP+Port组成的
	keys, err := client.Keys(ctx, pattern).Result()
	if err != nil {
		return nil, err
	}

	services := make([]map[string]string, 0, len(keys))
	for _, key := range keys {
		serviceInfo, err := client.HGetAll(ctx, key).Result()
		if err != nil {
			continue
		}
		services = append(services, serviceInfo)
	}

	return services, nil
}

func SyncRoutesWithRouteSyncRequest(client *redis.Client, req *pb.RouteSyncRequest) []*pb.RouteInfo {
	// TODO:同步的信息就放在1号redis数据库吧

	config.Logger.Println("[Info][discover]   SyncRoutesWithRouteSyncRequest时间戳：", req.LastSyncVersion)
	res := make([]*pb.RouteInfo, 0)
	// name、namePrefix需要比reqTime晚~
	for _, key := range req.Name {
		redisData := readCacheWithName(client, key, req.LastSyncVersion)
		if len(redisData) > 0 {
			res = append(res, redisData...)
		} else {
			// 从mysql里面读
			mysqlData, err := ReadFromMysqlWithName(key)
			if err != nil {
				config.Logger.Println("[Error][discover]   redis cache失败ReadFromMysqlWithName，从mysql读出错:", err)
				continue
			}
			if len(mysqlData) == 0 {
				config.Logger.Println("[Warning][discover]   mysql里面路由被删了(name):", key)
				continue
			}
			res = append(res, mysqlData...)
		}

	}

	for _, key := range req.NamePrefix {
		redisData := readCacheWithNamePrefix(client, key, req.LastSyncVersion)
		if len(redisData) > 0 {
			res = append(res, redisData...)
		} else {
			// 从mysql里面读

			// 提取出name和prefix
			parts := strings.Split(key, ":")
			if len(parts) != 2 {
				config.Logger.Println("[Error][discover]   redis cache失败,request.NamePrefix格式不对:", key)
				continue
			}
			name, prefix := parts[0], parts[1]
			mysqlData, err := ReadFromMysqlWithPrefix(name, prefix)
			if err != nil {
				config.Logger.Println("[Error][discover]   redis cache失败ReadFromMysqlWithName，从mysql读出错:", err)
				continue
			}
			if len(mysqlData) == 0 {
				config.Logger.Println("[Warning][discover]   mysql里面路由被删了(nameprefix):", key)
				continue
			}
			res = append(res, mysqlData...)
		}

	}

	// nameNew、namePrefixNew必须要读到不管新旧~
	for _, key := range req.Name {
		redisData := readCacheWithName(client, key, "")
		if len(redisData) > 0 {
			res = append(res, redisData...)
		} else {
			// 从mysql里面读
			mysqlData, err := ReadFromMysqlWithName(key)
			if err != nil {
				config.Logger.Println("[Error][discover]   redis cache失败ReadFromMysqlWithName，从mysql读出错:", err)
				continue
			}
			if len(mysqlData) == 0 {
				config.Logger.Println("[Warning][discover]   mysql里面路由被删了(name):", key)
				continue
			}
			res = append(res, mysqlData...)
		}

	}

	for _, key := range req.NamePrefix {
		redisData := readCacheWithNamePrefix(client, key, "")
		if len(redisData) > 0 {
			res = append(res, redisData...)
		} else {
			// 从mysql里面读

			// 提取出name和prefix
			parts := strings.Split(key, ":")
			if len(parts) != 2 {
				config.Logger.Println("[Error][discover]   redis cache失败,request.NamePrefix格式不对:", key)
				continue
			}
			name, prefix := parts[0], parts[1]
			mysqlData, err := ReadFromMysqlWithPrefix(name, prefix)
			if err != nil {
				config.Logger.Println("[Error][discover]   redis cache失败ReadFromMysqlWithName，从mysql读出错:", err)
				continue
			}
			if len(mysqlData) == 0 {
				config.Logger.Println("[Warning][discover]   mysql里面路由被删了(nameprefix):", key)
				continue
			}
			res = append(res, mysqlData...)
		}

	}

	return res
}

// 从redis缓存里面读取name的cache（所有名字里带name的）
func readCacheWithName(client *redis.Client, name string, lastSyncVersion string) []*pb.RouteInfo {
	return readCache(client, "*"+name+"*", lastSyncVersion) // 需要全量匹配，去重由sdk去做吧~
}

// 从redis缓存里面读取name:prefix的cache
func readCacheWithNamePrefix(client *redis.Client, namePrefix string, lastSyncVersion string) []*pb.RouteInfo {
	return readCache(client, namePrefix, lastSyncVersion)
}

// 从redis缓存里面读取带有特定前缀的name的cache
func readCache(client *redis.Client, pattern string, lastSyncVersion string) []*pb.RouteInfo {
	keys, err := client.Keys(ctx, pattern).Result()
	if err != nil {
		config.Logger.Printf("[Error][discover][redis] Error fetching keys with pattern %s: %v", pattern, err)
		return nil
	}
	var timeLimit time.Time
	if lastSyncVersion != "" { // 有时间的话需要更新一下
		timeLimit, err = time.Parse(time.RFC3339, lastSyncVersion)
		if err != nil {
			config.Logger.Println("[Error][discover]   SyncRoutesWithRouteSyncRequest 转换时间出错", err)
		}
	}
	res := make([]*pb.RouteInfo, 0, len(keys))
	for _, key := range keys {
		serviceInfo, err := client.HGetAll(ctx, key).Result()
		if err != nil {
			config.Logger.Printf("[Error][discover][redis] Error fetching redis cache data for key %s: %v\n", key, err)
			continue
		}
		if lastSyncVersion != "" { // 对比时间
			routeTime, err := time.Parse(time.RFC3339, serviceInfo["lastSyncVersion"])
			if err != nil || routeTime.Before(timeLimit) {
				config.Logger.Println("[Error][discover]   redis lastSyncVersion错误或者不是更新版。", err)
				continue
			}
		}
		route := &pb.RouteInfo{
			Name:     serviceInfo["name"],
			Host:     serviceInfo["host"],
			Port:     serviceInfo["port"],
			Prefix:   serviceInfo["prefix"],
			Metadata: serviceInfo["metadata"],
		}
		res = append(res, route)
	}

	return res
}

// 性能提优
// 使用 SCAN 命令从 Redis 缓存中读取带有特定前缀的 name 的 cache
func readCacheWithScan(client *redis.Client, pattern string) []*pb.RouteInfo {
	var cursor uint64
	var keys []string
	var err error

	res := make([]*pb.RouteInfo, 0)

	for {
		keys, cursor, err = client.Scan(ctx, cursor, pattern, 10).Result()
		if err != nil {
			config.Logger.Printf("[Error][discover][redis] Error scanning keys with pattern %s: %v", pattern, err)
			return nil
		}

		for _, key := range keys {
			serviceInfo, err := client.HGetAll(ctx, key).Result()
			if err != nil {
				config.Logger.Printf("[Error][discover][redis] Error fetching redis cache data for key %s: %v\n", key, err)
				continue
			}
			route := &pb.RouteInfo{
				Name:     serviceInfo["name"],
				Host:     serviceInfo["host"],
				Port:     serviceInfo["port"],
				Prefix:   serviceInfo["prefix"],
				Metadata: serviceInfo["metadata"],
			}
			res = append(res, route)
		}

		if cursor == 0 {
			break
		}
	}

	return res
}
