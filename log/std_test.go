package log

import (
	"bytes"
	"testing"

	"golang.org/x/sync/errgroup"
)

// TestStdLogger 测试标准日志记录器的基本功能
func TestStdLogger(_ *testing.T) {
	// 获取默认的日志记录器
	logger := DefaultLogger
	// 使用 With 函数为日志记录器添加额外的字段，如时间戳和调用者信息
	logger = With(logger, "caller", DefaultCaller, "ts", DefaultTimestamp)
	// 使用日志记录器记录一条信息级别为 LevelInfo 的日志
	_ = logger.Log(LevelInfo, "msg", "test debug")
	// 使用日志记录器记录一条信息级别为 LevelInfo 的日志
	_ = logger.Log(LevelInfo, "msg", "test info")
	// 使用日志记录器记录一条信息级别为 LevelInfo 的日志
	_ = logger.Log(LevelInfo, "msg", "test warn")
	// 使用日志记录器记录一条信息级别为 LevelInfo 的日志
	_ = logger.Log(LevelInfo, "msg", "test error")
	// 使用日志记录器记录一条调试级别为 LevelDebug 的日志
	_ = logger.Log(LevelDebug, "singular")

	// 再次获取默认的日志记录器
	logger2 := DefaultLogger
	// 使用日志记录器记录一条调试级别为 LevelDebug 的日志
	_ = logger2.Log(LevelDebug)
}

// TestStdLogger_Log 测试标准日志记录器的并发日志记录功能
func TestStdLogger_Log(t *testing.T) {
	// 创建一个 bytes.Buffer 用于捕获日志输出
	var b bytes.Buffer
	// 使用 NewStdLogger 函数创建一个新的标准日志记录器，并将其输出设置为捕获的缓冲区
	logger := NewStdLogger(&b)

	// 创建一个 errgroup.Group 用于并发执行多个函数
	var eg errgroup.Group
	// 启动两个并发的 goroutine，每个 goroutine 都调用日志记录器的 Log 方法记录一条日志
	eg.Go(func() error { return logger.Log(LevelInfo, "msg", "a", "k", "v") })
	eg.Go(func() error { return logger.Log(LevelInfo, "msg", "a", "k", "v") })

	// 等待所有 goroutine 完成，并检查是否有错误发生
	err := eg.Wait()
	// 如果有错误发生，记录错误信息并终止测试
	if err != nil {
		t.Fatalf("log error: %v", err)
	}

	// 检查捕获的日志输出是否与预期匹配
	if s := b.String(); s != "INFO msg=a k=v\nINFO msg=a k=v\n" {
		// 如果日志输出不匹配预期，记录错误信息并终止测试
		t.Fatalf("log not match: %q", s)
	}
}
