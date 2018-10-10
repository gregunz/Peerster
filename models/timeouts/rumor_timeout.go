package timeouts

import (
	"sync"
	"time"
)

type RumorTimeout struct {
	ticker   *time.Ticker
	duration time.Duration
	callback func()
	mux      sync.Mutex
}

func NewRumorTimeout(d time.Duration, callback func()) *RumorTimeout {
	timeout := &RumorTimeout{
		duration: d,
		ticker:   newTicker(d, callback),
		callback: callback,
	}
	return timeout
}

func newTicker(d time.Duration, callback func()) *time.Ticker {
	ticker := time.NewTicker(d)
	go func() {
		for range ticker.C {
			func() {
				ticker.Stop()
				go callback()
			}()
		}
	}()
	return ticker
}

func (timeout *RumorTimeout) Trigger() {
	timeout.mux.Lock()
	defer timeout.mux.Unlock()

	go timeout.callback()
	timeout.ticker.Stop()
}

func (timeout *RumorTimeout) Reset() {
	timeout.mux.Lock()
	defer timeout.mux.Unlock()

	timeout.ticker.Stop()
	timeout.ticker = newTicker(timeout.duration, timeout.callback)
}

func (timeout *RumorTimeout) Stop() {
	timeout.mux.Lock()
	defer timeout.mux.Unlock()

	timeout.ticker.Stop()
}
