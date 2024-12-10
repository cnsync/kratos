package log

import (
	"bytes"
	"context"
	"io"
	"strings"
	"sync"
	"testing"
	"time"
)

// 测试过滤器功能，过滤所有指定条件的日志
func TestFilterAll(_ *testing.T) {
	// 创建带默认时间戳和调用者信息的日志记录器
	logger := With(DefaultLogger, "ts", DefaultTimestamp, "caller", DefaultCaller)
	// 创建一个带多个过滤条件的日志助手
	log := NewHelper(NewFilter(logger,
		FilterLevel(LevelDebug),    // 过滤日志级别为 Debug 的日志
		FilterKey("username"),      // 过滤键为 "username" 的日志
		FilterValue("hello"),       // 过滤值为 "hello" 的日志
		FilterFunc(testFilterFunc), // 自定义过滤函数
	))
	// 记录不同级别的日志，测试过滤效果
	log.Log(LevelDebug, "msg", "test debug")
	log.Info("hello")
	log.Infow("password", "123456")
	log.Infow("username", "kratos")
	log.Warn("warn log")
}

// 测试日志级别过滤
func TestFilterLevel(_ *testing.T) {
	logger := With(DefaultLogger, "ts", DefaultTimestamp, "caller", DefaultCaller)
	log := NewHelper(NewFilter(NewFilter(logger, FilterLevel(LevelWarn)))) // 过滤掉低于警告级别的日志
	log.Log(LevelDebug, "msg1", "te1st debug")
	log.Debug("test debug")
	log.Debugf("test %s", "debug")
	log.Debugw("log", "test debug")
	log.Warn("warn log")
}

// 测试调用者信息过滤
func TestFilterCaller(_ *testing.T) {
	logger := With(DefaultLogger, "ts", DefaultTimestamp, "caller", DefaultCaller)
	log := NewFilter(logger)
	_ = log.Log(LevelDebug, "msg1", "te1st debug")
	logHelper := NewHelper(NewFilter(logger))
	logHelper.Log(LevelDebug, "msg1", "te1st debug")
}

// 测试根据键过滤
func TestFilterKey(_ *testing.T) {
	logger := With(DefaultLogger, "ts", DefaultTimestamp, "caller", DefaultCaller)
	log := NewHelper(NewFilter(logger, FilterKey("password"))) // 过滤键为 "password" 的日志
	log.Debugw("password", "123456")
}

// 测试根据值过滤
func TestFilterValue(_ *testing.T) {
	logger := With(DefaultLogger, "ts", DefaultTimestamp, "caller", DefaultCaller)
	log := NewHelper(NewFilter(logger, FilterValue("debug"))) // 过滤值为 "debug" 的日志
	log.Debugf("test %s", "debug")
}

// 测试自定义过滤函数
func TestFilterFunc(_ *testing.T) {
	logger := With(DefaultLogger, "ts", DefaultTimestamp, "caller", DefaultCaller)
	log := NewHelper(NewFilter(logger, FilterFunc(testFilterFunc))) // 使用自定义过滤函数
	log.Debug("debug level")
	log.Infow("password", "123456")
}

// 基准测试：按键过滤日志性能
func BenchmarkFilterKey(b *testing.B) {
	log := NewHelper(NewFilter(NewStdLogger(io.Discard), FilterKey("password")))
	for i := 0; i < b.N; i++ {
		log.Infow("password", "123456")
	}
}

// 基准测试：按值过滤日志性能
func BenchmarkFilterValue(b *testing.B) {
	log := NewHelper(NewFilter(NewStdLogger(io.Discard), FilterValue("password")))
	for i := 0; i < b.N; i++ {
		log.Infow("password")
	}
}

// 基准测试：自定义过滤函数性能
func BenchmarkFilterFunc(b *testing.B) {
	log := NewHelper(NewFilter(NewStdLogger(io.Discard), FilterFunc(testFilterFunc)))
	for i := 0; i < b.N; i++ {
		log.Info("password", "123456")
	}
}

