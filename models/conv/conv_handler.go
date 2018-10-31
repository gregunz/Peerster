package conv

import (
	"github.com/gregunz/Peerster/models/packets/packets_gossiper"
	"sync"
)

type convHandler struct {
	origin         string
	messages       []*packets_gossiper.PrivateMessage
	privateMsgChan PrivateMsgChan
	mux            sync.Mutex
}

func newConvHandler(origin string, privateMsgChan PrivateMsgChan) *convHandler {
	return &convHandler{
		origin:         origin,
		messages:       []*packets_gossiper.PrivateMessage{},
		privateMsgChan: privateMsgChan,
	}
}

func (handler *convHandler) save(msg *packets_gossiper.PrivateMessage) {
	handler.privateMsgChan.AddPrivateMsg(msg)
	handler.messages = append(handler.messages, msg)
}

func (handler *convHandler) Save(msg *packets_gossiper.PrivateMessage) {
	handler.mux.Lock()
	defer handler.mux.Unlock()

	handler.save(msg)
}

func (handler *convHandler) GetAll() []*packets_gossiper.PrivateMessage {
	handler.mux.Lock()
	defer handler.mux.Unlock()

	msgs := []*packets_gossiper.PrivateMessage{}
	for _, msg := range handler.messages {
		msgs = append(msgs, msg)
	}
	return msgs
}

func (handler *convHandler) CreateAndSaveNextMessage(content string, to string, hopLimit uint32) *packets_gossiper.PrivateMessage {
	handler.mux.Lock()
	defer handler.mux.Unlock()

	msg := &packets_gossiper.PrivateMessage{
		Origin:      handler.origin,
		ID:          0,
		Text:        content,
		Destination: to,
		HopLimit:    hopLimit,
	}
	handler.save(msg)
	return msg
}
