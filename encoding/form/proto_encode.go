package form

import (
	"encoding/base64"
	"fmt"
	"net/url"
	"reflect"
	"strconv"
	"strings"

	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/types/known/fieldmaskpb"
)

// EncodeValues 函数用于将一个 protobuf 消息编码为 URL 查询字符串格式。
// 参数：
//   - msg：要编码的消息，可以是 proto.Message 类型或其他类型。
//
// 返回值：
//   - url.Values：编码后的 URL 查询字符串。
//   - error：如果编码过程中发生错误，返回该错误。
func EncodeValues(msg interface{}) (url.Values, error) {
	// 检查 msg 是否为 nil 或者是一个指向 nil 的指针
	if msg == nil || (reflect.ValueOf(msg).Kind() == reflect.Ptr && reflect.ValueOf(msg).IsNil()) {
		// 如果是，则返回一个空的 url.Values 和 nil 错误
		return url.Values{}, nil
	}
	// 尝试将 msg 转换为 proto.Message 类型
	if v, ok := msg.(proto.Message); ok {
		// 创建一个新的 url.Values 对象
		u := make(url.Values)
		// 调用 encodeByField 函数，将消息编码到 URL 查询字符串中
		err := encodeByField(u, "", v.ProtoReflect())
		// 如果发生错误，返回 nil 和该错误
		if err != nil {
			return nil, err
		}
		// 返回编码后的 URL 查询字符串和 nil 错误
		return u, nil
	}
	// 如果 msg 不是 proto.Message 类型，则尝试使用默认的编码器进行编码
	return encoder.Encode(msg)
}

// encodeByField 函数用于将一个 protobuf 消息编码为 URL 查询字符串格式，并将结果存储在 url.Values 中。
// 参数：
//   - u：用于存储编码结果的 url.Values 对象。
//   - path：当前字段的路径，用于构建完整的字段名。
//   - m：要编码的 protobuf 消息。
//
// 返回值：
//   - error：如果编码过程中发生错误，返回该错误。
func encodeByField(u url.Values, path string, m protoreflect.Message) (finalErr error) {
	// 遍历消息中的每个字段
	m.Range(func(fd protoreflect.FieldDescriptor, v protoreflect.Value) bool {
		var (
			key     string
			newPath string
		)
		// 如果字段有 JSON 名称，则使用 JSON 名称作为键
		if fd.HasJSONName() {
			key = fd.JSONName()
		} else {
			// 否则，使用文本名称作为键
			key = fd.TextName()
		}
		// 如果路径为空，则新路径为键
		if path == "" {
			newPath = key
		} else {
			// 否则，新路径为路径和键的组合
			newPath = path + "." + key
		}
		// 如果字段是一个 oneof 字段的一部分，则检查当前字段是否是活动字段
		if of := fd.ContainingOneof(); of != nil {
			if f := m.WhichOneof(of); f != nil && f != fd {
				return true
			}
		}
		// 根据字段类型进行编码
		switch {
		case fd.IsList():
			// 如果字段是一个列表，则编码每个列表项
			if v.List().Len() > 0 {
				list, err := encodeRepeatedField(fd, v.List())
				if err != nil {
					finalErr = err
					return false
				}
				for _, item := range list {
					u.Add(newPath, item)
				}
			}
		case fd.IsMap():
			// 如果字段是一个映射，则编码每个映射项
			if v.Map().Len() > 0 {
				m, err := encodeMapField(fd, v.Map())
				if err != nil {
					finalErr = err
					return false
				}
				for k, value := range m {
					u.Set(fmt.Sprintf("%s[%s]", newPath, k), value)
				}
			}
		case (fd.Kind() == protoreflect.MessageKind) || (fd.Kind() == protoreflect.GroupKind):
			// 如果字段是一个消息或组，则递归编码该消息
			value, err := encodeMessage(fd.Message(), v)
			if err == nil {
				u.Set(newPath, value)
				return true
			}
			if err = encodeByField(u, newPath, v.Message()); err != nil {
				finalErr = err
				return false
			}
		default:
			// 对于其他类型的字段，直接编码其值
			value, err := EncodeField(fd, v)
			if err != nil {
				finalErr = err
				return false
			}
			u.Set(newPath, value)
		}
		return true
	})
	return
}

