package vector_clock

import (
	"github.com/gregunz/Peerster/models/packets/packets_gossiper"
	"math/rand"
	"sync"
)

type vectorClock struct {
	handlers  map[string]*vectorClockHandler
	rumorChan RumorChan
	mux       sync.Mutex
}

type VectorClock interface {
	GetOrCreateHandler(origin string) *vectorClockHandler
	ToStatusPacket() *packets_gossiper.StatusPacket
	Compare(statusMap map[string]uint32) (*packets_gossiper.RumorMessage, bool)
}

func NewVectorClock(rumorChan RumorChan) *vectorClock {
	return &vectorClock{
		handlers:  map[string]*vectorClockHandler{},
		rumorChan: rumorChan,
	}
}

func (vector *vectorClock) getOrCreateHandler(origin string) *vectorClockHandler {
	h, ok := vector.handlers[origin]
	if !ok {
		h = newVectorClockHandler(origin, vector.rumorChan)
		vector.handlers[origin] = h
	}
	return h
}

func (vector *vectorClock) GetOrCreateHandler(origin string) *vectorClockHandler {
	vector.mux.Lock()
	defer vector.mux.Unlock()

	return vector.getOrCreateHandler(origin)
}

func (vector *vectorClock) ToStatusPacket() *packets_gossiper.StatusPacket {
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

func (vector *vectorClock) Compare(statusMap map[string]uint32) (*packets_gossiper.RumorMessage, bool) {
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
