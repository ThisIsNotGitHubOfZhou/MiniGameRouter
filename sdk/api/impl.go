package api

import (
	"context"
	"fmt"
	discoverpb "github.com/ThisIsNotGitHubOfZhou/MiniGameRouter/sdk/proto/discover"
	"github.com/ThisIsNotGitHubOfZhou/MiniGameRouter/sdk/service"
	"github.com/ThisIsNotGitHubOfZhou/MiniGameRouter/sdk/tools"
	"sync"
)

type MiniClient struct {
	// 下面是实例信息
	name       string
	id         string
	host       string
	port       string
	protocol   string
	metadata   string
	weight     int
	timeout    int
	healthport string

	// TODO:后续整理:配置化、私有化
	// 下面是服务注册相关的
	registerFlag       int64
	RegisterServerInfo []string // register服务器IP:PORT集合,TODO:后续私有化
	RegisterGRPCPools  []*tools.GRPCPool
	registerLock       sync.Mutex
	registerPoolSize   int

	// 下面是健康检测
	healthCheckFlag       int64
	HealthCheckServerInfo []string // 服务器IP:PORT集合,TODO:后续私有化
	HealthCheckGRPCPools  []*tools.GRPCPool
	healthCheckLock       sync.Mutex
	healthCheckPoolSize   int

	// 下面是服务发现
	discoverFlag       int64
	DiscoverServerInfo []string // 服务器IP:PORT集合,TODO:后续私有化
	DiscoverGRPCPools  []*tools.GRPCPool
	discoverLock       sync.Mutex
	discoverPoolSize   int

	// TODO:缓存
	routeCacheMu   sync.RWMutex
	lastUpdateTime string
	cache          map[string][]*discoverpb.RouteInfo // service到路由信息组的map,TODO:去重和效率之间的取舍
	prefixToIndex  map[string][]int                   // servicename+":"+前缀  ->下标的映射(下标指在同名的service中，路由信息组的下标)
}

var _ service.RegisterService = (*MiniClient)(nil)

var _ service.DiscoverService = (*MiniClient)(nil)

var _ service.HealthCheckService = (*MiniClient)(nil)

var _ service.RouteAlgorithm = (*MiniClient)(nil)

func NewMiniClient(name, host, port, protocol, metadata string, weight, timeout int) *MiniClient {
	return &MiniClient{
		name:                name,
		host:                host,
		port:                port,
		protocol:            protocol,
		metadata:            metadata,
		weight:              weight,
		timeout:             timeout,
		registerFlag:        0,
		registerPoolSize:    500,
		healthCheckFlag:     0,
		healthCheckPoolSize: 500,
		discoverFlag:        0,
		discoverPoolSize:    1000,
		lastUpdateTime:      "",
	}
}

func (c *MiniClient) InitConfig() error { // 初始化配置
	// 初始化Register连接池
	// TODO:注意RegisterServerInfo要被初始化！
	if c.RegisterServerInfo == nil {
		fmt.Println("[Error][sdk] RegisterServerInfo没有初始化~")
		//return fmt.Errorf("RegisterServerInfo not init")
	}
	for i := 0; i < len(c.RegisterServerInfo); i++ {
		tool, err := tools.NewGRPCPool(c.RegisterServerInfo[i], c.registerPoolSize)
		if err != nil {
			fmt.Println("[Error][sdk] NewGRPCPool初始化出错：", err)
			return err
		}
		c.RegisterGRPCPools = append(c.RegisterGRPCPools, tool)
	}

	// 初始化HealthCheck连接池
	// TODO:注意HealthCheckServerInfo要被初始化！
	if c.HealthCheckServerInfo == nil {
		fmt.Println("[Error][sdk] HealthCheckServerInfo没有初始化~")
		//return fmt.Errorf("HealthCheckServerInfo not init")
	}
	for i := 0; i < len(c.HealthCheckServerInfo); i++ {
		tool, err := tools.NewGRPCPool(c.HealthCheckServerInfo[i], c.healthCheckPoolSize)
		if err != nil {
			fmt.Println("[Error][sdk] NewGRPCPool初始化出错：", err)
			return err
		}
		c.HealthCheckGRPCPools = append(c.HealthCheckGRPCPools, tool)
	}

	// 初始化discover连接池
	// TODO:注意DiscoverServerInfo要被初始化！
	if c.DiscoverServerInfo == nil {
		fmt.Println("[Error][sdk] DiscoverServerInfo没有初始化~")
		//return fmt.Errorf("DiscoverServerInfo not init")
	}
	for i := 0; i < len(c.DiscoverServerInfo); i++ {
		tool, err := tools.NewGRPCPool(c.DiscoverServerInfo[i], c.discoverPoolSize)
		if err != nil {
			fmt.Println("[Error][sdk] NewGRPCPool初始化出错：", err)
			return err
		}
		c.DiscoverGRPCPools = append(c.DiscoverGRPCPools, tool)
	}

	return nil
}

func (c *MiniClient) Close() {
	// 退出反注册
	ctx := context.Background()
	c.DeRegister(ctx, c.id, c.name, c.host, c.port)
	for i := 0; i < len(c.RegisterGRPCPools); i++ {
		c.RegisterGRPCPools[i].Close()
	}

	for i := 0; i < len(c.HealthCheckGRPCPools); i++ {
		c.HealthCheckGRPCPools[i].Close()
	}

	for i := 0; i < len(c.DiscoverGRPCPools); i++ {
		c.DiscoverGRPCPools[i].Close()
	}

}

func (c *MiniClient) Name() string {
	return c.name
}

func (c *MiniClient) ID() string {
	return c.id
}

func (c *MiniClient) Host() string {
	return c.host
}

func (c *MiniClient) Port() string {
	return c.port
}

func (c *MiniClient) Protocol() string {
	return c.protocol
}

func (c *MiniClient) Metadata() string {
	return c.metadata
}

func (c *MiniClient) Weight() int {
	return c.weight
}

func (c *MiniClient) Timeout() int {
	return c.timeout
}

func (c *MiniClient) HealthPort() string {
	return c.healthport
}