// encodeRepeatedField 函数用于将一个 protobuf 消息中的重复字段编码为 URL 查询字符串格式。
// 参数：
//   - fieldDescriptor：要编码的字段描述符。
//   - list：要编码的重复字段列表。
//
// 返回值：
//   - []string：编码后的 URL 查询字符串切片。
//   - error：如果编码过程中发生错误，返回该错误。
func encodeRepeatedField(fieldDescriptor protoreflect.FieldDescriptor, list protoreflect.List) ([]string, error) {
	var values []string
	for i := 0; i < list.Len(); i++ {
		// 对列表中的每个元素进行编码
		value, err := EncodeField(fieldDescriptor, list.Get(i))
		if err != nil {
			// 如果编码过程中发生错误，返回 nil 和该错误
			return nil, err
		}
		// 将编码后的字符串添加到 values 切片中
		values = append(values, value)
	}
	// 返回编码后的字符串切片和 nil 错误
	return values, nil
}

// encodeMapField 函数用于将一个 protobuf 消息中的映射字段编码为 URL 查询字符串格式。
// 参数：
//   - fieldDescriptor：要编码的字段描述符。
//   - mp：要编码的映射字段。
//
// 返回值：
//   - map[string]string：编码后的 URL 查询字符串映射。
//   - error：如果编码过程中发生错误，返回该错误。
func encodeMapField(fieldDescriptor protoreflect.FieldDescriptor, mp protoreflect.Map) (map[string]string, error) {
	// 创建一个新的 map 用于存储编码结果
	m := make(map[string]string)
	// 遍历映射中的每个键值对
	mp.Range(func(k protoreflect.MapKey, v protoreflect.Value) bool {
		// 对键进行编码
		key, err := EncodeField(fieldDescriptor.MapKey(), k.Value())
		if err != nil {
			// 如果编码过程中发生错误，返回 false 以停止遍历
			return false
		}
		// 对值进行编码
		value, err := EncodeField(fieldDescriptor.MapValue(), v)
		if err != nil {
			// 如果编码过程中发生错误，返回 false 以停止遍历
			return false
		}
		// 将编码后的键值对添加到 m 中
		m[key] = value
		// 返回 true 以继续遍历
		return true
	})

	// 返回编码后的映射和 nil 错误
	return m, nil
}

// EncodeField 函数用于将一个 protobuf 消息中的字段编码为 URL 查询字符串格式。
// 参数：
//   - fieldDescriptor：要编码的字段描述符。
//   - value：要编码的值。
//
// 返回值：
//   - string：编码后的 URL 查询字符串。
//   - error：如果编码过程中发生错误，返回该错误。
func EncodeField(fieldDescriptor protoreflect.FieldDescriptor, value protoreflect.Value) (string, error) {
	switch fieldDescriptor.Kind() {
	case protoreflect.BoolKind:
		// 如果字段类型是布尔类型，则将值转换为布尔字符串并返回
		return strconv.FormatBool(value.Bool()), nil
	case protoreflect.EnumKind:
		// 如果字段类型是枚举类型，则检查枚举的全名是否为 "google.protobuf.NullValue"
		if fieldDescriptor.Enum().FullName() == "google.protobuf.NullValue" {
			// 如果是，则返回 nullStr（可能是一个预定义的常量，表示空值）
			return nullStr, nil
		}
		// 否则，获取枚举值的描述，并将其名称转换为字符串并返回
		desc := fieldDescriptor.Enum().Values().ByNumber(value.Enum())
		return string(desc.Name()), nil
	case protoreflect.BytesKind:
		// 如果字段类型是字节类型，则将字节数组编码为 base64 字符串并返回
		return base64.URLEncoding.EncodeToString(value.Bytes()), nil
	case protoreflect.MessageKind, protoreflect.GroupKind:
		// 如果字段类型是消息或组类型，则调用 encodeMessage 函数进行编码并返回结果
		return encodeMessage(fieldDescriptor.Message(), value)
	default:
		// 对于其他类型的字段，直接将其值转换为字符串并返回
		return value.String(), nil
	}
}

