package timeouts

import (
	"sync"
	"time"
)

type Timeout struct {
	cancelChan  chan interface{}
	triggerChan chan interface{}
	active      bool
	mux         sync.Mutex
}

func NewTimeout() *Timeout {
	return &Timeout{
		cancelChan:  make(chan interface{}, 1),
		triggerChan: make(chan interface{}, 1),
		active:      false,
	}
}

func (timeout *Timeout) set(d time.Duration, callback func()) {

	if !timeout.active {
		timeout.active = true
		go func() {
			defer timeout.mux.Unlock()
			select {
			case <-timeout.triggerChan:
				timeout.mux.Lock()
				callback()
			case <-timeout.cancelChan:
				timeout.mux.Lock()
			case <-time.After(d):
				timeout.mux.Lock()
				callback()
			}
		}()
	}
}

func (timeout *Timeout) Set(d time.Duration, callback func()) {
	timeout.mux.Lock()
	defer timeout.mux.Unlock()
	timeout.set(d, callback)
}

func (timeout *Timeout) cancel() {
	if timeout.active {
		timeout.cancelChan <- nil
		timeout.active = false
	}
}

func (timeout *Timeout) Cancel() {
	timeout.mux.Lock()
	defer timeout.mux.Unlock()
	timeout.cancel()
}

func (timeout *Timeout) trigger() {
	if timeout.active {
		timeout.triggerChan <- nil
		timeout.active = false
	}
}

func (timeout *Timeout) Trigger() {
	timeout.mux.Lock()
	defer timeout.mux.Unlock()
	timeout.trigger()
}
