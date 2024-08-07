package test

import (
	"context"
	"fmt"
	"github.com/ThisIsNotGitHubOfZhou/MiniGameRouter/sdk/api"
	discoverpb "github.com/ThisIsNotGitHubOfZhou/MiniGameRouter/sdk/proto/discover"
	"math/rand"
	"runtime"
	"sync"
	"testing"
	"time"
)

// 定义字符集
const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

// 随机生成指定长度的字符串
func randomString(length int, rng *rand.Rand, mu *sync.Mutex) string { //加锁是防止生成相同字符
	result := make([]byte, length)
	mu.Lock()
	defer mu.Unlock()
	for i := range result {
		// 生成一个随机索引
		index := rng.Intn(len(charset))
		// 从字符集中选择字符
		result[i] = charset[index]
	}
	return string(result)
}

// NOTE 8_2 300s 100w,并行
// NOTE 8_2 30s 10w，并行
// NOTE 8_6 150s 50w,并行
// NOTE:这里只能线性，不线性randomString会出现很多重复
// TODO:线性会出错~

func TestDiscoverFunction(t *testing.T) {

	client := api.NewMiniClient("zcf_service", "10.76.143.", "6000", "grpc", "{'flag':true}", 10, 100000)
	defer client.Close()
	ctx := context.Background()

	client.DiscoverServerInfo = []string{"localhost:40001"} //,, "localhost:40002", "localhost:40003"

	err := client.InitConfig()
	if err != nil {
		t.Fatal(err)
	}

	//var wg sync.WaitGroup
	var mu sync.Mutex
	rng := rand.New(rand.NewSource(time.Now().UnixNano()))
	tempRoute := &discoverpb.RouteInfo{
		Name:     "zcf_service",
		Host:     "0.0.0.0",
		Port:     "666",
		Prefix:   "yesyes",
		Metadata: "{}",
	}

	for i := 0; i < 100000; i++ {
		// 并行~~~~~~~~~~~~~~~~~~~~~(由于并行会导致随机到相同值，所以分片就会退化成未分片)
		//wg.Add(1)
		//go func() {
		//	defer wg.Done()
		//	randStr := randomString(10, rng, &mu)
		//
		//	tempRoute.Name = randStr
		//
		//	randStr = randomString(10, rng, &mu)
		//
		//	tempRoute.Prefix = randStr
		//	err = client.SetRouteRule(ctx, tempRoute)
		//	if err != nil {
		//		t.Errorf("SetRouteRule error : %v", err)
		//	}
		//}()

		// 线性~~~~~~~~~~~~~~~~(线性会导致出bug~)
		randStr := randomString(10, rng, &mu)

		tempRoute.Name = randStr

		randStr = randomString(10, rng, &mu)

		tempRoute.Prefix = randStr
		err = client.SetRouteRule(ctx, tempRoute)
		fmt.Printf("~~~~~~~~~~~Number of goroutines: %d\n", runtime.NumGoroutine())
		if err != nil {
			t.Errorf("SetRouteRule error : %v", err)
			return
		}
	}

	//wg.Wait()

}
