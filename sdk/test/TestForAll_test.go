package test

import (
	"context"
	"github.com/ThisIsNotGitHubOfZhou/MiniGameRouter/sdk/api"
	discoverpb "github.com/ThisIsNotGitHubOfZhou/MiniGameRouter/sdk/proto/discover"
	"testing"
	"time"
)

// 全流程走一遍
func TestAllFunction(t *testing.T) {

	client := api.NewMiniClient("zcf_service", "localhost.", "6000", "grpc", "{'flag':true}", 10, 15)
	defer client.Close()
	ctx := context.Background()
	// 服务注册~~~~~~~~~~~~~~
	client.RegisterServerInfo = []string{"localhost:20001", "localhost:20002", "localhost:20003"}
	client.HealthCheckServerInfo = []string{"localhost:30001", "localhost:30002", "localhost:30003"}
	client.DiscoverServerInfo = []string{"localhost:40001", "localhost:40002", "localhost:40003"}

	err := client.InitConfig()
	if err != nil {
		t.Fatal(err)
	}

	// 将自己注册上去
	client.Register(ctx, "test", "localhost", "6000", "grpc", `{"flag": true}`, 10, 15)

	// 自己定时想svr发送心跳
	client.HealthCheckC(ctx, client.ID(), client.Name(), client.Host(), client.Port(), client.Timeout())

	t.Log("开始同步 ")
	go client.SyncCache() // 开始同步

	t.Log("设置了一条路由1")
	// 设置一条路由，这时候缓存里应该啥也没有
	err = client.SetRouteRule(ctx, &discoverpb.RouteInfo{
		Name:     "test",
		Host:     "localhost",
		Port:     "2222",
		Prefix:   "/test",
		Metadata: `{"flag": true}`, // 确保这是一个有效的 JSON 字符串
	})
	if err != nil {
		t.Fatal(err)
	}

	t.Log("读取路由信息")
	// 这时候应该会存到缓存

	routes, err := client.GetRouteInfoWithName(ctx, "test")
	if err != nil {
		t.Fatal(err)
	}
	t.Log("[Info][sdk][test] 路由信息(应该只有一条):", len(routes), routes)

	t.Log("设置了一条路由2")
	err = client.SetRouteRule(ctx, &discoverpb.RouteInfo{
		Name:     "test",
		Host:     "localhost",
		Port:     "3333",
		Prefix:   "/test",
		Metadata: `{"flag": true}`, // 确保这是一个有效的 JSON 字符串
	})

	t.Log("设置了一条路由3")
	// 设置同名，但是不同前缀，后续观察是否能通过同步到
	err = client.SetRouteRule(ctx, &discoverpb.RouteInfo{
		Name:     "test",
		Host:     "localhost",
		Port:     "3333",
		Prefix:   "/test123",
		Metadata: `{"flag": true}`, // 确保这是一个有效的 JSON 字符串
	})

	// 设置不同名的路由，这样不会被同步算法自动更新
	err = client.SetRouteRule(ctx, &discoverpb.RouteInfo{
		Name:     "test11",
		Host:     "localhost",
		Port:     "3333",
		Prefix:   "/test",
		Metadata: `{"flag": true}`, // 确保这是一个有效的 JSON 字符串
	})

	err = client.SetRouteRule(ctx, &discoverpb.RouteInfo{
		Name:     "test1133",
		Host:     "localhost",
		Port:     "3333",
		Prefix:   "/test",
		Metadata: `{"flag": true}`, // 确保这是一个有效的 JSON 字符串
	})

	time.Sleep(10 * time.Second) // 等待轮询的所有服务器都更新了自己的redis缓存
	routes, _ = client.GetRouteInfoWithName(ctx, "test")
	t.Log("[Info][sdk][test] 路由信息(应该只有3条):", len(routes), routes)

	routes, _ = client.GetRouteInfoWithPrefix(ctx, "test", "/test123")
	t.Log("[Info][sdk][test] 路由信息(应该只有一条):", len(routes), routes)
	svrInfo, _ := client.DiscoverServiceWithName(ctx, "test")
	t.Log("[Info][sdk][test] 服务信息:", svrInfo)
	routes, _ = client.GetRouteInfoWithName(ctx, "test11") // 执行后会被同步算法自动更新
	t.Log("[Info][sdk][test] 路由信息(应该只有一条):", len(routes), routes)

	// 最后mysql里面应该又5条记录
	// 但是sdk本地内存只有4条
	time.Sleep(time.Second * 500)

	client.DeRegister(ctx, client.ID(), client.Name(), client.Host(), client.Port())

}
