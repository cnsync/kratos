package config

import (
	"errors"
	"testing"

	"dario.cat/mergo"
)

const (
	// _testJSON 定义了一个用于测试的 JSON 字符串
	_testJSON = `
{
    "server":{
        "http":{
            "addr":"0.0.0.0",
			"port":80,
            "timeout":0.5,
			"enable_ssl":true
        },
        "grpc":{
            "addr":"0.0.0.0",
			"port":10080,
            "timeout":0.2
        }
    },
    "data":{
        "database":{
            "driver":"mysql",
            "source":"root:root@tcp(127.0.0.1:3306)/karta_id?parseTime=true"
        }
    },
	"endpoints":[
		"www.aaa.com",
		"www.bbb.org"
	]
}`
)

// testConfigStruct 定义了一个用于测试的配置结构体
type testConfigStruct struct {
	Server struct {
		HTTP struct {
			Addr      string  `json:"addr"`
			Port      int     `json:"port"`
			Timeout   float64 `json:"timeout"`
			EnableSSL bool    `json:"enable_ssl"`
		} `json:"http"`
		GRPC struct {
			Addr    string  `json:"addr"`
			Port    int     `json:"port"`
			Timeout float64 `json:"timeout"`
		} `json:"grpc"`
	} `json:"server"`
	Data struct {
		Database struct {
			Driver string `json:"driver"`
			Source string `json:"source"`
		} `json:"database"`
	} `json:"data"`
	Endpoints []string `json:"endpoints"`
}

// testJSONSource 定义了一个用于测试的 JSON 数据源
type testJSONSource struct {
	data string
	sig  chan struct{}
	err  chan struct{}
}

// newTestJSONSource 创建一个新的 testJSONSource 实例
func newTestJSONSource(data string) *testJSONSource {
	return &testJSONSource{data: data, sig: make(chan struct{}), err: make(chan struct{})}
}

// Load 方法从 testJSONSource 加载数据
func (p *testJSONSource) Load() ([]*KeyValue, error) {
	kv := &KeyValue{
		Key:    "json",
		Value:  []byte(p.data),
		Format: "json",
	}
	return []*KeyValue{kv}, nil
}

// Watch 方法监视 testJSONSource 的变化
func (p *testJSONSource) Watch() (Watcher, error) {
	return newTestWatcher(p.sig, p.err), nil
}

// testWatcher 定义了一个用于测试的监听器
type testWatcher struct {
	sig  chan struct{}
	err  chan struct{}
	exit chan struct{}
}

// newTestWatcher 创建一个新的 testWatcher 实例
func newTestWatcher(sig, err chan struct{}) Watcher {
	return &testWatcher{sig: sig, err: err, exit: make(chan struct{})}
}

// Next 方法等待下一个事件
func (w *testWatcher) Next() ([]*KeyValue, error) {
	select {
	case <-w.sig:
		return nil, nil
	case <-w.err:
		return nil, errors.New("error")
	case <-w.exit:
		return nil, nil
	}
}

// Stop 方法停止监听器
func (w *testWatcher) Stop() error {
	close(w.exit)
	return nil
}

// TestConfig 是一个单元测试，用于测试配置管理的功能
func TestConfig(t *testing.T) {
	var (
		err            error
		httpAddr       = "0.0.0.0"
		httpTimeout    = 0.5
		grpcPort       = 10080
		endpoint1      = "www.aaa.com"
		databaseDriver = "mysql"
	)

	// 创建一个新的配置实例
	c := New(
		WithSource(newTestJSONSource(_testJSON)),
		WithDecoder(defaultDecoder),
		WithResolver(defaultResolver),
	)
	// 关闭配置实例
	err = c.Close()
	if err != nil {
		t.Fatal(err)
	}

	// 创建一个新的 testJSONSource 实例
	jSource := newTestJSONSource(_testJSON)
	// 创建一个新的配置实例
	opts := options{
		sources:  []Source{jSource},
		decoder:  defaultDecoder,
		resolver: defaultResolver,
		merge: func(dst, src interface{}) error {
			return mergo.Map(dst, src, mergo.WithOverride)
		},
	}
	cf := &config{}
	cf.opts = opts
	cf.reader = newReader(opts)

	// 加载配置
	err = cf.Load()
	if err != nil {
		t.Fatal(err)
	}

	// 获取数据库驱动
	driver, err := cf.Value("data.database.driver").String()
	if err != nil {
		t.Fatal(err)
	}
	// 检查数据库驱动是否正确
	if databaseDriver != driver {
		t.Fatal("databaseDriver is not equal to val")
	}

	// 监视 endpoints 变化
	err = cf.Watch("endpoints", func(string, Value) {})
	if err != nil {
		t.Fatal(err)
	}

	// 发送信号
	jSource.sig <- struct{}{}
	// 发送错误信号
	jSource.err <- struct{}{}

	// 定义一个 testConfigStruct 实例
	var testConf testConfigStruct
	// 扫描配置到 testConfigStruct 实例
	err = cf.Scan(&testConf)
	if err != nil {
		t.Fatal(err)
	}
	// 检查 HTTP 地址是否正确
	if httpAddr != testConf.Server.HTTP.Addr {
		t.Errorf("testConf.Server.HTTP.Addr want: %s, got: %s", httpAddr, testConf.Server.HTTP.Addr)
	}
	// 检查 HTTP 超时时间是否正确
	if httpTimeout != testConf.Server.HTTP.Timeout {
		t.Errorf("testConf.Server.HTTP.Timeout want: %.1f, got: %.1f", httpTimeout, testConf.Server.HTTP.Timeout)
	}
	// 检查 HTTP 是否启用 SSL 是否正确
	if !testConf.Server.HTTP.EnableSSL {
		t.Error("testConf.Server.HTTP.EnableSSL is not equal to true")
	}
	// 检查 GRPC 端口是否正确
	if grpcPort != testConf.Server.GRPC.Port {
		t.Errorf("testConf.Server.GRPC.Port want: %d, got: %d", grpcPort, testConf.Server.GRPC.Port)
	}
	// 检查第一个端点是否正确
	if endpoint1 != testConf.Endpoints[0] {
		t.Errorf("testConf.Endpoints[0] want: %s, got: %s", endpoint1, testConf.Endpoints[0])
	}
	// 检查端点数量是否正确
	if len(testConf.Endpoints) != 2 {
		t.Error("len(testConf.Endpoints) is not equal to 2")
	}
}
