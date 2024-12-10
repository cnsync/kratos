package context

import (
	"context"
	"sync"
	"sync/atomic"
	"time"
)

// mergeCtx 结构体用于合并两个 context.Context 对象
type mergeCtx struct {
	// 父 context1
	parent1 context.Context
	// 父 context2
	parent2 context.Context

	// 用于通知合并后的 context 已完成的通道
	done chan struct{}
	// 标记 done 通道是否已关闭的原子变量
	doneMark uint32
	// 确保 finish 函数只被调用一次的 Once 对象
	doneOnce sync.Once
	// 合并后的 context 的错误
	doneErr error

	// 用于取消合并后的 context 的通道
	cancelCh chan struct{}
	// 确保 cancel 函数只被调用一次的 Once 对象
	cancelOnce sync.Once
}

// Merge 函数用于合并两个 context.Context 对象
func Merge(parent1, parent2 context.Context) (context.Context, context.CancelFunc) {
	// 创建一个 mergeCtx 实例
	mc := &mergeCtx{
		parent1:  parent1,
		parent2:  parent2,
		done:     make(chan struct{}),
		cancelCh: make(chan struct{}),
	}
	// 启动一个 goroutine 等待两个父 context 完成
	go mc.wait()
	// 检查两个父 context 是否已经完成
	select {
	case <-parent1.Done():
		// 如果 parent1 已经完成，调用 finish 函数处理结果
		_ = mc.finish(parent1.Err())
	case <-parent2.Done():
		// 如果 parent2 已经完成，调用 finish 函数处理结果
		_ = mc.finish(parent2.Err())
	default:
		// 如果两个父 context 都没有完成，返回合并后的 context 和取消函数
		return mc, mc.cancel
	}
	// 返回合并后的 context 和取消函数
	return mc, mc.cancel
}

// finish 函数用于标记合并后的 context 已完成，并设置错误信息
func (mc *mergeCtx) finish(err error) error {
	// 使用 doneOnce 确保 finish 函数只被调用一次
	mc.doneOnce.Do(func() {
		// 设置 doneErr 为传入的错误信息
		mc.doneErr = err
		// 将 doneMark 标记为 1，表示 done 通道已关闭
		atomic.StoreUint32(&mc.doneMark, 1)
		// 关闭 done 通道，通知所有等待的 goroutine
		close(mc.done)
	})
	// 返回合并后的 context 的错误信息
	return mc.doneErr
}

// wait 函数用于等待两个父 context 完成，并调用 finish 函数处理结果
func (mc *mergeCtx) wait() {
	// 声明一个错误变量
	var err error
	// 使用 select 等待两个父 context 完成或取消信号
	select {
	case <-mc.parent1.Done():
		// 如果 parent1 完成，获取其错误信息
		err = mc.parent1.Err()
	case <-mc.parent2.Done():
		// 如果 parent2 完成，获取其错误信息
		err = mc.parent2.Err()
	case <-mc.cancelCh:
		// 如果合并后的 context 被取消，设置错误为 context.Canceled
		err = context.Canceled
	}
	// 调用 finish 函数处理结果
	_ = mc.finish(err)
}

// cancel 函数用于取消合并后的 context
func (mc *mergeCtx) cancel() {
	// 使用 cancelOnce 确保 cancel 函数只被调用一次
	mc.cancelOnce.Do(func() {
		// 关闭 cancelCh 通道，通知所有等待的 goroutine
		close(mc.cancelCh)
	})
}

// Done 方法实现 context.Context 接口，返回一个通道，当合并后的 context 被取消或超时时会关闭
func (mc *mergeCtx) Done() <-chan struct{} {
	return mc.done
}

// Err 方法实现 context.Context 接口，返回合并后的 context 的错误信息
func (mc *mergeCtx) Err() error {
	// 检查 doneMark 是否为 1，表示 done 通道是否已关闭
	if atomic.LoadUint32(&mc.doneMark) != 0 {
		// 如果 done 通道已关闭，返回 doneErr 错误信息
		return mc.doneErr
	}
	// 声明一个错误变量
	var err error
	// 使用 select 等待两个父 context 完成或取消信号
	select {
	case <-mc.parent1.Done():
		// 如果 parent1 完成，获取其错误信息
		err = mc.parent1.Err()
	case <-mc.parent2.Done():
		// 如果 parent2 完成，获取其错误信息
		err = mc.parent2.Err()
	case <-mc.cancelCh:
		// 如果合并后的 context 被取消，设置错误为 context.Canceled
		err = context.Canceled
	default:
		// 如果没有错误，返回 nil
		return nil
	}
	// 调用 finish 函数处理结果
	return mc.finish(err)
}

// Deadline 方法实现 context.Context 接口，返回合并后的 context 的截止时间
func (mc *mergeCtx) Deadline() (time.Time, bool) {
	// 获取 parent1 的截止时间和是否有截止时间的布尔值
	d1, ok1 := mc.parent1.Deadline()
	// 获取 parent2 的截止时间和是否有截止时间的布尔值
	d2, ok2 := mc.parent2.Deadline()
	// 根据两个父 context 的截止时间决定合并后的 context 的截止时间
	switch {
	case !ok1:
		// 如果 parent1 没有截止时间，返回 parent2 的截止时间
		return d2, ok2
	case !ok2:
		// 如果 parent2 没有截止时间，返回 parent1 的截止时间
		return d1, ok1
	case d1.Before(d2):
		// 如果 parent1 的截止时间早于 parent2 的截止时间，返回 parent1 的截止时间
		return d1, true
	default:
		// 否则，返回 parent2 的截止时间
		return d2, true
	}
}

// Value 方法实现 context.Context 接口，返回合并后的 context 中指定键的值
func (mc *mergeCtx) Value(key interface{}) interface{} {
	// 首先从 parent1 中获取值
	if v := mc.parent1.Value(key); v != nil {
		// 如果 parent1 中有值，直接返回
		return v
	}
	// 如果 parent1 中没有值，从 parent2 中获取值
	return mc.parent2.Value(key)
}
