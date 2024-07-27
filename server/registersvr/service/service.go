package service

import (
	"registersvr/config"
	"registersvr/database"
	"time"
)

type Service interface {
	// 将服务实例注册到服务器上
	// metadata需要保证是键值对！
	// TODO:服务实例ID是应该int 还是 string？
	Register(name, host, port, protocol, metadata string, weight, timeout int) (string, error)

	// 将服务实例删除
	Deregister(id, name, host, port string) error
}

// 定义中间键服务
type ServiceMiddleware func(Service) Service

// 具体服务实现
type RegisterService struct{}

var _ Service = (*RegisterService)(nil)

// TODO：100ms一次？需要优化一下？？
func (s *RegisterService) Register(name, host, port, protocol, metadata string, weight, timeout int) (string, error) {

	// TODO:能否用host+port组合成为服务实例ID
	instanceInfo := map[string]interface{}{
		"service_name": name,
		"instance_id":  host + port,
		"host":         host,
		"port":         port,
		"protocol":     protocol,
		"weight":       weight,
		"timeout":      timeout,
		"metadata":     metadata,
	}
	err := database.RegisterServiceInstance(config.RedisClient, name, instanceInfo, time.Duration(timeout)*time.Second)
	if err != nil {
		config.Logger.Println("RegisterServiceInstance 出错:", err)
		return "", err
	}
	return host + port, nil
}

func (s *RegisterService) Deregister(id, name, host, port string) error {

	err := database.DeRegisterServiceInstance(config.RedisClient, name)
	if err != nil {
		config.Logger.Println("DeRegisterServiceInstance 出错:", err)
		return err
	}
	return nil
}
