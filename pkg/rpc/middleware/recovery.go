package middleware

import (
	"context"
	"fmt"

	rpcerrors "kama_chat_server/pkg/rpc/errors"
	"kama_chat_server/pkg/rpc/protocol"
)

// Recovery 捕获 panic 并转为 Internal 错误。
func Recovery() UnaryMiddleware {
	return func(next UnaryHandler) UnaryHandler {
		return func(ctx context.Context, request *protocol.Request) (response *protocol.Response, err error) {
			defer func() {
				if panicValue := recover(); panicValue != nil {
					err = rpcerrors.New(rpcerrors.Internal, fmt.Sprintf("handler panic: %v", panicValue), nil)
				}
			}()
			return next(ctx, request)
		}
	}
}

// Recover 统一处理 handler panic，返回 Internal 错误。
func Recover(ctx context.Context, request *protocol.Request, handler func(context.Context, *protocol.Request) (*protocol.Response, error)) (*protocol.Response, error) {
	return Recovery()(handler)(ctx, request)
}
