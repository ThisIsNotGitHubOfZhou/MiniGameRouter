package main

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/ThisIsNotGitHubOfZhou/MiniGameRouter/sdk/api"
)

func main() {
	client := api.NewMiniClient("zcf_service", "10.76.143.", "6000", "grpc", "{'flag':true}", 10, 100000)
	ctx := context.Background()
	client.Register(ctx, client.Name(), client.Host(), client.Port(), client.Protocol(), client.Metadata(), client.Weight(), client.Timeout())

	//client.HealthCheckS(ctx, "60001")
	//client.HealthCheckC(ctx, client.ID(), client.Name(), client.Host(), client.Port(), client.Timeout())
	for i := 0; i < 100; i++ {
		str, err := client.Register(ctx, "zcf_service"+strconv.Itoa(i), "10.76.143."+strconv.Itoa(i), "6000"+strconv.Itoa(i), "", "", 1, 100000)
		if err != nil {
			fmt.Println("[Error][test]SDK test Register error:", err)
		}
		fmt.Println("[Info][test]SDK test Register istance：", str)
	}
	service_name, err := client.DiscoverServiceWithName(ctx, "zcf_service")
	if err != nil {
		fmt.Println("[Error][test] DiscoverServiceWithName error", err)
	}
	for _, val := range service_name {
		fmt.Println("[Error][test] DiscoverServiceWithName val", val)
		time.Sleep(1 * time.Second)
	}
	time.Sleep(500 * time.Second)

	//client.DeRegister(ctx, client.ID(), client.Name(), client.Host(), client.Port())

}
