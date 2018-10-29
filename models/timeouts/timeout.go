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
		cancelChan:  make(chan interface{}),
		triggerChan: make(chan interface{}),
		IsActive:    false,
	}
}

func (timeout *Timeout) SetIfNotActive(d time.Duration, callback func()) {
	timeout.mux.Lock()
	defer timeout.mux.Unlock()

	if !timeout.IsActive {
		timeout.IsActive = true
		go func() {
			select {
			case <-timeout.cancelChan:
				// do nothing
			case <-timeout.triggerChan:
				go callback()
			case <-time.After(d):
				timeout.mux.Lock()
				if timeout.IsActive {
					go callback()
					timeout.IsActive = false
				}
				timeout.mux.Unlock()
			}
		}()
	}
}

func (timeout *Timeout) Cancel() {
	timeout.mux.Lock()
	defer timeout.mux.Unlock()

	if timeout.IsActive {
		timeout.cancelChan <- nil
		timeout.IsActive = false
	}
}

func (timeout *Timeout) Trigger() {
	timeout.mux.Lock()
	defer timeout.mux.Unlock()

	if timeout.IsActive {
		timeout.triggerChan <- nil
		timeout.IsActive = false
	}
}
