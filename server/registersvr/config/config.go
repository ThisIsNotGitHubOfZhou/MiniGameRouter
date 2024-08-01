package config

import (
	"flag"
	"github.com/go-redis/redis/v8"
	"log"
	"os"
	"time"

	kitlog "github.com/go-kit/log"
)

// 定义两个全局变量，用于存储日志记录器
var Logger *log.Logger
var KitLogger kitlog.Logger

var (
	RedisClient      *redis.Client
	RegisterGrpcHost string
	RegisterGrpcPort string
	IsK8s            bool
)

// init 函数在包初始化时自动执行
func init() {
	// TODO:监控
	//prometheus.MustRegister(RequestsVote)
	//prometheus.MustRegister(RequestsResult)

	// 初始化标准库日志记录器
	Logger = log.New(os.Stderr, "", log.LstdFlags)

	// 初始化 go-kit 日志记录器
	KitLogger = kitlog.NewLogfmtLogger(os.Stderr)
	// 为 go-kit 日志记录器添加时间戳字段
	KitLogger = kitlog.With(KitLogger, "ts", kitlog.DefaultTimestampUTC)
	// 为 go-kit 日志记录器添加调用者信息字段
	KitLogger = kitlog.With(KitLogger, "caller", kitlog.DefaultCaller)

	// 定义命令行标志
	flag.StringVar(&RegisterGrpcHost, "host", "10.76.143.1", "The host to register grpc")
	flag.StringVar(&RegisterGrpcPort, "port", "20001", "The port to register grpc")
	flag.BoolVar(&IsK8s, "k8s", false, "Is running in Kubernetes")

	// 解析命令行标志
	flag.Parse()

	// TODO:配置化
	RedisClient = redis.NewClient(&redis.Options{
		Addr:        "21.6.163.18:6380", // Redis 地址
		Password:    "664597599Zcf!",    // Redis 密码，没有则留空
		DB:          0,                  // 使用的数据库，默认为0
		PoolSize:    3000,               // 连接池大小
		PoolTimeout: 20 * time.Second,   //连接池等待时间
	})

}

// TODO:引入下面的东西
//var (
//	Client       *redis.Client
//	ZipkinTracer *zipkin.Tracer
//)

// promethues指标
//var (
//	RequestsVote = prometheus.NewCounter(
//		prometheus.CounterOpts{
//			Name: "vote_times_" + ServicePortString,
//			Help: "Total number of vote requests.",
//		},
//	)
//	RequestsResult = prometheus.NewCounter(
//		prometheus.CounterOpts{
//			Name: "result_times_" + ServicePortString,
//			Help: "Total number of vote result requests.",
//		},
//	)
//)
