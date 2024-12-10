package file

import (
	"errors"
	"os"
	"path/filepath"
	"reflect"
	"sync"
	"testing"
	"time"

	"github.com/cnsync/kratos/config"
)

const (
	_testJSON = `
{
    "test":{
        "settings":{
            "int_key":1000,
            "float_key":1000.1,
            "duration_key":10000,
            "string_key":"string_value"
        },
        "server":{
            "addr":"127.0.0.1",
            "port":8000
        }
    },
    "foo":[
        {
            "name":"nihao",
            "age":18
        },
        {
            "name":"nihao",
            "age":18
        }
    ]
}`

	_testJSONUpdate = `
{
    "test":{
        "settings":{
            "int_key":1000,
            "float_key":1000.1,
            "duration_key":10000,
            "string_key":"string_value"
        },
        "server":{
            "addr":"127.0.0.1",
            "port":8000
        }
    },
    "foo":[
        {
            "name":"nihao",
            "age":18
        },
        {
            "name":"nihao",
            "age":18
        }
    ],
	"bar":{
		"event":"update"
	}
}`

	//	_testYaml = `
	//Foo:
	//    bar :
	//        - {name: nihao,age: 1}
	//        - {name: nihao,age: 1}
	//
	//
	//`
)

// TestFile 测试文件源的加载和监视功能
func TestFile(t *testing.T) {
	// 定义一个临时目录路径
	var (
		path = filepath.Join(t.TempDir(), "test_config")
		// 定义一个文件路径
		file = filepath.Join(path, "test.json")
		// 定义一个 JSON 数据
		data = []byte(_testJSON)
	)
	// 测试结束后删除临时目录
	defer os.Remove(path)
	// 创建临时目录
	if err := os.MkdirAll(path, 0o700); err != nil {
		// 如果创建目录失败，记录错误信息
		t.Error(err)
	}
	// 将 JSON 数据写入文件
	if err := os.WriteFile(file, data, 0o666); err != nil {
		// 如果写入文件失败，记录错误信息
		t.Error(err)
	}
	// 测试文件源的加载功能
	testSource(t, file, data)
	// 测试目录源的加载功能
	testSource(t, path, data)
	// 测试文件的监视功能
	testWatchFile(t, file)
	// 测试目录的监视功能
	testWatchDir(t, path, file)
}

// testWatchFile 测试文件的监视功能
func testWatchFile(t *testing.T, path string) {
	// 记录测试文件路径
	t.Log(path)

	// 创建一个文件源实例
	s := NewSource(path)
	// 创建一个监视实例
	watch, err := s.Watch()
	// 如果创建监视实例失败，记录错误信息
	if err != nil {
		t.Error(err)
	}

	// 打开文件
	f, err := os.OpenFile(path, os.O_RDWR, 0)
	// 如果打开文件失败，记录错误信息
	if err != nil {
		t.Error(err)
	}
	// 测试结束后关闭文件
	defer f.Close()
	// 将更新后的 JSON 数据写入文件
	_, err = f.WriteString(_testJSONUpdate)
	// 如果写入文件失败，记录错误信息
	if err != nil {
		t.Error(err)
	}
	// 等待监视实例获取到更新后的配置
	kvs, err := watch.Next()
	// 如果获取配置失败，记录错误信息
	if err != nil {
		t.Errorf("watch.Next() error(%v)", err)
	}
	// 检查获取到的配置是否与更新后的 JSON 数据一致
	if !reflect.DeepEqual(string(kvs[0].Value), _testJSONUpdate) {
		// 如果不一致，记录错误信息
		t.Errorf("string(kvs[0].Value(%v) is  not equal to _testJSONUpdate(%v)", kvs[0].Value, _testJSONUpdate)
	}

	// 定义一个新的文件路径
	newFilepath := filepath.Join(filepath.Dir(path), "test1.json")
	// 将测试文件重命名为新的文件路径
	if err = os.Rename(path, newFilepath); err != nil {
		// 如果重命名文件失败，记录错误信息
		t.Error(err)
	}
	// 等待监视实例获取到重命名事件
	kvs, err = watch.Next()
	// 如果获取配置失败，记录错误信息
	if err == nil {
		// 如果没有错误，记录错误信息
		t.Errorf("watch.Next() error(%v)", err)
	}
	// 检查获取到的配置是否为 nil
	if kvs != nil {
		// 如果不为 nil，记录错误信息
		t.Errorf("watch.Next() error(%v)", err)
	}

	// 停止监视实例
	err = watch.Stop()
	// 如果停止监视实例失败，记录错误信息
	if err != nil {
		t.Errorf("watch.Stop() error(%v)", err)
	}

	// 将新的文件路径重命名为原来的文件路径
	if err := os.Rename(newFilepath, path); err != nil {
		// 如果重命名文件失败，记录错误信息
		t.Error(err)
	}
}

