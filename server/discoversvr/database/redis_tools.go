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

// ~~~~~~~~~~~~~~~~~~~~cache相关~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~
// 写一份cache到redis，带时间戳的，读取的时候写入

func WriteSyncRoutes(client *redis.Client, routes []*pb.RouteInfo) {
	ctx := context.Background() // 创建一个上下文
	startTime := time.Now()     // 记录开始时间

	//config.Logger.Printf("[Info][discover][redis] WriteSyncRoutes len: %v\n", len(routes))
	for _, route := range routes {
		cacheInfo := make(map[string]string)
		cacheInfo["lastSyncVersion"] = time.Now().Format(time.RFC3339)
		cacheInfo["name"] = route.Name
		cacheInfo["host"] = route.Host
		cacheInfo["port"] = route.Port
		cacheInfo["prefix"] = route.Prefix
		cacheInfo["metadata"] = route.Metadata

		// 将cacheInfo写入cache，并设置一个小时的过期时间
		var key string
		if route.Prefix != "" {
			key = route.Host + "-" + route.Port + ":" + route.Name + ":" + route.Prefix
		} else {
			key = route.Host + "-" + route.Port + ":" + route.Name
		}

		// 使用 HMSet 写入哈希表
		err := client.HMSet(ctx, key, cacheInfo).Err()
		if err != nil {
			// 处理错误
			config.Logger.Printf("[Error][discover][reds]Failed to write to Redis: %v\n", err)
			continue
		}

		// 设置过期时间为1小时
		// TODO:配置化
		err = client.Expire(ctx, key, time.Hour).Err()
		if err != nil {
			// 处理错误
			config.Logger.Printf("[Error][discover][reds]Failed to set expiration for key %s: %v\n", key, err)
		}
	}
	elapsedTime := time.Since(startTime) // 计算耗时
	config.Logger.Printf("~~~~~~~~~~~~~~~~~~~~~~WriteSyncRoutes耗时: %v\n", elapsedTime)
}

// NOTE：!!!!!!!!!!超级大优化！！！！！！！！！
//func WriteSyncRoutes(client *redis.Client, routes []*pb.RouteInfo) {
//	ctx := context.Background() // 创建一个上下文
//	startTime := time.Now()     // 记录开始时间
//
//	// 使用管道批量执行 Redis 命令
//	pipe := client.Pipeline()
//
//	for _, route := range routes {
//		cacheInfo := make(map[string]string)
//		cacheInfo["lastSyncVersion"] = time.Now().Format(time.RFC3339)
//		cacheInfo["name"] = route.Name
//		cacheInfo["host"] = route.Host
//		cacheInfo["port"] = route.Port
//		cacheInfo["prefix"] = route.Prefix
//		cacheInfo["metadata"] = route.Metadata
//
//		// 生成 Redis 键
//		var key string
//		if route.Prefix != "" {
//			key = route.Host + "-" + route.Port + ":" + route.Name + ":" + route.Prefix
//		} else {
//			key = route.Host + "-" + route.Port + ":" + route.Name
//		}
//
//		// 使用 HMSet 写入哈希表
//		pipe.HMSet(ctx, key, cacheInfo)
//
//		// 设置过期时间为1小时
//		pipe.Expire(ctx, key, time.Hour)
//	}
//
//	// 执行管道中的命令
//	_, err := pipe.Exec(ctx)
//	if err != nil {
//		config.Logger.Printf("[Error][discover][redis] Failed to execute pipeline: %v\n", err)
//	}
//
//	elapsedTime := time.Since(startTime) // 计算耗时
//	config.Logger.Printf("~~~~~~~~~~~~~~~~~~~~~~WriteSyncRoutes耗时: %v\n", elapsedTime)
//}

