package p2c

import (
	"context"
	"fmt"
	"math/rand"
	"reflect"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/cnsync/kratos/registry"
	"github.com/cnsync/kratos/selector"
	"github.com/cnsync/kratos/selector/filter"
)

// TestWrr3 测试加权轮询算法的实现
func TestWrr3(t *testing.T) {
	// 创建一个新的加权轮询实例
	p2c := New()
	// 创建一个节点列表
	var nodes []selector.Node
	// 添加三个新的节点到列表中，每个节点使用 HTTP 协议，地址分别为 127.0.0.0:8080, 127.0.0.1:8080, 127.0.0.2:8080，服务实例 ID 和版本分别为相应的地址和 v2.0.0，权重为 10
	for i := 0; i < 3; i++ {
		addr := fmt.Sprintf("127.0.0.%d:8080", i)
		nodes = append(nodes, selector.NewNode(
			"http",
			addr,
			&registry.ServiceInstance{
				ID:       addr,
				Version:  "v2.0.0",
				Metadata: map[string]string{"weight": "10"},
			}))
	}
	// 将节点列表应用到加权轮询实例中
	p2c.Apply(nodes)
	// 初始化三个计数器，分别用于统计三个节点被选中的次数
	var count1, count2, count3 int64
	// 创建一个 WaitGroup，用于等待所有 goroutine 完成
	group := &sync.WaitGroup{}
	// 创建一个互斥锁，用于保护计数器
	var lk sync.Mutex
	// 模拟 9000 次选择操作
	for i := 0; i < 9000; i++ {
		// 为每个选择操作添加一个 WaitGroup 计数器
		group.Add(1)
		// 启动一个 goroutine 来执行选择操作
		go func() {
			// 在 goroutine 结束时，减少 WaitGroup 计数器
			defer group.Done()
			// 加锁以保护计数器
			lk.Lock()
			// 随机等待一段时间，模拟网络延迟
			d := time.Duration(rand.Intn(500)) * time.Millisecond
			// 解锁
			lk.Unlock()
			// 等待随机时间
			time.Sleep(d)
			// 调用加权轮询实例的 Select 方法，传入一个空的上下文和一个版本过滤器，过滤出版本为 v2.0.0 的节点
			n, done, err := p2c.Select(context.Background(), selector.WithNodeFilter(filter.Version("v2.0.0")))
			// 检查是否返回了错误
			if err != nil {
				// 如果返回了错误，记录错误信息
				t.Errorf("expect %v, got %v", nil, err)
			}
			// 检查是否返回了节点
			if n == nil {
				// 如果没有返回节点，记录错误信息
				t.Errorf("expect %v, got %v", nil, n)
			}
			// 检查是否返回了完成函数
			if done == nil {
				// 如果没有返回完成函数，记录错误信息
				t.Errorf("expect %v, got %v", nil, done)
			}
			// 等待 10 毫秒，模拟请求处理时间
			time.Sleep(time.Millisecond * 10)
			// 调用完成函数，传入一个空的上下文和一个空的完成信息
			done(context.Background(), selector.DoneInfo{})
			// 根据返回的节点地址，更新相应的计数器
			if n.Address() == "127.0.0.0:8080" {
				// 使用原子操作增加计数器的值
				atomic.AddInt64(&count1, 1)
			} else if n.Address() == "127.0.0.1:8080" {
				atomic.AddInt64(&count2, 1)
			} else if n.Address() == "127.0.0.2:8080" {
				atomic.AddInt64(&count3, 1)
			}
		}()
	}
	// 等待所有 goroutine 完成
	group.Wait()
	// 检查第一个节点被选中的次数是否符合预期
	if count1 <= int64(1500) {
		// 如果不符合预期，记录错误信息
		t.Errorf("count1(%v) <= int64(1500)", count1)
	}
	// 检查第一个节点被选中的次数是否符合预期
	if count1 >= int64(4500) {
		// 如果不符合预期，记录错误信息
		t.Errorf("count1(%v) >= int64(4500),", count1)
	}
	// 检查第二个节点被选中的次数是否符合预期
	if count2 <= int64(1500) {
		// 如果不符合预期，记录错误信息
		t.Errorf("count2(%v) <= int64(1500)", count2)
	}
	// 检查第二个节点被选中的次数是否符合预期
	if count2 >= int64(4500) {
		// 如果不符合预期，记录错误信息
		t.Errorf("count2(%v) >= int64(4500),", count2)
	}
	// 检查第三个节点被选中的次数是否符合预期
	if count3 <= int64(1500) {
		// 如果不符合预期，记录错误信息
		t.Errorf("count3(%v) <= int64(1500)", count3)
	}
	// 检查第三个节点被选中的次数是否符合预期
	if count3 >= int64(4500) {
		// 如果不符合预期，记录错误信息
		t.Errorf("count3(%v) >= int64(4500),", count3)
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
		t.Errorf("expect %v, got %v", nil, err)
	}
}

// TestOne 测试加权轮询算法的实现
func TestOne(t *testing.T) {
	// 创建一个新的加权轮询实例
	p2c := New()
	// 创建一个节点列表
	var nodes []selector.Node
	// 添加一个新的节点到列表中，该节点使用 HTTP 协议，地址为 127.0.0.0:8080，服务实例 ID 为 127.0.0.0:8080，版本为 v2.0.0，权重为 10
	for i := 0; i < 1; i++ {
		addr := fmt.Sprintf("127.0.0.%d:8080", i)
		nodes = append(nodes, selector.NewNode(
			"http",
			addr,
			&registry.ServiceInstance{
				ID:       addr,
				Version:  "v2.0.0",
				Metadata: map[string]string{"weight": "10"},
			}))
	}
	// 将节点列表应用到加权轮询实例中
	p2c.Apply(nodes)
	// 调用加权轮询实例的 Select 方法，传入一个空的上下文和一个版本过滤器，过滤出版本为 v2.0.0 的节点
	n, done, err := p2c.Select(context.Background(), selector.WithNodeFilter(filter.Version("v2.0.0")))
	// 检查是否返回了错误
	if err != nil {
		// 如果返回了错误，记录错误信息
		t.Errorf("expect %v, got %v", nil, err)
	}
	// 检查是否返回了节点
	if n == nil {
		// 如果没有返回节点，记录错误信息
		t.Errorf("expect %v, got %v", nil, n)
	}
	// 检查是否返回了完成函数
	if done == nil {
		// 如果没有返回完成函数，记录错误信息
		t.Errorf("expect %v, got %v", nil, done)
	}
	// 检查返回的节点地址是否符合预期
	if !reflect.DeepEqual("127.0.0.0:8080", n.Address()) {
		// 如果不符合预期，记录错误信息
		t.Errorf("expect %v, got %v", "127.0.0.0:8080", n.Address())
	}
}
