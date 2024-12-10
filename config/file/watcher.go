package file

import (
	"context"
	"os"
	"path/filepath"

	"github.com/cnsync/kratos/config"

	"github.com/fsnotify/fsnotify"
)

// 定义一个名为 watcher 的结构体，实现了 config.Watcher 接口
var _ config.Watcher = (*watcher)(nil)

type watcher struct {
	f  *file
	fw *fsnotify.Watcher

	ctx    context.Context
	cancel context.CancelFunc
}

// newWatcher 函数创建一个新的 watcher 实例，用于监视文件或目录的变化
func newWatcher(f *file) (config.Watcher, error) {
	fw, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, err
	}
	if err := fw.Add(f.path); err != nil {
		return nil, err
	}
	ctx, cancel := context.WithCancel(context.Background())
	return &watcher{f: f, fw: fw, ctx: ctx, cancel: cancel}, nil
}

// Next 方法等待文件或目录的变化，并在发生变化时加载新的配置
func (w *watcher) Next() ([]*config.KeyValue, error) {
	select {
	case <-w.ctx.Done():
		return nil, w.ctx.Err()
	case event := <-w.fw.Events:
		// 如果文件被重命名，则添加新的文件到监听器中
		if event.Op == fsnotify.Rename {
			if _, err := os.Stat(event.Name); err == nil || os.IsExist(err) {
				if err := w.fw.Add(event.Name); err != nil {
					return nil, err
				}
			}
		}
		// 获取文件或目录的信息
		fi, err := os.Stat(w.f.path)
		if err != nil {
			return nil, err
		}
		path := w.f.path
		// 如果是目录，则使用事件中的文件名作为路径
		if fi.IsDir() {
			path = filepath.Join(w.f.path, filepath.Base(event.Name))
		}
		// 加载文件或目录的配置
		kv, err := w.f.loadFile(path)
		if err != nil {
			return nil, err
		}
		return []*config.KeyValue{kv}, nil
	case err := <-w.fw.Errors:
		return nil, err
	}
}

// Stop 方法停止监视，并关闭所有相关资源
func (w *watcher) Stop() error {
	w.cancel()
	return w.fw.Close()
}
