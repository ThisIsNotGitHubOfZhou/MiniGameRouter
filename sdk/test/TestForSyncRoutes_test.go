package test

import (
	"github.com/ThisIsNotGitHubOfZhou/MiniGameRouter/sdk/api"
	"testing"
	"time"
)

func TestSyncRoutesFunction(t *testing.T) {
	client := api.NewMiniClient("zcf_service", "10.76.143.", "6000", "grpc", "{'flag':true}", 10, 100000)

	// 服务注册~~~~~~~~~~~~~~
	//client.RegisterServerInfo = []string{"localhost:20001", "localhost:20002", "localhost:20003", "localhost:20004", "localhost:20005"}
	client.DiscoverServerInfo = []string{"localhost:40001"} //, "localhost:40002", "localhost:40003"

	err := client.InitConfig()
	if err != nil {
		t.Fatal(err)
	}
	client.SyncCache()

	time.Sleep(time.Second * 500)
	client.Close()

	//client.DeRegister(ctx, client.ID(), client.Name(), client.Host(), client.Port())

}
