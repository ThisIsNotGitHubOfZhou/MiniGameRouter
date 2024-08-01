package tools

import (
	"fmt"
	"sync"
	"time"

	"google.golang.org/grpc"
)

type GRPCPoolNew struct {
	pool     *sync.Pool
	maxConns int
	target   string
}

func NewGRPCPoolNew(target string, maxConns int) (*GRPCPoolNew, error) {
	pool := &GRPCPoolNew{
		maxConns: maxConns,
		target:   target,
		pool: &sync.Pool{
			New: func() interface{} {
				conn, err := grpc.Dial(target, grpc.WithInsecure(), grpc.WithBlock(), grpc.WithTimeout(5*time.Second))
				if err != nil {
					return fmt.Errorf("failed to create connection: %v", err)
				}
				return conn
			},
		},
	}

	// 预先创建连接
	for i := 0; i < maxConns; i++ {
		conn := pool.pool.Get()
		if _, ok := conn.(error); ok {
			return nil, conn.(error)
		}
		pool.pool.Put(conn)
	}

	return pool, nil
}

func (p *GRPCPoolNew) Get() (*grpc.ClientConn, error) {
	conn := p.pool.Get()
	if err, ok := conn.(error); ok {
		return nil, err
	}
	return conn.(*grpc.ClientConn), nil
}

func (p *GRPCPoolNew) Put(conn *grpc.ClientConn) {
	p.pool.Put(conn)
}

func (p *GRPCPoolNew) Close() {
	for i := 0; i < p.maxConns; i++ {
		conn := p.pool.Get().(*grpc.ClientConn)
		conn.Close()
	}
}
