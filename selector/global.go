package selector

// 定义一个全局的选择器实例
var globalSelector = &wrapSelector{}

// 确保 wrapSelector 结构体实现了 Builder 接口
var _ Builder = (*wrapSelector)(nil)

// wrapSelector 结构体，用于包装选择器，帮助覆盖全局选择器的实现
type wrapSelector struct{ Builder }

// GlobalSelector 函数，返回全局选择器的构建器
func GlobalSelector() Builder {
	// 如果全局选择器的构建器不为空，则返回全局选择器
	if globalSelector.Builder != nil {
		return globalSelector
	}
	// 如果全局选择器的构建器为空，则返回 nil
	return nil
}

// SetGlobalSelector 函数，设置全局选择器的构建器
func SetGlobalSelector(builder Builder) {
	// 设置全局选择器的构建器为传入的构建器
	globalSelector.Builder = builder
}
