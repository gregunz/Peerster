package peers

import (
	"fmt"
	"github.com/gregunz/Peerster/common"
	"github.com/gregunz/Peerster/models/timeouts"
	"sync"
	"time"
)

type Peer struct {
	Addr    *Address
	timeout *timeouts.RumorTimeout
	mux     sync.Mutex
}

func NewPeer(s string) *Peer {
	p := &Peer{
		Addr:    NewAddress(s),
		timeout: nil,
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

func (p *Peer) SetOrResetTimeout(d time.Duration, callback func()) {
	p.mux.Lock()
	defer p.mux.Unlock()

	if p.timeout != nil {
		common.HandleAbort(fmt.Sprintf("only one timeout per peer handled, discarding new timeout"), nil)
		p.timeout.Stop()
		p.timeout = timeouts.NewRumorTimeout(d, callback)
	} else {
		p.timeout = timeouts.NewRumorTimeout(d, callback)
	}
}

func (p *Peer) StopTimeout() {
	p.mux.Lock()
	defer p.mux.Unlock()

	if p.timeout != nil {
		p.timeout.Stop()
	}
}

func (p *Peer) TriggerTimeout() {
	p.mux.Lock()
	defer p.mux.Unlock()

	if p.timeout != nil {
		p.timeout.Trigger()
	}
}

/*

func (p *Peer) TriggerAndStopTimeout() {
	p.mux.Lock()
	defer p.mux.Unlock()

	if p.timeout != nil {
		p.timeout.TriggerAndStop()
	}
	//	p.timeout = nil
	//} else {
		//common.HandleError(fmt.Errorf("TriggerAndStop called on nil timeout"))
	//}
}

func (p *Peer) ResetTimeout() {
	p.mux.Lock()
	defer p.mux.Unlock()

	if p.timeout != nil {
		p.timeout.ResetIfTriggered()
	} else {
		//common.HandleError(fmt.Errorf("ResetIfTriggered called on nil timeout"))
	}
}


*/
