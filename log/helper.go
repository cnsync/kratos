package log

import (
	"context"
	"fmt"
	"os"
)

// DefaultMessageKey 默认的消息键名。
var DefaultMessageKey = "msg"

// Option 是用于配置 Helper 的选项类型。
type Option func(*Helper)

// Helper 是一个日志工具类，用于简化日志操作。
type Helper struct {
	logger  Logger                                       // 日志记录器接口
	msgKey  string                                       // 消息键名
	sprint  func(...interface{}) string                  // 格式化函数：将参数拼接成字符串
	sprintf func(format string, a ...interface{}) string // 格式化函数：根据格式化字符串生成日志内容
}

// WithMessageKey 设置自定义消息键名。
func WithMessageKey(k string) Option {
	return func(opts *Helper) {
		opts.msgKey = k
	}
}

// WithSprint 设置自定义的 sprint 函数，用于拼接日志消息。
func WithSprint(sprint func(...interface{}) string) Option {
	return func(opts *Helper) {
		opts.sprint = sprint
	}
}

// WithSprintf 设置自定义的 sprintf 函数，用于格式化日志消息。
func WithSprintf(sprintf func(format string, a ...interface{}) string) Option {
	return func(opts *Helper) {
		opts.sprintf = sprintf
	}
}

// NewHelper 创建一个新的日志 Helper 实例。
func NewHelper(logger Logger, opts ...Option) *Helper {
	options := &Helper{
		msgKey:  DefaultMessageKey, // 默认使用 "msg" 作为消息键名
		logger:  logger,            // 指定的日志记录器
		sprint:  fmt.Sprint,        // 默认拼接函数
		sprintf: fmt.Sprintf,       // 默认格式化函数
	}
	// 应用所有传入的选项
	for _, o := range opts {
		o(options)
	}
	return options
}

// WithContext 返回一个绑定了新的上下文 (context) 的 Helper 实例。
// 提供的上下文 ctx 必须为非 nil。
func (h *Helper) WithContext(ctx context.Context) *Helper {
	return &Helper{
		msgKey:  h.msgKey,
		logger:  WithContext(ctx, h.logger), // 将新的上下文绑定到 logger 上
		sprint:  h.sprint,
		sprintf: h.sprintf,
	}
}

// Enabled 判断指定的日志级别是否启用。
// 如果日志记录器是 *Filter 类型，则根据过滤条件判断是否启用。
func (h *Helper) Enabled(level Level) bool {
	if l, ok := h.logger.(*Filter); ok {
		return level >= l.level // 仅当日志级别大于或等于过滤级别时启用
	}
	return true
}

// Logger 返回 Helper 内部的日志记录器。
func (h *Helper) Logger() Logger {
	return h.logger
}

// Log 根据日志级别和键值对输出日志。
func (h *Helper) Log(level Level, keyvals ...interface{}) {
	_ = h.logger.Log(level, keyvals...)
}

// Debug 输出 debug 级别的日志消息。
func (h *Helper) Debug(a ...interface{}) {
	if !h.Enabled(LevelDebug) {
		return
	}
	_ = h.logger.Log(LevelDebug, h.msgKey, h.sprint(a...))
}

// Debugf 格式化输出 debug 级别的日志消息。
func (h *Helper) Debugf(format string, a ...interface{}) {
	if !h.Enabled(LevelDebug) {
		return
	}
	_ = h.logger.Log(LevelDebug, h.msgKey, h.sprintf(format, a...))
}

// Debugw 输出 debug 级别的日志消息，包含键值对。
func (h *Helper) Debugw(keyvals ...interface{}) {
	_ = h.logger.Log(LevelDebug, keyvals...)
}

// Info 输出 info 级别的日志消息。
func (h *Helper) Info(a ...interface{}) {
	if !h.Enabled(LevelInfo) {
		return
	}
	_ = h.logger.Log(LevelInfo, h.msgKey, h.sprint(a...))
}

// Infof 格式化输出 info 级别的日志消息。
func (h *Helper) Infof(format string, a ...interface{}) {
	if !h.Enabled(LevelInfo) {
		return
	}
	_ = h.logger.Log(LevelInfo, h.msgKey, h.sprintf(format, a...))
}

// Infow 输出 info 级别的日志消息，包含键值对。
func (h *Helper) Infow(keyvals ...interface{}) {
	_ = h.logger.Log(LevelInfo, keyvals...)
}

// Warn 输出 warn 级别的日志消息。
func (h *Helper) Warn(a ...interface{}) {
	if !h.Enabled(LevelWarn) {
		return
	}
	_ = h.logger.Log(LevelWarn, h.msgKey, h.sprint(a...))
}

// Warnf 格式化输出 warn 级别的日志消息。
func (h *Helper) Warnf(format string, a ...interface{}) {
	if !h.Enabled(LevelWarn) {
		return
	}
	_ = h.logger.Log(LevelWarn, h.msgKey, h.sprintf(format, a...))
}

// Warnw 输出 warn 级别的日志消息，包含键值对。
func (h *Helper) Warnw(keyvals ...interface{}) {
	_ = h.logger.Log(LevelWarn, keyvals...)
}

// Error 输出 error 级别的日志消息。
func (h *Helper) Error(a ...interface{}) {
	if !h.Enabled(LevelError) {
		return
	}
	_ = h.logger.Log(LevelError, h.msgKey, h.sprint(a...))
}

// Errorf 格式化输出 error 级别的日志消息。
func (h *Helper) Errorf(format string, a ...interface{}) {
	if !h.Enabled(LevelError) {
		return
	}
	_ = h.logger.Log(LevelError, h.msgKey, h.sprintf(format, a...))
}

// Errorw 输出 error 级别的日志消息，包含键值对。
func (h *Helper) Errorw(keyvals ...interface{}) {
	_ = h.logger.Log(LevelError, keyvals...)
}

// Fatal 输出 fatal 级别的日志消息，并终止程序。
func (h *Helper) Fatal(a ...interface{}) {
	_ = h.logger.Log(LevelFatal, h.msgKey, h.sprint(a...))
	os.Exit(1) // 退出程序
}

// Fatalf 格式化输出 fatal 级别的日志消息，并终止程序。
func (h *Helper) Fatalf(format string, a ...interface{}) {
	_ = h.logger.Log(LevelFatal, h.msgKey, h.sprintf(format, a...))
	os.Exit(1) // 退出程序
}

// Fatalw 输出 fatal 级别的日志消息，包含键值对，并终止程序。
func (h *Helper) Fatalw(keyvals ...interface{}) {
	_ = h.logger.Log(LevelFatal, keyvals...)
	os.Exit(1) // 退出程序
}
