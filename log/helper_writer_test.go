package log

import (
	"bytes"
	"io"
	"strings"
	"testing"
)

func TestWriterWrapper(t *testing.T) {
	// 定义一个 bytes.Buffer 类型的变量 buf，用于存储日志输出
	var buf bytes.Buffer
	// 创建一个新的标准日志记录器 logger，将日志输出到 buf 中
	logger := NewStdLogger(&buf)
	// 定义一个测试用的日志消息内容
	content := "ThisIsSomeTestLogMessage"
	// 定义一个测试用例结构体，包含 io.Writer 接口的实现、期望的日志级别和期望的消息键
	testCases := []struct {
		w                io.Writer
		acceptLevel      Level
		acceptMessageKey string
	}{
		// 默认情况下，使用默认的日志级别和消息键
		{
			w:                NewWriter(logger),
			acceptLevel:      LevelInfo, // default level
			acceptMessageKey: DefaultMessageKey,
		},
		// 自定义日志级别为 LevelDebug
		{
			w:                NewWriter(logger, WithWriterLevel(LevelDebug)),
			acceptLevel:      LevelDebug,
			acceptMessageKey: DefaultMessageKey,
		},
		// 自定义消息键为 "XxXxX"
		{
			w:                NewWriter(logger, WithWriteMessageKey("XxXxX")),
			acceptLevel:      LevelInfo, // default level
			acceptMessageKey: "XxXxX",
		},
		// 同时自定义日志级别为 LevelError 和消息键为 "XxXxX"
		{
			w:                NewWriter(logger, WithWriterLevel(LevelError), WithWriteMessageKey("XxXxX")),
			acceptLevel:      LevelError,
			acceptMessageKey: "XxXxX",
		},
	}
	// 遍历测试用例
	for _, tc := range testCases {
		// 将 content 写入到 tc.w 中，即调用 writerWrapper 的 Write 方法
		_, err := tc.w.Write([]byte(content))
		// 如果发生错误，记录错误并终止测试
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		// 检查 buf 中的字符串是否包含期望的日志级别
		if !strings.Contains(buf.String(), tc.acceptLevel.String()) {
			// 如果不包含，记录错误信息
			t.Errorf("expected level: %s, got: %s", tc.acceptLevel, buf.String())
		}
		// 检查 buf 中的字符串是否包含期望的消息键
		if !strings.Contains(buf.String(), tc.acceptMessageKey) {
			// 如果不包含，记录错误信息
			t.Errorf("expected message key: %s, got: %s", tc.acceptMessageKey, buf.String())
		}
	}
}
