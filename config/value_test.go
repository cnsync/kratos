package config

import (
	"fmt"
	"testing"
	"time"
)

// TestAtomicValue_Bool 测试 atomicValue 类型的 Bool 方法
func TestAtomicValue_Bool(t *testing.T) {
	// 定义一个包含多种类型的切片，这些类型都应该被正确地转换为布尔值
	vlist := []interface{}{"1", "t", "T", "true", "TRUE", "True", true, 1, int32(1)}
	// 遍历切片中的每个元素
	for _, x := range vlist {
		// 创建一个新的 atomicValue 实例
		v := atomicValue{}
		// 将当前元素存储到 atomicValue 实例中
		v.Store(x)
		// 调用 Bool 方法并检查返回的布尔值和错误
		b, err := v.Bool()
		// 如果发生错误，测试失败
		if err != nil {
			t.Fatal(err)
		}
		// 如果返回的布尔值不是 true，测试失败
		if !b {
			t.Fatal("b is not equal to true")
		}
	}

	// 定义一个包含多种类型的切片，这些类型都应该被正确地转换为布尔值 false
	vlist = []interface{}{"0", "f", "F", "false", "FALSE", "False", false, 0, int32(0)}
	// 遍历切片中的每个元素
	for _, x := range vlist {
		// 创建一个新的 atomicValue 实例
		v := atomicValue{}
		// 将当前元素存储到 atomicValue 实例中
		v.Store(x)
		// 调用 Bool 方法并检查返回的布尔值和错误
		b, err := v.Bool()
		// 如果发生错误，测试失败
		if err != nil {
			t.Fatal(err)
		}
		// 如果返回的布尔值不是 false，测试失败
		if b {
			t.Fatal("b is not equal to false")
		}
	}

	// 定义一个包含多种类型的切片，这些类型都不应该被正确地转换为布尔值
	vlist = []interface{}{"bbb", "-1"}
	// 遍历切片中的每个元素
	for _, x := range vlist {
		// 创建一个新的 atomicValue 实例
		v := atomicValue{}
		// 将当前元素存储到 atomicValue 实例中
		v.Store(x)
		// 调用 Bool 方法并检查返回的错误
		_, err := v.Bool()
		// 如果没有发生错误，测试失败
		if err == nil {
			t.Fatal("err is nil")
		}
	}
}

// TestAtomicValue_Int 测试 atomicValue 类型的 Int 方法
func TestAtomicValue_Int(t *testing.T) {
	// 定义一个包含多种类型的切片，这些类型都应该被正确地转换为整数
	vlist := []interface{}{"123123", float64(123123), int64(123123), int32(123123), 123123}
	// 遍历切片中的每个元素
	for _, x := range vlist {
		// 创建一个新的 atomicValue 实例
		v := atomicValue{}
		// 将当前元素存储到 atomicValue 实例中
		v.Store(x)
		// 调用 Int 方法并检查返回的整数值和错误
		b, err := v.Int()
		// 如果发生错误，测试失败
		if err != nil {
			t.Fatal(err)
		}
		// 如果返回的整数值不是 123123，测试失败
		if b != 123123 {
			t.Fatal("b is not equal to 123123")
		}
	}

	// 定义一个包含多种类型的切片，这些类型都不应该被正确地转换为整数
	vlist = []interface{}{"bbb", "-x1", true}
	// 遍历切片中的每个元素
	for _, x := range vlist {
		// 创建一个新的 atomicValue 实例
		v := atomicValue{}
		// 将当前元素存储到 atomicValue 实例中
		v.Store(x)
		// 调用 Int 方法并检查返回的错误
		_, err := v.Int()
		// 如果没有发生错误，测试失败
		if err == nil {
			t.Fatal("err is nil")
		}
	}
}

// TestAtomicValue_Float 测试 atomicValue 类型的 Float 方法
func TestAtomicValue_Float(t *testing.T) {
	// 定义一个包含多种类型的切片，这些类型都应该被正确地转换为浮点数
	vlist := []interface{}{"123123.1", 123123.1}
	// 遍历切片中的每个元素
	for _, x := range vlist {
		// 创建一个新的 atomicValue 实例
		v := atomicValue{}
		// 将当前元素存储到 atomicValue 实例中
		v.Store(x)
		// 调用 Float 方法并检查返回的浮点数值和错误
		b, err := v.Float()
		// 如果发生错误，测试失败
		if err != nil {
			t.Fatal(err)
		}
		// 如果返回的浮点数值不是 123123.1，测试失败
		if b != 123123.1 {
			t.Fatal("b is not equal to 123123.1")
		}
	}

	// 定义一个包含多种类型的切片，这些类型都不应该被正确地转换为浮点数
	vlist = []interface{}{"bbb", "-x1"}
	// 遍历切片中的每个元素
	for _, x := range vlist {
		// 创建一个新的 atomicValue 实例
		v := atomicValue{}
		// 将当前元素存储到 atomicValue 实例中
		v.Store(x)
		// 调用 Float 方法并检查返回的错误
		_, err := v.Float()
		// 如果没有发生错误，测试失败
		if err == nil {
			t.Fatal("err is nil")
		}
	}
}

// ts 是一个自定义结构体，实现了 String 方法
type ts struct {
	Name string
	Age  int
}

// String 方法返回结构体的字符串表示
func (t ts) String() string {
	return fmt.Sprintf("%s%d", t.Name, t.Age)
}

