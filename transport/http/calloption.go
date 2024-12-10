package http

import (
	"net/http"
)

// CallOption 配置在调用开始前或调用完成后提取信息的接口。
// 在调用发起前可以设置某些选项，在调用完成后也可以提取一些信息。
type CallOption interface {
	// before 在调用发送到任何服务器之前调用。如果 before 返回非 nil 错误，RPC 调用会失败并返回该错误。
	before(*callInfo) error

	// after 在调用完成后调用。after 不能返回错误，因此任何失败都应通过输出参数报告。
	after(*callInfo, *csAttempt)
}

// callInfo 包含有关 HTTP 调用的信息。
// 主要用于存储请求的内容类型、操作名称、路径模板等信息。
type callInfo struct {
	contentType   string       // 请求的内容类型（例如 "application/json"）
	operation     string       // 操作名称，通常为 HTTP 请求的路径或方法
	pathTemplate  string       // 请求路径的模板
	headerCarrier *http.Header // 请求的 HTTP 头信息
}

// EmptyCallOption 不会改变调用的配置。
// 它可以嵌入到其他结构中，用来携带拦截器使用的附加数据。
type EmptyCallOption struct{}

// before 是 EmptyCallOption 的默认实现，什么也不做。
func (EmptyCallOption) before(*callInfo) error { return nil }

// after 是 EmptyCallOption 的默认实现，什么也不做。
func (EmptyCallOption) after(*callInfo, *csAttempt) {}

// csAttempt 用来表示一次 HTTP 调用的尝试，存储了 HTTP 响应。
type csAttempt struct {
	res *http.Response // HTTP 响应
}

// ContentType 是一个设置请求内容类型的调用选项。
// 例如，可以指定请求的内容类型是 "application/json"。
func ContentType(contentType string) CallOption {
	return ContentTypeCallOption{ContentType: contentType}
}

// ContentTypeCallOption 是设置请求内容类型的调用选项。
type ContentTypeCallOption struct {
	EmptyCallOption
	ContentType string // 请求的内容类型
}

// before 设置内容类型到 callInfo 中。
func (o ContentTypeCallOption) before(c *callInfo) error {
	c.contentType = o.ContentType
	return nil
}

// defaultCallInfo 返回一个默认的 callInfo 对象，默认内容类型是 "application/json"。
func defaultCallInfo(path string) callInfo {
	return callInfo{
		contentType:  "application/json", // 默认内容类型
		operation:    path,               // 默认操作名称
		pathTemplate: path,               // 默认路径模板
	}
}

// Operation 是一个设置操作名称的调用选项。
// 操作名称通常是 HTTP 请求的路径。
func Operation(operation string) CallOption {
	return OperationCallOption{Operation: operation}
}

// OperationCallOption 是设置操作名称的调用选项。
type OperationCallOption struct {
	EmptyCallOption
	Operation string // 操作名称
}

// before 设置操作名称到 callInfo 中。
func (o OperationCallOption) before(c *callInfo) error {
	c.operation = o.Operation
	return nil
}

// PathTemplate 是一个设置路径模板的调用选项。
// 路径模板通常用于动态构建请求的路径。
func PathTemplate(pattern string) CallOption {
	return PathTemplateCallOption{Pattern: pattern}
}

// PathTemplateCallOption 是设置路径模板的调用选项。
type PathTemplateCallOption struct {
	EmptyCallOption
	Pattern string // 路径模板
}

// before 设置路径模板到 callInfo 中。
func (o PathTemplateCallOption) before(c *callInfo) error {
	c.pathTemplate = o.Pattern
	return nil
}

// Header 返回一个调用选项，该选项可以获取服务器响应的 HTTP 头。
func Header(header *http.Header) CallOption {
	return HeaderCallOption{header: header}
}

// HeaderCallOption 是获取 HTTP 响应头的调用选项。
type HeaderCallOption struct {
	EmptyCallOption
	header *http.Header // 存储响应头的地方
}

// before 设置 callInfo 中的 headerCarrier 为传入的 header。
func (o HeaderCallOption) before(c *callInfo) error {
	c.headerCarrier = o.header
	return nil
}

// after 在调用完成后提取响应头，并存储到传入的 header 中。
func (o HeaderCallOption) after(_ *callInfo, cs *csAttempt) {
	if cs.res != nil && cs.res.Header != nil {
		*o.header = cs.res.Header
	}
}
