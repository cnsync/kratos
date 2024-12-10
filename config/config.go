package config

import (
	"context"
	"errors"
	"reflect"
	"sync"
	"time"

	"dario.cat/mergo"

	// 初始化编码格式支持
	_ "github.com/cnsync/kratos/encoding/json"
	_ "github.com/cnsync/kratos/encoding/proto"
	_ "github.com/cnsync/kratos/encoding/xml"
	_ "github.com/cnsync/kratos/encoding/yaml"
	"github.com/cnsync/kratos/log"
)

// 确保 config 实现了 Config 接口
var _ Config = (*config)(nil)

// ErrNotFound 表示在配置中未找到指定键。
var ErrNotFound = errors.New("key not found")

// Observer 是配置观察者的类型定义。
type Observer func(string, Value)

// Config 是配置接口。
type Config interface {
	Load() error                        // 加载配置
	Scan(v interface{}) error           // 将配置解析到目标结构体
	Value(key string) Value             // 获取指定键的配置值
	Watch(key string, o Observer) error // 监听指定键的变化
	Close() error                       // 关闭配置监听器
}

type config struct {
	opts      options   // 配置选项
	reader    Reader    // 配置读取器
	cached    sync.Map  // 缓存的配置键值对
	observers sync.Map  // 监听器（键 -> Observer）
	watchers  []Watcher // 配置源的监听器列表
}

// New 创建一个配置实例并应用选项。
func New(opts ...Option) Config {
	o := options{
		decoder:  defaultDecoder,
		resolver: defaultResolver,
		merge: func(dst, src interface{}) error {
			return mergo.Map(dst, src, mergo.WithOverride) // 使用 mergo 合并配置
		},
	}
	for _, opt := range opts {
		opt(&o)
	}
	return &config{
		opts:   o,
		reader: newReader(o),
	}
}

// watch 启动对配置源的监听，处理变更。
func (c *config) watch(w Watcher) {
	for {
		kvs, err := w.Next() // 获取下一个变更
		if err != nil {
			if errors.Is(err, context.Canceled) {
				log.Infof("watcher's ctx cancel : %v", err)
				return
			}
			time.Sleep(time.Second) // 遇到错误时延迟重试
			log.Errorf("failed to watch next config: %v", err)
			continue
		}
		if err := c.reader.Merge(kvs...); err != nil {
			log.Errorf("failed to merge next config: %v", err)
			continue
		}
		if err := c.reader.Resolve(); err != nil {
			log.Errorf("failed to resolve next config: %v", err)
			continue
		}
		// 遍历缓存并更新值
		c.cached.Range(func(key, value interface{}) bool {
			k := key.(string)
			v := value.(Value)
			if n, ok := c.reader.Value(k); ok && reflect.TypeOf(n.Load()) == reflect.TypeOf(v.Load()) && !reflect.DeepEqual(n.Load(), v.Load()) {
				v.Store(n.Load())                     // 更新缓存
				if o, ok := c.observers.Load(k); ok { // 通知监听器
					o.(Observer)(k, v)
				}
			}
			return true
		})
	}
}

// Load 加载配置并启动监听。
func (c *config) Load() error {
	for _, src := range c.opts.sources {
		kvs, err := src.Load()
		if err != nil {
			return err
		}
		for _, v := range kvs {
			log.Debugf("config loaded: %s format: %s", v.Key, v.Format)
		}
		if err = c.reader.Merge(kvs...); err != nil {
			log.Errorf("failed to merge config source: %v", err)
			return err
		}
		w, err := src.Watch() // 创建监听器
		if err != nil {
			log.Errorf("failed to watch config source: %v", err)
			return err
		}
		c.watchers = append(c.watchers, w)
		go c.watch(w) // 异步启动监听
	}
	if err := c.reader.Resolve(); err != nil {
		log.Errorf("failed to resolve config source: %v", err)
		return err
	}
	return nil
}

// Value 获取指定键的配置值。
func (c *config) Value(key string) Value {
	if v, ok := c.cached.Load(key); ok {
		return v.(Value)
	}
	if v, ok := c.reader.Value(key); ok {
		c.cached.Store(key, v) // 缓存值
		return v
	}
	return &errValue{err: ErrNotFound} // 未找到返回错误值
}

// Scan 将配置解析到指定结构体。
func (c *config) Scan(v interface{}) error {
	data, err := c.reader.Source() // 获取原始配置数据
	if err != nil {
		return err
	}
	return unmarshalJSON(data, v) // 使用 JSON 解码
}

// Watch 监听指定键的配置变化。
func (c *config) Watch(key string, o Observer) error {
	if v := c.Value(key); v.Load() == nil {
		return ErrNotFound
	}
	c.observers.Store(key, o) // 存储监听器
	return nil
}

// Close 关闭所有监听器。
func (c *config) Close() error {
	for _, w := range c.watchers {
		if err := w.Stop(); err != nil {
			return err
		}
	}
	return nil
}
