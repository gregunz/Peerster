package packets_client

import (
	"fmt"
	"github.com/gregunz/Peerster/logger"
)

type RequestFilePacket struct {
	Destination string `json:"destination"`
	Filename    string `json:"filename"`
	Request     string `json:"request"`
}

func (packet *RequestFilePacket) String() string {
	s := ""
	if packet.Destination != "" {
		s = fmt.Sprintf(" from %s", packet.Destination)
	}
	return fmt.Sprintf("REQUEST FILE %s%s with hash %s", packet.Filename, s, packet.Request)
}

func (packet *RequestFilePacket) AckPrint() {
	logger.Printlnf(packet.String())
}

func (packet *RequestFilePacket) ToClientPacket() *ClientPacket {
	return &ClientPacket{
		RequestFile: packet,
	}
}
