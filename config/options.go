package config

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/cnsync/kratos/encoding"
)

// Decoder 是配置的解码器类型，接受一个 *KeyValue 和目标 map，进行解码。
type Decoder func(*KeyValue, map[string]interface{}) error

// Resolver 是占位符解析器类型，用于解析配置中的占位符。
type Resolver func(map[string]interface{}) error

// Merge 是配置的合并函数类型，用于将源配置合并到目标配置中。
type Merge func(dst, src interface{}) error

// Option 是配置的选项函数，用于设置 options。
type Option func(*options)

type options struct {
	sources  []Source // 配置来源数组
	decoder  Decoder  // 配置解码器
	resolver Resolver // 占位符解析器
	merge    Merge    // 合并函数
}

// WithSource 设置配置来源。
func WithSource(s ...Source) Option {
	return func(o *options) {
		o.sources = s
	}
}

// WithDecoder 设置自定义的配置解码器。
// 默认解码器行为：
// 如果 KeyValue.Format 不为空，则使用指定的格式解码 Value 为 map[string]interface{}。
// 如果 KeyValue.Format 为空，则直接存储 {KeyValue.Key : KeyValue.Value}。
func WithDecoder(d Decoder) Option {
	return func(o *options) {
		o.decoder = d
	}
}

// WithResolveActualTypes 配置占位符解析器，启用将配置值转换为实际数据类型的功能。
func WithResolveActualTypes(enableConvertToType bool) Option {
	return func(o *options) {
		o.resolver = newActualTypesResolver(enableConvertToType)
	}
}

// WithResolver 设置自定义的占位符解析器。
func WithResolver(r Resolver) Option {
	return func(o *options) {
		o.resolver = r
	}
}

// WithMergeFunc 设置自定义的合并函数。
func WithMergeFunc(m Merge) Option {
	return func(o *options) {
		o.merge = m
	}
}

// defaultDecoder 函数用于将源 KeyValue 配置解码到目标 map[string]interface{} 中，使用 src.Format 编码格式。
func defaultDecoder(src *KeyValue, target map[string]interface{}) error {
	// 如果 src.Format 为空，则将 src.Key 展开为嵌套的 map 结构，并将 src.Value 作为最终键的值。
	if src.Format == "" {
		// 使用 strings.Split 将 src.Key 分割成多个部分，以 "." 为分隔符。
		keys := strings.Split(src.Key, ".")
		// 遍历分割后的键列表。
		for i, k := range keys {
			// 如果当前键是列表中的最后一个，则将其值设置为 src.Value。
			if i == len(keys)-1 {
				target[k] = src.Value
			} else {
				// 否则，创建一个新的 map[string]interface{} 作为子映射。
				sub := make(map[string]interface{})
				// 将当前键的值设置为子映射。
				target[k] = sub
				// 更新 target 为子映射，以便在下一次迭代中继续处理下一级键。
				target = sub
			}
		}
		// 返回 nil，表示解码成功。
		return nil
	}
	// 如果 src.Format 不为空，则尝试获取对应的编码解码器。
	if codec := encoding.GetCodec(src.Format); codec != nil {
		// 使用获取到的编码解码器将 src.Value 解码到 target 中。
		return codec.Unmarshal(src.Value, &target)
	}
	return fmt.Errorf("不支持的键: %s 格式: %s", src.Key, src.Format)
}

// newActualTypesResolver 创建一个解析器，根据需要将值转换为实际数据类型。
func newActualTypesResolver(enableConvertToType bool) func(map[string]interface{}) error {
	return func(input map[string]interface{}) error {
		mapper := mapper(input)
		return resolver(input, mapper, enableConvertToType)
	}
}

// defaultResolver 函数用于解析 map 类型的配置数据中的占位符，占位符的格式为 ${key:default}。
func defaultResolver(input map[string]interface{}) error {
	// 调用 mapper 函数，根据输入的 map 创建一个映射函数，用于将占位符中的 key 映射到实际的值。
	mapper := mapper(input)
	// 调用 resolver 函数，使用创建的映射函数解析输入的 map 中的占位符。
	return resolver(input, mapper, false)
}

