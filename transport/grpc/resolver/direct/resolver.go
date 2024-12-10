package direct

import "google.golang.org/grpc/resolver"

// directResolver 结构体，用于实现 gRPC 的解析器接口
type directResolver struct{}

// newDirectResolver 函数，创建一个新的 directResolver 实例
func newDirectResolver() resolver.Resolver {
	return &directResolver{}
}

// Close 方法，关闭解析器
func (r *directResolver) Close() {
}

// ResolveNow 方法，立即解析目标地址
func (r *directResolver) ResolveNow(_ resolver.ResolveNowOptions) {
}
