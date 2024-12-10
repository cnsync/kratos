package config

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strconv"
	"sync/atomic"
	"time"

	"google.golang.org/protobuf/proto"

	kratosjson "github.com/cnsync/kratos/encoding/json"
)

var (
	// 确保 atomicValue 和 errValue 结构体实现了 Value 接口
	_ Value = (*atomicValue)(nil)
	_ Value = (*errValue)(nil)
)

// Value 是配置值的接口
type Value interface {
	Bool() (bool, error)
	Int() (int64, error)
	Float() (float64, error)
	String() (string, error)
	Duration() (time.Duration, error)
	Slice() ([]Value, error)
	Map() (map[string]Value, error)
	Scan(interface{}) error
	Load() interface{}
	Store(interface{})
}

// atomicValue 是一个原子值，它实现了 Value 接口
type atomicValue struct {
	atomic.Value
}

// typeAssertError 返回一个错误，表明类型断言失败
func (v *atomicValue) typeAssertError() error {
	return fmt.Errorf("type assert to %v failed", reflect.TypeOf(v.Load()))
}

// Bool 返回值的布尔表示
func (v *atomicValue) Bool() (bool, error) {
	switch val := v.Load().(type) {
	case bool:
		return val, nil
	case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64, float32, float64, string:
		return strconv.ParseBool(fmt.Sprint(val))
	}
	return false, v.typeAssertError()
}

// Int 返回值的整数表示
func (v *atomicValue) Int() (int64, error) {
	switch val := v.Load().(type) {
	case int:
		return int64(val), nil
	case int8:
		return int64(val), nil
	case int16:
		return int64(val), nil
	case int32:
		return int64(val), nil
	case int64:
		return val, nil
	case uint:
		return int64(val), nil
	case uint8:
		return int64(val), nil
	case uint16:
		return int64(val), nil
	case uint32:
		return int64(val), nil
	case uint64:
		return int64(val), nil
	case float32:
		return int64(val), nil
	case float64:
		return int64(val), nil
	case string:
		return strconv.ParseInt(val, 10, 64)
	}
	return 0, v.typeAssertError()
}

// Slice 返回值的切片表示
func (v *atomicValue) Slice() ([]Value, error) {
	vals, ok := v.Load().([]interface{})
	if !ok {
		return nil, v.typeAssertError()
	}
	slices := make([]Value, 0, len(vals))
	// 遍历切片 vals 中的每个元素
	for _, val := range vals {
		// 为每个元素创建一个新的 atomicValue 实例
		a := new(atomicValue)
		// 将原始值存储在 atomicValue 实例中
		a.Store(val)
		// 将 atomicValue 实例添加到新的切片 slices 中
		slices = append(slices, a)
	}
	// 返回转换后的切片和 nil 表示没有错误
	return slices, nil
}

// Map 返回值的映射表示
func (v *atomicValue) Map() (map[string]Value, error) {
	vals, ok := v.Load().(map[string]interface{})
	if !ok {
		return nil, v.typeAssertError()
	}
	// 将 map[string]interface{} 转换为 map[string]Value
	m := make(map[string]Value, len(vals))
	for key, val := range vals {
		// 为每个值创建一个新的 atomicValue 实例
		a := new(atomicValue)
		// 将原始值存储在 atomicValue 实例中
		a.Store(val)
		// 将 atomicValue 实例添加到新的 map 中
		m[key] = a
	}
	// 返回转换后的 map 和 nil 表示没有错误
	return m, nil
}

// Float 返回值的浮点表示
func (v *atomicValue) Float() (float64, error) {
	switch val := v.Load().(type) {
	case int:
		return float64(val), nil
	case int8:
		return float64(val), nil
	case int16:
		return float64(val), nil
	case int32:
		return float64(val), nil
	case int64:
		return float64(val), nil
	case uint:
		return float64(val), nil
	case uint8:
		return float64(val), nil
	case uint16:
		return float64(val), nil
	case uint32:
		return float64(val), nil
	case uint64:
		return float64(val), nil
	case float32:
		return float64(val), nil
	case float64:
		return val, nil
	case string:
		return strconv.ParseFloat(val, 64)
	}
	return 0.0, v.typeAssertError()
}

// String 返回值的字符串表示
func (v *atomicValue) String() (string, error) {
	switch val := v.Load().(type) {
	case string:
		return val, nil
	case bool, int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64, float32, float64:
		return fmt.Sprint(val), nil
	case []byte:
		return string(val), nil
	case fmt.Stringer:
		return val.String(), nil
	}
	return "", v.typeAssertError()
}

// Duration 返回值的持续时间表示
func (v *atomicValue) Duration() (time.Duration, error) {
	val, err := v.Int()
	if err != nil {
		return 0, err
	}
	return time.Duration(val), nil
}

// Scan 将值扫描到目标对象中
func (v *atomicValue) Scan(obj interface{}) error {
	data, err := json.Marshal(v.Load())
	if err != nil {
		return err
	}
	if pb, ok := obj.(proto.Message); ok {
		return kratosjson.UnmarshalOptions.Unmarshal(data, pb)
	}
	return json.Unmarshal(data, obj)
}

// errValue 是一个错误值，它实现了 Value 接口
type errValue struct {
	err error
}

// Bool 返回错误
func (v errValue) Bool() (bool, error) { return false, v.err }

// Int 返回错误
func (v errValue) Int() (int64, error) { return 0, v.err }

// Float 返回错误
func (v errValue) Float() (float64, error) { return 0.0, v.err }

// Duration 返回错误
func (v errValue) Duration() (time.Duration, error) { return 0, v.err }

// String 返回错误
func (v errValue) String() (string, error) { return "", v.err }

// Scan 返回错误
func (v errValue) Scan(interface{}) error { return v.err }

// Load 返回 nil
func (v errValue) Load() interface{} { return nil }

// Store 不做任何操作
func (v errValue) Store(interface{}) {}

// Slice 返回错误
func (v errValue) Slice() ([]Value, error) { return nil, v.err }

// Map 返回错误
func (v errValue) Map() (map[string]Value, error) { return nil, v.err }
