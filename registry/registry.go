package registry

import (
	"context"
	"fmt"
	"sort"
)

// Registrar 是服务注册器。
type Registrar interface {
	// Register 注册服务实例。
	Register(ctx context.Context, service *ServiceInstance) error
	// Deregister 注销服务实例。
	Deregister(ctx context.Context, service *ServiceInstance) error
}

// Discovery 是服务发现接口。
type Discovery interface {
	// GetService 根据服务名称返回内存中的服务实例。
	GetService(ctx context.Context, serviceName string) ([]*ServiceInstance, error)
	// Watch 根据服务名称创建一个监视器。
	Watch(ctx context.Context, serviceName string) (Watcher, error)
}

// Watcher 是服务的监视器。
type Watcher interface {
	// Next 在以下两种情况下返回服务实例列表：
	// 1. 第一次监视且服务实例列表非空。
	// 2. 检测到服务实例发生变更。
	// 如果以上两种条件均不满足，将阻塞直到 context 超时或被取消。
	Next() ([]*ServiceInstance, error)
	// Stop 关闭监视器。
	Stop() error
}

// ServiceInstance 是服务发现系统中的服务实例。
type ServiceInstance struct {
	// ID 是注册时唯一的实例 ID。
	ID string `json:"id"`
	// Name 是注册的服务名称。
	Name string `json:"name"`
	// Version 是编译时的版本号。
	Version string `json:"version"`
	// Metadata 是与服务实例关联的键值对元数据。
	Metadata map[string]string `json:"metadata"`
	// Endpoints 是服务实例的端点地址。
	// 格式：
	//   http://127.0.0.1:8000?isSecure=false
	//   grpc://127.0.0.1:9000?isSecure=false
	Endpoints []string `json:"endpoints"`
}

// String 方法将 ServiceInstance 实例转换为字符串表示
func (i *ServiceInstance) String() string {
	// 使用 fmt.Sprintf 函数格式化字符串，其中 %s 是格式化动词，表示将参数转换为字符串并插入到格式化字符串中的相应位置
	// i.Name 和 i.ID 是 ServiceInstance 实例的两个属性，分别表示服务的名称和唯一标识符
	return fmt.Sprintf("%s-%s", i.Name, i.ID)
}

// Equal 方法用于比较两个 ServiceInstance 实例是否相等
func (i *ServiceInstance) Equal(o interface{}) bool {
	// 如果两个实例都为 nil，则它们相等
	if i == nil && o == nil {
		return true
	}

	// 如果其中一个实例为 nil，而另一个不为 nil，则它们不相等
	if i == nil || o == nil {
		return false
	}

	// 将接口类型 o 转换为 *ServiceInstance 类型
	t, ok := o.(*ServiceInstance)
	// 如果转换失败，则它们不相等
	if !ok {
		return false
	}

	// 比较两个实例的 Endpoints 字段，如果长度不同，则它们不相等
	if len(i.Endpoints) != len(t.Endpoints) {
		return false
	}

	// 对两个实例的 Endpoints 字段进行排序
	sort.Strings(i.Endpoints)
	sort.Strings(t.Endpoints)
	// 比较两个实例的 Endpoints 字段，如果有任何一个元素不同，则它们不相等
	for j := 0; j < len(i.Endpoints); j++ {
		if i.Endpoints[j] != t.Endpoints[j] {
			return false
		}
	}

	// 比较两个实例的 Metadata 字段，如果长度不同，则它们不相等
	if len(i.Metadata) != len(t.Metadata) {
		return false
	}

	// 比较两个实例的 Metadata 字段，如果有任何一个键值对不同，则它们不相等
	for k, v := range i.Metadata {
		if v != t.Metadata[k] {
			return false
		}
	}

	// 比较两个实例的 ID、Name 和 Version 字段，如果有任何一个字段不同，则它们不相等
	return i.ID == t.ID && i.Name == t.Name && i.Version == t.Version
}
