package metadata

import (
	"context"
	"fmt"
	"strings"
)

// Metadata 是我们用来在内部表示请求头的方式。
// 它们用于 RPC 层，并可以在传输头和内部表示之间相互转换。
type Metadata map[string][]string

// New 从给定的键值映射创建一个 Metadata。
func New(mds ...map[string][]string) Metadata {
	md := Metadata{}
	for _, m := range mds {
		for k, vList := range m {
			for _, v := range vList {
				md.Add(k, v)
			}
		}
	}
	return md
}

// Add 将键值对添加到 Metadata 中。
func (m Metadata) Add(key, value string) {
	if len(key) == 0 {
		return
	}

	m[strings.ToLower(key)] = append(m[strings.ToLower(key)], value)
}

// Get 返回与指定键关联的值。
func (m Metadata) Get(key string) string {
	v := m[strings.ToLower(key)]
	if len(v) == 0 {
		return ""
	}
	return v[0]
}

// Set 存储一个键值对到 Metadata 中。
func (m Metadata) Set(key string, value string) {
	if key == "" || value == "" {
		return
	}
	m[strings.ToLower(key)] = []string{value}
}

// Range 遍历 Metadata 中的元素。
func (m Metadata) Range(f func(k string, v []string) bool) {
	for k, v := range m {
		if !f(k, v) {
			break
		}
	}
}

// Values 返回与指定键关联的所有值的切片。
func (m Metadata) Values(key string) []string {
	return m[strings.ToLower(key)]
}

// Clone 返回 Metadata 的一个深拷贝。
func (m Metadata) Clone() Metadata {
	md := make(Metadata, len(m))
	for k, v := range m {
		md[k] = v
	}
	return md
}

type serverMetadataKey struct{}

// NewServerContext 创建一个带有服务端 Metadata 的新上下文。
func NewServerContext(ctx context.Context, md Metadata) context.Context {
	return context.WithValue(ctx, serverMetadataKey{}, md)
}

// FromServerContext 返回上下文中的服务端 Metadata（如果存在）。
func FromServerContext(ctx context.Context) (Metadata, bool) {
	md, ok := ctx.Value(serverMetadataKey{}).(Metadata)
	return md, ok
}

type clientMetadataKey struct{}

// NewClientContext 创建一个带有客户端 Metadata 的新上下文。
func NewClientContext(ctx context.Context, md Metadata) context.Context {
	return context.WithValue(ctx, clientMetadataKey{}, md)
}

// FromClientContext 返回上下文中的客户端 Metadata（如果存在）。
func FromClientContext(ctx context.Context) (Metadata, bool) {
	md, ok := ctx.Value(clientMetadataKey{}).(Metadata)
	return md, ok
}

// AppendToClientContext 返回一个新的上下文，该上下文将提供的键值对与现有的客户端 Metadata 合并。
func AppendToClientContext(ctx context.Context, kv ...string) context.Context {
	if len(kv)%2 == 1 {
		panic(fmt.Sprintf("metadata: AppendToClientContext 收到了奇数个键值对：%d", len(kv)))
	}
	md, _ := FromClientContext(ctx)
	md = md.Clone()
	for i := 0; i < len(kv); i += 2 {
		md.Set(kv[i], kv[i+1])
	}
	return NewClientContext(ctx, md)
}

// MergeToClientContext 将新的 Metadata 合并到上下文中。
func MergeToClientContext(ctx context.Context, cmd Metadata) context.Context {
	md, _ := FromClientContext(ctx)
	md = md.Clone()
	for k, v := range cmd {
		md[k] = v
	}
	return NewClientContext(ctx, md)
}
