package www

import (
	"github.com/gregunz/Peerster/models/packets/packets_client"
)

type ClientChannelElement struct {
	Packet *packets_client.ClientPacket
	Writer Writer
}

type Writer interface {
	WriteJSON(v interface{}) error
}

type ProtoWriter struct {
	writeJSON func(v interface{}) error
}

func (w *ProtoWriter) WriteJSON(v interface{}) error {
	return w.writeJSON(v)
}
