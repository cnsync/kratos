package grpc

import (
	"context"
	"reflect"
	"testing"

	"google.golang.org/grpc/metadata"

	"github.com/cnsync/kratos/selector"
)

// TestTrailer 测试 Trailer 类型的 Get 方法
func TestTrailer(t *testing.T) {
	// 创建一个 Trailer 实例，包含一个键值对 "a": "b"
	trailer := Trailer(metadata.New(map[string]string{"a": "b"}))
	// 测试 Get 方法是否能正确获取存在的键的值
	if !reflect.DeepEqual("b", trailer.Get("a")) {
		t.Errorf("expect %v, got %v", "b", trailer.Get("a"))
	}
	// 测试 Get 方法是否能正确处理不存在的键
	if !reflect.DeepEqual("", trailer.Get("notfound")) {
		t.Errorf("expect %v, got %v", "", trailer.Get("notfound"))
	}
}

// TestFilters 测试 WithNodeFilter 函数是否能正确添加节点过滤器
func TestFilters(t *testing.T) {
	// 创建一个 clientOptions 实例
	o := &clientOptions{}
	// 使用 WithNodeFilter 函数添加一个节点过滤器
	WithNodeFilter(func(_ context.Context, nodes []selector.Node) []selector.Node {
		return nodes
	})(o)
	// 测试 filters 切片的长度是否为 1
	if !reflect.DeepEqual(1, len(o.filters)) {
		t.Errorf("expect %v, got %v", 1, len(o.filters))
	}
}
