package packets_client

import "fmt"

type RequestFilePacket struct {
	Destination string `json:"destination"`
	FileName    string `json:"filename"`
	Request     string `json:"request"`
	HashValue   string `json:"hash-value"`
}

func (packet RequestFilePacket) String() string {
	return fmt.Sprintf("REQUEST FILE %s of %s with hash %s", packet.FileName, packet.Destination, packet.Request)
}

func (packet *RequestFilePacket) AckPrint() {
	fmt.Println(packet.String())
}

func (packet *RequestFilePacket) ToClientPacket() *ClientPacket {
	return &ClientPacket{
		RequestFile: packet,
	}
}
