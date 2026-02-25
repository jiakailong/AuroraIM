package client

import (
	"context"
	"errors"
	"fmt"
	"net"
	"strings"
	"sync"
	"sync/atomic"

	rpcerrors "kama_chat_server/pkg/rpc/errors"
	"kama_chat_server/pkg/rpc/protocol"
	"kama_chat_server/pkg/rpc/registry"
	"kama_chat_server/pkg/rpc/transport"
)

var ErrClientClosed = errors.New("client: closed")

type pendingResult struct {
	frame protocol.Frame
	err   error
}

type Client struct {
	options Options

	instanceAddresses []string
	instanceCursor    atomic.Uint64

	poolsMu sync.Mutex
	pools   map[string]*transport.ConnPool

	requestID atomic.Uint64

	pendingMu sync.Mutex
	pending   map[uint64]chan pendingResult

	stateMu sync.RWMutex
	closed  bool
}

func NewClient(targetAddress string, options ...Option) (*Client, error) {
	clientOptions := defaultOptions()
	for _, option := range options {
		option(&clientOptions)
	}

	instanceAddresses, err := resolveTargetAddresses(targetAddress, clientOptions)
	if err != nil {
		return nil, err
	}

	client := &Client{
		options:           clientOptions,
		instanceAddresses: instanceAddresses,
		pending:           make(map[uint64]chan pendingResult),
		pools:             make(map[string]*transport.ConnPool),
	}
	return client, nil
}

func resolveTargetAddresses(targetAddress string, options Options) ([]string, error) {
	if targetAddress != "" {
		return []string{targetAddress}, nil
	}
	if options.ServiceName == "" {
		return nil, rpcerrors.New(rpcerrors.BadRequest, "target address and service name are both empty", nil)
	}

	instances := make([]registry.Instance, 0)
	if options.Registry != nil {
		listed, err := options.Registry.List(options.ServiceName)
		if err != nil {
			return nil, rpcerrors.New(rpcerrors.Unavailable, fmt.Sprintf("registry list failed: %v", err), nil)
		}
		instances = append(instances, listed...)
	}
	if len(instances) == 0 {
		instances = append(instances, options.ServiceInstances[options.ServiceName]...)
	}
	if len(instances) == 0 {
		return nil, rpcerrors.New(rpcerrors.NotFound, fmt.Sprintf("no instance found for service: %s", options.ServiceName), nil)
	}

	addresses := make([]string, 0, len(instances))
	seen := make(map[string]struct{}, len(instances))
	for _, instance := range instances {
		address := strings.TrimSpace(instance.Address)
		if address == "" {
			continue
		}
		if _, ok := seen[address]; ok {
			continue
		}
		seen[address] = struct{}{}
		addresses = append(addresses, address)
	}
	if len(addresses) == 0 {
		return nil, rpcerrors.New(rpcerrors.NotFound, fmt.Sprintf("no valid instance found for service: %s", options.ServiceName), nil)
	}
	return addresses, nil
}

func (client *Client) getConnPool(address string) (*transport.ConnPool, error) {
	if address == "" {
		return nil, rpcerrors.New(rpcerrors.BadRequest, "address is empty", nil)
	}

	client.poolsMu.Lock()
	if pool, ok := client.pools[address]; ok {
		client.poolsMu.Unlock()
		return pool, nil
	}
	client.poolsMu.Unlock()

	pool, err := transport.NewConnPool(
		client.options.PoolMaxConn,
		client.options.PoolMaxIdleConn,
		client.options.PoolIdleTimeout,
		func() (*transport.Conn, error) {
			rawConn, dialErr := net.DialTimeout(client.options.Network, address, client.options.DialTimeout)
			if dialErr != nil {
				return nil, fmt.Errorf("client: dial target %s: %w", address, dialErr)
			}
			conn := transport.NewConn(rawConn, client.onFrame, client.onClose)
			conn.Start()
			return conn, nil
		},
	)
	if err != nil {
		return nil, err
	}

	client.poolsMu.Lock()
	defer client.poolsMu.Unlock()
	if existed, ok := client.pools[address]; ok {
		_ = pool.Close()
		return existed, nil
	}
	client.pools[address] = pool
	return pool, nil
}

