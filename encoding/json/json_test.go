package json

import (
	"encoding/json"
	"reflect"
	"strings"
	"testing"

	testData "github.com/cnsync/kratos/internal/testdata/encoding"
)

// 定义一个嵌套结构体 testEmbed，包含三个整型字段，分别名为 a、b、c
type testEmbed struct {
	Level1a int `json:"a"`
	Level1b int `json:"b"`
	Level1c int `json:"c"`
}

// 定义一个结构体 testMessage，包含三个字符串字段，分别名为 a、b、c，以及一个嵌套结构体 testEmbed 类型的指针字段 embed
type testMessage struct {
	Field1 string     `json:"a"`
	Field2 string     `json:"b"`
	Field3 string     `json:"c"`
	Embed  *testEmbed `json:"embed,omitempty"`
}

// 定义一个结构体 mock，包含一个整型字段 value
type mock struct {
	value int
}

// 定义一个常量枚举，包含三个值：Unknown、Gopher、Zebra
const (
	Unknown = iota
	Gopher
	Zebra
)

// 实现了 json.Unmarshaler 接口，用于将 JSON 数据解析到 mock 结构体中
func (a *mock) UnmarshalJSON(b []byte) error {
	var s string
	if err := json.Unmarshal(b, &s); err != nil {
		return err
	}
	switch strings.ToLower(s) {
	default:
		a.value = Unknown
	case "gopher":
		a.value = Gopher
	case "zebra":
		a.value = Zebra
	}

	return nil
}

// 实现了 json.Marshaler 接口，用于将 mock 结构体编码为 JSON 数据
func (a *mock) MarshalJSON() ([]byte, error) {
	var s string
	switch a.value {
	default:
		s = "unknown"
	case Gopher:
		s = "gopher"
	case Zebra:
		s = "zebra"
	}

	return json.Marshal(s)
}

// 测试 codec 结构体的 Marshal 方法
func TestJSON_Marshal(t *testing.T) {
	tests := []struct {
		input  interface{}
		expect string
	}{
		// 测试空的 testMessage 结构体的 JSON 编码
		{
			input:  &testMessage{},
			expect: `{"a":"","b":"","c":""}`,
		},
		// 测试包含字段值的 testMessage 结构体的 JSON 编码
		{
			input:  &testMessage{Field1: "a", Field2: "b", Field3: "c"},
			expect: `{"a":"a","b":"b","c":"c"}`,
		},
		// 测试包含嵌套结构体的 testData.TestModel 结构体的 JSON 编码
		{
			input:  &testData.TestModel{Id: 1, Name: "go-kratos", Hobby: []string{"1", "2"}},
			expect: `{"id":"1","name":"go-kratos","hobby":["1","2"],"attrs":{}}`,
		},
		// 测试 mock 结构体的 JSON 编码
		{
			input:  &mock{value: Gopher},
			expect: `"gopher"`,
		},
	}
	for _, v := range tests {
		// 调用 codec 结构体的 Marshal 方法，将输入数据编码为 JSON 格式
		data, err := (codec{}).Marshal(v.input)
		if err != nil {
			t.Errorf("marshal(%#v): %s", v.input, err)
		}
		// 检查编码后的 JSON 字符串是否与预期一致
		if got, want := string(data), v.expect; strings.ReplaceAll(got, " ", "") != want {
			if strings.Contains(want, "\n") {
				t.Errorf("marshal(%#v):\nHAVE:\n%s\nWANT:\n%s", v.input, got, want)
			} else {
				t.Errorf("marshal(%#v):\nhave %#q\nwant %#q", v.input, got, want)
			}
		}
	}
}

// 测试 codec 结构体的 Unmarshal 方法
func TestJSON_Unmarshal(t *testing.T) {
	p := testMessage{}
	p2 := testData.TestModel{}
	p3 := &testData.TestModel{}
	p4 := &mock{}
	tests := []struct {
		input  string
		expect interface{}
	}{
		// 测试空的 JSON 字符串解码到 testMessage 结构体
		{
			input:  `{"a":"","b":"","c":""}`,
			expect: &testMessage{},
		},
		// 测试包含字段值的 JSON 字符串解码到 testMessage 结构体
		{
			input:  `{"a":"a","b":"b","c":"c"}`,
			expect: &p,
		},
		// 测试包含嵌套结构体的 JSON 字符串解码到 testData.TestModel 结构体
		{
			input:  `{"id":"1","name":"go-kratos","hobby":["1","2"],"attrs":{}}`,
			expect: &p2,
		},
		// 测试包含字段值的 JSON 字符串解码到 testData.TestModel 结构体指针
		{
			input:  `{"id":1,"name":"go-kratos","hobby":["1","2"]}`,
			expect: &p3,
		},
		// 测试 JSON 字符串解码到 mock 结构体
		{
			input:  `"zebra"`,
			expect: p4,
		},
	}
	for _, v := range tests {
		want := []byte(v.input)
		// 调用 codec 结构体的 Unmarshal 方法，将 JSON 字符串解码到目标结构体
		err := (codec{}).Unmarshal(want, v.expect)
		if err != nil {
			t.Errorf("marshal(%#v): %s", v.input, err)
		}
		// 检查解码后的结构体是否与预期一致
		got, err := codec{}.Marshal(v.expect)
		if err != nil {
			t.Errorf("marshal(%#v): %s", v.input, err)
		}
		if !reflect.DeepEqual(strings.ReplaceAll(string(got), " ", ""), strings.ReplaceAll(string(want), " ", "")) {
			t.Errorf("marshal(%#v):\nhave %#q\nwant %#q", v.input, got, want)
		}
	}
}
