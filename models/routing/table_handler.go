package routing

import (
	"fmt"
	"github.com/gregunz/Peerster/models/packets/packets_gossiper"
	"github.com/gregunz/Peerster/models/peers"
	"sync"
)

type tableHandler struct {
	origin   string
	peer     *peers.Peer
	latestID uint32
	mux      sync.Mutex
}

func newRoutingTableHandler(origin string) *tableHandler {
	return &tableHandler{
		origin:   origin,
		peer:     nil,
		latestID: 0,
	}
}

func (handler tableHandler) AckRumor(rumor *packets_gossiper.RumorMessage, fromPeer *peers.Peer) {
	handler.mux.Lock()
	defer handler.mux.Unlock()

	if rumor.Origin == handler.origin && rumor.ID > handler.latestID {
		handler.latestID = rumor.ID
		handler.peer = fromPeer
		fmt.Printf("DSDV %s %s\n", handler.origin, handler.peer.Addr.ToIpPort())
	}
}
