package origin_handlers

import (
	"github.com/gregunz/Peerster/models/packets/packets_gossiper"
	"github.com/gregunz/Peerster/models/peers"
)

type RoutingTable interface {
	AckRumor(rumor *packets_gossiper.RumorMessage, fromPeer *peers.Peer)
	Get(origin string) *peers.Peer
}

type ProtoRoutingTable struct {
	originsHandler *OriginsHandler
}

func (table ProtoRoutingTable) getOrCreateHandler(origin string) *routingHandler {
	return table.originsHandler.GetOrCreateHandler(origin).routingHandler
}

func (table ProtoRoutingTable) AckRumor(rumor *packets_gossiper.RumorMessage, fromPeer *peers.Peer) {
	if rumor.Origin != table.originsHandler.MyOrigin {
		table.getOrCreateHandler(rumor.Origin).AckRumor(rumor, fromPeer)
	}
}

func (table ProtoRoutingTable) Get(origin string) *peers.Peer {
	return table.getOrCreateHandler(origin).peer
}
