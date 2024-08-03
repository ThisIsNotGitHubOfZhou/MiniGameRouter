package api

import (
	"context"
	"fmt"
	discoverpb "github.com/ThisIsNotGitHubOfZhou/MiniGameRouter/sdk/proto/discover"
	grpctransport "github.com/go-kit/kit/transport/grpc"
	"io"
	"log"
	"time"
)

// TODO：后面别使用轮询了，建议使用一致性哈希~
// 根据服务名发现
func (c *MiniClient) DiscoverServiceWithName(ctx context.Context, name string) ([]*discoverpb.ServiceInfo, error) {
	fmt.Println("[Info][sdk] DiscoverServiceWithName，发现服务:", name)
	// 轮询服务
	c.discoverLock.Lock()
	c.discoverFlag++
	tempFlag := c.discoverFlag
	c.discoverLock.Unlock()

	if len(c.DiscoverGRPCPools) == 0 {
		fmt.Println("[Error][sdk] DiscoverGRPCPools为空")
		return nil, fmt.Errorf("DiscoverGRPCPools empty")
	}
	conn, err := c.DiscoverGRPCPools[tempFlag%(int64(len(c.DiscoverGRPCPools)))].Get() // 优化后
	defer c.DiscoverGRPCPools[tempFlag%(int64(len(c.DiscoverGRPCPools)))].Put(conn)
	if err != nil {
		return nil, err
	}
	//defer conn.Close()

	//// 使用 go-kit 的 gRPC 客户端传输层~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~
	var ep = grpctransport.NewClient(
		conn,
		"discover.DiscoverService", // 服务名称,注意前面要带包名！！！！！包名+在proto文件里定义的服务名
		"DiscoverServiceWithName",  // 方法名称
		encodeGRPCDiscoverServiceWithNameRequest,
		decodeGRPCDiscoverServiceWithNameResponse,
		discoverpb.DiscoverServiceWithNameRequest{},
		//clientTracer,
	).Endpoint()

	// 使用端点进行调用grpc
	request := &discoverpb.DiscoverServiceWithNameRequest{
		Name: name,
	}
	response, err := ep(ctx, request)
	if err != nil {
		fmt.Println("[Error][sdk][discover] grpc error", err)
		return nil, err
	}
	r := response.(*discoverpb.DiscoverServiceResponse)

	fmt.Println("[Info][sdk]  DiscoverServiceWithName 结果", r)
	if r.ErrorMes != "" {
		fmt.Println("[Info][sdk]  DiscoverServiceWithName error", r.ErrorMes)
	}
	return r.Services, nil
}
func encodeGRPCDiscoverServiceWithNameRequest(_ context.Context, request interface{}) (interface{}, error) {
	req := request.(*discoverpb.DiscoverServiceWithNameRequest)
	return req, nil
}
func decodeGRPCDiscoverServiceWithNameResponse(_ context.Context, response interface{}) (interface{}, error) {
	resp := response.(*discoverpb.DiscoverServiceResponse)
	return resp, nil
}

