package service

import (
	"fmt"
	"registersvr/config"
	"registersvr/database"
	"time"
)

type Service interface {
	// 将服务实例注册到服务器上
	// metadata需要保证是键值对！
	Register(name, host, port, protocol, metadata string, weight, timeout int) (string, error)

	// 将服务实例删除
	Deregister(id, name, host, port string) error
}

// 定义中间键服务
type ServiceMiddleware func(Service) Service

// 具体服务实现
type RegisterService struct{}

var _ Service = (*RegisterService)(nil)

// TODO：200ms一次？需要优化一下？？
func (s *RegisterService) Register(name, host, port, protocol, metadata string, weight, timeout int) (string, error) {

	// TODO:需要跟discoversvr.convertMapToServiceInfo对齐~
	instanceInfo := map[string]interface{}{
		"service_name": name,
		"instance_id":  generateInstanceID(name, host, port),
		"host":         host,
		"port":         port,
		"protocol":     protocol,
		"weight":       weight,
		"timeout":      timeout,
		"metadata":     metadata,
	}
	config.Logger.Printf("[Info][register] 注册实例,名称：%v，信息：%v\n", name, instanceInfo)

	// 异步不处理错误版
	go func() {
		err := database.RegisterServiceInstance(config.RedisClient, instanceInfo["instance_id"].(string), instanceInfo, time.Duration(timeout)*time.Second*3+5*time.Second)
		if err != nil {
			config.Logger.Println("[Error][register] database.RegisterServiceInstance 出错:", err)
		}
	}()

	// 原始同步版（有问题后启用）~~~~~~~~~~~~~~~~~~~~~~
	// TODO:异步or同步设置一个开关？？
	//err := database.RegisterServiceInstance(config.RedisClient, instanceInfo["instance_id"].(string), instanceInfo, time.Duration(timeout)*time.Second*3+5*time.Second)
	//if err != nil {
	//	config.Logger.Println("[Error][register] database.RegisterServiceInstance 出错:", err)
	//	return "", err
	//}
	// ~~~~~~~~~~~~~~
	return instanceInfo["instance_id"].(string), nil
}

// 这样生成的唯一也方便，直接根据服务名查询~
func generateInstanceID(name, host, port string) string {
	res := name + host + port
	config.Logger.Println("[Info][register] 生成InstanceID : ", res)
	return res
}

func (s *RegisterService) Deregister(id, name, host, port string) error {
	config.Logger.Printf("[Info][register] 删除实例,名称：%v，id：%v\n", name, id)
	if id != generateInstanceID(name, host, port) {
		config.Logger.Printf("[Error][register] 服务实例ID: %v与生成不一样：%v\n", id, generateInstanceID(name, host, port))
		return fmt.Errorf("服务实例ID: %v与生成不一样：%v", id, generateInstanceID(name, host, port))
	}

	go func() {
		err := database.DeRegisterServiceInstance(config.RedisClient, id)
		if err != nil {
			config.Logger.Println("[Error][register] database.DeRegisterServiceInstance 出错:", err)
		}
	}()

	// 原始同步版（异步出错后启用）~~~~~~~~~~~~~~~~~
	err := database.DeRegisterServiceInstance(config.RedisClient, id)
	if err != nil {
		config.Logger.Println("[Error][register] database.DeRegisterServiceInstance 出错:", err)
		return err
	}
	// ~~~~~~~~~~~~~~~~~~~
	return nil
}
