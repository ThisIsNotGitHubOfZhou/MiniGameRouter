package tools

import (
	pb "discoversvr/proto"
	"time"
)

var (
	SyncRouteUpdates chan *pb.RouteInfo
)

func init() {
	SyncRouteUpdates = make(chan *pb.RouteInfo, 1000)
}

// TODO :删除~~~~~~这个文件

// TODO: 从MySQL同步数据
func StartSyncFromMysql() {
	SyncRouteUpdates = make(chan *pb.RouteInfo, 1000)
	go syncRouteUpdatesFromMysql()

}

func syncRouteUpdatesFromMysql() {
	// 模拟从MySQL同步数据
	for {
		// 这里应该是从MySQL读取数据并发送到通道
		routeInfo := &pb.RouteInfo{
			Name: "test",
		}
		SyncRouteUpdates <- routeInfo
		time.Sleep(2 * time.Second)
	}
}
