# 项目名称

## 目录结构

### ClientServerA
- **描述**：作为客户端的服务，后续会调用 SDK，目前还没开始写。

### sdk
- **描述**：包含 SDK 接口。

#### sdk/api
- **描述**：存储对 service 接口的实现。

#### sdk/proto
- **描述**：存储与服务器通信的协议。

#### sdk/service
- **描述**：存储提供的服务接口。

#### sdk/test
- **描述**：存储模拟服务调用相关接口的测试函数。

#### sdk/tools
- **描述**：存储 SDK 侧可能用到的工具，目前包含了 gRPC 的连接池。

### server
- **描述**：包括 registersvr、healthchecksvr 以及 discoversvr 的具体实现，所有服务都是基于 go-kit 框架。

#### server/demo
- **描述**：存储一些想加在服务里面的功能，先写一个 demo 测试能不能达到预期效果。

#### server/xxxsvr/config
- **描述**：存储配置相关以及配置的初始化。

#### server/xxxsvr/database
- **描述**：存储该服务与数据库交互的函数。

#### server/xxxsvr/endpoint
- **描述**：该服务的端点层，定义服务的业务逻辑接口。

#### server/xxxsvr/plugins
- **描述**：该服务的中间件，目前只有 log 中间件，后续可能会添加其他中间件。

#### server/xxxsvr/proto
- **描述**：该服务提供的服务协议。

#### server/xxxsvr/service
- **描述**：该服务的服务层，实现具体的业务逻辑。

#### server/xxxsvr/tools
- **描述**：该服务可能用到的工具，目前暂时没有。

#### server/xxxsvr/transport
- **描述**：该服务的传输层，处理网络传输和协议相关的逻辑。

#### server/xxxsvr/main.go
- **描述**：该服务的入口函数。

## 使用说明

### 安装依赖
```bash
go mod tidy
