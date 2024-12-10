package ewma

import (
	"context"
	"math"
	"net"
	"sync/atomic"
	"time"

	"github.com/cnsync/kratos/errors"
	"github.com/cnsync/kratos/selector"
)

// 定义常量
const (
	// `cost` 的平均生命周期，在 Tau*ln(2) 后达到其半衰期。
	tau = int64(time.Millisecond * 600)
	// 如果没有收集到统计信息，则为端点添加一个较大的延迟惩罚值
	penalty = uint64(time.Microsecond * 100)
)

var (
	_ selector.WeightedNode        = (*Node)(nil)
	_ selector.WeightedNodeBuilder = (*Builder)(nil)
)

// Node 是服务实例的表示
type Node struct {
	selector.Node

	// 客户端统计的数据
	lag       int64      // 平均延迟时间
	success   uint64     // 成功率
	inflight  int64      // 当前正在处理的请求数
	inflights [200]int64 // 记录请求开始的时间戳，用于计算延迟
	// 最后一次收集统计的时间戳
	stamp int64
	// 在一段时间内的请求数
	reqs int64
	// 上次选择该节点的时间戳
	lastPick int64

	errHandler   func(err error) (isErr bool) // 错误处理函数
	cachedWeight *atomic.Value                // 用于缓存权重的原子变量
}

type nodeWeight struct {
	value    float64 // 权重值
	updateAt int64   // 更新时间戳
}

// Builder 是用于构建加权节点的构建器
type Builder struct {
	ErrHandler func(err error) (isErr bool) // 自定义错误处理函数
}

// Build 方法根据给定的节点创建一个新的加权节点实例
func (b *Builder) Build(n selector.Node) selector.WeightedNode {
	// 创建一个新的 Node 实例 s
	s := &Node{
		// 设置 Node 实例的基本属性
		Node: n,
		// 初始化平均延迟为 0
		lag: 0,
		// 初始化成功率为 1000（代表100%）
		success: 1000,
		// 初始化并发请求数为 1
		inflight: 1,
		// 设置错误处理函数为构建器的错误处理函数
		errHandler: b.ErrHandler,
		// 创建一个新的 atomic.Value 实例用于缓存权重
		cachedWeight: &atomic.Value{},
	}
	// 返回新创建的加权节点实例
	return s
}

// health 获取节点的成功率
func (n *Node) health() uint64 {
	return atomic.LoadUint64(&n.success)
}

// load 计算节点的负载
func (n *Node) load() (load uint64) {
	now := time.Now().UnixNano()
	avgLag := atomic.LoadInt64(&n.lag)
	predict := n.predict(avgLag, now)

	if avgLag == 0 {
		// 如果节点刚开始运行且没有数据，使用惩罚值作为负载
		load = penalty * uint64(atomic.LoadInt64(&n.inflight))
		return
	}
	if predict > avgLag {
		avgLag = predict
	}
	// 加5ms以消除不同区域之间的延迟差异
	avgLag += int64(time.Millisecond * 5)
	avgLag = int64(math.Sqrt(float64(avgLag)))
	load = uint64(avgLag) * uint64(atomic.LoadInt64(&n.inflight))
	return load
}

// predict 方法根据当前 inflight 请求的延迟情况，预测下一个请求的延迟
func (n *Node) predict(avgLag int64, now int64) (predict int64) {
	var (
		total    int64 // 总延迟
		slowNum  int   // 慢请求数量
		totalNum int   // 总请求数量
	)
	// 遍历 inflights 数组，统计当前 inflight 请求的延迟情况
	for i := range n.inflights {
		start := atomic.LoadInt64(&n.inflights[i])
		// 如果请求开始时间不为 0，则计算延迟
		if start != 0 {
			totalNum++
			lag := now - start
			// 如果延迟大于平均延迟，则认为是慢请求
			if lag > avgLag {
				slowNum++
				total += lag
			}
		}
	}
	// 如果慢请求数量超过总请求数量的一半，则预测延迟为总延迟除以慢请求数量
	if slowNum >= (totalNum/2 + 1) {
		predict = total / int64(slowNum)
	}
	return
}

