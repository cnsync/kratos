package http

import (
	"bytes"
	"io"
	"net/http"
	"testing"

	"github.com/cnsync/kratos/errors"
)

// TestDefaultRequestDecoder 测试默认请求解码器
func TestDefaultRequestDecoder(t *testing.T) {
	var (
		// 定义请求体字符串
		bodyStr = `{"a":"1", "b": 2}`
		// 创建一个 HTTP POST 请求，请求体为 JSON 格式
		r, _ = http.NewRequest(http.MethodPost, "", io.NopCloser(bytes.NewBufferString(bodyStr)))
	)
	// 设置请求头的 Content-Type 为 application/json
	r.Header.Set("Content-Type", "application/json")

	// 定义一个结构体，用于测试请求编码器
	v1 := &struct {
		A string `json:"a"`
		B int64  `json:"b"`
	}{}
	// 使用默认请求解码器将请求体解码到结构体中
	err := DefaultRequestDecoder(r, &v1)
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

	// 读取请求体中的数据
	data, err := io.ReadAll(r.Body)
	// 如果读取过程中发生错误，测试将失败
	if err != nil {
		t.Fatal(err)
	}
	// 比较读取到的数据和原始请求体字符串是否相等
	if bodyStr != string(data) {
		t.Errorf("expected %v, got %v", bodyStr, string(data))
	}
}

// mockResponseWriter 是一个用于模拟 HTTP 响应写入器的结构体
type mockResponseWriter struct {
	// 响应的状态码
	StatusCode int
	// 响应的数据
	Data []byte
	// 响应头
	header http.Header
}

// Header 返回响应头
func (w *mockResponseWriter) Header() http.Header {
	return w.header
}

// Write 将数据写入响应体
func (w *mockResponseWriter) Write(b []byte) (int, error) {
	w.Data = b
	return len(b), nil
}

// WriteHeader 设置响应的状态码
func (w *mockResponseWriter) WriteHeader(statusCode int) {
	w.StatusCode = statusCode
}

// TestDefaultResponseEncoder 测试默认响应编码器
func TestDefaultResponseEncoder(t *testing.T) {
	var (
		// 创建一个模拟的响应写入器
		w = &mockResponseWriter{StatusCode: 200, header: make(http.Header)}
		// 创建一个 HTTP 请求
		r, _ = http.NewRequest(http.MethodPost, "", nil)
		// 定义一个结构体，用于测试响应编码器
		v = &struct {
			A string `json:"a"`
			B int64  `json:"b"`
		}{
			A: "1",
			B: 2,
		}
	)
	// 设置请求头的 Content-Type 为 application/json
	r.Header.Set("Content-Type", "application/json")

	// 使用默认响应编码器将结构体编码为 JSON 格式并写入响应
	err := DefaultResponseEncoder(w, r, v)
	// 如果编码过程中发生错误，测试将失败
	if err != nil {
		t.Fatal(err)
	}
	// 检查响应头的 Content-Type 是否正确
	if w.Header().Get("Content-Type") != "application/json" {
		t.Errorf("expected %v, got %v", "application/json", w.Header().Get("Content-Type"))
	}
	// 检查响应的状态码是否正确
	if w.StatusCode != 200 {
		t.Errorf("expected %v, got %v", 200, w.StatusCode)
	}
	// 检查响应体是否不为空
	if w.Data == nil {
		t.Errorf("expected not nil, got %v", w.Data)
	}
}

// TestDefaultErrorEncoder 测试默认错误编码器
func TestDefaultErrorEncoder(t *testing.T) {
	var (
		// 创建一个模拟的响应写入器
		w = &mockResponseWriter{header: make(http.Header)}
		// 创建一个 HTTP 请求
		r, _ = http.NewRequest(http.MethodPost, "", nil)
		// 定义一个错误对象
		err = errors.New(511, "", "")
	)
	// 设置请求头的 Content-Type 为 application/json
	r.Header.Set("Content-Type", "application/json")

	// 使用默认错误编码器将错误对象编码为 JSON 格式并写入响应
	DefaultErrorEncoder(w, r, err)
	// 检查响应头的 Content-Type 是否正确
	if w.Header().Get("Content-Type") != "application/json" {
		t.Errorf("expected %v, got %v", "application/json", w.Header().Get("Content-Type"))
	}
	// 检查响应的状态码是否正确
	if w.StatusCode != 511 {
		t.Errorf("expected %v, got %v", 511, w.StatusCode)
	}
	// 检查响应体是否不为空
	if w.Data == nil {
		t.Errorf("expected not nil, got %v", w.Data)
	}
}

// TestDefaultResponseEncoderEncodeNil 测试默认响应编码器在编码空值时的行为
func TestDefaultResponseEncoderEncodeNil(t *testing.T) {
	var (
		// 创建一个模拟的响应写入器，状态码为 204
		w = &mockResponseWriter{StatusCode: 204, header: make(http.Header)}
		// 创建一个 HTTP 请求，请求体为 XML 格式
		r, _ = http.NewRequest(http.MethodPost, "", io.NopCloser(bytes.NewBufferString("<xml></xml>")))
	)
	// 设置请求头的 Content-Type 为 application/json
	r.Header.Set("Content-Type", "application/json")

	// 使用默认响应编码器将 nil 编码并写入响应
	err := DefaultResponseEncoder(w, r, nil)
	// 如果编码过程中发生错误，测试将失败
	if err != nil {
		t.Fatal(err)
	}
	// 检查响应头的 Content-Type 是否为空字符串
	if w.Header().Get("Content-Type") != "" {
		t.Errorf("expected empty string, got %v", w.Header().Get("Content-Type"))
	}
	// 检查响应的状态码是否正确
	if w.StatusCode != 204 {
		t.Errorf("expected %v, got %v", 204, w.StatusCode)
	}
	// 检查响应体是否为 nil
	if w.Data != nil {
		t.Errorf("expected nil, got %v", w.Data)
	}
}

// TestCodecForRequest 测试根据请求头的 Content-Type 获取编解码器
func TestCodecForRequest(t *testing.T) {
	// 创建一个 HTTP POST 请求，请求体为 XML 格式
	r, _ := http.NewRequest(http.MethodPost, "", io.NopCloser(bytes.NewBufferString("<xml></xml>")))
	// 设置请求头的 Content-Type 为 application/xml
	r.Header.Set("Content-Type", "application/xml")
	// 根据请求头的 Content-Type 获取对应的编解码器
	c, ok := CodecForRequest(r, "Content-Type")
	// 如果获取失败，测试将失败
	if !ok {
		t.Fatalf("expected true, got %v", ok)
	}
	// 检查获取到的编解码器名称是否为 xml
	if c.Name() != "xml" {
		t.Errorf("expected %v, got %v", "xml", c.Name())
	}

	// 创建一个 HTTP POST 请求，请求体为 JSON 格式
	r, _ = http.NewRequest(http.MethodPost, "", io.NopCloser(bytes.NewBufferString(`{"a":"1", "b": 2}`)))
	// 设置请求头的 Content-Type 为一个无效的值
	r.Header.Set("Content-Type", "blablablabla")
	// 根据请求头的 Content-Type 获取对应的编解码器
	c, ok = CodecForRequest(r, "Content-Type")
	// 如果获取成功，测试将失败
	if ok {
		t.Fatalf("expected false, got %v", ok)
	}
	// 检查获取到的编解码器名称是否为 json
	if c.Name() != "json" {
		t.Errorf("expected %v, got %v", "json", c.Name())
	}
}
