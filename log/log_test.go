package log

import (
	"testing"
)

// TestInfo 测试日志记录器的基本功能
func TestInfo(t *testing.T) {
	// 获取默认的日志记录器
	logger := DefaultLogger
	// 使用 With 函数为日志记录器添加额外的字段，如时间戳和调用者信息
	logger = With(logger, "ts", DefaultTimestamp)
	logger = With(logger, "caller", DefaultCaller)
	// 使用日志记录器记录一条信息级别为 LevelInfo 的日志
	_ = logger.Log(LevelInfo, "key1", "value1")
}
