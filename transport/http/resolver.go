package http

import (
	"context"
	"errors"
	"net/url"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/cnsync/kratos/internal/endpoint"
	"github.com/cnsync/kratos/log"
	"github.com/cnsync/kratos/registry"
	"github.com/cnsync/kratos/selector"
	"github.com/go-kratos/aegis/subset"
)

// Target 表示解析后的目标信息，包含协议（Scheme）、授权信息（Authority）和终端（Endpoint）
// 该结构体用于描述服务的目标地址
type Target struct {
	Scheme    string // 协议（http 或 https）
	Authority string // 授权信息（主机名或IP地址）
	Endpoint  string // 服务的终端路径
}

// parseTarget 函数用于解析服务的目标地址。
// 根据是否为不安全连接（insecure），它会为地址添加 http 或 https 前缀。
// 返回解析后的 Target 对象及可能的错误信息。
func parseTarget(endpoint string, insecure bool) (*Target, error) {
	// 如果目标地址中不包含 "://", 则根据 insecure 决定使用 http 还是 https
	if !strings.Contains(endpoint, "://") {
		if insecure {
			endpoint = "http://" + endpoint
		} else {
			endpoint = "https://" + endpoint
		}
	}

	// 解析 URL 地址
	u, err := url.Parse(endpoint)
	if err != nil {
		return nil, err
	}

	// 构建 Target 对象，提取协议、主机和路径
	target := &Target{Scheme: u.Scheme, Authority: u.Host}
	if len(u.Path) > 1 {
		target.Endpoint = u.Path[1:] // 获取去掉前导 "/" 的路径部分
	}
	return target, nil
}

// resolver 是一个解析器，它负责从服务发现系统中获取服务实例并进行负载均衡。
// 该结构体包含目标信息、服务发现的 Watcher 和负载均衡器等。
type resolver struct {
	rebalancer selector.Rebalancer // 负载均衡器

	target      *Target          // 目标服务的解析信息
	watcher     registry.Watcher // 服务发现的观察者
	selectorKey string           // 选择器的唯一标识符
	subsetSize  int              // 子集大小，用于筛选服务实例
	insecure    bool             // 是否使用不安全的 HTTP（http://）
}

// newResolver 创建并初始化一个新的 resolver。
// 它会启动一个 goroutine 用于实时获取并更新服务实例。
func newResolver(ctx context.Context, discovery registry.Discovery, target *Target,
	rebalancer selector.Rebalancer, block, insecure bool, subsetSize int,
) (*resolver, error) {
	// 使用服务发现系统的 Watch 方法来监控目标服务的实例变化
	watcher, err := discovery.Watch(ctx, target.Endpoint)
	if err != nil {
		return nil, err
	}

	// 创建 resolver 实例
	r := &resolver{
		target:      target,
		watcher:     watcher,
		rebalancer:  rebalancer,
		insecure:    insecure,
		selectorKey: uuid.New().String(),
		subsetSize:  subsetSize,
	}

	// 如果 block 为 true，则阻塞直到获取到服务实例并更新
	if block {
		done := make(chan error, 1)
		go func() {
			for {
				// 获取下一个服务实例列表
				services, err := watcher.Next()
				if err != nil {
					done <- err
					return
				}
				// 更新服务实例
				if r.update(services) {
					done <- nil
					return
				}
			}
		}()
		// 等待直到获取到更新，或发生错误或上下文超时
		select {
		case err := <-done:
			if err != nil {
				stopErr := watcher.Stop()
				if stopErr != nil {
					log.Errorf("failed to http client watch stop: %v, error: %+v", target, stopErr)
				}
				return nil, err
			}
		case <-ctx.Done():
			// 上下文超时，停止 watcher
			log.Errorf("http client watch service %v reaching context deadline!", target)
			stopErr := watcher.Stop()
			if stopErr != nil {
				log.Errorf("failed to http client watch stop: %v, error: %+v", target, stopErr)
			}
			return nil, ctx.Err()
		}
	}

	// 启动一个 goroutine 来持续监听服务实例的变化
	go func() {
		for {
			services, err := watcher.Next()
			if err != nil {
				if errors.Is(err, context.Canceled) {
					return
				}
				log.Errorf("http client watch service %v got unexpected error:=%v", target, err)
				time.Sleep(time.Second)
				continue
			}
			// 更新服务实例
			r.update(services)
		}
	}()
	return r, nil
}

// update 根据从服务发现系统中获取到的服务实例列表，更新负载均衡节点。
// 它会过滤掉无法解析或无效的服务实例，并应用负载均衡策略。
func (r *resolver) update(services []*registry.ServiceInstance) bool {
	// 过滤服务实例，移除无效的服务实例
	filtered := make([]*registry.ServiceInstance, 0, len(services))
	for _, ins := range services {
		ept, err := endpoint.ParseEndpoint(ins.Endpoints, endpoint.Scheme("http", !r.insecure))
		if err != nil {
			log.Errorf("Failed to parse (%v) discovery endpoint: %v error %v", r.target, ins.Endpoints, err)
			continue
		}
		if ept == "" {
			continue
		}
		filtered = append(filtered, ins)
	}

	// 如果 subsetSize 不为零，则使用 subset 策略从中选择一部分服务实例
	if r.subsetSize != 0 {
		filtered = subset.Subset(r.selectorKey, filtered, r.subsetSize)
	}

	// 构造负载均衡所需的节点列表
	nodes := make([]selector.Node, 0, len(filtered))
	for _, ins := range filtered {
		ept, _ := endpoint.ParseEndpoint(ins.Endpoints, endpoint.Scheme("http", !r.insecure))
		nodes = append(nodes, selector.NewNode("http", ept, ins))
	}

	// 如果没有有效的服务节点，则记录警告日志并返回 false
	if len(nodes) == 0 {
		log.Warnf("[http resolver]Zero endpoint found,refused to write,set: %s ins: %v", r.target.Endpoint, nodes)
		return false
	}

	// 将节点应用到负载均衡器
	r.rebalancer.Apply(nodes)
	return true
}

// Close 停止服务观察者 watcher，并释放相关资源。
func (r *resolver) Close() error {
	return r.watcher.Stop()
}
