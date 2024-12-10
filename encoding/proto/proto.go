// Package proto 定义了 protobuf 编解码器。导入该包时会自动注册该编解码器。
package proto

import (
	"errors"
	"github.com/cnsync/kratos/encoding"
	"reflect"

	"google.golang.org/protobuf/proto"
)

// Name 是为 proto 编解码器注册的名称。
const Name = "proto"

func init() {
	// 注册一个名为 proto 的编解码器
	encoding.RegisterCodec(codec{})
}

// codec 是基于 protobuf 的 Codec 实现。它是 Transport 的默认编解码器。
type codec struct{}

// Marshal 方法将一个 Go 语言的值序列化为 Protocol Buffers 格式的字节切片
func (codec) Marshal(v interface{}) ([]byte, error) {
	// 使用 protobuf 包中的 Marshal 函数将值 v 序列化为 Protocol Buffers 格式
	return proto.Marshal(v.(proto.Message))
}

// Unmarshal 方法将一个 Protocol Buffers 格式的字节切片反序列化为 Go 语言中的值
func (codec) Unmarshal(data []byte, v interface{}) error {
	// 获取 protobuf 消息对象
	pm, err := getProtoMessage(v)
	if err != nil {
		return err
	}
	// 使用 protobuf 包中的 Unmarshal 函数将字节切片 data 反序列化为值 v
	return proto.Unmarshal(data, pm)
}

// Name 方法返回编解码器的名称
func (codec) Name() string {
	// 返回编解码器的名称 "proto"
	return Name
}

// getProtoMessage 方法从接口类型中获取 protobuf 消息对象
// 如果不是，则返回错误。
func getProtoMessage(v interface{}) (proto.Message, error) {
	// 检查 v 是否已经是 protobuf 消息对象
	if msg, ok := v.(proto.Message); ok {
		return msg, nil
	}
	// 获取 v 的反射值
	val := reflect.ValueOf(v)
	// 如果 v 不是指针类型，则返回错误
	if val.Kind() != reflect.Ptr {
		return nil, errors.New("not proto message")
	}
	// 获取指针指向的值
	val = val.Elem()
	// 递归调用 getProtoMessage 方法，获取 protobuf 消息对象
	return getProtoMessage(val.Interface())
}
