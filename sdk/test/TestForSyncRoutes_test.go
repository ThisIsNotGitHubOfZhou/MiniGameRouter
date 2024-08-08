package test

import (
	"context"
	"github.com/ThisIsNotGitHubOfZhou/MiniGameRouter/sdk/api"
	discoverpb "github.com/ThisIsNotGitHubOfZhou/MiniGameRouter/sdk/proto/discover"
	"testing"
	"time"
)

// NOTE:测试本地路由能不能按照预期更新
func TestSyncRoutesFunction(t *testing.T) {

	client := api.NewMiniClient("zcf_service", "10.76.143.", "6000", "grpc", "{'flag':true}", 10, 100000)
	defer client.Close()
	ctx := context.Background()
	// 服务注册~~~~~~~~~~~~~~
	//client.RegisterServerInfo = []string{"localhost:20001", "localhost:20002", "localhost:20003", "localhost:20004", "localhost:20005"}
	client.DiscoverServerInfo = []string{"localhost:40001"} //, "localhost:40002", "localhost:40003"

	err := client.InitConfig()
	if err != nil {
		t.Fatal(err)
	}

	t.Log("开始同步 ")
	go client.SyncCache() // 开始同步

	t.Log("设置了一条路由 ")
	time.Sleep(5 * time.Second)
	// 设置一条路由，这时候缓存里应该啥也没有
	err = client.SetRouteRule(ctx, &discoverpb.RouteInfo{
		Name:     "test8_8_1",
		Host:     "localhost",
		Port:     "2222",
		Prefix:   "/test",
		Metadata: `{"flag": true}`, // 确保这是一个有效的 JSON 字符串
	})
	if err != nil {
		t.Fatal(err)
	}

	t.Log("读取路由信息")
	time.Sleep(5 * time.Second)
	// 这时候应该会存到缓存
	//client.DiscoverServiceWithName(ctx, "test")
	routes, err := client.GetRouteInfoWithName(ctx, "test8_8_1")
	if err != nil {
		t.Fatal(err)
	}

	t.Log("[Info][sdk][test] 路由信息:", routes)

	t.Log("应该存到缓存了现在 ")
	time.Sleep(5 * time.Second)

	err = client.SetRouteRule(ctx, &discoverpb.RouteInfo{
		Name:     "test8_8_1",
		Host:     "localhost",
		Port:     "3333",
		Prefix:   "/test",
		Metadata: `{"flag": true}`, // 确保这是一个有效的 JSON 字符串
	})

	t.Log("缓存过一会就会更新应该 ")
	time.Sleep(5 * time.Second)

	time.Sleep(time.Second * 500)

	//client.DeRegister(ctx, client.ID(), client.Name(), client.Host(), client.Port())

}
