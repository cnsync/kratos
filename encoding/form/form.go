package form

import (
	"net/url"
	"reflect"

	"github.com/cnsync/kratos/encoding"
	"github.com/go-playground/form/v4"
	"google.golang.org/protobuf/proto"
)

const (
	// Name 是表单编解码器的名称
	Name = "x-www-form-urlencoded"
	// nullStr 是一个表示空值的字符串
	nullStr = "null"
)

var (
	// 创建表单编码器和解码器实例
	encoder = form.NewEncoder() // 用于将数据编码为表单格式
	decoder = form.NewDecoder() // 用于将表单数据解码为结构体
)

// 可以通过 -ldflags 传递该变量的值，例如：
// go build "-ldflags=-X grove/encoding/form.tagName=form"
var tagName = "json" // 用于标记结构体字段的标签名，默认是 "json"

func init() {
	// 设置编码器和解码器使用的标签名称
	decoder.SetTagName(tagName)
	encoder.SetTagName(tagName)
	// 注册表单编解码器
	encoding.RegisterCodec(codec{encoder: encoder, decoder: decoder})
}

// codec 结构体实现了 form 编解码器的功能
type codec struct {
	encoder *form.Encoder // 表单编码器
	decoder *form.Decoder // 表单解码器
}

// Marshal 方法将数据编码为表单格式（x-www-form-urlencoded）
func (c codec) Marshal(v interface{}) ([]byte, error) {
	var vs url.Values // 存储编码后的表单数据
	var err error

	// 如果 v 是 protobuf 消息类型，则进行专门的编码
	if m, ok := v.(proto.Message); ok {
		vs, err = EncodeValues(m)
		if err != nil {
			return nil, err
		}
	} else {
		// 否则使用默认的表单编码器进行编码
		vs, err = c.encoder.Encode(v)
		if err != nil {
			return nil, err
		}
	}

	// 删除表单中值为空的字段
	for k, v := range vs {
		if len(v) == 0 {
			delete(vs, k)
		}
	}

	// 返回编码后的表单数据
	return []byte(vs.Encode()), nil
}

// Unmarshal 方法将表单数据解码为指定类型的对象
func (c codec) Unmarshal(data []byte, v interface{}) error {
	// 解析 URL 编码的表单数据
	vs, err := url.ParseQuery(string(data))
	if err != nil {
		return err
	}

	// 获取 v 的反射值
	rv := reflect.ValueOf(v)

	// 如果 v 是指针，解引用它
	for rv.Kind() == reflect.Ptr {
		if rv.IsNil() {
			rv.Set(reflect.New(rv.Type().Elem()))
		}
		rv = rv.Elem()
	}

	// 如果 v 是 protobuf 消息类型，则进行专门的解码
	if m, ok := v.(proto.Message); ok {
		return DecodeValues(m, vs)
	}

	// 如果 v 的类型是 protobuf 消息，则解码
	if m, ok := rv.Interface().(proto.Message); ok {
		return DecodeValues(m, vs)
	}

	// 否则使用表单解码器进行解码
	return c.decoder.Decode(v, vs)
}

// Name 方法返回表单编解码器的名称
func (codec) Name() string {
	return Name
}
