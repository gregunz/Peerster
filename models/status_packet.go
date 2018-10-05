package models

import (
	"fmt"
	"strings"
)

type StatusPacket struct {
	Want []PeerStatus
}

func (packet *StatusPacket) AckPrint(fromPeer *Peer) {
	ls := []string{}
	for _, ps := range packet.Want {
		ls = append(ls, fmt.Sprintf("peer %s nextID %d", ps.Identifier, ps.NextID))
	}
	fmt.Printf("STATUS from %s %s",
		fromPeer.Addr.ToIpPort(), strings.Join(ls, " "))
}

func (packet *StatusPacket) ToGossipPacket() *GossipPacket {
	return &GossipPacket{
		Status: packet,
	}
}
