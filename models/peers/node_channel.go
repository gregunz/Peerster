package peers

type NodeChan interface {
	AddNode(*Peer)
	GetNode() *Peer
}
