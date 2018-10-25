package packets

import "fmt"

type GetMessagePacket struct{}

func (packet *GetMessagePacket) AckPrint() {
	fmt.Printf(packet.String())
}

func (packet GetMessagePacket) String() string {
	return fmt.Sprintf("GET MESSAGE\n")
}
