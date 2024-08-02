package api

import (
	"context"
	"fmt"
	discoverpb "github.com/ThisIsNotGitHubOfZhou/MiniGameRouter/sdk/proto/discover"
	grpctransport "github.com/go-kit/kit/transport/grpc"
	"io"
	"log"
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
		index, ok2 := c.prefixToIndex[prefix]
		if ok1 && ok2 {
			var res []*discoverpb.RouteInfo
			for _, i := range index {
				res = append(res, routesWithName[i])
			}
			return res
		}
	}
	return nil
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

// cache同步线程
func (c *MiniClient) SyncCache() error {
	// TODO: 同步cache
	// 利用stream流实现

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
	conn, err := c.DiscoverGRPCPools[tempFlag%(int64(len(c.DiscoverGRPCPools)))].Get() // 优化后
	defer c.DiscoverGRPCPools[tempFlag%(int64(len(c.DiscoverGRPCPools)))].Put(conn)
	if err != nil {
		return err
	}

	client := discoverpb.NewDiscoverServiceClient(conn)

	req := &discoverpb.RouteSyncRequest{
		// 填充请求数据
	}
	// 调用 SyncRoutes 方法
	stream, err := client.SyncRoutes(context.Background(), req)
	if err != nil {
		log.Fatalf("could not call SyncRoutes: %v", err)
	}

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
		log.Printf("Received route: %v", routeInfo)
	}
	return nil
}
