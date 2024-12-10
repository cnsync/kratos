package selector

import (
	"context"
	"errors"
	"math/rand"
	"reflect"
	"sync/atomic"
	"testing"
	"time"

	"github.com/cnsync/kratos/registry"
)

// 定义一个错误，表示节点不匹配
var errNodeNotMatch = errors.New("node is not match")

// mockWeightedNode 是一个模拟的加权节点
type mockWeightedNode struct {
	Node

	lastPick int64
}

// Raw 返回原始的节点
func (n *mockWeightedNode) Raw() Node {
	return n.Node
}

// Weight 是运行时计算的权重
func (n *mockWeightedNode) Weight() float64 {
	if n.InitialWeight() != nil {
		return float64(*n.InitialWeight())
	}
	return 100
}

// Pick 选择节点
func (n *mockWeightedNode) Pick() DoneFunc {
	now := time.Now().UnixNano()
	atomic.StoreInt64(&n.lastPick, now)
	return func(context.Context, DoneInfo) {}
}

// PickElapsed 是自上次选择以来的时间
func (n *mockWeightedNode) PickElapsed() time.Duration {
	return time.Duration(time.Now().UnixNano() - atomic.LoadInt64(&n.lastPick))
}

// mockWeightedNodeBuilder 是一个模拟的加权节点构建器
type mockWeightedNodeBuilder struct{}

// Build 创建一个新的加权节点
func (b *mockWeightedNodeBuilder) Build(n Node) WeightedNode {
	return &mockWeightedNode{Node: n}
}

// mockFilter 函数根据给定的版本号过滤节点列表
func mockFilter(version string) NodeFilter {
	return func(_ context.Context, nodes []Node) []Node {
		// 创建一个新的切片来存储过滤后的节点
		newNodes := nodes[:0]
		// 遍历传入的节点列表
		for _, n := range nodes {
			// 如果节点的版本号与给定的版本号相匹配
			if n.Version() == version {
				// 将该节点添加到新的节点列表中
				newNodes = append(newNodes, n)
			}
		}
		// 返回过滤后的节点列表
		return newNodes
	}
}

// mockBalancerBuilder 是一个模拟的负载均衡器构建器
type mockBalancerBuilder struct{}

// Build 创建一个新的负载均衡器
func (b *mockBalancerBuilder) Build() Balancer {
	return &mockBalancer{}
}

// mockBalancer 是一个模拟的负载均衡器
type mockBalancer struct{}

// Pick 方法从给定的加权节点列表中选择一个节点
func (b *mockBalancer) Pick(_ context.Context, nodes []WeightedNode) (selected WeightedNode, done DoneFunc, err error) {
	// 如果节点列表为空，则返回 ErrNoAvailable 错误
	if len(nodes) == 0 {
		err = ErrNoAvailable
		return
	}
	// 从节点列表中随机选择一个索引
	cur := rand.Intn(len(nodes))
	// 获取选中的节点
	selected = nodes[cur]
	// 获取选中节点的完成函数
	done = selected.Pick()
	return
}

// mockMustErrorBalancerBuilder 是一个模拟的负载均衡器构建器，它总是返回错误
type mockMustErrorBalancerBuilder struct{}

// Build 创建一个新的负载均衡器
func (b *mockMustErrorBalancerBuilder) Build() Balancer {
	return &mockMustErrorBalancer{}
}

// mockMustErrorBalancer 是一个模拟的负载均衡器，它总是返回错误
type mockMustErrorBalancer struct{}

// Pick 选择一个节点
func (b *mockMustErrorBalancer) Pick(_ context.Context, _ []WeightedNode) (selected WeightedNode, done DoneFunc, err error) {
	return nil, nil, errNodeNotMatch
}

