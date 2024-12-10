package grpc

import (
	"google.golang.org/grpc/metadata"

	"github.com/cnsync/kratos/selector"
	"github.com/cnsync/kratos/transport"
)

var _ transport.Transporter = (*Transport)(nil)

// Transport 是一个 gRPC 传输器。
type Transport struct {
	endpoint    string                // 端点地址
	operation   string                // 操作名称
	reqHeader   headerCarrier         // 请求头
	replyHeader headerCarrier         // 回复头
	nodeFilters []selector.NodeFilter // 节点过滤器
}

// Kind 返回传输器的类型。
func (tr *Transport) Kind() transport.Kind {
	return transport.KindGRPC
}

// Endpoint 返回传输器的端点。
func (tr *Transport) Endpoint() string {
	return tr.endpoint
}

// Operation 返回传输器的操作。
func (tr *Transport) Operation() string {
	return tr.operation
}

// RequestHeader 返回请求头。
func (tr *Transport) RequestHeader() transport.Header {
	return tr.reqHeader
}

// ReplyHeader 返回回复头。
func (tr *Transport) ReplyHeader() transport.Header {
	return tr.replyHeader
}

// NodeFilters 返回客户端选择过滤器。
func (tr *Transport) NodeFilters() []selector.NodeFilter {
	return tr.nodeFilters
}

// headerCarrier 是一个用于携带 gRPC 元数据的载体。
type headerCarrier metadata.MD

// Get 返回与给定键关联的值。
func (mc headerCarrier) Get(key string) string {
	vals := metadata.MD(mc).Get(key)
	if len(vals) > 0 {
		return vals[0]
	}
	return ""
}

// Set 存储键值对。
func (mc headerCarrier) Set(key string, value string) {
	metadata.MD(mc).Set(key, value)
}

// Add 将值附加到键值对。
func (mc headerCarrier) Add(key string, value string) {
	metadata.MD(mc).Append(key, value)
}

// Keys 列出此载体中存储的键。
func (mc headerCarrier) Keys() []string {
	keys := make([]string, 0, len(mc))
	for k := range metadata.MD(mc) {
		keys = append(keys, k)
	}
	return keys
}

// Values 返回与给定键关联的值的切片。
func (mc headerCarrier) Values(key string) []string {
	return metadata.MD(mc).Get(key)
}
