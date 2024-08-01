package tools

import (
	"fmt"
	"sync"
	"time"

	"google.golang.org/grpc"
)

type GRPCPool struct {
	mu      sync.Mutex
	cond    *sync.Cond
	clients []*grpc.ClientConn

	maxConns int
	target   string
}

func NewGRPCPool(target string, maxConns int) (*GRPCPool, error) {
	pool := &GRPCPool{
		clients:  make([]*grpc.ClientConn, 0, maxConns),
		maxConns: maxConns,
		target:   target,
	}
	pool.cond = sync.NewCond(&pool.mu)

	for i := 0; i < maxConns; i++ {
		conn, err := grpc.Dial(target, grpc.WithInsecure(), grpc.WithBlock(), grpc.WithTimeout(5*time.Second)) // 阻塞5秒直到连接成功
		if err != nil {
			// 关闭已经创建的连接
			for _, c := range pool.clients {
				c.Close()
			}
			return nil, fmt.Errorf("failed to create connection: %v", err)
		}
		pool.clients = append(pool.clients, conn)
	}

	return pool, nil
}

func (p *GRPCPool) Get() (*grpc.ClientConn, error) {
	p.mu.Lock()
	defer p.mu.Unlock()

	for len(p.clients) == 0 {
		p.cond.Wait()
	}

	conn := p.clients[0]
	p.clients = p.clients[1:]
	return conn, nil
}

func (p *GRPCPool) Put(conn *grpc.ClientConn) {
	p.mu.Lock()
	defer p.mu.Unlock()

	p.clients = append(p.clients, conn)
	p.cond.Signal()
}

func (p *GRPCPool) Close() {
	p.mu.Lock()
	defer p.mu.Unlock()

	for _, conn := range p.clients {
		conn.Close()
	}
	p.clients = nil
}
