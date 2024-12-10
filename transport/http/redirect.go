package http

// redirect 结构体表示一个 HTTP 重定向。
type redirect struct {
	// URL 是重定向的目标 URL。
	URL string
	// Code 是重定向的状态码。
	Code int
}

// Redirect 方法返回重定向的 URL 和状态码。
func (r *redirect) Redirect() (string, int) {
	return r.URL, r.Code
}

// NewRedirect 函数创建一个新的重定向实例。
// 参数：
//   - url：重定向的目标 URL，可以是相对于请求路径的路径。
//   - code：重定向的状态码，应该在 3xx 范围内，通常是 StatusMovedPermanently、StatusFound 或 StatusSeeOther。
//
// 返回值：
//   - Redirector：创建的重定向实例。
func NewRedirect(url string, code int) Redirector {
	return &redirect{URL: url, Code: code}
}
