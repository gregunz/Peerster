package files

import "github.com/gregunz/Peerster/models/updates"

type StoredFileChan interface {
	Get() *FileType
	Push(file *FileType)
}

type fileChan struct {
	updates.Chan
}

func NewFileChan(activated bool) StoredFileChan {
	return &fileChan{Chan: updates.NewChan(activated)}
}

func (ch *fileChan) Push(file *FileType) {
	ch.Chan.Push(file)
}

func (ch *fileChan) Get() *FileType {
	file, ok := ch.Chan.Get().(*FileType)
	if !ok {
		return nil
	}
	return file
}
