package packets_gossiper

import (
	"fmt"
	"github.com/gregunz/Peerster/models/peers"
	"strings"
)

type GossipPacket struct {
	Simple      *SimpleMessage  `json:"simple"`
	Rumor       *RumorMessage   `json:"rumor"`
	Status      *StatusPacket   `json:"status"`
	Private     *PrivateMessage `json:"private"`
	DataRequest *DataRequest    `json:"data-request"`
	DataReply   *DataReply      `json:"data-reply"`
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
	if packet.IsPrivate() {
		counter += 1
	}
	if packet.IsDataRequest() {
		counter += 1
	}
	if packet.IsDataReply() {
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

func (packet *GossipPacket) IsPrivate() bool {
	return packet.Private != nil
}

func (packet *GossipPacket) IsDataRequest() bool {
	return packet.DataRequest != nil
}

func (packet *GossipPacket) IsDataReply() bool {
	return packet.DataReply != nil
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
	if packet.IsPrivate() {
		ls = append(ls, packet.Private.String())
	}
	if packet.IsDataRequest() {
		ls = append(ls, packet.DataRequest.String())
	}
	if packet.IsDataReply() {
		ls = append(ls, packet.DataReply.String())
	}
	if len(ls) == 0 {
		return "<empty>"
	}
	return strings.Join(ls, " + ")
}

func (packet *GossipPacket) AckPrint(fromPeer *peers.Peer, myOrigin string) {
	if packet.IsSimple() {
		packet.Simple.AckPrint()
	}
	if packet.IsRumor() {
		packet.Rumor.AckPrint(fromPeer)
	}
	if packet.IsStatus() {
		packet.Status.AckPrint(fromPeer)
	}
	if packet.IsPrivate() {
		packet.Private.AckPrint(myOrigin)
	}
	if packet.IsDataRequest() {
		packet.DataRequest.AckPrint(myOrigin)
	}
	if packet.IsDataReply() {
		packet.DataReply.AckPrint(myOrigin)
	}
}
