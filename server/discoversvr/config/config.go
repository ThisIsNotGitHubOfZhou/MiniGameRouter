package config

import (
	"database/sql"
	"flag"
	"fmt"
	kitlog "github.com/go-kit/log"
	"github.com/go-redis/redis/v8"
	_ "github.com/go-sql-driver/mysql"
	"github.com/prometheus/client_golang/prometheus"
	"log"
	"os"
	"time"
)

// 定义两个全局变量，用于存储日志记录器
var Logger *log.Logger
var KitLogger kitlog.Logger

var (
	// Redis相关
	RedisClient     *redis.Client
	SyncRedisClient *redis.Client // 用Host-Port:Name:prefix的格式作为Key
	RedisPoolSize   int
	RedisTimeout    int

	// Mysql相关
	MysqlClient      *sql.DB
	MysqlConnNum     int
	MysqlIdleConnNum int
	NameSplitSize    int
	PrefixSplitSize  int

	DiscoverGrpcHost string
	DiscoverGrpcPort string
	IsK8s            bool

	// prometheus指标
	PrometheusPort          string
	ReadRouteTotalTimes     prometheus.Counter
	ReadRouteFromMysqlTimes prometheus.Counter
	ReadRouteFromRedisTimes prometheus.Counter
)

// TODO:引入下面的东西
//var (
//	Client       *redis.Client
//	ZipkinTracer *zipkin.Tracer
//)

// init 函数在包初始化时自动执行
func init() {

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

	// redis相关
	flag.IntVar(&RedisPoolSize, "redisPoolSize", 3000, "The RedisPoolSize to discover grpc")
	flag.IntVar(&RedisTimeout, "redisTimeout", 20, "The RedisTimeout to discover grpc")

	// mysql相关
	flag.IntVar(&MysqlConnNum, "mysqlConnNum", 5000, "The MysqlConnNum to discover grpc")
	flag.IntVar(&MysqlIdleConnNum, "mysqlIdleConnNum", 4500, "The MysqlIdleConnNum to discover grpc")
	flag.IntVar(&NameSplitSize, "nameSplitSize", 2, "The NameSplitSize to discover grpc")
	flag.IntVar(&PrefixSplitSize, "prefixSplitSize", 3, "The PrefixSplitSize to discover grpc")

	// 云服务
	flag.BoolVar(&IsK8s, "k8s", false, "Is running in Kubernetes")

	// prometheus相关
	flag.StringVar(&PrometheusPort, "prometheusPort", "42112", "The port to prometheus")

	// 解析命令行标志
	flag.Parse()

	RedisClient = redis.NewClient(&redis.Options{
		Addr:        "21.6.163.18:6380",                        // Redis 地址
		Password:    "664597599Zcf!",                           // Redis 密码，没有则留空
		DB:          0,                                         // 使用的数据库，默认为0
		PoolSize:    RedisPoolSize,                             // 连接池大小
		PoolTimeout: time.Duration(RedisTimeout) * time.Second, //连接池等待时间
	})

	SyncRedisClient = redis.NewClient(&redis.Options{
		Addr:        "21.6.163.18:6380",                        // Redis 地址
		Password:    "664597599Zcf!",                           // Redis 密码，没有则留空
		DB:          1,                                         // 使用的数据库，同步数据库
		PoolSize:    RedisPoolSize,                             // 连接池大小
		PoolTimeout: time.Duration(RedisTimeout) * time.Second, //连接池等待时间
	})

	// mysql
	dsn := fmt.Sprintf("root:664597599Zcf!@tcp(9.134.206.110:3306)/route_db")
	var err error
	MysqlClient, err = sql.Open("mysql", dsn)
	if err != nil {
		Logger.Println("[Error][discover]Failed to connect to database route_db: ", err)
	}

	MysqlClient.SetMaxOpenConns(MysqlConnNum)        // 设置最大打开连接数
	MysqlClient.SetMaxIdleConns(MysqlIdleConnNum)    // 设置最大空闲连接数
	MysqlClient.SetConnMaxLifetime(30 * time.Minute) // 设置连接的最大生命周期

	// prometheus监控
	if !IsK8s {
		ReadRouteFromMysqlTimes = prometheus.NewCounter(
			prometheus.CounterOpts{
				Name: "read_from_mysql_times_" + DiscoverGrpcPort,
				Help: "Total times of read from mysql.",
			},
		)
		ReadRouteFromRedisTimes = prometheus.NewCounter(
			prometheus.CounterOpts{
				Name: "read_from_redis_times_" + DiscoverGrpcPort,
				Help: "Total times of read from redis.",
			},
		)
		ReadRouteTotalTimes = prometheus.NewCounter(
			prometheus.CounterOpts{
				Name: "read_total_times_" + DiscoverGrpcPort,
				Help: "Total times of read.",
			},
		)

		prometheus.MustRegister(ReadRouteFromMysqlTimes)
		prometheus.MustRegister(ReadRouteFromRedisTimes)
		prometheus.MustRegister(ReadRouteTotalTimes)
	}

}
