package http

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"reflect"
	"strconv"
	"testing"
	"time"

	kratoserrors "github.com/cnsync/kratos/errors"
	"github.com/cnsync/kratos/middleware"
	"github.com/cnsync/kratos/registry"
	"github.com/cnsync/kratos/selector"
)

// mockRoundTripper 是一个模拟的 RoundTripper 实现
type mockRoundTripper struct{}

// RoundTrip 方法实现了 http.RoundTripper 接口
func (rt *mockRoundTripper) RoundTrip(_ *http.Request) (resp *http.Response, err error) {
	return
}

// mockCallOption 是一个模拟的调用选项
type mockCallOption struct {
	needErr bool
}

// before 方法在调用前执行
func (x *mockCallOption) before(_ *callInfo) error {
	if x.needErr {
		return errors.New("option need return err")
	}
	return nil
}

// after 方法在调用后执行
func (x *mockCallOption) after(_ *callInfo, _ *csAttempt) {
	log.Println("run in mockCallOption.after")
}

// TestWithSubset 测试 WithSubset 选项
func TestWithSubset(t *testing.T) {
	co := &clientOptions{}
	o := WithSubset(1)
	o(co)
	if co.subsetSize != 1 {
		t.Error("expected subset size to be 1")
	}
}

// TestWithTransport 测试 WithTransport 选项
func TestWithTransport(t *testing.T) {
	ov := &mockRoundTripper{}
	o := WithTransport(ov)
	co := &clientOptions{}
	o(co)
	if !reflect.DeepEqual(co.transport, ov) {
		t.Errorf("expected transport to be %v, got %v", ov, co.transport)
	}
}

// TestWithTimeout 测试 WithTimeout 选项
func TestWithTimeout(t *testing.T) {
	ov := 1 * time.Second
	o := WithTimeout(ov)
	co := &clientOptions{}
	o(co)
	if !reflect.DeepEqual(co.timeout, ov) {
		t.Errorf("expected timeout to be %v, got %v", ov, co.timeout)
	}
}

// TestWithBlock 测试 WithBlock 选项
func TestWithBlock(t *testing.T) {
	o := WithBlock()
	co := &clientOptions{}
	o(co)
	if !co.block {
		t.Errorf("expected block to be true, got %v", co.block)
	}
}

// TestWithTLSConfig 测试 WithTLSConfig 选项
func TestWithTLSConfig(t *testing.T) {
	ov := &tls.Config{}
	o := WithTLSConfig(ov)
	co := &clientOptions{}
	o(co)
	if !reflect.DeepEqual(co.tlsConf, ov) {
		t.Errorf("expected tls config to be %v, got %v", ov, co.tlsConf)
	}
}

// TestWithUserAgent 测试 WithUserAgent 选项
func TestWithUserAgent(t *testing.T) {
	ov := "kratos"
	o := WithUserAgent(ov)
	co := &clientOptions{}
	o(co)
	if !reflect.DeepEqual(co.userAgent, ov) {
		t.Errorf("expected user agent to be %v, got %v", ov, co.userAgent)
	}
}

// TestWithMiddleware 测试 WithMiddleware 选项
func TestWithMiddleware(t *testing.T) {
	o := &clientOptions{}
	v := []middleware.Middleware{
		func(middleware.Handler) middleware.Handler { return nil },
	}
	WithMiddleware(v...)(o)
	if !reflect.DeepEqual(o.middleware, v) {
		t.Errorf("expected middleware to be %v, got %v", v, o.middleware)
	}
}

// TestWithEndpoint 测试 WithEndpoint 选项
func TestWithEndpoint(t *testing.T) {
	ov := "some-endpoint"
	o := WithEndpoint(ov)
	co := &clientOptions{}
	o(co)
	if !reflect.DeepEqual(co.endpoint, ov) {
		t.Errorf("expected endpoint to be %v, got %v", ov, co.endpoint)
	}
}

