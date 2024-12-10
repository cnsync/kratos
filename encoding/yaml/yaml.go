package yaml

import (
	"github.com/cnsync/kratos/encoding"
	"gopkg.in/yaml.v3"
)

// Name 是为 yaml 编解码器注册的名称。
const Name = "yaml"

func init() {
	// 注册一个名为 yaml 的编解码器
	encoding.RegisterCodec(codec{})
}

// codec 是一个使用 yaml 实现的编解码器
type codec struct{}

// Marshal 方法将一个 Go 语言的值序列化为 YAML 格式的字节切片
func (codec) Marshal(v interface{}) ([]byte, error) {
	// 使用 gopkg.in/yaml.v3 包中的 Marshal 函数将值 v 序列化为 YAML 格式
	return yaml.Marshal(v)
}

// Unmarshal 方法将一个 YAML 格式的字节切片反序列化为 Go 语言中的值
func (codec) Unmarshal(data []byte, v interface{}) error {
	// 使用 gopkg.in/yaml.v3 包中的 Unmarshal 函数将字节切片 data 反序列化为值 v
	return yaml.Unmarshal(data, v)
}

// Name 方法返回编解码器的名称
func (codec) Name() string {
	// 返回编解码器的名称 "yaml"
	return Name
}
