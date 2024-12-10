package selector

// SelectOptions 是选择器的选项。
type SelectOptions struct {
	// NodeFilters 是节点过滤器列表。
	NodeFilters []NodeFilter
}

// SelectOption 是选择器的选项函数。
type SelectOption func(*SelectOptions)

// WithNodeFilter 是一个选项函数，用于设置节点过滤器。
func WithNodeFilter(fn ...NodeFilter) SelectOption {
	return func(opts *SelectOptions) {
		// 将传入的节点过滤器添加到选项的 NodeFilters 列表中。
		opts.NodeFilters = fn
	}
}
