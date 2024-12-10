package yaml

import (
	"math"
	"reflect"
	"testing"
)

// TestCodec_Unmarshal 测试 codec 的 Unmarshal 方法
func TestCodec_Unmarshal(t *testing.T) {
	tests := []struct {
		data  string
		value interface{}
	}{
		// 测试空数据
		{"", (*struct{})(nil)},
		// 测试空对象
		{"{}", &struct{}{}},
		// 测试字符串值
		{"v: hi", map[string]string{"v": "hi"}},
		// 测试字符串值
		{"v: hi", map[string]interface{}{"v": "hi"}},
		// 测试布尔值
		{"v: true", map[string]string{"v": "true"}},
		// 测试布尔值
		{"v: true", map[string]interface{}{"v": true}},
		// 测试整数值
		{"v: 10", map[string]interface{}{"v": 10}},
		// 测试二进制整数值
		{"v: 0b10", map[string]interface{}{"v": 2}},
		// 测试十六进制整数值
		{"v: 0xA", map[string]interface{}{"v": 10}},
		// 测试大整数值
		{"v: 4294967296", map[string]int64{"v": 4294967296}},
		// 测试小数值
		{"v: 0.1", map[string]interface{}{"v": 0.1}},
		// 测试小数值
		{"v:.1", map[string]interface{}{"v": 0.1}},
		// 测试正无穷大值
		{"v:.Inf", map[string]interface{}{"v": math.Inf(+1)}},
		// 测试负无穷大值
		{"v: -.Inf", map[string]interface{}{"v": math.Inf(-1)}},
		// 测试负整数值
		{"v: -10", map[string]interface{}{"v": -10}},
		// 测试负小数值
		{"v: -.1", map[string]interface{}{"v": -0.1}},
	}
	for _, tt := range tests {
		v := reflect.ValueOf(tt.value).Type()
		value := reflect.New(v)
		err := (codec{}).Unmarshal([]byte(tt.data), value.Interface())
		if err != nil {
			t.Fatalf("(codec{}).Unmarshal should not return err")
		}
	}
	// 测试嵌套对象
	spec := struct {
		A string
		B map[string]interface{}
	}{A: "a"}
	err := (codec{}).Unmarshal([]byte("v: hi"), &spec.B)
	if err != nil {
		t.Fatalf("(codec{}).Unmarshal should not return err")
	}
}

// TestCodec_Marshal 测试 codec 的 Marshal 方法
func TestCodec_Marshal(t *testing.T) {
	value := map[string]string{"v": "hi"}
	got, err := (codec{}).Marshal(value)
	if err != nil {
		t.Fatalf("should not return err")
	}
	if string(got) != "v: hi\n" {
		t.Fatalf("want \"v: hi\n\" return \"%s\"", string(got))
	}
}
