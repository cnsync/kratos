package ewma

import (
	"context"
	"net"
	"reflect"
	"testing"
	"time"

	"github.com/cnsync/kratos/registry"
	"github.com/cnsync/kratos/selector"
)

// TestDirect 测试直接选择算法的实现
func TestDirect(t *testing.T) {
	// 创建一个 Builder 实例
	b := &Builder{}
	// 使用 Builder 实例构建一个加权节点
	wn := b.Build(selector.NewNode(
		"http",
		"127.0.0.1:9090",
		&registry.ServiceInstance{
			ID:        "127.0.0.1:9090",
			Name:      "helloworld",
			Version:   "v1.0.0",
			Endpoints: []string{"http://127.0.0.1:9090"},
			Metadata:  map[string]string{"weight": "10"},
		}))

	// 检查加权节点的权重是否符合预期
	if !reflect.DeepEqual(float64(100), wn.Weight()) {
		// 如果不符合预期，记录错误信息
		t.Errorf("expect %v, got %v", 100, wn.Weight())
	}
	// 调用加权节点的 Pick 方法，获取一个完成函数
	done := wn.Pick()
	// 检查完成函数是否为 nil
	if done == nil {
		// 如果为 nil，记录错误信息
		t.Errorf("done is equal to nil")
	}
	// 再次调用加权节点的 Pick 方法，获取另一个完成函数
	done2 := wn.Pick()
	// 检查完成函数是否为 nil
	if done2 == nil {
		// 如果为 nil，记录错误信息
		t.Errorf("done2 is equal to nil")
	}

	// 等待 15 毫秒
	time.Sleep(time.Millisecond * 15)
	// 调用第一个完成函数，传入一个空的上下文和一个空的完成信息
	done(context.Background(), selector.DoneInfo{})
	// 检查加权节点的权重是否在预期范围内
	if float64(70) >= wn.Weight() {
		// 如果权重小于等于 70，记录错误信息
		t.Errorf("float64(30000) >= wn.Weight()(%v)", wn.Weight())
	}
	// 检查加权节点的权重是否在预期范围内
	if float64(1200) <= wn.Weight() {
		// 如果权重大于等于 1200，记录错误信息
		t.Errorf("float64(1000) <= wn.Weight()(%v)", wn.Weight())
	}
	// 检查加权节点的选择延迟是否在预期范围内
	if time.Millisecond*30 <= wn.PickElapsed() {
		// 如果选择延迟小于等于 30 毫秒，记录错误信息
		t.Errorf("time.Millisecond*30 <= wn.PickElapsed()(%v)", wn.PickElapsed())
	}
	// 检查加权节点的选择延迟是否在预期范围内
	if time.Millisecond*5 >= wn.PickElapsed() {
		// 如果选择延迟大于等于 5 毫秒，记录错误信息
		t.Errorf("time.Millisecond*5 >= wn.PickElapsed()(%v)", wn.PickElapsed())
	}
}

// TestDirectError 测试直接选择算法在处理错误时的实现
func TestDirectError(t *testing.T) {
	// 创建一个 Builder 实例
	b := &Builder{}
	// 使用 Builder 实例构建一个加权节点
	wn := b.Build(selector.NewNode(
		"http",
		"127.0.0.1:9090",
		&registry.ServiceInstance{
			ID:        "127.0.0.1:9090",
			Name:      "helloworld",
			Version:   "v1.0.0",
			Endpoints: []string{"http://127.0.0.1:9090"},
			Metadata:  map[string]string{"weight": "10"},
		}))

	// 模拟 5 次请求，其中第一次请求没有错误，其余请求设置为超时错误
	for i := 0; i < 5; i++ {
		var err error
		if i != 0 {
			err = context.DeadlineExceeded
		}
		// 调用加权节点的 Pick 方法，获取一个完成函数
		done := wn.Pick()
		// 检查完成函数是否为 nil
		if done == nil {
			// 如果为 nil，记录错误信息
			t.Errorf("expect not nil, got nil")
		}
		// 等待 20 毫秒
		time.Sleep(time.Millisecond * 20)
		// 调用完成函数，传入一个空的上下文和一个包含错误信息的完成信息
		done(context.Background(), selector.DoneInfo{Err: err})
	}
	// 检查加权节点的权重是否在预期范围内
	if float64(1000) >= wn.Weight() {
		// 如果权重小于等于 1000，记录错误信息
		t.Errorf("float64(1000) >= wn.Weight()(%v)", wn.Weight())
	}
	// 检查加权节点的权重是否在预期范围内
	if float64(2000) <= wn.Weight() {
		// 如果权重大于等于 2000，记录错误信息
		t.Errorf("float64(2000) <= wn.Weight()(%v)", wn.Weight())
	}
}

