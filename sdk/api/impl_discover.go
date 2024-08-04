package api

import (
	"context"
	"fmt"
	discoverpb "github.com/ThisIsNotGitHubOfZhou/MiniGameRouter/sdk/proto/discover"
	grpctransport "github.com/go-kit/kit/transport/grpc"
	"hash/fnv"
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
		discoverpb.DiscoverServiceResponse{},
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

// 从cache里面读取数据
func (c *MiniClient) getRouteWithNameFromCache(name string) []*discoverpb.RouteInfo {
	// TODO:先用cache~
	c.routeCacheMu.RLock()
	defer c.routeCacheMu.RUnlock()
	if c.cache != nil {
		routes, ok := c.cache[name]
		c.cacheTime[name] = time.Now()
		if ok {
			return routes
		}
	}
	return nil
}

// 从grpc请求后更新cache
func (c *MiniClient) putRouteWithNameToCache(name string, routes []*discoverpb.RouteInfo) {
	c.routeCacheMu.Lock()
	defer c.routeCacheMu.Unlock()
	if c.cache == nil {
		c.cache = make(map[string][]*discoverpb.RouteInfo)
	}
	if c.cacheTime == nil {
		c.cacheTime = make(map[string]time.Time)
	}
	if c.prefixToIndex == nil {
		c.prefixToIndex = make(map[string][]int)
	}
	c.cache[name] = append(c.cache[name], routes...)
	c.cacheTime[name] = time.Now()
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

// 从cache里面读取数据
func (c *MiniClient) getRouteWithPrefixFromCache(name string, prefix string) []*discoverpb.RouteInfo {
	c.routeCacheMu.RLock()
	defer c.routeCacheMu.RUnlock()
	if c.cache != nil && c.prefixToIndex != nil {
		routesWithName, ok1 := c.cache[name]
		indexs, ok2 := c.prefixToIndex[name+":"+prefix]
		if ok1 && ok2 {
			c.cacheTime[name] = time.Now()
			var res []*discoverpb.RouteInfo
			for _, i := range indexs {
				res = append(res, routesWithName[i])
			}
			return res
		}
	}
	return nil
}

// 从grpc请求后更新cache
func (c *MiniClient) putRouteWithPrefixToCache(name string, prefix string, routes []*discoverpb.RouteInfo) {
	c.routeCacheMu.Lock()
	defer c.routeCacheMu.Unlock()
	if c.cache == nil {
		c.cache = make(map[string][]*discoverpb.RouteInfo)
	}
	if c.cacheTime == nil {
		c.cacheTime = make(map[string]time.Time)
	}
	if c.prefixToIndex == nil {
		c.prefixToIndex = make(map[string][]int)
	}
	c.cache[name] = append(c.cache[name], routes...) // 会不会有冗余数据：TODO,去冗余
	c.cacheTime[name] = time.Now()
	for i, route := range routes {
		if route.Prefix != "" {
			if route.Prefix != prefix || name != route.Name {
				fmt.Println("[Error][sdk] 服务名、前缀不一致~")
				continue
			}
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

		// TODO: NameNew、NamePrefixNew

	}
}

func getCacheStringKeys(m map[string][]*discoverpb.RouteInfo) []string {
	result := make([]string, 0)
	for k, _ := range m {
		result = append(result, k)
	}
	return result
}
func getPrefixToIndexKeys(m map[string][]int) []string {
	result := make([]string, 0)
	for k, _ := range m {
		result = append(result, k)
	}
	return result
}

// 用于去重
func hashRouteInfo(route *discoverpb.RouteInfo) uint64 {
	h := fnv.New64a()
	h.Write([]byte(route.Name))
	h.Write([]byte(route.Host))
	h.Write([]byte(route.Port))
	h.Write([]byte(route.Prefix))
	h.Write([]byte(route.Metadata))
	return h.Sum64()
}

// 去重本地缓存，注意不能加锁！！！！每次同步后都会调用
func (c *MiniClient) deduplicate() {
	// 对c.cache进行去重,并清理过期键

	// 对 c.cache 进行去重
	deleteKey := []string{}
	// 更新c.prefixToIndex
	totalSize := 0
	c.prefixToIndex = map[string][]int{}
	for key, routes := range c.cache {
		if c.cacheTime[key].Before(time.Now().Add(-10 * time.Minute)) { // 删除十分钟前的key
			deleteKey = append(deleteKey, key)
			continue
		}
		seen := make(map[uint64]bool)
		uniqueRoutes := []*discoverpb.RouteInfo{}

		for _, route := range routes {
			if route.Name != key {
				fmt.Println("[Error][sdk][cache] 存在cache中的路由名不一致")
				continue // 处理错误，跳过不一致的路由
			}
			hash := hashRouteInfo(route)
			if !seen[hash] {
				seen[hash] = true
				uniqueRoutes = append(uniqueRoutes, route)
				totalSize++
				if route.Prefix != "" {
					c.prefixToIndex[route.Name+":"+route.Prefix] = append(c.prefixToIndex[key], len(uniqueRoutes)-1)
				}
			}
		}
		c.cache[key] = uniqueRoutes
	}
	for _, val := range deleteKey {
		delete(c.cache, val)
		delete(c.cacheTime, val)
	}

	fmt.Println("[Info][sdk] 本地路由总数", totalSize)
}

// TODO:是否会让时间太长？
func (c *MiniClient) syncRoute(resp *discoverpb.RouteSyncResponse) {
	c.routeCacheMu.Lock()
	defer c.routeCacheMu.Unlock()
	c.lastUpdateTime = resp.NewVersion // 只有这里会更新时间戳，putRouteWithPrefixToCache、putRouteWithNameToCache不会更新~
	// TODO:resp.Routes去重,同步cache

	for _, route := range resp.Routes {
		c.cache[route.Name] = append(c.cache[route.Name], route)
	}
	c.deduplicate()
}

// cache同步线程
func (c *MiniClient) SyncCache() error {
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
			req := c.getRouteSyncRequest()
			//fmt.Printf("~~~~~~~~~~~~~~~TODO~~~~~~~~~~~~~同步请求:\n name :  %v %v \n nameprfix : %v %v\n", req.Name, len(req.Name), req.NamePrefix, len(req.NamePrefix))
			if err := stream.Send(req); err != nil {
				if err == io.EOF {
					return
				}
				log.Fatalf("failed to send request: %v", err)
			}
			time.Sleep(5 * time.Second) // 模拟定期发送请求,配置化
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

	}
	return nil
}
