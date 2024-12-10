package config

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/cnsync/kratos/encoding"

	"dario.cat/mergo"
)

// TestReader_Merge 测试 Reader 的 Merge 方法
func TestReader_Merge(t *testing.T) {
	var (
		err error
		ok  bool
	)
	// 配置选项
	opts := options{
		decoder: func(kv *KeyValue, v map[string]interface{}) error {
			if codec := encoding.GetCodec(kv.Format); codec != nil {
				return codec.Unmarshal(kv.Value, &v)
			}
			return fmt.Errorf("不支持的键: %s 格式: %s", kv.Key, kv.Format)
		},
		resolver: defaultResolver,
		merge: func(dst, src interface{}) error {
			return mergo.Map(dst, src, mergo.WithOverride)
		},
	}
	r := newReader(opts)

	// 测试无效的 JSON 数据
	err = r.Merge(&KeyValue{
		Key:    "a",
		Value:  []byte("bad"),
		Format: "json",
	})
	if err == nil {
		t.Fatal("错误应非空，但得到 nil")
	}

	// 测试有效的 JSON 数据
	err = r.Merge(&KeyValue{
		Key:    "b",
		Value:  []byte(`{"nice": "boat", "x": 1}`),
		Format: "json",
	})
	if err != nil {
		t.Fatal(err)
	}
	vv, ok := r.Value("nice")
	if !ok {
		t.Fatal("未找到键值 'nice'")
	}
	vvv, err := vv.String()
	if err != nil {
		t.Fatal(err)
	}
	if vvv != "boat" {
		t.Fatalf("期望值为 'boat'，但得到 %s", vvv)
	}

	// 测试合并后的覆盖行为
	err = r.Merge(&KeyValue{
		Key:    "b",
		Value:  []byte(`{"x": 2}`),
		Format: "json",
	})
	if err != nil {
		t.Fatal(err)
	}
	vv, ok = r.Value("x")
	if !ok {
		t.Fatal("未找到键值 'x'")
	}
	vvx, err := vv.Int()
	if err != nil {
		t.Fatal(err)
	}
	if vvx != 2 {
		t.Fatalf("期望值为 2，但得到 %d", vvx)
	}
}

// TestReader_Value 测试 Reader 的 Value 方法
func TestReader_Value(t *testing.T) {
	opts := options{
		decoder: func(kv *KeyValue, v map[string]interface{}) error {
			if codec := encoding.GetCodec(kv.Format); codec != nil {
				return codec.Unmarshal(kv.Value, &v)
			}
			return fmt.Errorf("不支持的键: %s 格式: %s", kv.Key, kv.Format)
		},
		resolver: defaultResolver,
		merge: func(dst, src interface{}) error {
			return mergo.Map(dst, src, mergo.WithOverride)
		},
	}

	ymlval := `
a: 
  b: 
    X: 1
    Y: "lol"
    z: true
`
	tests := []struct {
		name string
		kv   KeyValue
	}{
		{
			name: "JSON 数据",
			kv: KeyValue{
				Key:    "config",
				Value:  []byte(`{"a": {"b": {"X": 1, "Y": "lol", "z": true}}}`),
				Format: "json",
			},
		},
		{
			name: "YAML 数据",
			kv: KeyValue{
				Key:    "config",
				Value:  []byte(ymlval),
				Format: "yaml",
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			r := newReader(opts)
			err := r.Merge(&test.kv)
			if err != nil {
				t.Fatal(err)
			}

			// 测试路径 "a.b.X"
			vv, ok := r.Value("a.b.X")
			if !ok {
				t.Fatal("未找到路径 'a.b.X'")
			}
			vvv, err := vv.Int()
			if err != nil {
				t.Fatal(err)
			}
			if int64(1) != vvv {
				t.Fatalf("期望值为 1，但得到 %d", vvv)
			}

			// 测试路径 "a.b.Y"
			vv, ok = r.Value("a.b.Y")
			if !ok {
				t.Fatal("未找到路径 'a.b.Y'")
			}
			vvy, err := vv.String()
			if err != nil {
				t.Fatal(err)
			}
			if vvy != "lol" {
				t.Fatalf("期望值为 'lol'，但得到 %s", vvy)
			}

			// 测试路径 "a.b.z"
			vv, ok = r.Value("a.b.z")
			if !ok {
				t.Fatal("未找到路径 'a.b.z'")
			}
			vvz, err := vv.Bool()
			if err != nil {
				t.Fatal(err)
			}
			if !vvz {
				t.Fatal("期望值为 true，但得到 false")
			}
		})
	}
}

