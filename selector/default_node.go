package selector

import (
	"strconv"

	"github.com/cnsync/kratos/registry"
)

var _ Node = (*DefaultNode)(nil)

// DefaultNode 是选择器节点
type DefaultNode struct {
	// scheme 是节点的通信协议，例如 "http" 或 "grpc"
	scheme string
	// addr 是节点的网络地址，例如 "127.0.0.1:9090"
	addr string
	// weight 是节点的权重，用于负载均衡
	weight *int64
	// version 是节点的版本号，用于区分不同版本的服务实例
	version string
	// name 是节点的服务名称，用于标识服务
	name string
	// metadata 是节点的元数据，包含一些额外的信息，例如标签、环境变量等
	metadata map[string]string
}

// Scheme 是节点方案
func (n *DefaultNode) Scheme() string {
	return n.scheme
}

// Address 是节点地址
func (n *DefaultNode) Address() string {
	return n.addr
}

// ServiceName 是节点服务名
func (n *DefaultNode) ServiceName() string {
	return n.name
}

// InitialWeight 是节点初始权重
func (n *DefaultNode) InitialWeight() *int64 {
	return n.weight
}

// Version 是节点版本
func (n *DefaultNode) Version() string {
	return n.version
}

// Metadata 是节点元数据
func (n *DefaultNode) Metadata() map[string]string {
	return n.metadata
}

// NewNode 函数根据给定的参数创建一个新的 DefaultNode 实例
func NewNode(scheme, addr string, ins *registry.ServiceInstance) Node {
	// 初始化一个新的 DefaultNode 实例 n
	n := &DefaultNode{
		// 设置节点的 scheme
		scheme: scheme,
		// 设置节点的 addr
		addr: addr,
	}

	// 如果 ins 不为 nil
	if ins != nil {
		// 设置节点的 name
		n.name = ins.Name
		// 设置节点的 version
		n.version = ins.Version
		// 设置节点的 metadata
		n.metadata = ins.Metadata
		// 检查 metadata 中是否存在 "weight" 键
		if str, ok := ins.Metadata["weight"]; ok {
			// 将 "weight" 键对应的值转换为 int64 类型
			if weight, err := strconv.ParseInt(str, 10, 64); err == nil {
				// 设置节点的 weight
				n.weight = &weight
			}
		}
	}
	// 返回新创建的 DefaultNode 实例
	return n
}
