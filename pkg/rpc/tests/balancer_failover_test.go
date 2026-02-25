package tests

import (
	"context"
	"net"
	"testing"

	rpcclient "kama_chat_server/pkg/rpc/client"
	"kama_chat_server/pkg/rpc/registry"
)

func TestClientLoadBalanceWithRoundRobin(t *testing.T) {
	server1Addr, closeServer1 := startRouteRPCServer(t, "rr-server-1")
	defer closeServer1()
	server2Addr, closeServer2 := startRouteRPCServer(t, "rr-server-2")
	defer closeServer2()

	memoryRegistry := registry.NewMemoryRegistry()
	if err := memoryRegistry.Set("routeService", []registry.Instance{{Address: server1Addr}, {Address: server2Addr}}); err != nil {
		t.Fatalf("set memory registry failed: %v", err)
	}

	client, err := rpcclient.NewClient("",
		rpcclient.WithServiceName("routeService"),
		rpcclient.WithRegistry(memoryRegistry),
		rpcclient.WithPoolMaxConn(1),
		rpcclient.WithPoolMaxIdleConn(1),
	)
	if err != nil {
		t.Fatalf("new client failed: %v", err)
	}
	defer func() {
		_ = client.Close()
	}()

	hit := map[string]int{}
	for i := 0; i < 6; i++ {
		var response routeResponse
		if err = client.Call(context.Background(), "routeService.Echo", &routeRequest{Value: "rr"}, &response); err != nil {
			t.Fatalf("rpc call failed: %v", err)
		}
		hit[response.Source]++
	}

	if hit["rr-server-1"] == 0 || hit["rr-server-2"] == 0 {
		t.Fatalf("expected load balanced calls across both instances, got: %#v", hit)
	}
}

func TestClientFailoverToNextInstance(t *testing.T) {
	unavailableAddr := nextClosedAddress(t)
	serverAddr, closeServer := startRouteRPCServer(t, "failover-server")
	defer closeServer()

	client, err := rpcclient.NewClient("",
		rpcclient.WithServiceName("routeService"),
		rpcclient.WithServiceInstances(map[string][]registry.Instance{
			"routeService": {
				{Address: unavailableAddr},
				{Address: serverAddr},
			},
		}),
		rpcclient.WithPoolMaxConn(1),
		rpcclient.WithPoolMaxIdleConn(1),
	)
	if err != nil {
		t.Fatalf("new client failed: %v", err)
	}
	defer func() {
		_ = client.Close()
	}()

	var response routeResponse
	if err = client.Call(context.Background(), "routeService.Echo", &routeRequest{Value: "ok"}, &response); err != nil {
		t.Fatalf("rpc call with failover failed: %v", err)
	}
	if response.Source != "failover-server" {
		t.Fatalf("unexpected response source=%s", response.Source)
	}
}

func nextClosedAddress(t *testing.T) string {
	t.Helper()
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen temp address failed: %v", err)
	}
	address := listener.Addr().String()
	if err = listener.Close(); err != nil {
		t.Fatalf("close temp listener failed: %v", err)
	}
	return address
}