// TestWithRequestEncoder 测试 WithRequestEncoder 选项
func TestWithRequestEncoder(t *testing.T) {
	o := &clientOptions{}
	v := func(context.Context, string, interface{}) (body []byte, err error) {
		return nil, nil
	}
	WithRequestEncoder(v)(o)
	if o.encoder == nil {
		t.Errorf("expected encoder to be not nil")
	}
}

// TestWithResponseDecoder 测试 WithResponseDecoder 选项
func TestWithResponseDecoder(t *testing.T) {
	o := &clientOptions{}
	v := func(context.Context, *http.Response, interface{}) error { return nil }
	WithResponseDecoder(v)(o)
	if o.decoder == nil {
		t.Errorf("expected encoder to be not nil")
	}
}

// TestWithErrorDecoder 测试 WithErrorDecoder 选项
func TestWithErrorDecoder(t *testing.T) {
	o := &clientOptions{}
	v := func(context.Context, *http.Response) error { return nil }
	WithErrorDecoder(v)(o)
	if o.errorDecoder == nil {
		t.Errorf("expected encoder to be not nil")
	}
}

// mockDiscovery 是一个模拟的服务发现实现
type mockDiscovery struct{}

// GetService 方法实现了 registry.Discovery 接口
func (*mockDiscovery) GetService(_ context.Context, _ string) ([]*registry.ServiceInstance, error) {
	return nil, nil
}

// Watch 方法实现了 registry.Discovery 接口
func (*mockDiscovery) Watch(_ context.Context, _ string) (registry.Watcher, error) {
	return &mockWatcher{}, nil
}

// mockWatcher 是一个模拟的服务实例监听器
type mockWatcher struct{}

// Next 方法实现了 registry.Watcher 接口
func (m *mockWatcher) Next() ([]*registry.ServiceInstance, error) {
	instance := &registry.ServiceInstance{
		ID:        "1",
		Name:      "kratos",
		Version:   "v1",
		Metadata:  map[string]string{},
		Endpoints: []string{fmt.Sprintf("http://127.0.0.1:9001?isSecure=%s", strconv.FormatBool(false))},
	}
	time.Sleep(time.Millisecond * 500)
	return []*registry.ServiceInstance{instance}, nil
}

// Stop 方法实现了 registry.Watcher 接口
func (*mockWatcher) Stop() error {
	return nil
}

// TestWithDiscovery 测试 WithDiscovery 选项
func TestWithDiscovery(t *testing.T) {
	ov := &mockDiscovery{}
	o := WithDiscovery(ov)
	co := &clientOptions{}
	o(co)
	if !reflect.DeepEqual(co.discovery, ov) {
		t.Errorf("expected discovery to be %v, got %v", ov, co.discovery)
	}
}

// TestWithNodeFilter 测试 WithNodeFilter 选项
func TestWithNodeFilter(t *testing.T) {
	ov := func(context.Context, []selector.Node) []selector.Node {
		return []selector.Node{&selector.DefaultNode{}}
	}
	o := WithNodeFilter(ov)
	co := &clientOptions{}
	o(co)
	for _, n := range co.nodeFilters {
		ret := n(context.Background(), nil)
		if len(ret) != 1 {
			t.Errorf("expected node  length to be 1, got %v", len(ret))
		}
	}
}

func TestDefaultRequestEncoder(t *testing.T) {
	// 创建一个 HTTP POST 请求，请求体为 JSON 格式
	r, _ := http.NewRequest(http.MethodPost, "", io.NopCloser(bytes.NewBufferString(`{"a":"1", "b": 2}`)))
	// 设置请求头的 Content-Type 为 application/xml
	r.Header.Set("Content-Type", "application/xml")

	// 定义一个结构体，用于测试请求编码器
	v1 := &struct {
		A string `json:"a"`
		B int64  `json:"b"`
	}{"a", 1}
	// 使用默认请求编码器将结构体编码为 JSON 格式的字节数组
	b, err := DefaultRequestEncoder(context.TODO(), "application/json", v1)
	// 如果编码过程中发生错误，测试将失败
	if err != nil {
		t.Fatal(err)
	}
	// 定义一个结构体，用于测试解码后的结果
	v1b := &struct {
		A string `json:"a"`
		B int64  `json:"b"`
	}{}
	// 将编码后的字节数组解码回结构体
	err = json.Unmarshal(b, v1b)
	// 如果解码过程中发生错误，测试将失败
	if err != nil {
		t.Fatal(err)
	}
	// 比较解码后的结构体和原始结构体是否相等
	if !reflect.DeepEqual(v1b, v1) {
		t.Errorf("expected %v, got %v", v1, v1b)
	}
}

