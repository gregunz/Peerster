package clients

import (
	"github.com/gregunz/Peerster/models/packets/responses_client"
	"sync"
)

type Writer interface {
	WriteJSON(v *responses_client.ClientResponse) error
}

func NewWriter(writeJSON func(v *responses_client.ClientResponse) error) Writer {
	return &writer{
		writeJSON: writeJSON,
	}
}

type writer struct {
	writeJSON func(v *responses_client.ClientResponse) error
	mux       sync.Mutex
}

func (w *writer) WriteJSON(v *responses_client.ClientResponse) error {
	w.mux.Lock()
	defer w.mux.Unlock()

	return w.writeJSON(v)
}
