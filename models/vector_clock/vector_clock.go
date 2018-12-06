package vector_clock

import (
	"github.com/gregunz/Peerster/models/packets/packets_gossiper"
	"math/rand"
	"sync"
)

type VectorClock struct {
	handlers  map[string]*vectorClockHandler
	RumorChan RumorChan
	mux       sync.Mutex
}

func NewVectorClock(activateChan bool) *VectorClock {
	return &VectorClock{
		handlers:  map[string]*vectorClockHandler{},
		RumorChan: NewRumorChan(activateChan),
	}
}

func (vector *VectorClock) getOrCreateHandler(origin string) *vectorClockHandler {
	h, ok := vector.handlers[origin]
	if !ok {
		h = newVectorClockHandler(origin, vector.RumorChan)
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
		want = append(want, *h.ToPeerStatus()) // safe access (ToPeerStatus does a lock)
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
		h.mux.Lock() // needs to lock -> writing
		if h.latestID < nextID-1 {
			remoteHasMsg = true
		}
		h.mux.Unlock()
	}
	for _, h := range vector.handlers {
		h.mux.Lock() // needs to lock -> reading
		nextID, ok := statusMap[h.origin]
		if !ok && h.latestID > 0 {
			msgToSend = append(msgToSend, h.messages[1])
		} else if h.latestID >= nextID {
			msgToSend = append(msgToSend, h.messages[nextID])
		}
		h.mux.Unlock()
	}

	if len(msgToSend) > 0 {
		return msgToSend[rand.Int()%len(msgToSend)], remoteHasMsg
	}
	return nil, remoteHasMsg
}
