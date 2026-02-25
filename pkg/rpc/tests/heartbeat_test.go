package tests

import (
	"errors"
	"io"
	"net"
	"sync/atomic"
	"testing"
	"time"

	"kama_chat_server/pkg/rpc/transport"
)

func TestHeartbeatTimeout(t *testing.T) {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen failed: %v", err)
	}
	defer func() {
		_ = listener.Close()
	}()

	serverClosed := make(chan struct{})
	go func() {
		defer close(serverClosed)
		conn, acceptErr := listener.Accept()
		if acceptErr != nil {
			return
		}
		defer func() {
			_ = conn.Close()
		}()
		_, _ = io.Copy(io.Discard, conn)
	}()

	rawConn, err := net.Dial("tcp", listener.Addr().String())
	if err != nil {
		t.Fatalf("dial failed: %v", err)
	}

	errCh := make(chan error, 1)
	conn := transport.NewConn(rawConn, nil, func(conn *transport.Conn, closeErr error) {
		errCh <- closeErr
	})
	conn.Start()

	if err = conn.EnableHeartbeat(20*time.Millisecond, 60*time.Millisecond); err != nil {
		t.Fatalf("enable heartbeat failed: %v", err)
	}

	select {
	case closeErr := <-errCh:
		if !errors.Is(closeErr, transport.ErrHeartbeatTimeout) {
			t.Fatalf("expected ErrHeartbeatTimeout, got=%v", closeErr)
		}
	case <-time.After(800 * time.Millisecond):
		t.Fatal("wait heartbeat timeout exceeded")
	}

	if conn.IsHealthy() {
		t.Fatal("connection should be unhealthy after heartbeat timeout")
	}

	_ = conn.Close()
	<-serverClosed
}

func TestHeartbeatPingPongKeepsConnectionHealthy(t *testing.T) {
	srv := transport.NewServer(nil, nil)
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen failed: %v", err)
	}
	defer func() {
		_ = srv.Close()
	}()
	go func() {
		_ = srv.Serve(listener)
	}()

	rawConn, err := net.Dial("tcp", listener.Addr().String())
	if err != nil {
		t.Fatalf("dial failed: %v", err)
	}

	var closed atomic.Bool
	conn := transport.NewConn(rawConn, nil, func(conn *transport.Conn, closeErr error) {
		closed.Store(true)
	})
	conn.Start()
	defer func() {
		_ = conn.Close()
	}()

	if err = conn.EnableHeartbeat(20*time.Millisecond, 200*time.Millisecond); err != nil {
		t.Fatalf("enable heartbeat failed: %v", err)
	}

	time.Sleep(140 * time.Millisecond)
	if closed.Load() {
		t.Fatal("connection should stay open when pong is received")
	}
	if !conn.IsHealthy() {
		t.Fatal("connection should be healthy when pong is received")
	}
}
