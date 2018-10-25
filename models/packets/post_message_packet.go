package packets

import "fmt"

type PostMessagePacket struct {
	Message string `json:"message"`
}

func (packet *PostMessagePacket) AckPrint() {
	fmt.Printf("CLIENT MESSAGE %s\n", packet.Message)
}

func (packet PostMessagePacket) String() string {
	return fmt.Sprintf("POST MESSAGE %s\n", packet.Message)
}
