package errors

import (
	"errors"
	"fmt"
	"testing"
)

// mockErr 是一个自定义错误类型，用于测试错误包装和 unwrap。
type mockErr struct{}

// Error 方法实现了 error 接口，返回 mockErr 的错误消息。
func (*mockErr) Error() string {
	return "mock error"
}

// TestWarp 函数测试了错误包装和 unwrap 的功能。
func TestWarp(t *testing.T) {
	// 创建一个 mockErr 类型的错误实例。
	var err error = &mockErr{}
	// 使用 fmt.Errorf 函数将 err 包装成一个新的错误 err2。
	err2 := fmt.Errorf("wrap %w", err)
	// 检查 err 是否是 err2 的 unwrap 结果。
	if !errors.Is(err, Unwrap(err2)) {
		// 如果不是，记录错误信息。
		t.Errorf("got %v want: %v", err, Unwrap(err2))
	}
	// 检查 err2 是否是 err 的包装错误。
	if !Is(err2, err) {
		// 如果不是，记录错误信息。
		t.Errorf("Is(err2, err) got %v want: %v", Is(err2, err), true)
	}
	// 创建另一个 mockErr 类型的错误实例 err3。
	err3 := &mockErr{}
	// 检查 err2 是否可以转换为 *mockErr 类型，并赋值给 err3。
	if !As(err2, &err3) {
		// 如果不是，记录错误信息。
		t.Errorf("As(err2, &err3) got %v want: %v", As(err2, &err3), true)
	}
}
