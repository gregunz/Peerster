package packets_client

import (
	"fmt"
	"github.com/gregunz/Peerster/logger"
	"github.com/gregunz/Peerster/models/peers"
)

type PostNodePacket struct {
	Node string `json:"node"`
}

func (packet *PostNodePacket) ToPeer() *peers.Peer {
	return peers.NewPeer(packet.Node)
}

func (packet *PostNodePacket) AckPrint() {
	logger.Printlnf(packet.String())
}

func (packet *PostNodePacket) String() string {
	return fmt.Sprintf("ADD NODE %s", packet.Node)
}
