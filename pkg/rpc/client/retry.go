package client

import (
	"errors"
	"math/rand"
	"time"

	rpcerrors "kama_chat_server/pkg/rpc/errors"
)

func (client *Client) canRetry(method string, request any, invokeErr error, attempt int) bool {
	if attempt >= client.options.RetryMax {
		return false
	}
	if !isRetryableError(invokeErr) {
		return false
	}
	if !client.isIdempotent(method, request) {
		return false
	}
	return true
}

func (client *Client) isIdempotent(method string, request any) bool {
	if _, ok := client.options.IdempotentMethods[method]; ok {
		return true
	}
	if client.options.IdempotentMatcher != nil && client.options.IdempotentMatcher(method, request) {
		return true
	}
	return false
}

func isRetryableError(invokeErr error) bool {
	var rpcErr *rpcerrors.RpcError
	if !errors.As(invokeErr, &rpcErr) || rpcErr == nil {
		return false
	}
	return rpcErr.Code == rpcerrors.Unavailable || rpcErr.Code == rpcerrors.Timeout
}

func (client *Client) nextRetryDelay(attempt int) time.Duration {
	delay := client.options.RetryBaseDelay
	if delay <= 0 {
		delay = 20 * time.Millisecond
	}
	maxDelay := client.options.RetryMaxDelay
	if maxDelay <= 0 {
		maxDelay = 500 * time.Millisecond
	}

	for index := 0; index < attempt; index++ {
		delay *= 2
		if delay >= maxDelay {
			delay = maxDelay
			break
		}
	}

	jitter := client.options.RetryJitter
	if jitter <= 0 {
		return delay
	}
	if jitter > 1 {
		jitter = 1
	}

	factor := 1 + (rand.Float64()*2-1)*jitter
	if factor < 0 {
		factor = 0
	}
	jittered := time.Duration(float64(delay) * factor)
	if jittered <= 0 {
		return time.Millisecond
	}
	return jittered
}

func sleepWithContext(ctxDone <-chan struct{}, delay time.Duration) bool {
	if delay <= 0 {
		return true
	}
	timer := time.NewTimer(delay)
	defer timer.Stop()

	select {
	case <-ctxDone:
		return false
	case <-timer.C:
		return true
	}
}
