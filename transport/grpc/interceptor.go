package grpc

import (
	"context"
	"fmt"

	"google.golang.org/grpc"
	grpcmd "google.golang.org/grpc/metadata"

	ic "github.com/cnsync/kratos/internal/context"
	"github.com/cnsync/kratos/internal/matcher"
	"github.com/cnsync/kratos/middleware"
	"github.com/cnsync/kratos/transport"
)

// unaryServerInterceptor 是一个 gRPC 的单次 RPC 拦截器
func (s *Server) unaryServerInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		// 合并用户的上下文和基本上下文
		ctx, cancel := ic.Merge(ctx, s.baseCtx)
		defer cancel()

		// 获取请求的元数据
		md, _ := grpcmd.FromIncomingContext(ctx)

		// 创建一个用于传输的 Transport 对象，包含请求和响应的元数据
		replyHeader := grpcmd.MD{}
		tr := &Transport{
			operation:   info.FullMethod,
			reqHeader:   headerCarrier(md),
			replyHeader: headerCarrier(replyHeader),
		}

		// 如果有端点信息，设置它
		if s.endpoint != nil {
			tr.endpoint = s.endpoint.String()
		}

		// 将 transport 信息存入上下文中
		ctx = transport.NewServerContext(ctx, tr)

		// 如果有超时限制，设置超时
		if s.timeout > 0 {
			ctx, cancel = context.WithTimeout(ctx, s.timeout)
			defer cancel()
		}

		// 定义请求处理函数
		h := func(ctx context.Context, req interface{}) (interface{}, error) {
			return handler(ctx, req)
		}

		// 如果有中间件匹配当前操作，链式调用中间件
		if next := s.middleware.Match(tr.Operation()); len(next) > 0 {
			h = middleware.Chain(next...)(h)
		}

		// 调用处理函数并返回结果
		reply, err := h(ctx, req)

		// 如果有回复头信息，设置它
		if len(replyHeader) > 0 {
			_ = grpc.SetHeader(ctx, replyHeader)
		}
		return reply, err
	}
}

// wrappedStream 用于包装 gRPC 流式请求的上下文
type wrappedStream struct {
	grpc.ServerStream
	ctx        context.Context
	middleware matcher.Matcher
}

// NewWrappedStream 创建一个新的 wrappedStream 实例
func NewWrappedStream(ctx context.Context, stream grpc.ServerStream, m matcher.Matcher) grpc.ServerStream {
	return &wrappedStream{
		ServerStream: stream,
		ctx:          ctx,
		middleware:   m,
	}
}

// Context 获取包装后的流的上下文
func (w *wrappedStream) Context() context.Context {
	return w.ctx
}

// streamServerInterceptor 是一个 gRPC 的流式 RPC 拦截器
func (s *Server) streamServerInterceptor() grpc.StreamServerInterceptor {
	return func(srv interface{}, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
		// 合并用户的上下文和基本上下文
		ctx, cancel := ic.Merge(ss.Context(), s.baseCtx)
		defer cancel()

		// 获取请求的元数据
		md, _ := grpcmd.FromIncomingContext(ctx)

		// 创建一个用于传输的 Transport 对象，包含请求和响应的元数据
		replyHeader := grpcmd.MD{}
		ctx = transport.NewServerContext(ctx, &Transport{
			endpoint:    s.endpoint.String(),
			operation:   info.FullMethod,
			reqHeader:   headerCarrier(md),
			replyHeader: headerCarrier(replyHeader),
		})

		// 定义流式请求的处理函数
		h := func(_ context.Context, _ interface{}) (interface{}, error) {
			return handler(srv, ss), nil
		}

		// 如果有中间件匹配当前流操作，链式调用中间件
		if next := s.streamMiddleware.Match(info.FullMethod); len(next) > 0 {
			middleware.Chain(next...)(h)
		}

		// 将流的上下文存入 context 中
		ctx = context.WithValue(ctx, stream{
			ServerStream:     ss,
			streamMiddleware: s.streamMiddleware,
		}, ss)

		// 创建包装后的流实例
		ws := NewWrappedStream(ctx, ss, s.streamMiddleware)

		// 调用处理函数并返回结果
		err := handler(srv, ws)

		// 如果有回复头信息，设置它
		if len(replyHeader) > 0 {
			_ = grpc.SetHeader(ctx, replyHeader)
		}
		return err
	}
}

// stream 类型用于存储流的中间件信息
type stream struct {
	grpc.ServerStream
	streamMiddleware matcher.Matcher
}

// GetStream 从上下文中获取流实例
func GetStream(ctx context.Context) grpc.ServerStream {
	return ctx.Value(stream{}).(grpc.ServerStream)
}

// SendMsg 通过包装的流发送消息，支持中间件链式处理
func (w *wrappedStream) SendMsg(m interface{}) error {
	h := func(_ context.Context, req interface{}) (interface{}, error) {
		return req, w.ServerStream.SendMsg(m)
	}

	// 从上下文中获取 transport 信息
	info, ok := transport.FromServerContext(w.ctx)
	if !ok {
		return fmt.Errorf("transport value stored in ctx returns: %v", ok)
	}

	// 如果有中间件匹配当前操作，链式调用中间件
	if next := w.middleware.Match(info.Operation()); len(next) > 0 {
		h = middleware.Chain(next...)(h)
	}

	// 调用处理函数并返回结果
	_, err := h(w.ctx, m)
	return err
}

// RecvMsg 通过包装的流接收消息，支持中间件链式处理
func (w *wrappedStream) RecvMsg(m interface{}) error {
	h := func(_ context.Context, req interface{}) (interface{}, error) {
		return req, w.ServerStream.RecvMsg(m)
	}

	// 从上下文中获取 transport 信息
	info, ok := transport.FromServerContext(w.ctx)
	if !ok {
		return fmt.Errorf("transport value stored in ctx returns: %v", ok)
	}

	// 如果有中间件匹配当前操作，链式调用中间件
	if next := w.middleware.Match(info.Operation()); len(next) > 0 {
		h = middleware.Chain(next...)(h)
	}

	// 调用处理函数并返回结果
	_, err := h(w.ctx, m)
	return err
}