func TestDefaultResponseDecoder(t *testing.T) {
	// 创建一个 HTTP 响应，响应体为 JSON 格式
	resp1 := &http.Response{
		Header:     make(http.Header),
		StatusCode: 200,
		Body:       io.NopCloser(bytes.NewBufferString(`{"a":"1", "b": 2}`)),
	}
	// 定义一个结构体，用于测试响应解码器
	v1 := &struct {
		A string `json:"a"`
		B int64  `json:"b"`
	}{}
	// 使用默认响应解码器将响应体解码到结构体中
	err := DefaultResponseDecoder(context.TODO(), resp1, &v1)
	// 如果解码过程中发生错误，测试将失败
	if err != nil {
		t.Fatal(err)
	}
	// 检查解码后的结构体中的值是否正确
	if v1.A != "1" {
		t.Errorf("expected %v, got %v", "1", v1.A)
	}
	if v1.B != int64(2) {
		t.Errorf("expected %v, got %v", 2, v1.B)
	}

	// 创建一个 HTTP 响应，响应体为格式不正确的 JSON
	resp2 := &http.Response{
		Header:     make(http.Header),
		StatusCode: 200,
		Body:       io.NopCloser(bytes.NewBufferString("{badjson}")),
	}
	// 定义一个结构体，用于测试响应解码器
	v2 := &struct {
		A string `json:"a"`
		B int64  `json:"b"`
	}{}
	// 使用默认响应解码器将响应体解码到结构体中
	err = DefaultResponseDecoder(context.TODO(), resp2, &v2)
	// 预期会发生 JSON 语法错误，检查错误类型是否正确
	syntaxErr := &json.SyntaxError{}
	if !errors.As(err, &syntaxErr) {
		t.Errorf("expected %v, got %v", syntaxErr, err)
	}
}

func TestDefaultErrorDecoder(t *testing.T) {
	// 测试 200 到 299 之间的 HTTP 状态码，预期不会有错误
	for i := 200; i < 300; i++ {
		resp := &http.Response{Header: make(http.Header), StatusCode: i}
		if DefaultErrorDecoder(context.TODO(), resp) != nil {
			t.Errorf("expected no error, got %v", DefaultErrorDecoder(context.TODO(), resp))
		}
	}
	// 创建一个 HTTP 响应，状态码为 300，预期不会有错误
	resp1 := &http.Response{
		Header:     make(http.Header),
		StatusCode: 300,
		Body:       io.NopCloser(bytes.NewBufferString("{\"foo\":\"bar\"}")),
	}
	if DefaultErrorDecoder(context.TODO(), resp1) == nil {
		t.Errorf("expected error, got nil")
	}

	// 创建一个 HTTP 响应，状态码为 500，响应体包含错误信息
	resp2 := &http.Response{
		Header:     make(http.Header),
		StatusCode: 500,
		Body:       io.NopCloser(bytes.NewBufferString(`{"code":54321, "message": "hi", "reason": "FOO"}`)),
	}
	// 使用默认错误解码器解析响应体中的错误信息
	err := DefaultErrorDecoder(context.TODO(), resp2)
	// 如果没有解析出错误，测试将失败
	if err == nil {
		t.Errorf("expected error, got nil")
	}
	// 检查解析出的错误信息是否正确
	if err.(*kratoserrors.Error).Code != int32(500) {
		t.Errorf("expected %v, got %v", 500, err.(*kratoserrors.Error).Code)
	}
	if err.(*kratoserrors.Error).Message != "hi" {
		t.Errorf("expected %v, got %v", "hi", err.(*kratoserrors.Error).Message)
	}
	if err.(*kratoserrors.Error).Reason != "FOO" {
		t.Errorf("expected %v, got %v", "FOO", err.(*kratoserrors.Error).Reason)
	}
}

