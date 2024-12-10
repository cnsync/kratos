package pprof

import (
	"net/http"
	"net/http/pprof"
)

// NewHandler 初始化一个新的 HTTP 处理器，用于处理 pprof 相关的请求。
// 该处理器会将请求分发到不同的 pprof 处理函数，如 /debug/pprof/、/debug/pprof/cmdline 等。
func NewHandler() http.Handler {
	mux := http.NewServeMux()
	// 处理 /debug/pprof/ 请求，返回 pprof 索引页面。
	mux.HandleFunc("/debug/pprof/", pprof.Index)
	// 处理 /debug/pprof/cmdline 请求，返回程序启动命令行参数。
	mux.HandleFunc("/debug/pprof/cmdline", pprof.Cmdline)
	// 处理 /debug/pprof/profile 请求，启动 CPU 性能分析，并返回分析数据。
	mux.HandleFunc("/debug/pprof/profile", pprof.Profile)
	// 处理 /debug/pprof/symbol 请求，用于解析程序中的符号信息。
	mux.HandleFunc("/debug/pprof/symbol", pprof.Symbol)
	// 处理 /debug/pprof/trace 请求，启动跟踪分析，并返回跟踪数据。
	mux.HandleFunc("/debug/pprof/trace", pprof.Trace)
	return mux
}
