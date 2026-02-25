package middleware

import (
	"context"

	"kama_chat_server/pkg/rpc/protocol"
)

// UnaryHandler 表示一次 RPC 请求处理函数。
type UnaryHandler func(ctx context.Context, request *protocol.Request) (*protocol.Response, error)

// UnaryMiddleware 表示对 UnaryHandler 的包装函数。
type UnaryMiddleware func(next UnaryHandler) UnaryHandler

// Chain 将多个中间件按顺序组合为一个中间件。
func Chain(middlewares ...UnaryMiddleware) UnaryMiddleware {
	return func(finalHandler UnaryHandler) UnaryHandler {
		wrapped := finalHandler
		for index := len(middlewares) - 1; index >= 0; index-- {
			wrapped = middlewares[index](wrapped)
		}
		return wrapped
	}
}
