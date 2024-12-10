package file

import (
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/cnsync/kratos/config"
)

// 定义一个名为 file 的结构体，实现了 config.Source 接口
var _ config.Source = (*file)(nil)

// file 结构体表示一个文件或目录源，用于加载和监视配置文件
type file struct {
	// path 字段存储文件或目录的路径
	path string
}

// NewSource 函数创建一个新的 file 源实例，该实例可以从文件或目录中加载配置
func NewSource(path string) config.Source {
	return &file{path: path}
}

// loadFile 方法从指定的文件路径加载配置，并将其作为键值对返回
func (f *file) loadFile(path string) (*config.KeyValue, error) {
	// 打开文件
	file, err := os.Open(path)
	if err != nil {
		// 如果打开文件失败，返回错误
		return nil, err
	}
	// 延迟关闭文件
	defer file.Close()
	// 读取文件内容
	data, err := io.ReadAll(file)
	if err != nil {
		// 如果读取文件内容失败，返回错误
		return nil, err
	}
	// 获取文件信息
	info, err := file.Stat()
	if err != nil {
		// 如果获取文件信息失败，返回错误
		return nil, err
	}
	// 返回键值对
	return &config.KeyValue{
		Key:    info.Name(),
		Format: format(info.Name()),
		Value:  data,
	}, nil
}

// loadDir 方法从指定的目录路径加载配置，并将其作为键值对列表返回
func (f *file) loadDir(path string) (kvs []*config.KeyValue, err error) {
	// 读取目录下的所有文件和目录
	files, err := os.ReadDir(path)
	if err != nil {
		// 如果读取目录失败，返回错误
		return nil, err
	}
	// 遍历所有文件和目录
	for _, file := range files {
		// 忽略隐藏文件和目录
		if file.IsDir() || strings.HasPrefix(file.Name(), ".") {
			continue
		}
		// 加载文件
		kv, err := f.loadFile(filepath.Join(path, file.Name()))
		if err != nil {
			// 如果加载文件失败，返回错误
			return nil, err
		}
		// 将键值对添加到列表中
		kvs = append(kvs, kv)
	}
	// 返回键值对列表
	return
}

// Load 方法从文件或目录中加载配置，并将其作为键值对列表返回
func (f *file) Load() (kvs []*config.KeyValue, err error) {
	// 获取文件或目录的信息
	fi, err := os.Stat(f.path)
	// 如果获取信息失败，返回错误
	if err != nil {
		return nil, err
	}
	// 如果是目录，调用 loadDir 方法加载配置
	if fi.IsDir() {
		return f.loadDir(f.path)
	}
	// 如果是文件，调用 loadFile 方法加载配置
	kv, err := f.loadFile(f.path)
	// 如果加载失败，返回错误
	if err != nil {
		return nil, err
	}
	// 返回加载的键值对列表
	return []*config.KeyValue{kv}, nil
}

// Watch 方法创建一个新的 Watcher 实例，用于监视文件或目录的变化
func (f *file) Watch() (config.Watcher, error) {
	// 创建一个新的 watcher 实例
	return newWatcher(f)
}