func TestDefault(t *testing.T) {
	// 创建一个 DefaultBuilder 实例，其中包含一个 mockWeightedNodeBuilder 和一个 mockBalancerBuilder
	builder := DefaultBuilder{
		Node:     &mockWeightedNodeBuilder{},
		Balancer: &mockBalancerBuilder{},
	}
	// 使用 builder 构建一个 selector
	selector := builder.Build()

	// 创建一个 Node 列表
	var nodes []Node
	// 添加一个 Node 实例，其服务实例的版本为 "v2.0.0"
	nodes = append(nodes, NewNode(
		"http",
		"127.0.0.1:8080",
		&registry.ServiceInstance{
			ID:        "127.0.0.1:8080",
			Name:      "helloworld",
			Version:   "v2.0.0",
			Endpoints: []string{"http://127.0.0.1:8080"},
			Metadata:  map[string]string{"weight": "10"},
		}))
	// 添加一个 Node 实例，其服务实例的版本为 "v1.0.0"
	nodes = append(nodes, NewNode(
		"http",
		"127.0.0.1:9090",
		&registry.ServiceInstance{
			ID:        "127.0.0.1:9090",
			Name:      "helloworld",
			Version:   "v1.0.0",
			Endpoints: []string{"http://127.0.0.1:9090"},
			Metadata:  map[string]string{"weight": "10"},
		}))

	// 将 nodes 应用到 selector 中
	selector.Apply(nodes)

	// 使用 WithNodeFilter 选项选择版本为 "v2.0.0" 的服务实例
	n, done, err := selector.Select(context.Background(), WithNodeFilter(mockFilter("v2.0.0")))
	// 如果发生错误，记录错误信息
	if err != nil {
		t.Errorf("expect %v, got %v", nil, err)
	}
	// 如果返回的 Node 实例为 nil，记录错误信息
	if n == nil {
		t.Errorf("expect %v, got %v", nil, n)
	}
	// 如果返回的 done 函数为 nil，记录错误信息
	if done == nil {
		t.Errorf("expect %v, got %v", nil, done)
	}
	// 如果返回的 Node 实例的版本不是 "v2.0.0"，记录错误信息
	if !reflect.DeepEqual("v2.0.0", n.Version()) {
		t.Errorf("expect %v, got %v", "v2.0.0", n.Version())
	}
	// 如果返回的 Node 实例的 Scheme 为空，记录错误信息
	if n.Scheme() == "" {
		t.Errorf("expect %v, got %v", "", n.Scheme())
	}
	// 如果返回的 Node 实例的 Address 为空，记录错误信息
	if n.Address() == "" {
		t.Errorf("expect %v, got %v", "", n.Address())
	}
	// 如果返回的 Node 实例的初始权重不是 10，记录错误信息
	if !reflect.DeepEqual(int64(10), *n.InitialWeight()) {
		t.Errorf("expect %v, got %v", 10, *n.InitialWeight())
	}
	// 如果返回的 Node 实例的元数据为 nil，记录错误信息
	if n.Metadata() == nil {
		t.Errorf("expect %v, got %v", nil, n.Metadata())
	}
	// 如果返回的 Node 实例的服务名为 "helloworld"，记录错误信息
	if !reflect.DeepEqual("helloworld", n.ServiceName()) {
		t.Errorf("expect %v, got %v", "helloworld", n.ServiceName())
	}
	// 调用 done 函数
	done(context.Background(), DoneInfo{})

	// 在上下文中设置 peer
	ctx := NewPeerContext(context.Background(), &Peer{
		Node: mockWeightedNode{},
	})
	// 从上下文中选择一个 Node 实例
	n, done, err = selector.Select(ctx)
	// 如果发生错误，记录错误信息
	if err != nil {
		t.Errorf("expect %v, got %v", ErrNoAvailable, err)
	}
	// 如果返回的 done 函数为 nil，记录错误信息
	if done == nil {
		t.Errorf("expect %v, got %v", nil, done)
	}
	// 如果返回的 Node 实例为 nil，记录错误信息
	if n == nil {
		t.Errorf("expect %v, got %v", nil, n)
	}

	// 选择不存在的版本 "v3.0.0" 的服务实例
	n, done, err = selector.Select(context.Background(), WithNodeFilter(mockFilter("v3.0.0")))
	// 如果错误不是 ErrNoAvailable，记录错误信息
	if !errors.Is(ErrNoAvailable, err) {
		t.Errorf("expect %v, got %v", ErrNoAvailable, err)
	}
	// 如果 done 函数不为 nil，记录错误信息
	if done != nil {
		t.Errorf("expect %v, got %v", nil, done)
	}
	// 如果 Node 实例不为 nil，记录错误信息
	if n != nil {
		t.Errorf("expect %v, got %v", nil, n)
	}

	// 应用零个实例
	selector.Apply([]Node{})
	// 选择版本为 "v2.0.0" 的服务实例
	n, done, err = selector.Select(context.Background(), WithNodeFilter(mockFilter("v2.0.0")))
	// 如果错误不是 ErrNoAvailable，记录错误信息
	if !errors.Is(ErrNoAvailable, err) {
		t.Errorf("expect %v, got %v", ErrNoAvailable, err)
	}
	// 如果 done 函数不为 nil，记录错误信息
	if done != nil {
		t.Errorf("expect %v, got %v", nil, done)
	}
	// 如果 Node 实例不为 nil，记录错误信息
	if n != nil {
		t.Errorf("expect %v, got %v", nil, n)
	}

	// 应用 nil 个实例
	selector.Apply(nil)
	// 选择版本为 "v2.0.0" 的服务实例
	n, done, err = selector.Select(context.Background(), WithNodeFilter(mockFilter("v2.0.0")))
	// 如果错误不是 ErrNoAvailable，记录错误信息
	if !errors.Is(ErrNoAvailable, err) {
		t.Errorf("expect %v, got %v", ErrNoAvailable, err)
	}
	// 如果 done 函数不为 nil，记录错误信息
	if done != nil {
		t.Errorf("expect %v, got %v", nil, done)
	}
	// 如果 Node 实例不为 nil，记录错误信息
	if n != nil {
		t.Errorf("expect %v, got %v", nil, n)
	}

	// 不使用 node_filters
	n, done, err = selector.Select(context.Background())
	// 如果错误不是 ErrNoAvailable，记录错误信息
	if !errors.Is(ErrNoAvailable, err) {
		t.Errorf("expect %v, got %v", ErrNoAvailable, err)
	}
	// 如果 done 函数不为 nil，记录错误信息
	if done != nil {
		t.Errorf("expect %v, got %v", nil, done)
	}
	// 如果 Node 实例不为 nil，记录错误信息
	if n != nil {
		t.Errorf("expect %v, got %v", nil, n)
	}
}

