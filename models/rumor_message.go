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
		msg.Origin, fromPeer.Addr.ToIpPort(), msg.ID, msg.Text)
}

func (msg *RumorMessage) SendPrint(toPeer *Peer, flipped bool) {
	if flipped {
		fmt.Printf("FLIPPED COIN sending rumor to %s\n", toPeer.Addr.ToIpPort())
	} else {
		fmt.Printf("MONGERING with %s\n", toPeer.Addr.ToIpPort())
	}
}

func (msg *RumorMessage) ToGossipPacket() *GossipPacket {
	return &GossipPacket{
		Rumor: msg,
	}
}

func (msg RumorMessage) String() string {
	return fmt.Sprintf("RUMOR origin %s ID %d contents %s",
		msg.Origin, msg.Origin, msg.Text)
}