package models

import (
	"fmt"
)

type RumorMessage struct {
	Origin string
	ID     uint32
	Text   string
}

func (msg *RumorMessage) AckPrint(fromPeer *Peer) {
	fmt.Printf("RUMOR origin %s from %s ID %d contents %s\n",
		msg.Origin, fromPeer.Addr.ToIpPort(), msg.Origin, msg.Text)
}

func (msg *RumorMessage) SendPrint(toPeer *Peer, flipped bool) {
	if flipped {
		fmt.Printf("FLIPPED COIN sending rumor to %s", toPeer.Addr.ToIpPort())
	} else {
		fmt.Printf("MONGERING with %s", toPeer.Addr.ToIpPort())
	}
}

func (msg *RumorMessage) ToGossipPacket() *GossipPacket {
	return &GossipPacket{
		Rumor: msg,
	}
}
