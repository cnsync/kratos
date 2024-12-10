package selector

import (
	"context"
	"time"
)

// Balancer 是负载均衡器接口
type Balancer interface {
	// Pick 方法从给定的节点列表中选择一个节点，并返回一个完成函数和可能的错误
	Pick(ctx context.Context, nodes []WeightedNode) (selected WeightedNode, done DoneFunc, err error)
}

// BalancerBuilder 是负载均衡器构建器接口
type BalancerBuilder interface {
	// Build 方法构建一个负载均衡器实例
	Build() Balancer
}

// WeightedNode 是实时计算调度权重的接口
type WeightedNode interface {
	Node

	// Raw 方法返回原始节点
	Raw() Node

	// Weight 方法返回运行时计算的权重
	Weight() float64

	// Pick 方法选择节点
	Pick() DoneFunc

	// PickElapsed 方法返回自上次选择以来的时间间隔
	PickElapsed() time.Duration
}

// WeightedNodeBuilder 是加权节点构建器接口
type WeightedNodeBuilder interface {
	// Build 方法构建一个加权节点实例
	Build(Node) WeightedNode
}
