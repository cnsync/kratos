package grpc

import (
	"fmt"

	"google.golang.org/grpc/encoding"
	"google.golang.org/protobuf/proto"

	enc "github.com/cnsync/kratos/encoding"
	"github.com/cnsync/kratos/encoding/json"
)

func init() {
	// 注册自定义的编解码器
	encoding.RegisterCodec(codec{})
}

// codec 是一个使用 protobuf 的 Codec 实现。它是 gRPC 的默认编解码器。
type codec struct{}

// Marshal 方法将 proto.Message 编码为字节切片。
func (codec) Marshal(v interface{}) ([]byte, error) {
	// 确保输入的对象是 proto.Message 类型
	vv, ok := v.(proto.Message)
	if !ok {
		// 如果不是，返回错误
		return nil, fmt.Errorf("failed to marshal, message is %T, want proto.Message", v)
	}
	// 使用 JSON 编解码器将 proto.Message 编码为字节切片
	return enc.GetCodec(json.Name).Marshal(vv)
}

// Unmarshal 方法将字节切片解码为 proto.Message。
func (codec) Unmarshal(data []byte, v interface{}) error {
	// 确保输入的对象是 proto.Message 类型
	vv, ok := v.(proto.Message)
	if !ok {
		// 如果不是，返回错误
		return fmt.Errorf("failed to unmarshal, message is %T, want proto.Message", v)
	}
	// 使用 JSON 编解码器将字节切片解码为 proto.Message
	return enc.GetCodec(json.Name).Unmarshal(data, vv)
}

// Name 方法返回编解码器的名称。
func (codec) Name() string {
	// 返回 JSON 编解码器的名称
	return json.Name
}
