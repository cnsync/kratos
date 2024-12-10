package config

import (
	"bytes"
	"encoding/gob"
	"encoding/json"
	"fmt"
	"strings"
	"sync"

	"github.com/cnsync/kratos/log"

	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
)

// Reader 是配置读取器接口。
type Reader interface {
	Merge(...*KeyValue) error   // 合并配置数据
	Value(string) (Value, bool) // 获取指定路径的配置值
	Source() ([]byte, error)    // 获取所有配置的 JSON 表示
	Resolve() error             // 执行配置解析器
}

// 内部的配置读取器实现
type reader struct {
	opts   options                // 配置选项
	values map[string]interface{} // 配置键值存储
	lock   sync.Mutex             // 用于保护并发访问的锁
}

// 创建一个新的 Reader 实例
func newReader(opts options) Reader {
	return &reader{
		opts:   opts,
		values: make(map[string]interface{}),
		lock:   sync.Mutex{},
	}
}

// Merge 将多个 KeyValue 合并到当前配置中
func (r *reader) Merge(kvs ...*KeyValue) error {
	merged, err := r.cloneMap() // 克隆当前配置
	if err != nil {
		return err
	}
	for _, kv := range kvs {
		next := make(map[string]interface{})
		// 使用解码器解码 KeyValue
		if err := r.opts.decoder(kv, next); err != nil {
			log.Errorf("配置解码失败: %v 键: %s 值: %s", err, kv.Key, string(kv.Value))
			return err
		}
		// 使用合并器合并配置
		if err := r.opts.merge(&merged, convertMap(next)); err != nil {
			log.Errorf("配置合并失败: %v 键: %s 值: %s", err, kv.Key, string(kv.Value))
			return err
		}
	}
	// 更新配置存储
	r.lock.Lock()
	r.values = merged
	r.lock.Unlock()
	return nil
}

// Value 获取指定路径的配置值
func (r *reader) Value(path string) (Value, bool) {
	r.lock.Lock()
	defer r.lock.Unlock()
	return readValue(r.values, path)
}

// Source 获取当前配置的 JSON 表示
func (r *reader) Source() ([]byte, error) {
	r.lock.Lock()
	defer r.lock.Unlock()
	return marshalJSON(convertMap(r.values))
}

// Resolve 调用解析器处理配置
func (r *reader) Resolve() error {
	r.lock.Lock()
	defer r.lock.Unlock()
	return r.opts.resolver(r.values)
}

// 克隆当前配置
func (r *reader) cloneMap() (map[string]interface{}, error) {
	r.lock.Lock()
	defer r.lock.Unlock()
	return cloneMap(r.values)
}

// cloneMap 克隆一个 map[string]interface{} 的深拷贝
func cloneMap(src map[string]interface{}) (map[string]interface{}, error) {
	// 使用 gob 编码和解码进行深拷贝
	var buf bytes.Buffer
	// 注册 map[string]interface{} 类型，确保可以正确编码和解码
	gob.Register(map[string]interface{}{})
	// 注册 []interface{} 类型，确保可以正确编码和解码
	gob.Register([]interface{}{})
	// 创建一个新的 gob 编码器，将数据编码到 buf 中
	enc := gob.NewEncoder(&buf)
	// 创建一个新的 gob 解码器，从 buf 中解码数据
	dec := gob.NewDecoder(&buf)
	// 将 src 编码到 buf 中
	err := enc.Encode(src)
	// 如果编码过程中发生错误，返回 nil 和错误信息
	if err != nil {
		return nil, err
	}
	// 创建一个新的 map[string]interface{} 用于存储克隆的数据
	var clone map[string]interface{}
	// 从 buf 中解码数据到 clone 中
	err = dec.Decode(&clone)
	// 如果解码过程中发生错误，返回 nil 和错误信息
	if err != nil {
		return nil, err
	}
	// 返回克隆后的 map 和 nil 表示没有错误
	return clone, nil
}

// convertMap 将 map 的键从任意类型转换为 string 类型
func convertMap(src interface{}) interface{} {
	switch m := src.(type) {
	case map[string]interface{}:
		// 如果输入已经是 map[string]interface{} 类型，则直接返回
		dst := make(map[string]interface{}, len(m))
		for k, v := range m {
			dst[k] = convertMap(v)
		}
		return dst
	case map[interface{}]interface{}:
		// 如果输入是 map[interface{}]interface{} 类型，则将键转换为字符串
		dst := make(map[string]interface{}, len(m))
		for k, v := range m {
			dst[fmt.Sprint(k)] = convertMap(v)
		}
		return dst
	case []interface{}:
		// 如果输入是 []interface{} 类型，则递归处理每个元素
		dst := make([]interface{}, len(m))
		for k, v := range m {
			dst[k] = convertMap(v)
		}
		return dst
	case []byte:
		// 配置数据中不应包含二进制数据，将其转换为字符串
		return string(m)
	default:
		// 如果输入不是上述类型，则直接返回
		return src
	}
}

// readValue 函数根据给定的路径从 map 中读取值
func readValue(values map[string]interface{}, path string) (Value, bool) {
	// 初始化变量，next 指向 values，keys 是路径分割后的字符串数组，last 是 keys 的最后一个索引
	var (
		next = values
		keys = strings.Split(path, ".")
		last = len(keys) - 1
	)
	// 遍历 keys 数组
	for idx, key := range keys {
		// 尝试从 next 中获取 key 对应的值
		value, ok := next[key]
		// 如果没有找到对应的值，返回 nil 和 false
		if !ok {
			return nil, false
		}
		// 如果是最后一个 key，创建一个 atomicValue 并存储 value，然后返回
		if idx == last {
			av := &atomicValue{}
			av.Store(value)
			return av, true
		}
		// 根据 value 的类型进行不同的处理
		switch vm := value.(type) {
		// 如果 value 是 map[string]interface{} 类型，将 next 指向这个 map
		case map[string]interface{}:
			next = vm
		// 如果 value 是其他类型，返回 nil 和 false
		default:
			return nil, false
		}
	}
	// 如果没有找到对应的值，返回 nil 和 false
	return nil, false
}

// marshalJSON 将值编码为 JSON
func marshalJSON(v interface{}) ([]byte, error) {
	if m, ok := v.(proto.Message); ok {
		// 使用 protobuf JSON 序列化选项
		return protojson.MarshalOptions{EmitUnpopulated: true}.Marshal(m)
	}
	return json.Marshal(v)
}

// unmarshalJSON 将 JSON 解码为值
func unmarshalJSON(data []byte, v interface{}) error {
	if m, ok := v.(proto.Message); ok {
		// 使用 protobuf JSON 反序列化选项
		return protojson.UnmarshalOptions{DiscardUnknown: true}.Unmarshal(data, m)
	}
	return json.Unmarshal(data, v)
}
