package log

import "strings"

// Level 是日志记录器的级别。
type Level int8

// LevelKey 是日志记录器级别的键。
const LevelKey = "level"

const (
	// LevelDebug 是日志记录器的调试级别。
	LevelDebug Level = iota - 1
	// LevelInfo 是日志记录器的信息级别。
	LevelInfo
	// LevelWarn 是日志记录器的警告级别。
	LevelWarn
	// LevelError 是日志记录器的错误级别。
	LevelError
	// LevelFatal 是日志记录器的致命级别。
	LevelFatal
)

// Key 方法返回日志记录器级别的键。
func (l Level) Key() string {
	return LevelKey
}

// String 方法返回日志记录器级别的字符串表示。
func (l Level) String() string {
	switch l {
	case LevelDebug:
		return "DEBUG"
	case LevelInfo:
		return "INFO"
	case LevelWarn:
		return "WARN"
	case LevelError:
		return "ERROR"
	case LevelFatal:
		return "FATAL"
	default:
		return ""
	}
}

// ParseLevel 方法将一个级别字符串解析为日志记录器的 Level 值。
func ParseLevel(s string) Level {
	switch strings.ToUpper(s) {
	case "DEBUG":
		return LevelDebug
	case "INFO":
		return LevelInfo
	case "WARN":
		return LevelWarn
	case "ERROR":
		return LevelError
	case "FATAL":
		return LevelFatal
	}
	return LevelInfo
}
