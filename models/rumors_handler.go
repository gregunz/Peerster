package models

import (
	"math/rand"
	"sync"
)

type RumorHandlers struct {
	handlers map[string]*rumorHandler
	mux      sync.Mutex
}

func NewRumorsHandler() *RumorHandlers {
	handlers := map[string]*rumorHandler{}
	return &RumorHandlers{
		handlers: handlers,
	}
}


func (handlers *RumorHandlers) ToStatusPacket() *StatusPacket {
	handlers.mux.Lock()
	defer handlers.mux.Unlock()

	want := []PeerStatus{}
	for _, h := range handlers.handlers {
		want = append(want, *h.ToPeerStatus())
	}
	return &StatusPacket{
		Want: want,
	}
}

func (handlers *RumorHandlers) getOrCreateHandler(origin string) *rumorHandler {
	h, ok := handlers.handlers[origin]
	if !ok {
		h = NewRumorHandler(origin)
		handlers.handlers[origin] = h
	}
	return h
}

func (handlers *RumorHandlers) GetOrCreateHandler(origin string) *rumorHandler {
	handlers.mux.Lock()
	defer handlers.mux.Unlock()
	return handlers.getOrCreateHandler(origin)
}

func (handlers *RumorHandlers) Save(msg *RumorMessage) {
	h := handlers.getOrCreateHandler(msg.Origin)
	h.Save(msg)
}

func (handlers *RumorHandlers) Compare(want []PeerStatus) (*RumorMessage, bool) {
	handlers.mux.Lock()
	defer handlers.mux.Unlock()

	msgToSend := []*RumorMessage{}
	remoteHasMsg := false

	for _, ps := range want {
		handler := handlers.getOrCreateHandler(ps.Identifier)
		if handler.latestID >= ps.NextID {
			msgToSend = append(msgToSend, handler.messages[ps.NextID])
		} else if handler.latestID < ps.NextID - 1 {
			remoteHasMsg = true
		}
	}

	if len(msgToSend) > 0 {
		return msgToSend[rand.Int() % len(msgToSend)], remoteHasMsg
	}
	return nil, remoteHasMsg
}