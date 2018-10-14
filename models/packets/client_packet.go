package packets

import "fmt"

type ClientPacket struct {
	Message string `json:"message"`
}

func (packet *ClientPacket) AckPrint() {
	fmt.Printf("CLIENT MESSAGE %s\n", packet.Message)
}
