package rumors

import (
	"github.com/gregunz/Peerster/models/packets"
	"math/rand"
	"sync"
)

type VectorClock struct {
	myOrigin string
	handlers map[string]*rumorHandler
	mux      sync.Mutex
}

func NewVectorClock(myOrigin string) *VectorClock {
	handlers := map[string]*rumorHandler{}
	return &VectorClock{
		myOrigin: myOrigin,
		handlers: handlers,
	}
}

func (vectorClock *VectorClock) ToStatusPacket() *packets.StatusPacket {
	vectorClock.mux.Lock()
	defer vectorClock.mux.Unlock()

	want := []packets.PeerStatus{}
	for _, h := range vectorClock.handlers {
		//if h.origin != vectorClock.myOrigin {
		want = append(want, *h.ToPeerStatus())
		//}
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

func (vectorClock *VectorClock) Save(msg *packets.RumorMessage) {
	h := vectorClock.getOrCreateHandler(msg.Origin)
	h.Save(msg)
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
		if !ok {
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
