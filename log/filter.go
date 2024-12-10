package log

// FilterOption 是过滤器的选项。
type FilterOption func(*Filter)

const fuzzyStr = "***"

// FilterLevel 用于设置过滤级别。
func FilterLevel(level Level) FilterOption {
	return func(opts *Filter) {
		opts.level = level
	}
}

// FilterKey 用于设置过滤键。
func FilterKey(key ...string) FilterOption {
	return func(o *Filter) {
		for _, v := range key {
			o.key[v] = struct{}{}
		}
	}
}

// FilterValue 用于设置过滤值。
func FilterValue(value ...string) FilterOption {
	return func(o *Filter) {
		for _, v := range value {
			o.value[v] = struct{}{}
		}
	}
}

// FilterFunc 用于设置过滤函数。
func FilterFunc(f func(level Level, keyvals ...interface{}) bool) FilterOption {
	return func(o *Filter) {
		o.filter = f
	}
}

// Filter 是一个日志过滤器。
type Filter struct {
	logger Logger
	level  Level
	key    map[interface{}]struct{}
	value  map[interface{}]struct{}
	filter func(level Level, keyvals ...interface{}) bool
}

// NewFilter 新建一个日志过滤器。
func NewFilter(logger Logger, opts ...FilterOption) *Filter {
	options := Filter{
		logger: logger,
		key:    make(map[interface{}]struct{}),
		value:  make(map[interface{}]struct{}),
	}
	for _, o := range opts {
		o(&options)
	}
	return &options
}

// Log 根据级别和键值对打印日志。
func (f *Filter) Log(level Level, keyvals ...interface{}) error {
	// 如果日志级别低于过滤器设置的级别，则不记录日志
	if level < f.level {
		return nil
	}

	// prefixkv 包含在日志初始化期间定义为前缀的参数切片
	var prefixkv []interface{}
	// 如果日志记录器实现了 logger 接口，并且有前缀，则将前缀添加到 prefixkv 中
	l, ok := f.logger.(*logger)
	if ok && len(l.prefix) > 0 {
		prefixkv = make([]interface{}, 0, len(l.prefix))
		prefixkv = append(prefixkv, l.prefix...)
	}

	// 如果过滤器函数存在，并且它对前缀或键值对返回 true，则不记录日志
	if f.filter != nil && (f.filter(level, prefixkv...) || f.filter(level, keyvals...)) {
		return nil
	}

	// 如果过滤器设置了键或值，则遍历键值对，将匹配的键或值替换为模糊字符串
	if len(f.key) > 0 || len(f.value) > 0 {
		for i := 0; i < len(keyvals); i += 2 {
			v := i + 1
			if v >= len(keyvals) {
				continue
			}
			if _, ok := f.key[keyvals[i]]; ok {
				keyvals[v] = fuzzyStr
			}
			if _, ok := f.value[keyvals[v]]; ok {
				keyvals[v] = fuzzyStr
			}
		}
	}

	// 记录过滤后的日志
	return f.logger.Log(level, keyvals...)
}
