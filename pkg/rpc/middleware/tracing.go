package middleware

import (
	"context"

	"kama_chat_server/pkg/rpc/api"
	"kama_chat_server/pkg/rpc/observability"
	"kama_chat_server/pkg/rpc/protocol"
)

// Tracing 在上下文与请求 metadata 中保证 trace_id 存在并可透传。
func Tracing() UnaryMiddleware {
	return func(next UnaryHandler) UnaryHandler {
		return func(ctx context.Context, request *protocol.Request) (*protocol.Response, error) {
			traceCtx, traceID := observability.EnsureTraceID(ctx)
			if request.Metadata == nil {
				request.Metadata = protocol.Metadata{}
			}
			request.Metadata[api.MetadataKeyTraceID] = traceID
			traceCtx = api.SetTraceID(traceCtx, traceID)
			return next(traceCtx, request)
		}
	}
}
