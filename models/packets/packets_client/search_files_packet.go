package packets_client

import (
	"fmt"
	"github.com/gregunz/Peerster/logger"
	"strings"
)

type SearchFilesPacket struct {
	Keywords []string `json:"keywords"`
	Budget   uint64   `json:"budget"`
}

func (packet *SearchFilesPacket) String() string {
	return fmt.Sprintf("SEARCH FILES with keywords=%s and budget=%d",
		strings.Join(packet.Keywords, ","), packet.Budget)
}

func (packet *SearchFilesPacket) AckPrint() {
	logger.Printlnf(packet.String())
}

func (packet *SearchFilesPacket) ToClientPacket() *ClientPacket {
	return &ClientPacket{
		SearchFiles: packet,
	}
}
