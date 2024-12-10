package http

import (
	"context"
	"encoding/json"
	"encoding/xml"
	"io"
	"net/http"
	"net/url"
	"time"

	"github.com/gorilla/mux"

	"github.com/cnsync/kratos/middleware"
	"github.com/cnsync/kratos/transport"
	"github.com/cnsync/kratos/transport/http/binding"
)

// Context 是一个HTTP请求上下文接口，包含了访问请求和响应的多种方法。
type Context interface {
	context.Context                                   // 嵌入 Go 的 context.Context，允许获取请求的上下文信息
	Vars() url.Values                                 // 获取URL中的路径参数
	Query() url.Values                                // 获取URL中的查询参数
	Form() url.Values                                 // 获取请求中的表单数据
	Header() http.Header                              // 获取请求头
	Request() *http.Request                           // 获取原始的HTTP请求
	Response() http.ResponseWriter                    // 获取HTTP响应
	Middleware(middleware.Handler) middleware.Handler // 获取请求的中间件
	Bind(interface{}) error                           // 绑定请求体到指定的对象
	BindVars(interface{}) error                       // 绑定路径参数到指定的对象
	BindQuery(interface{}) error                      // 绑定查询参数到指定的对象
	BindForm(interface{}) error                       // 绑定表单数据到指定的对象
	Returns(interface{}, error) error                 // 返回响应结果
	Result(int, interface{}) error                    // 返回一个特定HTTP状态码的响应
	JSON(int, interface{}) error                      // 返回JSON格式的响应
	XML(int, interface{}) error                       // 返回XML格式的响应
	String(int, string) error                         // 返回纯文本响应
	Blob(int, string, []byte) error                   // 返回二进制流响应
	Stream(int, string, io.Reader) error              // 返回流式数据响应
	Reset(http.ResponseWriter, *http.Request)         // 重置上下文为新的请求和响应
}

// responseWriter 用于包装 http.ResponseWriter，支持状态码设置。
type responseWriter struct {
	code int                 // 响应状态码
	w    http.ResponseWriter // 实际的 HTTP 响应
}

// reset 重置 responseWriter，初始化 HTTP 响应
func (w *responseWriter) reset(res http.ResponseWriter) {
	w.w = res
	w.code = http.StatusOK
}

// Header 返回响应的头部信息
func (w *responseWriter) Header() http.Header { return w.w.Header() }

// WriteHeader 设置响应的状态码
func (w *responseWriter) WriteHeader(statusCode int) { w.code = statusCode }

// Write 写入响应数据，并设置相应的状态码
func (w *responseWriter) Write(data []byte) (int, error) {
	w.w.WriteHeader(w.code)
	return w.w.Write(data)
}

// Unwrap 返回实际的 http.ResponseWriter
func (w *responseWriter) Unwrap() http.ResponseWriter { return w.w }

// wrapper 实现了 Context 接口，封装了 HTTP 请求和响应的相关信息。
type wrapper struct {
	router *Router             // 路由器
	req    *http.Request       // HTTP 请求
	res    http.ResponseWriter // HTTP 响应
	w      responseWriter      // 包装后的响应写入器
}

// Header 返回请求头部信息
func (c *wrapper) Header() http.Header {
	return c.req.Header
}

// Vars 返回URL中的路径变量（使用 gorilla/mux 路由时，路径参数将被解析并返回）
func (c *wrapper) Vars() url.Values {
	raws := mux.Vars(c.req)
	vars := make(url.Values, len(raws))
	for k, v := range raws {
		vars[k] = []string{v}
	}
	return vars
}

// Form 返回解析后的表单数据
func (c *wrapper) Form() url.Values {
	// 解析请求表单数据，若失败返回空的 url.Values
	if err := c.req.ParseForm(); err != nil {
		return url.Values{}
	}
	return c.req.Form
}

// Query 返回URL中的查询参数
func (c *wrapper) Query() url.Values {
	return c.req.URL.Query()
}

// Request 返回原始的 HTTP 请求
func (c *wrapper) Request() *http.Request {
	return c.req
}