func (client *Client) pickAddress(candidates []string, key string) (string, error) {
	if len(candidates) == 0 {
		return "", rpcerrors.New(rpcerrors.NotFound, "no available target address", nil)
	}
	if len(candidates) == 1 {
		return candidates[0], nil
	}

	if client.options.Balancer != nil {
		instances := make([]registry.Instance, 0, len(candidates))
		for _, address := range candidates {
			instances = append(instances, registry.Instance{Address: address})
		}
		picked, err := client.options.Balancer.Pick(instances, key)
		if err == nil {
			address := strings.TrimSpace(picked.Address)
			for _, candidate := range candidates {
				if candidate == address {
					return candidate, nil
				}
			}
		}
	}

	index := client.instanceCursor.Add(1) - 1
	return candidates[index%uint64(len(candidates))], nil
}

func (client *Client) pickAttemptAddresses(method string, maxAttempts int) ([]string, error) {
	if len(client.instanceAddresses) == 0 {
		return nil, rpcerrors.New(rpcerrors.NotFound, "no available target address", nil)
	}

	limit := len(client.instanceAddresses)
	if maxAttempts > 0 && maxAttempts < limit {
		limit = maxAttempts
	}

	remaining := make([]string, len(client.instanceAddresses))
	copy(remaining, client.instanceAddresses)

	picked := make([]string, 0, limit)
	for len(picked) < limit && len(remaining) > 0 {
		address, err := client.pickAddress(remaining, method)
		if err != nil {
			return nil, err
		}
		picked = append(picked, address)
		nextRemaining := make([]string, 0, len(remaining)-1)
		for _, candidate := range remaining {
			if candidate != address {
				nextRemaining = append(nextRemaining, candidate)
			}
		}
		remaining = nextRemaining
	}
	return picked, nil
}

func (client *Client) Close() error {
	client.stateMu.Lock()
	if client.closed {
		client.stateMu.Unlock()
		return nil
	}
	client.closed = true
	client.stateMu.Unlock()

	client.poolsMu.Lock()
	pools := make([]*transport.ConnPool, 0, len(client.pools))
	for _, pool := range client.pools {
		pools = append(pools, pool)
	}
	client.pools = make(map[string]*transport.ConnPool)
	client.poolsMu.Unlock()

	var firstErr error
	for _, pool := range pools {
		if err := pool.Close(); err != nil && firstErr == nil {
			firstErr = err
		}
	}
	return firstErr
}

func (client *Client) Call(ctx context.Context, method string, request any, response any) error {
	return client.Invoke(ctx, method, request, response)
}

func (client *Client) PendingCount() int {
	client.pendingMu.Lock()
	defer client.pendingMu.Unlock()
	return len(client.pending)
}

func (client *Client) isClosed() bool {
	client.stateMu.RLock()
	defer client.stateMu.RUnlock()
	return client.closed
}

func (client *Client) nextRequestID() uint64 {
	return client.requestID.Add(1)
}

func (client *Client) addPending(requestID uint64, resultCh chan pendingResult) {
	client.pendingMu.Lock()
	defer client.pendingMu.Unlock()
	client.pending[requestID] = resultCh
}

func (client *Client) removePending(requestID uint64) {
	client.pendingMu.Lock()
	defer client.pendingMu.Unlock()
	delete(client.pending, requestID)
}

func (client *Client) onFrame(conn *transport.Conn, frame protocol.Frame) {
	if frame.Header.MsgType != protocol.MsgTypeResponse {
		return
	}

	client.pendingMu.Lock()
	resultCh, ok := client.pending[frame.Header.RequestID]
	client.pendingMu.Unlock()
	if !ok {
		return
	}

	select {
	case resultCh <- pendingResult{frame: frame}:
	default:
	}
}

func (client *Client) onClose(conn *transport.Conn, err error) {
	client.pendingMu.Lock()
	pendingSnapshot := make([]chan pendingResult, 0, len(client.pending))
	for requestID, resultCh := range client.pending {
		pendingSnapshot = append(pendingSnapshot, resultCh)
		delete(client.pending, requestID)
	}
	client.pendingMu.Unlock()

	closeErr := err
	if closeErr == nil {
		closeErr = ErrClientClosed
	}
	for _, resultCh := range pendingSnapshot {
		select {
		case resultCh <- pendingResult{err: closeErr}:
		default:
		}
	}
}
