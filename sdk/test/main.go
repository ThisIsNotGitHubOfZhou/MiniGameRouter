package main

import (
	"context"
	"strconv"
	"time"

	"github.com/ThisIsNotGitHubOfZhou/MiniGameRouter/sdk/api"
)

func main() {
	client := api.NewMiniClient()
	ctx := context.Background()
	client.Register(ctx, "zcf", "123888", ":"+strconv.Itoa(2), "", "", 1, 15)

	time.Sleep(5 * time.Second)

	client.DeRegister(ctx, "id", "zcf", "", "")
	//for i := 0; i < 1000000; i++ {
	//	str, err := client.Register(ctx, strconv.Itoa(i), "123888", ":"+strconv.Itoa(i), "", "", 1, 15)
	//	if err != nil {
	//		fmt.Println("SDK test Register error:", err)
	//	}
	//	fmt.Println("SDK test Register istance：", str)
	//}
}
