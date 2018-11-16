package packets_client

import (
	"fmt"
	"github.com/gregunz/Peerster/logger"
)

type GetIdPacket struct{}

func (packet *GetIdPacket) AckPrint() {
	logger.Printlnf(packet.String())
}

func (packet *GetIdPacket) String() string {
	return fmt.Sprintf("GET ID")
}
