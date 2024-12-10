package filter

import (
	"context"

	"github.com/cnsync/kratos/selector"
)

// Version 函数根据指定的版本号过滤节点列表
func Version(version string) selector.NodeFilter {
	// 返回一个函数，该函数接受上下文和节点列表作为参数，并返回过滤后的节点列表
	return func(_ context.Context, nodes []selector.Node) []selector.Node {
		// 创建一个新的节点列表，用于存储过滤后的节点
		newNodes := make([]selector.Node, 0, len(nodes))
		// 遍历输入的节点列表
		for _, n := range nodes {
			// 如果节点的版本号与指定的版本号相同，则将该节点添加到新的节点列表中
			if n.Version() == version {
				newNodes = append(newNodes, n)
			}
		}
		// 返回过滤后的节点列表
		return newNodes
	}
}
