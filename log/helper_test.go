package log

import (
	"context"
	"io"
	"os"
	"testing"
)

// TestHelper 测试 Helper 的基本功能。
func TestHelper(_ *testing.T) {
	// 创建一个带有默认配置的日志记录器
	logger := With(
		DefaultLogger,
		"ts", DefaultTimestamp, // 时间戳
		"caller", DefaultCaller, // 调用者信息
		"module", "test", // 模块名称
	)
	// 创建日志辅助工具
	log := NewHelper(logger)

	// 使用不同的日志级别输出日志
	log.Log(LevelDebug, "msg", "test debug") // 直接输出 debug 级别日志
	log.Debug("test debug")                  // 使用 Debug 方法输出日志
	log.Debugf("test %s", "debug")           // 格式化输出 Debug 日志
	log.Debugw("log", "test debug")          // 使用键值对输出 Debug 日志

	log.Warn("test warn")         // 输出 warn 日志
	log.Warnf("test %s", "warn")  // 格式化输出 warn 日志
	log.Warnw("log", "test warn") // 使用键值对输出 warn 日志

	// 创建子日志记录器
	subLogger := With(log.Logger(),
		"module", "sub", // 子模块名称
	)
	subLog := NewHelper(subLogger)
	subLog.Infof("sub logger test with level %s", "info") // 使用子日志记录器输出 info 日志
}

// TestHelperWithMsgKey 测试自定义消息键名。
func TestHelperWithMsgKey(_ *testing.T) {
	// 创建带默认配置的日志记录器
	logger := With(DefaultLogger, "ts", DefaultTimestamp, "caller", DefaultCaller)
	// 创建日志辅助工具，并自定义消息键名为 "message"
	log := NewHelper(logger, WithMessageKey("message"))
	log.Debugf("test %s", "debug")  // 格式化输出 Debug 日志
	log.Debugw("log", "test debug") // 使用键值对输出 Debug 日志
}

// TestHelperLevel 测试 Helper 不同日志级别的功能。
func TestHelperLevel(_ *testing.T) {
	// 创建默认日志记录器
	log := NewHelper(DefaultLogger)
	log.Debug("test debug")         // Debug 级别日志
	log.Info("test info")           // Info 级别日志
	log.Infof("test %s", "info")    // 格式化输出 Info 日志
	log.Warn("test warn")           // Warn 级别日志
	log.Error("test error")         // Error 级别日志
	log.Errorf("test %s", "error")  // 格式化输出 Error 日志
	log.Errorw("log", "test error") // 使用键值对输出 Error 日志
}

// BenchmarkHelperPrint 测试 Helper 打印性能。
func BenchmarkHelperPrint(b *testing.B) {
	// 使用标准日志记录器，将日志丢弃（不输出）
	log := NewHelper(NewStdLogger(io.Discard))
	for i := 0; i < b.N; i++ {
		log.Debug("test") // 测试 Debug 日志性能
	}
}

// BenchmarkHelperPrintFilterLevel 测试带日志级别过滤的 Helper 打印性能。
func BenchmarkHelperPrintFilterLevel(b *testing.B) {
	// 创建带日志级别过滤的日志记录器
	log := NewHelper(NewFilter(NewStdLogger(io.Discard), FilterLevel(LevelDebug)))
	for i := 0; i < b.N; i++ {
		log.Debug("test") // 测试 Debug 日志性能
	}
}

// BenchmarkHelperPrintf 测试 Helper 格式化日志的性能。
func BenchmarkHelperPrintf(b *testing.B) {
	// 使用标准日志记录器，将日志丢弃（不输出）
	log := NewHelper(NewStdLogger(io.Discard))
	for i := 0; i < b.N; i++ {
		log.Debugf("%s", "test") // 测试格式化 Debug 日志性能
	}
}

// BenchmarkHelperPrintfFilterLevel 测试带日志级别过滤的 Helper 格式化日志性能。
func BenchmarkHelperPrintfFilterLevel(b *testing.B) {
	// 创建带日志级别过滤的日志记录器（过滤级别为 Info）
	log := NewHelper(NewFilter(NewStdLogger(io.Discard), FilterLevel(LevelInfo)))
	for i := 0; i < b.N; i++ {
		log.Debugf("%s", "test") // 测试格式化 Debug 日志性能
	}
}

// BenchmarkHelperPrintw 测试键值对日志输出的性能。
func BenchmarkHelperPrintw(b *testing.B) {
	// 使用标准日志记录器，将日志丢弃（不输出）
	log := NewHelper(NewStdLogger(io.Discard))
	for i := 0; i < b.N; i++ {
		log.Debugw("key", "value") // 测试键值对 Debug 日志性能
	}
}

// traceKey 是用于上下文中的键类型。
type traceKey struct{}

// TestContext 测试日志工具与上下文的结合使用。
func TestContext(_ *testing.T) {
	// 创建一个带 Trace 功能的日志记录器
	logger := With(NewStdLogger(os.Stdout),
		"trace", Trace(), // 添加 Trace 键值到日志输出中
	)
	log := NewHelper(logger)

	// 创建带 traceKey 的上下文
	ctx := context.WithValue(context.Background(), traceKey{}, "2233")
	log.WithContext(ctx).Info("got trace!") // 使用带上下文的日志工具输出日志
}

// Trace 是一个返回上下文中 trace 值的函数。
func Trace() Valuer {
	return func(ctx context.Context) interface{} {
		// 从上下文中获取 traceKey 的值
		s, ok := ctx.Value(traceKey{}).(string)
		if !ok {
			return nil
		}
		return s
	}
}
