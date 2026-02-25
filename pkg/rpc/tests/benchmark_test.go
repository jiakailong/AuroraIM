package tests

import (
	"context"
	"net"
	"testing"
	"time"

	rpcclient "kama_chat_server/pkg/rpc/client"
	rpcserver "kama_chat_server/pkg/rpc/server"
)

type benchEchoService struct{}

type benchEchoRequest struct {
	Value string `json:"value"`
}

type benchEchoResponse struct {
	Value string `json:"value"`
}

func (service *benchEchoService) Echo(ctx context.Context, request *benchEchoRequest) (*benchEchoResponse, error) {
	return &benchEchoResponse{Value: request.Value}, nil
}

func BenchmarkLocal_Echo(b *testing.B) {
	svc := &benchEchoService{}
	ctx := context.Background()
	request := &benchEchoRequest{Value: "ok"}

	b.ReportAllocs()
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			response, err := svc.Echo(ctx, request)
			if err != nil {
				b.Fatalf("local call failed: %v", err)
			}
			if response == nil || response.Value != "ok" {
				b.Fatalf("unexpected response: %+v", response)
			}
		}
	})
}

func BenchmarkRPC_Echo(b *testing.B) {
	srv := rpcserver.NewServer()
	if err := srv.RegisterService(&benchEchoService{}); err != nil {
		b.Fatalf("register service failed: %v", err)
	}

	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		b.Fatalf("listen failed: %v", err)
	}
	go func() {
		_ = srv.Serve(listener)
	}()
	defer func() {
		_ = srv.Close()
	}()

	client, err := rpcclient.NewClient(
		listener.Addr().String(),
		rpcclient.WithTimeout(3*time.Second),
		rpcclient.WithPoolMaxConn(16),
		rpcclient.WithPoolMaxIdleConn(16),
	)
	if err != nil {
		b.Fatalf("new client failed: %v", err)
	}
	defer func() {
		_ = client.Close()
	}()

	b.ReportAllocs()
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			var response benchEchoResponse
			if callErr := client.Call(context.Background(), "benchEchoService.Echo", &benchEchoRequest{Value: "ok"}, &response); callErr != nil {
				b.Fatalf("rpc call failed: %v", callErr)
			}
			if response.Value != "ok" {
				b.Fatalf("unexpected response value: %s", response.Value)
			}
		}
	})
}

func BenchmarkRPC_NoPool_Echo(b *testing.B) {
	srv := rpcserver.NewServer()
	if err := srv.RegisterService(&benchEchoService{}); err != nil {
		b.Fatalf("register service failed: %v", err)
	}

	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		b.Fatalf("listen failed: %v", err)
	}
	go func() {
		_ = srv.Serve(listener)
	}()
	defer func() {
		_ = srv.Close()
	}()

	client, err := rpcclient.NewClient(
		listener.Addr().String(),
		rpcclient.WithTimeout(3*time.Second),
		rpcclient.WithPoolMaxConn(1),
		rpcclient.WithPoolMaxIdleConn(0),
	)
	if err != nil {
		b.Fatalf("new client failed: %v", err)
	}
	defer func() {
		_ = client.Close()
	}()

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var response benchEchoResponse
		if callErr := client.Call(context.Background(), "benchEchoService.Echo", &benchEchoRequest{Value: "ok"}, &response); callErr != nil {
			b.Fatalf("rpc no-pool call failed: %v", callErr)
		}
		if response.Value != "ok" {
			b.Fatalf("unexpected response value: %s", response.Value)
		}
	}
}
