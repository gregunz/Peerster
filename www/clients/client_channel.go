package clients

import (
	"github.com/gregunz/Peerster/models/packets/packets_client"
)

type ClientChannelElement struct {
	Packet *packets_client.ClientPacket
	Writer Writer
}
