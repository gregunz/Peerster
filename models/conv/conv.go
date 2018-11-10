package conv

import (
	"github.com/gregunz/Peerster/models/packets/packets_gossiper"
	"sync"
)

type conversations struct {
	handlers       map[string]*convHandler
	privateMsgChan PrivateMsgChan
	mux            sync.Mutex
}

type Conversation interface {
	GetOrCreateHandler(origin string) *convHandler
	GetAll() []*packets_gossiper.PrivateMessage
}

func NewConversations(privateMsgChan PrivateMsgChan) *conversations {
	return &conversations{
		handlers:       map[string]*convHandler{},
		privateMsgChan: privateMsgChan,
	}
}

func (conv *conversations) getOrCreateHandler(origin string) *convHandler {
	h, ok := conv.handlers[origin]
	if !ok {
		h = newConvHandler(origin, conv.privateMsgChan)
		conv.handlers[origin] = h
	}
	return h
}

func (conv *conversations) GetOrCreateHandler(origin string) *convHandler {
	conv.mux.Lock()
	defer conv.mux.Unlock()

	return conv.getOrCreateHandler(origin)
}

func (conv *conversations) GetAllOf(origin string) []*packets_gossiper.PrivateMessage {
	conv.mux.Lock()
	defer conv.mux.Unlock()

	return conv.getOrCreateHandler(origin).GetAll()
}

func (conv *conversations) GetAll() []*packets_gossiper.PrivateMessage {
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
