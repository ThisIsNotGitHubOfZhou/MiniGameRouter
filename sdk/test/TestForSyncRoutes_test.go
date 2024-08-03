package test

import (
	"context"
	"github.com/ThisIsNotGitHubOfZhou/MiniGameRouter/sdk/api"
	discoverpb "github.com/ThisIsNotGitHubOfZhou/MiniGameRouter/sdk/proto/discover"
	"testing"
	"time"
)

func TestSyncRoutesFunction(t *testing.T) {
	client := api.NewMiniClient("zcf_service", "10.76.143.", "6000", "grpc", "{'flag':true}", 10, 100000)
	ctx := context.Background()
	// 服务注册~~~~~~~~~~~~~~
	//client.RegisterServerInfo = []string{"localhost:20001", "localhost:20002", "localhost:20003", "localhost:20004", "localhost:20005"}
	client.DiscoverServerInfo = []string{"localhost:40001"} //, "localhost:40002", "localhost:40003"

	err := client.InitConfig()
	if err != nil {
		t.Fatal(err)
	}
	err = client.SetRouteRule(ctx, &discoverpb.RouteInfo{
		Name:     "test",
		Host:     "localhost",
		Port:     "20001",
		Prefix:   "/test",
		Metadata: "{'flag':true}",
	})
	if err != nil {
		t.Fatal(err)
	}

	routes, err := client.DiscoverServiceWithName(ctx, "test")
	if err != nil {
		t.Fatal(err)
	}
	t.Log("[Info][sdk][test] 路由长度:", len(routes))
	client.SyncCache()

	time.Sleep(time.Second * 500)
	client.Close()

	//client.DeRegister(ctx, client.ID(), client.Name(), client.Host(), client.Port())

}