// Response 返回原始的 HTTP 响应
func (c *wrapper) Response() http.ResponseWriter {
	return c.res
}

// Middleware 处理请求的中间件
func (c *wrapper) Middleware(h middleware.Handler) middleware.Handler {
	// 如果请求中包含 server 上下文，则使用该上下文的操作匹配中间件，否则使用请求的路径匹配中间件
	if tr, ok := transport.FromServerContext(c.req.Context()); ok {
		return middleware.Chain(c.router.srv.middleware.Match(tr.Operation())...)(h)
	}
	return middleware.Chain(c.router.srv.middleware.Match(c.req.URL.Path)...)(h)
}

// Bind 将请求体绑定到给定的结构体
func (c *wrapper) Bind(v interface{}) error {
	return c.router.srv.decBody(c.req, v)
}

// BindVars 将路径变量绑定到给定的结构体
func (c *wrapper) BindVars(v interface{}) error {
	return c.router.srv.decVars(c.req, v)
}

// BindQuery 将查询参数绑定到给定的结构体
func (c *wrapper) BindQuery(v interface{}) error {
	return c.router.srv.decQuery(c.req, v)
}

// BindForm 将表单数据绑定到给定的结构体
func (c *wrapper) BindForm(v interface{}) error {
	return binding.BindForm(c.req, v)
}

// Returns 返回响应体，如果没有错误，则将给定值编码为响应体
func (c *wrapper) Returns(v interface{}, err error) error {
	if err != nil {
		return err
	}
	return c.router.srv.enc(&c.w, c.req, v)
}

// Result 返回一个特定HTTP状态码的响应
func (c *wrapper) Result(code int, v interface{}) error {
	c.w.WriteHeader(code)
	return c.router.srv.enc(&c.w, c.req, v)
}

// JSON 返回 JSON 格式的响应
func (c *wrapper) JSON(code int, v interface{}) error {
	c.res.Header().Set("Content-Type", "application/json")
	c.res.WriteHeader(code)
	return json.NewEncoder(c.res).Encode(v)
}

// XML 返回 XML 格式的响应
func (c *wrapper) XML(code int, v interface{}) error {
	c.res.Header().Set("Content-Type", "application/xml")
	c.res.WriteHeader(code)
	return xml.NewEncoder(c.res).Encode(v)
}

// String 返回纯文本格式的响应
func (c *wrapper) String(code int, text string) error {
	c.res.Header().Set("Content-Type", "text/plain")
	c.res.WriteHeader(code)
	_, err := c.res.Write([]byte(text))
	if err != nil {
		return err
	}
	return nil
}

// Blob 返回二进制流数据作为响应
func (c *wrapper) Blob(code int, contentType string, data []byte) error {
	c.res.Header().Set("Content-Type", contentType)
	c.res.WriteHeader(code)
	_, err := c.res.Write(data)
	if err != nil {
		return err
	}
	return nil
}

// Stream 返回流式数据响应
func (c *wrapper) Stream(code int, contentType string, rd io.Reader) error {
	c.res.Header().Set("Content-Type", contentType)
	c.res.WriteHeader(code)
	_, err := io.Copy(c.res, rd)
	return err
}

// Reset 重置当前的 HTTP 请求和响应
func (c *wrapper) Reset(res http.ResponseWriter, req *http.Request) {
	c.w.reset(res)
	c.res = res
	c.req = req
}

// Deadline 返回请求的截止时间
func (c *wrapper) Deadline() (time.Time, bool) {
	if c.req == nil {
		return time.Time{}, false
	}
	return c.req.Context().Deadline()
}

// Done 返回请求的 Done channel，用于通知请求是否已完成
func (c *wrapper) Done() <-chan struct{} {
	if c.req == nil {
		return nil
	}
	return c.req.Context().Done()
}

// Err 返回请求的错误信息
func (c *wrapper) Err() error {
	if c.req == nil {
		return context.Canceled
	}
	return c.req.Context().Err()
}

// Value 返回请求上下文中的某个值
func (c *wrapper) Value(key interface{}) interface{} {
	if c.req == nil {
		return nil
	}
	return c.req.Context().Value(key)
}
