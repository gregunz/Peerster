package models

import (
	"fmt"
	"github.com/gregunz/Peerster/common"
	"sync"
	"time"
)

type Peer struct {
	Addr     *Address
	timeout  *RumorTimeout
	mux      sync.Mutex
}

func NewPeer(s string) *Peer {
	return &Peer{
		Addr:     NewAddress(s),
		timeout:  nil,
	}
}

func (p *Peer) ID() string {
	return p.Addr.ToIpPort()
}

func (p *Peer) Equals(other *Peer) bool {
	return p.Addr.Equals(other.Addr)
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