package config

import (
	"database/sql"
	"flag"
	"fmt"
	kitlog "github.com/go-kit/log"
	"github.com/go-redis/redis/v8"
	_ "github.com/go-sql-driver/mysql"
	"log"
	"os"
	"time"
)

// 定义两个全局变量，用于存储日志记录器
var Logger *log.Logger
var KitLogger kitlog.Logger

var (
	RedisClient      *redis.Client
	SyncRedisClient  *redis.Client
	MysqlClient      *sql.DB
	DiscoverGrpcHost string
	DiscoverGrpcPort string
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
	flag.StringVar(&DiscoverGrpcHost, "host", "10.76.143.1", "The host to discover grpc")
	flag.StringVar(&DiscoverGrpcPort, "port", "40001", "The port to discover grpc")
	flag.BoolVar(&IsK8s, "k8s", false, "Is running in Kubernetes")

	// 解析命令行标志
	flag.Parse()

	// TODO:参数配置化
	RedisClient = redis.NewClient(&redis.Options{
		Addr:        "21.6.163.18:6380", // Redis 地址
		Password:    "664597599Zcf!",    // Redis 密码，没有则留空
		DB:          0,                  // 使用的数据库，默认为0
		PoolSize:    3000,               // 连接池大小
		PoolTimeout: 20 * time.Second,   //连接池等待时间
	})

	SyncRedisClient = redis.NewClient(&redis.Options{
		Addr:        "21.6.163.18:6380", // Redis 地址
		Password:    "664597599Zcf!",    // Redis 密码，没有则留空
		DB:          1,                  // 使用的数据库，同步数据库
		PoolSize:    3000,               // 连接池大小
		PoolTimeout: 20 * time.Second,   //连接池等待时间
	})

	// mysql
	// TODO:参数配置化
	dsn := fmt.Sprintf("root:664597599Zcf!@tcp(9.134.206.110:3306)/route_db")
	var err error
	MysqlClient, err = sql.Open("mysql", dsn)
	if err != nil {
		Logger.Println("[Error][discover]Failed to connect to database route_db: ", err)
	}

	MysqlClient.SetMaxOpenConns(1000)         // 设置最大打开连接数
	MysqlClient.SetMaxIdleConns(800)          // 设置最大空闲连接数
	MysqlClient.SetConnMaxLifetime(time.Hour) // 设置连接的最大生命周期

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
