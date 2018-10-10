package rumors

import (
	"github.com/gregunz/Peerster/models/packets"
	"sync"
)

type rumorHandler struct {
	origin   string
	latestID uint32
	messages map[uint32]*packets.RumorMessage
	mux      sync.Mutex
}

func NewRumorHandler(origin string) *rumorHandler {
	return &rumorHandler{
		origin:   origin,
		latestID: 0,
		messages: map[uint32]*packets.RumorMessage{},
	}
}

func (handler *rumorHandler) save(msg *packets.RumorMessage) {
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
		// discading overwriting message
	}
}

func (handler *rumorHandler) Save(msg *packets.RumorMessage) {
	handler.mux.Lock()
	defer handler.mux.Unlock()

	handler.save(msg)
}

func (handler *rumorHandler) NextMessage(content string) *packets.RumorMessage {
	handler.mux.Lock()
	defer handler.mux.Unlock()

	handler.latestID += 1
	msg := &packets.RumorMessage{
		Origin: handler.origin,
		ID:     handler.latestID,
		Text:   content,
	}
	handler.messages[handler.latestID] = msg
	return msg
}

func (handler *rumorHandler) ToPeerStatus() *packets.PeerStatus {
	handler.mux.Lock()
	defer handler.mux.Unlock()

	return &packets.PeerStatus{
		Identifier: handler.origin,
		NextID:     handler.latestID + 1,
	}
}

/*
func (handler *RumorHandler) Get(id uint32) *RumorMessage {
	handler.mux.Lock()
	defer handler.mux.Unlock()

	msg, ok := handler.messages[id]
	if !ok {
		return nil
	}
	return msg
}
*/
