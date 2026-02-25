package middleware

import (
	"context"
	"errors"
	"time"

	rpcerrors "kama_chat_server/pkg/rpc/errors"
	"kama_chat_server/pkg/rpc/observability"
	"kama_chat_server/pkg/rpc/protocol"
)

// Logging 记录 method/requestID/trace_id/latency/code。
func Logging(logger observability.Logger) UnaryMiddleware {
	if logger == nil {
		logger = observability.GetDefaultLogger()
	}

	return func(next UnaryHandler) UnaryHandler {
		return func(ctx context.Context, request *protocol.Request) (*protocol.Response, error) {
			start := time.Now()
			response, err := next(ctx, request)
			latencyMs := time.Since(start).Milliseconds()

			code := statusCodeFromResult(response, err)
			traceID, _ := observability.TraceIDFromContext(ctx)
			fields := observability.Fields{
				"method":     request.Method,
				"request_id": request.RequestID,
				"trace_id":   traceID,
				"latency_ms": latencyMs,
				"code":       code,
			}
			if err != nil {
				fields["error"] = err.Error()
				logger.Error("rpc server handled request", fields)
				return response, err
			}
			logger.Info("rpc server handled request", fields)
			return response, nil
		}
	}
}

func statusCodeFromResult(response *protocol.Response, err error) uint16 {
	if err != nil {
		var rpcErr *rpcerrors.RpcError
		if errors.As(err, &rpcErr) {
			return rpcErr.Code.Uint16()
		}
		return rpcerrors.Internal.Uint16()
	}
	if response == nil || response.Status == 0 {
		return protocol.StatusOK
	}
	return response.Status
}