// 根据服务InstanceID返回
func (c *MiniClient) DiscoverServiceWithID(ctx context.Context, instanceID string) ([]*discoverpb.ServiceInfo, error) {
	fmt.Println("[Info][sdk] DiscoverServiceWithID，发现服务:", instanceID)
	// 轮询服务
	c.discoverLock.Lock()
	c.discoverFlag++
	tempFlag := c.discoverFlag
	c.discoverLock.Unlock()

	if len(c.DiscoverGRPCPools) == 0 {
		fmt.Println("[Error][sdk] DiscoverGRPCPools为空")
		return nil, fmt.Errorf("DiscoverGRPCPools empty")
	}
	conn, err := c.DiscoverGRPCPools[tempFlag%(int64(len(c.DiscoverGRPCPools)))].Get() // 优化后
	defer c.DiscoverGRPCPools[tempFlag%(int64(len(c.DiscoverGRPCPools)))].Put(conn)
	if err != nil {
		return nil, err
	}

	//// 使用 go-kit 的 gRPC 客户端传输层~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~
	var ep = grpctransport.NewClient(
		conn,
		"discover.DiscoverService", // 服务名称,注意前面要带包名！！！！！包名+在proto文件里定义的服务名
		"DiscoverServiceWithID",    // 方法名称
		encodeGRPCDiscoverServiceWithIDRequest,
		decodeGRPCDiscoverServiceWithIDResponse,
		discoverpb.DiscoverServiceResponse{},
		//clientTracer,
	).Endpoint()

	// 使用端点进行调用grpc
	request := &discoverpb.DiscoverServiceWithIDRequest{
		InstanceId: instanceID,
	}
	response, err := ep(ctx, request)
	if err != nil {
		fmt.Println("[Error][sdk] DiscoverServiceWithID grpc error", err)
		return nil, err
	}
	r := response.(*discoverpb.DiscoverServiceResponse)

	fmt.Println("[Info][sdk]  DiscoverServiceWithID 结果", r)
	if r.ErrorMes != "" {
		fmt.Println("[Info][sdk]  DiscoverServiceWithID error", r.ErrorMes)
	}
	return r.Services, nil
}

func encodeGRPCDiscoverServiceWithIDRequest(_ context.Context, request interface{}) (interface{}, error) {
	req := request.(*discoverpb.DiscoverServiceWithIDRequest)
	return req, nil
}
func decodeGRPCDiscoverServiceWithIDResponse(_ context.Context, response interface{}) (interface{}, error) {
	resp := response.(*discoverpb.DiscoverServiceResponse)
	return resp, nil
}

func (c *MiniClient) getRouteWithNameFromCache(name string) []*discoverpb.RouteInfo {
	// TODO:先用cache~
	c.routeCacheMu.RLock()
	defer c.routeCacheMu.RUnlock()
	if c.cache != nil {
		routes, ok := c.cache[name]
		if ok {
			return routes
		}
	}
	return nil
}

func (c *MiniClient) putRouteWithNameToCache(name string, routes []*discoverpb.RouteInfo) {
	// TODO:存入cache
	c.routeCacheMu.Lock()
	defer c.routeCacheMu.Unlock()
	if c.cache == nil {
		c.cache = make(map[string][]*discoverpb.RouteInfo)
	}
	// 这里是不是应该增量？
	c.cache[name] = append(c.cache[name], routes...)
	for i, route := range routes {
		if name != route.Name {
			fmt.Println("[Error][sdk] 服务名、前缀不一致~")
			continue
		}
		if route.Prefix != "" {
			c.prefixToIndex[route.Name+":"+route.Prefix] = append(c.prefixToIndex[route.Name], i) // 可能有前缀相同的多个服务
		}
	}

}

// 根据服务名返回路由
func (c *MiniClient) GetRouteInfoWithName(ctx context.Context, name string) ([]*discoverpb.RouteInfo, error) {
	// 先从缓存中获取
	cacheRes := c.getRouteWithNameFromCache(name)
	if cacheRes != nil {
		fmt.Println("[Info][sdk] GetRouteInfoWithName，cache 命中获取路由:", name)
		return cacheRes, nil
	}

	fmt.Println("[Info][sdk] GetRouteInfoWithName，获取路由:", name)
	// 轮询服务
	c.discoverLock.Lock()
	c.discoverFlag++
	tempFlag := c.discoverFlag
	c.discoverLock.Unlock()

	if len(c.DiscoverGRPCPools) == 0 {
		fmt.Println("[Error][sdk] DiscoverGRPCPools为空")
		return nil, fmt.Errorf("DiscoverGRPCPools empty")
	}
	conn, err := c.DiscoverGRPCPools[tempFlag%(int64(len(c.DiscoverGRPCPools)))].Get() // 优化后
	defer c.DiscoverGRPCPools[tempFlag%(int64(len(c.DiscoverGRPCPools)))].Put(conn)
	if err != nil {
		return nil, err
	}

	//clientTracer := kitzipkin.GRPCClientTrace(config.ZipkinTracer)

	//// 使用 go-kit 的 gRPC 客户端传输层~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~
	var ep = grpctransport.NewClient(
		conn,
		"discover.DiscoverService", // 服务名称,注意前面要带包名！！！！！包名+在proto文件里定义的服务名
		"GetRouteInfoWithName",     // 方法名称
		encodeGRPCGetRouteInfoWithNameRequest,
		decodeGRPCGetRouteInfoWithNameResponse,
		discoverpb.RouteInfosResponse{},
		//clientTracer,
	).Endpoint()

	// 使用端点进行调用grpc
	request := &discoverpb.GetRouteInfoWithNameRequest{
		Name: name,
	}
	response, err := ep(ctx, request)
	if err != nil {
		fmt.Println("[Error][sdk] GetRouteInfoWithName grpc error", err)
		return nil, err
	}
	r := response.(*discoverpb.RouteInfosResponse)

	fmt.Println("[Info][sdk]  GetRouteInfoWithName 结果", r)
	if r.ErrorMes != "" {
		fmt.Println("[Info][sdk]  GetRouteInfoWithName error", r.ErrorMes)
	}

	// 存入cache
	c.putRouteWithNameToCache(name, r.Routes)
	return r.Routes, nil
}