// 定义一个测试结构体，用于测试 Reader 类型的 Source 方法
func TestReader_Source(t *testing.T) {
	// 声明一个错误变量，用于捕获可能出现的错误
	var err error
	// 定义一个 options 结构体，用于配置 Reader 的行为
	opts := options{
		// 定义一个解码器函数，用于将 KeyValue 结构体中的 Value 字段解码为 map[string]interface{} 类型
		decoder: func(kv *KeyValue, v map[string]interface{}) error {
			// 尝试获取与 KeyValue 结构体中的 Format 字段对应的编解码器
			if codec := encoding.GetCodec(kv.Format); codec != nil {
				// 如果找到了对应的编解码器，则使用它来解码 Value 字段，并将结果存储在 v 中
				return codec.Unmarshal(kv.Value, &v)
			}
			// 如果没有找到对应的编解码器，则返回一个错误
			return fmt.Errorf("unsupported key: %s format: %s", kv.Key, kv.Format)
		},
		// 定义一个解析器函数，用于解析 KeyValue 结构体中的 Key 字段
		resolver: defaultResolver,
		// 定义一个合并函数，用于将两个 map[string]interface{} 类型的数据源合并
		merge: func(dst, src interface{}) error {
			// 使用 mergo 库的 Map 方法将 src 合并到 dst 中，并覆盖 dst 中已存在的键值对
			return mergo.Map(dst, src, mergo.WithOverride)
		},
	}
	// 创建一个新的 Reader 实例，使用之前定义的 options 结构体进行配置
	r := newReader(opts)
	// 调用 Reader 实例的 Merge 方法，将一个 KeyValue 结构体合并到 Reader 的数据源中
	err = r.Merge(&KeyValue{
		Key:    "b",
		Value:  []byte(`{"a": {"b": {"X": 1}}}`),
		Format: "json",
	})
	// 如果在合并过程中出现错误，则记录错误信息并终止测试
	if err != nil {
		t.Fatal(err)
	}
	// 调用 Reader 实例的 Source 方法，获取合并后的数据源
	b, err := r.Source()
	// 如果在获取数据源的过程中出现错误，则记录错误信息并终止测试
	if err != nil {
		t.Fatal(err)
	}
	// 使用 reflect.DeepEqual 函数比较获取到的数据源和预期的数据源是否相等
	if !reflect.DeepEqual([]byte(`{"a":{"b":{"X":1}}}`), b) {
		// 如果不相等，则记录错误信息并终止测试
		t.Fatal("[]byte(`{\"a\":{\"b\":{\"X\":1}}}`) is not equal to b")
	}
}

// TestCloneMap 测试 cloneMap 函数是否能正确复制 map
func TestCloneMap(t *testing.T) {
	tests := []struct {
		input map[string]interface{}
		want  map[string]interface{}
	}{
		// 测试用例 1：包含多种类型元素的 map
		{
			input: map[string]interface{}{
				"a": 1,
				"b": "2",
				"c": true,
			},
			want: map[string]interface{}{
				"a": 1,
				"b": "2",
				"c": true,
			},
		},
		// 测试用例 2：空 map
		{
			input: map[string]interface{}{},
			want:  map[string]interface{}{},
		},
		// 测试用例 3：输入为 nil 的情况
		{
			input: nil,
			want:  map[string]interface{}{},
		},
	}
	// 遍历测试用例
	for _, tt := range tests {
		// 调用 cloneMap 函数，并检查是否有错误
		if got, err := cloneMap(tt.input); err != nil {
			// 如果有错误，记录错误信息
			t.Errorf("expect no err, got %v", err)
		} else if !reflect.DeepEqual(got, tt.want) {
			// 如果没有错误，检查返回的 map 是否与预期一致
			t.Errorf("cloneMap(%v) = %v, want %v", tt.input, got, tt.want)
		}
	}
}

// TestReadValue 测试 readValue 函数
func TestReadValue(t *testing.T) {
	// 定义一个测试用的 map，包含了嵌套的 map
	m := map[string]interface{}{
		"a": 1,
		"b": map[string]interface{}{
			"c": "3",
			"d": map[string]interface{}{
				"e": true,
			},
		},
	}
	// 定义了三个 atomicValue 类型的变量，用于存储测试用例的期望值
	va := atomicValue{}
	va.Store(1)

	vbc := atomicValue{}
	vbc.Store("3")

	vbde := atomicValue{}
	vbde.Store(true)

	// 定义了一个测试用例的切片，包含了路径和期望值
	tests := []struct {
		path string
		want atomicValue
	}{
		// 测试路径 "a"，期望值为 va
		{
			path: "a",
			want: va,
		},
		// 测试路径 "b.c"，期望值为 vbc
		{
			path: "b.c",
			want: vbc,
		},
		// 测试路径 "b.d.e"，期望值为 vbde
		{
			path: "b.d.e",
			want: vbde,
		},
	}
	// 遍历测试用例
	for _, tt := range tests {
		// 调用 readValue 函数，传入测试用的 map 和路径，获取返回值和是否找到的标志
		if got, found := readValue(m, tt.path); !found {
			// 如果没有找到，记录错误信息
			t.Errorf("expect found %v in %v, but not.", tt.path, m)
		} else if got.Load() != tt.want.Load() {
			// 如果找到的值不等于期望值，记录错误信息
			t.Errorf("readValue(%v, %v) = %v, want %v", m, tt.path, got, tt.want)
		}
	}
}
