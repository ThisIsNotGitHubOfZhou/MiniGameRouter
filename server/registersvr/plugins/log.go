package plugins

import (
	"registersvr/service"
	"time"

	"github.com/go-kit/log"
)

// 添加中间件
type loggingMiddleware struct {
	service.Service // 如果没有重新定义service则直接会调用这个service
	logger          log.Logger
}

// LoggingMiddleware make logging middleware
func LoggingMiddleware(logger log.Logger) service.ServiceMiddleware {
	return func(next service.Service) service.Service {
		return &loggingMiddleware{next, logger}
	}
}

func (mw *loggingMiddleware) Register(name, host, port, protocol, metadata string, weight, timeout int) (instanceID string, err error) {
	// 函数执行结束后打印日志
	defer func(begin time.Time) {
		mw.logger.Log(
			"function", "Register",
			"name", name,
			"host", host,
			"port", port,
			"protocol", protocol,
			"weight", weight,
			"metadata", metadata,
			"took", time.Since(begin),
		)
	}(time.Now())

	instanceID, err = mw.Service.Register(name, host, port, protocol, metadata, weight, timeout)
	return
}

func (mw loggingMiddleware) Deregister(id, name, host, port string) (err error) {
	defer func(begin time.Time) {
		mw.logger.Log(
			"function", "Deregister",
			"took", time.Since(begin),
		)
	}(time.Now())
	err = mw.Service.Deregister(id, name, host, port)
	return
}

//func (mw loggingMiddleware) HealthCheck() (result bool) {
//	defer func(begin time.Time) {
//		mw.logger.Log(
//			"function", "HealthChcek",
//			"result", result,
//			"took", time.Since(begin),
//		)
//	}(time.Now())
//	result = mw.Service.HealthCheck()
//	return
//}
