package clients

import (
	"github.com/gregunz/Peerster/common"
	"sync"
)

type Map interface {
	Add(w Writer)
	Get(w Writer) Client
	Remove(w Writer)
	Iterate(func(Writer, Client))
}

type list struct {
	clients map[Writer]Client
	mux     sync.RWMutex
}

func NewList() Map {
	return &list{
		clients: map[Writer]Client{},
	}
}

func (l *list) Add(w Writer) {
	l.mux.Lock()
	defer l.mux.Unlock()

	l.clients[w] = NewClient()
}

func (l *list) Get(w Writer) Client {
	l.mux.RLock()
	defer l.mux.RUnlock()

	if c, ok := l.clients[w]; ok {
		return c
	}
	common.HandleAbort("trying to get client that is not stored", nil)
	return nil
}

func (l *list) Remove(w Writer) {
	l.mux.Lock()
	defer l.mux.Unlock()
	delete(l.clients, w)
}

func (l *list) Iterate(callback func(w Writer, c Client)) {
	l.mux.RLock()
	defer l.mux.RUnlock()

	for w, c := range l.clients {
		callback(w, c)
	}
}
