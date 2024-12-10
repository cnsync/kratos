package transport

import (
	"context"
	"reflect"
	"testing"
)

// mockTransport 是一个 gRPC 传输的模拟实现。
type mockTransport struct {
	endpoint  string
	operation string
}

// Kind 返回传输的类型。
func (tr *mockTransport) Kind() Kind {
	return KindGRPC
}

// Endpoint 返回传输的端点。
func (tr *mockTransport) Endpoint() string {
	return tr.endpoint
}

// Operation 返回传输的操作。
func (tr *mockTransport) Operation() string {
	return tr.operation
}

// RequestHeader 返回请求头。
func (tr *mockTransport) RequestHeader() Header {
	return nil
}

// ReplyHeader 返回回复头。
func (tr *mockTransport) ReplyHeader() Header {
	return nil
}

// TestServerTransport 测试服务器传输的功能。
func TestServerTransport(t *testing.T) {
	ctx := context.Background()

	// 创建一个新的服务器上下文，并将 mockTransport 实例作为值传递给它。
	ctx = NewServerContext(ctx, &mockTransport{endpoint: "test_endpoint"})

	// 从服务器上下文中获取传输实例。
	tr, ok := FromServerContext(ctx)
	if !ok {
		// 如果获取传输实例失败，则记录错误信息。
		t.Errorf("expected:%v got:%v", true, ok)
	}
	if tr == nil {
		// 如果获取的传输实例为 nil，则记录错误信息。
		t.Errorf("expected:%v got:%v", nil, tr)
	}

	// 将获取的传输实例转换为 mockTransport 类型。
	mtr, ok := tr.(*mockTransport)
	if !ok {
		// 如果转换失败，则记录错误信息。
		t.Errorf("expected:%v got:%v", true, ok)
	}
	if mtr == nil {
		// 如果转换后的 mockTransport 实例为 nil，则记录错误信息。
		t.Fatalf("expected:%v got:%v", nil, mtr)
	}

	// 检查 mockTransport 实例的类型是否正确。
	if mtr.Kind().String() != KindGRPC.String() {
		// 如果类型不正确，则记录错误信息。
		t.Errorf("expected:%v got:%v", KindGRPC.String(), mtr.Kind().String())
	}

	// 检查 mockTransport 实例的端点是否正确。
	if !reflect.DeepEqual(mtr.endpoint, "test_endpoint") {
		// 如果端点不正确，则记录错误信息。
		t.Errorf("expected:%v got:%v", "test_endpoint", mtr.endpoint)
	}
}

// TestClientTransport 测试客户端传输的功能。
func TestClientTransport(t *testing.T) {
	ctx := context.Background()

	// 创建一个新的客户端上下文，并将 mockTransport 实例作为值传递给它。
	ctx = NewClientContext(ctx, &mockTransport{endpoint: "test_endpoint"})

	// 从客户端上下文中获取传输实例。
	tr, ok := FromClientContext(ctx)
	if !ok {
		// 如果获取传输实例失败，则记录错误信息。
		t.Errorf("expected:%v got:%v", true, ok)
	}
	if tr == nil {
		// 如果获取的传输实例为 nil，则记录错误信息。
		t.Errorf("expected:%v got:%v", nil, tr)
	}

	// 将获取的传输实例转换为 mockTransport 类型。
	mtr, ok := tr.(*mockTransport)
	if !ok {
		t.Errorf("expected:%v got:%v", true, ok)
		return
	}
	if mtr == nil {
		t.Errorf("expected:%v got:%v", nil, mtr)
		return
	}
	// 检查 mockTransport 实例的端点是否正确。
	if !reflect.DeepEqual(mtr.endpoint, "test_endpoint") {
		// 如果端点不正确，则记录错误信息。
		t.Errorf("expected:%v got:%v", "test_endpoint", mtr.endpoint)
	}
}
