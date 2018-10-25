package packets

import "fmt"

type GetIdPacket struct{}

func (packet *GetIdPacket) AckPrint() {
	fmt.Printf(packet.String())
}

func (packet GetIdPacket) String() string {
	return fmt.Sprintf("GET ID\n")
}
