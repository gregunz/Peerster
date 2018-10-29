package clock

import (
	"github.com/gregunz/Peerster/models/packets/packets_gossiper"
	"sync"
)

type rumorHandler struct {
	origin   string
	latestID uint32
	messages map[uint32]*packets_gossiper.RumorMessage
	mux      sync.Mutex
}

func NewRumorHandler(origin string) *rumorHandler {
	return &rumorHandler{
		origin:   origin,
		latestID: 0,
		messages: map[uint32]*packets_gossiper.RumorMessage{},
	}
}

func (handler *rumorHandler) save(msg *packets_gossiper.RumorMessage) bool {
	_, ok := handler.messages[msg.ID]
	if !ok {
		handler.messages[msg.ID] = msg
		for {
			_, ok := handler.messages[handler.latestID+1]
			if ok {
				handler.latestID += 1
			} else {
				break
			}
		}
	} else {
		// discarding overwriting message
	}
	return !ok
}

func (handler *rumorHandler) Save(msg *packets_gossiper.RumorMessage) bool {
	handler.mux.Lock()
	defer handler.mux.Unlock()

	return handler.save(msg)
}

func (handler *rumorHandler) CreateNextMessage(content string) *packets_gossiper.RumorMessage {
	handler.mux.Lock()
	defer handler.mux.Unlock()

	handler.latestID += 1
	msg := &packets_gossiper.RumorMessage{
		Origin: handler.origin,
		ID:     handler.latestID,
		Text:   content,
	}
	//handler.messages[handler.latestID] = msg
	return msg
}

func (handler *rumorHandler) ToPeerStatus() *packets_gossiper.PeerStatus {
	handler.mux.Lock()
	defer handler.mux.Unlock()

	return &packets_gossiper.PeerStatus{
		Identifier: handler.origin,
		NextID:     handler.latestID + 1,
	}
}
