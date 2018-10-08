package models

import (
	"fmt"
	"github.com/gregunz/Peerster/common"
	"sync"
	"time"
)

type Peer struct {
	Addr     *Address
	LatestID uint32
	Rumors   map[uint32]*RumorMessage
	timeout  *RumorTimeout
	mux      sync.Mutex
}

func NewPeer(s string) *Peer {
	return &Peer{
		Addr:     NewAddress(s),
		LatestID: 0,
		Rumors:   map[uint32]*RumorMessage{},
		timeout:  nil,
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
		NextID:     p.LatestID + 1,
	}
}

func (p *Peer) SaveRumor(msg *RumorMessage) {
	p.mux.Lock()
	defer p.mux.Unlock()

	/* THIS VERSION DISCARDS PREVIOUS & LATER MESSAGES
	if msg.ID == p.LatestID + 1 {
		p.Rumors[msg.ID] = msg
		p.LatestID += 1
	}
	//*/

	//* THIS VERSION DISCARDS PREVIOUS BUT ADDS LATER MESSAGES
	_, ok := p.Rumors[msg.ID]
	if ok {
		// message already stored -> discard
	} else {
		p.Rumors[msg.ID] = msg
		if p.LatestID + 1 == msg.ID {
			p.LatestID += 1
			for {
				_, ok := p.Rumors[p.LatestID + 1]
				if ok {
					p.LatestID += 1
				} else {
					break
				}
			}
		}
		// we update latest id knowing that we might have added later than latestID messages hence -> loop
	}
	//*/
}

func (p *Peer) ID() string {
	return p.Addr.ToIpPort()
}

func (p *Peer) SetTimeout(d time.Duration, callback func()) {
	p.mux.Lock()
	defer p.mux.Unlock()

	if p.timeout != nil {
		common.HandleError(fmt.Errorf("only one timeout per peer handled, discarding new timeout"))
	} else {
		p.timeout = NewRumorTimeout(d, callback)
	}
}

func (p *Peer) TriggerTimeout() {
	if p.timeout != nil {
		p.timeout.Trigger()
	} else {
		//common.HandleError(fmt.Errorf("Trigger called on nil timeout"))
	}
}

func (p *Peer) ResetTimeout() {
	if p.timeout != nil {
		p.timeout.Reset()
	} else {
		//common.HandleError(fmt.Errorf("Reset called on nil timeout"))
	}
}

func (p *Peer) StopTimeout() {
	if p.timeout != nil {
		p.timeout.Stop()
	} else {
		//common.HandleError(fmt.Errorf("Stop called on nil timeout"))
	}
}
