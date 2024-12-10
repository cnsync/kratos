package log

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"strings"
	"testing"
)

func TestGlobalLog(t *testing.T) {
	// 创建一个字节缓冲区，用于捕获日志输出
	buffer := &bytes.Buffer{}
	// 创建一个新的标准日志记录器，将日志输出到缓冲区
	logger := NewStdLogger(buffer)
	// 设置全局日志记录器为新创建的日志记录器
	SetLogger(logger)

	// 检查全局日志记录器是否与设置的日志记录器相等
	if global.Logger != logger {
		// 如果不相等，记录错误信息
		t.Error("GetLogger() is not equal to logger")
	}

	// 定义一个测试用例结构体，包含日志级别和内容
	testCases := []struct {
		level   Level
		content []interface{}
	}{
		{
			// 调试级别
			LevelDebug,
			// 日志内容
			[]interface{}{"test debug"},
		},
		{
			// 信息级别
			LevelInfo,
			// 日志内容
			[]interface{}{"test info"},
		},
		{
			// 信息级别
			LevelInfo,
			// 日志内容，包含格式化字符串
			[]interface{}{"test %s", "info"},
		},
		{
			// 警告级别
			LevelWarn,
			// 日志内容
			[]interface{}{"test warn"},
		},
		{
			// 错误级别
			LevelError,
			// 日志内容
			[]interface{}{"test error"},
		},
		{
			// 错误级别
			LevelError,
			// 日志内容，包含格式化字符串
			[]interface{}{"test %s", "error"},
		},
	}

	// 初始化一个切片，用于存储预期的日志输出
	var expected []string
	// 遍历测试用例
	for _, tc := range testCases {
		// 将日志内容格式化为字符串
		msg := fmt.Sprintf(tc.content[0].(string), tc.content[1:]...)
		// 根据日志级别进行处理
		switch tc.level {
		case LevelDebug:
			// 记录调试级别的日志
			Debug(msg)
			// 将预期的日志输出添加到切片中
			expected = append(expected, fmt.Sprintf("%s msg=%s", "DEBUG", msg))
			// 记录格式化的调试级别的日志
			Debugf(tc.content[0].(string), tc.content[1:]...)
			// 将预期的日志输出添加到切片中
			expected = append(expected, fmt.Sprintf("%s msg=%s", "DEBUG", msg))
			// 记录带有键值对的调试级别的日志
			Debugw("log", msg)
			// 将预期的日志输出添加到切片中
			expected = append(expected, fmt.Sprintf("%s log=%s", "DEBUG", msg))
		case LevelInfo:
			// 记录信息级别的日志
			Info(msg)
			// 将预期的日志输出添加到切片中
			expected = append(expected, fmt.Sprintf("%s msg=%s", "INFO", msg))
			// 记录格式化的信息级别的日志
			Infof(tc.content[0].(string), tc.content[1:]...)
			// 将预期的日志输出添加到切片中
			expected = append(expected, fmt.Sprintf("%s msg=%s", "INFO", msg))
			// 记录带有键值对的信息级别的日志
			Infow("log", msg)
			// 将预期的日志输出添加到切片中
			expected = append(expected, fmt.Sprintf("%s log=%s", "INFO", msg))
		case LevelWarn:
			// 记录警告级别的日志
			Warn(msg)
			// 将预期的日志输出添加到切片中
			expected = append(expected, fmt.Sprintf("%s msg=%s", "WARN", msg))
			// 记录格式化的警告级别的日志
			Warnf(tc.content[0].(string), tc.content[1:]...)
			// 将预期的日志输出添加到切片中
			expected = append(expected, fmt.Sprintf("%s msg=%s", "WARN", msg))
			// 记录带有键值对的警告级别的日志
			Warnw("log", msg)
			// 将预期的日志输出添加到切片中
			expected = append(expected, fmt.Sprintf("%s log=%s", "WARN", msg))
		case LevelError:
			// 记录错误级别的日志
			Error(msg)
			// 将预期的日志输出添加到切片中
			expected = append(expected, fmt.Sprintf("%s msg=%s", "ERROR", msg))
			// 记录格式化的错误级别的日志
			Errorf(tc.content[0].(string), tc.content[1:]...)
			// 将预期的日志输出添加到切片中
			expected = append(expected, fmt.Sprintf("%s msg=%s", "ERROR", msg))
			// 记录带有键值对的错误级别的日志
			Errorw("log", msg)
			// 将预期的日志输出添加到切片中
			expected = append(expected, fmt.Sprintf("%s log=%s", "ERROR", msg))
		}
	}
	// 记录一条信息级别的日志
	Log(LevelInfo, DefaultMessageKey, "test log")
	// 将预期的日志输出添加到切片中
	expected = append(expected, fmt.Sprintf("%s msg=%s", "INFO", "test log"))

	// 在预期的日志输出后添加一个空行
	expected = append(expected, "")

	// 打印缓冲区中的内容
	t.Logf("Content: %s", buffer.String())
	// 检查缓冲区中的内容是否与预期的日志输出一致
	if buffer.String() != strings.Join(expected, "\n") {
		// 如果不一致，记录错误信息
		t.Errorf("Expected: %s, got: %s", strings.Join(expected, "\n"), buffer.String())
	}
}

