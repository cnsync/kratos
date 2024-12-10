package discovery

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"
	"google.golang.org/grpc/resolver"

	"github.com/cnsync/kratos/registry"
)

const name = "discovery"

var ErrWatcherCreateTimeout = errors.New("discovery create watcher overtime")

// Option 是构建器的选项。
type Option func(o *builder)

// WithTimeout 带有超时选项。
func WithTimeout(timeout time.Duration) Option {
	return func(b *builder) {
		b.timeout = timeout
	}
}

// WithInsecure 带有是否安全的选项。
func WithInsecure(insecure bool) Option {
	return func(b *builder) {
		b.insecure = insecure
	}
}

// WithSubset 带有子集大小的选项。
func WithSubset(size int) Option {
	return func(b *builder) {
		b.subsetSize = size
	}
}

// PrintDebugLog 打印 gRPC 解析器观察服务日志
func PrintDebugLog(p bool) Option {
	return func(b *builder) {
		b.debugLog = p
	}
}

type builder struct {
	discoverer registry.Discovery
	timeout    time.Duration
	insecure   bool
	subsetSize int
	debugLog   bool
}

// NewBuilder 创建一个构建器，用于生成注册解析器。
func NewBuilder(d registry.Discovery, opts ...Option) resolver.Builder {
	b := &builder{
		discoverer: d,
		timeout:    time.Second * 10,
		insecure:   false,
		debugLog:   true,
		subsetSize: 25,
	}
	for _, o := range opts {
		o(b)
	}
	return b
}

// Build 根据目标地址和客户端连接创建一个解析器实例
func (b *builder) Build(target resolver.Target, cc resolver.ClientConn, _ resolver.BuildOptions) (resolver.Resolver, error) {
	// 定义一个结构体用于存储 Watch 操作的结果
	watchRes := &struct {
		err error
		w   registry.Watcher
	}{}

	// 创建一个通道用于通知 Watch 操作完成
	done := make(chan struct{}, 1)

	// 创建一个可取消的上下文
	ctx, cancel := context.WithCancel(context.Background())

	// 启动一个 goroutine 来执行 Watch 操作
	go func() {
		// 调用发现服务的 Watch 方法，监听目标路径的变化
		w, err := b.discoverer.Watch(ctx, strings.TrimPrefix(target.URL.Path, "/"))
		// 将 Watcher 和错误信息存储在 watchRes 中
		watchRes.w = w
		watchRes.err = err
		// 关闭 done 通道，表示 Watch 操作完成
		close(done)
	}()

	// 等待 Watch 操作完成或超时
	var err error
	select {
	case <-done:
		// 如果 done 通道关闭，获取 Watch 操作的错误信息
		err = watchRes.err
	case <-time.After(b.timeout):
		// 如果超时，设置错误为创建观察者超时
		err = ErrWatcherCreateTimeout
	}

	// 如果发生错误，取消上下文并返回错误
	if err != nil {
		cancel()
		return nil, err
	}

	// 创建一个新的 discoveryResolver 实例
	r := &discoveryResolver{
		w:           watchRes.w,
		cc:          cc,
		ctx:         ctx,
		cancel:      cancel,
		insecure:    b.insecure,
		debugLog:    b.debugLog,
		subsetSize:  b.subsetSize,
		selectorKey: uuid.New().String(), // 生成一个唯一的选择器键
	}

	// 启动一个 goroutine 来监视服务实例的变化
	go r.watch()

	// 返回解析器实例和 nil 错误
	return r, nil
}

// Scheme 返回发现的方案
func (*builder) Scheme() string {
	return name
}
