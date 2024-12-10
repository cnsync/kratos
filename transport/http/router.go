package http

import (
	"net/http"
	"path"
)

// WalkRouteFunc 是在遍历路由时，为每个访问的路由调用的函数类型。
// 它接收一个 RouteInfo 参数，返回一个 error，用于处理路由信息。
type WalkRouteFunc func(RouteInfo) error

// RouteInfo 是 HTTP 路由的信息结构体。
// 它包含路由的路径（Path）和方法（Method），用于描述一个 HTTP 路由。
type RouteInfo struct {
	Path   string // 路由的 URL 路径
	Method string // HTTP 请求方法（如 GET、POST、PUT 等）
}

// HandlerFunc 定义了一个处理 HTTP 请求的函数类型。
// 它接收一个 Context 类型的参数，并返回一个 error，表示请求的处理逻辑。
type HandlerFunc func(Context) error

// Router 是一个 HTTP 路由器，负责将 URL 路径和 HTTP 方法与对应的处理函数进行匹配。
// 它维护一个路由前缀、服务器实例和中间件过滤器。
type Router struct {
	prefix  string       // 路由前缀，所有路由的 URL 会基于这个前缀匹配
	srv     *Server      // 服务器实例，用于注册路由
	filters []FilterFunc // 路由的过滤器（中间件），用于在请求处理过程中执行
}

// newRouter 用于创建一个新的路由器实例。
// 它接受路由前缀、服务器实例和中间件列表，返回一个新的 Router 实例。
func newRouter(prefix string, srv *Server, filters ...FilterFunc) *Router {
	r := &Router{
		prefix:  prefix,  // 设置路由的前缀
		srv:     srv,     // 设置服务器实例
		filters: filters, // 设置中间件过滤器
	}
	return r
}

// Group 返回一个新的路由组。
// 它将创建一个新的路由器，并将当前路由的前缀和中间件过滤器与新组合并。
func (r *Router) Group(prefix string, filters ...FilterFunc) *Router {
	var newFilters []FilterFunc
	newFilters = append(newFilters, r.filters...)                       // 保留当前路由的过滤器
	newFilters = append(newFilters, filters...)                         // 添加新路由组的过滤器
	return newRouter(path.Join(r.prefix, prefix), r.srv, newFilters...) // 创建新的路由器组
}

// Handle 注册一个新的路由，匹配 URL 路径和 HTTP 方法。
// 它接收 HTTP 方法、相对路径、处理函数和过滤器（中间件）列表，并将其添加到路由器中。
func (r *Router) Handle(method, relativePath string, h HandlerFunc, filters ...FilterFunc) {
	// 将处理函数包裹为 http.Handler，并处理错误
	next := http.Handler(http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
		ctx := &wrapper{router: r} // 创建一个 Context 包装器
		ctx.Reset(res, req)        // 重置上下文
		if err := h(ctx); err != nil {
			r.srv.ene(res, req, err) // 如果处理函数返回错误，调用错误编码器
		}
	}))
	// 应用过滤器链
	next = FilterChain(filters...)(next)
	next = FilterChain(r.filters...)(next) // 将路由器的过滤器应用到处理函数
	// 注册路由到服务器
	r.srv.router.Handle(path.Join(r.prefix, relativePath), next).Methods(method)
}

// GET 注册一个新的 GET 请求路由，并将其与处理函数绑定。
func (r *Router) GET(path string, h HandlerFunc, m ...FilterFunc) {
	r.Handle(http.MethodGet, path, h, m...)
}

// HEAD 注册一个新的 HEAD 请求路由，并将其与处理函数绑定。
func (r *Router) HEAD(path string, h HandlerFunc, m ...FilterFunc) {
	r.Handle(http.MethodHead, path, h, m...)
}

// POST 注册一个新的 POST 请求路由，并将其与处理函数绑定。
func (r *Router) POST(path string, h HandlerFunc, m ...FilterFunc) {
	r.Handle(http.MethodPost, path, h, m...)
}

// PUT 注册一个新的 PUT 请求路由，并将其与处理函数绑定。
func (r *Router) PUT(path string, h HandlerFunc, m ...FilterFunc) {
	r.Handle(http.MethodPut, path, h, m...)
}

// PATCH 注册一个新的 PATCH 请求路由，并将其与处理函数绑定。
func (r *Router) PATCH(path string, h HandlerFunc, m ...FilterFunc) {
	r.Handle(http.MethodPatch, path, h, m...)
}

// DELETE 注册一个新的 DELETE 请求路由，并将其与处理函数绑定。
func (r *Router) DELETE(path string, h HandlerFunc, m ...FilterFunc) {
	r.Handle(http.MethodDelete, path, h, m...)
}

// CONNECT 注册一个新的 CONNECT 请求路由，并将其与处理函数绑定。
func (r *Router) CONNECT(path string, h HandlerFunc, m ...FilterFunc) {
	r.Handle(http.MethodConnect, path, h, m...)
}

// OPTIONS 注册一个新的 OPTIONS 请求路由，并将其与处理函数绑定。
func (r *Router) OPTIONS(path string, h HandlerFunc, m ...FilterFunc) {
	r.Handle(http.MethodOptions, path, h, m...)
}

// TRACE 注册一个新的 TRACE 请求路由，并将其与处理函数绑定。
func (r *Router) TRACE(path string, h HandlerFunc, m ...FilterFunc) {
	r.Handle(http.MethodTrace, path, h, m...)
}
