package tests

import (
	"context"
	"errors"
	"net"
	"sync/atomic"
	"testing"
	"time"

	rpcclient "kama_chat_server/pkg/rpc/client"
	rpcerrors "kama_chat_server/pkg/rpc/errors"
	rpcserver "kama_chat_server/pkg/rpc/server"
)

type sleepService struct{}

type sleepRequest struct {
	Ms int `json:"ms"`
}

type sleepResponse struct {
	Done bool `json:"done"`
}

func (service *sleepService) Sleep(ctx context.Context, request *sleepRequest) (*sleepResponse, error) {
	time.Sleep(time.Duration(request.Ms) * time.Millisecond)
	return &sleepResponse{Done: true}, nil
}

type unstableService struct {
	failFirstN int32
	calls      atomic.Int32
}

type unstableRequest struct {
	Value string `json:"value"`
}

type unstableResponse struct {
	Value string `json:"value"`
}

func (service *unstableService) Fetch(ctx context.Context, request *unstableRequest) (*unstableResponse, error) {
	callNum := service.calls.Add(1)
	if callNum <= service.failFirstN {
		return nil, rpcerrors.New(rpcerrors.Unavailable, "temporary unavailable", nil)
	}
	return &unstableResponse{Value: request.Value}, nil
}

func (service *unstableService) CallCount() int32 {
	return service.calls.Load()
}

func TestTimeoutAndPendingCleanup(t *testing.T) {
	srv := rpcserver.NewServer()
	if err := srv.RegisterService(&sleepService{}); err != nil {
		t.Fatalf("register service failed: %v", err)
	}

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

	client, err := rpcclient.NewClient(listener.Addr().String(), rpcclient.WithTimeout(50*time.Millisecond))
	if err != nil {
		t.Fatalf("new client failed: %v", err)
	}
	defer func() {
		_ = client.Close()
	}()

	var response sleepResponse
	err = client.Call(context.Background(), "sleepService.Sleep", &sleepRequest{Ms: 200}, &response)
	if err == nil {
		t.Fatal("expected timeout error, got nil")
	}

	var rpcErr *rpcerrors.RpcError
	if !errors.As(err, &rpcErr) {
		t.Fatalf("expected *RpcError, got %T", err)
	}
	if rpcErr.Code != rpcerrors.Timeout {
		t.Fatalf("expected timeout code, got %d", rpcErr.Code)
	}

	time.Sleep(20 * time.Millisecond)
	if pending := client.PendingCount(); pending != 0 {
		t.Fatalf("pending requests should be cleaned up, got=%d", pending)
	}
}

func TestRetrySucceedsForIdempotentMethod(t *testing.T) {
	service := &unstableService{failFirstN: 1}
	srv := rpcserver.NewServer()
	if err := srv.RegisterService(service); err != nil {
		t.Fatalf("register service failed: %v", err)
	}

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

	client, err := rpcclient.NewClient(
		listener.Addr().String(),
		rpcclient.WithTimeout(500*time.Millisecond),
		rpcclient.WithRetryMax(2),
		rpcclient.WithRetryBackoff(10*time.Millisecond, 20*time.Millisecond, 0),
		rpcclient.WithIdempotentMethod("unstableService.Fetch"),
	)
	if err != nil {
		t.Fatalf("new client failed: %v", err)
	}
	defer func() {
		_ = client.Close()
	}()

	var response unstableResponse
	err = client.Call(context.Background(), "unstableService.Fetch", &unstableRequest{Value: "ok"}, &response)
	if err != nil {
		t.Fatalf("expected retry success, got err=%v", err)
	}
	if response.Value != "ok" {
		t.Fatalf("unexpected response value=%s", response.Value)
	}
	if service.CallCount() < 2 {
		t.Fatalf("expected retried call count >= 2, got=%d", service.CallCount())
	}
}

func TestNonIdempotentMethodDoesNotRetry(t *testing.T) {
	service := &unstableService{failFirstN: 1}
	srv := rpcserver.NewServer()
	if err := srv.RegisterService(service); err != nil {
		t.Fatalf("register service failed: %v", err)
	}

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

	client, err := rpcclient.NewClient(
		listener.Addr().String(),
		rpcclient.WithTimeout(500*time.Millisecond),
		rpcclient.WithRetryMax(2),
		rpcclient.WithRetryBackoff(10*time.Millisecond, 20*time.Millisecond, 0),
	)
	if err != nil {
		t.Fatalf("new client failed: %v", err)
	}
	defer func() {
		_ = client.Close()
	}()

	var response unstableResponse
	err = client.Call(context.Background(), "unstableService.Fetch", &unstableRequest{Value: "ok"}, &response)
	if err == nil {
		t.Fatal("expected unavailable error, got nil")
	}

	var rpcErr *rpcerrors.RpcError
	if !errors.As(err, &rpcErr) {
		t.Fatalf("expected *RpcError, got %T", err)
	}
	if rpcErr.Code != rpcerrors.Unavailable {
		t.Fatalf("expected unavailable code, got %d", rpcErr.Code)
	}
	if service.CallCount() != 1 {
		t.Fatalf("expected no retry for non-idempotent method, calls=%d", service.CallCount())
	}
}
