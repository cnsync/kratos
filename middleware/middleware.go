package middleware

import (
	"context"
)

// Handler 定义了中间件调用的处理程序。
type Handler func(ctx context.Context, req interface{}) (interface{}, error)

// Middleware 是 HTTP/gRPC 传输中间件。
type Middleware func(Handler) Handler

// Chain 返回一个中间件，该中间件指定了端点的链式处理程序。
func Chain(m ...Middleware) Middleware {
	return func(next Handler) Handler {
		// 遍历传入的中间件列表，从最后一个开始
		for i := len(m) - 1; i >= 0; i-- {
			// 将当前中间件应用于下一个处理程序，形成链式调用
			next = m[i](next)
		}
		return next
	}
}
