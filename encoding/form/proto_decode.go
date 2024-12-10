package form

import (
	"encoding/base64"
	"errors"
	"fmt"
	"net/url"
	"strconv"
	"strings"
	"time"

	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/reflect/protoregistry"
	"google.golang.org/protobuf/types/known/durationpb"
	"google.golang.org/protobuf/types/known/fieldmaskpb"
	"google.golang.org/protobuf/types/known/structpb"
	"google.golang.org/protobuf/types/known/timestamppb"
	"google.golang.org/protobuf/types/known/wrapperspb"
)

// fieldSeparator 是用于分隔字段路径的分隔符。
const fieldSeparator = "."

// errInvalidFormatMapKey 是一个错误变量，用于表示在解析 URL 查询字符串中的映射字段键时发生的错误。
var errInvalidFormatMapKey = errors.New("invalid formatting for map key")

func DecodeValues(msg proto.Message, values url.Values) error {
	// 遍历 values 中的所有键值对
	for key, values := range values {
		// 将键按照 "." 分割成多个部分，以便于填充到嵌套的消息字段中
		fieldPath := strings.Split(key, ".")
		// 调用 populateFieldValues 函数，将值填充到消息对象的相应字段中
		if err := populateFieldValues(msg.ProtoReflect(), fieldPath, values); err != nil {
			// 如果填充过程中发生错误，返回该错误
			return err
		}
	}
	// 如果所有键值对都成功填充到消息对象中，返回 nil 表示没有错误
	return nil
}

// populateFieldValues 函数用于将值填充到消息对象的字段中。
// 参数：
//   - v：消息对象。
//   - fieldPath：字段路径，以点号分隔的字符串切片。
//   - values：要填充的值，字符串切片。
//
// 返回值：
//   - error：如果填充过程中发生错误，返回该错误。
func populateFieldValues(v protoreflect.Message, fieldPath []string, values []string) error {
	// 如果字段路径长度小于 1，则返回错误。
	if len(fieldPath) < 1 {
		return errors.New("no field path")
	}
	// 如果值的数量小于 1，则返回错误。
	if len(values) < 1 {
		return errors.New("no value provided")
	}

	// 初始化字段描述符。
	var fd protoreflect.FieldDescriptor
	// 遍历字段路径中的每个字段名。
	for i, fieldName := range fieldPath {
		// 获取消息对象中对应字段名的字段描述符。
		if fd = getFieldDescriptor(v, fieldName); fd == nil {
			// 如果字段描述符为空，则忽略该意外字段。
			return nil
		}
		// 如果字段是映射类型且字段路径长度为 2，则填充映射字段。
		if fd.IsMap() && len(fieldPath) == 2 {
			return populateMapField(fd, v.Mutable(fd).Map(), fieldPath, values)
		}
		// 如果已经到达字段路径的最后一个字段，则停止遍历。
		if i == len(fieldPath)-1 {
			break
		}
		// 如果字段不是消息类型或者字段的基数不是重复的，则返回错误。
		if fd.Message() == nil || fd.Cardinality() == protoreflect.Repeated {
			// 如果字段是映射类型且字段路径长度大于 1，则填充映射字段的子字段。
			if fd.IsMap() && len(fieldPath) > 1 {
				// 填充映射字段的子字段。
				return populateMapField(fd, v.Mutable(fd).Map(), []string{fieldPath[1]}, values)
			}
			return fmt.Errorf("invalid path: %q is not a message", fieldName)
		}
		// 获取字段的消息对象，并继续遍历下一个字段。
		v = v.Mutable(fd).Message()
	}
	// 检查字段是否属于某个 oneof 字段。
	if of := fd.ContainingOneof(); of != nil {
		// 检查 oneof 字段是否已经设置了值。
		if f := v.WhichOneof(of); f != nil {
			return fmt.Errorf("field already set for oneof %q", of.FullName().Name())
		}
	}
	// 根据字段类型进行填充。
	switch {
	case fd.IsList():
		// 填充重复字段。
		return populateRepeatedField(fd, v.Mutable(fd).List(), values)
	case fd.IsMap():
		// 填充映射字段。
		return populateMapField(fd, v.Mutable(fd).Map(), fieldPath, values)
	}
	// 如果值的数量大于 1，则返回错误。
	if len(values) > 1 {
		return fmt.Errorf("too many values for field %q: %s", fd.FullName().Name(), strings.Join(values, ", "))
	}
	// 填充单个字段。
	return populateField(fd, v, values[0])
}

