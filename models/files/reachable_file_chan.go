package files

import "github.com/gregunz/Peerster/models/updates"

type T = *SearchMatch

type ReachableFileChan interface {
	Get() T
	Push(file T)
}

type matchChan struct {
	updates.Chan
}

func NewMatchChan(activated bool) ReachableFileChan {
	return &matchChan{Chan: updates.NewChan(activated)}
}

func (ch *matchChan) Push(match T) {
	ch.Chan.Push(match)
}

func (ch *matchChan) Get() T {
	match, ok := ch.Chan.Get().(T)
	if !ok {
		return nil
	}
	return match
}
