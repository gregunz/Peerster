package packets_gossiper

import (
	"fmt"
	"github.com/gregunz/Peerster/logger"
)

type TxPublish struct {
	File     File   `json:"file"`
	HopLimit uint32 `json:"hop-limit"`
}

func (packet *TxPublish) AckPrint() {
	logger.Printlnf(packet.String())
}

func (packet *TxPublish) ToGossipPacket() *GossipPacket {
	return &GossipPacket{
		TxPublish: packet,
	}
}

func (packet *TxPublish) String() string {
	return fmt.Sprintf("TX PUBLISH file <%s> with hop-limit %d", packet.File.String(), packet.HopLimit)
}

func (msg TxPublish) Hopped() Transmittable {
	msg.HopLimit -= 1
	return &msg
}

func (msg *TxPublish) Dest() string {
	return ""
}

func (msg *TxPublish) IsTransmittable() bool {
	return msg.HopLimit > 0
}
