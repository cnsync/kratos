package log

import (
	"bytes"
	"fmt"
	"io"
	"sync"
)

var _ Logger = (*stdLogger)(nil)

// stdLogger 对应于标准库的 [log.Logger]，并提供类似的功能。
// 它还可以被多个 goroutine 同时使用。
type stdLogger struct {
	// w 是日志写入器
	w io.Writer
	// isDiscard 标记是否为丢弃写入器
	isDiscard bool
	// mu 用于同步对日志写入器的访问
	mu sync.Mutex
	// pool 是一个用于缓存 bytes.Buffer 的对象池
	pool *sync.Pool
}

// NewStdLogger 使用指定的写入器创建一个新的标准日志记录器。
func NewStdLogger(w io.Writer) Logger {
	return &stdLogger{
		w:         w,
		isDiscard: w == io.Discard,
		pool: &sync.Pool{
			New: func() interface{} {
				return new(bytes.Buffer)
			},
		},
	}
}

// Log 打印键值对日志。
func (l *stdLogger) Log(level Level, keyvals ...interface{}) error {
	// 如果是丢弃写入器或没有键值对，则不进行任何操作
	if l.isDiscard || len(keyvals) == 0 {
		return nil
	}
	// 如果键值对数量为奇数，则添加一个默认值
	if (len(keyvals) & 1) == 1 {
		keyvals = append(keyvals, "KEYVALS UNPAIRED")
	}

	// 从对象池中获取一个 bytes.Buffer
	buf := l.pool.Get().(*bytes.Buffer)
	defer l.pool.Put(buf)

	// 写入日志级别
	buf.WriteString(level.String())
	// 写入键值对
	for i := 0; i < len(keyvals); i += 2 {
		_, _ = fmt.Fprintf(buf, " %s=%v", keyvals[i], keyvals[i+1])
	}
	// 写入换行符
	buf.WriteByte('\n')
	defer buf.Reset()

	// 加锁以确保线程安全
	l.mu.Lock()
	defer l.mu.Unlock()
	// 将缓冲区内容写入日志写入器
	_, err := l.w.Write(buf.Bytes())
	return err
}

// Close 关闭日志记录器。
func (l *stdLogger) Close() error {
	return nil
}
