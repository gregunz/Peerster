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