// testWatchDir 测试目录的监视功能
func testWatchDir(t *testing.T, path, file string) {
	// 记录测试目录路径
	t.Log(path)
	// 记录测试文件路径
	t.Log(file)

	// 创建一个目录源实例
	s := NewSource(path)
	// 创建一个监视实例
	watch, err := s.Watch()
	// 如果创建监视实例失败，记录错误信息
	if err != nil {
		t.Error(err)
	}

	// 打开文件
	f, err := os.OpenFile(file, os.O_RDWR, 0)
	// 如果打开文件失败，记录错误信息
	if err != nil {
		t.Error(err)
	}
	// 测试结束后关闭文件
	defer f.Close()
	// 将更新后的 JSON 数据写入文件
	_, err = f.WriteString(_testJSONUpdate)
	// 如果写入文件失败，记录错误信息
	if err != nil {
		t.Error(err)
	}

	// 等待监视实例获取到更新后的配置
	kvs, err := watch.Next()
	// 如果获取配置失败，记录错误信息
	if err != nil {
		t.Errorf("watch.Next() error(%v)", err)
	}
	// 检查获取到的配置是否与更新后的 JSON 数据一致
	if !reflect.DeepEqual(string(kvs[0].Value), _testJSONUpdate) {
		// 如果不一致，记录错误信息
		t.Errorf("string(kvs[0].Value(%s) is  not equal to _testJSONUpdate(%v)", kvs[0].Value, _testJSONUpdate)
	}
}

// testSource 测试文件源的加载功能
func testSource(t *testing.T, path string, data []byte) {
	// 记录测试文件路径
	t.Log(path)

	// 创建一个文件源实例
	s := NewSource(path)
	// 加载文件源的配置
	kvs, err := s.Load()
	// 如果加载配置失败，记录错误信息
	if err != nil {
		t.Error(err)
	}
	// 检查加载的配置是否与预期数据一致
	if string(kvs[0].Value) != string(data) {
		// 如果不一致，记录错误信息
		t.Errorf("no expected: %s, but got: %s", kvs[0].Value, data)
	}
}

func TestConfig(t *testing.T) {
	// 获取临时目录并创建一个名为 test_config.json 的文件路径
	path := filepath.Join(t.TempDir(), "test_config.json")
	// 测试结束后删除该文件
	defer os.Remove(path)
	// 将测试用的 JSON 数据写入文件
	if err := os.WriteFile(path, []byte(_testJSON), 0o666); err != nil {
		// 如果写入文件失败，记录错误信息
		t.Error(err)
	}
	// 创建一个新的配置对象，并添加一个文件源
	c := config.New(config.WithSource(
		NewSource(path),
	))
	// 测试扫描功能
	testScan(t, c)
	// 测试配置加载和解析功能
	testConfig(t, c)
}

func testConfig(t *testing.T, c config.Config) {
	// 定义一个预期的配置值映射
	expected := map[string]interface{}{
		"test.settings.int_key":      int64(1000),
		"test.settings.float_key":    1000.1,
		"test.settings.string_key":   "string_value",
		"test.settings.duration_key": time.Duration(10000),
		"test.server.addr":           "127.0.0.1",
		"test.server.port":           int64(8000),
	}
	// 加载配置
	if err := c.Load(); err != nil {
		// 如果加载配置失败，记录错误信息
		t.Error(err)
	}
	// 遍历预期的配置值映射，检查实际加载的配置值是否与预期一致
	for key, value := range expected {
		// 根据值的类型进行不同的处理
		switch value.(type) {
		case int64:
			// 尝试获取整数类型的值，并检查是否与预期一致
			if v, err := c.Value(key).Int(); err != nil {
				// 如果获取值失败，记录错误信息
				t.Error(key, value, err)
			} else if v != value {
				// 如果获取的值与预期不一致，记录错误信息
				t.Errorf("no expect key: %s value: %v, but got: %v", key, value, v)
			}
		case float64:
			// 尝试获取浮点类型的值，并检查是否与预期一致
			if v, err := c.Value(key).Float(); err != nil {
				// 如果获取值失败，记录错误信息
				t.Error(key, value, err)
			} else if v != value {
				// 如果获取的值与预期不一致，记录错误信息
				t.Errorf("no expect key: %s value: %v, but got: %v", key, value, v)
			}
		case string:
			// 尝试获取字符串类型的值，并检查是否与预期一致
			if v, err := c.Value(key).String(); err != nil {
				// 如果获取值失败，记录错误信息
				t.Error(key, value, err)
			} else if v != value {
				// 如果获取的值与预期不一致，记录错误信息
				t.Errorf("no expect key: %s value: %v, but got: %v", key, value, v)
			}
		case time.Duration:
			// 尝试获取时间类型的值，并检查是否与预期一致
			if v, err := c.Value(key).Duration(); err != nil {
				// 如果获取值失败，记录错误信息
				t.Error(key, value, err)
			} else if v != value {
				// 如果获取的值与预期不一致，记录错误信息
				t.Errorf("no expect key: %s value: %v, but got: %v", key, value, v)
			}
		}
	}
	// 测试结构体扫描功能
	var settings struct {
		IntKey      int64         `json:"int_key"`
		FloatKey    float64       `json:"float_key"`
		StringKey   string        `json:"string_key"`
		DurationKey time.Duration `json:"duration_key"`
	}
	// 尝试将配置值扫描到结构体中，并检查是否成功
	if err := c.Value("test.settings").Scan(&settings); err != nil {
		// 如果扫描失败，记录错误信息
		t.Error(err)
	}
	// 检查结构体中的值是否与预期一致
	if v := expected["test.settings.int_key"]; settings.IntKey != v {
		// 如果不一致，记录错误信息
		t.Errorf("no expect int_key value: %v, but got: %v", settings.IntKey, v)
	}
	if v := expected["test.settings.float_key"]; settings.FloatKey != v {
		// 如果不一致，记录错误信息
		t.Errorf("no expect float_key value: %v, but got: %v", settings.FloatKey, v)
	}
	if v := expected["test.settings.string_key"]; settings.StringKey != v {
		// 如果不一致，记录错误信息
		t.Errorf("no expect string_key value: %v, but got: %v", settings.StringKey, v)
	}
	if v := expected["test.settings.duration_key"]; settings.DurationKey != v {
		// 如果不一致，记录错误信息
		t.Errorf("no expect duration_key value: %v, but got: %v", settings.DurationKey, v)
	}
	// 测试未找到键的情况
	if _, err := c.Value("not_found_key").Bool(); errors.Is(err, config.ErrNotFound) {
		// 如果未找到键，记录日志信息
		t.Logf("not_found_key not match: %v", err)
	}
}

