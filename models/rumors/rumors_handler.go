package rumors

import (
	"github.com/gregunz/Peerster/models/packets"
	"math/rand"
	"sync"
)

type RumorHandlers struct {
	myOrigin string
	handlers map[string]*rumorHandler
	mux      sync.Mutex
}

func NewRumorsHandler(myOrigin string) *RumorHandlers {
	handlers := map[string]*rumorHandler{}
	return &RumorHandlers{
		myOrigin: myOrigin,
		handlers: handlers,
	}
}

func (handlers *RumorHandlers) ToStatusPacket() *packets.StatusPacket {
	handlers.mux.Lock()
	defer handlers.mux.Unlock()

	want := []packets.PeerStatus{}
	for _, h := range handlers.handlers {
		if h.origin != handlers.myOrigin {
			want = append(want, *h.ToPeerStatus())
		}
	}
	return &packets.StatusPacket{
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

func (handlers *RumorHandlers) Save(msg *packets.RumorMessage) {
	h := handlers.getOrCreateHandler(msg.Origin)
	h.Save(msg)
}

func (handlers *RumorHandlers) Compare(want []packets.PeerStatus) (*packets.RumorMessage, bool) {
	handlers.mux.Lock()
	defer handlers.mux.Unlock()

	msgToSend := []*packets.RumorMessage{}
	remoteHasMsg := false

	for _, ps := range want {
		handler := handlers.getOrCreateHandler(ps.Identifier)
		if handler.latestID >= ps.NextID {
			msgToSend = append(msgToSend, handler.messages[ps.NextID])
		} else if handler.latestID < ps.NextID-1 {
			remoteHasMsg = true
		}
	}

	if len(msgToSend) > 0 {
		return msgToSend[rand.Int()%len(msgToSend)], remoteHasMsg
	}
	return nil, remoteHasMsg
}
