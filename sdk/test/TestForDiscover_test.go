package test

import (
	"context"
	"github.com/ThisIsNotGitHubOfZhou/MiniGameRouter/sdk/api"
	discoverpb "github.com/ThisIsNotGitHubOfZhou/MiniGameRouter/sdk/proto/discover"
	"math/rand"
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

// TODO：
// TODO 8_2 300s 100w,并行
// TODO 8_2 30s 10w，并行
// TODO:数据打到数据库不均匀~
// TODO:这里只能线性，不线性randomString会出现很多重复

func TestDiscoverFunction(t *testing.T) {
	client := api.NewMiniClient("zcf_service", "10.76.143.", "6000", "grpc", "{'flag':true}", 10, 100000)
	ctx := context.Background()

	// 服务注册~~~~~~~~~~~~~~
	//client.RegisterServerInfo = []string{"localhost:20001", "localhost:20002", "localhost:20003", "localhost:20004", "localhost:20005"}
	client.DiscoverServerInfo = []string{"localhost:40001"} //, "localhost:40002", "localhost:40003"

	err := client.InitConfig()
	if err != nil {
		t.Fatal(err)
	}

	//服务发现~~~~~~~~~~~~~~~
	//service_name, err := client.DiscoverServiceWithName(ctx, "zcf_service")
	//if err != nil {
	//	t.Fatal("[Error][test] DiscoverServiceWithName error", err)
	//}
	//for _, val := range service_name {
	//	t.Log("[Info][test] DiscoverServiceWithName val", val)
	//	time.Sleep(1 * time.Second)
	//}
	//~~~~~~~~~~~~~~~~

	// 路由设置~~~~~~~~~~~~~~~~~~
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
		//wg.Add(1)
		//go func() {
		//	defer wg.Done()
		//	//randStr := randomString(5, rng, &mu)
		//	//
		//	//tempRoute.Name = randStr
		//	//
		//	//randStr = randomString(5, rng, &mu)
		//	//
		//	//tempRoute.Prefix = randStr
		//	//err = client.SetRouteRule(ctx, tempRoute)
		//	//if err != nil {
		//	//	t.Errorf("SetRouteRule error : %v", err)
		//	//}
		//}()
		randStr := randomString(10, rng, &mu)

		tempRoute.Name = randStr

		randStr = randomString(10, rng, &mu)

		tempRoute.Prefix = randStr
		err = client.SetRouteRule(ctx, tempRoute)
		if err != nil {
			t.Errorf("SetRouteRule error : %v", err)
		}
	}

	//wg.Wait()

	// 查询路由
	//tempRoute = &discoverpb.RouteInfo{
	//	Name:     "zcf_service",
	//	Host:     "0.0.0.0",
	//	Port:     "666",
	//	Prefix:   "yesyes",
	//	Metadata: "{}",
	//}
	//err = client.SetRouteRule(ctx, tempRoute)
	//if err != nil {
	//	t.Errorf("SetRouteRule error : %v", err)
	//}
	//
	//routeInfo, err := client.GetRouteInfoWithName(ctx, "zcf_service")
	//if err != nil {
	//	t.Errorf("[Error][test] DiscoverServiceWithName error : %v", err)
	//}
	//for _, val := range routeInfo {
	//	t.Log("[Info][test] GetRouteInfoWithName val", val)
	//	time.Sleep(1 * time.Second)
	//}

	//~~~~~~~~~~~~~~~~~~~~~~~

	client.Close()

	//client.DeRegister(ctx, client.ID(), client.Name(), client.Host(), client.Port())

}
