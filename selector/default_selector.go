package selector

import (
	"context"
	"sync/atomic"
)

var (
	// 确保 Default 结构体实现了 Rebalancer 接口
	_ Rebalancer = (*Default)(nil)
	// 确保 DefaultBuilder 结构体实现了 Builder 接口
	_ Builder = (*DefaultBuilder)(nil)
)

// Default 是一个组合选择器。
type Default struct {
	// NodeBuilder 是一个加权节点构建器。
	NodeBuilder WeightedNodeBuilder
	// Balancer 是一个负载均衡器。
	Balancer Balancer

	// nodes 是一个原子值，用于存储节点列表。
	nodes atomic.Value
}

// Select 选择一个节点。
func (d *Default) Select(ctx context.Context, opts ...SelectOption) (selected Node, done DoneFunc, err error) {
	var (
		// options 是一个选择选项。
		options SelectOptions
		// candidates 是一个加权节点列表。
		candidates []WeightedNode
	)
	// 从原子值中加载节点列表。
	nodes, ok := d.nodes.Load().([]WeightedNode)
	if !ok {
		// 如果加载失败，返回没有可用节点的错误。
		return nil, nil, ErrNoAvailable
	}
	// 遍历选择选项，应用到 options 中。
	for _, o := range opts {
		o(&options)
	}
	// 如果有节点过滤器，则应用过滤器。
	if len(options.NodeFilters) > 0 {
		// 创建一个新的节点列表，用于存储过滤后的节点。
		newNodes := make([]Node, len(nodes))
		for i, wc := range nodes {
			// 将加权节点转换为节点。
			newNodes[i] = wc
		}
		// 遍历节点过滤器，对节点列表进行过滤。
		for _, filter := range options.NodeFilters {
			// 应用过滤器，得到过滤后的节点列表。
			newNodes = filter(ctx, newNodes)
		}
		// 将过滤后的节点列表转换为加权节点列表。
		candidates = make([]WeightedNode, len(newNodes))
		for i, n := range newNodes {
			// 将节点转换为加权节点。
			candidates[i] = n.(WeightedNode)
		}
	} else {
		// 如果没有节点过滤器，则直接使用原始节点列表。
		candidates = nodes
	}

	// 如果没有候选节点，返回没有可用节点的错误。
	if len(candidates) == 0 {
		return nil, nil, ErrNoAvailable
	}
	// 使用负载均衡器选择一个加权节点。
	wn, done, err := d.Balancer.Pick(ctx, candidates)
	if err != nil {
		// 如果选择失败，返回错误。
		return nil, nil, err
	}
	// 从上下文中获取对等节点信息。
	p, ok := FromPeerContext(ctx)
	if ok {
		// 如果获取成功，设置对等节点的信息。
		p.Node = wn.Raw()
	}
	// 返回选择的节点、完成函数和错误（如果有）。
	return wn.Raw(), done, nil
}

// Apply 更新节点信息。
func (d *Default) Apply(nodes []Node) {
	// 创建一个加权节点列表，用于存储更新后的节点。
	weightedNodes := make([]WeightedNode, 0, len(nodes))
	// 遍历节点列表，构建加权节点。
	for _, n := range nodes {
		// 使用加权节点构建器构建加权节点。
		weightedNodes = append(weightedNodes, d.NodeBuilder.Build(n))
	}
	// 将更新后的加权节点列表存储到原子值中。
	d.nodes.Store(weightedNodes)
}

// DefaultBuilder 是 Default 选择器的构建器。
type DefaultBuilder struct {
	// Node 是一个加权节点构建器。
	Node WeightedNodeBuilder
	// Balancer 是一个负载均衡器构建器。
	Balancer BalancerBuilder
}

// Build 创建 Default 选择器。
func (db *DefaultBuilder) Build() Selector {
	// 返回一个新的 Default 选择器实例。
	return &Default{
		NodeBuilder: db.Node,
		Balancer:    db.Balancer.Build(),
	}
}
