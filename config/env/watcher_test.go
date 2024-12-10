package env

import (
	"testing"
)

// Test_watcher_next 测试 watcher 的 Next 方法
func Test_watcher_next(t *testing.T) {
	// 测试用例：在停止后调用 Next 方法应该返回错误
	t.Run("next after stop should return err", func(t *testing.T) {
		// 创建一个新的 watcher
		w, err := NewWatcher()
		// 检查是否有错误
		if err != nil {
			// 如果有错误，记录错误信息
			t.Errorf("expect no error, got %v", err)
		}

		// 停止 watcher
		_ = w.Stop()
		// 调用 Next 方法
		_, err = w.Next()
		// 检查是否有错误
		if err == nil {
			// 如果没有错误，记录错误信息
			t.Error("expect error, actual nil")
		}
	})
}

// Test_watcher_stop 测试 watcher 的 Stop 方法
func Test_watcher_stop(t *testing.T) {
	// 测试用例：多次调用 Stop 方法不应该导致 panic
	t.Run("stop multiple times should not panic", func(t *testing.T) {
		// 创建一个新的 watcher
		w, err := NewWatcher()
		// 检查是否有错误
		if err != nil {
			// 如果有错误，记录错误信息
			t.Errorf("expect no error, got %v", err)
		}

		// 停止 watcher
		_ = w.Stop()
		// 再次停止 watcher
		_ = w.Stop()
	})
}
