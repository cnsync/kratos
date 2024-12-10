package env

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/cnsync/kratos/config"
	"github.com/cnsync/kratos/config/file"
)

const _testJSON = `
{
    "test":{
        "server":{
			"name":"${SERVICE_NAME}",
            "addr":"${ADDR:127.0.0.1}",
            "port":"${PORT:8080}"
        }
    },
    "foo":[
        {
            "name":"Tom",
            "age":"${AGE}"
        }
    ]
}`

func TestEnvWithPrefix(t *testing.T) {
	// 定义临时目录路径
	path := filepath.Join(t.TempDir(), "test_config")
	// 定义配置文件名
	filename := filepath.Join(path, "test.json")
	// 定义配置文件内容
	data := []byte(_testJSON)

	// 测试结束后删除临时目录
	defer os.Remove(path)

	// 创建临时目录
	if err := os.MkdirAll(path, 0o700); err != nil {
		// 如果创建目录失败，记录错误信息
		t.Error(err)
	}
	// 将配置文件内容写入临时目录
	if err := os.WriteFile(filename, data, 0o666); err != nil {
		// 如果写入文件失败，记录错误信息
		t.Error(err)
	}

	// 设置环境变量前缀
	prefix1, prefix2 := "KRATOS_", "FOO"
	// 定义环境变量
	envs := map[string]string{
		prefix1 + "SERVICE_NAME": "kratos_app",
		prefix2 + "ADDR":         "192.168.0.1",
		prefix1 + "AGE":          "20",
		// 只有前缀
		prefix2:       "foo",
		prefix2 + "_": "foo_",
	}

	// 设置环境变量
	for k, v := range envs {
		os.Setenv(k, v)
	}

	// 创建一个新的配置对象，并添加文件源和环境变量源
	c := config.New(config.WithSource(
		file.NewSource(path),
		NewSource(prefix1, prefix2),
	))

	// 加载配置
	if err := c.Load(); err != nil {
		// 如果加载配置失败，记录错误信息
		t.Fatal(err)
	}

	// 定义测试用例
	tests := []struct {
		name   string
		path   string
		expect interface{}
	}{
		{
			name:   "test $KEY",
			path:   "test.server.name",
			expect: "kratos_app",
		},
		{
			name:   "test ${KEY:DEFAULT} without default",
			path:   "test.server.addr",
			expect: "192.168.0.1",
		},
		{
			name:   "test ${KEY:DEFAULT} with default",
			path:   "test.server.port",
			expect: "8080",
		},
		{
			name: "test ${KEY} in array",
			path: "foo",
			expect: []interface{}{
				map[string]interface{}{
					"name": "Tom",
					"age":  "20",
				},
			},
		},
	}

	// 遍历测试用例
	for _, test := range tests {
		// 运行测试用例
		t.Run(test.name, func(t *testing.T) {
			var err error
			// 获取配置值
			v := c.Value(test.path)
			// 加载配置值
			if v.Load() != nil {
				var actual interface{}
				// 根据期望类型转换配置值
				switch test.expect.(type) {
				case int:
					if actual, err = v.Int(); err == nil {
						// 如果实际值与期望值不相等，记录错误信息
						if !reflect.DeepEqual(test.expect.(int), int(actual.(int64))) {
							t.Errorf("expect %v, actual %v", test.expect, actual)
						}
					}
				case string:
					if actual, err = v.String(); err == nil {
						// 如果实际值与期望值不相等，记录错误信息
						if !reflect.DeepEqual(test.expect.(string), actual.(string)) {
							t.Errorf(`expect %v, actual %v`, test.expect, actual)
						}
					}
				case bool:
					if actual, err = v.Bool(); err == nil {
						// 如果实际值与期望值不相等，记录错误信息
						if !reflect.DeepEqual(test.expect.(bool), actual.(bool)) {
							t.Errorf(`expect %v, actual %v`, test.expect, actual)
						}
					}
				case float64:
					if actual, err = v.Float(); err == nil {
						// 如果实际值与期望值不相等，记录错误信息
						if !reflect.DeepEqual(test.expect.(float64), actual.(float64)) {
							t.Errorf(`expect %v, actual %v`, test.expect, actual)
						}
					}
				default:
					// 如果类型不匹配，记录错误信息
					actual = v.Load()
					if !reflect.DeepEqual(test.expect, actual) {
						t.Logf("\nexpect: %#v\nactural: %#v", test.expect, actual)
						t.Fail()
					}
				}
				// 如果发生错误，记录错误信息
				if err != nil {
					t.Error(err)
				}
			} else {
				// 如果配置值未找到，记录错误信息
				t.Error("value path not found")
			}
		})
	}
}

