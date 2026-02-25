package transport

import (
	"errors"
	"fmt"
	"sync"
	"time"
)

var (
	ErrConnPoolClosed    = errors.New("transport: connection pool closed")
	ErrConnPoolExhausted = errors.New("transport: connection pool exhausted")
)

type ConnFactory func() (*Conn, error)

type ConnPool struct {
	mu sync.Mutex

	idle []*Conn

	totalConn int

	maxConn     int
	maxIdleConn int
	idleTimeout time.Duration

	factory ConnFactory
	closed  bool
}

func NewConnPool(maxConn, maxIdleConn int, idleTimeout time.Duration, factory ConnFactory) (*ConnPool, error) {
	if maxConn <= 0 {
		return nil, errors.New("transport: maxConn must be greater than 0")
	}
	if maxIdleConn < 0 {
		return nil, errors.New("transport: maxIdleConn must be greater than or equal to 0")
	}
	if maxIdleConn > maxConn {
		maxIdleConn = maxConn
	}
	if idleTimeout < 0 {
		return nil, errors.New("transport: idleTimeout must be greater than or equal to 0")
	}
	if factory == nil {
		return nil, errors.New("transport: factory is nil")
	}

	return &ConnPool{
		idle:        make([]*Conn, 0, maxIdleConn),
		maxConn:     maxConn,
		maxIdleConn: maxIdleConn,
		idleTimeout: idleTimeout,
		factory:     factory,
	}, nil
}

func (pool *ConnPool) Get() (*Conn, error) {
	pool.mu.Lock()
	if pool.closed {
		pool.mu.Unlock()
		return nil, ErrConnPoolClosed
	}

	for len(pool.idle) > 0 {
		last := len(pool.idle) - 1
		conn := pool.idle[last]
		pool.idle = pool.idle[:last]
		if pool.isConnReusable(conn) {
			pool.mu.Unlock()
			return conn, nil
		}
		pool.totalConn--
		pool.mu.Unlock()
		_ = conn.Close()
		pool.mu.Lock()
		if pool.closed {
			pool.mu.Unlock()
			return nil, ErrConnPoolClosed
		}
	}

	if pool.totalConn >= pool.maxConn {
		pool.mu.Unlock()
		return nil, ErrConnPoolExhausted
	}
	pool.totalConn++
	pool.mu.Unlock()

	conn, err := pool.factory()
	if err != nil {
		pool.mu.Lock()
		pool.totalConn--
		pool.mu.Unlock()
		return nil, fmt.Errorf("transport: create connection from factory: %w", err)
	}
	return conn, nil
}

func (pool *ConnPool) Put(conn *Conn) {
	if conn == nil {
		return
	}

	pool.mu.Lock()
	if pool.closed || !pool.isConnReusable(conn) || len(pool.idle) >= pool.maxIdleConn {
		pool.totalConn--
		pool.mu.Unlock()
		_ = conn.Close()
		return
	}
	pool.idle = append(pool.idle, conn)
	pool.mu.Unlock()
}

func (pool *ConnPool) Discard(conn *Conn) {
	if conn == nil {
		return
	}

	pool.mu.Lock()
	pool.totalConn--
	pool.mu.Unlock()
	_ = conn.Close()
}

func (pool *ConnPool) Close() error {
	pool.mu.Lock()
	if pool.closed {
		pool.mu.Unlock()
		return nil
	}
	pool.closed = true
	idleSnapshot := make([]*Conn, 0, len(pool.idle))
	idleSnapshot = append(idleSnapshot, pool.idle...)
	pool.idle = nil
	pool.totalConn = 0
	pool.mu.Unlock()

	for _, conn := range idleSnapshot {
		_ = conn.Close()
	}
	return nil
}

func (pool *ConnPool) isConnReusable(conn *Conn) bool {
	if conn == nil || !conn.IsHealthy() {
		return false
	}
	if pool.idleTimeout <= 0 {
		return true
	}
	return time.Since(conn.LastActiveAt()) <= pool.idleTimeout
}
