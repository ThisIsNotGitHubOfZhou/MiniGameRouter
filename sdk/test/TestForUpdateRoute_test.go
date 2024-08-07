package test

import (
	"context"
	"github.com/ThisIsNotGitHubOfZhou/MiniGameRouter/sdk/api"
	discoverpb "github.com/ThisIsNotGitHubOfZhou/MiniGameRouter/sdk/proto/discover"
	"testing"
	"time"
)

func TestUpdateRouteFunction(t *testing.T) {
	client := api.NewMiniClient("zcf_service", "10.76.143.", "6000", "grpc", "{'flag':true}", 10, 100000)
	ctx := context.Background()

	client.DiscoverServerInfo = []string{"localhost:40001"} //,, "localhost:40002", "localhost:40003"

	err := client.InitConfig()
	if err != nil {
		t.Fatal(err)
	}

	tempRoute := &discoverpb.RouteInfo{
		Name:     "000zcf_service",
		Host:     "0.0.0.0",
		Port:     "666",
		Prefix:   "yesyes",
		Metadata: "{}",
	}

	err = client.SetRouteRule(ctx, tempRoute)

	time.Sleep(time.Second * 50)
	if err != nil {
		t.Errorf("SetRouteRule error : %v", err)
	}
	tempRoute.Host = "1.1.2.3"
	tempRoute.Port = "123"
	tempRoute.Metadata = `{"flag": true}`
	err = client.UpdateRouteRule(ctx, "000zcf_service", "1.1.2.3", "123", "yesyes", tempRoute)

	if err != nil {
		t.Fatal(err)
	}
	//wg.Wait()

	client.Close()

}