func TestGlobalLogUpdate(t *testing.T) {
	// 创建一个新的日志记录器设备
	l := &loggerAppliance{}
	// 设置日志记录器设备的日志记录器为标准输出日志记录器
	l.SetLogger(NewStdLogger(os.Stdout))
	// 获取全局日志记录器的帮助器
	LOG := NewHelper(l)
	// 使用帮助器记录一条信息到标准输出
	LOG.Info("Log to stdout")

	// 创建一个字节缓冲区
	buffer := &bytes.Buffer{}
	// 设置日志记录器设备的日志记录器为缓冲区日志记录器
	l.SetLogger(NewStdLogger(buffer))
	// 使用帮助器记录一条信息到缓冲区
	LOG.Info("Log to buffer")

	// 预期的日志输出
	expected := "INFO msg=Log to buffer\n"
	// 检查缓冲区中的内容是否与预期的日志输出一致
	if buffer.String() != expected {
		// 如果不一致，记录错误信息
		t.Errorf("Expected: %s, got: %s", expected, buffer.String())
	}
}

func TestGlobalContext(t *testing.T) {
	// 创建一个字节缓冲区，用于捕获日志输出
	buffer := &bytes.Buffer{}
	// 设置全局日志记录器为新创建的日志记录器
	SetLogger(NewStdLogger(buffer))
	// 使用上下文记录器记录一条格式化的信息
	Context(context.Background()).Infof("111")
	// 检查缓冲区中的内容是否与预期的日志输出一致
	if buffer.String() != "INFO msg=111\n" {
		// 如果不一致，记录错误信息
		t.Errorf("Expected:%s, got:%s", "INFO msg=111", buffer.String())
	}
}

func TestContextWithGlobalLog(t *testing.T) {
	// 创建一个字节缓冲区，用于捕获日志输出
	buffer := &bytes.Buffer{}

	// 定义一个 traceKey 类型，用于在上下文中存储和检索 trace-id
	type traceKey struct{}

	// 设置 "trace-id" 值提供器
	// 这个值提供器会从上下文中提取 trace-id 的值
	newLogger := With(NewStdLogger(buffer), "trace-id", Valuer(func(ctx context.Context) interface{} {
		// 从上下文中获取 traceKey 对应的值
		return ctx.Value(traceKey{})
	}))

	// 设置全局日志记录器为新创建的日志记录器
	SetLogger(newLogger)

	// 向上下文中添加 trace-id 的值
	ctx := context.WithValue(context.Background(), traceKey{}, "test-trace-id")

	// 使用带有上下文的日志记录器记录一条信息级别的日志
	// 这条日志应该包含 trace-id 的值
	_ = WithContext(ctx, GetLogger()).Log(LevelInfo)

	// 预期的日志输出格式
	expected := "INFO trace-id=test-trace-id\n"

	// 检查缓冲区中的内容是否与预期的日志输出一致
	if buffer.String() != expected {
		// 如果不一致，记录错误信息
		t.Errorf("Expected:%s, got:%s", expected, buffer.String())
	}
}
