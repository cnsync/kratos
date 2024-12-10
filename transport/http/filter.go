package http

import "net/http"

// FilterFunc 是一个函数，它接收一个 http.Handler 并返回另一个 http.Handler。
type FilterFunc func(http.Handler) http.Handler

// FilterChain 返回一个 FilterFunc，它指定了 HTTP 路由器的链式处理程序。
func FilterChain(filters ...FilterFunc) FilterFunc {
	return func(next http.Handler) http.Handler {
		// 遍历 filters 切片，从最后一个元素开始向前遍历。
		for i := len(filters) - 1; i >= 0; i-- {
			// 将当前的 next 处理程序作为参数传递给 filters 中的函数，得到一个新的处理程序。
			next = filters[i](next)
		}
		// 返回最终的处理程序。
		return next
	}
}
