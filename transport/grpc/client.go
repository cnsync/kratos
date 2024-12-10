package grpc

import (
	"context"
	"crypto/tls"
	"fmt"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	grpcinsecure "google.golang.org/grpc/credentials/insecure"
	grpcmd "google.golang.org/grpc/metadata"

	"github.com/cnsync/kratos/internal/matcher"
	"github.com/cnsync/kratos/log"
	"github.com/cnsync/kratos/middleware"
	"github.com/cnsync/kratos/registry"
	"github.com/cnsync/kratos/selector"
	"github.com/cnsync/kratos/selector/wrr"
	"github.com/cnsync/kratos/transport"
	"github.com/cnsync/kratos/transport/grpc/resolver/discovery"

	// 初始化 resolver
	_ "github.com/cnsync/kratos/transport/grpc/resolver/direct"
)

func init() {
	// 如果全局选择器为空，初始化为 WRR 负载均衡器
	if selector.GlobalSelector() == nil {
		selector.SetGlobalSelector(wrr.NewBuilder())
	}
}

// ClientOption 是 gRPC 客户端的配置选项类型
type ClientOption func(o *clientOptions)

// WithEndpoint 设置客户端的服务端点
func WithEndpoint(endpoint string) ClientOption {
	return func(o *clientOptions) {
		o.endpoint = endpoint
	}
}

// WithSubset 设置客户端的发现子集大小
// 默认为 0，表示不启用子集过滤
func WithSubset(size int) ClientOption {
	return func(o *clientOptions) {
		o.subsetSize = size
	}
}

// WithTimeout 设置客户端的超时时间
func WithTimeout(timeout time.Duration) ClientOption {
	return func(o *clientOptions) {
		o.timeout = timeout
	}
}

// WithMiddleware 设置客户端的中间件
func WithMiddleware(m ...middleware.Middleware) ClientOption {
	return func(o *clientOptions) {
		o.middleware = m
	}
}

// WithDiscovery 设置客户端的服务发现接口
func WithDiscovery(d registry.Discovery) ClientOption {
	return func(o *clientOptions) {
		o.discovery = d
	}
}

// WithTLSConfig 设置客户端的 TLS 配置
func WithTLSConfig(c *tls.Config) ClientOption {
	return func(o *clientOptions) {
		o.tlsConf = c
	}
}

// WithUnaryInterceptor 设置客户端单次 RPC 的拦截器
func WithUnaryInterceptor(in ...grpc.UnaryClientInterceptor) ClientOption {
	return func(o *clientOptions) {
		o.ints = in
	}
}

// WithStreamInterceptor 设置客户端流式 RPC 的拦截器
func WithStreamInterceptor(in ...grpc.StreamClientInterceptor) ClientOption {
	return func(o *clientOptions) {
		o.streamInts = in
	}
}

// WithOptions 设置 gRPC 连接的其他选项
func WithOptions(opts ...grpc.DialOption) ClientOption {
	return func(o *clientOptions) {
		o.grpcOpts = opts
	}
}

// WithNodeFilter 设置节点选择过滤器
func WithNodeFilter(filters ...selector.NodeFilter) ClientOption {
	return func(o *clientOptions) {
		o.filters = filters
	}
}

// WithHealthCheck 设置是否启用健康检查
func WithHealthCheck(healthCheck bool) ClientOption {
	return func(o *clientOptions) {
		if !healthCheck {
			o.healthCheckConfig = ""
		}
	}
}

// WithLogger 设置日志记录器
// Deprecated: 请使用全局日志记录器
func WithLogger(log.Logger) ClientOption {
	return func(*clientOptions) {}
}

// WithPrintDiscoveryDebugLog 设置是否打印服务发现的调试日志
func WithPrintDiscoveryDebugLog(p bool) ClientOption {
	return func(o *clientOptions) {
		o.printDiscoveryDebugLog = p
	}
}

