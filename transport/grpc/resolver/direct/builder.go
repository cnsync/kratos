package direct

import (
	"strings"

	"google.golang.org/grpc/resolver"
)

const name = "direct"

func init() {
	// 注册 directBuilder 到 gRPC 解析器中
	resolver.Register(NewBuilder())
}

// directBuilder 结构体，用于创建直接解析器
type directBuilder struct{}

// NewBuilder 创建一个新的 directBuilder 实例
func NewBuilder() resolver.Builder {
	return &directBuilder{}
}

// Build 根据目标地址和客户端连接创建一个解析器实例
func (d *directBuilder) Build(target resolver.Target, cc resolver.ClientConn, _ resolver.BuildOptions) (resolver.Resolver, error) {
	// 解析目标地址中的路径部分，获取逗号分隔的地址列表
	addrs := make([]resolver.Address, 0)
	for _, addr := range strings.Split(strings.TrimPrefix(target.URL.Path, "/"), ",") {
		// 将每个地址添加到解析器的地址列表中
		addrs = append(addrs, resolver.Address{Addr: addr})
	}
	// 更新客户端连接的状态，设置解析器的地址列表
	err := cc.UpdateState(resolver.State{
		Addresses: addrs,
	})
	if err != nil {
		// 如果更新状态失败，返回错误
		return nil, err
	}
	// 创建并返回一个新的 directResolver 实例
	return newDirectResolver(), nil
}

// Scheme 返回解析器的方案名称
func (d *directBuilder) Scheme() string {
	return name
}
