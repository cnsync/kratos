package selector

import "context"

// NodeFilter 是一个选择过滤器。
type NodeFilter func(context.Context, []Node) []Node
