package transport

import (
	"errors"
	"fmt"
	"net"
	"sync"
)

type Server struct {
	listener net.Listener

	onFrame OnFrameFunc
	onClose OnCloseFunc

	mu        sync.Mutex
	conns     map[*Conn]struct{}
	closed    bool
	closeOnce sync.Once
}

func NewServer(onFrame OnFrameFunc, onClose OnCloseFunc) *Server {
	return &Server{
		onFrame: onFrame,
		onClose: onClose,
		conns:   make(map[*Conn]struct{}),
	}
}

func (server *Server) ListenAndServe(network, address string) error {
	listener, err := net.Listen(network, address)
	if err != nil {
		return fmt.Errorf("transport: listen %s %s: %w", network, address, err)
	}
	return server.Serve(listener)
}

func (server *Server) Serve(listener net.Listener) error {
	server.mu.Lock()
	if server.closed {
		server.mu.Unlock()
		_ = listener.Close()
		return ErrServerClosed
	}
	server.listener = listener
	server.mu.Unlock()

	for {
		rawConn, err := listener.Accept()
		if err != nil {
			if server.isClosed() {
				return ErrServerClosed
			}
			var netErr net.Error
			if errors.As(err, &netErr) && netErr.Temporary() {
				continue
			}
			return fmt.Errorf("transport: accept connection: %w", err)
		}

		conn := NewConn(rawConn, server.onFrame, server.wrapOnClose())
		server.addConn(conn)
		conn.Start()
	}
}

func (server *Server) Close() error {
	var closeErr error
	server.closeOnce.Do(func() {
		server.mu.Lock()
		server.closed = true
		listener := server.listener
		conns := make([]*Conn, 0, len(server.conns))
		for conn := range server.conns {
			conns = append(conns, conn)
		}
		server.mu.Unlock()

		if listener != nil {
			if err := listener.Close(); err != nil {
				closeErr = fmt.Errorf("transport: close listener: %w", err)
			}
		}
		for _, conn := range conns {
			_ = conn.Close()
		}
	})
	return closeErr
}

func (server *Server) addConn(conn *Conn) {
	server.mu.Lock()
	defer server.mu.Unlock()
	if server.closed {
		_ = conn.Close()
		return
	}
	server.conns[conn] = struct{}{}
}

func (server *Server) removeConn(conn *Conn) {
	server.mu.Lock()
	defer server.mu.Unlock()
	delete(server.conns, conn)
}

func (server *Server) wrapOnClose() OnCloseFunc {
	return func(conn *Conn, err error) {
		server.removeConn(conn)
		if server.onClose != nil {
			server.onClose(conn, err)
		}
	}
}

func (server *Server) isClosed() bool {
	server.mu.Lock()
	defer server.mu.Unlock()
	return server.closed
}

var ErrServerClosed = errors.New("transport: server closed")
