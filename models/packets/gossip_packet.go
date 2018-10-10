package packets

import (
	"fmt"
	"github.com/gregunz/Peerster/common"
	"github.com/gregunz/Peerster/models/peers"
	"strings"
)

type GossipPacket struct {
	Simple *SimpleMessage
	Rumor  *RumorMessage
	Status *StatusPacket
}

func (packet *GossipPacket) Check() error {
	var counter uint = 0
	if packet.IsSimple() {
		counter += 1
	}
	if packet.IsRumor() {
		counter += 1
	}
	if packet.IsStatus() {
		counter += 1
	}
	if counter == 1 {
		return nil
	}
	return fmt.Errorf("GossipPacket should have at least and at most one entry not nil instead of %s", packet.String())
}

func (packet *GossipPacket) IsSimple() bool {
	return packet.Simple != nil
}

func (packet *GossipPacket) IsRumor() bool {
	return packet.Rumor != nil
}

func (packet *GossipPacket) IsStatus() bool {
	return packet.Status != nil
}

func (packet GossipPacket) String() string {
	ls := []string{}
	if packet.IsSimple() {
		ls = append(ls, packet.Simple.String())
	}
	if packet.IsRumor() {
		ls = append(ls, packet.Rumor.String())
	}
	if packet.IsStatus() {
		ls = append(ls, packet.Status.String())
	}
	if len(ls) == 0 {
		common.HandleError(fmt.Errorf("empty gossip packet"))
		return ""
	}
	return strings.Join(ls, " + ")
}

func (packet *GossipPacket) AckPrint(fromPeer *peers.Peer) {
	if packet.IsSimple() {
		packet.Simple.AckPrint()
	}
	if packet.IsRumor() {
		packet.Rumor.AckPrint(fromPeer)
	}
	if packet.IsStatus() {
		packet.Status.AckPrint(fromPeer)
	}
}
