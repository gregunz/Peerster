package packets

import (
	"fmt"
	"github.com/gregunz/Peerster/models/peers"
)

type AddNodePacket struct {
	Node string `json:"node"`
}

func (packet AddNodePacket) ToPeer() *peers.Peer {
	return peers.NewPeer(packet.Node)
}

func (packet *AddNodePacket) AckPrint() {
	fmt.Printf(packet.String())
}

func (packet AddNodePacket) String() string {
	return fmt.Sprintf("ADD NODE %s\n", packet.Node)
}
