package selector

import (
	"context"
	"testing"
)

// TestPeer 测试 Peer 结构体和相关方法
func TestPeer(t *testing.T) {
	// 创建一个 Peer 实例，其中包含一个模拟的加权节点
	p := Peer{
		Node: mockWeightedNode{},
	}
	// 使用 NewPeerContext 函数创建一个新的上下文，并将 Peer 实例添加到上下文中
	ctx := NewPeerContext(context.Background(), &p)
	// 使用 FromPeerContext 函数从上下文中获取 Peer 实例
	p2, ok := FromPeerContext(ctx)
	// 检查是否成功获取到 Peer 实例，并且实例中的节点是否不为 nil
	if !ok || p2.Node == nil {
		// 如果没有获取到 Peer 实例或者实例中的节点为 nil，则记录错误信息并终止测试
		t.Fatalf("no peer found!")
	}
}

// TestNotPeer 测试在没有 Peer 实例的上下文中，FromPeerContext 函数的行为
func TestNotPeer(t *testing.T) {
	// 使用 FromPeerContext 函数从一个没有 Peer 实例的上下文中获取 Peer 实例
	_, ok := FromPeerContext(context.Background())
	// 检查是否没有获取到 Peer 实例
	if ok {
		// 如果获取到了 Peer 实例，则记录错误信息并终止测试
		t.Fatalf("test no peer found peer!")
	}
}
