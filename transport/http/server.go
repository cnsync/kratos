package http

import (
	"context"
	"crypto/tls"
	"errors"
	"net"
	"net/http"
	"net/url"
	"time"

	"github.com/gorilla/mux"

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
	_ http.Handler               = (*Server)(nil)
)

// ServerOption 是用于配置 HTTP 服务器的选项。
type ServerOption func(*Server)

// Network 配置服务器的网络类型（如 TCP、UDP）。
func Network(network string) ServerOption {
	return func(s *Server) {
		s.network = network
	}
}

// Address 配置服务器的监听地址。
func Address(addr string) ServerOption {
	return func(s *Server) {
		s.address = addr
	}
}

// Endpoint 配置服务器的端点（URL）。
func Endpoint(endpoint *url.URL) ServerOption {
	return func(s *Server) {
		s.endpoint = endpoint
	}
}

// Timeout 配置服务器的超时时间。
func Timeout(timeout time.Duration) ServerOption {
	return func(s *Server) {
		s.timeout = timeout
	}
}

// Logger 配置服务器的日志记录器。
// Deprecated: 使用全局日志记录器。
func Logger(log.Logger) ServerOption {
	return func(*Server) {}
}

// Middleware 配置服务中间件。
func Middleware(m ...middleware.Middleware) ServerOption {
	return func(o *Server) {
		o.middleware.Use(m...)
	}
}

// Filter 配置 HTTP 中间件。
func Filter(filters ...FilterFunc) ServerOption {
	return func(o *Server) {
		o.filters = filters
	}
}

// RequestVarsDecoder 配置请求参数解码器。
func RequestVarsDecoder(dec DecodeRequestFunc) ServerOption {
	return func(o *Server) {
		o.decVars = dec
	}
}

// RequestQueryDecoder 配置查询参数解码器。
func RequestQueryDecoder(dec DecodeRequestFunc) ServerOption {
	return func(o *Server) {
		o.decQuery = dec
	}
}

// RequestDecoder 配置请求体解码器。
func RequestDecoder(dec DecodeRequestFunc) ServerOption {
	return func(o *Server) {
		o.decBody = dec
	}
}

// ResponseEncoder 配置响应编码器。
func ResponseEncoder(en EncodeResponseFunc) ServerOption {
	return func(o *Server) {
		o.enc = en
	}
}

// ErrorEncoder 配置错误编码器。
func ErrorEncoder(en EncodeErrorFunc) ServerOption {
	return func(o *Server) {
		o.ene = en
	}
}

// TLSConfig 配置 TLS 配置。
func TLSConfig(c *tls.Config) ServerOption {
	return func(o *Server) {
		o.tlsConf = c
	}
}

// StrictSlash 配置 mux 的 StrictSlash。
// 如果为 true，当访问 "/path" 时，自动重定向到 "/path/"，反之亦然。
func StrictSlash(strictSlash bool) ServerOption {
	return func(o *Server) {
		o.strictSlash = strictSlash
	}
}

// Listener 配置服务器的监听器。
func Listener(lis net.Listener) ServerOption {
	return func(s *Server) {
		s.lis = lis
	}
}

// PathPrefix 配置 mux 的 PathPrefix。
// 路由器会被替换为一个以 prefix 为前缀的子路由。
func PathPrefix(prefix string) ServerOption {
	return func(s *Server) {
		s.router = s.router.PathPrefix(prefix).Subrouter()
	}
}

// NotFoundHandler 配置 404 请求的处理器。
func NotFoundHandler(handler http.Handler) ServerOption {
	return func(s *Server) {
		s.router.NotFoundHandler = handler
	}
}

// MethodNotAllowedHandler 配置 405 请求的处理器。
func MethodNotAllowedHandler(handler http.Handler) ServerOption {
	return func(s *Server) {
		s.router.MethodNotAllowedHandler = handler
	}
}

// Server 是 HTTP 服务器的封装，提供了更灵活的配置和中间件支持。
type Server struct {
	*http.Server
	lis         net.Listener       // 网络监听器
	tlsConf     *tls.Config        // TLS 配置
	endpoint    *url.URL           // 服务器的端点 URL
	err         error              // 错误信息
	network     string             // 网络类型（TCP、UDP）
	address     string             // 服务器地址
	timeout     time.Duration      // 请求超时
	filters     []FilterFunc       // 过滤器（中间件）
	middleware  matcher.Matcher    // 中间件匹配器
	decVars     DecodeRequestFunc  // 请求变量解码器
	decQuery    DecodeRequestFunc  // 查询参数解码器
	decBody     DecodeRequestFunc  // 请求体解码器
	enc         EncodeResponseFunc // 响应编码器
	ene         EncodeErrorFunc    // 错误编码器
	strictSlash bool               // 是否启用严格斜杠
	router      *mux.Router        // 路由器
}

// NewServer 创建一个新的 HTTP 服务器，接受配置选项。
func NewServer(opts ...ServerOption) *Server {
	srv := &Server{
		network:     "tcp",
		address:     ":0",
		timeout:     1 * time.Second,
		middleware:  matcher.New(),
		decVars:     DefaultRequestVars,
		decQuery:    DefaultRequestQuery,
		decBody:     DefaultRequestDecoder,
		enc:         DefaultResponseEncoder,
		ene:         DefaultErrorEncoder,
		strictSlash: true,
		router:      mux.NewRouter(),
	}
	srv.router.NotFoundHandler = http.DefaultServeMux
	srv.router.MethodNotAllowedHandler = http.DefaultServeMux
	// 应用配置选项
	for _, o := range opts {
		o(srv)
	}
	// 启用严格斜杠选项
	srv.router.StrictSlash(srv.strictSlash)
	// 添加中间件
	srv.router.Use(srv.filter())
	// 创建 HTTP 服务器
	srv.Server = &http.Server{
		Handler:   FilterChain(srv.filters...)(srv.router),
		TLSConfig: srv.tlsConf,
	}
	return srv
}

