package packets_client

import "fmt"

type RequestFilePacket struct {
	Destination string `json:"destination"`
	File        string `json:"file"`
	Request     string `json:"request"`
}

func (packet RequestFilePacket) String() string {
	return fmt.Sprintf("REQUEST FILE %s of %s with hash %s", packet.File, packet.Destination, packet.Request)
}

func (packet *RequestFilePacket) AckPrint() {
	fmt.Println(packet.String())
}

func (packet *RequestFilePacket) ToClientPacket() *ClientPacket {
	return &ClientPacket{
		RequestFile: packet,
	}
}