// 自定义过滤函数示例
func testFilterFunc(level Level, keyvals ...interface{}) bool {
	if level == LevelWarn { // 只允许警告级别日志通过
		return true
	}
	for i := 0; i < len(keyvals); i++ {
		if keyvals[i] == "password" { // 模糊处理 password 值
			keyvals[i+1] = fuzzyStr
		}
	}
	return false
}

// 测试带日志前缀的过滤函数
func TestFilterFuncWitchLoggerPrefix(t *testing.T) {
	buf := new(bytes.Buffer)
	tests := []struct {
		logger Logger
		want   string
	}{
		{
			logger: NewFilter(With(NewStdLogger(buf), "caller", "caller", "prefix", "whatever"), FilterFunc(testFilterFuncWithLoggerPrefix)),
			want:   "",
		},
		{
			logger: NewFilter(With(NewStdLogger(buf), "caller", "caller"), FilterFunc(testFilterFuncWithLoggerPrefix)),
			want:   "INFO caller=caller msg=msg filtered=***\n",
		},
		{
			logger: NewFilter(With(NewStdLogger(buf)), FilterFunc(testFilterFuncWithLoggerPrefix)),
			want:   "INFO msg=msg filtered=***\n",
		},
	}

	for _, tt := range tests {
		err := tt.logger.Log(LevelInfo, "msg", "msg", "filtered", "true")
		if err != nil {
			t.Fatal("err should be nil")
		}
		got := buf.String()
		if got != tt.want {
			t.Fatalf("filter should catch prefix, want %s, got %s.", tt.want, got)
		}
		buf.Reset()
	}
}

// 带日志前缀的自定义过滤函数
func testFilterFuncWithLoggerPrefix(level Level, keyvals ...interface{}) bool {
	if level == LevelWarn {
		return true
	}
	for i := 0; i < len(keyvals); i += 2 {
		if keyvals[i] == "prefix" {
			return true
		}
		if keyvals[i] == "filtered" {
			keyvals[i+1] = fuzzyStr
		}
	}
	return false
}

// 测试上下文中日志过滤
func TestFilterWithContext(t *testing.T) {
	type CtxKey struct {
		Key string
	}
	ctxKey := CtxKey{Key: "context"}
	ctxValue := "filter test value"

	v1 := func() Valuer {
		return func(ctx context.Context) interface{} {
			return ctx.Value(ctxKey)
		}
	}

	info := &bytes.Buffer{}

	logger := With(NewStdLogger(info), "request_id", v1())
	filter := NewFilter(logger, FilterLevel(LevelError))

	ctx := context.WithValue(context.Background(), ctxKey, ctxValue)

	_ = WithContext(ctx, filter).Log(LevelInfo, "kind", "test")

	if info.String() != "" {
		t.Error("filter is not working")
		return
	}

	_ = WithContext(ctx, filter).Log(LevelError, "kind", "test")
	if !strings.Contains(info.String(), ctxValue) {
		t.Error("don't read ctx value")
	}
}

type traceIDKey struct{}

// 设置 trace ID
func setTraceID(ctx context.Context, tid string) context.Context {
	return context.WithValue(ctx, traceIDKey{}, tid)
}

// 获取 trace ID 的 Valuer
func traceIDValuer() Valuer {
	return func(ctx context.Context) any {
		if ctx == nil {
			return ""
		}
		if tid := ctx.Value(traceIDKey{}); tid != nil {
			return tid
		}
		return ""
	}
}

// 测试并发场景下的日志过滤
func TestFilterWithContextConcurrent(t *testing.T) {
	var buf bytes.Buffer
	pctx := context.Background()
	l := NewFilter(
		With(NewStdLogger(&buf), "trace-id", traceIDValuer()),
		FilterLevel(LevelInfo),
	)

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		time.Sleep(time.Second)
		NewHelper(l).Info("done1")
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		tid := "world"
		ctx := setTraceID(pctx, tid)
		NewHelper(WithContext(ctx, l)).Info("done2")
	}()

	wg.Wait()
	expected := "INFO trace-id=world msg=done2\nINFO trace-id= msg=done1\n"
	if got := buf.String(); got != expected {
		t.Errorf("got: %#v", got)
	}
}
