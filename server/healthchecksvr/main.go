package main

import (
	"fmt"
	"google.golang.org/grpc"
	"healthchecksvr/config"
	"healthchecksvr/endpoint"
	"healthchecksvr/plugins"
	pb "healthchecksvr/proto"
	"healthchecksvr/service"
	"healthchecksvr/transport"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"
)

func main() {

	errChan := make(chan error)

	var svc service.Service             // 接口
	svc = &service.HealthCheckService{} // 具体定义
	// add logging middleware
	svc = plugins.LoggingMiddleware(config.KitLogger)(svc)

	// TODO:tracer
	//var (
	//	err           error
	//	hostPort      = config.ServiceHost + ":" + config.ServicePortString
	//	serviceName   = "vote-grpc"
	//	useNoopTracer = (config.ZipkinURL == "")
	//	reporter      = zipkinhttp.NewReporter(config.ZipkinURL)
	//)
	//defer reporter.Close()
	//zEP, _ := zipkin.NewEndpoint(serviceName, hostPort)
	//config.ZipkinTracer, err = zipkin.NewTracer(
	//	reporter, zipkin.WithLocalEndpoint(zEP), zipkin.WithNoopTracer(useNoopTracer),
	//)
	//if err != nil {
	//	config.Logger.Println("err", err)
	//	os.Exit(1)
	//}
	//if !useNoopTracer {
	//	config.Logger.Println("tracer=Zipkin , type=Native,URL=", config.ZipkinURL)
	//}

	// 下面是grpc服务
	// 定义endpoint层

	healthCheckSEnpt := endpoint.MakeHealthCheckSEndpoint(svc)
	//voteToReis = kitzipkin.TraceEndpoint(config.ZipkinTracer, "voteToRedis-endpoint")(voteToReis) // 添加trace信息

	healthCheckCEnpt := endpoint.MakeHealthCheckCEndpoint(svc)
	//voteResult = kitzipkin.TraceEndpoint(config.ZipkinTracer, "voteResult-endpoint")(voteResult)

	endpoints := endpoint.HealthCheckEndpoint{
		HealthCheckS: healthCheckSEnpt,
		HealthCheckC: healthCheckCEnpt,
	}

	// 定义transport层的trace

	// 定义transport层
	grpcServer := transport.NewGRPCServer(endpoints)

	// 连接grpc
	go func() {
		lis, err := net.Listen("tcp", "0.0.0.0:"+config.HealthCheckGrpcPort)
		if err != nil {
			log.Fatalf("failed to listen: %v", err)
		}
		s := grpc.NewServer()
		// s := grpc.NewServer(grpc.UnaryInterceptor(grpctransport.Interceptor)) // 这是 go-kit 提供的一个拦截器，用于将 go-kit 的中间件集成到 gRPC 服务器中。这个拦截器可以在 gRPC 调用的前后执行一些额外的逻辑，比如日志记录、度量、认证等

		pb.RegisterHealthCheckServiceServer(s, grpcServer)
		config.Logger.Println("[Info][healthcheck] Grpc Server start at port:" + config.HealthCheckGrpcPort)
		if err := s.Serve(lis); err != nil {
			log.Fatalf("[Error] healthcheck failed to serve: %v", err)
		}
		config.Logger.Printf("[Info] healthcheck  listening at %v\n", lis.Addr())

	}()

	// TODO:pprof
	// 启动 pprof 服务器
	// http.ListenAndServe("0.0.0.0:6060", nil)
	//mux := http.NewServeMux()
	//mux.HandleFunc("/debug/pprof/", pprof.Index)
	//mux.HandleFunc("/debug/pprof/cmdline", pprof.Cmdline)
	//mux.HandleFunc("/debug/pprof/profile", pprof.Profile)
	//mux.HandleFunc("/debug/pprof/symbol", pprof.Symbol)
	//mux.HandleFunc("/debug/pprof/trace", pprof.Trace)
	//go func() { log.Fatal(http.ListenAndServe(config.PprofPort, mux)) }() // TODO:修改端口信息

	go func() {
		c := make(chan os.Signal, 1)
		signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)
		errChan <- fmt.Errorf("%s", <-c)
	}()

	error := <-errChan

	config.Logger.Println(error)

}
