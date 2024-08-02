package main

import (
	"context"
	"fmt"
	"github.com/go-redis/redis/v8"
	"log"
	"time"
)

var ctx = context.Background()

// 注册服务实例
func registerServiceInstance(client *redis.Client, instanceID string, instanceInfo map[string]interface{}, ttl time.Duration) error {
	// 使用HMSet存储服务实例信息
	err := client.HMSet(ctx, instanceID, instanceInfo).Err()
	if err != nil {
		return err
	}
	// 设置过期时间
	return client.Expire(ctx, instanceID, ttl).Err()
}

// 续约服务实例
func renewServiceInstance(client *redis.Client, instanceID string, ttl time.Duration) {
	ticker := time.NewTicker(ttl / 2)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			err := client.Expire(ctx, instanceID, ttl).Err()
			if err != nil {
				fmt.Println("Failed to renew service instance:", err)
			} else {
				fmt.Println("Service instance renewed successfully")
			}
		}
	}
}

// 发现服务实例
func discoverServices(client *redis.Client, pattern string) ([]map[string]string, error) {
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

func flushAll(client *redis.Client) {
	// 使用 SCAN 命令逐步扫描键
	var cursor uint64
	for {
		var keys []string
		var err error
		keys, cursor, err = client.Scan(ctx, cursor, "*", 100).Result()
		if err != nil {
			log.Fatalf("Failed to scan keys: %v", err)
		}

		if len(keys) > 0 {
			// 使用 DEL 命令删除扫描到的键
			if err := client.Del(ctx, keys...).Err(); err != nil {
				log.Fatalf("Failed to delete keys: %v", err)
			}
		}

		// 如果 cursor 为 0，表示扫描结束
		if cursor == 0 {
			break
		}
	}

	log.Println("All keys in the current database have been deleted.")
}

func main() {
	client := redis.NewClient(&redis.Options{
		Addr:     "21.6.163.18:6380", // Redis 地址
		Password: "664597599Zcf!",    // Redis 密码，没有则留空
		DB:       0,                  // 使用的数据库，默认为0
	})

	// instanceID := "instance_1"
	// instanceInfo := map[string]interface{}{
	// 	"service":  "my_service",
	// 	"host":     "127.0.0.1",
	// 	"port":     8080,
	// 	"protocol": "http",
	// 	"weight":   10,
	// 	"healthy":  true,
	// 	"metadata": `{"version": "1.0", "region": "us-east"}`,
	// }
	// ttl := 10 * time.Second

	// // 注册服务实例
	// err := registerServiceInstance(client, instanceID, instanceInfo, ttl)
	// if err != nil {
	// 	fmt.Println("Failed to register service instance:", err)
	// 	return
	// }
	// fmt.Println("Service instance registered successfully")

	// // 启动续约协程
	// go renewServiceInstance(client, instanceID, ttl)

	// // 模拟服务发现
	// time.Sleep(5 * time.Second)
	// services, err := discoverServices(client, "instance_*")
	// if err != nil {
	// 	fmt.Println("Failed to discover services:", err)
	// } else {
	// 	fmt.Println("Discovered services:", services)
	// }

	// 保持主函数运行
	//select {}

	// 删除redis所有键
	flushAll(client)
}
