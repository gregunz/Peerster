package peers

import "github.com/gregunz/Peerster/models/updates"

type NodeChan interface {
	Get() *Peer
	Push(peer *Peer)
}

type nodeChan struct {
	updates.Chan
}

func NewNodeChan(activated bool) NodeChan {
	return &nodeChan{Chan: updates.NewChan(activated)}
}

func (ch *nodeChan) Push(s *Peer) {
	ch.Chan.Push(s)
}

func (ch *nodeChan) Get() *Peer {
	s, ok := ch.Chan.Get().(*Peer)
	if !ok {
		return nil
	}
	return s
}
