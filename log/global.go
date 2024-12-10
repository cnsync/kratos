package log

import (
	"context"
	"fmt"
	"os"
	"sync"
)

// globalLogger 设置为当前进程中的全局日志记录器。
var global = &loggerAppliance{}

// loggerAppliance 是 `Logger` 的代理，
// 使得日志记录器的更改会影响所有子日志记录器。
type loggerAppliance struct {
	lock sync.RWMutex
	Logger
}

func init() {
	// 初始化时设置默认日志记录器
	global.SetLogger(DefaultLogger)
}

func (a *loggerAppliance) SetLogger(in Logger) {
	// 设置日志记录器，此操作不是线程安全的
	a.lock.Lock()
	defer a.lock.Unlock()
	a.Logger = in
}

// SetLogger 应该在任何其他日志调用之前调用。
// 并且它不是线程安全的。
func SetLogger(logger Logger) {
	global.SetLogger(logger)
}

// GetLogger 返回当前进程中的全局日志记录器。
func GetLogger() Logger {
	global.lock.RLock()
	defer global.lock.RUnlock()
	return global.Logger
}

// Log 根据级别和键值对打印日志。
func Log(level Level, keyvals ...interface{}) {
	// 记录日志，忽略返回值
	_ = global.Log(level, keyvals...)
}

// Context 带有上下文日志记录器。
func Context(ctx context.Context) *Helper {
	return NewHelper(WithContext(ctx, global.Logger))
}

// Debug 记录调试级别的日志。
func Debug(a ...interface{}) {
	// 记录日志，忽略返回值
	_ = global.Log(LevelDebug, DefaultMessageKey, fmt.Sprint(a...))
}

// Debugf 记录格式化的调试级别的日志。
func Debugf(format string, a ...interface{}) {
	// 记录日志，忽略返回值
	_ = global.Log(LevelDebug, DefaultMessageKey, fmt.Sprintf(format, a...))
}

// Debugw 记录带有键值对的调试级别的日志。
func Debugw(keyvals ...interface{}) {
	// 记录日志，忽略返回值
	_ = global.Log(LevelDebug, keyvals...)
}

// Info 记录信息级别的日志。
func Info(a ...interface{}) {
	// 记录日志，忽略返回值
	_ = global.Log(LevelInfo, DefaultMessageKey, fmt.Sprint(a...))
}

// Infof 记录格式化的信息级别的日志。
func Infof(format string, a ...interface{}) {
	// 记录日志，忽略返回值
	_ = global.Log(LevelInfo, DefaultMessageKey, fmt.Sprintf(format, a...))
}

// Infow 记录带有键值对的信息级别的日志。
func Infow(keyvals ...interface{}) {
	// 记录日志，忽略返回值
	_ = global.Log(LevelInfo, keyvals...)
}

// Warn 记录警告级别的日志。
func Warn(a ...interface{}) {
	// 记录日志，忽略返回值
	_ = global.Log(LevelWarn, DefaultMessageKey, fmt.Sprint(a...))
}

// Warnf 记录格式化的警告级别的日志。
func Warnf(format string, a ...interface{}) {
	// 记录日志，忽略返回值
	_ = global.Log(LevelWarn, DefaultMessageKey, fmt.Sprintf(format, a...))
}

// Warnw 记录带有键值对的警告级别的日志。
func Warnw(keyvals ...interface{}) {
	// 记录日志，忽略返回值
	_ = global.Log(LevelWarn, keyvals...)
}

// Error 记录错误级别的日志。
func Error(a ...interface{}) {
	// 记录日志，忽略返回值
	_ = global.Log(LevelError, DefaultMessageKey, fmt.Sprint(a...))
}

// Errorf 记录格式化的错误级别的日志。
func Errorf(format string, a ...interface{}) {
	// 记录日志，忽略返回值
	_ = global.Log(LevelError, DefaultMessageKey, fmt.Sprintf(format, a...))
}

// Errorw 记录带有键值对的错误级别的日志。
func Errorw(keyvals ...interface{}) {
	// 记录日志，忽略返回值
	_ = global.Log(LevelError, keyvals...)
}

// Fatal 记录致命级别的日志并退出程序。
func Fatal(a ...interface{}) {
	// 记录日志，忽略返回值
	_ = global.Log(LevelFatal, DefaultMessageKey, fmt.Sprint(a...))
	// 退出程序
	os.Exit(1)
}

// Fatalf 记录格式化的致命级别的日志并退出程序。
func Fatalf(format string, a ...interface{}) {
	// 记录日志，忽略返回值
	_ = global.Log(LevelFatal, DefaultMessageKey, fmt.Sprintf(format, a...))
	// 退出程序
	os.Exit(1)
}

// Fatalw 记录带有键值对的致命级别的日志并退出程序。
func Fatalw(keyvals ...interface{}) {
	// 记录日志，忽略返回值
	_ = global.Log(LevelFatal, keyvals...)
	// 退出程序
	os.Exit(1)
}
