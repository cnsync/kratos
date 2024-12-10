package endpoint

import (
	"net/url"
)

// NewEndpoint 函数用于创建一个新的 URL 端点。
// 参数：
//   - scheme：URL 的协议，如 "http" 或 "https"。
//   - host：URL 的主机名或 IP 地址。
//
// 返回值：
//   - *url.URL：新创建的 URL 端点。
func NewEndpoint(scheme, host string) *url.URL {
	return &url.URL{Scheme: scheme, Host: host}
}

// ParseEndpoint 函数用于解析一组端点字符串，并返回指定协议的第一个匹配的主机名。
// 参数：
//   - endpoints：包含端点字符串的切片。
//   - scheme：要匹配的协议。
//
// 返回值：
//   - string：匹配的主机名，如果没有找到匹配的端点，则返回空字符串。
//   - error：如果在解析过程中发生错误，则返回错误。
func ParseEndpoint(endpoints []string, scheme string) (string, error) {
	for _, e := range endpoints {
		u, err := url.Parse(e)
		if err != nil {
			return "", err
		}

		if u.Scheme == scheme {
			return u.Host, nil
		}
	}
	return "", nil
}

// Scheme 函数用于根据是否需要安全连接来确定 URL 的协议。
// 参数：
//   - scheme：基础协议，如 "http"。
//   - isSecure：是否需要安全连接。
//
// 返回值：
//   - string：根据需要返回的协议，如 "http" 或 "https"。
func Scheme(scheme string, isSecure bool) string {
	if isSecure {
		return scheme + "s"
	}
	return scheme
}