func encodeGRPCGetRouteInfoWithNameRequest(_ context.Context, request interface{}) (interface{}, error) {
	req := request.(*discoverpb.GetRouteInfoWithNameRequest)
	return req, nil
}
func decodeGRPCGetRouteInfoWithNameResponse(_ context.Context, response interface{}) (interface{}, error) {
	resp := response.(*discoverpb.RouteInfosResponse)
	return resp, nil
}

func (c *MiniClient) getRouteWithPrefixFromCache(name string, prefix string) []*discoverpb.RouteInfo {
	c.routeCacheMu.RLock()
	defer c.routeCacheMu.RUnlock()
	if c.cache != nil && c.prefixToIndex != nil {
		routesWithName, ok1 := c.cache[name]
		indexs, ok2 := c.prefixToIndex[name+":"+prefix]
		if ok1 && ok2 {
			var res []*discoverpb.RouteInfo
			for _, i := range indexs {
				res = append(res, routesWithName[i])
			}
			return res
		}
	}
	return nil
}

func (c *MiniClient) putRouteWithPrefixToCache(name string, prefix string, routes []*discoverpb.RouteInfo) {
	c.routeCacheMu.Lock()
	defer c.routeCacheMu.Unlock()
	if c.cache == nil {
		c.cache = make(map[string][]*discoverpb.RouteInfo)
	}
	c.cache[name] = append(c.cache[name], routes...) // 会不会有冗余数据：TODO,去冗余
	for i, route := range routes {
		if route.Prefix != "" {
			if route.Prefix != prefix || name != route.Name {
				fmt.Println("[Error][sdk] 服务名、前缀不一致~")
				continue
			}
			// TODO:下标更新是错误的~
			c.prefixToIndex[route.Name+":"+route.Prefix] = append(c.prefixToIndex[route.Name+":"+route.Prefix], i) // 可能有前缀相同的多个服务
		}
	}
}

