package api

import (
	"flag"
)

var (
	RegisteGrpcrHost    string
	RegisterGrpcPort    string
	HealthCheckGrpcHost string
	HealthCheckGrpcPort string
	IsK8s               bool
)

func init() {
	// 定义命令行标志
	flag.StringVar(&RegisteGrpcrHost, "registerhost", "9.135.95.71", "The host to register grpc")
	flag.StringVar(&RegisterGrpcPort, "registerport", "20001", "The port to register grpc")
	flag.StringVar(&HealthCheckGrpcHost, "healthcheckhost", "9.135.95.71", "The host to register grpc")
	flag.StringVar(&HealthCheckGrpcPort, "healthcheckport", "30001", "The port to register grpc")
	flag.BoolVar(&IsK8s, "k8s", false, "Is running in Kubernetes")
}
