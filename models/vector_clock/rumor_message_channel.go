package vector_clock

import (
	"github.com/gregunz/Peerster/models/packets/packets_gossiper"
	"github.com/gregunz/Peerster/models/updates"
)

type T = *packets_gossiper.RumorMessage

type RumorChan interface {
	Push(T)
	Get() T
}

type rumorChan struct {
	updates.Chan
}

func NewRumorChan(activated bool) RumorChan {
	return &rumorChan{Chan: updates.NewChan(activated)}
}

func (ch *rumorChan) Push(s T) {
	ch.Chan.Push(s)
}

func (ch *rumorChan) Get() T {
	s, ok := ch.Chan.Get().(T)
	if !ok {
		return nil
	}
	return s
}
