package test

import (
	"context"
	"github.com/ThisIsNotGitHubOfZhou/MiniGameRouter/sdk/api"
	"strconv"
	"sync"
	"testing"
	"time"
)

// NOTE:服务注册测试并发
// NOTE:8.1 21.20s5w
func TestRegisterFunction(t *testing.T) {
	var wg sync.WaitGroup
	client := api.NewMiniClient("zcf_service", "10.76.143.", "6000", "grpc", "{'flag':true}", 10, 100000)
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second) // 设置超时
	defer cancel()

	client.RegisterServerInfo = []string{"localhost:20001", "localhost:20002", "localhost:20003"} //, "localhost:20004", "localhost:20005"

	err := client.InitConfig()
	if err != nil {
		t.Fatalf("Failed to initialize config: %v", err)
	}

	errorChan := make(chan error, 100000) // 用于收集错误

	for i := 0; i < 50000; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			_, err := client.Register(ctx, "zcf_service"+strconv.Itoa(i), "10.76.143."+strconv.Itoa(i), "6000"+strconv.Itoa(i), "", "", 1, 100000)
			if err != nil {
				errorChan <- err
			}
		}(i)
	}

	//time.Sleep(30 * time.Second)
	//for i := 0; i < 50000; i++ {
	//	wg.Add(1)
	//	go func(i int) {
	//		defer wg.Done()
	//		_, err := client.Register(ctx, "zcf_service"+strconv.Itoa(i), "10.76.143."+strconv.Itoa(i), "6000"+strconv.Itoa(i), "", "", 1, 100000)
	//		if err != nil {
	//			errorChan <- err
	//		}
	//	}(i)
	//}
	wg.Wait()
	close(errorChan)

	// 检查是否有错误
	for err := range errorChan {
		if err != nil {
			t.Errorf("Error occurred during registration: %v", err)
		}
	}

	//client.Close()
	t.Log("All goroutines have completed.")
}
