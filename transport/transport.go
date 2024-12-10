package transport

import (
	"context"
	"net/url"

	// 初始化编码格式
	_ "github.com/cnsync/kratos/encoding/form"
	_ "github.com/cnsync/kratos/encoding/json"
	_ "github.com/cnsync/kratos/encoding/proto"
	_ "github.com/cnsync/kratos/encoding/xml"
	_ "github.com/cnsync/kratos/encoding/yaml"
)

// Server 是传输服务接口
type Server interface {
	// Start 启动服务
	Start(context.Context) error
	// Stop 停止服务
	Stop(context.Context) error
}

// EndpointProvider  是用于注册的端点接口
type EndpointProvider interface {
	// Endpoint 返回端点的 URL
	Endpoint() (*url.URL, error)
}

// Header 是存储头部数据的接口
type Header interface {
	// Get 获取指定 key 的值
	Get(key string) string
	// Set 设置指定 key 的值
	Set(key string, value string)
	// Add 添加指定 key 的值
	Add(key string, value string)
	// Keys 返回所有头部的键名
	Keys() []string
	// Values 返回指定 key 对应的所有值
	Values(key string) []string
}

// Transporter 是传输上下文值的接口
type Transporter interface {
	// Kind 返回传输类型
	// grpc 或 http
	Kind() Kind
	// Endpoint 返回服务端或客户端的端点
	// 服务端传输: grpc://127.0.0.1:9000
	// 客户端传输: discovery:///provider-demo
	Endpoint() string
	// Operation 返回服务的完整方法选择器（由 protobuf 生成）
	// 示例: /helloworld.Greeter/SayHello
	Operation() string
	// RequestHeader 返回传输请求头部
	// http: http.Header
	// grpc: metadata.MD
	RequestHeader() Header
	// ReplyHeader 返回传输响应头部（仅适用于服务端传输）
	// http: http.Header
	// grpc: metadata.MD
	ReplyHeader() Header
}

// Kind 定义传输的类型
type Kind string

// 返回 Kind 类型的字符串形式
func (k Kind) String() string { return string(k) }

// 定义一组传输类型
const (
	KindGRPC Kind = "grpc"
	KindHTTP Kind = "http"
)

type (
	serverTransportKey struct{}
	clientTransportKey struct{}
)

// NewServerContext 返回一个包含传输值的新上下文
func NewServerContext(ctx context.Context, tr Transporter) context.Context {
	return context.WithValue(ctx, serverTransportKey{}, tr)
}

// FromServerContext 返回上下文中存储的传输值（如果有）
func FromServerContext(ctx context.Context) (tr Transporter, ok bool) {
	tr, ok = ctx.Value(serverTransportKey{}).(Transporter)
	return
}

// NewClientContext 返回一个包含传输值的新上下文
func NewClientContext(ctx context.Context, tr Transporter) context.Context {
	return context.WithValue(ctx, clientTransportKey{}, tr)
}

// FromClientContext 返回上下文中存储的传输值（如果有）
func FromClientContext(ctx context.Context) (tr Transporter, ok bool) {
	tr, ok = ctx.Value(clientTransportKey{}).(Transporter)
	return
}
