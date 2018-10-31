package updates

import (
	"github.com/gregunz/Peerster/models/packets/packets_gossiper"
	"github.com/gregunz/Peerster/models/peers"
)

type UpdateChannels struct {
	rumorMsg   chan *packets_gossiper.RumorMessage
	node       chan *peers.Peer
	privateMsg chan *packets_gossiper.PrivateMessage
	origin     chan string
	activated  bool
}

func NewChannels() *UpdateChannels {
	return &UpdateChannels{
		rumorMsg:   make(chan *packets_gossiper.RumorMessage),
		node:       make(chan *peers.Peer),
		privateMsg: make(chan *packets_gossiper.PrivateMessage),
		origin:     make(chan string),
		activated:  true,
	}
}

func (ch *UpdateChannels) AddRumor(msg *packets_gossiper.RumorMessage) {
	if ch.activated {
		go func() { ch.rumorMsg <- msg }()
	}
}

func (ch *UpdateChannels) GetRumor() *packets_gossiper.RumorMessage {
	if ch.activated {
		msg, ok := <-ch.rumorMsg
		if ok {
			return msg
		}
	}
	return nil
}

func (ch *UpdateChannels) AddPrivateMsg(msg *packets_gossiper.PrivateMessage) {
	if ch.activated {
		go func() { ch.privateMsg <- msg }()
	}
}

func (ch *UpdateChannels) GetPrivateMsg() *packets_gossiper.PrivateMessage {
	if ch.activated {
		msg, ok := <-ch.privateMsg
		if ok {
			return msg
		}
	}
	return nil
}

func (ch *UpdateChannels) AddNode(node *peers.Peer) {
	if ch.activated {
		go func() { ch.node <- node }()
	}
}

func (ch *UpdateChannels) GetNode() *peers.Peer {
	if ch.activated {
		node, ok := <-ch.node
		if ok {
			return node
		}
	}
	return nil
}

func (ch *UpdateChannels) AddOrigin(o string) {
	if ch.activated {
		go func() { ch.origin <- o }()
	}
}

func (ch *UpdateChannels) GetOrigin() string {
	if ch.activated {
		o, ok := <-ch.origin
		if ok {
			return o
		}
	}
	return ""
}
