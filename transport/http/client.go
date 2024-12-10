package http

import (
	"bytes"
	"context"
	"crypto/tls"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/cnsync/kratos/encoding"
	"github.com/cnsync/kratos/errors"
	"github.com/cnsync/kratos/internal/host"
	"github.com/cnsync/kratos/internal/httputil"
	"github.com/cnsync/kratos/middleware"
	"github.com/cnsync/kratos/registry"
	"github.com/cnsync/kratos/selector"
	"github.com/cnsync/kratos/selector/wrr"
	"github.com/cnsync/kratos/transport"
)

func init() {
	// 初始化全局选择器，如果尚未设置，则使用 WRR 选择器作为默认值。
	if selector.GlobalSelector() == nil {
		selector.SetGlobalSelector(wrr.NewBuilder())
	}
}

// DecodeErrorFunc 是解码错误的函数类型，用于处理解码时的错误。
type DecodeErrorFunc func(ctx context.Context, res *http.Response) error

// EncodeRequestFunc 是请求编码函数的类型，用于将请求数据编码成字节数组。
type EncodeRequestFunc func(ctx context.Context, contentType string, in interface{}) (body []byte, err error)

// DecodeResponseFunc 是响应解码函数的类型，用于将响应数据解码到指定结构中。
type DecodeResponseFunc func(ctx context.Context, res *http.Response, out interface{}) error

// ClientOption 是 HTTP 客户端的配置选项函数类型。
type ClientOption func(*clientOptions)

// clientOptions 用于存储 HTTP 客户端的配置选项。
type clientOptions struct {
	ctx          context.Context         // 上下文对象，用于控制超时等
	tlsConf      *tls.Config             // TLS 配置，用于启用 HTTPS
	timeout      time.Duration           // 请求超时时间
	endpoint     string                  // 目标服务的地址
	userAgent    string                  // 用户代理字符串
	encoder      EncodeRequestFunc       // 请求编码器
	decoder      DecodeResponseFunc      // 响应解码器
	errorDecoder DecodeErrorFunc         // 错误解码器
	transport    http.RoundTripper       // HTTP 请求的传输器
	nodeFilters  []selector.NodeFilter   // 节点选择器过滤器
	discovery    registry.Discovery      // 服务发现接口
	middleware   []middleware.Middleware // 中间件列表
	block        bool                    // 是否阻塞
	subsetSize   int                     // 客户端发现的子集大小
}

// WithSubset 设置客户端发现的子集大小。零值表示禁用子集过滤。
func WithSubset(size int) ClientOption {
	return func(o *clientOptions) {
		o.subsetSize = size
	}
}

// WithTransport 设置客户端的传输器。
func WithTransport(trans http.RoundTripper) ClientOption {
	return func(o *clientOptions) {
		o.transport = trans
	}
}

// WithTimeout 设置客户端请求超时时间。
func WithTimeout(d time.Duration) ClientOption {
	return func(o *clientOptions) {
		o.timeout = d
	}
}

// WithUserAgent 设置客户端的用户代理。
func WithUserAgent(ua string) ClientOption {
	return func(o *clientOptions) {
		o.userAgent = ua
	}
}

// WithMiddleware 设置客户端的中间件。
func WithMiddleware(m ...middleware.Middleware) ClientOption {
	return func(o *clientOptions) {
		o.middleware = m
	}
}

// WithEndpoint 设置客户端的目标地址。
func WithEndpoint(endpoint string) ClientOption {
	return func(o *clientOptions) {
		o.endpoint = endpoint
	}
}

// WithRequestEncoder 设置客户端的请求编码器。
func WithRequestEncoder(encoder EncodeRequestFunc) ClientOption {
	return func(o *clientOptions) {
		o.encoder = encoder
	}
}

// WithResponseDecoder 设置客户端的响应解码器。
func WithResponseDecoder(decoder DecodeResponseFunc) ClientOption {
	return func(o *clientOptions) {
		o.decoder = decoder
	}
}

// WithErrorDecoder 设置客户端的错误解码器。
func WithErrorDecoder(errorDecoder DecodeErrorFunc) ClientOption {
	return func(o *clientOptions) {
		o.errorDecoder = errorDecoder
	}
}

// WithDiscovery 设置客户端的服务发现机制。
func WithDiscovery(d registry.Discovery) ClientOption {
	return func(o *clientOptions) {
		o.discovery = d
	}
}

// WithNodeFilter 设置节点选择过滤器。
func WithNodeFilter(filters ...selector.NodeFilter) ClientOption {
	return func(o *clientOptions) {
		o.nodeFilters = filters
	}
}

