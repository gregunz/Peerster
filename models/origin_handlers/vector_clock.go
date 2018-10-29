package origin_handlers

import (
	"github.com/gregunz/Peerster/models/packets/packets_gossiper"
	"math/rand"
	"sync"
)

type VectorClock interface {
	GetLatestRumorChan() chan *packets_gossiper.RumorMessage

	ToStatusPacket() *packets_gossiper.StatusPacket
	GetOrCreateHandler(origin string) *saveRumorHandler
	Save(msg *packets_gossiper.RumorMessage)
	Compare(statusMap map[string]uint32) (*packets_gossiper.RumorMessage, bool)
}

type ProtoVectorClock struct {
	originsHandler  *OriginsHandler
	latestRumorChan chan *packets_gossiper.RumorMessage
	mux             sync.Mutex
}

func (vectorClock *ProtoVectorClock) GetLatestRumorChan() chan *packets_gossiper.RumorMessage {
	return vectorClock.latestRumorChan
}

func (vectorClock *ProtoVectorClock) ToStatusPacket() *packets_gossiper.StatusPacket {
	vectorClock.mux.Lock()
	defer vectorClock.mux.Unlock()

	want := []packets_gossiper.PeerStatus{}
	for _, h := range vectorClock.originsHandler.Map {
		want = append(want, *h.saveRumor.ToPeerStatus())
	}
	return &packets_gossiper.StatusPacket{
		Want: want,
	}
}

func (vectorClock *ProtoVectorClock) getOrCreateHandler(origin string) *saveRumorHandler {
	return vectorClock.originsHandler.GetOrCreateHandler(origin).saveRumor
}

func (vectorClock *ProtoVectorClock) GetOrCreateHandler(origin string) *saveRumorHandler {
	vectorClock.mux.Lock()
	defer vectorClock.mux.Unlock()
	return vectorClock.getOrCreateHandler(origin)
}

func (vectorClock *ProtoVectorClock) Save(msg *packets_gossiper.RumorMessage) {
	vectorClock.mux.Lock()
	defer vectorClock.mux.Unlock()

	h := vectorClock.getOrCreateHandler(msg.Origin)
	if h.Save(msg) { // if it is a new message
		go func() {
			vectorClock.latestRumorChan <- msg
		}()
	}
}

func (vectorClock *ProtoVectorClock) Compare(statusMap map[string]uint32) (*packets_gossiper.RumorMessage, bool) {
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
	for _, handler := range vectorClock.originsHandler.Map {
		nextID, ok := statusMap[handler.saveRumor.origin]
		if !ok && handler.saveRumor.latestID > 0 {
			msgToSend = append(msgToSend, handler.saveRumor.messages[1])
		} else if handler.saveRumor.latestID >= nextID {
			msgToSend = append(msgToSend, handler.saveRumor.messages[nextID])
		}
	}

	if len(msgToSend) > 0 {
		return msgToSend[rand.Int()%len(msgToSend)], remoteHasMsg
	}
	return nil, remoteHasMsg
}
