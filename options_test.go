package kratos

import (
	"context"
	"log"
	"net/url"
	"os"
	"reflect"
	"testing"
	"time"

	xlog "github.com/cnsync/kratos/log"
	"github.com/cnsync/kratos/registry"
	"github.com/cnsync/kratos/transport"
)

// TestID 测试 ID 设置方法
func TestID(t *testing.T) {
	// 创建一个 options 实例
	o := &options{}
	// 设置 ID
	v := "123"
	ID(v)(o)
	// 检查 ID 是否设置正确
	if !reflect.DeepEqual(v, o.id) {
		t.Fatalf("o.id:%s is not equal to v:%s", o.id, v)
	}
}

// TestName 测试名称设置方法
func TestName(t *testing.T) {
	// 创建一个 options 实例
	o := &options{}
	// 设置名称
	v := "abc"
	Name(v)(o)
	// 检查名称是否设置正确
	if !reflect.DeepEqual(v, o.name) {
		t.Fatalf("o.name:%s is not equal to v:%s", o.name, v)
	}
}

// TestVersion 测试版本设置方法
func TestVersion(t *testing.T) {
	// 创建一个 options 实例
	o := &options{}
	// 设置版本
	v := "123"
	Version(v)(o)
	// 检查版本是否设置正确
	if !reflect.DeepEqual(v, o.version) {
		t.Fatalf("o.version:%s is not equal to v:%s", o.version, v)
	}
}

// TestMetadata 测试元数据设置方法
func TestMetadata(t *testing.T) {
	// 创建一个 options 实例
	o := &options{}
	// 设置元数据
	v := map[string]string{
		"a": "1",
		"b": "2",
	}
	Metadata(v)(o)
	// 检查元数据是否设置正确
	if !reflect.DeepEqual(v, o.metadata) {
		t.Fatalf("o.metadata:%s is not equal to v:%s", o.metadata, v)
	}
}

// TestEndpoint 测试端点设置方法
func TestEndpoint(t *testing.T) {
	// 创建一个 options 实例
	o := &options{}
	// 设置端点
	v := []*url.URL{
		{Host: "example.com"},
		{Host: "foo.com"},
	}
	Endpoint(v...)(o)
	// 检查端点是否设置正确
	if !reflect.DeepEqual(v, o.endpoints) {
		t.Fatalf("o.endpoints:%s is not equal to v:%s", o.endpoints, v)
	}
}

// TestContext 测试上下文设置方法
func TestContext(t *testing.T) {
	type ctxKey struct {
		Key string
	}
	// 创建一个 options 实例
	o := &options{}
	// 设置上下文
	v := context.WithValue(context.TODO(), ctxKey{Key: "context"}, "b")
	Context(v)(o)
	// 检查上下文是否设置正确
	if !reflect.DeepEqual(v, o.ctx) {
		t.Fatalf("o.ctx:%s is not equal to v:%s", o.ctx, v)
	}
}

// TestLogger 测试日志记录器设置方法
func TestLogger(t *testing.T) {
	// 创建一个 options 实例
	o := &options{}
	// 设置日志记录器
	v := xlog.NewStdLogger(log.Writer())
	Logger(v)(o)
	// 检查日志记录器是否设置正确
	if !reflect.DeepEqual(v, o.logger) {
		t.Fatalf("o.logger:%v is not equal to xlog.NewHelper(v):%v", o.logger, xlog.NewHelper(v))
	}
}

type mockServer struct{}

func (m *mockServer) Start(_ context.Context) error { return nil }
func (m *mockServer) Stop(_ context.Context) error  { return nil }

// TestServer 测试服务器设置方法
func TestServer(t *testing.T) {
	// 创建一个 options 实例
	o := &options{}
	// 设置服务器
	v := []transport.Server{
		&mockServer{}, &mockServer{},
	}
	Server(v...)(o)
	// 检查服务器是否设置正确
	if !reflect.DeepEqual(v, o.servers) {
		t.Fatalf("o.servers:%s is not equal to xlog.NewHelper(v):%s", o.servers, v)
	}
}

type mockSignal struct{}

func (m *mockSignal) String() string { return "sig" }
func (m *mockSignal) Signal()        {}

// TestSignal 测试信号设置方法
func TestSignal(t *testing.T) {
	// 创建一个 options 实例
	o := &options{}
	// 设置信号
	v := []os.Signal{
		&mockSignal{}, &mockSignal{},
	}
	Signal(v...)(o)
	// 检查信号是否设置正确
	if !reflect.DeepEqual(v, o.sigs) {
		t.Fatal("o.sigs is not equal to v")
	}
}

type mockRegistrar struct{}

func (m *mockRegistrar) Register(_ context.Context, _ *registry.ServiceInstance) error {
	return nil
}

func (m *mockRegistrar) Deregister(_ context.Context, _ *registry.ServiceInstance) error {
	return nil
}

// TestRegistrar 测试注册器设置方法
func TestRegistrar(t *testing.T) {
	// 创建一个 options 实例
	o := &options{}
	// 设置注册器
	v := &mockRegistrar{}
	Registrar(v)(o)
	// 检查注册器是否设置正确
	if !reflect.DeepEqual(v, o.registrar) {
		t.Fatal("o.registrar is not equal to v")
	}
}

// TestRegistrarTimeout 测试注册器超时设置方法
func TestRegistrarTimeout(t *testing.T) {
	// 创建一个 options 实例
	o := &options{}
	// 设置注册器超时
	v := time.Duration(123)
	RegistrarTimeout(v)(o)
	// 检查注册器超时是否设置正确
	if !reflect.DeepEqual(v, o.registrarTimeout) {
		t.Fatal("o.registrarTimeout is not equal to v")
	}
}

// TestStopTimeout 测试停止超时设置方法
func TestStopTimeout(t *testing.T) {
	// 创建一个 options 实例
	o := &options{}
	// 设置停止超时
	v := time.Duration(123)
	StopTimeout(v)(o)
	// 检查停止超时是否设置正确
	if !reflect.DeepEqual(v, o.stopTimeout) {
		t.Fatal("o.stopTimeout is not equal to v")
	}
}

// TestBeforeStart 测试启动前回调设置方法
func TestBeforeStart(t *testing.T) {
	// 创建一个 options 实例
	o := &options{}
	// 设置启动前回调
	v := func(_ context.Context) error {
		t.Log("BeforeStart...")
		return nil
	}
	BeforeStart(v)(o)
}

// TestBeforeStop 测试停止前回调设置方法
func TestBeforeStop(t *testing.T) {
	// 创建一个 options 实例
	o := &options{}
	// 设置停止前回调
	v := func(_ context.Context) error {
		t.Log("BeforeStop...")
		return nil
	}
	BeforeStop(v)(o)
}

// TestAfterStart 测试启动后回调设置方法
func TestAfterStart(t *testing.T) {
	// 创建一个 options 实例
	o := &options{}
	// 设置启动后回调
	v := func(_ context.Context) error {
		t.Log("AfterStart...")
		return nil
	}
	AfterStart(v)(o)
}

// TestAfterStop 测试停止后回调设置方法
func TestAfterStop(t *testing.T) {
	// 创建一个 options 实例
	o := &options{}
	// 设置停止后回调
	v := func(_ context.Context) error {
		t.Log("AfterStop...")
		return nil
	}
	AfterStop(v)(o)
}
