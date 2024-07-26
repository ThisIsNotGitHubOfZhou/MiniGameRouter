package api

import "flag"

var (
	RegisteGrpcrHost string
	RegisterGrpcPort string
	IsK8s            bool
)

func init() {
	// 定义命令行标志
	flag.StringVar(&RegisteGrpcrHost, "host", "10.76.143.1", "The host to register grpc")
	flag.StringVar(&RegisterGrpcPort, "port", "20001", "The port to register grpc")
	flag.BoolVar(&IsK8s, "k8s", false, "Is running in Kubernetes")
}

type MiniClient struct {
}

func NewMiniClient() *MiniClient {
	return &MiniClient{}
}
