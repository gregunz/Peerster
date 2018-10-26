package clock

import (
	"github.com/gregunz/Peerster/models/packets"
	"math/rand"
	"sync"
)

type VectorClock struct {
	handlers     map[string]*rumorHandler
	latestRumors []*packets.RumorMessage
	mux          sync.Mutex
}

func NewVectorClock(myOrigin string) *VectorClock {
	handlers := map[string]*rumorHandler{}
	handlers[myOrigin] = NewRumorHandler(myOrigin)
	return &VectorClock{
		handlers:     handlers,
		latestRumors: []*packets.RumorMessage{},
	}
}

func (vectorClock *VectorClock) ToStatusPacket() *packets.StatusPacket {
	vectorClock.mux.Lock()
	defer vectorClock.mux.Unlock()

	want := []packets.PeerStatus{}
	for _, h := range vectorClock.handlers {
		want = append(want, *h.ToPeerStatus())
	}
	return &packets.StatusPacket{
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

func (vectorClock *VectorClock) Save(msg *packets.RumorMessage) bool {
	vectorClock.mux.Lock()
	defer vectorClock.mux.Unlock()

	h := vectorClock.getOrCreateHandler(msg.Origin)
	return h.Save(msg)
}

func (vectorClock *VectorClock) SaveLatest(msg *packets.RumorMessage) {
	vectorClock.mux.Lock()
	defer vectorClock.mux.Unlock()

	vectorClock.latestRumors = append(vectorClock.latestRumors, msg)
}

func (vectorClock *VectorClock) GetAllMessages() []*packets.RumorMessage {
	vectorClock.mux.Lock()
	defer vectorClock.mux.Unlock()

	/*
		rumorsCopy := []*packets.RumorMessage{}
		for _, r := range vectorClock.latestRumors {
			rumorsCopy = append(rumorsCopy, r)
		}

		// resetting the list of all messages
		vectorClock.latestRumors = []*packets.RumorMessage{}

		return rumorsCopy
	*/
	return vectorClock.latestRumors
}

func (vectorClock *VectorClock) Compare(statusMap map[string]uint32) (*packets.RumorMessage, bool) {
	vectorClock.mux.Lock()
	defer vectorClock.mux.Unlock()

	msgToSend := []*packets.RumorMessage{}
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