// resolver 是通用解析函数，用于递归解析占位符并替换。
func resolver(input map[string]interface{}, mapper func(name string) string, toType bool) error {
	var resolve func(map[string]interface{}) error
	resolve = func(sub map[string]interface{}) error {
		for k, v := range sub {
			switch vt := v.(type) {
			case string:
				sub[k] = expand(vt, mapper, toType)
			case map[string]interface{}:
				if err := resolve(vt); err != nil {
					return err
				}
			case []interface{}:
				for i, iface := range vt {
					switch it := iface.(type) {
					case string:
						vt[i] = expand(it, mapper, toType)
					case map[string]interface{}:
						if err := resolve(it); err != nil {
							return err
						}
					}
				}
				sub[k] = vt
			}
		}
		return nil
	}
	return resolve(input)
}

// mapper 返回一个函数，用于映射占位符名称为对应的值。
func mapper(input map[string]interface{}) func(name string) string {
	mapper := func(name string) string {
		args := strings.SplitN(strings.TrimSpace(name), ":", 2) // 支持占位符的默认值
		if v, has := readValue(input, args[0]); has {
			s, _ := v.String()
			return s
		} else if len(args) > 1 { // 提供默认值
			return args[1]
		}
		return ""
	}
	return mapper
}

// convertToType 尝试将字符串转换为具体的数据类型（如 bool、int64、float64 或 string）。
func convertToType(input string) interface{} {
	if strings.HasPrefix(input, "\"") && strings.HasSuffix(input, "\"") {
		return strings.Trim(input, "\"") // 如果是带引号的字符串，去掉引号返回。
	}
	if input == "true" || input == "false" {
		b, _ := strconv.ParseBool(input)
		return b
	}
	// 尝试将输入转换为 float64
	// 如果字符串中包含小数点，则尝试转换为 float64
	if strings.Contains(input, ".") {
		// 调用 strconv.ParseFloat 函数将字符串转换为 float64
		// 第二个参数 64 表示转换为 64 位浮点数
		// 如果转换成功，将结果赋值给 f
		if f, err := strconv.ParseFloat(input, 64); err == nil {
			// 返回转换后的 float64 类型的值
			return f
		}
	}

	// 尝试将输入转换为 int64
	// 调用 strconv.ParseInt 函数将字符串转换为 int64
	// 第二个参数 10 表示十进制
	// 第三个参数 64 表示转换为 64 位整数
	// 如果转换成功，将结果赋值给 i
	if i, err := strconv.ParseInt(input, 10, 64); err == nil {
		// 返回转换后的 int64 类型的值
		return i
	}

	// 如果无法转换为其他类型，则默认返回字符串
	return input
}

// expand 函数用于解析字符串 s 中的占位符，并根据 mapping 函数提供的值进行替换。
// 参数 s 是要解析的字符串，mapping 是一个函数，用于将占位符中的键映射到实际的值，toType 是一个布尔值，指示是否将替换后的值转换为相应的类型。
func expand(s string, mapping func(string) string, toType bool) interface{} {
	// 使用正则表达式 r 查找字符串 s 中所有匹配的子串，并返回一个二维切片 re，其中每个子切片包含两个元素：完整的匹配字符串和捕获组的内容。
	r := regexp.MustCompile(`\${(.*?)}`)
	re := r.FindAllStringSubmatch(s, -1)
	// 初始化一个变量 ct，用于存储转换后的类型。
	var ct interface{}
	// 遍历 re 中的每个匹配项。
	for _, i := range re {
		// 检查每个匹配项是否包含两个元素，即完整的匹配字符串和捕获组的内容。
		if len(i) == 2 { //nolint:mnd
			// 调用 mapping 函数，将捕获组的内容（即占位符中的键）映射到实际的值，并存储在变量 m 中。
			m := mapping(i[1])
			// 如果 toType 为真，则将 m 转换为相应的类型，并存储在变量 ct 中。
			if toType {
				ct = convertToType(m)
				// 返回转换后的类型 ct。
				return ct
			}
			// 如果 toType 为假，则将字符串 s 中的占位符替换为实际的值 m，并存储在变量 s 中。
			s = strings.ReplaceAll(s, i[0], m)
		}
	}
	// 如果 toType 为假，则返回替换后的字符串 s。
	return s
}
