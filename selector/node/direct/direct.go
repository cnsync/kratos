package direct

import (
	"context"
	"sync/atomic"
	"time"

	"github.com/cnsync/kratos/selector"
)

const (
	// 默认权重值
	defaultWeight = 100
)

var (
	// 确保 Node 类型实现了 selector.WeightedNode 接口
	_ selector.WeightedNode = (*Node)(nil)
	// 确保 Builder 类型实现了 selector.WeightedNodeBuilder 接口
	_ selector.WeightedNodeBuilder = (*Builder)(nil)
)

// Node 是端点实例
type Node struct {
	// 嵌入 selector.Node 类型
	selector.Node

	// 最后一次选择的时间戳
	lastPick int64
}

// Builder 是直接节点构建器
type Builder struct{}

// Build 创建一个新的节点
func (*Builder) Build(n selector.Node) selector.WeightedNode {
	// 返回一个新的 Node 实例，初始 lastPick 时间戳为 0
	return &Node{Node: n, lastPick: 0}
}

// Pick 选择一个节点并返回一个完成时调用的回调函数
func (n *Node) Pick() selector.DoneFunc {
	// 记录当前时间，作为请求开始时间
	now := time.Now().UnixNano()
	// 更新节点的 lastPick 时间为当前时间
	atomic.StoreInt64(&n.lastPick, now)
	// 返回一个空的回调函数，因为在这个简单的实现中，我们不需要在请求完成时做任何事情
	return func(context.Context, selector.DoneInfo) {}
}

// Weight 获取节点的有效权重
func (n *Node) Weight() float64 {
	// 如果节点有初始权重，则返回初始权重
	if n.InitialWeight() != nil {
		return float64(*n.InitialWeight())
	}
	// 如果没有初始权重，则返回默认权重
	return defaultWeight
}

// PickElapsed 获取自上次选择以来的时间
func (n *Node) PickElapsed() time.Duration {
	// 返回当前时间与 lastPick 时间戳之间的差值
	return time.Duration(time.Now().UnixNano() - atomic.LoadInt64(&n.lastPick))
}

// Raw 返回原始的 selector.Node 实例
func (n *Node) Raw() selector.Node {
	// 返回嵌入的 Node 实例
	return n.Node
}
