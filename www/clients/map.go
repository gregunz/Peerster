package clients

import (
	"github.com/gregunz/Peerster/common"
	"github.com/gregunz/Peerster/logger"
	"sync"
)

type Map struct {
	clients map[Writer]*Client

	sync.RWMutex
}

func NewMap() *Map {
	return &Map{
		clients: map[Writer]*Client{},
	}
}

func (l *Map) AddUnsafe(w Writer) {
	logger.Printlnf("<web-server> new client just arrived")
	l.clients[w] = NewClient()
}

func (l *Map) Add(w Writer) {
	l.Lock()
	defer l.Unlock()
	l.AddUnsafe(w)
}

func (l *Map) Get(w Writer) *Client {
	l.RLock()
	defer l.RUnlock()

	if c, ok := l.clients[w]; ok {
		return c
	}
	common.HandleAbort("trying to get client that is not stored", nil)
	return nil
}

func (l *Map) Remove(w Writer) {
	l.Lock()
	defer l.Unlock()

	logger.Printlnf("<web-server> a client left :'(")
	delete(l.clients, w)
}

func (l *Map) Iterate(callback func(w Writer, c *Client)) {
	l.RLock()
	defer l.RUnlock()

	for w, c := range l.clients {
		callback(w, c)
	}
}