// clientOptions 是 gRPC 客户端的配置选项结构体
type clientOptions struct {
	endpoint               string
	subsetSize             int
	tlsConf                *tls.Config
	timeout                time.Duration
	discovery              registry.Discovery
	middleware             []middleware.Middleware
	streamMiddleware       []middleware.Middleware
	ints                   []grpc.UnaryClientInterceptor
	streamInts             []grpc.StreamClientInterceptor
	grpcOpts               []grpc.DialOption
	balancerName           string
	filters                []selector.NodeFilter
	healthCheckConfig      string
	printDiscoveryDebugLog bool
}

// Dial 返回一个 gRPC 连接
func Dial(ctx context.Context, opts ...ClientOption) (*grpc.ClientConn, error) {
	return dial(ctx, false, opts...)
}

// DialInsecure 返回一个不安全的 gRPC 连接
func DialInsecure(ctx context.Context, opts ...ClientOption) (*grpc.ClientConn, error) {
	return dial(ctx, true, opts...)
}

func dial(ctx context.Context, insecure bool, opts ...ClientOption) (*grpc.ClientConn, error) {
	// 初始化默认的客户端配置
	options := clientOptions{
		timeout:                2000 * time.Millisecond,
		balancerName:           balancerName,
		subsetSize:             25,
		printDiscoveryDebugLog: true,
		healthCheckConfig:      `,"healthCheckConfig":{"serviceName":""}`,
	}

	// 应用客户端配置选项
	for _, o := range opts {
		o(&options)
	}

	// 设置单次 RPC 的拦截器
	ints := []grpc.UnaryClientInterceptor{
		unaryClientInterceptor(options.middleware, options.timeout, options.filters),
	}

	// 设置流式 RPC 的拦截器
	sints := []grpc.StreamClientInterceptor{
		streamClientInterceptor(options.streamMiddleware, options.filters),
	}

	// 添加用户自定义的拦截器
	if len(options.ints) > 0 {
		ints = append(ints, options.ints...)
	}
	if len(options.streamInts) > 0 {
		sints = append(sints, options.streamInts...)
	}

	// 配置 gRPC 连接选项
	grpcOpts := []grpc.DialOption{
		grpc.WithDefaultServiceConfig(fmt.Sprintf(`{"loadBalancingConfig": [{"%s":{}}]%s}`,
			options.balancerName, options.healthCheckConfig)),
		grpc.WithChainUnaryInterceptor(ints...),
		grpc.WithChainStreamInterceptor(sints...),
	}

	// 如果启用了服务发现，则添加解析器选项
	if options.discovery != nil {
		grpcOpts = append(grpcOpts,
			grpc.WithResolvers(
				discovery.NewBuilder(
					options.discovery,
					discovery.WithInsecure(insecure),
					discovery.WithTimeout(options.timeout),
					discovery.WithSubset(options.subsetSize),
					discovery.PrintDebugLog(options.printDiscoveryDebugLog),
				)))
	}

	// 如果是非安全连接，使用不安全的凭证
	if insecure {
		grpcOpts = append(grpcOpts, grpc.WithTransportCredentials(grpcinsecure.NewCredentials()))
	}

	// 如果设置了 TLS 配置，使用加密连接
	if options.tlsConf != nil {
		grpcOpts = append(grpcOpts, grpc.WithTransportCredentials(credentials.NewTLS(options.tlsConf)))
	}

	// 添加用户自定义的 gRPC 连接选项
	if len(options.grpcOpts) > 0 {
		grpcOpts = append(grpcOpts, options.grpcOpts...)
	}

	// 使用配置选项建立 gRPC 连接
	return grpc.DialContext(ctx, options.endpoint, grpcOpts...)
}

