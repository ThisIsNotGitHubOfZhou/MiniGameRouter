package service

import (
	"healthchecksvr/config"
	"healthchecksvr/database"
	"net/http"
	"time"
)

type Service interface {
	// 服务器帮助定时发送HealthCheck
	HealthCheckS(Url string, id string, second int) error

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
func (s *HealthCheckService) HealthCheckS(Url string, id string, second int) error {
	config.Logger.Printf("[Info][healthcheck] HealthCheckS服务器主动检查,URL:%v,id:%v,时长:%v\n", Url, id, second)
	go func() {
		retry := 0
		for {
			time.Sleep(time.Duration(second) * 3 * time.Second)
			if ok := healthCheck(Url); !ok {
				retry++
				if retry >= 3 {
					go database.DeRegisterServiceInstance(config.RedisClient, id) // 更新redis,反注册
					break
				}
				continue
			}
			go database.RenewServiceInstance(config.RedisClient, id, time.Duration(second)*3*time.Second+5*time.Second)
		}
	}()

	return nil
}

func healthCheck(Url string) bool {
	// 访问url的地址确定是否返回true
	resp, err := http.Get(Url)
	if err != nil {
		config.Logger.Printf("[Error][healthcheck] HealthCheckS服务器主动检查,访问接口URL出错,URL:%v,错误:%v\n", Url, err)
		return false
	}
	defer resp.Body.Close()

	return resp.StatusCode >= 200 && resp.StatusCode < 300
}

func (s *HealthCheckService) HealthCheckC(id, name, host, port string, second int) error {
	config.Logger.Printf("[Info][healthcheck] HealthCheckC客户端主动发送,id:%v,name:%v,时长:%v\n", id, name, second)
	go database.RenewServiceInstance(config.RedisClient, id, time.Duration(second)*3*time.Second+5*time.Second)
	return nil
}
