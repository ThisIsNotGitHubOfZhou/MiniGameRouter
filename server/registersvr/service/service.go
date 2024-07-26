package service

import "registersvr/config"

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

type RegisterService struct{}

var _ Service = (*RegisterService)(nil)

func (s *RegisterService) Register(name, host, port, protocol, metadata string, weight, timeout int) (string, error) {
	// TODO:能否用host+port组合成为服务实例ID
	config.Logger.Println("service~~~~~~hello world")
	return "465798222", nil
}

func (s *RegisterService) Deregister(id, name, host, port string) error {
	return nil
}
