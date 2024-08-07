package test

import (
	"context"
	"github.com/ThisIsNotGitHubOfZhou/MiniGameRouter/sdk/api"
	"testing"
	"time"
)

// NOTE:测试万个实例同时去请求同步内存~
// NOTE：GRPCPool连接数无法支持上万,100个完全可以,1000个会报错
func TestMutiSyncRoutesFunction(t *testing.T) {

	for i := 0; i < 100; i++ {
		go func() {
			client := api.NewMiniClient("zcf_service", "10.76.143.", "6000", "grpc", "{'flag':true}", 10, 100000)
			defer client.Close()
			ctx := context.Background()

			client.DiscoverServerInfo = []string{"localhost:40001"} //, "localhost:40002", "localhost:40003"

			err := client.InitConfig()
			if err != nil {
				t.Fatal(err)
			}

			go client.SyncCache() // 开始同步

			_, err = client.GetRouteInfoWithName(ctx, "test")
			if err != nil {
				t.Fatal(err)
			}

			time.Sleep(time.Second * 500)
		}()
	}
	time.Sleep(time.Second * 10000)
}
