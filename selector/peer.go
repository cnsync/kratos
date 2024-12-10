package selector

import (
	"context"
)

// peerKey 是一个用于在上下文中存储和检索对等节点信息的键。
type peerKey struct{}

// Peer 包含 RPC 的对等信息，如地址和认证信息。
type Peer struct {
	// node 是对等节点。
	Node Node
}

// NewPeerContext 创建一个新的上下文，并附加对等信息。
func NewPeerContext(ctx context.Context, p *Peer) context.Context {
	return context.WithValue(ctx, peerKey{}, p)
}

// FromPeerContext 从上下文中返回对等信息，如果存在的话。
func FromPeerContext(ctx context.Context) (p *Peer, ok bool) {
	p, ok = ctx.Value(peerKey{}).(*Peer)
	return
}
