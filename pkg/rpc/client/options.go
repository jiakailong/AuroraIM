package client

import (
	"time"

	"kama_chat_server/pkg/rpc/balancer"
	"kama_chat_server/pkg/rpc/protocol"
	"kama_chat_server/pkg/rpc/registry"
)

type Options struct {
	Timeout     time.Duration
	DialTimeout time.Duration
	Network     string
	CodecID     uint8
	RetryMax    int

	RetryBaseDelay time.Duration
	RetryMaxDelay  time.Duration
	RetryJitter    float64

	Registry registry.Registry
	Balancer balancer.Balancer

	ServiceName      string
	ServiceInstances map[string][]registry.Instance

	IdempotentMethods map[string]struct{}
	IdempotentMatcher func(method string, request any) bool

	PoolMaxConn     int
	PoolMaxIdleConn int
	PoolIdleTimeout time.Duration
}

type Option func(*Options)

func defaultOptions() Options {
	return Options{
		Timeout:           3 * time.Second,
		DialTimeout:       3 * time.Second,
		Network:           "tcp",
		CodecID:           protocol.DefaultCodecID,
		RetryMax:          0,
		RetryBaseDelay:    20 * time.Millisecond,
		RetryMaxDelay:     500 * time.Millisecond,
		RetryJitter:       0.2,
		Balancer:          balancer.NewRoundRobin(),
		ServiceInstances:  make(map[string][]registry.Instance),
		IdempotentMethods: make(map[string]struct{}),

		PoolMaxConn:     1,
		PoolMaxIdleConn: 1,
		PoolIdleTimeout: 60 * time.Second,
	}
}

func WithTimeout(timeout time.Duration) Option {
	return func(options *Options) {
		if timeout > 0 {
			options.Timeout = timeout
		}
	}
}

func WithDialTimeout(timeout time.Duration) Option {
	return func(options *Options) {
		if timeout > 0 {
			options.DialTimeout = timeout
		}
	}
}

func WithNetwork(network string) Option {
	return func(options *Options) {
		if network != "" {
			options.Network = network
		}
	}
}

func WithCodecID(codecID uint8) Option {
	return func(options *Options) {
		options.CodecID = codecID
	}
}

func WithRetryMax(retryMax int) Option {
	return func(options *Options) {
		if retryMax >= 0 {
			options.RetryMax = retryMax
		}
	}
}

func WithRetryBackoff(baseDelay time.Duration, maxDelay time.Duration, jitter float64) Option {
	return func(options *Options) {
		if baseDelay > 0 {
			options.RetryBaseDelay = baseDelay
		}
		if maxDelay > 0 {
			options.RetryMaxDelay = maxDelay
		}
		if jitter >= 0 && jitter <= 1 {
			options.RetryJitter = jitter
		}
	}
}

func WithRegistry(reg registry.Registry) Option {
	return func(options *Options) {
		options.Registry = reg
	}
}

func WithBalancer(loadBalancer balancer.Balancer) Option {
	return func(options *Options) {
		if loadBalancer != nil {
			options.Balancer = loadBalancer
		}
	}
}

func WithServiceName(serviceName string) Option {
	return func(options *Options) {
		if serviceName != "" {
			options.ServiceName = serviceName
		}
	}
}

func WithServiceInstances(serviceInstances map[string][]registry.Instance) Option {
	return func(options *Options) {
		if len(serviceInstances) == 0 {
			return
		}
		if options.ServiceInstances == nil {
			options.ServiceInstances = make(map[string][]registry.Instance)
		}
		for serviceName, instances := range serviceInstances {
			if serviceName == "" {
				continue
			}
			copied := make([]registry.Instance, len(instances))
			copy(copied, instances)
			options.ServiceInstances[serviceName] = copied
		}
	}
}

func WithIdempotentMethod(method string) Option {
	return func(options *Options) {
		if method == "" {
			return
		}
		if options.IdempotentMethods == nil {
			options.IdempotentMethods = make(map[string]struct{})
		}
		options.IdempotentMethods[method] = struct{}{}
	}
}

func WithIdempotentMethods(methods []string) Option {
	return func(options *Options) {
		if options.IdempotentMethods == nil {
			options.IdempotentMethods = make(map[string]struct{})
		}
		for _, method := range methods {
			if method == "" {
				continue
			}
			options.IdempotentMethods[method] = struct{}{}
		}
	}
}

func WithIdempotentMatcher(matcher func(method string, request any) bool) Option {
	return func(options *Options) {
		options.IdempotentMatcher = matcher
	}
}

func WithPoolMaxConn(maxConn int) Option {
	return func(options *Options) {
		if maxConn > 0 {
			options.PoolMaxConn = maxConn
		}
	}
}

func WithPoolMaxIdleConn(maxIdleConn int) Option {
	return func(options *Options) {
		if maxIdleConn >= 0 {
			options.PoolMaxIdleConn = maxIdleConn
		}
	}
}

func WithPoolIdleTimeout(idleTimeout time.Duration) Option {
	return func(options *Options) {
		if idleTimeout >= 0 {
			options.PoolIdleTimeout = idleTimeout
		}
	}
}
