package env

import (
	"os"
	"strings"

	"github.com/cnsync/kratos/config"
)

// env 结构体表示一个环境变量源，它包含一个前缀列表。
type env struct {
	prefixes []string
}

// NewSource 函数创建一个新的 env 源实例，该实例可以从环境变量中加载配置。
func NewSource(prefixes ...string) config.Source {
	return &env{prefixes: prefixes}
}

// Load 方法从环境变量中加载配置，并将其作为键值对列表返回。
func (e *env) Load() (kv []*config.KeyValue, err error) {
	return e.load(os.Environ()), nil
}

// load 方法从给定的环境变量列表中加载配置，并将其作为键值对列表返回。
func (e *env) load(envs []string) []*config.KeyValue {
	var kv []*config.KeyValue
	// 遍历环境变量列表
	for _, env := range envs {
		var k, v string
		// 将环境变量分割成键和值
		subs := strings.SplitN(env, "=", 2) //nolint:mnd
		k = subs[0]
		// 如果有值，则赋值给 v
		if len(subs) > 1 {
			v = subs[1]
		}

		// 如果有前缀，则检查键是否以前缀开头
		if len(e.prefixes) > 0 {
			p, ok := matchPrefix(e.prefixes, k)
			// 如果键不以前缀开头，或者前缀长度等于键长度，则跳过
			if !ok || len(p) == len(k) {
				continue
			}
			// 去除键的前缀
			k = strings.TrimPrefix(k, p)
			k = strings.TrimPrefix(k, "_")
		}

		// 如果键不为空，则将键值对添加到 kv 列表中
		if len(k) != 0 {
			kv = append(kv, &config.KeyValue{
				Key:   k,
				Value: []byte(v),
			})
		}
	}
	return kv
}

// Watch 方法创建一个新的 Watcher 实例，用于监视环境变量的变化
func (e *env) Watch() (config.Watcher, error) {
	// 创建一个新的 watcher 实例
	w, err := NewWatcher()
	// 如果创建实例失败，返回错误
	if err != nil {
		return nil, err
	}
	// 返回 watcher 实例
	return w, nil
}

// matchPrefix 函数检查给定的字符串是否以任何一个前缀开头，并返回匹配的前缀。
func matchPrefix(prefixes []string, s string) (string, bool) {
	// 遍历前缀列表
	for _, p := range prefixes {
		// 如果字符串 s 以前缀 p 开头
		if strings.HasPrefix(s, p) {
			// 返回匹配的前缀 p 和 true
			return p, true
		}
	}
	// 如果没有匹配的前缀，返回空字符串和 false
	return "", false
}