// TestWithoutApply 测试在没有应用任何节点的情况下，选择器的行为
func TestWithoutApply(t *testing.T) {
	// 创建一个默认构建器，其中包含一个模拟的加权节点构建器和一个模拟的负载均衡器构建器
	builder := DefaultBuilder{
		Node:     &mockWeightedNodeBuilder{},
		Balancer: &mockBalancerBuilder{},
	}
	// 使用默认构建器构建一个选择器
	selector := builder.Build()
	// 调用选择器的 Select 方法，传入一个空的上下文
	n, done, err := selector.Select(context.Background())
	// 检查返回的错误是否是 ErrNoAvailable
	if !errors.Is(ErrNoAvailable, err) {
		// 如果不是，记录错误信息
		t.Errorf("expect %v, got %v", ErrNoAvailable, err)
	}
	// 检查返回的完成函数是否为 nil
	if done != nil {
		// 如果不是，记录错误信息
		t.Errorf("expect %v, got %v", nil, done)
	}
	// 检查返回的节点是否为 nil
	if n != nil {
		// 如果不是，记录错误信息
		t.Errorf("expect %v, got %v", nil, n)
	}
}

// TestNoPick 测试在应用了节点但负载均衡器总是返回错误的情况下，选择器的行为
func TestNoPick(t *testing.T) {
	// 创建一个默认构建器，其中包含一个模拟的加权节点构建器和一个模拟的总是返回错误的负载均衡器构建器
	builder := DefaultBuilder{
		Node:     &mockWeightedNodeBuilder{},
		Balancer: &mockMustErrorBalancerBuilder{},
	}

	// 创建一个节点列表
	var nodes []Node
	// 添加一个新的节点到列表中，该节点使用 HTTP 协议，地址为 127.0.0.1:8080，服务实例 ID 为 127.0.0.1:8080，服务名为 helloworld，版本为 v2.0.0，端点为 http://127.0.0.1:8080，元数据为 {"weight": "10"}
	nodes = append(nodes, NewNode(
		"http",
		"127.0.0.1:8080",
		&registry.ServiceInstance{
			ID:        "127.0.0.1:8080",
			Name:      "helloworld",
			Version:   "v2.0.0",
			Endpoints: []string{"http://127.0.0.1:8080"},
			Metadata:  map[string]string{"weight": "10"},
		}))
	// 添加另一个新的节点到列表中，该节点使用 HTTP 协议，地址为 127.0.0.1:9090，服务实例 ID 为 127.0.0.1:9090，服务名为 helloworld，版本为 v1.0.0，端点为 http://127.0.0.1:9090，元数据为 {"weight": "10"}
	nodes = append(nodes, NewNode(
		"http",
		"127.0.0.1:9090",
		&registry.ServiceInstance{
			ID:        "127.0.0.1:9090",
			Name:      "helloworld",
			Version:   "v1.0.0",
			Endpoints: []string{"http://127.0.0.1:9090"},
			Metadata:  map[string]string{"weight": "10"},
		}))

	// 使用默认构建器构建一个选择器
	selector := builder.Build()
	// 将节点列表应用到选择器中
	selector.Apply(nodes)

	// 调用选择器的 Select 方法，传入一个空的上下文
	n, done, err := selector.Select(context.Background())
	// 检查返回的错误是否是 errNodeNotMatch
	if !errors.Is(errNodeNotMatch, err) {
		// 如果不是，记录错误信息
		t.Errorf("expect %v, got %v", errNodeNotMatch, err)
	}
	// 检查返回的完成函数是否为 nil
	if done != nil {
		// 如果不是，记录错误信息
		t.Errorf("expect %v, got %v", nil, done)
	}
	// 检查返回的节点是否为 nil
	if n != nil {
		// 如果不是，记录错误信息
		t.Errorf("expect %v, got %v", nil, n)
	}
}

// TestGlobalSelector 测试全局选择器的设置和获取
func TestGlobalSelector(t *testing.T) {
	// 创建一个默认构建器，其中包含一个模拟的加权节点构建器和一个模拟的负载均衡器构建器
	builder := DefaultBuilder{
		Node:     &mockWeightedNodeBuilder{},
		Balancer: &mockBalancerBuilder{},
	}
	// 设置全局选择器为创建的默认构建器
	SetGlobalSelector(&builder)

	// 获取全局选择器
	gBuilder := GlobalSelector()
	// 检查获取到的全局选择器是否为 nil
	if gBuilder == nil {
		// 如果是，记录错误信息
		t.Errorf("expect %v, got %v", nil, gBuilder)
	}
}
