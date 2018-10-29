package clock

import (
	"github.com/gregunz/Peerster/models/packets/packets_gossiper"
	"math/rand"
	"sync"
)

type VectorClock struct {
	handlers        map[string]*rumorHandler
	LatestRumorChan chan *packets_gossiper.RumorMessage
	mux             sync.Mutex
}

func NewVectorClock(myOrigin string) *VectorClock {
	handlers := map[string]*rumorHandler{}
	handlers[myOrigin] = NewRumorHandler(myOrigin)
	return &VectorClock{
		handlers:        handlers,
		LatestRumorChan: make(chan *packets_gossiper.RumorMessage, 1),
	}
}

func (vectorClock *VectorClock) ToStatusPacket() *packets_gossiper.StatusPacket {
	vectorClock.mux.Lock()
	defer vectorClock.mux.Unlock()

	want := []packets_gossiper.PeerStatus{}
	for _, h := range vectorClock.handlers {
		want = append(want, *h.ToPeerStatus())
	}
	return &packets_gossiper.StatusPacket{
		Want: want,
	}
}

func (vectorClock *VectorClock) getOrCreateHandler(origin string) *rumorHandler {
	h, ok := vectorClock.handlers[origin]
	if !ok {
		h = NewRumorHandler(origin)
		vectorClock.handlers[origin] = h
	}
	return h
}

func (vectorClock *VectorClock) GetOrCreateHandler(origin string) *rumorHandler {
	vectorClock.mux.Lock()
	defer vectorClock.mux.Unlock()
	return vectorClock.getOrCreateHandler(origin)
}

func (vectorClock *VectorClock) Save(msg *packets_gossiper.RumorMessage) {
	vectorClock.mux.Lock()
	defer vectorClock.mux.Unlock()

	h := vectorClock.getOrCreateHandler(msg.Origin)
	if h.Save(msg) { // if it is a new message
		go func() {
			vectorClock.LatestRumorChan <- msg
		}()
	}
}

func (vectorClock *VectorClock) Compare(statusMap map[string]uint32) (*packets_gossiper.RumorMessage, bool) {
	vectorClock.mux.Lock()
	defer vectorClock.mux.Unlock()

	msgToSend := []*packets_gossiper.RumorMessage{}
	remoteHasMsg := false

	for origin, nextID := range statusMap {
		h := vectorClock.getOrCreateHandler(origin)
		if h.latestID < nextID-1 {
			remoteHasMsg = true
		}
	}
	for _, handler := range vectorClock.handlers {
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