// WithBlock 设置客户端为阻塞模式。
func WithBlock() ClientOption {
	return func(o *clientOptions) {
		o.block = true
	}
}

// WithTLSConfig 设置 TLS 配置，用于支持 HTTPS。
func WithTLSConfig(c *tls.Config) ClientOption {
	return func(o *clientOptions) {
		o.tlsConf = c
	}
}

// Client 是 HTTP 客户端的结构体，封装了 HTTP 请求的配置和操作。
type Client struct {
	opts     clientOptions     // 客户端配置选项
	target   *Target           // 目标服务的信息
	r        *resolver         // 解析器，用于服务发现
	cc       *http.Client      // 内部 HTTP 客户端
	insecure bool              // 是否使用不安全的 HTTP（非 HTTPS）
	selector selector.Selector // 服务选择器
}

// NewClient 返回一个新的 HTTP 客户端实例。
func NewClient(ctx context.Context, opts ...ClientOption) (*Client, error) {
	options := clientOptions{
		ctx:          ctx,
		timeout:      2000 * time.Millisecond, // 默认超时时间为 2000 毫秒
		encoder:      DefaultRequestEncoder,   // 默认请求编码器
		decoder:      DefaultResponseDecoder,  // 默认响应解码器
		errorDecoder: DefaultErrorDecoder,     // 默认错误解码器
		transport:    http.DefaultTransport,   // 默认 HTTP 传输器
		subsetSize:   25,                      // 默认子集大小为 25
	}
	// 处理传入的客户端配置选项
	for _, o := range opts {
		o(&options)
	}
	// 如果配置了 TLS 配置，则更新传输器的 TLS 设置
	if options.tlsConf != nil {
		if tr, ok := options.transport.(*http.Transport); ok {
			tr.TLSClientConfig = options.tlsConf
		}
	}
	insecure := options.tlsConf == nil
	// 解析目标地址
	target, err := parseTarget(options.endpoint, insecure)
	if err != nil {
		return nil, err
	}
	// 使用全局选择器构建一个服务选择器
	selector := selector.GlobalSelector().Build()
	var r *resolver
	// 如果配置了服务发现，则创建解析器
	if options.discovery != nil {
		if target.Scheme == "discovery" {
			if r, err = newResolver(ctx, options.discovery, target, selector, options.block, insecure, options.subsetSize); err != nil {
				return nil, fmt.Errorf("[http client] new resolver failed!err: %v", options.endpoint)
			}
		} else if _, _, err := host.ExtractHostPort(options.endpoint); err != nil {
			return nil, fmt.Errorf("[http client] invalid endpoint format: %v", options.endpoint)
		}
	}
	// 返回配置好的客户端实例
	return &Client{
		opts:     options,
		target:   target,
		insecure: insecure,
		r:        r,
		cc: &http.Client{
			Timeout:   options.timeout,
			Transport: options.transport,
		},
		selector: selector,
	}, nil
}

// Invoke 调用远程服务的 RPC 方法。
func (client *Client) Invoke(ctx context.Context, method, path string, args interface{}, reply interface{}, opts ...CallOption) error {
	var (
		contentType string
		body        io.Reader
	)
	c := defaultCallInfo(path)
	// 处理传入的调用选项
	for _, o := range opts {
		if err := o.before(&c); err != nil {
			return err
		}
	}
	// 如果有请求参数，则进行编码
	if args != nil {
		data, err := client.opts.encoder(ctx, c.contentType, args)
		if err != nil {
			return err
		}
		contentType = c.contentType
		body = bytes.NewReader(data)
	}
	// 构建请求 URL
	url := fmt.Sprintf("%s://%s%s", client.target.Scheme, client.target.Authority, path)
	// 创建 HTTP 请求对象
	req, err := http.NewRequest(method, url, body)
	if err != nil {
		return err
	}
	// 设置请求头
	if c.headerCarrier != nil {
		req.Header = *c.headerCarrier
	}

	// 设置请求的 Content-Type
	if contentType != "" {
		req.Header.Set("Content-Type", c.contentType)
	}
	// 设置客户端的用户代理
	if client.opts.userAgent != "" {
		req.Header.Set("User-Agent", client.opts.userAgent)
	}
	// 将请求传递给传输层
	ctx = transport.NewClientContext(ctx, &Transport{
		endpoint:     client.opts.endpoint,
		reqHeader:    headerCarrier(req.Header),
		operation:    c.operation,
		request:      req,
		pathTemplate: c.pathTemplate,
	})
	// 调用客户端的内部请求方法
	return client.invoke(ctx, req, args, reply, c, opts...)
}