// encodeMessage 函数用于将一个 protobuf 消息编码为 URL 查询字符串格式。
// 参数：
//   - msgDescriptor：要编码的消息描述符。
//   - value：要编码的值。
//
// 返回值：
//   - string：编码后的 URL 查询字符串。
//   - error：如果编码过程中发生错误，返回该错误。
func encodeMessage(msgDescriptor protoreflect.MessageDescriptor, value protoreflect.Value) (string, error) {
	// 根据消息描述符的全名进行不同的处理
	switch msgDescriptor.FullName() {
	case timestampMessageFullname:
		// 如果是时间戳消息，则调用 marshalTimestamp 函数进行编码
		return marshalTimestamp(value.Message())
	case durationMessageFullname:
		// 如果是持续时间消息，则调用 marshalDuration 函数进行编码
		return marshalDuration(value.Message())
	case bytesMessageFullname:
		// 如果是字节消息，则调用 marshalBytes 函数进行编码
		return marshalBytes(value.Message())
	case "google.protobuf.DoubleValue", "google.protobuf.FloatValue", "google.protobuf.Int64Value", "google.protobuf.Int32Value",
		"google.protobuf.UInt64Value", "google.protobuf.UInt32Value", "google.protobuf.BoolValue", "google.protobuf.StringValue":
		// 如果是基本类型值消息，则获取其值字段并返回其字符串表示
		fd := msgDescriptor.Fields()
		v := value.Message().Get(fd.ByName("value"))
		return fmt.Sprint(v.Interface()), nil
	case fieldMaskFullName:
		// 如果是字段掩码消息，则将其路径转换为驼峰命名法并返回逗号分隔的字符串
		m, ok := value.Message().Interface().(*fieldmaskpb.FieldMask)
		if !ok || m == nil {
			return "", nil
		}
		for i, v := range m.Paths {
			m.Paths[i] = jsonCamelCase(v)
		}
		return strings.Join(m.Paths, ","), nil
	default:
		// 如果是不支持的消息类型，则返回错误
		return "", fmt.Errorf("unsupported message type: %q", string(msgDescriptor.FullName()))
	}
}

// EncodeFieldMask 函数用于将字段掩码消息编码为查询字符串格式。
// 参数：
//   - m：要编码的字段掩码消息。
//
// 返回值：
//   - string：编码后的查询字符串，格式为 "name=paths"。
func EncodeFieldMask(m protoreflect.Message) (query string) {
	// 遍历消息中的每个字段
	m.Range(func(fd protoreflect.FieldDescriptor, v protoreflect.Value) bool {
		// 如果字段是一个消息类型，并且其全名是 "google.protobuf.FieldMask"
		if fd.Kind() == protoreflect.MessageKind {
			if msg := fd.Message(); msg.FullName() == fieldMaskFullName {
				// 编码该消息字段
				value, err := encodeMessage(msg, v)
				// 如果编码过程中发生错误，返回 false 以停止遍历
				if err != nil {
					return false
				}
				// 如果字段有 JSON 名称，则使用 JSON 名称作为键
				if fd.HasJSONName() {
					query = fd.JSONName() + "=" + value
				} else {
					// 否则，使用文本名称作为键
					query = fd.TextName() + "=" + value
				}
				// 编码完成，返回 false 以停止遍历
				return false
			}
		}
		// 继续遍历下一个字段
		return true
	})
	// 返回编码后的查询字符串
	return
}

// jsonCamelCase 函数用于将一个使用下划线分隔的字符串（如 snake_case）转换为驼峰命名法的字符串（如 camelCase）。
// 参数：
//   - s：要转换的字符串。
//
// 返回值：
//   - string：转换后的驼峰命名法字符串。
func jsonCamelCase(s string) string {
	var b []byte
	var wasUnderscore bool
	for i := 0; i < len(s); i++ { // proto identifiers are always ASCII
		c := s[i]
		if c != '_' {
			if wasUnderscore && isASCIILower(c) {
				c -= 'a' - 'A' // convert to uppercase
			}
			b = append(b, c)
		}
		wasUnderscore = c == '_'
	}
	return string(b)
}

// isASCIILower 函数用于检查一个字符是否是 ASCII 编码中的小写字母。
// 参数：
//   - c：要检查的字符。
//
// 返回值：
//   - bool：如果字符是小写字母，返回 true；否则返回 false。
func isASCIILower(c byte) bool {
	return 'a' <= c && c <= 'z'
}