func unaryClientInterceptor(ms []middleware.Middleware, timeout time.Duration, filters []selector.NodeFilter) grpc.UnaryClientInterceptor {
	return func(ctx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
		// 为每个 RPC 请求创建新的上下文
		ctx = transport.NewClientContext(ctx, &Transport{
			endpoint:    cc.Target(),
			operation:   method,
			reqHeader:   headerCarrier{},
			nodeFilters: filters,
		})

		// 设置超时
		if timeout > 0 {
			var cancel context.CancelFunc
			ctx, cancel = context.WithTimeout(ctx, timeout)
			defer cancel()
		}

		// 处理 RPC 调用
		h := func(ctx context.Context, req interface{}) (interface{}, error) {
			if tr, ok := transport.FromClientContext(ctx); ok {
				header := tr.RequestHeader()
				keys := header.Keys()
				keyvals := make([]string, 0, len(keys))
				for _, k := range keys {
					keyvals = append(keyvals, k, header.Get(k))
				}
				// 将请求头添加到 gRPC 上下文中
				ctx = grpcmd.AppendToOutgoingContext(ctx, keyvals...)
			}
			return reply, invoker(ctx, method, req, reply, cc, opts...)
		}

		// 应用中间件链
		if len(ms) > 0 {
			h = middleware.Chain(ms...)(h)
		}

		var p selector.Peer
		ctx = selector.NewPeerContext(ctx, &p)
		_, err := h(ctx, req)
		return err
	}
}

// wrappedClientStream 包装了 grpc.ClientStream，并应用中间件
type wrappedClientStream struct {
	grpc.ClientStream
	ctx        context.Context
	middleware matcher.Matcher
}

// Context 返回包装后的流的上下文
func (w *wrappedClientStream) Context() context.Context {
	return w.ctx
}

// SendMsg 发送消息，应用中间件
func (w *wrappedClientStream) SendMsg(m interface{}) error {
	h := func(_ context.Context, req interface{}) (interface{}, error) {
		return req, w.ClientStream.SendMsg(m)
	}

	info, ok := transport.FromClientContext(w.ctx)
	if !ok {
		return fmt.Errorf("transport value stored in ctx returns: %v", ok)
	}

	// 如果有匹配的中间件，应用它们
	if next := w.middleware.Match(info.Operation()); len(next) > 0 {
		h = middleware.Chain(next...)(h)
	}

	_, err := h(w.ctx, m)
	return err
}

// RecvMsg 接收消息，应用中间件
func (w *wrappedClientStream) RecvMsg(m interface{}) error {
	h := func(_ context.Context, req interface{}) (interface{}, error) {
		return req, w.ClientStream.RecvMsg(m)
	}

	info, ok := transport.FromClientContext(w.ctx)
	if !ok {
		return fmt.Errorf("transport value stored in ctx returns: %v", ok)
	}

	// 如果有匹配的中间件，应用它们
	if next := w.middleware.Match(info.Operation()); len(next) > 0 {
		h = middleware.Chain(next...)(h)
	}

	_, err := h(w.ctx, m)
	return err
}

// streamClientInterceptor 为流式 RPC 设置拦截器，并应用中间件
func streamClientInterceptor(ms []middleware.Middleware, filters []selector.NodeFilter) grpc.StreamClientInterceptor {
	return func(ctx context.Context, desc *grpc.StreamDesc, cc *grpc.ClientConn, method string, streamer grpc.Streamer, opts ...grpc.CallOption) (grpc.ClientStream, error) {
		// 为每个流式 RPC 请求创建新的上下文
		ctx = transport.NewClientContext(ctx, &Transport{
			endpoint:    cc.Target(),
			operation:   method,
			reqHeader:   headerCarrier{},
			nodeFilters: filters,
		})

		var p selector.Peer
		ctx = selector.NewPeerContext(ctx, &p)

		// 创建流式 RPC 流
		clientStream, err := streamer(ctx, desc, cc, method, opts...)
		if err != nil {
			return nil, err
		}

		h := func(_ context.Context, _ interface{}) (interface{}, error) {
			return streamer, nil
		}

		m := matcher.New()
		// 如果有中间件，应用它们
		if len(ms) > 0 {
			m.Use(ms...)
			middleware.Chain(ms...)(h)
		}

		// 返回包装后的流
		wrappedStream := &wrappedClientStream{
			ClientStream: clientStream,
			ctx:          ctx,
			middleware:   m,
		}

		return wrappedStream, nil
	}
}
