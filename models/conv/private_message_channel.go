package conv

import (
	"github.com/gregunz/Peerster/models/packets/packets_gossiper"
	"github.com/gregunz/Peerster/models/updates"
)

type T = *packets_gossiper.PrivateMessage

type PrivateMsgChan interface {
	Push(T)
	Get() T
}

func NewPrivateChan(activated bool) PrivateMsgChan {
	return &_chan{Chan: updates.NewChan(activated)}
}

type _chan struct {
	updates.Chan
}

func (ch *_chan) Push(s T) {
	ch.Chan.Push(s)
}

func (ch *_chan) Get() T {
	s, ok := ch.Chan.Get().(T)
	if !ok {
		return nil
	}
	return s
}