// getFieldDescriptor 函数用于获取消息对象中指定字段名的字段描述符。
// 参数：
//   - v：消息对象。
//   - fieldName：字段名。
//
// 返回值：
//   - protoreflect.FieldDescriptor：字段描述符，如果未找到则返回 nil。
func getFieldDescriptor(v protoreflect.Message, fieldName string) protoreflect.FieldDescriptor {
	// 获取消息对象的字段描述符集合。
	var fields = v.Descriptor().Fields()
	// 初始化字段描述符为 nil。
	var fd = getDescriptorByFieldAndName(fields, fieldName)
	// 如果字段描述符为空，则根据不同情况进行处理。
	if fd == nil {
		// 如果消息对象的全名等于 structMessageFullname，则获取 structFieldsFieldNumber 字段的描述符。
		switch {
		case v.Descriptor().FullName() == structMessageFullname:
			fd = fields.ByNumber(structFieldsFieldNumber)
		// 如果字段名以 "[]" 结尾，则去除 "[]" 后再次尝试获取字段描述符。
		case len(fieldName) > 2 && strings.HasSuffix(fieldName, "[]"):
			fd = getDescriptorByFieldAndName(fields, strings.TrimSuffix(fieldName, "[]"))
		default:
			// 如果字段名是 map 类型，则解析出 map 的键名，并获取该键名对应的字段描述符。
			// 例如，对于字段名 "map[kratos]"，解析出键名为 "kratos"，并获取 "map" 字段的描述符。
			field, _, err := parseURLQueryMapKey(fieldName)
			if err != nil {
				// 如果解析失败，则中断处理。
				break
			}
			// 获取解析后的字段名对应的字段描述符。
			fd = getDescriptorByFieldAndName(fields, field)
		}
	}
	// 返回最终获取到的字段描述符。
	return fd
}

// getDescriptorByFieldAndName 函数用于从字段描述符集合中获取指定字段名的字段描述符。
// 参数：
//   - fields：字段描述符集合。
//   - fieldName：字段名。
//
// 返回值：
//   - protoreflect.FieldDescriptor：字段描述符，如果未找到则返回 nil。
func getDescriptorByFieldAndName(fields protoreflect.FieldDescriptors, fieldName string) protoreflect.FieldDescriptor {
	// 初始化字段描述符为 nil。
	var fd protoreflect.FieldDescriptor
	// 首先尝试通过字段名获取字段描述符。
	if fd = fields.ByName(protoreflect.Name(fieldName)); fd == nil {
		// 如果通过字段名未找到，则尝试通过 JSON 名获取字段描述符。
		fd = fields.ByJSONName(fieldName)
	}
	// 返回最终获取到的字段描述符。
	return fd
}

// populateField 函数用于将单个值填充到消息对象的字段中。
// 参数：
//   - fd：字段描述符。
//   - v：消息对象。
//   - value：要填充的值。
//
// 返回值：
//   - error：如果填充过程中发生错误，返回该错误。
func populateField(fd protoreflect.FieldDescriptor, v protoreflect.Message, value string) error {
	// 如果值为空字符串，则直接返回 nil，表示没有错误。
	if value == "" {
		return nil
	}
	// 调用 parseField 函数，将值解析为字段描述符对应类型的值。
	val, err := parseField(fd, value)
	// 如果解析过程中发生错误，则返回该错误。
	if err != nil {
		return fmt.Errorf("parsing field %q: %w", fd.FullName().Name(), err)
	}
	// 将解析后的值设置到消息对象的对应字段中。
	v.Set(fd, val)
	// 如果所有操作都成功，则返回 nil 表示没有错误。
	return nil
}

