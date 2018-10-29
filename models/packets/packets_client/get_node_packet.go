package packets_client

import (
	"fmt"
)

type GetNodePacket struct{}

func (packet *GetNodePacket) AckPrint() {
	fmt.Printf(packet.String())
}

func (packet GetNodePacket) String() string {
	return fmt.Sprintf("GET NODE\n")
}
