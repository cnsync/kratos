package encoding

import (
	"encoding/xml"
	"runtime/debug"
	"testing"
)

// codec 是一个简单的编解码器实现，用于测试目的。
type codec struct{}

// Marshal 方法实现了 codec 接口的 Marshal 方法，但它会抛出一个 panic，因为它没有被正确实现。
func (c codec) Marshal(_ interface{}) ([]byte, error) {
	panic("implement me")
}

// Unmarshal 方法实现了 codec 接口的 Unmarshal 方法，但它会抛出一个 panic，因为它没有被正确实现。
func (c codec) Unmarshal(_ []byte, _ interface{}) error {
	panic("implement me")
}

// Name 方法返回一个空字符串，因为 codec 结构体没有正确实现 Name 方法。
func (c codec) Name() string {
	return ""
}

// codec2 是一个使用 XML 进行编码和解码的编解码器实现。
type codec2 struct{}

// Marshal 方法实现了 codec 接口的 Marshal 方法，它使用 encoding/xml 包来将给定的值序列化为 XML 格式。
func (codec2) Marshal(v interface{}) ([]byte, error) {
	return xml.Marshal(v)
}

// Unmarshal 方法实现了 codec 接口的 Unmarshal 方法，它使用 encoding/xml 包来将给定的 XML 数据反序列化为指定的值。
func (codec2) Unmarshal(data []byte, v interface{}) error {
	return xml.Unmarshal(data, v)
}

// Name 方法返回编解码器的名称 "xml"。
func (codec2) Name() string {
	return "xml"
}

// TestRegisterCodec 测试了 RegisterCodec 函数的行为。
func TestRegisterCodec(t *testing.T) {
	// 测试注册一个 nil 编解码器是否会导致 panic。
	f := func() { RegisterCodec(nil) }
	funcDidPanic, panicValue, _ := didPanic(f)
	if !funcDidPanic {
		t.Fatalf("func should panic\n\tPanic value:\t%#v", panicValue)
	}
	if panicValue != "cannot register a nil Codec" {
		t.Fatalf("panic error got %s want cannot register a nil Codec", panicValue)
	}

	// 测试注册一个 Name 方法返回空字符串的编解码器是否会导致 panic。
	f = func() {
		RegisterCodec(codec{})
	}
	funcDidPanic, panicValue, _ = didPanic(f)
	if !funcDidPanic {
		t.Fatalf("func should panic\n\tPanic value:\t%#v", panicValue)
	}
	if panicValue != "cannot register Codec with empty string result for Name()" {
		t.Fatalf("panic error got %s want cannot register Codec with empty string result for Name()", panicValue)
	}

	// 测试注册一个有效的编解码器，并检查是否可以正确获取。
	codec := codec2{}
	RegisterCodec(codec)
	got := GetCodec("xml")
	if got != codec {
		t.Fatalf("RegisterCodec(%v) want %v got %v", codec, codec, got)
	}
}

// PanicTestFunc 定义了一个应该传递给 assert.Panics 和 assert.NotPanics 方法的函数，它表示一个不带参数且不返回任何内容的简单函数。
type PanicTestFunc func()

// didPanic 函数用于检查给定的函数是否发生了 panic。如果发生了 panic，它将返回 true，否则返回 false。
func didPanic(f PanicTestFunc) (bool, interface{}, string) {
	didPanic := false
	var message interface{}
	var stack string
	func() {
		defer func() {
			if message = recover(); message != nil {
				didPanic = true
				stack = string(debug.Stack())
			}
		}()

		// 调用目标函数
		f()
	}()

	return didPanic, message, stack
}
