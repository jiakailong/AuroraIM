package middleware

import (
	"context"
	"time"

	"kama_chat_server/pkg/rpc/observability"
	"kama_chat_server/pkg/rpc/protocol"
)

func Metrics(metrics observability.Metrics) UnaryMiddleware {
	if metrics == nil {
		metrics = observability.GetDefaultMetrics()
	}
	return func(next UnaryHandler) UnaryHandler {
		return func(ctx context.Context, request *protocol.Request) (*protocol.Response, error) {
			method := request.Method
			metrics.IncInFlight(method)
			start := time.Now()
			response, err := next(ctx, request)
			latency := time.Since(start)
			code := statusCodeFromResult(response, err)
			metrics.Observe(method, code, latency)
			metrics.DecInFlight(method)
			return response, err
		}
	}
}
