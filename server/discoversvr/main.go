package main

import (
	"discoversvr/config"
	"discoversvr/database"
	"discoversvr/endpoint"
	"discoversvr/plugins"
	pb "discoversvr/proto"
	"discoversvr/service"
	"discoversvr/transport"
	"fmt"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"google.golang.org/grpc"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	// 初始化DB
	database.InitMysql()

	errChan := make(chan error)

	var svc service.Service          // 接口
	svc = &service.DiscoverService{} // 具体定义
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

	discoverSvcWithNameEnpt := endpoint.MakeDiscoverServiceWithNameEndpoint(svc)
	//voteToReis = kitzipkin.TraceEndpoint(config.ZipkinTracer, "voteToRedis-endpoint")(voteToReis) // 添加trace信息

	discoverSvcWithIDEnpt := endpoint.MakeDiscoverServiceWithIDEndpoint(svc)
	//voteResult = kitzipkin.TraceEndpoint(config.ZipkinTracer, "voteResult-endpoint")(voteResult)

	getRouteInfoWithNameEnpt := endpoint.MakeGetRouteInfoWithNameEndpoint(svc)
	//voteToReis = kitzipkin.TraceEndpoint(config.ZipkinTracer, "voteToRedis-endpoint")(voteToReis) // 添加trace信息

	getRouteInfoWithPrefixEnpt := endpoint.MakeGetRouteInfoWithPrefixEndpoint(svc)
	//voteResult = kitzipkin.TraceEndpoint(config.ZipkinTracer, "voteResult-endpoint")(voteResult)

	setRouteRuleEnpt := endpoint.MakeSetRouteRuleEndpoint(svc)
	//voteResult = kitzipkin.TraceEndpoint(config.ZipkinTracer, "voteResult-endpoint")(voteResult)

	syncRoutesEnpt := endpoint.MakeSyncRoutesEndpoint(svc)

	updateRouteRule := endpoint.MakeUpdateRouteRuleEndpoint(svc)

	endpoints := endpoint.DiscoverEndpoint{
		DiscoverServiceWithName: discoverSvcWithNameEnpt,
		DiscoverServiceWithID:   discoverSvcWithIDEnpt,
		GetRouteInfoWithName:    getRouteInfoWithNameEnpt,
		GetRouteInfoWithPrefix:  getRouteInfoWithPrefixEnpt,
		SetRouteRule:            setRouteRuleEnpt,
		SyncRoutes:              syncRoutesEnpt,
		UpdateRouteRule:         updateRouteRule,
	}

	// 定义transport层的trace

	// 定义transport层
	grpcServer := transport.NewGRPCServer(endpoints)

	// 连接grpc
	go func() {
		lis, err := net.Listen("tcp", "0.0.0.0:"+config.DiscoverGrpcPort)
		if err != nil {
			log.Fatalf("failed to listen: %v", err)
		}
		s := grpc.NewServer()
		// s := grpc.NewServer(grpc.UnaryInterceptor(grpctransport.Interceptor)) // 这是 go-kit 提供的一个拦截器，用于将 go-kit 的中间件集成到 gRPC 服务器中。这个拦截器可以在 gRPC 调用的前后执行一些额外的逻辑，比如日志记录、度量、认证等

		pb.RegisterDiscoverServiceServer(s, grpcServer)
		config.Logger.Println("[Info][discover] Grpc Server start at port:" + config.DiscoverGrpcPort)
		if err := s.Serve(lis); err != nil {
			log.Fatalf("failed to serve: %v", err)
		}
		log.Printf("[Info][discover] server listening at %v", lis.Addr())

	}()

	// 启动同步线程：定时从mysql里面读取最新信息（从消息队列里读取）
	go database.LoopRefreshSvrCache(config.SyncRedisClient)

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

	// 启动http服务暴露prometheus指标
	// 启动一个 HTTP 服务器来暴露 Prometheus 指标
	if !config.IsK8s {
		go func() {
			http.Handle("/metrics", promhttp.Handler())
			err := http.ListenAndServe(":"+config.PrometheusPort, nil) // 2112 是常用的 Prometheus 端口
			if err != nil {
				config.Logger.Println("[Info][discover] prometheus指标服务器启动失败", err)
			}
		}()
	}

	go func() {
		c := make(chan os.Signal, 1)
		signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)
		errChan <- fmt.Errorf("%s", <-c)
	}()

	error := <-errChan
	config.Logger.Println("[Info][discover] redis连接池关闭")
	err := config.RedisClient.Close()
	if err != nil {
		config.Logger.Println("[Info][discover] redis连接池关闭出错", err)
	}
	config.Logger.Println(error)

}
