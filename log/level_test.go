package log

import "testing"

// TestLevel_Key 测试日志记录器级别的 Key 方法
func TestLevel_Key(t *testing.T) {
	// 检查 LevelInfo 级别的 Key 是否与预设的 LevelKey 一致
	if LevelInfo.Key() != LevelKey {
		// 如果不一致，记录错误信息
		t.Errorf("want: %s, got: %s", LevelKey, LevelInfo.Key())
	}
}

// TestLevel_String 测试日志记录器级别的 String 方法
func TestLevel_String(t *testing.T) {
	// 定义一个测试用例结构体，包含名称、级别和预期字符串表示
	tests := []struct {
		name string
		l    Level
		want string
	}{
		{
			// 测试用例名称：DEBUG
			name: "DEBUG",
			// 测试级别：LevelDebug
			l: LevelDebug,
			// 预期字符串表示：DEBUG
			want: "DEBUG",
		},
		{
			// 测试用例名称：INFO
			name: "INFO",
			// 测试级别：LevelInfo
			l: LevelInfo,
			// 预期字符串表示：INFO
			want: "INFO",
		},
		{
			// 测试用例名称：WARN
			name: "WARN",
			// 测试级别：LevelWarn
			l: LevelWarn,
			// 预期字符串表示：WARN
			want: "WARN",
		},
		{
			// 测试用例名称：ERROR
			name: "ERROR",
			// 测试级别：LevelError
			l: LevelError,
			// 预期字符串表示：ERROR
			want: "ERROR",
		},
		{
			// 测试用例名称：FATAL
			name: "FATAL",
			// 测试级别：LevelFatal
			l: LevelFatal,
			// 预期字符串表示：FATAL
			want: "FATAL",
		},
		{
			// 测试用例名称：other
			name: "other",
			// 测试级别：自定义级别 10
			l: 10,
			// 预期字符串表示：空字符串
			want: "",
		},
	}
	// 遍历测试用例
	for _, tt := range tests {
		// 运行每个测试用例
		t.Run(tt.name, func(t *testing.T) {
			// 检查实际字符串表示是否与预期一致
			if got := tt.l.String(); got != tt.want {
				// 如果不一致，记录错误信息
				t.Errorf("String() = %v, want %v", got, tt.want)
			}
		})
	}
}

// TestParseLevel 测试日志记录器级别的 ParseLevel 方法
func TestParseLevel(t *testing.T) {
	// 定义一个测试用例结构体，包含名称、字符串表示和预期级别
	tests := []struct {
		name string
		s    string
		want Level
	}{
		{
			// 测试用例名称：DEBUG
			name: "DEBUG",
			// 预期级别：LevelDebug
			want: LevelDebug,
			// 字符串表示：DEBUG
			s: "DEBUG",
		},
		{
			// 测试用例名称：INFO
			name: "INFO",
			// 预期级别：LevelInfo
			want: LevelInfo,
			// 字符串表示：INFO
			s: "INFO",
		},
		{
			// 测试用例名称：WARN
			name: "WARN",
			// 预期级别：LevelWarn
			want: LevelWarn,
			// 字符串表示：WARN
			s: "WARN",
		},
		{
			// 测试用例名称：ERROR
			name: "ERROR",
			// 预期级别：LevelError
			want: LevelError,
			// 字符串表示：ERROR
			s: "ERROR",
		},
		{
			// 测试用例名称：FATAL
			name: "FATAL",
			// 预期级别：LevelFatal
			want: LevelFatal,
			// 字符串表示：FATAL
			s: "FATAL",
		},
		{
			// 测试用例名称：other
			name: "other",
			// 预期级别：LevelInfo
			want: LevelInfo,
			// 字符串表示：other
			s: "other",
		},
	}
	// 遍历测试用例
	for _, tt := range tests {
		// 运行每个测试用例
		t.Run(tt.name, func(t *testing.T) {
			// 检查解析后的级别是否与预期一致
			if got := ParseLevel(tt.s); got != tt.want {
				// 如果不一致，记录错误信息
				t.Errorf("ParseLevel() = %v, want %v", got, tt.want)
			}
		})
	}
}
