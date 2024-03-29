package packets_gossiper

import (
	"fmt"
	"github.com/gregunz/Peerster/logger"
	"github.com/gregunz/Peerster/utils"
)

type DataRequest struct {
	Origin      string `json:"origin"`
	Destination string `json:"destination"`
	HopLimit    uint32 `json:"hop-limit"`
	HashValue   []byte `json:"hash-value"`
}

func (packet *DataRequest) String() string {
	return fmt.Sprintf("DATA REQUEST from %s hop-limit %d hash %s to %s",
		packet.Origin, packet.HopLimit, utils.HashToHex(packet.HashValue), packet.Destination)
}

func (packet *DataRequest) AckPrint(myOrigin string) {
	if myOrigin == packet.Destination {
		logger.Printlnf(packet.String())
	}
}

func (packet *DataRequest) ToGossipPacket() *GossipPacket {
	return &GossipPacket{
		DataRequest: packet,
	}
}

func (msg DataRequest) Hopped() Transmittable {
	msg.HopLimit -= 1
	return &msg
}

func (msg *DataRequest) Dest() string {
	return msg.Destination
}

func (msg *DataRequest) IsTransmittable() bool {
	return msg.HopLimit > 0
}
