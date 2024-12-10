package wrr

import (
	"context"
	"sync"

	"github.com/cnsync/kratos/selector"
	"github.com/cnsync/kratos/selector/node/direct"
)

const (
	// Name 是 wrr(Weighted Round Robin) 均衡器的名称
	Name = "wrr"
)

var _ selector.Balancer = (*Balancer)(nil) // Name 是均衡器的名称

// Option 是 wrr 构建器的选项。
type Option func(o *options)

// options 是 wrr 构建器的选项。
type options struct{}

// Balancer 是一个 wrr 均衡器。
type Balancer struct {
	mu            sync.Mutex
	currentWeight map[string]float64
}

// New 随机选择一个选择器。
func New(opts ...Option) selector.Selector {
	return NewBuilder(opts...).Build()
}

// Pick 方法实现了从给定的节点列表中选择一个节点的逻辑
func (p *Balancer) Pick(_ context.Context, nodes []selector.WeightedNode) (selector.WeightedNode, selector.DoneFunc, error) {
	// 如果节点列表为空，则返回错误
	if len(nodes) == 0 {
		return nil, nil, selector.ErrNoAvailable
	}
	// 初始化总权重为 0
	var totalWeight float64
	// 初始化选中的节点为空
	var selected selector.WeightedNode
	// 初始化选中的权重为 0
	var selectWeight float64

	// 使用互斥锁保证线程安全
	p.mu.Lock()
	// 遍历节点列表
	for _, node := range nodes {
		// 累加总权重
		totalWeight += node.Weight()
		// 获取当前节点的当前权重
		cwt := p.currentWeight[node.Address()]
		// 当前权重加上有效权重
		cwt += node.Weight()
		// 更新当前节点的当前权重
		p.currentWeight[node.Address()] = cwt
		// 如果当前节点的权重大于选中节点的权重，则更新选中节点
		if selected == nil || selectWeight < cwt {
			selectWeight = cwt
			selected = node
		}
	}
	// 更新选中节点的当前权重为选中权重减去总权重
	p.currentWeight[selected.Address()] = selectWeight - totalWeight
	// 释放互斥锁
	p.mu.Unlock()

	// 调用选中节点的 Pick 方法获取完成函数
	d := selected.Pick()
	// 返回选中的节点、完成函数和 nil 错误
	return selected, d, nil
}

// NewBuilder 函数根据给定的选项创建一个新的选择器构建器实例
func NewBuilder(opts ...Option) selector.Builder {
	// 初始化一个新的 options 实例 option
	var option options
	// 遍历所有传入的选项
	for _, opt := range opts {
		// 将每个选项应用到 option 实例上
		opt(&option)
	}
	// 返回一个新的 DefaultBuilder 实例，其中包含了 wrr 均衡器构建器和直接节点构建器
	return &selector.DefaultBuilder{
		// 设置 Balancer 为 Builder 实例
		Balancer: &Builder{},
		// 设置 Node 为 direct.Builder 实例
		Node: &direct.Builder{},
	}
}

// Builder 是 wrr 构建器。
type Builder struct{}

// Build 创建 Balancer。
func (b *Builder) Build() selector.Balancer {
	return &Balancer{currentWeight: make(map[string]float64)}
}
