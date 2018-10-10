package packets

import (
	"fmt"
	"github.com/gregunz/Peerster/models/peers"
	"strings"
)

type StatusPacket struct {
	Want []PeerStatus
}

func (packet *StatusPacket) AckPrint(fromPeer *peers.Peer) {
	ls := packet.wantString()
	fmt.Printf("STATUS from %s %s\n",
		fromPeer.Addr.ToIpPort(), strings.Join(ls, " "))
}

func (packet *StatusPacket) ToGossipPacket() *GossipPacket {
	return &GossipPacket{
		Status: packet,
	}
}

func (packet StatusPacket) String() string {
	ls := packet.wantString()
	return fmt.Sprintf("STATUS %s", strings.Join(ls, " "))
}

func (packet *StatusPacket) wantString() []string {
	ls := []string{}
	for _, ps := range packet.Want {
		ls = append(ls, ps.String())
	}
	return ls
}

func (packet *StatusPacket) ToMap() map[string]uint32 {
	statusMap := map[string]uint32{}
	for _, ps := range packet.Want {
		statusMap[ps.Identifier] = ps.NextID
	}
	return statusMap
}
