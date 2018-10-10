package timeouts

import (
	"fmt"
	"sync"
	"time"
)

type RumorTimeout struct {
	ticker    *time.Ticker
	duration  time.Duration
	callback  func()
	triggered bool
	mux       sync.Mutex
}

func NewRumorTimeout(d time.Duration, callback func()) *RumorTimeout {
	timeout := &RumorTimeout{
		duration:  d,
		callback:  callback,
		triggered: false,
	}
	tickerCallback := func() {
		timeout.mux.Lock()
		defer timeout.mux.Unlock()
		fmt.Println("called 2")

		if !timeout.triggered {
			fmt.Println("called 3")

			timeout.triggered = true
			callback()
		}
	}
	timeout.ticker = newTicker(d, tickerCallback)
	return timeout
}

func newTicker(d time.Duration, callback func()) *time.Ticker {
	ticker := time.NewTicker(d)
	go func() {
		for range ticker.C {
			func() {
				fmt.Println("called 1")
				ticker.Stop()
				callback()
			}()
		}
	}()
	return ticker
}

func (timeout *RumorTimeout) Trigger() {
	timeout.mux.Lock()
	defer timeout.mux.Unlock()

	go timeout.callback() //without "go" routine, the lock is not released
}

/*
func (timeout *RumorTimeout) ResetIfTriggered() {
	timeout.mux.Lock()
	oldTimeout := *timeout
	defer oldTimeout.mux.Unlock()

	if oldTimeout.triggered {
		oldTimeout.ticker.Stop()
		oldTimeout.triggered = false
		*timeout = *NewRumorTimeout(timeout.duration, timeout.callback)
	}
}
*/

func (timeout *RumorTimeout) Stop() {
	timeout.mux.Lock()
	defer timeout.mux.Unlock()

	timeout.ticker.Stop()
}
