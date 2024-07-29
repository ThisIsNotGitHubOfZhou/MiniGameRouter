package plugins

import (
	"healthchecksvr/service"
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

func (mw *loggingMiddleware) HealthCheckS(Url string, name string, second int) (err error) {
	// 函数执行结束后打印日志
	defer func(begin time.Time) {
		mw.logger.Log(
			"function", "HealthCheckS",
			"instance_id", name,
			"Url", Url,
			"second", second,
			"took", time.Since(begin),
		)
	}(time.Now())

	err = mw.Service.HealthCheckS(Url, name, second)
	return
}

func (mw loggingMiddleware) HealthCheckC(id, name, host, port string, second int) (err error) {
	defer func(begin time.Time) {
		mw.logger.Log(
			"function", "HealthCheckC",
			"instance_id", name,
			"second", second,
			"took", time.Since(begin),
		)
	}(time.Now())
	err = mw.Service.HealthCheckC(id, name, host, port, second)
	return
}
