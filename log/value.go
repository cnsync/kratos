package log

import (
	"context"
	"runtime"
	"strconv"
	"strings"
	"time"
)

var (
	// DefaultCaller 是一个 Valuer，它返回调用者的文件和行号。
	DefaultCaller = Caller(4)

	// DefaultTimestamp 是一个 Valuer，它返回当前的时间戳。
	DefaultTimestamp = Timestamp(time.RFC3339)
)

// Valuer 是一个函数类型，它接受一个 context.Context 参数并返回一个 interface{} 类型的值。
type Valuer func(ctx context.Context) interface{}

// Value 函数接受一个 context.Context 和一个 interface{} 类型的值，如果这个值是一个 Valuer 类型，它将调用这个 Valuer 并返回其结果，否则直接返回这个值。
func Value(ctx context.Context, v interface{}) interface{} {
	if v, ok := v.(Valuer); ok {
		return v(ctx)
	}
	return v
}

// Caller 函数返回一个 Valuer，这个 Valuer 会返回调用者的文件名和行号。
func Caller(depth int) Valuer {
	return func(context.Context) interface{} {
		// 获取调用者的信息
		_, file, line, _ := runtime.Caller(depth)
		// 找到文件名中最后一个 / 的位置
		idx := strings.LastIndexByte(file, '/')
		if idx == -1 {
			// 如果没有找到 /，则直接返回文件名和行号
			return file[idx+1:] + ":" + strconv.Itoa(line)
		}
		// 找到文件名中倒数第二个 / 的位置
		idx = strings.LastIndexByte(file[:idx], '/')
		// 返回文件名和行号
		return file[idx+1:] + ":" + strconv.Itoa(line)
	}
}

// Timestamp 函数返回一个 Valuer，这个 Valuer 会返回当前的时间戳。
func Timestamp(layout string) Valuer {
	return func(context.Context) interface{} {
		// 返回当前时间的格式化字符串
		return time.Now().Format(layout)
	}
}

// bindValues 函数遍历一个键值对切片，并将其中的 Valuer 类型的值替换为其调用结果。
func bindValues(ctx context.Context, keyvals []interface{}) {
	for i := 1; i < len(keyvals); i += 2 {
		if v, ok := keyvals[i].(Valuer); ok {
			keyvals[i] = v(ctx)
		}
	}
}

// containsValuer 函数检查一个键值对切片中是否包含 Valuer 类型的值。
func containsValuer(keyvals []interface{}) bool {
	for i := 1; i < len(keyvals); i += 2 {
		if _, ok := keyvals[i].(Valuer); ok {
			return true
		}
	}
	return false
}
