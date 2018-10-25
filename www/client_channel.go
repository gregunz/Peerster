package www

import (
	"github.com/gregunz/Peerster/models/packets"
)

type ClientChannelElement struct {
	Packet *packets.ClientPacket
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