// TestAtomicValue_String 测试 atomicValue 类型的 String 方法
func TestAtomicValue_String(t *testing.T) {
	// 定义一个包含多种类型的切片，这些类型都应该被正确地转换为字符串
	vlist := []interface{}{"1", float64(1), int64(1), 1, int64(1)}
	// 遍历切片中的每个元素
	for _, x := range vlist {
		// 创建一个新的 atomicValue 实例
		v := atomicValue{}
		// 将当前元素存储到 atomicValue 实例中
		v.Store(x)
		// 调用 String 方法并检查返回的字符串和错误
		b, err := v.String()
		// 如果发生错误，测试失败
		if err != nil {
			t.Fatal(err)
		}
		// 如果返回的字符串不是 "1"，测试失败
		if b != "1" {
			t.Fatal("b is not equal to 1")
		}
	}

	// 创建一个新的 atomicValue 实例
	v := atomicValue{}
	// 将布尔值 true 存储到 atomicValue 实例中
	v.Store(true)
	// 调用 String 方法并检查返回的字符串和错误
	b, err := v.String()
	// 如果发生错误，测试失败
	if err != nil {
		t.Fatal(err)
	}
	// 如果返回的字符串不是 "true"，测试失败
	if b != "true" {
		t.Fatal(`b is not equal to "true"`)
	}

	// 创建一个新的 atomicValue 实例
	v = atomicValue{}
	// 将自定义结构体实例存储到 atomicValue 实例中
	v.Store(ts{
		Name: "test",
		Age:  10,
	})
	// 调用 String 方法并检查返回的字符串和错误
	b, err = v.String()
	// 如果发生错误，测试失败
	if err != nil {
		t.Fatal(err)
	}
	// 如果返回的字符串不是 "test10"，测试失败
	if b != "test10" {
		t.Fatal(`b is not equal to "test10"`)
	}
}

// TestAtomicValue_Duration 测试 atomicValue 类型的 Duration 方法
func TestAtomicValue_Duration(t *testing.T) {
	// 定义一个包含多种类型的切片，这些类型都应该被正确地转换为 Duration
	vlist := []interface{}{int64(5)}
	// 遍历切片中的每个元素
	for _, x := range vlist {
		// 创建一个新的 atomicValue 实例
		v := atomicValue{}
		// 将当前元素存储到 atomicValue 实例中
		v.Store(x)
		// 调用 Duration 方法并检查返回的 Duration 值和错误
		b, err := v.Duration()
		// 如果发生错误，测试失败
		if err != nil {
			t.Fatal(err)
		}
		// 如果返回的 Duration 值不是 5 纳秒，测试失败
		if b != time.Duration(5) {
			t.Fatal("b is not equal to time.Duration(5)")
		}
	}
}

// TestAtomicValue_Slice 测试 atomicValue 类型的 Slice 方法
func TestAtomicValue_Slice(t *testing.T) {
	// 定义一个包含多种类型的切片，这些类型都应该被正确地转换为切片
	vlist := []interface{}{int64(5)}
	// 创建一个新的 atomicValue 实例
	v := atomicValue{}
	// 将当前切片存储到 atomicValue 实例中
	v.Store(vlist)
	// 调用 Slice 方法并检查返回的切片和错误
	slices, err := v.Slice()
	// 如果发生错误，测试失败
	if err != nil {
		t.Fatal(err)
	}
	// 遍历返回的切片中的每个元素
	for _, v := range slices {
		// 调用 Duration 方法并检查返回的 Duration 值和错误
		b, err := v.Duration()
		// 如果发生错误，测试失败
		if err != nil {
			t.Fatal(err)
		}
		// 如果返回的 Duration 值不是 5 纳秒，测试失败
		if b != time.Duration(5) {
			t.Fatal("b is not equal to time.Duration(5)")
		}
	}
}

// TestAtomicValue_Map 测试 atomicValue 类型的 Map 方法
func TestAtomicValue_Map(t *testing.T) {
	// 定义一个包含多种类型的映射，这些类型都应该被正确地转换为映射
	vlist := make(map[string]interface{})
	vlist["5"] = int64(5)
	vlist["text"] = "text"
	// 创建一个新的 atomicValue 实例
	v := atomicValue{}
	// 将当前映射存储到 atomicValue 实例中
	v.Store(vlist)
	// 调用 Map 方法并检查返回的映射和错误
	m, err := v.Map()
	// 如果发生错误，测试失败
	if err != nil {
		t.Fatal(err)
	}
	// 遍历返回的映射中的每个键值对
	for k, v := range m {
		// 如果键为 "5"，调用 Duration 方法并检查返回的 Duration 值和错误
		if k == "5" {
			b, err := v.Duration()
			// 如果发生错误，测试失败
			if err != nil {
				t.Fatal(err)
			}
			// 如果返回的 Duration 值不是 5 纳秒，测试失败
			if b != time.Duration(5) {
				t.Fatal("b is not equal to time.Duration(5)")
			}
			// 如果键为 "text"，调用 String 方法并检查返回的字符串和错误
		} else {
			b, err := v.String()
			// 如果发生错误，测试失败
			if err != nil {
				t.Fatal(err)
			}
			// 如果返回的字符串不是 "text"，测试失败
			if b != "text" {
				t.Fatal(`b is not equal to "text"`)
			}
		}
	}
}

// TestAtomicValue_Scan 测试 atomicValue 类型的 Scan 方法
func TestAtomicValue_Scan(t *testing.T) {
	// 创建一个新的 atomicValue 实例
	v := atomicValue{}
	// 调用 Scan 方法并检查返回的错误
	err := v.Scan(&struct {
		A string `json:"a"`
	}{"a"})
	// 如果发生错误，测试失败
	if err != nil {
		t.Fatal(err)
	}

	// 调用 Scan 方法并检查返回的错误
	err = v.Scan(&struct {
		A string `json:"a"`
	}{"a"})
	// 如果发生错误，测试失败
	if err != nil {
		t.Fatal(err)
	}
}
