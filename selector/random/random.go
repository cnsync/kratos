package random

import (
	"context"
	"math/rand"

	"github.com/cnsync/kratos/selector"
	"github.com/cnsync/kratos/selector/node/direct"
)

const (
	// Name 是随机均衡器的名称
	Name = "random"
)

var _ selector.Balancer = (*Balancer)(nil) // Name 是均衡器的名称

// Option 是随机构建器的选项。
type Option func(o *options)

// options 是随机构建器的选项。
type options struct{}

// Balancer 是一个随机均衡器。
type Balancer struct{}

// New 随机选择一个选择器。
func New(opts ...Option) selector.Selector {
	return NewBuilder(opts...).Build()
}

// Pick 是选择一个加权节点。
func (p *Balancer) Pick(_ context.Context, nodes []selector.WeightedNode) (selector.WeightedNode, selector.DoneFunc, error) {
	if len(nodes) == 0 {
		return nil, nil, selector.ErrNoAvailable
	}
	// 生成一个随机索引
	cur := rand.Intn(len(nodes))
	// 选择随机索引对应的节点
	selected := nodes[cur]
	// 调用节点的 Pick 方法获取完成函数
	d := selected.Pick()
	// 返回选中的节点、完成函数和 nil 错误
	return selected, d, nil
}

// NewBuilder 返回一个带有随机均衡器的选择器构建器。
func NewBuilder(opts ...Option) selector.Builder {
	var option options
	for _, opt := range opts {
		opt(&option)
	}
	return &selector.DefaultBuilder{
		Balancer: &Builder{},
		Node:     &direct.Builder{},
	}
}

// Builder 是随机构建器。
type Builder struct{}

// Build 创建 Balancer。
func (b *Builder) Build() selector.Balancer {
	return &Balancer{}
}
