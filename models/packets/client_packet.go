package packets

import "fmt"

type ClientPacket struct {
	Message string
}

func (packet *ClientPacket) AckPrint() {
	fmt.Printf("CLIENT MESSAGE %s\n", packet.Message)
}
