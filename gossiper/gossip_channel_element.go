package gossiper

import (
	"github.com/gregunz/Peerster/models/packets"
	"github.com/gregunz/Peerster/models/peers"
)

type GossipChannelElement struct {
	Packet *packets.GossipPacket
	From   *peers.Peer
}