// 根据服务名+前缀返回路由
func (c *MiniClient) GetRouteInfoWithPrefix(ctx context.Context, name string, prefix string) ([]*discoverpb.RouteInfo, error) {
	// 先从缓存中获取
	cacheRes := c.getRouteWithPrefixFromCache(name, prefix)
	if cacheRes != nil {
		fmt.Println("[Info][sdk] GetRouteInfoWithPrefix，cache 命中获取路由:", name)
		return cacheRes, nil
	}

	fmt.Println("[Info][sdk] GetRouteInfoWithPrefix，发现路由:", name)
	// 轮询服务
	c.discoverLock.Lock()
	c.discoverFlag++
	tempFlag := c.discoverFlag
	c.discoverLock.Unlock()

	if len(c.DiscoverGRPCPools) == 0 {
		fmt.Println("[Error][sdk] DiscoverGRPCPools为空")
		return nil, fmt.Errorf("DiscoverGRPCPools empty")
	}
	conn, err := c.DiscoverGRPCPools[tempFlag%(int64(len(c.DiscoverGRPCPools)))].Get() // 优化后
	defer c.DiscoverGRPCPools[tempFlag%(int64(len(c.DiscoverGRPCPools)))].Put(conn)
	if err != nil {
		return nil, err
	}

	//clientTracer := kitzipkin.GRPCClientTrace(config.ZipkinTracer)

	//// 使用 go-kit 的 gRPC 客户端传输层~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~
	var ep = grpctransport.NewClient(
		conn,
		"discover.DiscoverService", // 服务名称,注意前面要带包名！！！！！包名+在proto文件里定义的服务名
		"GetRouteInfoWithPrefix",   // 方法名称
		encodeGRPCGetRouteInfoWithPrefixRequest,
		decodeGRPCGetRouteInfoWithPrefixResponse,
		discoverpb.RouteInfosResponse{},
		//clientTracer,
	).Endpoint()

	// 使用端点进行调用grpc
	request := &discoverpb.GetRouteInfoWithPrefixRequest{
		Name:   name,
		Prefix: prefix,
	}
	response, err := ep(ctx, request)
	if err != nil {
		fmt.Println("[Error][sdk] GetRouteInfoWithPrefix grpc error", err)
		return nil, err
	}
	r := response.(*discoverpb.RouteInfosResponse)

	fmt.Println("[Info][sdk]  GetRouteInfoWithPrefix 结果", r)
	if r.ErrorMes != "" {
		fmt.Println("[Info][sdk]  GetRouteInfoWithPrefix error", r.ErrorMes)
	}
	c.putRouteWithPrefixToCache(name, prefix, r.Routes) // 将prefix路由写入缓存
	return r.Routes, nil
}

func encodeGRPCGetRouteInfoWithPrefixRequest(_ context.Context, request interface{}) (interface{}, error) {
	req := request.(*discoverpb.GetRouteInfoWithPrefixRequest)
	return req, nil
}
func decodeGRPCGetRouteInfoWithPrefixResponse(_ context.Context, response interface{}) (interface{}, error) {
	resp := response.(*discoverpb.RouteInfosResponse)
	return resp, nil
}

// 前缀路由(prefix)or定向路由(metadata)
func (c *MiniClient) SetRouteRule(ctx context.Context, info *discoverpb.RouteInfo) error {
	fmt.Println("[Info][sdk] SetRouteRule，設置路由:", info)
	// 轮询服务
	c.discoverLock.Lock()
	c.discoverFlag++
	tempFlag := c.discoverFlag
	c.discoverLock.Unlock()

	if len(c.DiscoverGRPCPools) == 0 {
		fmt.Println("[Error][sdk] DiscoverGRPCPools为空")
		return fmt.Errorf("DiscoverGRPCPools empty")
	}
	conn, err := c.DiscoverGRPCPools[tempFlag%(int64(len(c.DiscoverGRPCPools)))].Get() // 优化后
	defer c.DiscoverGRPCPools[tempFlag%(int64(len(c.DiscoverGRPCPools)))].Put(conn)
	if err != nil {
		return err
	}

	//clientTracer := kitzipkin.GRPCClientTrace(config.ZipkinTracer)

	//// 使用 go-kit 的 gRPC 客户端传输层~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~
	var ep = grpctransport.NewClient(
		conn,
		"discover.DiscoverService", // 服务名称,注意前面要带包名！！！！！包名+在proto文件里定义的服务名
		"SetRouteRule",             // 方法名称
		encodeGRPCSetRouteRuleRequest,
		decodeGRPCSetRouteRuleResponse,
		discoverpb.SetRouteRuleResponse{},
		//clientTracer,
	).Endpoint()

	// 使用端点进行调用grpc
	request := &discoverpb.RouteInfo{
		Name:     info.Name,
		Host:     info.Host,
		Port:     info.Port,
		Prefix:   info.Prefix,
		Metadata: info.Metadata,
	}
	response, err := ep(ctx, request)
	if err != nil {
		fmt.Println("[Error][sdk] SetRouteRule grpc error", err)
		return err
	}
	r := response.(*discoverpb.SetRouteRuleResponse)

	if r.ErrorMes != "" {
		fmt.Println("[Info][sdk]  SetRouteRule error", r.ErrorMes)
		return fmt.Errorf(r.ErrorMes)
	}
	return nil
}

