package xml

import (
	"encoding/xml"
	"github.com/cnsync/kratos/encoding"
)

// Name 是为 xml 编解码器注册的名称。
const Name = "xml"

func init() {
	// 注册一个名为 xml 的编解码器
	encoding.RegisterCodec(codec{})
}

// codec 是一个使用 xml 实现的编解码器
type codec struct{}

// Marshal 方法将一个 Go 语言的值序列化为 XML 格式的字节切片
func (codec) Marshal(v interface{}) ([]byte, error) {
	// 使用 encoding/xml 包中的 Marshal 函数将值 v 序列化为 XML 格式
	return xml.Marshal(v)
}

// Unmarshal 方法将一个 XML 格式的字节切片反序列化为 Go 语言中的值
func (codec) Unmarshal(data []byte, v interface{}) error {
	// 使用 encoding/xml 包中的 Unmarshal 函数将字节切片 data 反序列化为值 v
	return xml.Unmarshal(data, v)
}

// Name 方法返回编解码器的名称
func (codec) Name() string {
	// 返回编解码器的名称 "xml"
	return Name
}
