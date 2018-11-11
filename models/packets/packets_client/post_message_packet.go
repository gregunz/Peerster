package packets_client

import "fmt"

type PostMessagePacket struct {
	Message     string `json:"message"`
	Destination string `json:"destination"`
}

func (packet *PostMessagePacket) AckPrint() {
	fmt.Printf("CLIENT MESSAGE %s\n", packet.Message)
}

func (packet PostMessagePacket) String() string {
	toStr := ""
	if packet.Destination != "" {
		toStr = fmt.Sprintf("to %s", packet.Destination)
	}
	return fmt.Sprintf("POST MESSAGE %s%s", packet.Message, toStr)
}

func (packet *PostMessagePacket) ToClientPacket() *ClientPacket {
	return &ClientPacket{
		PostMessage: packet,
	}
}
