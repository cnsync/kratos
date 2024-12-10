package log

import (
	"context"
	"log"
)

// DefaultLogger 是默认的日志记录器。
var DefaultLogger = NewStdLogger(log.Writer())

// Logger 是一个日志接口，用于定义日志记录的方法。
type Logger interface {
	// Log 方法记录日志。
	// 参数:
	// - level: 日志级别。
	// - keyvals: 键值对，表示日志的内容。
	Log(level Level, keyvals ...interface{}) error
}

// logger 是 Logger 接口的一个实现。
type logger struct {
	logger    Logger          // 实际的日志记录器。
	prefix    []interface{}   // 日志的前缀键值对。
	hasValuer bool            // 标识是否包含值函数（Valuer）。
	ctx       context.Context // 上下文，用于绑定值函数。
}

// Log 方法记录日志。
func (c *logger) Log(level Level, keyvals ...interface{}) error {
	// 创建一个新的切片来存储所有的键值对。
	kvs := make([]interface{}, 0, len(c.prefix)+len(keyvals))
	// 将前缀添加到键值对切片中。
	kvs = append(kvs, c.prefix...)
	// 如果有值函数（Valuer），则将其绑定到上下文中。
	if c.hasValuer {
		bindValues(c.ctx, kvs)
	}
	// 将传入的键值对添加到切片中。
	kvs = append(kvs, keyvals...)
	// 使用底层日志记录器记录日志。
	return c.logger.Log(level, kvs...)
}

// With 方法用于创建一个新的日志记录器，并为其添加额外的键值对。
func With(l Logger, kv ...interface{}) Logger {
	// 尝试将日志记录器转换为 logger 类型。
	c, ok := l.(*logger)
	// 如果转换失败，则创建一个新的 logger 实例。
	if !ok {
		return &logger{logger: l, prefix: kv, hasValuer: containsValuer(kv), ctx: context.Background()}
	}
	// 创建一个新的切片来存储所有的键值对。
	kvs := make([]interface{}, 0, len(c.prefix)+len(kv))
	// 将前缀添加到键值对切片中。
	kvs = append(kvs, c.prefix...)
	// 将传入的键值对添加到切片中。
	kvs = append(kvs, kv...)
	// 返回一个新的 logger 实例。
	return &logger{
		logger:    c.logger,
		prefix:    kvs,
		hasValuer: containsValuer(kvs),
		ctx:       c.ctx,
	}
}

// WithContext 方法用于创建一个新的日志记录器，并为其绑定一个上下文。
func WithContext(ctx context.Context, l Logger) Logger {
	// 根据日志记录器的类型，返回一个新的日志记录器实例。
	switch v := l.(type) {
	default:
		// 如果是默认类型，直接创建一个新的 logger 实例。
		return &logger{logger: l, ctx: ctx}
	case *logger:
		// 如果是 logger 类型，复制当前实例并设置新的上下文。
		lv := *v
		lv.ctx = ctx
		return &lv
	case *Filter:
		// 如果是 Filter 类型，递归绑定上下文到其内部的 logger。
		fv := *v
		fv.logger = WithContext(ctx, fv.logger)
		return &fv
	}
}
