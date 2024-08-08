package test

import (
	"context"
	"github.com/ThisIsNotGitHubOfZhou/MiniGameRouter/sdk/api"
	discoverpb "github.com/ThisIsNotGitHubOfZhou/MiniGameRouter/sdk/proto/discover"
	"testing"
	"time"
)

// NOTE:设置路由，观察是否服务器是否能感知到~
func TestMqFunction(t *testing.T) {

	client := api.NewMiniClient("zcf_service", "10.76.143.", "6000", "grpc", "{'flag':true}", 10, 100000)
	defer client.Close()
	ctx := context.Background()

	client.DiscoverServerInfo = []string{"localhost:40001"} //, "localhost:40002", "localhost:40003"

	err := client.InitConfig()
	if err != nil {
		t.Fatal(err)
	}

	t.Log("设置了一条路由 ")

	err = client.SetRouteRule(ctx, &discoverpb.RouteInfo{
		Name:     "test",
		Host:     "0.0.0.0",
		Port:     "4879",
		Prefix:   "/test",
		Metadata: `{"flag": true}`, // 确保这是一个有效的 JSON 字符串
	})
	if err != nil {
		t.Fatal(err)
	}
	time.Sleep(50 * time.Second)

	err = client.SetRouteRule(ctx, &discoverpb.RouteInfo{
		Name:     "test",
		Host:     "0.0.0.0",
		Port:     "8794",
		Prefix:   "/test",
		Metadata: `{"flag": true}`, // 确保这是一个有效的 JSON 字符串
	})
	if err != nil {
		t.Fatal(err)
	}
	time.Sleep(50 * time.Second)

}