// populateRepeatedField 函数用于将多个值填充到消息对象的重复字段中。
// 参数：
//   - fd：字段描述符。
//   - list：消息对象的重复字段列表。
//   - values：要填充的值，字符串切片。
//
// 返回值：
//   - error：如果填充过程中发生错误，返回该错误。
func populateRepeatedField(fd protoreflect.FieldDescriptor, list protoreflect.List, values []string) error {
	// 遍历值切片中的每个值。
	for _, value := range values {
		// 调用 parseField 函数，将值解析为字段描述符对应类型的值。
		v, err := parseField(fd, value)
		// 如果解析过程中发生错误，则返回该错误。
		if err != nil {
			return fmt.Errorf("parsing list %q: %w", fd.FullName().Name(), err)
		}
		// 将解析后的值添加到列表中。
		list.Append(v)
	}
	// 如果所有操作都成功，则返回 nil 表示没有错误。
	return nil
}

// populateMapField 函数用于将值填充到消息对象的映射字段中。
// 参数：
//   - fd：字段描述符。
//   - mp：消息对象的映射字段。
//   - fieldPath：字段路径，以点号分隔的字符串切片。
//   - values：要填充的值，字符串切片。
//
// 返回值：
//   - error：如果填充过程中发生错误，返回该错误。
func populateMapField(fd protoreflect.FieldDescriptor, mp protoreflect.Map, fieldPath []string, values []string) error {
	// 解析字段路径，获取映射的键名。
	_, keyName, err := parseURLQueryMapKey(strings.Join(fieldPath, fieldSeparator))
	// 如果解析过程中发生错误，则返回该错误。
	if err != nil {
		return err
	}
	// 解析键名，获取映射的键值。
	key, err := parseField(fd.MapKey(), keyName)
	// 如果解析过程中发生错误，则返回该错误。
	if err != nil {
		return fmt.Errorf("parsing map key %q: %w", fd.FullName().Name(), err)
	}
	// 解析值，获取映射的值。
	value, err := parseField(fd.MapValue(), values[len(values)-1])
	// 如果解析过程中发生错误，则返回该错误。
	if err != nil {
		return fmt.Errorf("parsing map value %q: %w", fd.FullName().Name(), err)
	}
	// 将键值对设置到映射中。
	mp.Set(key.MapKey(), value)
	// 如果所有操作都成功，则返回 nil 表示没有错误。
	return nil
}

