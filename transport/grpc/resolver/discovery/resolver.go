package discovery

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"google.golang.org/grpc/attributes"
	"google.golang.org/grpc/resolver"

	"github.com/cnsync/kratos/internal/endpoint"
	"github.com/cnsync/kratos/log"
	"github.com/cnsync/kratos/registry"
	"github.com/go-kratos/aegis/subset"
)

// discoveryResolver 结构体，用于实现 gRPC 的解析器接口
type discoveryResolver struct {
	w  registry.Watcher    // 服务发现的观察者
	cc resolver.ClientConn // 客户端连接

	ctx    context.Context    // 上下文
	cancel context.CancelFunc // 取消函数

	insecure    bool   // 是否为不安全连接
	debugLog    bool   // 是否打印调试日志
	selectorKey string // 选择器键
	subsetSize  int    // 子集大小
}

// watch 方法，用于监视服务实例的变化
func (r *discoveryResolver) watch() {
	for {
		select {
		case <-r.ctx.Done():
			return
		default:
		}
		ins, err := r.w.Next()
		if err != nil {
			if errors.Is(err, context.Canceled) {
				return
			}
			log.Errorf("[resolver] Failed to watch discovery endpoint: %v", err)
			time.Sleep(time.Second)
			continue
		}
		r.update(ins)
	}
}

// update 方法，用于更新客户端连接的状态
func (r *discoveryResolver) update(ins []*registry.ServiceInstance) {
	// 创建一个 map 用于存储已经处理过的 endpoints
	endpoints := make(map[string]struct{})
	// 创建一个切片用于存储过滤后的服务实例
	filtered := make([]*registry.ServiceInstance, 0, len(ins))

	// 遍历所有的服务实例
	for _, in := range ins {
		// 解析服务实例的 endpoints，获取 grpc 协议的 endpoint
		ept, err := endpoint.ParseEndpoint(in.Endpoints, endpoint.Scheme("grpc", !r.insecure))
		// 如果解析失败，记录错误并继续处理下一个实例
		if err != nil {
			log.Errorf("[resolver] Failed to parse discovery endpoint: %v", err)
			continue
		}
		// 如果解析后的 endpoint 为空，则忽略该实例
		if ept == "" {
			continue
		}
		// 过滤掉重复的 endpoints
		if _, ok := endpoints[ept]; ok {
			continue
		}
		// 将 endpoint 加入到 endpoints map 中
		endpoints[ept] = struct{}{}
		// 将服务实例加入到 filtered 切片中
		filtered = append(filtered, in)
	}

	// 如果设置了子集大小，则对服务实例进行子集划分
	if r.subsetSize != 0 {
		filtered = subset.Subset(r.selectorKey, filtered, r.subsetSize)
	}

	// 创建一个切片用于存储最终的 resolver.Address 列表
	addrs := make([]resolver.Address, 0, len(filtered))
	// 遍历过滤后的服务实例
	for _, in := range filtered {
		// 解析服务实例的 endpoints，获取 grpc 协议的 endpoint
		ept, _ := endpoint.ParseEndpoint(in.Endpoints, endpoint.Scheme("grpc", !r.insecure))
		// 创建一个 resolver.Address 对象
		addr := resolver.Address{
			ServerName: in.Name,
			Attributes: parseAttributes(in.Metadata).WithValue("rawServiceInstance", in),
			Addr:       ept,
		}
		// 将 resolver.Address 对象加入到 addrs 切片中
		addrs = append(addrs, addr)
	}

	// 如果没有找到任何 endpoint，则记录警告并返回
	if len(addrs) == 0 {
		log.Warnf("[resolver] Zero endpoint found,refused to write, instances: %v", ins)
		return
	}

	// 更新客户端连接的状态
	err := r.cc.UpdateState(resolver.State{Addresses: addrs})
	// 如果更新失败，记录错误
	if err != nil {
		log.Errorf("[resolver] failed to update state: %s", err)
	}

	// 如果开启了调试日志，则记录更新的实例信息
	if r.debugLog {
		b, _ := json.Marshal(filtered)
		log.Infof("[resolver] update instances: %s", b)
	}
}

// Close 方法，用于关闭解析器
func (r *discoveryResolver) Close() {
	r.cancel()
	err := r.w.Stop()
	if err != nil {
		log.Errorf("[resolver] failed to watch top: %s", err)
	}
}

// ResolveNow 方法，用于立即解析目标地址
func (r *discoveryResolver) ResolveNow(_ resolver.ResolveNowOptions) {}

// parseAttributes 函数用于解析服务实例的元数据，并将其转换为 attributes.Attributes 类型
func parseAttributes(md map[string]string) (a *attributes.Attributes) {
	// 遍历传入的元数据映射 md
	for k, v := range md {
		if a == nil {
			// 初始化 attributes.Attributes，使用第一个键值对
			a = attributes.New(k, v)
		} else {
			// 使用 WithValue 方法将后续的键值对添加到 attributes.Attributes
			a = a.WithValue(k, v)
		}
	}
	// 返回最终的 attributes.Attributes 类型实例
	return a
}
