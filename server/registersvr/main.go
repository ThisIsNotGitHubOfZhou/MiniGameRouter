package main

import (
	"fmt"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"google.golang.org/grpc"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"registersvr/config"
	"registersvr/endpoint"
	"registersvr/plugins"
	pb "registersvr/proto"
	"registersvr/service"
	"registersvr/transport"
	"syscall"
)

func main() {

	errChan := make(chan error)

	var svc service.Service          // 接口
	svc = &service.RegisterService{} // 具体定义
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

	registerEnpt := endpoint.MakeRegisterEndpoint(svc)
	//voteToReis = kitzipkin.TraceEndpoint(config.ZipkinTracer, "voteToRedis-endpoint")(voteToReis) // 添加trace信息

	deregisterEnpt := endpoint.MakeDeRegisterEndpoint(svc)
	//voteResult = kitzipkin.TraceEndpoint(config.ZipkinTracer, "voteResult-endpoint")(voteResult)

	endpoints := endpoint.RegisterEndpoint{
		Register:   registerEnpt,
		DeRegister: deregisterEnpt,
	}

	// 定义transport层的trace

	// 定义transport层
	grpcServer := transport.NewGRPCServer(endpoints)

	// 连接grpc
	go func() {
		lis, err := net.Listen("tcp", "0.0.0.0:"+config.RegisterGrpcPort)
		if err != nil {
			log.Fatalf("failed to listen: %v", err)
		}
		s := grpc.NewServer()
		// s := grpc.NewServer(grpc.UnaryInterceptor(grpctransport.Interceptor)) // 这是 go-kit 提供的一个拦截器，用于将 go-kit 的中间件集成到 gRPC 服务器中。这个拦截器可以在 gRPC 调用的前后执行一些额外的逻辑，比如日志记录、度量、认证等

		pb.RegisterRegisterServiceServer(s, grpcServer)
		config.Logger.Println("[Info][register] Grpc Server start at port:" + config.RegisterGrpcPort)
		if err := s.Serve(lis); err != nil {
			log.Fatalf("failed to serve: %v", err)
		}
		log.Printf("server listening at %v", lis.Addr())

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
	//go func() { log.Fatal(http.ListenAndServe(config.PprofPort, mux)) }()

	if !config.IsK8s {
		go func() {
			http.Handle("/metrics", promhttp.Handler())
			err := http.ListenAndServe(":"+config.PrometheusPort, nil) // 2112 是常用的 Prometheus 端口
			if err != nil {
				config.Logger.Println("[Info][healthcheck] prometheus指标服务器启动失败", err)
			}
		}()
	}

	go func() {
		c := make(chan os.Signal, 1)
		signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)
		errChan <- fmt.Errorf("%s", <-c)
	}()

	error := <-errChan
	config.Logger.Println("[Info][register] redis连接池关闭")
	err := config.RedisClient.Close()
	if err != nil {
		config.Logger.Println("[Info][register] redis连接池关闭出错", err)
	}
	config.Logger.Println(error)

}