// parseField 函数用于将字符串值解析并转换为字段描述符对应类型的值。
// 参数：
//   - fd：字段描述符。
//   - value：要解析的值。
//
// 返回值：
//   - protoreflect.Value：解析后的值。
//   - error：如果解析过程中发生错误，返回该错误。
func parseField(fd protoreflect.FieldDescriptor, value string) (protoreflect.Value, error) {
	switch fd.Kind() {
	case protoreflect.BoolKind:
		// 解析布尔值
		v, err := strconv.ParseBool(value)
		if err != nil {
			return protoreflect.Value{}, err
		}
		return protoreflect.ValueOfBool(v), nil
	case protoreflect.EnumKind:
		// 解析枚举值
		enum, err := protoregistry.GlobalTypes.FindEnumByName(fd.Enum().FullName())
		switch {
		case errors.Is(err, protoregistry.NotFound):
			return protoreflect.Value{}, fmt.Errorf("enum %q is not registered", fd.Enum().FullName())
		case err != nil:
			return protoreflect.Value{}, fmt.Errorf("failed to look up enum: %w", err)
		}
		v := enum.Descriptor().Values().ByName(protoreflect.Name(value))
		if v == nil {
			i, err := strconv.ParseInt(value, 10, 32) //nolint:mnd
			if err != nil {
				return protoreflect.Value{}, fmt.Errorf("%q is not a valid value", value)
			}
			v = enum.Descriptor().Values().ByNumber(protoreflect.EnumNumber(i))
			if v == nil {
				return protoreflect.Value{}, fmt.Errorf("%q is not a valid value", value)
			}
		}
		return protoreflect.ValueOfEnum(v.Number()), nil
	case protoreflect.Int32Kind, protoreflect.Sint32Kind, protoreflect.Sfixed32Kind:
		// 解析 int32、sint32、sfixed32 值
		v, err := strconv.ParseInt(value, 10, 32) //nolint:mnd
		if err != nil {
			return protoreflect.Value{}, err
		}
		return protoreflect.ValueOfInt32(int32(v)), nil
	case protoreflect.Int64Kind, protoreflect.Sint64Kind, protoreflect.Sfixed64Kind:
		// 解析 int64、sint64、sfixed64 值
		v, err := strconv.ParseInt(value, 10, 64) //nolint:mnd
		if err != nil {
			return protoreflect.Value{}, err
		}
		return protoreflect.ValueOfInt64(v), nil
	case protoreflect.Uint32Kind, protoreflect.Fixed32Kind:
		// 解析 uint32、fixed32 值
		v, err := strconv.ParseUint(value, 10, 32) //nolint:mnd
		if err != nil {
			return protoreflect.Value{}, err
		}
		return protoreflect.ValueOfUint32(uint32(v)), nil
	case protoreflect.Uint64Kind, protoreflect.Fixed64Kind:
		// 解析 uint64、fixed64 值
		v, err := strconv.ParseUint(value, 10, 64) //nolint:mnd
		if err != nil {
			return protoreflect.Value{}, err
		}
		return protoreflect.ValueOfUint64(v), nil
	case protoreflect.FloatKind:
		// 解析 float32 值
		v, err := strconv.ParseFloat(value, 32) //nolint:mnd
		if err != nil {
			return protoreflect.Value{}, err
		}
		return protoreflect.ValueOfFloat32(float32(v)), nil
	case protoreflect.DoubleKind:
		// 解析 float64 值
		v, err := strconv.ParseFloat(value, 64) //nolint:mnd
		if err != nil {
			return protoreflect.Value{}, err
		}
		return protoreflect.ValueOfFloat64(v), nil
	case protoreflect.StringKind:
		// 解析字符串值
		return protoreflect.ValueOfString(value), nil
	case protoreflect.BytesKind:
		// 解析字节数组值
		v, err := base64.StdEncoding.DecodeString(value)
		if err != nil {
			return protoreflect.Value{}, err
		}
		return protoreflect.ValueOfBytes(v), nil
	case protoreflect.MessageKind, protoreflect.GroupKind:
		// 解析消息或组值
		return parseMessage(fd.Message(), value)
	default:
		// 如果遇到未知的字段类型，则抛出异常
		panic(fmt.Sprintf("unknown field kind: %v", fd.Kind()))
	}
}

// parseMessage 函数用于将字符串值解析并转换为消息描述符对应类型的消息对象。
// 参数：
//   - md：消息描述符。
//   - value：要解析的值。
//
// 返回值：
//   - protoreflect.Value：解析后的值。
//   - error：如果解析过程中发生错误，返回该错误。
func parseMessage(md protoreflect.MessageDescriptor, value string) (protoreflect.Value, error) {
	var msg proto.Message
	switch md.FullName() {
	case "google.protobuf.Timestamp":
		if value == nullStr {
			break
		}
		t, err := time.ParseInLocation(time.RFC3339Nano, value, time.Local)
		if err != nil {
			return protoreflect.Value{}, err
		}
		msg = timestamppb.New(t)
	case "google.protobuf.Duration":
		if value == nullStr {
			break
		}
		d, err := time.ParseDuration(value)
		if err != nil {
			return protoreflect.Value{}, err
		}
		msg = durationpb.New(d)
	case "google.protobuf.DoubleValue":
		v, err := strconv.ParseFloat(value, 64) //nolint:mnd
		if err != nil {
			return protoreflect.Value{}, err
		}
		msg = wrapperspb.Double(v)
	case "google.protobuf.FloatValue":
		v, err := strconv.ParseFloat(value, 32) //nolint:mnd
		if err != nil {
			return protoreflect.Value{}, err
		}
		msg = wrapperspb.Float(float32(v))
	case "google.protobuf.Int64Value":
		v, err := strconv.ParseInt(value, 10, 64) //nolint:mnd
		if err != nil {
			return protoreflect.Value{}, err
		}
		msg = wrapperspb.Int64(v)
	case "google.protobuf.Int32Value":
		v, err := strconv.ParseInt(value, 10, 32) //nolint:mnd
		if err != nil {
			return protoreflect.Value{}, err
		}
		msg = wrapperspb.Int32(int32(v))
	case "google.protobuf.UInt64Value":
		v, err := strconv.ParseUint(value, 10, 64) //nolint:mnd
		if err != nil {
			return protoreflect.Value{}, err
		}
		msg = wrapperspb.UInt64(v)
	case "google.protobuf.UInt32Value":
		v, err := strconv.ParseUint(value, 10, 32) //nolint:mnd
		if err != nil {
			return protoreflect.Value{}, err
		}
		msg = wrapperspb.UInt32(uint32(v))
	case "google.protobuf.BoolValue":
		v, err := strconv.ParseBool(value)
		if err != nil {
			return protoreflect.Value{}, err
		}
		msg = wrapperspb.Bool(v)
	case "google.protobuf.StringValue":
		msg = wrapperspb.String(value)
	case "google.protobuf.BytesValue":
		v, err := base64.StdEncoding.DecodeString(value)
		if err != nil {
			if v, err = base64.URLEncoding.DecodeString(value); err != nil {
				return protoreflect.Value{}, err
			}
		}
		msg = wrapperspb.Bytes(v)
	case "google.protobuf.FieldMask":
		fm := &fieldmaskpb.FieldMask{}
		for _, fv := range strings.Split(value, ",") {
			fm.Paths = append(fm.Paths, jsonSnakeCase(fv))
		}
		msg = fm
	case "google.protobuf.Value":
		fm, err := structpb.NewValue(value)
		if err != nil {
			return protoreflect.Value{}, err
		}
		msg = fm
	case "google.protobuf.Struct":
		var v structpb.Struct
		if err := protojson.Unmarshal([]byte(value), &v); err != nil {
			return protoreflect.Value{}, err
		}
		msg = &v
	default:
		return protoreflect.Value{}, fmt.Errorf("unsupported message type: %q", string(md.FullName()))
	}
	return protoreflect.ValueOfMessage(msg.ProtoReflect()), nil
}

