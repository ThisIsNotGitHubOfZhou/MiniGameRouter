package test

import (
	"context"
	"github.com/ThisIsNotGitHubOfZhou/MiniGameRouter/sdk/api"
	"sync"
	"testing"
)

// NOTE:并发读取路由
func TestGetRouteFunction(t *testing.T) {

	client := api.NewMiniClient("zcf_service", "10.76.143.", "6000", "grpc", "{'flag':true}", 10, 100000)
	defer client.Close()
	ctx := context.Background()

	client.DiscoverServerInfo = []string{"localhost:40001"}

	err := client.InitConfig()
	if err != nil {
		t.Fatal(err)
	}

	var wg sync.WaitGroup
	for i := 0; i < 10000; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			routes, err := client.GetRouteInfoWithName(ctx, "test")
			if err != nil {
				t.Fatal(err)
			}
			t.Log("get route num:", len(routes))
		}()
	}

	wg.Wait()
}
