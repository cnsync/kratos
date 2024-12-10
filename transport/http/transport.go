package http

import (
	"context"
	"net/http"

	"github.com/cnsync/kratos/transport"
)

var _ Transporter = (*Transport)(nil)

// Transporter 是 HTTP 协议的 Transporter 接口，定义了 HTTP 请求的基本操作。
type Transporter interface {
	transport.Transporter   // 嵌套 transport.Transporter，继承了 transport 包中的功能
	Request() *http.Request // 获取 HTTP 请求
	PathTemplate() string   // 获取 HTTP 路径模板
}

// Transport 是一个 HTTP 协议的 Transport 实现，封装了 HTTP 请求和响应相关的字段和操作。
type Transport struct {
	endpoint     string              // 请求的端点（URL）
	operation    string              // 操作名称
	reqHeader    headerCarrier       // 请求头
	replyHeader  headerCarrier       // 响应头
	request      *http.Request       // HTTP 请求对象
	response     http.ResponseWriter // HTTP 响应对象
	pathTemplate string              // 请求路径模板
}

// Kind 返回当前 Transport 的协议类型，这里是 HTTP。
func (tr *Transport) Kind() transport.Kind {
	return transport.KindHTTP
}

// Endpoint 返回当前 HTTP 请求的端点（URL）。
func (tr *Transport) Endpoint() string {
	return tr.endpoint
}

// Operation 返回当前操作的名称。
func (tr *Transport) Operation() string {
	return tr.operation
}

// Request 返回当前的 HTTP 请求对象。
func (tr *Transport) Request() *http.Request {
	return tr.request
}

// RequestHeader 返回请求头信息。
func (tr *Transport) RequestHeader() transport.Header {
	return tr.reqHeader
}

// ReplyHeader 返回响应头信息。
func (tr *Transport) ReplyHeader() transport.Header {
	return tr.replyHeader
}

// PathTemplate 返回请求的路径模板。
func (tr *Transport) PathTemplate() string {
	return tr.pathTemplate
}

// SetOperation 设置当前的操作名称。此函数会从上下文中获取 Transport 对象并设置操作名称。
func SetOperation(ctx context.Context, op string) {
	if tr, ok := transport.FromServerContext(ctx); ok {
		if tr, ok := tr.(*Transport); ok {
			tr.operation = op
		}
	}
}

// SetCookie 向 HTTP 响应头添加一个 Set-Cookie 信息。
// 提供的 Cookie 必须有有效的 Name，若无效则会被忽略。
func SetCookie(ctx context.Context, cookie *http.Cookie) {
	if tr, ok := transport.FromServerContext(ctx); ok {
		if tr, ok := tr.(*Transport); ok {
			http.SetCookie(tr.response, cookie)
		}
	}
}

// RequestFromServerContext 从上下文中提取出 HTTP 请求对象。
// 返回值是请求对象和一个布尔值，表示是否成功获取。
func RequestFromServerContext(ctx context.Context) (*http.Request, bool) {
	if tr, ok := transport.FromServerContext(ctx); ok {
		if tr, ok := tr.(*Transport); ok {
			return tr.request, true
		}
	}
	return nil, false
}

// headerCarrier 是一个自定义类型，它实现了 transport.Header 接口，用于封装 HTTP 请求和响应头。
type headerCarrier http.Header

// Get 获取指定头部键的值。
func (hc headerCarrier) Get(key string) string {
	return http.Header(hc).Get(key)
}

// Set 设置指定头部键的值。
func (hc headerCarrier) Set(key string, value string) {
	http.Header(hc).Set(key, value)
}

// Add 向指定头部键添加一个值。
func (hc headerCarrier) Add(key string, value string) {
	http.Header(hc).Add(key, value)
}

// Keys 返回所有的头部键。
func (hc headerCarrier) Keys() []string {
	keys := make([]string, 0, len(hc))
	for k := range http.Header(hc) {
		keys = append(keys, k)
	}
	return keys
}

// Values 返回指定键的所有值。
func (hc headerCarrier) Values(key string) []string {
	return http.Header(hc).Values(key)
}
