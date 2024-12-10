package direct

import (
	"context"
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

	// 调用加权节点的 Pick 方法，获取一个完成函数
	done := wn.Pick()
	// 检查完成函数是否为 nil
	if done == nil {
		// 如果为 nil，记录错误信息
		t.Errorf("expect %v, got %v", nil, done)
	}
	// 等待 10 毫秒
	time.Sleep(time.Millisecond * 10)
	// 调用完成函数，传入一个空的上下文和一个包含错误信息的完成信息
	done(context.Background(), selector.DoneInfo{})
	// 检查加权节点的权重是否在预期范围内
	if !reflect.DeepEqual(float64(10), wn.Weight()) {
		// 如果不符合预期，记录错误信息
		t.Errorf("expect %v, got %v", float64(10), wn.Weight())
	}
	// 检查加权节点的选择延迟是否在预期范围内
	if time.Millisecond*20 <= wn.PickElapsed() {
		// 如果选择延迟小于等于 20 毫秒，记录错误信息
		t.Errorf("20ms <= wn.PickElapsed()(%s)", wn.PickElapsed())
	}
	// 检查加权节点的选择延迟是否在预期范围内
	if time.Millisecond*10 >= wn.PickElapsed() {
		// 如果选择延迟大于等于 10 毫秒，记录错误信息
		t.Errorf("10ms >= wn.PickElapsed()(%s)", wn.PickElapsed())
	}
}

// TestDirectDefaultWeight 测试直接选择算法在默认权重下的实现
func TestDirectDefaultWeight(t *testing.T) {
	// 创建一个 Builder 实例
	b := &Builder{}
	// 使用 Builder 实例构建一个加权节点，该节点的权重默认为 100
	wn := b.Build(selector.NewNode(
		"http",
		"127.0.0.1:9090",
		&registry.ServiceInstance{
			ID:        "127.0.0.1:9090",
			Name:      "helloworld",
			Version:   "v1.0.0",
			Endpoints: []string{"http://127.0.0.1:9090"},
		}))

	// 调用加权节点的 Pick 方法，获取一个完成函数
	done := wn.Pick()
	// 检查完成函数是否为 nil
	if done == nil {
		// 如果为 nil，记录错误信息
		t.Errorf("expect %v, got %v", nil, done)
	}
	// 等待 10 毫秒
	time.Sleep(time.Millisecond * 10)
	// 调用完成函数，传入一个空的上下文和一个包含错误信息的完成信息
	done(context.Background(), selector.DoneInfo{})
	// 检查加权节点的权重是否在预期范围内
	if !reflect.DeepEqual(float64(100), wn.Weight()) {
		// 如果不符合预期，记录错误信息
		t.Errorf("expect %v, got %v", float64(100), wn.Weight())
	}
	// 检查加权节点的选择延迟是否在预期范围内
	if time.Millisecond*20 <= wn.PickElapsed() {
		// 如果选择延迟小于等于 20 毫秒，记录错误信息
		t.Errorf("time.Millisecond*20 <= wn.PickElapsed()(%s)", wn.PickElapsed())
	}
	// 检查加权节点的选择延迟是否在预期范围内
	if time.Millisecond*5 >= wn.PickElapsed() {
		// 如果选择延迟大于等于 5 毫秒，记录错误信息
		t.Errorf("time.Millisecond*5 >= wn.PickElapsed()(%s)", wn.PickElapsed())
	}
}
