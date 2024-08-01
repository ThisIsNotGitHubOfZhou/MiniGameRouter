package database

import (
	"context"
	"github.com/go-redis/redis/v8"
	"registersvr/config"
	"time"
)

var ctx = context.Background()

// 注册服务实例
func RegisterServiceInstance(client *redis.Client, instanceID string, instanceInfo map[string]interface{}, ttl time.Duration) error {
	// 使用HMSet存储服务实例信息
	// TODO:下面有用吗？
	//ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	//defer cancel()
	config.Logger.Printf("[Info][register] 注册实例database.RegisterServiceInstance,名称：%v，\n", instanceID)
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
	config.Logger.Printf("[Info][register] 删除实例database.DeRegisterServiceInstance,名称：%v，\n", instanceID)
	return client.Del(ctx, instanceID).Err()
}

// 续约服务实例
func RenewServiceInstance(client *redis.Client, instanceID string, ttl time.Duration) {

	cmd := client.Expire(ctx, instanceID, ttl)
	err := cmd.Err()
	exist := cmd.Val()
	if !exist {
		config.Logger.Println("[Warning] 续约服务的时候键不存在!")
	}

	if err != nil {
		config.Logger.Println("Failed to renew service instance:", err)
	} else {
		config.Logger.Println("Service instance renewed successfully")
	}
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
