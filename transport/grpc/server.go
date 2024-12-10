package grpc

import (
	"context"
	"crypto/tls"
	"net"
	"net/url"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/admin"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/health"
	"google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/reflection"

	apimd "github.com/cnsync/kratos/api/metadata"
	"github.com/cnsync/kratos/internal/endpoint"
	"github.com/cnsync/kratos/internal/host"
	"github.com/cnsync/kratos/internal/matcher"
	"github.com/cnsync/kratos/log"
	"github.com/cnsync/kratos/middleware"
	"github.com/cnsync/kratos/transport"
)

var (
	_ transport.Server           = (*Server)(nil)
	_ transport.EndpointProvider = (*Server)(nil)
)

// ServerOption 是 gRPC 服务器配置的选项类型
type ServerOption func(o *Server)

// Network 设置服务器的网络类型（如 tcp）
func Network(network string) ServerOption {
	return func(s *Server) {
		s.network = network
	}
}

// Address 设置服务器的监听地址
func Address(addr string) ServerOption {
	return func(s *Server) {
		s.address = addr
	}
}

// Endpoint 设置服务器的端点（URL）
func Endpoint(endpoint *url.URL) ServerOption {
	return func(s *Server) {
		s.endpoint = endpoint
	}
}

// Timeout 设置服务器的超时时间
func Timeout(timeout time.Duration) ServerOption {
	return func(s *Server) {
		s.timeout = timeout
	}
}

// Logger 设置服务器的日志记录器
// Deprecated: 请使用全局日志记录器
func Logger(log.Logger) ServerOption {
	return func(*Server) {}
}

// Middleware 设置服务器的中间件
func Middleware(m ...middleware.Middleware) ServerOption {
	return func(s *Server) {
		s.middleware.Use(m...)
	}
}

// StreamMiddleware 设置服务器的流式中间件
func StreamMiddleware(m ...middleware.Middleware) ServerOption {
	return func(s *Server) {
		s.streamMiddleware.Use(m...)
	}
}

// CustomHealth 启用自定义健康检查
func CustomHealth() ServerOption {
	return func(s *Server) {
		s.customHealth = true
	}
}

// TLSConfig 设置 TLS 配置
func TLSConfig(c *tls.Config) ServerOption {
	return func(s *Server) {
		s.tlsConf = c
	}
}

// Listener 设置服务器的监听器
func Listener(lis net.Listener) ServerOption {
	return func(s *Server) {
		s.lis = lis
	}
}

// UnaryInterceptor 设置服务器的单次 RPC 拦截器
func UnaryInterceptor(in ...grpc.UnaryServerInterceptor) ServerOption {
	return func(s *Server) {
		s.unaryInts = in
	}
}

// StreamInterceptor 设置服务器的流式 RPC 拦截器
func StreamInterceptor(in ...grpc.StreamServerInterceptor) ServerOption {
	return func(s *Server) {
		s.streamInts = in
	}
}

// Options 设置 gRPC 连接的其他选项
func Options(opts ...grpc.ServerOption) ServerOption {
	return func(s *Server) {
		s.grpcOpts = opts
	}
}

// Server 是一个 gRPC 服务器包装器
type Server struct {
	*grpc.Server
	baseCtx          context.Context
	tlsConf          *tls.Config
	lis              net.Listener
	err              error
	network          string
	address          string
	endpoint         *url.URL
	timeout          time.Duration
	middleware       matcher.Matcher
	streamMiddleware matcher.Matcher
	unaryInts        []grpc.UnaryServerInterceptor
	streamInts       []grpc.StreamServerInterceptor
	grpcOpts         []grpc.ServerOption
	health           *health.Server
	customHealth     bool
	metadata         *apimd.Server
	adminClean       func()
}

