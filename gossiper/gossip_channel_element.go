package gossiper

import (
	"github.com/gregunz/Peerster/models/packets/packets_gossiper"
	"github.com/gregunz/Peerster/models/peers"
)

type GossipChannelElement struct {
	Packet *packets_gossiper.GossipPacket
	From   *peers.Peer
}
