package matcher

import (
	"sort"
	"strings"

	"github.com/cnsync/kratos/middleware"
)

// Matcher 是一个中间件匹配器。
type Matcher interface {
	// Use 设置默认的中间件。
	Use(ms ...middleware.Middleware)
	// Add 添加特定选择器的中间件。
	Add(selector string, ms ...middleware.Middleware)
	// Match 根据操作字符串匹配并返回相应的中间件。
	Match(operation string) []middleware.Middleware
}

// New 创建一个新的中间件匹配器。
func New() Matcher {
	return &matcher{
		matches: make(map[string][]middleware.Middleware),
	}
}

// matcher 是 Matcher 接口的实现。
type matcher struct {
	// prefix 存储前缀匹配的选择器。
	prefix []string
	// defaults 存储默认的中间件。
	defaults []middleware.Middleware
	// matches 存储选择器和对应的中间件。
	matches map[string][]middleware.Middleware
}

// Use 设置默认的中间件。
func (m *matcher) Use(ms ...middleware.Middleware) {
	m.defaults = ms
}

// Add 添加特定选择器的中间件。
func (m *matcher) Add(selector string, ms ...middleware.Middleware) {
	if strings.HasSuffix(selector, "*") {
		selector = strings.TrimSuffix(selector, "*")
		m.prefix = append(m.prefix, selector)
		// 对前缀进行排序：
		//  - /foo/bar
		//  - /foo
		sort.Slice(m.prefix, func(i, j int) bool {
			return m.prefix[i] > m.prefix[j]
		})
	}
	m.matches[selector] = ms
}

// Match 根据操作字符串匹配并返回相应的中间件。
func (m *matcher) Match(operation string) []middleware.Middleware {
	ms := make([]middleware.Middleware, 0, len(m.defaults))
	if len(m.defaults) > 0 {
		ms = append(ms, m.defaults...)
	}
	if next, ok := m.matches[operation]; ok {
		return append(ms, next...)
	}
	for _, prefix := range m.prefix {
		if strings.HasPrefix(operation, prefix) {
			return append(ms, m.matches[prefix]...)
		}
	}
	return ms
}
