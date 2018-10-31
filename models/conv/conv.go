package conv

import (
	"github.com/gregunz/Peerster/models/packets/packets_gossiper"
	"sync"
)

type Conversations struct {
	handlers       map[string]*convHandler
	privateMsgChan PrivateMsgChan
	mux            sync.Mutex
}

func NewConversations(privateMsgChan PrivateMsgChan) *Conversations {
	return &Conversations{
		handlers:       map[string]*convHandler{},
		privateMsgChan: privateMsgChan,
	}
}

func (conv *Conversations) getOrCreateHandler(origin string) *convHandler {
	h, ok := conv.handlers[origin]
	if !ok {
		h = newConvHandler(origin, conv.privateMsgChan)
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
