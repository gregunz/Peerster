package packets

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