// TestDirectErrorHandler 测试直接选择算法在处理错误时的实现，并且使用了自定义的错误处理函数
func TestDirectErrorHandler(t *testing.T) {
	// 创建一个 Builder 实例，并设置错误处理函数
	b := &Builder{
		// 定义错误处理函数，当错误不为 nil 时返回 true
		ErrHandler: func(err error) bool {
			return err != nil
		},
	}
	// 使用 Builder 实例构建一个加权节点
	wn := b.Build(selector.NewNode(
		"http",
		"127.0.0.1:9090",
		&registry.ServiceInstance{
			ID:        "127.0.0.1:9090",
			Name:      "helloworld",
			Version:   "v1.0.0",
			Endpoints: []string{"http://127.0.0.1:9090"},
			Metadata:  map[string]string{"weight": "10"},
		}))
	// 定义一个错误列表，包含多种类型的错误
	errs := []error{
		context.DeadlineExceeded,
		context.Canceled,
		net.ErrClosed,
	}
	// 模拟 5 次请求，其中第一次请求没有错误，其余请求设置为超时错误
	for i := 0; i < 5; i++ {
		var err error
		if i != 0 {
			// 从错误列表中获取一个错误
			err = errs[i%len(errs)]
		}
		// 调用加权节点的 Pick 方法，获取一个完成函数
		done := wn.Pick()
		// 检查完成函数是否为 nil
		if done == nil {
			// 如果为 nil，记录错误信息
			t.Errorf("expect not nil, got nil")
		}
		// 等待 20 毫秒
		time.Sleep(time.Millisecond * 20)
		// 调用完成函数，传入一个空的上下文和一个包含错误信息的完成信息
		done(context.Background(), selector.DoneInfo{Err: err})
	}
	// 检查加权节点的权重是否在预期范围内
	if float64(1000) >= wn.Weight() {
		// 如果权重小于等于 1000，记录错误信息
		t.Errorf("float64(100) >= wn.Weight()(%v)", wn.Weight())
	}
	// 检查加权节点的权重是否在预期范围内
	if float64(2000) <= wn.Weight() {
		// 如果权重大于等于 2000，记录错误信息
		t.Errorf("float64(200) <= wn.Weight()(%v)", wn.Weight())
	}
}

// BenchmarkPickAndWeight 测试直接选择算法的实现的性能
func BenchmarkPickAndWeight(b *testing.B) {
	// 创建一个 Builder 实例
	bu := &Builder{}
	// 使用 Builder 实例构建一个加权节点
	node := bu.Build(selector.NewNode(
		"http",
		"127.0.0.1:9090",
		&registry.ServiceInstance{
			ID:        "127.0.0.1:9090",
			Name:      "helloworld",
			Version:   "v1.0.0",
			Endpoints: []string{"http://127.0.0.1:9090"},
			Metadata:  map[string]string{"weight": "10"},
		}))
	// 定义一个完成信息
	di := selector.DoneInfo{}
	// 使用 b.RunParallel 进行并行测试
	b.RunParallel(func(pb *testing.PB) {
		// 循环执行测试
		for pb.Next() {
			// 调用加权节点的 Pick 方法，获取一个完成函数
			done := node.Pick()
			// 获取加权节点的权重
			node.Weight()
			// 调用完成函数，传入一个空的上下文和一个包含错误信息的完成信息
			done(context.Background(), di)
		}
	})
}
