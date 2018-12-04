package packets_gossiper

import (
	"fmt"
	"github.com/gregunz/Peerster/models/peers"
	"strings"
)

type GossipPacket struct {
	Simple        *SimpleMessage  `json:"simple"`
	Rumor         *RumorMessage   `json:"rumor"`
	Status        *StatusPacket   `json:"status"`
	Private       *PrivateMessage `json:"private"`
	DataRequest   *DataRequest    `json:"data-request"`
	DataReply     *DataReply      `json:"data-reply"`
	SearchRequest *SearchRequest  `json:"search-request"`
	SearchReply   *SearchReply    `json:"search-reply"`
	TxPublish     *TxPublish      `json:"tx-publish"`
	BlockPublish  *BlockPublish   `json:"block-publish"`
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

func (packet *GossipPacket) IsSearchRequest() bool {
	return packet.SearchRequest != nil
}

func (packet *GossipPacket) IsSearchReply() bool {
	return packet.SearchReply != nil
}

func (packet *GossipPacket) IsTxPublish() bool {
	return packet.TxPublish != nil
}

func (packet *GossipPacket) IsBlockPublish() bool {
	return packet.BlockPublish != nil
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
	if packet.IsSearchRequest() {
		counter += 1
	}
	if packet.IsSearchReply() {
		counter += 1
	}
	if packet.IsTxPublish() {
		counter += 1
	}
	if packet.IsBlockPublish() {
		counter += 1
	}
	if counter == 1 {
		return nil
	}
	return fmt.Errorf("unexpected gossip packet format: %s", packet.String())
}

func (packet *GossipPacket) String() string {
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
	if packet.IsSearchRequest() {
		ls = append(ls, packet.SearchRequest.String())
	}
	if packet.IsSearchReply() {
		ls = append(ls, packet.SearchReply.String())
	}
	if packet.IsTxPublish() {
		ls = append(ls, packet.TxPublish.String())
	}
	if packet.IsBlockPublish() {
		ls = append(ls, packet.BlockPublish.String())
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

	// for now on, some of the rest is commented for better performance, might setup a verbosity level to enable them anyway

	if packet.IsDataRequest() { //not required by homework (to check)
		//packet.DataRequest.AckPrint(myOrigin)
	}
	if packet.IsDataReply() { //not required by homework (to check)
		//packet.DataReply.AckPrint(myOrigin)
	}
	if packet.IsSearchRequest() { //not required by homework (to check)
		//packet.SearchRequest.AckPrint()
	}
	if packet.IsSearchReply() {
		packet.SearchReply.AckPrint()
	}
	if packet.IsTxPublish() { //not required by homework (to check)
		//packet.TxPublish.AckPrint()
	}
	if packet.IsBlockPublish() { //not required by homework (to check)
		//packet.BlockPublish.AckPrint()
	}
}
