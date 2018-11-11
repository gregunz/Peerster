package files

import "github.com/gregunz/Peerster/models/updates"

type FileChan interface {
	Get() *FileType
	Push(file *FileType)
}

type fileChan struct {
	ch updates.Chan
}

func NewFileChan(activated bool) FileChan {
	return &fileChan{ch: updates.NewChan(activated)}
}

func (ch *fileChan) Push(file *FileType) {
	ch.ch.Push(file)
}

func (ch *fileChan) Get() *FileType {
	file, ok := ch.ch.Get().(*FileType)
	if !ok {
		return nil
	}
	return file
}
