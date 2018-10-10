package peers

import (
	"github.com/gregunz/Peerster/models/timeouts"
	"sync"
)

type Peer struct {
	Addr    *Address
	Timeout *timeouts.Timeout
	mux     sync.Mutex
}

func NewPeer(s string) *Peer {
	p := &Peer{
		Addr:    NewAddress(s),
		Timeout: timeouts.NewTimeout(),
	}
	if p.Addr.IsEmpty() {
		return nil
	}
	return p
}

func (p *Peer) ID() string {
	return p.Addr.ToIpPort()
}

func (p *Peer) Equals(other *Peer) bool {
	return p.Addr.Equals(other.Addr)
}
