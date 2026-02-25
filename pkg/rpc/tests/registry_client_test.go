package tests

import (
	"context"
	"net"
	"testing"

	rpcclient "kama_chat_server/pkg/rpc/client"
	"kama_chat_server/pkg/rpc/registry"
	rpcserver "kama_chat_server/pkg/rpc/server"
)

type routeService struct {
	source string
}

type routeRequest struct {
	Value string `json:"value"`
}

type routeResponse struct {
	Source string `json:"source"`
	Value  string `json:"value"`
}

func (service *routeService) Echo(ctx context.Context, request *routeRequest) (*routeResponse, error) {
	return &routeResponse{Source: service.source, Value: request.Value}, nil
}

func TestClientSelectsInstancesFromMemoryRegistry(t *testing.T) {
	server1Addr, closeServer1 := startRouteRPCServer(t, "server-1")
	defer closeServer1()
	server2Addr, closeServer2 := startRouteRPCServer(t, "server-2")
	defer closeServer2()

	memoryRegistry := registry.NewMemoryRegistry()
	if err := memoryRegistry.Set("routeService", []registry.Instance{{Address: server1Addr}, {Address: server2Addr}}); err != nil {
		t.Fatalf("set memory registry failed: %v", err)
	}

	client, err := rpcclient.NewClient("",
		rpcclient.WithServiceName("routeService"),
		rpcclient.WithRegistry(memoryRegistry),
		rpcclient.WithPoolMaxConn(1),
		rpcclient.WithPoolMaxIdleConn(0),
	)
	if err != nil {
		t.Fatalf("new client failed: %v", err)
	}
	defer func() {
		_ = client.Close()
	}()

	var resp1 routeResponse
	if err = client.Call(context.Background(), "routeService.Echo", &routeRequest{Value: "a"}, &resp1); err != nil {
		t.Fatalf("first rpc call failed: %v", err)
	}
	var resp2 routeResponse
	if err = client.Call(context.Background(), "routeService.Echo", &routeRequest{Value: "b"}, &resp2); err != nil {
		t.Fatalf("second rpc call failed: %v", err)
	}

	if resp1.Source == resp2.Source {
		t.Fatalf("expected calls routed to different instances, got same source=%s", resp1.Source)
	}
}

func TestClientSelectsInstancesFromServiceMap(t *testing.T) {
	serverAddr, closeServer := startRouteRPCServer(t, "map-server")
	defer closeServer()

	client, err := rpcclient.NewClient("",
		rpcclient.WithServiceName("routeService"),
		rpcclient.WithServiceInstances(map[string][]registry.Instance{
			"routeService": []registry.Instance{{Address: serverAddr}},
		}),
	)
	if err != nil {
		t.Fatalf("new client failed: %v", err)
	}
	defer func() {
		_ = client.Close()
	}()

	var response routeResponse
	if err = client.Call(context.Background(), "routeService.Echo", &routeRequest{Value: "ok"}, &response); err != nil {
		t.Fatalf("rpc call failed: %v", err)
	}
	if response.Source != "map-server" {
		t.Fatalf("unexpected response source=%s", response.Source)
	}
}

func startRouteRPCServer(t *testing.T, source string) (string, func()) {
	t.Helper()

	srv := rpcserver.NewServer()
	if err := srv.RegisterService(&routeService{source: source}); err != nil {
		t.Fatalf("register service failed: %v", err)
	}

	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen failed: %v", err)
	}
	go func() {
		_ = srv.Serve(listener)
	}()

	closeFunc := func() {
		_ = srv.Close()
	}
	return listener.Addr().String(), closeFunc
}
