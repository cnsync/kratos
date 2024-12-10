package log

import (
	"context"
	"testing"
)

// TestValue 测试 Value 函数
func TestValue(t *testing.T) {
	// 获取默认的日志记录器
	logger := DefaultLogger
	// 使用 With 函数为日志记录器添加额外的字段，如时间戳和调用者信息
	logger = With(logger, "ts", DefaultTimestamp, "caller", DefaultCaller)
	// 使用日志记录器记录一条信息级别为 LevelInfo 的日志
	_ = logger.Log(LevelInfo, "msg", "helloworld")

	// 再次获取默认的日志记录器
	logger = DefaultLogger
	// 使用 With 函数为日志记录器添加额外的字段，但不指定任何字段
	logger = With(logger)
	// 使用日志记录器记录一条调试级别为 LevelDebug 的日志
	_ = logger.Log(LevelDebug, "msg", "helloworld")

	// 定义一个空的接口变量 v1
	var v1 interface{}
	// 调用 Value 函数，传入上下文和 v1 变量
	got := Value(context.Background(), v1)
	// 检查返回值是否与 v1 变量相同
	if got != v1 {
		// 如果返回值与 v1 变量不同，记录错误信息
		t.Errorf("Value() = %v, want %v", got, v1)
	}

	// 定义一个 Valuer 类型的变量 v2，该变量返回一个整数值 3
	var v2 Valuer = func(context.Context) interface{} {
		return 3
	}
	// 调用 Value 函数，传入上下文和 v2 变量
	got = Value(context.Background(), v2)
	// 将返回值转换为整数类型
	res := got.(int)
	// 检查返回值是否为 3
	if res != 3 {
		// 如果返回值不是 3，记录错误信息
		t.Errorf("Value() = %v, want %v", res, 3)
	}
}