// Use 添加服务中间件，并使用选择器进行匹配。
// 选择器可以是具体的路径或 API 方法。
func (s *Server) Use(selector string, m ...middleware.Middleware) {
	s.middleware.Add(selector, m...)
}

// WalkRoute 遍历路由器及其子路由，调用提供的回调函数处理每个路由。
func (s *Server) WalkRoute(fn WalkRouteFunc) error {
	return s.router.Walk(func(route *mux.Route, _ *mux.Router, _ []*mux.Route) error {
		methods, err := route.GetMethods()
		if err != nil {
			return nil // 忽略没有方法的路由
		}
		path, err := route.GetPathTemplate()
		if err != nil {
			return err
		}
		for _, method := range methods {
			if err := fn(RouteInfo{Method: method, Path: path}); err != nil {
				return err
			}
		}
		return nil
	})
}

// WalkHandle 遍历路由器及其子路由，调用提供的回调函数处理每个路由的处理器。
func (s *Server) WalkHandle(handle func(method, path string, handler http.HandlerFunc)) error {
	return s.WalkRoute(func(r RouteInfo) error {
		handle(r.Method, r.Path, s.ServeHTTP)
		return nil
	})
}

// Route 注册一个 HTTP 路由。
func (s *Server) Route(prefix string, filters ...FilterFunc) *Router {
	return newRouter(prefix, s, filters...)
}

// Handle 注册一个新路由。
func (s *Server) Handle(path string, h http.Handler) {
	s.router.Handle(path, h)
}

// HandlePrefix 注册一个带有路径前缀的路由。
func (s *Server) HandlePrefix(prefix string, h http.Handler) {
	s.router.PathPrefix(prefix).Handler(h)
}

// HandleFunc 注册一个带有路径匹配的路由，处理器为 http.HandlerFunc。
func (s *Server) HandleFunc(path string, h http.HandlerFunc) {
	s.router.HandleFunc(path, h)
}

// HandleHeader 根据请求头注册路由。
func (s *Server) HandleHeader(key, val string, h http.HandlerFunc) {
	s.router.Headers(key, val).Handler(h)
}

// ServeHTTP 处理 HTTP 请求并返回响应。
func (s *Server) ServeHTTP(res http.ResponseWriter, req *http.Request) {
	s.Handler.ServeHTTP(res, req)
}

// filter 返回一个中间件函数，设置请求的上下文和其他过滤器。
func (s *Server) filter() mux.MiddlewareFunc {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			var (
				ctx    context.Context
				cancel context.CancelFunc
			)
			if s.timeout > 0 {
				ctx, cancel = context.WithTimeout(req.Context(), s.timeout)
			} else {
				ctx, cancel = context.WithCancel(req.Context())
			}
			defer cancel()

			// 获取路径模板，可能包含占位符
			pathTemplate := req.URL.Path
			if route := mux.CurrentRoute(req); route != nil {
				pathTemplate, _ = route.GetPathTemplate()
			}

			// 创建一个 Transport 对象封装 HTTP 请求和响应
			tr := &Transport{
				operation:    pathTemplate,
				pathTemplate: pathTemplate,
				reqHeader:    headerCarrier(req.Header),
				replyHeader:  headerCarrier(w.Header()),
				request:      req,
				response:     w,
			}
			if s.endpoint != nil {
				tr.endpoint = s.endpoint.String()
			}
			tr.request = req.WithContext(transport.NewServerContext(ctx, tr))
			// 调用下一个中间件或处理器
			next.ServeHTTP(w, tr.request)
		})
	}
}

// Endpoint 返回实际的服务器地址和端点。
func (s *Server) Endpoint() (*url.URL, error) {
	if err := s.listenAndEndpoint(); err != nil {
		return nil, err
	}
	return s.endpoint, nil
}

// Start 启动 HTTP 服务器。
func (s *Server) Start(ctx context.Context) error {
	if err := s.listenAndEndpoint(); err != nil {
		return err
	}
	// 设置上下文
	s.BaseContext = func(net.Listener) context.Context {
		return ctx
	}
	log.Infof("[HTTP] server listening on: %s", s.lis.Addr().String())
	var err error
	// 启动服务器（支持 TLS 和非 TLS）
	if s.tlsConf != nil {
		err = s.ServeTLS(s.lis, "", "")
	} else {
		err = s.Serve(s.lis)
	}
	if !errors.Is(err, http.ErrServerClosed) {
		return err
	}
	return nil
}

// Stop 停止 HTTP 服务器。
func (s *Server) Stop(ctx context.Context) error {
	log.Info("[HTTP] server stopping")
	return s.Shutdown(ctx)
}

// listenAndEndpoint 初始化监听器并确定服务器的端点地址。
func (s *Server) listenAndEndpoint() error {
	if s.lis == nil {
		lis, err := net.Listen(s.network, s.address)
		if err != nil {
			s.err = err
			return err
		}
		s.lis = lis
	}
	if s.endpoint == nil {
		addr, err := host.Extract(s.address, s.lis)
		if err != nil {
			s.err = err
			return err
		}
		s.endpoint = endpoint.NewEndpoint(endpoint.Scheme("http", s.tlsConf != nil), addr)
	}
	return s.err
}