// Pick 选择一个节点并返回完成时调用的回调函数
func (n *Node) Pick() selector.DoneFunc {
	// 记录当前时间，作为请求开始时间
	start := time.Now().UnixNano()
	// 更新节点的 lastPick 时间为当前时间
	atomic.StoreInt64(&n.lastPick, start)
	// 增加节点的 inflight 请求数量
	atomic.AddInt64(&n.inflight, 1)
	// 增加节点的总请求数量
	reqs := atomic.AddInt64(&n.reqs, 1)
	// 计算请求在 inflights 数组中的索引
	slot := reqs % 200
	// 尝试将 inflights 数组中的对应位置设置为当前时间
	swapped := atomic.CompareAndSwapInt64(&n.inflights[slot], 0, start)
	// 返回一个回调函数，该回调函数在请求完成时被调用
	return func(_ context.Context, di selector.DoneInfo) {
		// 如果成功设置了 inflights 数组中的值，则在请求完成时将其重置为 0
		if swapped {
			atomic.CompareAndSwapInt64(&n.inflights[slot], start, 0)
		}
		// 减少节点的 inflight 请求数量
		atomic.AddInt64(&n.inflight, -1)

		// 获取当前时间
		now := time.Now().UnixNano()
		// 获取移动平均比率 w
		stamp := atomic.SwapInt64(&n.stamp, now)
		td := now - stamp
		if td < 0 {
			td = 0
		}
		w := math.Exp(float64(-td) / float64(tau))

		lag := now - start
		if lag < 0 {
			lag = 0
		}
		oldLag := atomic.LoadInt64(&n.lag)
		if oldLag == 0 {
			w = 0.0
		}
		lag = int64(float64(oldLag)*w + float64(lag)*(1.0-w))
		atomic.StoreInt64(&n.lag, lag)

		success := uint64(1000) // 默认成功率是100%
		if di.Err != nil {
			if n.errHandler != nil && n.errHandler(di.Err) {
				success = 0
			}
			var netErr net.Error
			if errors.Is(context.DeadlineExceeded, di.Err) || errors.Is(context.Canceled, di.Err) ||
				errors.IsServiceUnavailable(di.Err) || errors.IsGatewayTimeout(di.Err) || errors.As(di.Err, &netErr) {
				success = 0
			}
		}
		oldSuc := atomic.LoadUint64(&n.success)
		success = uint64(float64(oldSuc)*w + float64(success)*(1.0-w))
		atomic.StoreUint64(&n.success, success)
	}
}

// Weight 获取节点的有效权重
func (n *Node) Weight() (weight float64) {
	// 尝试从 cachedWeight 中加载节点权重
	w, ok := n.cachedWeight.Load().(*nodeWeight)
	// 获取当前时间的纳秒表示
	now := time.Now().UnixNano()
	// 如果权重未找到或权重更新时间超过 5 毫秒
	if !ok || time.Duration(now-w.updateAt) > (time.Millisecond*5) {
		// 获取节点的健康度
		health := n.health()
		// 获取节点的负载
		load := n.load()
		// 计算权重，健康度乘以 10 微秒，再除以负载
		weight = float64(health*uint64(time.Microsecond)*10) / float64(load)
		// 将新计算的权重存储到 cachedWeight 中
		n.cachedWeight.Store(&nodeWeight{
			value:    weight,
			updateAt: now,
		})
	} else {
		// 如果权重是有效的，则直接使用缓存中的权重
		weight = w.value
	}
	// 返回计算得到的权重
	return
}

// PickElapsed 获取自上次选取节点以来经过的时间
func (n *Node) PickElapsed() time.Duration {
	return time.Duration(time.Now().UnixNano() - atomic.LoadInt64(&n.lastPick))
}

// Raw 返回原始的 Node 实例
func (n *Node) Raw() selector.Node {
	return n.Node
}
