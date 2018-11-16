package packets_gossiper

import (
	"fmt"
	"github.com/gregunz/Peerster/logger"
	"github.com/gregunz/Peerster/models/peers"
)

type RumorMessage struct {
	Origin string `json:"origin"`
	ID     uint32 `json:"id"`
	Text   string `json:"text"`
}

func (msg *RumorMessage) AckPrint(fromPeer *peers.Peer) {
	logger.Printlnf("RUMOR origin %s from %s ID %d contents %s",
		msg.Origin, fromPeer.Addr.ToIpPort(), msg.ID, msg.Text)
}

func (msg *RumorMessage) SendPrintMongering(toPeer *peers.Peer) {
	logger.Printlnf("MONGERING with %s", toPeer.Addr.ToIpPort())
}

func (msg *RumorMessage) SendPrintFlipped(toPeer *peers.Peer) {
	logger.Printlnf("FLIPPED COIN sending rumor to %s", toPeer.Addr.ToIpPort())
}

func (msg *RumorMessage) ToGossipPacket() *GossipPacket {
	return &GossipPacket{
		Rumor: msg,
	}
}

func (msg RumorMessage) String() string {
	return fmt.Sprintf("RUMOR origin %s ID %d contents %s",
		msg.Origin, msg.ID, msg.Text)
}
