package database

import (
	"context"
	"github.com/go-redis/redis/v8"
)

var ctx = context.Background()

// 发现服务实例
func DiscoverServices(client *redis.Client, pattern string) ([]map[string]string, error) {
	pattern = "*" + pattern + "*"
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
