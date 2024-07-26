package main

import (
	"context"
	"fmt"

	"github.com/ThisIsNotGitHubOfZhou/MiniGameRouter/sdk/api"
)

func main() {
	client := api.NewMiniClient()
	ctx := context.Background()
	str, err := client.Register(ctx, "", "", "", "", "", 1, 2)
	if err != nil {
		fmt.Println("client~~~~~~~~~~", err)
	}
	fmt.Println("client~~~~~~~~~~str", str)
}
