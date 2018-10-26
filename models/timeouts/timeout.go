package timeouts

import (
	"sync"
	"time"
)

type Timeout struct {
	cancelChan  chan interface{}
	triggerChan chan interface{}
	IsActive    bool
	mux         sync.Mutex
}

func NewTimeout() *Timeout {
	return &Timeout{
		cancelChan:  make(chan interface{}, 1),
		triggerChan: make(chan interface{}, 1),
		IsActive:    false,
	}
}

func (timeout *Timeout) set(d time.Duration, callback func()) {
	if !timeout.IsActive {
		timeout.IsActive = true
		go func() {
			defer timeout.mux.Unlock()
			select {
			case <-timeout.cancelChan:
				timeout.mux.Lock()
				// do nothing
			case <-timeout.triggerChan:
				timeout.mux.Lock()
				callback()
			case <-time.After(d):
				timeout.mux.Lock()
				callback()
			}
		}()
	}
}

func (timeout *Timeout) SetIfNotActive(d time.Duration, callback func()) {
	timeout.mux.Lock()
	defer timeout.mux.Unlock()
	timeout.set(d, callback)
}

func (timeout *Timeout) cancel() {
	if timeout.IsActive {
		timeout.cancelChan <- nil
		timeout.IsActive = false
	}
}

func (timeout *Timeout) Cancel() {
	timeout.mux.Lock()
	defer timeout.mux.Unlock()
	timeout.cancel()
}

func (timeout *Timeout) trigger() {
	if timeout.IsActive {
		timeout.triggerChan <- nil
		timeout.IsActive = false
	}
}

func (timeout *Timeout) Trigger() {
	timeout.mux.Lock()
	defer timeout.mux.Unlock()
	timeout.trigger()
}
