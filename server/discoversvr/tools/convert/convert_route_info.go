package convert

import (
	pb "discoversvr/proto"
	"encoding/json"
)

// 中间结构体
type RouteInfoJSON struct {
	Name     string `json:"name,omitempty"`
	Host     string `json:"host,omitempty"`
	Port     string `json:"port,omitempty"`
	Prefix   string `json:"prefix,omitempty"`
	Metadata string `json:"metadata,omitempty"`
}

// 将从rabbitmq里面读取的数据转换为RouteInfo
func ByteToRouteInfo(info []byte) (*pb.RouteInfo, error) {
	var route pb.RouteInfo
	//var routeMid RouteInfoJSON
	err := json.Unmarshal(info, &route)
	if err != nil {
		return nil, err
	}
	//route.Name = routeMid.Name
	//route.Host = routeMid.Host
	//route.Port = routeMid.Port
	//route.Prefix = routeMid.Prefix
	//route.Metadata = routeMid.Metadata
	return &route, nil
}

// 将RouteInfo转成[]byte
func RouteInfoToByte(info *pb.RouteInfo) ([]byte, error) {
	//var routeMid RouteInfoJSON
	//routeMid.Name = info.Name
	//routeMid.Host = info.Host
	//routeMid.Port = info.Port
	//routeMid.Prefix = info.Prefix
	//routeMid.Metadata = info.Metadata
	str, err := json.Marshal(info)
	if err != nil {
		return nil, err
	}
	return str, nil
}