// jsonSnakeCase 函数将一个字符串从驼峰命名法转换为蛇形命名法。
// 例如，"camelCase" 将被转换为 "camel_case"。
// 参数：
//   - s：要转换的字符串。
//
// 返回值：
//   - string：转换后的蛇形命名法字符串。
func jsonSnakeCase(s string) string {
	var b []byte
	for i := 0; i < len(s); i++ { // proto 标识符总是 ASCII 字符
		c := s[i]
		if isASCIIUpper(c) {
			b = append(b, '_')
			c += 'a' - 'A' // 转换为小写字母
		}
		b = append(b, c)
	}
	return string(b)
}

// isASCIIUpper 函数检查一个字节是否是 ASCII 大写字母。
// 参数：
//   - c：要检查的字节。
//
// 返回值：
//   - bool：如果字节是 ASCII 大写字母，则返回 true；否则返回 false。
func isASCIIUpper(c byte) bool {
	return 'A' <= c && c <= 'Z'
}

// parseURLQueryMapKey 函数用于解析 URL 查询字符串中的映射字段键。
// 它接受一个字符串参数 key，并返回一个三元组，包含映射字段名、键名和可能的错误。
// 参数：
//   - key：要解析的字符串。
//
// 返回值：
//   - string：解析出的映射字段名。
//   - string：解析出的键名。
//   - error：如果解析过程中发生错误，返回该错误。
func parseURLQueryMapKey(key string) (string, string, error) {
	var (
		startIndex = strings.IndexByte(key, '[') // 查找 '[' 的位置
		endIndex   = strings.IndexByte(key, ']') // 查找 ']' 的位置
	)
	if startIndex < 0 {
		// 如果没有找到 '['，则尝试使用字段分隔符 '.' 分割字符串
		values := strings.Split(key, fieldSeparator)
		// 如果分割后的字符串数量不等于 2，则返回错误
		if len(values) != 2 {
			return "", "", errInvalidFormatMapKey
		}
		return values[0], values[1], nil
	}
	// 如果 '[' 的位置小于等于 0，或者 '[' 的位置大于等于 ']' 的位置，或者 key 的长度不等于 ']' 的位置加 1，则返回错误
	if startIndex <= 0 || startIndex >= endIndex || len(key) != endIndex+1 {
		return "", "", errInvalidFormatMapKey
	}
	// 返回解析出的映射字段名和键名
	return key[:startIndex], key[startIndex+1 : endIndex], nil
}
