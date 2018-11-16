package packets_gossiper

import (
	"fmt"
	"github.com/gregunz/Peerster/logger"
	"github.com/gregunz/Peerster/utils"
)

type DataReply struct {
	Origin      string `json:"origin"`
	Destination string `json:"destination"`
	HopLimit    uint32 `json:"hop-limit"`
	HashValue   []byte `json:"hash-value"`
	Data        []byte `json:"data"`
}

func (packet *DataReply) String() string {
	return fmt.Sprintf("DATA REPLY from %s hop-limit %d hash %s to %s",
		packet.Origin, packet.HopLimit, utils.HashToHex(packet.HashValue), packet.Destination)
}

func (packet *DataReply) AckPrint(myOrigin string) {
	if myOrigin == packet.Destination {
		logger.Printlnf(packet.String())
	}
}

func (packet *DataReply) ToGossipPacket() *GossipPacket {
	return &GossipPacket{
		DataReply: packet,
	}
}

func (msg DataReply) Hopped() Transmittable {
	msg.HopLimit -= 1
	return &msg
}

func (msg *DataReply) Dest() string {
	return msg.Destination
}

func (msg *DataReply) IsTransmittable() bool {
	return msg.HopLimit > 0
}
