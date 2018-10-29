package routing

import (
	"github.com/gregunz/Peerster/models/packets/packets_gossiper"
	"github.com/gregunz/Peerster/models/peers"
	"sync"
)

type Table struct {
	myOrigin string
	handlers map[string]*tableHandler
	mux      sync.Mutex
}

func NewTable(myOrigin string) *Table {
	return &Table{
		myOrigin: myOrigin,
		handlers: map[string]*tableHandler{},
	}
}

func (handler *Table) getOrCreateHandler(origin string) *tableHandler {
	h, ok := handler.handlers[origin]
	if !ok {
		h = newRoutingTableHandler(origin)
		handler.handlers[origin] = h
	}
	return h
}

func (handler *Table) GetOrCreateHandler(origin string) *tableHandler {
	handler.mux.Lock()
	defer handler.mux.Unlock()

	return handler.getOrCreateHandler(origin)
}

func (handler *Table) AckRumor(rumor *packets_gossiper.RumorMessage, fromPeer *peers.Peer) {
	if rumor.Origin != handler.myOrigin {
		handler.getOrCreateHandler(rumor.Origin).AckRumor(rumor, fromPeer)
	}
}

func (handler *Table) Get(origin string) *peers.Peer {
	return handler.getOrCreateHandler(origin).peer
}