func TestCodecForResponse(t *testing.T) {
	// 创建一个 HTTP 响应，设置响应头的 Content-Type 为 application/xml
	resp := &http.Response{Header: make(http.Header)}
	resp.Header.Set("Content-Type", "application/xml")
	// 根据响应头的 Content-Type 获取对应的编解码器
	c := CodecForResponse(resp)
	// 检查获取到的编解码器名称是否为 xml
	if !reflect.DeepEqual("xml", c.Name()) {
		t.Errorf("expected %v, got %v", "xml", c.Name())
	}
}

func TestNewClient(t *testing.T) {
	// 测试使用默认配置创建一个 HTTP 客户端，连接到指定的端点
	_, err := NewClient(context.Background(), WithEndpoint("127.0.0.1:8888"))
	if err != nil {
		t.Error(err)
	}

	// 测试使用自定义 TLS 配置创建一个 HTTP 客户端，连接到指定的端点
	_, err = NewClient(context.Background(), WithEndpoint("127.0.0.1:9999"), WithTLSConfig(&tls.Config{ServerName: "www.kratos.com", RootCAs: nil}))
	if err != nil {
		t.Error(err)
	}

	// 测试使用服务发现创建一个 HTTP 客户端，连接到指定的服务
	_, err = NewClient(context.Background(), WithDiscovery(&mockDiscovery{}), WithEndpoint("discovery:///go-kratos"))
	if err != nil {
		t.Error(err)
	}

	// 测试使用服务发现和自定义端点创建一个 HTTP 客户端
	_, err = NewClient(context.Background(), WithDiscovery(&mockDiscovery{}), WithEndpoint("127.0.0.1:8888"))
	if err != nil {
		t.Error(err)
	}

	// 测试使用无效的端点地址创建一个 HTTP 客户端，预期会发生错误
	_, err = NewClient(context.Background(), WithEndpoint("127.0.0.1:8888:xxxxa"))
	if err == nil {
		t.Error("except a parseTarget error")
	}

	// 测试使用服务发现和无效的端点地址创建一个 HTTP 客户端，预期会发生错误
	_, err = NewClient(context.Background(), WithDiscovery(&mockDiscovery{}), WithEndpoint("https://go-kratos.dev/"))
	if err == nil {
		t.Error("err should not be equal to nil")
	}

	// 测试创建一个带有中间件的客户端
	client, err := NewClient(
		context.Background(),
		WithDiscovery(&mockDiscovery{}),
		WithEndpoint("discovery:///go-kratos"),
		WithMiddleware(func(handler middleware.Handler) middleware.Handler {
			t.Logf("handle in middleware")
			return func(ctx context.Context, req interface{}) (interface{}, error) {
				return handler(ctx, req)
			}
		}),
	)
	if err != nil {
		t.Fatal(err)
	}

	// 测试客户端的调用功能，预期会发生错误
	err = client.Invoke(context.Background(), http.MethodPost, "/go", map[string]string{"name": "kratos"}, nil, EmptyCallOption{}, &mockCallOption{})
	if err == nil {
		t.Error("err should not be equal to nil")
	}

	// 测试客户端的调用功能，预期会发生错误
	err = client.Invoke(context.Background(), http.MethodPost, "/go", map[string]string{"name": "kratos"}, nil, EmptyCallOption{}, &mockCallOption{needErr: true})
	if err == nil {
		t.Error("err should be equal to callOption err")
	}

	// 测试客户端的编码器，预期会发生错误
	client.opts.encoder = func(context.Context, string, interface{}) (body []byte, err error) {
		return nil, errors.New("mock test encoder error")
	}
	err = client.Invoke(context.Background(), http.MethodPost, "/go", map[string]string{"name": "kratos"}, nil, EmptyCallOption{})
	if err == nil {
		t.Error("err should be equal to encoder error")
	}
}