// 定期从mysql里面读取新数据,强制刷新!!!!
func LoopRefreshSvrCache(client *redis.Client) {
	// 拿到redis里所有的key，并去mysql里面读取

	// 获取 Redis 中的所有键
	for {
		keys, err := client.Keys(ctx, "*").Result()
		if err != nil {
			config.Logger.Println("[Error][discover][redis]Failed to get keys from Redis: %v\n", err)
			return
		}
		config.Logger.Println("[Info][discover][redis]   LoopRefreshSvrCahce keys:", keys)
		for _, key := range keys {
			// 确定是那种类型的
			parts := strings.Split(key, ":")
			if len(parts) == 2 {
				// TODO: 这里会有问题：一定会更新redis cache里面的版本，变得更新,最好是通过消息队列完成
				ReadFromMysqlWithName(parts[1])
				continue
			} else if len(parts) == 3 {
				// TODO: 这里会有问题：一定会更新redis cache里面的版本，变得更新,最好是通过消息队列完成
				ReadFromMysqlWithName(parts[1]) // Note:这里是为了确保在所有服务都是带有前缀的时候，能正确拉取相同服务名的路由~
				// TODO: 这里会有问题：一定会更新redis cache里面的版本，变得更新
				ReadFromMysqlWithPrefix(parts[1], parts[2])
				continue
			} else {
				config.Logger.Println("[Error][discover][redis] 缓存里面的键有问题：", key)
			}
		}

		time.Sleep(2 * time.Second) // TODO：配置化
	}

}

func SyncRoutesWithRouteSyncRequest(client *redis.Client, req *pb.RouteSyncRequest) []*pb.RouteInfo {
	res := make([]*pb.RouteInfo, 0)
	// name、namePrefix需要比reqTime晚~
	for _, key := range req.Name {
		redisData, exist := readCacheWithName(client, key, req.LastSyncVersion)
		if len(redisData) > 0 || exist {
			config.Logger.Println("[Info][discover][redis][SyncRoutesWithRouteSyncRequest] name从redis同步成功：", len(redisData), key)
			res = append(res, redisData...)
		} else {
			// 从mysql里面读
			mysqlData, err := ReadFromMysqlWithName(key)
			if err != nil {
				config.Logger.Println("[Error][discover][redis][SyncRoutesWithRouteSyncRequest]   redis cache失败ReadFromMysqlWithName，从mysql读出错:", err)
				continue
			}
			if len(mysqlData) == 0 {
				config.Logger.Println("[Warning][discover][redis]   mysql里面路由被删了(name):", key)
				continue
			}
			config.Logger.Println("[Info][discover][redis][SyncRoutesWithRouteSyncRequest] name从mysql同步成功：", len(mysqlData), key)
			res = append(res, mysqlData...)
		}

	}

	for _, key := range req.NamePrefix {
		redisData, exist := readCacheWithNamePrefix(client, key, req.LastSyncVersion)
		if len(redisData) > 0 || exist {
			config.Logger.Println("[Info][discover][redis][SyncRoutesWithRouteSyncRequest] NamePrefix从redis同步成功：", len(redisData), key)
			res = append(res, redisData...)
		} else {
			// 从mysql里面读

			// 提取出name和prefix
			parts := strings.Split(key, ":")
			if len(parts) != 2 {
				config.Logger.Println("[Error][discover][redis][SyncRoutesWithRouteSyncRequest]   redis cache失败,request.NamePrefix格式不对:", key)
				continue
			}
			name, prefix := parts[0], parts[1]
			mysqlData, err := ReadFromMysqlWithPrefix(name, prefix)
			if err != nil {
				config.Logger.Println("[Error][discover][redis][SyncRoutesWithRouteSyncRequest]   redis cache失败ReadFromMysqlWithName，从mysql读出错:", err)
				continue
			}
			if len(mysqlData) == 0 {
				config.Logger.Println("[Warning][discover][redis][SyncRoutesWithRouteSyncRequest]   mysql里面路由被删了(nameprefix):", key)
				continue
			}
			config.Logger.Println("[Info][discover][redis][SyncRoutesWithRouteSyncRequest] NamePrefix，从mysql同步成功：", len(mysqlData), key)
			res = append(res, mysqlData...)
		}

	}

	// nameNew、namePrefixNew必须要读到不管新旧~
	for _, key := range req.NameNew {
		redisData, exist := readCacheWithName(client, key, "")
		if len(redisData) > 0 || exist {
			config.Logger.Println("[Info][discover][redis][SyncRoutesWithRouteSyncRequest]  NameNew从redis同步成功：", len(redisData), key)
			res = append(res, redisData...)
		} else {
			// 从mysql里面读
			mysqlData, err := ReadFromMysqlWithName(key)
			if err != nil {
				config.Logger.Println("[Error][discover][redis][SyncRoutesWithRouteSyncRequest]   redis cache失败ReadFromMysqlWithName，从mysql读出错:", err)
				continue
			}
			if len(mysqlData) == 0 {
				config.Logger.Println("[Warning][discover][redis][SyncRoutesWithRouteSyncRequest]   mysql里面路由被删了(name new):", key)
				continue
			}
			config.Logger.Println("[Info][discover][redis][SyncRoutesWithRouteSyncRequest] NameNew从mysql同步成功：", len(mysqlData), key)
			res = append(res, mysqlData...)
		}

	}

	for _, key := range req.NamePrefixNew {
		redisData, exist := readCacheWithNamePrefix(client, key, "")
		if len(redisData) > 0 || exist {
			config.Logger.Println("[Info][discover][redis][SyncRoutesWithRouteSyncRequest] NamePrefixNew，从redis同步成功：", len(redisData), key)
			res = append(res, redisData...)
		} else {
			// 从mysql里面读

			// 提取出name和prefix
			parts := strings.Split(key, ":")
			if len(parts) != 2 {
				config.Logger.Println("[Error][discover][redis][SyncRoutesWithRouteSyncRequest]   redis cache失败,request.NamePrefix格式不对:", key)
				continue
			}
			name, prefix := parts[0], parts[1]
			mysqlData, err := ReadFromMysqlWithPrefix(name, prefix)
			if err != nil {
				config.Logger.Println("[Error][discover][redis][SyncRoutesWithRouteSyncRequest]   redis cache失败ReadFromMysqlWithName，从mysql读出错:", err)
				continue
			}
			if len(mysqlData) == 0 {
				config.Logger.Println("[Warning][discover][redis][SyncRoutesWithRouteSyncRequest]   mysql里面路由被删了(nameprefix new):", key)
				continue
			}
			config.Logger.Println("[Info][discover][redis][SyncRoutesWithRouteSyncRequest] NamePrefixNew从mysql同步成功：", len(mysqlData), key)
			res = append(res, mysqlData...)
		}

	}

	config.Logger.Println("[Info][discover][redis][SyncRoutesWithRouteSyncRequest]   同步成功,总个数(未去重)：", len(res))
	return res
}

