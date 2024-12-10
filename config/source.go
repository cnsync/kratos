package config

// KeyValue 是配置的键值对。
type KeyValue struct {
	// Key 是配置的键。
	Key string
	// Value 是配置的值，以字节切片的形式存储。
	Value []byte
	// Format 是配置值的格式，例如 JSON、YAML 等。
	Format string
}

// Source 是配置源的接口。
type Source interface {
	// Load 方法从配置源加载配置键值对。
	Load() ([]*KeyValue, error)
	// Watch 方法创建一个监控器，用于监控配置源的变化。
	Watch() (Watcher, error)
}

// Watcher 是监控器的接口，用于监控配置源的变化。
type Watcher interface {
	// Next 方法等待并返回配置源的下一次变化。
	Next() ([]*KeyValue, error)
	// Stop 方法停止监控器。
	Stop() error
}
