package http

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"net/url"

	"github.com/gorilla/mux"

	"github.com/cnsync/kratos/encoding"
	"github.com/cnsync/kratos/errors"
	"github.com/cnsync/kratos/internal/httputil"
	"github.com/cnsync/kratos/transport/http/binding"
)

// SupportPackageIsVersion1 这些常量不应该被任何其他代码引用。
const SupportPackageIsVersion1 = true

// Redirector 接口用于回复请求，重定向到指定的 URL，该 URL 可以是相对于请求路径的路径。
type Redirector interface {
	Redirect() (string, int)
}

// Request 类型是 net/http 的 Request 类型。
type Request = http.Request

// ResponseWriter 类型是 net/http 的 ResponseWriter 类型。
type ResponseWriter = http.ResponseWriter

// Flusher 类型是 net/http 的 Flusher 类型。
type Flusher = http.Flusher

// DecodeRequestFunc 是解码请求的函数。
type DecodeRequestFunc func(*http.Request, interface{}) error

// EncodeResponseFunc 是编码响应的函数。
type EncodeResponseFunc func(http.ResponseWriter, *http.Request, interface{}) error

// EncodeErrorFunc 是编码错误的函数。
type EncodeErrorFunc func(http.ResponseWriter, *http.Request, error)

// DefaultRequestVars 解码请求变量到对象。
func DefaultRequestVars(r *http.Request, v interface{}) error {
	raws := mux.Vars(r)
	vars := make(url.Values, len(raws))
	for k, v := range raws {
		vars[k] = []string{v}
	}
	return binding.BindQuery(vars, v)
}

// DefaultRequestQuery 解码请求查询字符串到对象。
func DefaultRequestQuery(r *http.Request, v interface{}) error {
	return binding.BindQuery(r.URL.Query(), v)
}

// DefaultRequestDecoder 解码请求体到对象。
func DefaultRequestDecoder(r *http.Request, v interface{}) error {
	codec, ok := CodecForRequest(r, "Content-Type")
	if !ok {
		return errors.BadRequest("CODEC", fmt.Sprintf("unregister Content-Type: %s", r.Header.Get("Content-Type")))
	}
	data, err := io.ReadAll(r.Body)

	// 重置请求体。
	r.Body = io.NopCloser(bytes.NewBuffer(data))

	if err != nil {
		return errors.BadRequest("CODEC", err.Error())
	}
	if len(data) == 0 {
		return nil
	}
	if err = codec.Unmarshal(data, v); err != nil {
		return errors.BadRequest("CODEC", fmt.Sprintf("body unmarshal %s", err.Error()))
	}
	return nil
}

// DefaultResponseEncoder 编码对象到 HTTP 响应。
func DefaultResponseEncoder(w http.ResponseWriter, r *http.Request, v interface{}) error {
	if v == nil {
		return nil
	}
	if rd, ok := v.(Redirector); ok {
		url, code := rd.Redirect()
		http.Redirect(w, r, url, code)
		return nil
	}
	codec, _ := CodecForRequest(r, "Accept")
	data, err := codec.Marshal(v)
	if err != nil {
		return err
	}
	w.Header().Set("Content-Type", httputil.ContentType(codec.Name()))
	_, err = w.Write(data)
	if err != nil {
		return err
	}
	return nil
}

// DefaultErrorEncoder 编码错误到 HTTP 响应。
func DefaultErrorEncoder(w http.ResponseWriter, r *http.Request, err error) {
	se := errors.FromError(err)
	codec, _ := CodecForRequest(r, "Accept")
	body, err := codec.Marshal(se)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", httputil.ContentType(codec.Name()))
	w.WriteHeader(int(se.Code))
	_, _ = w.Write(body)
}

// CodecForRequest 通过 HTTP 请求获取编码解码器。
func CodecForRequest(r *http.Request, name string) (encoding.Codec, bool) {
	for _, accept := range r.Header[name] {
		codec := encoding.GetCodec(httputil.ContentSubtype(accept))
		if codec != nil {
			return codec, true
		}
	}
	return encoding.GetCodec("json"), false
}
