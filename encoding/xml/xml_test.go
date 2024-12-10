package xml

import (
	"reflect"
	"strings"
	"testing"
)

// Plain 是一个简单的结构体，包含一个接口类型的字段 V
type Plain struct {
	V interface{}
}

// NestedOrder 是一个嵌套的结构体，包含三个字符串字段 Field1、Field2 和 Field3
type NestedOrder struct {
	XMLName struct{} `xml:"result"`
	Field1  string   `xml:"parent>c"`
	Field2  string   `xml:"parent>b"`
	Field3  string   `xml:"parent>a"`
}

// TestCodec_Marshal 测试 codec 的 Marshal 方法
func TestCodec_Marshal(t *testing.T) {
	tests := []struct {
		Value     interface{}
		ExpectXML string
	}{
		// 测试值类型
		{Value: &Plain{true}, ExpectXML: `<Plain><V>true</V></Plain>`},
		{Value: &Plain{false}, ExpectXML: `<Plain><V>false</V></Plain>`},
		{Value: &Plain{42}, ExpectXML: `<Plain><V>42</V></Plain>`},
		{
			Value: &NestedOrder{Field1: "C", Field2: "B", Field3: "A"},
			ExpectXML: `<result>` +
				`<parent>` +
				`<c>C</c>` +
				`<b>B</b>` +
				`<a>A</a>` +
				`</parent>` +
				`</result>`,
		},
	}
	for _, tt := range tests {
		data, err := (codec{}).Marshal(tt.Value)
		if err != nil {
			t.Errorf("marshal(%#v): %s", tt.Value, err)
		}
		if got, want := string(data), tt.ExpectXML; got != want {
			if strings.Contains(want, "\n") {
				t.Errorf("marshal(%#v):\nHAVE:\n%s\nWANT:\n%s", tt.Value, got, want)
			} else {
				t.Errorf("marshal(%#v):\nhave %#q\nwant %#q", tt.Value, got, want)
			}
		}
	}
}

// TestCodec_Unmarshal 测试 codec 的 Unmarshal 方法
func TestCodec_Unmarshal(t *testing.T) {
	tests := []struct {
		want     interface{}
		InputXML string
	}{
		{
			want: &NestedOrder{Field1: "C", Field2: "B", Field3: "A"},
			InputXML: `<result>` +
				`<parent>` +
				`<c>C</c>` +
				`<b>B</b>` +
				`<a>A</a>` +
				`</parent>` +
				`</result>`,
		},
	}

	for _, tt := range tests {
		vt := reflect.TypeOf(tt.want)
		dest := reflect.New(vt.Elem()).Interface()
		data := []byte(tt.InputXML)
		err := (codec{}).Unmarshal(data, dest)
		if err != nil {
			t.Errorf("unmarshal(%#v, %#v): %s", tt.InputXML, dest, err)
		}
		if got, want := dest, tt.want; !reflect.DeepEqual(got, want) {
			t.Errorf("unmarshal(%q):\nhave %#v\nwant %#v", tt.InputXML, got, want)
		}
	}
}

// TestCodec_NilUnmarshal 测试 codec 的 Unmarshal 方法，当目标为 nil 时
func TestCodec_NilUnmarshal(t *testing.T) {
	tests := []struct {
		want     interface{}
		InputXML string
	}{
		{
			want: &NestedOrder{Field1: "C", Field2: "B", Field3: "A"},
			InputXML: `<result>` +
				`<parent>` +
				`<c>C</c>` +
				`<b>B</b>` +
				`<a>A</a>` +
				`</parent>` +
				`</result>`,
		},
	}

	for _, tt := range tests {
		s := struct {
			A string `xml:"a"`
			B *NestedOrder
		}{A: "a"}
		data := []byte(tt.InputXML)
		err := (codec{}).Unmarshal(data, &s.B)
		if err != nil {
			t.Errorf("unmarshal(%#v, %#v): %s", tt.InputXML, s.B, err)
		}
		if got, want := s.B, tt.want; !reflect.DeepEqual(got, want) {
			t.Errorf("unmarshal(%q):\nhave %#v\nwant %#v", tt.InputXML, got, want)
		}
	}
}
