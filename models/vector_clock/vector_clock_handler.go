package vector_clock

import (
	"github.com/gregunz/Peerster/models/packets/packets_gossiper"
	"sync"
)

type vectorClockHandler struct {
	origin    string
	latestID  uint32
	messages  map[uint32]*packets_gossiper.RumorMessage
	rumorChan RumorChan
	mux       sync.Mutex
}

func newVectorClockHandler(origin string, rumorChan RumorChan) *vectorClockHandler {
	return &vectorClockHandler{
		origin:    origin,
		latestID:  0,
		messages:  map[uint32]*packets_gossiper.RumorMessage{},
		rumorChan: rumorChan,
	}
}

func (handler *vectorClockHandler) save(msg *packets_gossiper.RumorMessage) bool {
	_, ok := handler.messages[msg.ID]
	if !ok {
		handler.messages[msg.ID] = msg
		handler.AddToLatest(msg)

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

func (handler *vectorClockHandler) Save(msg *packets_gossiper.RumorMessage) bool {
	handler.mux.Lock()
	defer handler.mux.Unlock()

	return handler.save(msg)
}

func (handler *vectorClockHandler) CreateAndSaveNextMessage(content string) *packets_gossiper.RumorMessage {
	handler.mux.Lock()
	defer handler.mux.Unlock()

	handler.latestID += 1
	msg := &packets_gossiper.RumorMessage{
		Origin: handler.origin,
		ID:     handler.latestID,
		Text:   content,
	}
	handler.save(msg)

	return msg
}

func (handler *vectorClockHandler) ToPeerStatus() *packets_gossiper.PeerStatus {
	handler.mux.Lock()
	defer handler.mux.Unlock()

	return &packets_gossiper.PeerStatus{
		Identifier: handler.origin,
		NextID:     handler.latestID + 1,
	}
}

func (handler *vectorClockHandler) AddToLatest(msg *packets_gossiper.RumorMessage) {
	go func() {
		if msg.Text != "" {
			handler.rumorChan.AddRumor(msg)
		}
	}()
}