// 从redis缓存里面读取name的cache（所有名字里带name的）, bool表示是否命中（一般命中且为空的话就是说明旧版数据）
func readCacheWithName(client *redis.Client, name string, lastSyncVersion string) ([]*pb.RouteInfo, bool) {
	return readCache(client, "*"+name+"*", lastSyncVersion) // 需要全量匹配，去重由sdk去做吧~
}

// 从redis缓存里面读取name:prefix的cache, bool表示是否命中（一般命中且为空的话就是说明旧版数据）
func readCacheWithNamePrefix(client *redis.Client, namePrefix string, lastSyncVersion string) ([]*pb.RouteInfo, bool) {
	return readCache(client, "*"+namePrefix, lastSyncVersion)
}

// 从redis缓存里面读取带有特定前缀的name的cache
func readCache(client *redis.Client, pattern string, lastSyncVersion string) ([]*pb.RouteInfo, bool) {
	keys, err := client.Keys(ctx, pattern).Result()
	exist := false // TODO:会不会有别的问题？
	if err != nil {
		config.Logger.Printf("[Error][discover][redis] Error fetching keys with pattern %s: %v", pattern, err)
		return nil, false
	}
	var timeLimit time.Time
	if lastSyncVersion != "" { // 有时间的话需要更新一下
		timeLimit, err = time.Parse(time.RFC3339, lastSyncVersion)
		if err != nil {
			config.Logger.Println("[Error][discover][redis]   readCache 转换时间出错", err)
		}
	}

	res := make([]*pb.RouteInfo, 0)
	for _, key := range keys {
		exist = true // cache命中了~
		serviceInfo, err := client.HGetAll(ctx, key).Result()
		if err != nil {
			config.Logger.Printf("[Error][discover][redis] Error fetching redis cache data for key %s: %v\n", key, err)
			continue
		}

		if lastSyncVersion != "" { // 对比时间
			routeTime, err := time.Parse(time.RFC3339, serviceInfo["lastSyncVersion"])
			if err != nil {
				config.Logger.Println("[Error][discover][redis]   readCache lastSyncVersion错误。", err)
				continue
			}
			if routeTime.Before(timeLimit) {
				// 旧版数据不处理
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

	return res, exist
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
