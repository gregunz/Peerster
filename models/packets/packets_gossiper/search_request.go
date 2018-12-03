package packets_gossiper

import (
	"fmt"
	"github.com/gregunz/Peerster/logger"
	"strings"
)

type SearchRequest struct {
	Origin   string   `json:"origin"`
	Budget   uint64   `json:"budget"`
	Keywords []string `json:"keywords"`
}

func (packet *SearchRequest) AckPrint() {
	logger.Printlnf(packet.String())
}

func (packet *SearchRequest) ToGossipPacket() *GossipPacket {
	return &GossipPacket{
		SearchRequest: packet,
	}
}

func (packet *SearchRequest) String() string {
	return fmt.Sprintf("SEARCH REQUEST origin %s budget %d with keywords %s",
		packet.Origin, packet.Budget, strings.Join(packet.Keywords, " "))
}
