package peers

type NodeChan interface {
	AddNode(r *Peer)
	GetNode() *Peer
}
