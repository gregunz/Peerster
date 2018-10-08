package models

import (
	"sync"
	"time"
)

type RumorTimeout struct {
	ticker   *time.Ticker
	duration time.Duration
	callback func()
	mux sync.Mutex
}

func NewRumorTimeout(d time.Duration, callback func()) *RumorTimeout {
	timeout := &RumorTimeout{
		duration: d,
		ticker:   time.NewTicker(d),
		callback: callback,
	}
	go func() {
		for range timeout.ticker.C {
			timeout.mux.Lock()
			timeout.ticker.Stop()
			go callback()
			timeout.mux.Unlock()
		}
	}()
	return timeout
}

func (timeout *RumorTimeout) Trigger() {
	timeout.mux.Lock()
	defer timeout.mux.Unlock()

	timeout.ticker.Stop()
	go timeout.callback()
}

func (timeout *RumorTimeout) Reset() {
	timeout.mux.Lock()
	defer timeout.mux.Unlock()

	timeout.ticker.Stop()
	*timeout = *NewRumorTimeout(timeout.duration, timeout.callback)
}

func (timeout *RumorTimeout) Stop() {
	timeout.mux.Lock()
	defer timeout.mux.Unlock()

	timeout.ticker.Stop()
}

