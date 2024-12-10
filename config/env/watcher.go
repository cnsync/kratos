package env

import (
	"context"
	"github.com/cnsync/kratos/config"
)

// 定义一个名为 watcher 的结构体，用于实现 config.Watcher 接口
var _ config.Watcher = (*watcher)(nil)

type watcher struct {
	ctx    context.Context
	cancel context.CancelFunc
}

// NewWatcher 函数创建一个新的 watcher 实例，用于监视环境变量的变化
func NewWatcher() (config.Watcher, error) {
	// 创建一个带有取消功能的上下文对象
	ctx, cancel := context.WithCancel(context.Background())
	// 返回 watcher 实例和 nil 错误
	return &watcher{ctx: ctx, cancel: cancel}, nil
}

// Next 方法会一直阻塞，直到 Stop 方法被调用
func (w *watcher) Next() ([]*config.KeyValue, error) {
	// 等待上下文被取消
	<-w.ctx.Done()
	// 返回 nil 和上下文的错误
	return nil, w.ctx.Err()
}

// Stop 方法用于停止监视，并取消上下文
func (w *watcher) Stop() error {
	// 调用 cancel 函数取消上下文
	w.cancel()
	// 返回 nil 表示没有错误
	return nil
}
