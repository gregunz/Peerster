package conv

import (
	"github.com/gregunz/Peerster/models/packets/packets_gossiper"
	"sync"
)

type Conversations struct {
	handlers       map[string]*convHandler
	PrivateMsgChan PrivateMsgChan
	mux            sync.Mutex
}

func NewConversations(activateChan bool) *Conversations {
	return &Conversations{
		handlers:       map[string]*convHandler{},
		PrivateMsgChan: NewPrivateChan(activateChan),
	}
}

func (conv *Conversations) getOrCreateHandler(origin string) *convHandler {
	h, ok := conv.handlers[origin]
	if !ok {
		h = newConvHandler(origin, conv.PrivateMsgChan)
		conv.handlers[origin] = h
	}
	return h
}

func (conv *Conversations) GetOrCreateHandler(origin string) *convHandler {
	conv.mux.Lock()
	defer conv.mux.Unlock()

	return conv.getOrCreateHandler(origin)
}

func (conv *Conversations) GetAllOf(origin string) []*packets_gossiper.PrivateMessage {
	conv.mux.Lock()
	defer conv.mux.Unlock()

	return conv.getOrCreateHandler(origin).GetAll()
}

func (conv *Conversations) GetAll() []*packets_gossiper.PrivateMessage {
	conv.mux.Lock()
	defer conv.mux.Unlock()

	msgs := []*packets_gossiper.PrivateMessage{}
	for _, h := range conv.handlers {
		for _, msg := range h.GetAll() {
			msgs = append(msgs, msg)
		}
	}
	return msgs
}