func TestEnvWithoutPrefix(t *testing.T) {
	// 定义临时目录路径
	path := filepath.Join(t.TempDir(), "test_config")
	// 定义配置文件名
	filename := filepath.Join(path, "test.json")
	// 定义配置文件内容
	data := []byte(_testJSON)

	// 测试结束后删除临时目录
	defer os.Remove(path)

	// 创建临时目录
	if err := os.MkdirAll(path, 0o700); err != nil {
		// 如果创建目录失败，记录错误信息
		t.Error(err)
	}
	// 将配置文件内容写入临时目录
	if err := os.WriteFile(filename, data, 0o666); err != nil {
		// 如果写入文件失败，记录错误信息
		t.Error(err)
	}

	// 设置环境变量
	envs := map[string]string{
		"SERVICE_NAME": "kratos_app",
		"ADDR":         "192.168.0.1",
		"AGE":          "20",
	}

	// 设置环境变量
	for k, v := range envs {
		os.Setenv(k, v)
	}

	// 创建一个新的配置对象，并添加文件源和环境变量源
	c := config.New(config.WithSource(
		NewSource(),
		file.NewSource(path),
	))

	// 加载配置
	if err := c.Load(); err != nil {
		// 如果加载配置失败，记录错误信息
		t.Fatal(err)
	}

	// 定义测试用例
	tests := []struct {
		name   string
		path   string
		expect interface{}
	}{
		{
			name:   "test $KEY",
			path:   "test.server.name",
			expect: "kratos_app",
		},
		{
			name:   "test ${KEY:DEFAULT} without default",
			path:   "test.server.addr",
			expect: "192.168.0.1",
		},
		{
			name:   "test ${KEY:DEFAULT} with default",
			path:   "test.server.port",
			expect: "8080",
		},
		{
			name: "test ${KEY} in array",
			path: "foo",
			expect: []interface{}{
				map[string]interface{}{
					"name": "Tom",
					"age":  "20",
				},
			},
		},
	}

	// 遍历测试用例
	for _, test := range tests {
		// 运行测试用例
		t.Run(test.name, func(t *testing.T) {
			var err error
			// 获取配置值
			v := c.Value(test.path)
			// 加载配置值
			if v.Load() != nil {
				var actual interface{}
				// 根据期望类型转换配置值
				switch test.expect.(type) {
				case int:
					if actual, err = v.Int(); err == nil {
						// 如果实际值与期望值不相等，记录错误信息
						if !reflect.DeepEqual(test.expect.(int), int(actual.(int64))) {
							t.Errorf("expect %v, actual %v", test.expect, actual)
						}
					}
				case string:
					if actual, err = v.String(); err == nil {
						// 如果实际值与期望值不相等，记录错误信息
						if !reflect.DeepEqual(test.expect.(string), actual.(string)) {
							t.Errorf(`expect %v, actual %v`, test.expect, actual)
						}
					}
				case bool:
					if actual, err = v.Bool(); err == nil {
						// 如果实际值与期望值不相等，记录错误信息
						if !reflect.DeepEqual(test.expect.(bool), actual.(bool)) {
							t.Errorf(`expect %v, actual %v`, test.expect, actual)
						}
					}
				case float64:
					if actual, err = v.Float(); err == nil {
						// 如果实际值与期望值不相等，记录错误信息
						if !reflect.DeepEqual(test.expect.(float64), actual.(float64)) {
							t.Errorf(`expect %v, actual %v`, test.expect, actual)
						}
					}
				default:
					// 如果类型不匹配，记录错误信息
					actual = v.Load()
					if !reflect.DeepEqual(test.expect, actual) {
						t.Logf("\nexpect: %#v\nactural: %#v", test.expect, actual)
						t.Fail()
					}
				}
				// 如果发生错误，记录错误信息
				if err != nil {
					t.Error(err)
				}
			} else {
				// 如果配置值未找到，记录错误信息
				t.Error("value path not found")
			}
		})
	}
}

