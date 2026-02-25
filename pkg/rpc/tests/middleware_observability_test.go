package tests

import (
	"context"
	"fmt"
	"net"
	"sync"
	"testing"
	"time"

	"kama_chat_server/pkg/rpc/api"
	rpcclient "kama_chat_server/pkg/rpc/client"
	"kama_chat_server/pkg/rpc/observability"
	rpcserver "kama_chat_server/pkg/rpc/server"
)

type captureLogger struct {
	mu      sync.Mutex
	entries []observability.Fields
}

func (logger *captureLogger) Info(message string, fields observability.Fields) {
	logger.add(fields)
}

func (logger *captureLogger) Error(message string, fields observability.Fields) {
	logger.add(fields)
}

func (logger *captureLogger) add(fields observability.Fields) {
	logger.mu.Lock()
	defer logger.mu.Unlock()
	copyFields := make(observability.Fields, len(fields))
	for key, value := range fields {
		copyFields[key] = value
	}
	logger.entries = append(logger.entries, copyFields)
}

func (logger *captureLogger) last() (observability.Fields, bool) {
	logger.mu.Lock()
	defer logger.mu.Unlock()
	if len(logger.entries) == 0 {
		return nil, false
	}
	return logger.entries[len(logger.entries)-1], true
}

type obsEchoService struct{}

type obsEchoRequest struct {
	Value string `json:"value"`
}

type obsEchoResponse struct {
	Value string `json:"value"`
}

func (service *obsEchoService) Echo(ctx context.Context, req *obsEchoRequest) (*obsEchoResponse, error) {
	return &obsEchoResponse{Value: req.Value}, nil
}

func TestLoggingAndTracingMiddleware(t *testing.T) {
	previousLogger := observability.GetDefaultLogger()
	logger := &captureLogger{}
	observability.SetDefaultLogger(logger)
	defer observability.SetDefaultLogger(previousLogger)

	srv := rpcserver.NewServer()
	if err := srv.RegisterService(&obsEchoService{}); err != nil {
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

	client, err := rpcclient.NewClient(listener.Addr().String())
	if err != nil {
		t.Fatalf("create client failed: %v", err)
	}
	defer func() {
		_ = client.Close()
	}()

	ctx := api.SetTraceID(context.Background(), "trace-day8-fixed")
	var response obsEchoResponse
	if err = client.Call(ctx, "obsEchoService.Echo", &obsEchoRequest{Value: "ok"}, &response); err != nil {
		t.Fatalf("rpc call failed: %v", err)
	}
	if response.Value != "ok" {
		t.Fatalf("unexpected response value: %s", response.Value)
	}

	entry, ok := logger.last()
	if !ok {
		t.Fatal("expected at least one log entry")
	}
	assertFieldExists(t, entry, "method")
	assertFieldExists(t, entry, "request_id")
	assertFieldExists(t, entry, "trace_id")
	assertFieldExists(t, entry, "latency_ms")
	assertFieldExists(t, entry, "code")
	if entry["trace_id"] != "trace-day8-fixed" {
		t.Fatalf("unexpected trace_id in log entry: %v", entry["trace_id"])
	}
}

func TestMetricsMiddleware(t *testing.T) {
	previousMetrics := observability.GetDefaultMetrics()
	metrics := observability.NewInMemoryMetrics()
	observability.SetDefaultMetrics(metrics)
	defer observability.SetDefaultMetrics(previousMetrics)

	srv := rpcserver.NewServer()
	if err := srv.RegisterService(&obsEchoService{}); err != nil {
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

	client, err := rpcclient.NewClient(listener.Addr().String())
	if err != nil {
		t.Fatalf("create client failed: %v", err)
	}
	defer func() {
		_ = client.Close()
	}()

	for i := 0; i < 3; i++ {
		var response obsEchoResponse
		if err = client.Call(context.Background(), "obsEchoService.Echo", &obsEchoRequest{Value: fmt.Sprintf("v-%d", i)}, &response); err != nil {
			t.Fatalf("rpc call failed: %v", err)
		}
	}

	time.Sleep(20 * time.Millisecond)
	snapshot := metrics.Snapshot()
	key := "obsEchoService.Echo|000"
	if snapshot.RequestsTotal[key] < 3 {
		t.Fatalf("expected at least 3 successful requests, got=%d snapshot=%v", snapshot.RequestsTotal[key], snapshot.RequestsTotal)
	}
	if snapshot.LatencyTotal[key] <= 0 {
		t.Fatalf("expected positive latency total, got=%v", snapshot.LatencyTotal[key])
	}
	if inFlight, ok := snapshot.InFlight["obsEchoService.Echo"]; !ok || inFlight != 0 {
		t.Fatalf("expected inflight to return 0, got=%d exists=%v", inFlight, ok)
	}
}

func assertFieldExists(t *testing.T, fields observability.Fields, key string) {
	t.Helper()
	if _, ok := fields[key]; !ok {
		t.Fatalf("expected field %s in log entry, got=%v", key, fields)
	}
}
