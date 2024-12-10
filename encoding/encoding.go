package encoding

import (
	"strings"
)

// Codec 定义了 Transport 用于编码和解码消息的接口。
// 注意：该接口的实现必须是线程安全的；Codec 的方法可以被并发的 goroutine 调用。
type Codec interface {
	// Marshal 返回 v 的线格式。
	Marshal(v interface{}) ([]byte, error)
	// Unmarshal 将线格式解析为 v。
	Unmarshal(data []byte, v interface{}) error
	// Name 返回 Codec 实现的名称。返回的字符串将在传输中用作内容类型的一部分。
	// 该结果必须是静态的；调用多次时结果不能发生变化。
	Name() string
}

var registeredCodecs = make(map[string]Codec)

// RegisterCodec 注册指定的 Codec，以供所有 Transport 客户端和服务端使用。
func RegisterCodec(codec Codec) {
	if codec == nil {
		panic("不能注册一个空的 Codec")
	}
	if codec.Name() == "" {
		panic("不能注册 Name() 结果为空字符串的 Codec")
	}
	contentSubtype := strings.ToLower(codec.Name())
	registeredCodecs[contentSubtype] = codec
}

// GetCodec 根据内容子类型获取已注册的 Codec，
// 如果该内容子类型没有注册对应的 Codec，则返回 nil。
//
// 内容子类型预期为小写格式。
func GetCodec(contentSubtype string) Codec {
	return registeredCodecs[contentSubtype]
}
