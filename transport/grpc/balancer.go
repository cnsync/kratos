package grpc

import (
	"google.golang.org/grpc/balancer"
	"google.golang.org/grpc/balancer/base"
	"google.golang.org/grpc/metadata"

	"github.com/cnsync/kratos/registry"
	"github.com/cnsync/kratos/selector"
	"github.com/cnsync/kratos/transport"
)

const (
	// 负载均衡器的名称
	balancerName = "selector"
)

var (
	// 确保 balancerBuilder 实现了 base.PickerBuilder 接口
	_ base.PickerBuilder = (*balancerBuilder)(nil)
	// 确保 balancerPicker 实现了 balancer.Picker 接口
	_ balancer.Picker = (*balancerPicker)(nil)
)

// 初始化函数，注册负载均衡器
func init() {
	// 创建一个新的负载均衡器构建器
	b := base.NewBalancerBuilder(
		balancerName,
		&balancerBuilder{
			builder: selector.GlobalSelector(),
		},
		base.Config{HealthCheck: true},
	)
	// 注册负载均衡器
	balancer.Register(b)
}

// balancerBuilder 结构体，实现了 base.PickerBuilder 接口
type balancerBuilder struct {
	// 选择器构建器
	builder selector.Builder
}

// Build 方法，创建一个 gRPC Picker
func (b *balancerBuilder) Build(info base.PickerBuildInfo) balancer.Picker {
	if len(info.ReadySCs) == 0 {
		// 如果没有可用的子连接，则返回一个错误
		return base.NewErrPicker(balancer.ErrNoSubConnAvailable)
	}
	// 创建一个节点列表
	nodes := make([]selector.Node, 0, len(info.ReadySCs))
	// 遍历所有准备好的子连接
	for conn, info := range info.ReadySCs {
		// 从地址属性中获取原始服务实例
		ins, _ := info.Address.Attributes.Value("rawServiceInstance").(*registry.ServiceInstance)
		// 创建一个新的 grpcNode 并添加到节点列表中
		nodes = append(nodes, &grpcNode{
			Node:    selector.NewNode("grpc", info.Address.Addr, ins),
			subConn: conn,
		})
	}
	// 创建一个新的 balancerPicker
	p := &balancerPicker{
		selector: b.builder.Build(),
	}
	// 应用节点列表到选择器中
	p.selector.Apply(nodes)
	// 返回 picker
	return p
}

// balancerPicker 结构体，实现了 balancer.Picker 接口
type balancerPicker struct {
	// 选择器实例
	selector selector.Selector
}

// Pick 方法，选择实例
func (p *balancerPicker) Pick(info balancer.PickInfo) (balancer.PickResult, error) {
	var filters []selector.NodeFilter
	// 从客户端上下文中获取传输实例
	if tr, ok := transport.FromClientContext(info.Ctx); ok {
		if gtr, ok := tr.(*Transport); ok {
			// 获取节点过滤器
			filters = gtr.NodeFilters()
		}
	}

	// 使用选择器选择节点
	n, done, err := p.selector.Select(info.Ctx, selector.WithNodeFilter(filters...))
	if err != nil {
		return balancer.PickResult{}, err
	}

	// 返回选择结果
	return balancer.PickResult{
		SubConn: n.(*grpcNode).subConn,
		Done: func(di balancer.DoneInfo) {
			// 调用 done 函数，处理完成信息
			done(info.Ctx, selector.DoneInfo{
				Err:           di.Err,
				BytesSent:     di.BytesSent,
				BytesReceived: di.BytesReceived,
				ReplyMD:       Trailer(di.Trailer),
			})
		},
	}, nil
}

// Trailer 结构体，封装了 gRPC 响应的元数据
type Trailer metadata.MD

// Get 方法，获取指定键的值
func (t Trailer) Get(k string) string {
	v := metadata.MD(t).Get(k)
	if len(v) > 0 {
		return v[0]
	}
	return ""
}

// grpcNode 结构体，表示一个 gRPC 节点
type grpcNode struct {
	// 节点实例
	selector.Node
	// 子连接实例
	subConn balancer.SubConn
}