// testScan 测试配置文件的扫描功能
func testScan(t *testing.T, c config.Config) {
	// 定义一个结构体，用于接收扫描后的配置数据
	type TestJSON struct {
		Test struct {
			Settings struct {
				IntKey      int     `json:"int_key"`
				FloatKey    float64 `json:"float_key"`
				DurationKey int     `json:"duration_key"`
				StringKey   string  `json:"string_key"`
			} `json:"settings"`
			Server struct {
				Addr string `json:"addr"`
				Port int    `json:"port"`
			} `json:"server"`
		} `json:"test"`
		Foo []struct {
			Name string `json:"name"`
			Age  int    `json:"age"`
		} `json:"foo"`
	}
	// 初始化结构体变量
	var conf TestJSON
	// 加载配置
	if err := c.Load(); err != nil {
		// 如果加载配置失败，记录错误信息
		t.Error(err)
	}
	// 扫描配置到结构体中
	if err := c.Scan(&conf); err != nil {
		// 如果扫描失败，记录错误信息
		t.Error(err)
	}
	// 记录扫描后的配置信息
	t.Log(conf)
}

// TestMergeDataRace 测试在多线程环境下配置加载和扫描的并发性能
func TestMergeDataRace(t *testing.T) {
	// 获取临时目录并创建一个名为 test_config.json 的文件路径
	path := filepath.Join(t.TempDir(), "test_config.json")
	// 测试结束后删除该文件
	defer os.Remove(path)
	// 将测试用的 JSON 数据写入文件
	if err := os.WriteFile(path, []byte(_testJSON), 0o666); err != nil {
		// 如果写入文件失败，记录错误信息
		t.Error(err)
	}
	// 创建一个新的配置对象，并添加一个文件源
	c := config.New(config.WithSource(
		NewSource(path),
	))
	// 定义并发线程数量
	const count = 80
	// 创建一个等待组，用于等待所有线程结束
	wg := &sync.WaitGroup{}
	// 初始化等待组计数器
	wg.Add(2)
	// 创建一个通道，用于同步线程启动
	startCh := make(chan struct{})
	// 启动第一个线程，用于扫描配置
	go func() {
		defer wg.Done()
		// 等待通道信号，开始执行
		<-startCh
		// 循环扫描配置
		for i := 0; i < count; i++ {
			var conf struct{}
			// 尝试将配置值扫描到结构体中，并检查是否成功
			if err := c.Scan(&conf); err != nil {
				// 如果扫描失败，记录错误信息
				t.Error(err)
			}
		}
	}()

	// 启动第二个线程，用于加载配置
	go func() {
		defer wg.Done()
		// 等待通道信号，开始执行
		<-startCh
		// 循环加载配置
		for i := 0; i < count; i++ {
			// 加载配置
			if err := c.Load(); err != nil {
				// 如果加载配置失败，记录错误信息
				t.Error(err)
			}
		}
	}()
	// 关闭通道，通知所有线程开始执行
	close(startCh)
	// 等待所有线程结束
	wg.Wait()
}
