package www

import (
	"github.com/gregunz/Peerster/models/packets/responses_client"
	"sync"
)

type Writer interface {
	WriteJSON(v *responses_client.ClientResponse) error
}

type ProtoWriter struct {
	writeJSON func(v *responses_client.ClientResponse) error
	mux       sync.Mutex
}

func (w *ProtoWriter) WriteJSON(v *responses_client.ClientResponse) error {
	w.mux.Lock()
	defer w.mux.Unlock()

	return w.writeJSON(v)
}