// invoke 实际执行 HTTP 请求并处理响应。
func (client *Client) invoke(ctx context.Context, req *http.Request, args interface{}, reply interface{}, c callInfo, opts ...CallOption) error {
	// 定义处理请求的函数
	h := func(ctx context.Context, _ interface{}) (interface{}, error) {
		res, err := client.do(req.WithContext(ctx)) // 发送请求
		if res != nil {
			cs := csAttempt{res: res}
			// 处理调用后的操作
			for _, o := range opts {
				o.after(&c, &cs)
			}
		}
		if err != nil {
			return nil, err
		}
		defer res.Body.Close()
		// 解码响应数据
		if err := client.opts.decoder(ctx, res, reply); err != nil {
			return nil, err
		}
		return reply, nil
	}
	var p selector.Peer
	// 在上下文中设置节点信息
	ctx = selector.NewPeerContext(ctx, &p)
	// 如果有中间件，链式处理
	if len(client.opts.middleware) > 0 {
		h = middleware.Chain(client.opts.middleware...)(h)
	}
	// 执行请求处理函数
	_, err := h(ctx, args)
	return err
}

// Do 发送 HTTP 请求并解码响应数据。
func (client *Client) Do(req *http.Request, opts ...CallOption) (*http.Response, error) {
	c := defaultCallInfo(req.URL.Path)
	// 处理传入的调用选项
	for _, o := range opts {
		if err := o.before(&c); err != nil {
			return nil, err
		}
	}

	// 发送请求
	return client.do(req)
}

// do 实际执行 HTTP 请求并返回响应。
func (client *Client) do(req *http.Request) (*http.Response, error) {
	var done func(context.Context, selector.DoneInfo)
	if client.r != nil {
		var (
			err  error
			node selector.Node
		)
		// 使用选择器选择节点
		if node, done, err = client.selector.Select(req.Context(), selector.WithNodeFilter(client.opts.nodeFilters...)); err != nil {
			return nil, errors.ServiceUnavailable("NODE_NOT_FOUND", err.Error())
		}
		// 根据安全性配置设置请求的 URL 协议（http 或 https）
		if client.insecure {
			req.URL.Scheme = "http"
		} else {
			req.URL.Scheme = "https"
		}
		// 设置请求的主机地址
		req.URL.Host = node.Address()
		req.Host = node.Address()
	}
	// 发送请求并获取响应
	resp, err := client.cc.Do(req)
	if err == nil {
		// 从上下文中提取传输对象并更新响应头
		t, ok := transport.FromClientContext(req.Context())
		if ok {
			ht, ok := t.(*Transport)
			if ok {
				ht.replyHeader = headerCarrier(resp.Header)
			}
		}
		// 解码错误响应
		err = client.opts.errorDecoder(req.Context(), resp)
	}
	// 调用结束时执行的操作
	if done != nil {
		done(req.Context(), selector.DoneInfo{Err: err})
	}
	if err != nil {
		return nil, err
	}
	return resp, nil
}

// Close 关闭客户端并释放相关资源。
func (client *Client) Close() error {
	if client.r != nil {
		return client.r.Close()
	}
	return nil
}

// DefaultRequestEncoder 是默认的请求编码器，将请求数据编码为字节数组。
func DefaultRequestEncoder(_ context.Context, contentType string, in interface{}) ([]byte, error) {
	name := httputil.ContentSubtype(contentType)
	body, err := encoding.GetCodec(name).Marshal(in)
	if err != nil {
		return nil, err
	}
	return body, err
}

// DefaultResponseDecoder 是默认的响应解码器，将响应数据解码到指定结构。
func DefaultResponseDecoder(_ context.Context, res *http.Response, v interface{}) error {
	defer res.Body.Close()
	data, err := io.ReadAll(res.Body)
	if err != nil {
		return err
	}
	return CodecForResponse(res).Unmarshal(data, v)
}

// DefaultErrorDecoder 是默认的错误解码器，检查响应状态码并解码错误信息。
func DefaultErrorDecoder(_ context.Context, res *http.Response) error {
	if res.StatusCode >= 200 && res.StatusCode <= 299 {
		return nil
	}
	defer res.Body.Close()
	data, err := io.ReadAll(res.Body)
	if err == nil {
		e := new(errors.Error)
		if err = CodecForResponse(res).Unmarshal(data, e); err == nil {
			e.Code = int32(res.StatusCode)
			return e
		}
	}
	return errors.Newf(res.StatusCode, errors.UnknownReason, "").WithCause(err)
}

// CodecForResponse 获取适用于响应的编码器。
func CodecForResponse(r *http.Response) encoding.Codec {
	codec := encoding.GetCodec(httputil.ContentSubtype(r.Header.Get("Content-Type")))
	if codec != nil {
		return codec
	}
	return encoding.GetCodec("json")
}
