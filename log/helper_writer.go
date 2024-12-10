package log

import "io"

// writerWrapper 是一个 io.Writer 接口的实现，它将写入的数据作为日志记录下来。
type writerWrapper struct {
	helper *Helper // 用于记录日志的帮助器
	level  Level   // 日志级别
}

// WriterOptionFn 是一个函数类型，用于设置 writerWrapper 的选项。
type WriterOptionFn func(w *writerWrapper)

// WithWriterLevel 设置 writerWrapper 的日志级别。
func WithWriterLevel(level Level) WriterOptionFn {
	return func(w *writerWrapper) {
		w.level = level
	}
}

// WithWriteMessageKey 设置 writerWrapper 的帮助器的消息键。
func WithWriteMessageKey(key string) WriterOptionFn {
	return func(w *writerWrapper) {
		w.helper.msgKey = key
	}
}

// NewWriter 创建一个新的 writerWrapper 实例。
func NewWriter(logger Logger, opts ...WriterOptionFn) io.Writer {
	ww := &writerWrapper{
		helper: NewHelper(logger, WithMessageKey(DefaultMessageKey)),
		level:  LevelInfo, // 默认级别为信息
	}
	for _, opt := range opts {
		opt(ww)
	}
	return ww
}

// Write 将数据写入到 writerWrapper 中，并将其作为日志记录下来。
func (ww *writerWrapper) Write(p []byte) (int, error) {
	ww.helper.Log(ww.level, ww.helper.msgKey, string(p))
	return 0, nil
}
