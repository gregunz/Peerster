package packets_gossiper

import "fmt"

type DataReply struct {
	Origin      string `json:"origin"`
	Destination string `json:"destination"`
	HopLimit    uint32 `json:"hop-limit"`
	HashValue   []byte `json:"hash-value"`
	Data        []byte `json:"data"`
}

func (packet DataReply) String() string {
	return fmt.Sprintf("DATA REPLY from %s hop-limit %d hash %s to %s",
		packet.Origin, packet.HopLimit, string(packet.HashValue), packet.Destination)
}

func (packet *DataReply) AckPrint(myOrigin string) {
	if myOrigin == packet.Destination {
		fmt.Println(packet.String())
	}
}
