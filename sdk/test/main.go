package main

import (
	"context"
	"time"

	"github.com/ThisIsNotGitHubOfZhou/MiniGameRouter/sdk/api"
)

func main() {
	client := api.NewMiniClient("zcf_test", "10.76.143.36", "60001", "grpc", "{'flag':true}", 10, 10)
	ctx := context.Background()
	client.Register(ctx, client.Name(), client.Host(), client.Port(), client.Protocol(), client.Metadata(), client.Weight(), client.Timeout())

	//client.HealthCheckS(ctx, "60001")
	client.HealthCheckC(ctx, client.ID(), client.Name(), client.Host(), client.Port(), client.Timeout())

	time.Sleep(500 * time.Second)

	client.DeRegister(ctx, client.ID(), client.Name(), client.Host(), client.Port())
	//for i := 0; i < 1000000; i++ {
	//	str, err := client.Register(ctx, strconv.Itoa(i), "123888", ":"+strconv.Itoa(i), "", "", 1, 15)
	//	if err != nil {
	//		fmt.Println("SDK test Register error:", err)
	//	}
	//	fmt.Println("SDK test Register istance：", str)
	//}
}
