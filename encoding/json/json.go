package json

import (
	"github.com/cnsync/kratos/encoding"

	"encoding/json"
	"reflect"

	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
)

// Name 是为 json 编解码器注册的名称。
const Name = "json"

var (
	// MarshalOptions 是一个可配置的 JSON 格式 marshaller。
	MarshalOptions = protojson.MarshalOptions{
		EmitUnpopulated: true,
	}
	// UnmarshalOptions 是一个可配置的 JSON 格式解析器。
	UnmarshalOptions = protojson.UnmarshalOptions{
		DiscardUnknown: true,
	}
)

func init() {
	// 注册一个名为 json 的编解码器
	encoding.RegisterCodec(codec{})
}

// codec 是一个使用 json 实现的编解码器。
type codec struct{}

// Marshal 方法将一个 Go 语言的值序列化为 JSON 格式的字节切片。
func (codec) Marshal(v interface{}) ([]byte, error) {
	switch m := v.(type) {
	case json.Marshaler:
		// 如果 v 实现了 json.Marshaler 接口，则调用其 MarshalJSON 方法。
		return m.MarshalJSON()
	case proto.Message:
		// 如果 v 是 proto.Message 类型，则使用 MarshalOptions 进行序列化。
		return MarshalOptions.Marshal(m)
	default:
		// 否则，使用默认的 json.Marshal 方法进行序列化。
		return json.Marshal(m)
	}
}

// Unmarshal 方法将一个 JSON 格式的字节切片反序列化为 Go 语言中的值。
func (codec) Unmarshal(data []byte, v interface{}) error {
	switch m := v.(type) {
	case json.Unmarshaler:
		// 如果 v 实现了 json.Unmarshaler 接口，则调用其 UnmarshalJSON 方法。
		return m.UnmarshalJSON(data)
	case proto.Message:
		// 如果 v 是 proto.Message 类型，则使用 UnmarshalOptions 进行反序列化。
		return UnmarshalOptions.Unmarshal(data, m)
	default:
		rv := reflect.ValueOf(v)
		for rv := rv; rv.Kind() == reflect.Ptr; {
			// 如果 v 是指针类型，且为空，则创建一个新的实例。
			if rv.IsNil() {
				rv.Set(reflect.New(rv.Type().Elem()))
			}
			rv = rv.Elem()
		}
		if m, ok := reflect.Indirect(rv).Interface().(proto.Message); ok {
			// 如果 v 是 proto.Message 类型，则使用 UnmarshalOptions 进行反序列化。
			return UnmarshalOptions.Unmarshal(data, m)
		}
		// 否则，使用默认的 json.Unmarshal 方法进行反序列化。
		return json.Unmarshal(data, m)
	}
}

// Name 方法返回编解码器的名称。
func (codec) Name() string {
	// 返回编解码器的名称 "json"
	return Name
}
