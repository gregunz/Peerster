package packets_gossiper

import "fmt"

type DataRequest struct {
	Origin      string `json:"origin"`
	Destination string `json:"destination"`
	HopLimit    uint32 `json:"hop-limit"`
	HashValue   []byte `json:"hash-value"`
}

func (packet DataRequest) String() string {
	return fmt.Sprintf("DATA REQUEST from %s hop-limit %d hash %s to %s",
		packet.Origin, packet.HopLimit, string(packet.HashValue), packet.Destination)
}

func (packet *DataRequest) AckPrint(myOrigin string) {
	if myOrigin == packet.Destination {
		fmt.Println(packet.String())
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
