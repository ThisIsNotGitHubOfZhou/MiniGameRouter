package database

import (
	"context"
	"fmt"
	"github.com/go-redis/redis/v8"
	"registersvr/config"
	"time"
)

var ctx = context.Background()

// 注册服务实例
func RegisterServiceInstance(client *redis.Client, instanceID string, instanceInfo map[string]interface{}, ttl time.Duration) error {
	// 使用HMSet存储服务实例信息
	config.Logger.Println("注册实例：", instanceID)
	err := client.HMSet(ctx, instanceID, instanceInfo).Err()
	if err != nil {
		return err
	}
	// 设置过期时间
	return client.Expire(ctx, instanceID, ttl).Err()

}

// 注册服务实例
func DeRegisterServiceInstance(client *redis.Client, instanceID string) error {
	// 使用Del删除服务实例信息
	config.Logger.Println("注销实例：", instanceID)
	return client.Del(ctx, instanceID).Err()
}

// 续约服务实例
func RenewServiceInstance(client *redis.Client, instanceID string, ttl time.Duration) {

	err := client.Expire(ctx, instanceID, ttl).Err()
	if err != nil {
		fmt.Println("Failed to renew service instance:", err)
	} else {
		fmt.Println("Service instance renewed successfully")
	}

	// TODO:可以将下面的东西写到SDK里面定时发送吗？
	//ticker := time.NewTicker(ttl / 2)
	//defer ticker.Stop()
	//for {
	//	select {
	//	case <-ticker.C:
	//
	//	}
	//}
}

// 发现服务实例
func DiscoverServices(client *redis.Client, pattern string) ([]map[string]string, error) {
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
