package main

import (
	"context"
	"fmt"
	"github.com/ThisIsNotGitHubOfZhou/MiniGameRouter/sdk/api"
	"strconv"
	"time"
)

func main() {
	client := api.NewMiniClient("zcf_service", "10.76.143.", "6000", "grpc", "{'flag':true}", 10, 100000)
	ctx := context.Background()

	// 服务注册~~~~~~~~~~~~~~
	client.RegisterServerInfo = []string{"localhost:20001", "localhost:20002", "localhost:20003", "localhost:20004", "localhost:20005"}
	//, "localhost:20005"
	//client.Register(ctx, "zbc", client.Host(), client.Port(), client.Protocol(), client.Metadata(), client.Weight(), 10)

	err := client.InitConfig()
	if err != nil {
		fmt.Println(err)
		return
	}
	for i := 0; i < 100000; i++ {
		go client.Register(ctx, "zcf_service"+strconv.Itoa(i), "10.76.143."+strconv.Itoa(i), "6000"+strconv.Itoa(i), "", "", 1, 100000)
	}

	fmt.Println("Hello world ~~~~~~~~~~~~~~~~~~~~~")
	// ~~~~~~~~~~~~~~

	// 服务检测~~~~~~~~~~~~~~~~~~~
	//client.HealthCheckS(ctx, "60001")
	//client.HealthCheckC(ctx, client.ID(), client.Name(), client.Host(), client.Port(), client.Timeout())
	//~~~~~~~~~~~~~~

	// 服务发现~~~~~~~~~~~~~~~
	//service_name, err := client.DiscoverServiceWithName(ctx, "zcf_service")
	//if err != nil {
	//	fmt.Println("[Error][test] DiscoverServiceWithName error", err)
	//}
	//for _, val := range service_name {
	//	fmt.Println("[Error][test] DiscoverServiceWithName val", val)
	//	time.Sleep(1 * time.Second)
	//}
	//~~~~~~~~~~~~~~~~

	//路由发现~~~~~~~~~~~~~~~~~~
	//tempRoute := &discoverpb.RouteInfo{
	//	Name:     "zcf_service",
	//	Host:     "0.0.0.0",
	//	Port:     "666",
	//	Prefix:   "yesyes",
	//	Metadata: "{}",
	//}
	//err := client.SetRouteRule(ctx, tempRoute)
	//if err != nil {
	//	fmt.Println("[Error][test] SetRouteRule error", err)
	//}

	//routeInfo, err := client.GetRouteInfoWithName(ctx, "zcf_service")
	//if err != nil {
	//	fmt.Println("[Error][test] DiscoverServiceWithName error", err)
	//}
	//for _, val := range routeInfo {
	//	fmt.Println("[Error][test] GetRouteInfoWithName val", val)
	//	time.Sleep(1 * time.Second)
	//}

	//~~~~~~~~~~~~~~~~~~~~~~~

	client.Close()
	time.Sleep(500 * time.Second)

	//client.DeRegister(ctx, client.ID(), client.Name(), client.Host(), client.Port())

}
