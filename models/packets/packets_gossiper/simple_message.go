package packets_gossiper

import (
	"fmt"
	"github.com/gregunz/Peerster/logger"
)

type SimpleMessage struct {
	OriginalName  string `json:"original-name"`
	RelayPeerAddr string `json:"relay-peer-address"`
	Contents      string `json:"contents"`
}

func (msg *SimpleMessage) AckPrint() {
	logger.Printlnf(msg.String())
}

func (msg *SimpleMessage) ToGossipPacket() *GossipPacket {
	return &GossipPacket{
		Simple: msg,
	}
}

func (msg *SimpleMessage) String() string {
	return fmt.Sprintf("SIMPLE MESSAGE origin %s from %s contents %s",
		msg.OriginalName, msg.RelayPeerAddr, msg.Contents)
}
