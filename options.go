package kratos

import (
	"context"
	"net/url"
	"os"
	"time"

	"github.com/cnsync/kratos/log"
	"github.com/cnsync/kratos/registry"
	"github.com/cnsync/kratos/transport"
)

// Option 是应用程序的选项。
type Option func(o *options)

// options 是应用程序的选项。
type options struct {
	id        string
	name      string
	version   string
	metadata  map[string]string
	endpoints []*url.URL

	ctx  context.Context
	sigs []os.Signal

	logger           log.Logger
	registrar        registry.Registrar
	registrarTimeout time.Duration
	stopTimeout      time.Duration
	servers          []transport.Server

	// 启动前和停止后的函数
	beforeStart []func(context.Context) error
	beforeStop  []func(context.Context) error
	afterStart  []func(context.Context) error
	afterStop   []func(context.Context) error
}

// ID 用于设置服务 ID。
func ID(id string) Option {
	return func(o *options) { o.id = id }
}

// Name 用于设置服务名称。
func Name(name string) Option {
	return func(o *options) { o.name = name }
}

// Version 用于设置服务版本。
func Version(version string) Option {
	return func(o *options) { o.version = version }
}

// Metadata 用于设置服务元数据。
func Metadata(md map[string]string) Option {
	return func(o *options) { o.metadata = md }
}

// Endpoint 用于设置服务端点。
func Endpoint(endpoints ...*url.URL) Option {
	return func(o *options) { o.endpoints = endpoints }
}

// Context 用于设置服务上下文。
func Context(ctx context.Context) Option {
	return func(o *options) { o.ctx = ctx }
}

// Logger 用于设置服务日志记录器。
func Logger(logger log.Logger) Option {
	return func(o *options) { o.logger = logger }
}

// Server 用于设置传输服务器。
func Server(srv ...transport.Server) Option {
	return func(o *options) { o.servers = srv }
}

// Signal 用于设置退出信号。
func Signal(sigs ...os.Signal) Option {
	return func(o *options) { o.sigs = sigs }
}

// Registrar 用于设置服务注册器。
func Registrar(r registry.Registrar) Option {
	return func(o *options) { o.registrar = r }
}

// RegistrarTimeout 用于设置注册器超时时间。
func RegistrarTimeout(t time.Duration) Option {
	return func(o *options) { o.registrarTimeout = t }
}

// StopTimeout 用于设置应用程序停止超时时间。
func StopTimeout(t time.Duration) Option {
	return func(o *options) { o.stopTimeout = t }
}

// 启动前和停止后的函数

// BeforeStart 在应用程序启动前运行函数。
func BeforeStart(fn func(context.Context) error) Option {
	return func(o *options) {
		o.beforeStart = append(o.beforeStart, fn)
	}
}

// BeforeStop 在应用程序停止前运行函数。
func BeforeStop(fn func(context.Context) error) Option {
	return func(o *options) {
		o.beforeStop = append(o.beforeStop, fn)
	}
}

// AfterStart 在应用程序启动后运行函数。
func AfterStart(fn func(context.Context) error) Option {
	return func(o *options) {
		o.afterStart = append(o.afterStart, fn)
	}
}

// AfterStop 在应用程序停止后运行函数。
func AfterStop(fn func(context.Context) error) Option {
	return func(o *options) {
		o.afterStop = append(o.afterStop, fn)
	}
}
