package tests

import (
	"context"
	"net"
	"sync/atomic"
	"testing"

	rpcclient "kama_chat_server/pkg/rpc/client"
	rpcserver "kama_chat_server/pkg/rpc/server"
)

type poolEchoService struct{}

type poolEchoRequest struct {
	Value string `json:"value"`
}

type poolEchoResponse struct {
	Value string `json:"value"`
}

func (service *poolEchoService) Echo(ctx context.Context, req *poolEchoRequest) (*poolEchoResponse, error) {
	return &poolEchoResponse{Value: req.Value}, nil
}

func TestClientConnectionPoolReusesConnection(t *testing.T) {
	srv := rpcserver.NewServer()
	if err := srv.RegisterService(&poolEchoService{}); err != nil {
		t.Fatalf("register service failed: %v", err)
	}

	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen failed: %v", err)
	}
	defer func() {
		_ = srv.Close()
	}()

	var acceptCount atomic.Int32
	wrappedListener := &countingListener{Listener: listener, counter: &acceptCount}
	go func() {
		_ = srv.Serve(wrappedListener)
	}()

	client, err := rpcclient.NewClient(listener.Addr().String(),
		rpcclient.WithPoolMaxConn(1),
		rpcclient.WithPoolMaxIdleConn(1),
	)
	if err != nil {
		t.Fatalf("create client failed: %v", err)
	}
	defer func() {
		_ = client.Close()
	}()

	for index := 0; index < 2; index++ {
		var response poolEchoResponse
		if err = client.Call(context.Background(), "poolEchoService.Echo", &poolEchoRequest{Value: "ok"}, &response); err != nil {
			t.Fatalf("rpc call failed at index=%d: %v", index, err)
		}
		if response.Value != "ok" {
			t.Fatalf("unexpected response at index=%d: %+v", index, response)
		}
	}

	if got := acceptCount.Load(); got != 1 {
		t.Fatalf("expected one accepted connection for two calls, got=%d", got)
	}
}

type countingListener struct {
	net.Listener
	counter *atomic.Int32
}

func (listener *countingListener) Accept() (net.Conn, error) {
	conn, err := listener.Listener.Accept()
	if err != nil {
		return nil, err
	}
	listener.counter.Add(1)
	return conn, nil
}
