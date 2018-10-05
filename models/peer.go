package models

import (
	"github.com/gregunz/Peerster/utils"
	"sync"
)

type Peer struct {
	Addr     *Address
	LatestID uint32
	Rumors   map[uint32]RumorMessage
	mux      sync.Mutex
}

func NewPeer(s string) *Peer {
	addr := utils.IpPortToUDPAddr(s)
	return &Peer{
		Addr:     &Address{UDPAddr: addr},
		LatestID: 0,
		Rumors:   map[uint32]RumorMessage{},
	}
}

func (p *Peer) SetSequenceNum(num uint32) {
	p.mux.Lock()
	p.LatestID = num
	p.mux.Unlock()
}

func (p *Peer) Equals(other *Peer) bool {
	return p.Addr.Equals(other.Addr)
}

func (p *Peer) ToPeerStatus() *PeerStatus {
	return &PeerStatus{
		Identifier: p.Addr.ToIpPort(),
		NextID:     p.LatestID,
	}
}

func (p *Peer) SaveRumor(msg RumorMessage) {
	p.mux.Lock()
	defer p.mux.Unlock()
	if msg.ID == p.LatestID+1 {
		p.LatestID += 1
	}
	p.Rumors[msg.ID] = msg
}

func (p *Peer) ID() string {
	return p.Addr.ToIpPort()
}
