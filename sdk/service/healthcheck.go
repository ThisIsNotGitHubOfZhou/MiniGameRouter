package service

import (
	"context"
)

type HealthCheckService interface {
	HealthCheckS(ctx context.Context, port string) error // 服务器那边帮忙轮询检查,需要本地暴露接口~

	HealthCheckC(ctx context.Context, id, name, port, ip string, timeout int) error // 本地定时发送心跳
}
