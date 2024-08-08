package database

import (
	"discoversvr/config"
	"discoversvr/tools/convert"
	"os"
	"os/signal"
	"syscall"
)

// 利用队列来快速感知数据库的变化
func ListToMQ() {

	err := config.Consumer.Bind(config.RabbitMQExch)
	if err != nil {
		config.Logger.Println("[Error][discover][mq] Failed to connect to RabbitMQ:", err)
	}
	defer config.Consumer.Close()
	msgs, err := config.Consumer.Consume()
	if err != nil {
		config.Logger.Println("[Error][discover][mq] Failed to subscribe to RabbitMQ exchange: ", err)
	}

	forever := make(chan bool)

	// 从通道中持续接收消息
	go func() {
		for d := range msgs {
			info, err := convert.ByteToRouteInfo(d.Body)
			if err != nil {
				config.Logger.Println("[Error][discover][mq] Failed to convert message to RouteInfo: ", err)
			}
			CheckAndUpdate(config.SyncRedisClient, info)

			// 判断是否需要这条路由信息
			// TODO: 与redis交互
		}
	}()

	// 捕捉系统信号以便优雅地退出
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan

	<-forever
}
