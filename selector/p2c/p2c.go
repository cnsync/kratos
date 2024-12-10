package p2c

import (
	"context"
	"math/rand"
	"sync"
	"sync/atomic"
	"time"

	"github.com/cnsync/kratos/selector"
	"github.com/cnsync/kratos/selector/node/ewma"
)

const (
	forcePick = time.Second * 3
	// Name 是 p2c(Pick of 2 choices) 均衡器的名称
	Name = "p2c"
)

var _ selector.Balancer = (*Balancer)(nil)

// Option 是 p2c 构建器的选项。
type Option func(o *options)

// options 是 p2c 构建器的选项。
type options struct{}

// New 创建一个 p2c 选择器。
func New(opts ...Option) selector.Selector {
	return NewBuilder(opts...).Build()
}

// Balancer 是 p2c 选择器。
type Balancer struct {
	mu     sync.Mutex
	r      *rand.Rand
	picked int64
}

// prePick 方法从给定的节点列表中随机选择两个不同的节点
func (s *Balancer) prePick(nodes []selector.WeightedNode) (nodeA selector.WeightedNode, nodeB selector.WeightedNode) {
	// 加锁，确保在同一时间只有一个 goroutine 可以访问这段代码
	s.mu.Lock()
	// 从节点列表中随机选择一个索引
	a := s.r.Intn(len(nodes))
	// 从节点列表中随机选择一个索引，但排除已经选择过的索引
	b := s.r.Intn(len(nodes) - 1)
	// 解锁，允许其他 goroutine 访问这段代码
	s.mu.Unlock()
	// 如果索引 b 大于等于索引 a，则将索引 b 加 1，以确保选择的是不同的节点
	if b >= a {
		b = b + 1
	}
	// 返回选择的两个节点
	nodeA, nodeB = nodes[a], nodes[b]
	return
}

// Pick 选择一个节点。
func (s *Balancer) Pick(_ context.Context, nodes []selector.WeightedNode) (selector.WeightedNode, selector.DoneFunc, error) {
	if len(nodes) == 0 {
		return nil, nil, selector.ErrNoAvailable
	}
	if len(nodes) == 1 {
		done := nodes[0].Pick()
		return nodes[0], done, nil
	}

	var pc, upc selector.WeightedNode
	nodeA, nodeB := s.prePick(nodes)
	// meta.Weight 是服务发布者在发现中设置的权重
	if nodeB.Weight() > nodeA.Weight() {
		pc, upc = nodeB, nodeA
	} else {
		pc, upc = nodeA, nodeB
	}

	// 如果失败的节点在 forceGap 期间从未被选择过一次，则强制选择一次
	// 利用强制机会触发成功率和延迟的更新
	if upc.PickElapsed() > forcePick && atomic.CompareAndSwapInt64(&s.picked, 0, 1) {
		pc = upc
		atomic.StoreInt64(&s.picked, 0)
	}
	done := pc.Pick()
	return pc, done, nil
}

// NewBuilder 返回一个带有 p2c 均衡器的选择器构建器。
func NewBuilder(opts ...Option) selector.Builder {
	var option options
	for _, opt := range opts {
		opt(&option)
	}
	return &selector.DefaultBuilder{
		Balancer: &Builder{},
		Node:     &ewma.Builder{},
	}
}

// Builder 是 p2c 构建器。
type Builder struct{}

// Build 创建 Balancer。
func (b *Builder) Build() selector.Balancer {
	return &Balancer{r: rand.New(rand.NewSource(time.Now().UnixNano()))}
}