func encodeGRPCSetRouteRuleRequest(_ context.Context, request interface{}) (interface{}, error) {
	req := request.(*discoverpb.RouteInfo)
	return req, nil
}
func decodeGRPCSetRouteRuleResponse(_ context.Context, response interface{}) (interface{}, error) {
	resp := response.(*discoverpb.SetRouteRuleResponse)
	return resp, nil
}

func (c *MiniClient) getRouteSyncRequest() *discoverpb.RouteSyncRequest {
	c.routeCacheMu.RLock()
	defer c.routeCacheMu.RUnlock()
	return &discoverpb.RouteSyncRequest{
		Name:            getCacheStringKeys(c.cache),
		NamePrefix:      getPrefixToIndexKeys(c.prefixToIndex),
		LastSyncVersion: c.lastUpdateTime,
	}
}

func getCacheStringKeys(m map[string][]*discoverpb.RouteInfo) []string {
	result := make([]string, len(m))
	for k, _ := range m {
		result = append(result, k)
	}
	return result
}
func getPrefixToIndexKeys(m map[string][]int) []string {
	result := make([]string, len(m))
	for k, _ := range m {
		result = append(result, k)
	}
	return result
}

func (c *MiniClient) syncRoute(resp *discoverpb.RouteSyncResponse) {
	c.routeCacheMu.RLock()
	defer c.routeCacheMu.RUnlock()
	c.lastUpdateTime = resp.NewVersion
	// TODO:去重,同步cache

}

// cache同步线程
func (c *MiniClient) SyncCache() error {
	// TODO: 同步cache
	// 利用stream流实现
	// TODO:服务如果有新的要访问的路由数据如何只增量读？如何设计？

	fmt.Println("[Info][sdk] SyncCache，开始:")
	// 轮询服务
	c.discoverLock.Lock()
	c.discoverFlag++
	tempFlag := c.discoverFlag
	c.discoverLock.Unlock()

	if len(c.DiscoverGRPCPools) == 0 {
		fmt.Println("[Error][sdk] DiscoverGRPCPools为空")
		return fmt.Errorf("DiscoverGRPCPools empty")
	}
	conn, err := c.DiscoverGRPCPools[tempFlag%(int64(len(c.DiscoverGRPCPools)))].Get()
	defer c.DiscoverGRPCPools[tempFlag%(int64(len(c.DiscoverGRPCPools)))].Put(conn)
	if err != nil {
		return err
	}

	client := discoverpb.NewDiscoverServiceClient(conn)
	// ~~~~~~~~~~~~~~~~~上面是固定操作，下面才是真正同步逻辑~~~~~~~~~~~~~~~~~~

	// 创建双向流
	stream, err := client.SyncRoutes(context.Background())
	if err != nil {
		log.Fatalf("SyncRoutes could not create stream: %v", err)
	}

	// 启动一个 goroutine 来发送请求
	go func() {
		for {
			req := c.getRouteSyncRequest() // TODO:新增路由请求该怎么办？
			if err := stream.Send(req); err != nil {
				if err == io.EOF {
					return
				}
				log.Fatalf("failed to send request: %v", err)
			}
			time.Sleep(2 * time.Second) // 模拟定期发送请求
		}
	}()

	// 处理返回的流
	for {
		routeInfo, err := stream.Recv()
		if err == io.EOF {
			// 流结束
			break
		}
		if err != nil {
			log.Fatalf("error receiving stream: %v", err)
		}

		// 处理接收到的 RouteInfo
		c.syncRoute(routeInfo)

		// TODO:记录同步时间
		log.Printf("~~~~~~~~~~~~~~~TODO~~~~~~~~~~~~~Received route: %v", len(routeInfo.Routes))
	}
	return nil
}
