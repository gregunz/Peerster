package packets_gossiper

import (
	"fmt"
	"github.com/gregunz/Peerster/logger"
)

type BlockPublish struct {
	Block    Block  `json:"block"`
	HopLimit uint32 `json:"hop-limit"`
}

func (packet *BlockPublish) AckPrint() {
	logger.Printlnf(packet.String())
}

func (packet *BlockPublish) ToGossipPacket() *GossipPacket {
	return &GossipPacket{
		BlockPublish: packet,
	}
}

func (packet *BlockPublish) String() string {
	return fmt.Sprintf("BLOCK PUBLISH block %s hop-limit %d",
		packet.Block.String(), packet.HopLimit)
}

func (msg BlockPublish) Hopped() Transmittable {
	msg.HopLimit -= 1
	return &msg
}

func (msg *BlockPublish) Dest() string {
	return ""
}

func (msg *BlockPublish) IsTransmittable() bool {
	return msg.HopLimit > 0
}
