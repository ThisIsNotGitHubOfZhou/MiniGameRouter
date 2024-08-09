package service

import (
	"context"
)

type RegisterService interface {
	// 返回注册完成的ID
	Register(ctx context.Context, name, host, port, protocol, metadata string, weight, timeout int) (string, error)

	DeRegister(ctx context.Context, id, name, host, port string) error
}
