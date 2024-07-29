package service

import (
	"healthchecksvr/config"
	"healthchecksvr/database"
	"net/http"
	"time"
)

type Service interface {
	// 服务器帮助定时发送HealthCheck
	HealthCheckS(Url string, name string, second int) error

	// 客户端主动发送的~
	// TODO:修改参数
	HealthCheckC(id, name, host, port string, second int) error
}

// 定义中间键服务
type ServiceMiddleware func(Service) Service

// 具体服务实现
type HealthCheckService struct {
}

var _ Service = (*HealthCheckService)(nil)

// TODO：暂时用name当作redis键

// TODO:性能检测~
func (s *HealthCheckService) HealthCheckS(Url string, name string, second int) error {

	go func() {
		retry := 0
		for {
			time.Sleep(time.Duration(second) * 3 * time.Second)
			if ok := healthCheck(Url); !ok {
				retry++
				if retry >= 3 {
					go database.DeRegisterServiceInstance(config.RedisClient, name) // 更新redis,反注册
					break
				}
				continue
			}
			go database.RenewServiceInstance(config.RedisClient, name, time.Duration(second)*3+5*time.Second)
		}
	}()

	return nil
}

func healthCheck(Url string) bool {
	// 访问url的地址确定是否返回true
	resp, err := http.Get(Url)
	if err != nil {
		return false
	}
	defer resp.Body.Close()

	return resp.StatusCode >= 200 && resp.StatusCode < 300
}

func (s *HealthCheckService) HealthCheckC(id, name, host, port string, second int) error {
	go database.RenewServiceInstance(config.RedisClient, name, time.Duration(second)*3+5*time.Second)
	return nil
}
