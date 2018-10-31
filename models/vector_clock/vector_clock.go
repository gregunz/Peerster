package vector_clock

import (
	"github.com/gregunz/Peerster/models/packets/packets_gossiper"
	"math/rand"
	"sync"
)

type VectorClock struct {
	handlers  map[string]*vectorClockHandler
	rumorChan RumorChan
	mux       sync.Mutex
}

func NewVectorClock(rumorChan RumorChan) *VectorClock {
	return &VectorClock{
		handlers:  map[string]*vectorClockHandler{},
		rumorChan: rumorChan,
	}
}

func (vector *VectorClock) getOrCreateHandler(origin string) *vectorClockHandler {
	h, ok := vector.handlers[origin]
	if !ok {
		h = newVectorClockHandler(origin, vector.rumorChan)
		vector.handlers[origin] = h
	}
	return h
}

func (vector *VectorClock) GetOrCreateHandler(origin string) *vectorClockHandler {
	vector.mux.Lock()
	defer vector.mux.Unlock()

	return vector.getOrCreateHandler(origin)
}

func (vector *VectorClock) ToStatusPacket() *packets_gossiper.StatusPacket {
	vector.mux.Lock()
	defer vector.mux.Unlock()

	want := []packets_gossiper.PeerStatus{}
	for _, h := range vector.handlers {
		want = append(want, *h.ToPeerStatus())
	}
	return &packets_gossiper.StatusPacket{
		Want: want,
	}
}

func (vector *VectorClock) Compare(statusMap map[string]uint32) (*packets_gossiper.RumorMessage, bool) {
	vector.mux.Lock()
	defer vector.mux.Unlock()

	msgToSend := []*packets_gossiper.RumorMessage{}
	remoteHasMsg := false

	for origin, nextID := range statusMap {
		h := vector.getOrCreateHandler(origin)
		if h.latestID < nextID-1 {
			remoteHasMsg = true
		}
	}
	for _, handler := range vector.handlers {
		nextID, ok := statusMap[handler.origin]
		if !ok && handler.latestID > 0 {
			msgToSend = append(msgToSend, handler.messages[1])
		} else if handler.latestID >= nextID {
			msgToSend = append(msgToSend, handler.messages[nextID])
		}
	}

	if len(msgToSend) > 0 {
		return msgToSend[rand.Int()%len(msgToSend)], remoteHasMsg
	}
	return nil, remoteHasMsg
}
