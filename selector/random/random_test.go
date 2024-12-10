package random

import (
	"context"
	"testing"

	"github.com/cnsync/kratos/registry"
	"github.com/cnsync/kratos/selector"
	"github.com/cnsync/kratos/selector/filter"
)

// TestWrr 测试加权轮询算法的实现
func TestWrr(t *testing.T) {
	// 创建一个新的加权轮询实例
	random := New()
	// 创建一个节点列表
	var nodes []selector.Node
	// 添加一个新的节点到列表中，该节点使用 HTTP 协议，地址为 127.0.0.1:8080，服务实例 ID 为 127.0.0.1:8080，版本为 v2.0.0，权重为 10
	nodes = append(nodes, selector.NewNode(
		"http",
		"127.0.0.1:8080",
		&registry.ServiceInstance{
			ID:       "127.0.0.1:8080",
			Version:  "v2.0.0",
			Metadata: map[string]string{"weight": "10"},
		}))
	// 添加另一个新的节点到列表中，该节点使用 HTTP 协议，地址为 127.0.0.1:9090，服务实例 ID 为 127.0.0.1:9090，版本为 v2.0.0，权重为 20
	nodes = append(nodes, selector.NewNode(
		"http",
		"127.0.0.1:9090",
		&registry.ServiceInstance{
			ID:       "127.0.0.1:9090",
			Version:  "v2.0.0",
			Metadata: map[string]string{"weight": "20"},
		}))
	// 将节点列表应用到加权轮询实例中
	random.Apply(nodes)
	// 初始化两个计数器，分别用于统计两个节点被选中的次数
	var count1, count2 int
	// 模拟 200 次选择操作
	for i := 0; i < 200; i++ {
		// 调用加权轮询实例的 Select 方法，传入一个空的上下文和一个版本过滤器，过滤出版本为 v2.0.0 的节点
		n, done, err := random.Select(context.Background(), selector.WithNodeFilter(filter.Version("v2.0.0")))
		// 检查是否返回了错误
		if err != nil {
			// 如果返回了错误，记录错误信息
			t.Errorf("expect no error, got %v", err)
		}
		// 检查是否返回了完成函数
		if done == nil {
			// 如果没有返回完成函数，记录错误信息
			t.Errorf("expect not nil, got:%v", done)
		}
		// 检查是否返回了节点
		if n == nil {
			// 如果没有返回节点，记录错误信息
			t.Errorf("expect not nil, got:%v", n)
		}
		// 调用完成函数，传入一个空的上下文和一个空的完成信息
		done(context.Background(), selector.DoneInfo{})
		// 根据返回的节点地址，更新相应的计数器
		if n.Address() == "127.0.0.1:8080" {
			count1++
		} else if n.Address() == "127.0.0.1:9090" {
			count2++
		}
	}
	// 检查第一个节点被选中的次数是否符合预期
	if count1 <= 80 {
		// 如果不符合预期，记录错误信息
		t.Errorf("count1(%v) <= 80", count1)
	}
	// 检查第一个节点被选中的次数是否符合预期
	if count1 >= 120 {
		// 如果不符合预期，记录错误信息
		t.Errorf("count1(%v) >= 120", count1)
	}
	// 检查第二个节点被选中的次数是否符合预期
	if count2 <= 80 {
		// 如果不符合预期，记录错误信息
		t.Errorf("count2(%v) <= 80", count2)
	}
	// 检查第二个节点被选中的次数是否符合预期
	if count2 >= 120 {
		// 如果不符合预期，记录错误信息
		t.Errorf("count2(%v) >= 120", count2)
	}
}

// TestEmpty 测试在没有节点的情况下，Balancer 接口的 Pick 方法是否返回 ErrNoAvailable 错误
func TestEmpty(t *testing.T) {
	// 创建一个 Balancer 实例
	b := &Balancer{}
	// 调用 Balancer 实例的 Pick 方法，传入一个空的上下文和一个空的加权节点列表
	_, _, err := b.Pick(context.Background(), []selector.WeightedNode{})
	// 检查是否返回了错误
	if err == nil {
		// 如果没有返回错误，记录错误信息
		t.Errorf("expect nil, got %v", err)
	}
}
