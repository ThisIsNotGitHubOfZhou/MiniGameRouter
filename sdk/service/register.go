package service

import (
	"context"
)

type RegisterService interface {
	Register(ctx context.Context, name, host, port, protocol, metadata string, weight, timeout int) (string, error)

	Deregister(username, password string) error
}