func Test_env_load(t *testing.T) {
	// 定义 fields 结构体，包含 prefixes 字段
	type fields struct {
		prefixes []string
	}
	// 定义 args 结构体，包含 envStrings 字段
	type args struct {
		envStrings []string
	}
	// 定义测试用例结构体
	tests := []struct {
		name   string
		fields fields
		args   args
		want   []*config.KeyValue
	}{
		// 测试用例 1：没有前缀
		{
			name: "without prefixes",
			fields: fields{
				prefixes: nil,
			},
			args: args{
				envStrings: []string{
					"SERVICE_NAME=kratos_app",
					"ADDR=192.168.0.1",
					"AGE=20",
				},
			},
			want: []*config.KeyValue{
				{Key: "SERVICE_NAME", Value: []byte("kratos_app"), Format: ""},
				{Key: "ADDR", Value: []byte("192.168.0.1"), Format: ""},
				{Key: "AGE", Value: []byte("20"), Format: ""},
			},
		},
		// 测试用例 2：空前缀
		{
			name: "empty prefix",
			fields: fields{
				prefixes: []string{""},
			},
			args: args{
				envStrings: []string{
					"__SERVICE_NAME=kratos_app",
					"__ADDR=192.168.0.1",
					"__AGE=20",
				},
			},
			want: []*config.KeyValue{
				{Key: "_SERVICE_NAME", Value: []byte("kratos_app"), Format: ""},
				{Key: "_ADDR", Value: []byte("192.168.0.1"), Format: ""},
				{Key: "_AGE", Value: []byte("20"), Format: ""},
			},
		},
		// 测试用例 3：下划线前缀
		{
			name: "underscore prefix",
			fields: fields{
				prefixes: []string{"_"},
			},
			args: args{
				envStrings: []string{
					"__SERVICE_NAME=kratos_app",
					"__ADDR=192.168.0.1",
					"__AGE=20",
				},
			},
			want: []*config.KeyValue{
				{Key: "SERVICE_NAME", Value: []byte("kratos_app"), Format: ""},
				{Key: "ADDR", Value: []byte("192.168.0.1"), Format: ""},
				{Key: "AGE", Value: []byte("20"), Format: ""},
			},
		},
		// 测试用例 4：带有前缀
		{
			name: "with prefixes",
			fields: fields{
				prefixes: []string{"KRATOS_", "FOO"},
			},
			args: args{
				envStrings: []string{
					"KRATOS_SERVICE_NAME=kratos_app",
					"KRATOS_ADDR=192.168.0.1",
					"FOO_AGE=20",
				},
			},
			want: []*config.KeyValue{
				{Key: "SERVICE_NAME", Value: []byte("kratos_app"), Format: ""},
				{Key: "ADDR", Value: []byte("192.168.0.1"), Format: ""},
				{Key: "AGE", Value: []byte("20"), Format: ""},
			},
		},
		// 测试用例 5：不应该 panic #1
		{
			name: "should not panic #1",
			fields: fields{
				prefixes: []string{"FOO"},
			},
			args: args{
				envStrings: []string{
					"FOO=123",
				},
			},
			want: nil,
		},
		// 测试用例 6：不应该 panic #2
		{
			name: "should not panic #2",
			fields: fields{
				prefixes: []string{"FOO=1"},
			},
			args: args{
				envStrings: []string{
					"FOO=123",
				},
			},
			want: nil,
		},
	}
	// 遍历测试用例
	for _, tt := range tests {
		// 运行测试用例
		t.Run(tt.name, func(t *testing.T) {
			// 创建 env 实例
			e := &env{
				prefixes: tt.fields.prefixes,
			}
			// 调用 load 方法
			got := e.load(tt.args.envStrings)
			// 比较实际结果和预期结果
			if !reflect.DeepEqual(tt.want, got) {
				// 如果不相等，记录错误信息
				t.Errorf("env.load() = %v, want %v", got, tt.want)
			}
		})
	}
}

// Test_matchPrefix 测试 matchPrefix 函数是否能够正确匹配前缀
func Test_matchPrefix(t *testing.T) {
	// 定义 args 结构体，包含 prefixes 字段和 s 字段
	type args struct {
		prefixes []string
		s        string
	}
	// 定义测试用例结构体
	tests := []struct {
		name   string
		args   args
		want   string
		wantOk bool
	}{
		// 测试用例 1：没有前缀，应该返回空字符串和 false
		{args: args{prefixes: nil, s: "foo=123"}, want: "", wantOk: false},
		// 测试用例 2：前缀为空字符串，应该返回空字符串和 true
		{args: args{prefixes: []string{""}, s: "foo=123"}, want: "", wantOk: true},
		// 测试用例 3：前缀为 foo，应该返回 foo 和 true
		{args: args{prefixes: []string{"foo"}, s: "foo=123"}, want: "foo", wantOk: true},
		// 测试用例 4：前缀为 foo=1，应该返回 foo=1 和 true
		{args: args{prefixes: []string{"foo=1"}, s: "foo=123"}, want: "foo=1", wantOk: true},
		// 测试用例 5：前缀为 foo=1234，应该返回空字符串和 false
		{args: args{prefixes: []string{"foo=1234"}, s: "foo=123"}, want: "", wantOk: false},
		// 测试用例 6：前缀为 bar，应该返回空字符串和 false
		{args: args{prefixes: []string{"bar"}, s: "foo=123"}, want: "", wantOk: false},
	}
	// 遍历测试用例
	for _, tt := range tests {
		// 运行测试用例
		t.Run(tt.name, func(t *testing.T) {
			// 调用 matchPrefix 函数
			got, gotOk := matchPrefix(tt.args.prefixes, tt.args.s)
			// 比较实际结果和预期结果
			if got != tt.want {
				// 如果不相等，记录错误信息
				t.Errorf("matchPrefix() got = %v, want %v", got, tt.want)
			}
			// 比较实际结果和预期结果
			if gotOk != tt.wantOk {
				// 如果不相等，记录错误信息
				t.Errorf("matchPrefix() gotOk = %v, wantOk %v", gotOk, tt.wantOk)
			}
		})
	}
}

// Test_env_watch 测试 env 结构体的 watch 方法
func Test_env_watch(t *testing.T) {
	// 定义前缀列表
	prefixes := []string{"BAR", "FOO"}
	// 创建一个新的 env 源，并指定前缀
	source := NewSource(prefixes...)
	// 调用 watch 方法，并检查是否有错误
	w, err := source.Watch()
	// 如果有错误，记录错误信息
	if err != nil {
		t.Errorf("expect no err, got %v", err)
	}
	// 调用 Stop 方法停止监控
	_ = w.Stop()
}
