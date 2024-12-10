package selector

import (
	"context"

	"github.com/cnsync/kratos/errors"
)

// ErrNoAvailable 表示没有可用的节点。
var ErrNoAvailable = errors.ServiceUnavailable("no_available_node", "")

// Selector 是节点选择均衡器。
type Selector interface {
	Rebalancer

	// Select 选择节点。
	// 如果 err == nil，则 selected 和 done 不能为空。
	Select(ctx context.Context, opts ...SelectOption) (selected Node, done DoneFunc, err error)
}

// Rebalancer 是节点重新均衡器。
type Rebalancer interface {
	// Apply 在任何更改发生时应用所有节点。
	Apply(nodes []Node)
}

// Builder 构建选择器。
type Builder interface {
	Build() Selector
}

// Node 是节点接口。
type Node interface {
	// Scheme 是服务节点方案。
	Scheme() string

	// Address 是同一服务下的唯一地址。
	Address() string

	// ServiceName 是服务名称。
	ServiceName() string

	// InitialWeight 是调度权重的初始值。
	// 如果未设置，则返回 nil。
	InitialWeight() *int64

	// Version 是服务节点版本。
	Version() string

	// Metadata 是与服务实例关联的 kv 对元数据。
	// 版本、命名空间、区域、协议等。
	Metadata() map[string]string
}

// DoneInfo 是 RPC 调用完成时的回调信息。
type DoneInfo struct {
	// 响应错误。
	Err error
	// 响应元数据。
	ReplyMD ReplyMD

	// BytesSent 指示是否已向服务器发送任何字节。
	BytesSent bool
	// BytesReceived 指示是否已从服务器接收任何字节。
	BytesReceived bool
}

// ReplyMD 是回复元数据。
type ReplyMD interface {
	Get(key string) string
}

// DoneFunc 是 RPC 调用完成时的回调函数。
type DoneFunc func(ctx context.Context, di DoneInfo)