// NewServer 创建一个 gRPC 服务器，并应用给定的选项
func NewServer(opts ...ServerOption) *Server {
	srv := &Server{
		baseCtx:          context.Background(),
		network:          "tcp", // 默认使用 TCP 网络
		address:          ":0",  // 默认监听地址为 :0
		timeout:          1 * time.Second,
		health:           health.NewServer(),
		middleware:       matcher.New(),
		streamMiddleware: matcher.New(),
	}
	// 应用给定的选项
	for _, o := range opts {
		o(srv)
	}

	// 配置默认的 RPC 拦截器
	unaryInts := []grpc.UnaryServerInterceptor{
		srv.unaryServerInterceptor(),
	}
	streamInts := []grpc.StreamServerInterceptor{
		srv.streamServerInterceptor(),
	}

	// 合并用户设置的拦截器
	if len(srv.unaryInts) > 0 {
		unaryInts = append(unaryInts, srv.unaryInts...)
	}
	if len(srv.streamInts) > 0 {
		streamInts = append(streamInts, srv.streamInts...)
	}

	// 配置 gRPC 服务器选项
	grpcOpts := []grpc.ServerOption{
		grpc.ChainUnaryInterceptor(unaryInts...),
		grpc.ChainStreamInterceptor(streamInts...),
	}

	// 如果启用了 TLS，添加 TLS 认证
	if srv.tlsConf != nil {
		grpcOpts = append(grpcOpts, grpc.Creds(credentials.NewTLS(srv.tlsConf)))
	}

	// 添加用户自定义的 gRPC 选项
	if len(srv.grpcOpts) > 0 {
		grpcOpts = append(grpcOpts, srv.grpcOpts...)
	}

	// 创建 gRPC 服务器实例
	srv.Server = grpc.NewServer(grpcOpts...)
	srv.metadata = apimd.NewServer(srv.Server)

	// 注册内部服务
	if !srv.customHealth {
		grpc_health_v1.RegisterHealthServer(srv.Server, srv.health)
	}
	apimd.RegisterMetadataServer(srv.Server, srv.metadata)
	reflection.Register(srv.Server)

	// 注册管理员接口
	srv.adminClean, _ = admin.Register(srv.Server)

	return srv
}

// Use 使用服务中间件，并指定选择器
// 选择器可以是：
//   - '/*'
//   - '/helloworld.v1.Greeter/*'
//   - '/helloworld.v1.Greeter/SayHello'
func (s *Server) Use(selector string, m ...middleware.Middleware) {
	s.middleware.Add(selector, m...)
}

// Endpoint 返回真实的服务端点地址
// 示例：
//
//	grpc://127.0.0.1:9000?isSecure=false
func (s *Server) Endpoint() (*url.URL, error) {
	if err := s.listenAndEndpoint(); err != nil {
		return nil, s.err
	}
	return s.endpoint, nil
}

// Start 启动 gRPC 服务器
func (s *Server) Start(ctx context.Context) error {
	if err := s.listenAndEndpoint(); err != nil {
		return s.err
	}
	s.baseCtx = ctx
	log.Infof("[gRPC] server listening on: %s", s.lis.Addr().String())
	s.health.Resume()
	return s.Serve(s.lis)
}

// Stop 停止 gRPC 服务器
func (s *Server) Stop(_ context.Context) error {
	if s.adminClean != nil {
		s.adminClean()
	}
	s.health.Shutdown()
	s.GracefulStop()
	log.Info("[gRPC] server stopping")
	return nil
}

// listenAndEndpoint 启动监听并设置服务端点
func (s *Server) listenAndEndpoint() error {
	if s.lis == nil {
		// 如果没有提供监听器，默认使用网络和地址创建
		lis, err := net.Listen(s.network, s.address)
		if err != nil {
			s.err = err
			return err
		}
		s.lis = lis
	}
	if s.endpoint == nil {
		// 如果没有提供服务端点，自动提取服务地址
		addr, err := host.Extract(s.address, s.lis)
		if err != nil {
			s.err = err
			return err
		}
		s.endpoint = endpoint.NewEndpoint(endpoint.Scheme("grpc", s.tlsConf != nil), addr)
	}
	return s.err
}
