package tools

import (
	"context"
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
	closed   bool
}

func NewGRPCPool(target string, maxConns int) (*GRPCPool, error) {
	pool := &GRPCPool{
		clients:  make([]*grpc.ClientConn, 0, maxConns),
		maxConns: maxConns,
		target:   target,
	}
	pool.cond = sync.NewCond(&pool.mu)

	for i := 0; i < maxConns; i++ {

		// 使用 context.WithTimeout 来设置连接超时
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		conn, err := grpc.DialContext(ctx, target, grpc.WithInsecure(), grpc.WithBlock())
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

	// 超时机制
	timer := time.NewTimer(5 * time.Second)
	defer timer.Stop()

	for len(p.clients) == 0 && !p.closed {
		p.cond.Wait()
		if !timer.Stop() {
			<-timer.C
		}
		timer.Reset(5 * time.Second)
	}

	if p.closed {
		return nil, fmt.Errorf("connection pool is closed")
	}

	if len(p.clients) == 0 {
		return nil, fmt.Errorf("timeout waiting for connection")
	}

	conn := p.clients[0]
	p.clients = p.clients[1:]
	return conn, nil
}

func (p *GRPCPool) Put(conn *grpc.ClientConn) {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.closed {
		conn.Close()
		return
	}

	if len(p.clients) < p.maxConns {
		p.clients = append(p.clients, conn)
		p.cond.Signal()
	} else {
		conn.Close() // 如果池已满，关闭连接
	}
}

func (p *GRPCPool) Close() {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.closed {
		return
	}

	for _, conn := range p.clients {
		conn.Close()
	}
	p.clients = nil
	p.closed = true
	p.cond.Broadcast()
}
