package files

import "github.com/gregunz/Peerster/models/updates"

type FileChan interface {
	Get() string
	Push(filename string)
}

type fileChan struct {
	ch updates.Chan
}

func NewFileChan(activated bool) FileChan {
	return &fileChan{ch: updates.NewChan(activated)}
}

func (ch *fileChan) Push(filename string) {
	ch.ch.Push(filename)
}

func (ch *fileChan) Get() string {
	filename, ok := ch.ch.Get().(string)
	if !ok {
		return ""
	}
	return filename
}
