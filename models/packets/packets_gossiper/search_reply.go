package packets_gossiper

import (
	"fmt"
	"github.com/gregunz/Peerster/logger"
	"github.com/gregunz/Peerster/utils"
	"strings"
)

type SearchReply struct {
	Origin      string          `json:"origin"`
	Destination string          `json:"destination"`
	HopLimit    uint32          `json:"hop-limit"`
	Results     []*SearchResult `json:"results"`
}

func (packet *SearchReply) AckPrint() {
	for _, res := range packet.Results {
		chunkList := []string{}
		for _, chunk := range res.ChunkMap {
			chunkList = append(chunkList, fmt.Sprintf("%d", chunk))
		}
		logger.Printlnf("FOUND match %s at %s metafile=%s chunks=%s",
			res.FileName, packet.Origin, utils.HashToHex(res.MetafileHash), strings.Join(chunkList, ","))
	}
}

func (packet *SearchReply) ToGossipPacket() *GossipPacket {
	return &GossipPacket{
		SearchReply: packet,
	}
}

func (packet *SearchReply) String() string {
	results := []string{}
	for _, res := range packet.Results {
		results = append(results, fmt.Sprintf("<%s>", res.String()))
	}
	return fmt.Sprintf("SEARCH REPLY origin %s hop-limit %d to %s with results %s",
		packet.Origin, packet.HopLimit, packet.Destination, strings.Join(results, " "))
}

func (msg SearchReply) Hopped() Transmittable {
	msg.HopLimit -= 1
	return &msg
}

func (msg *SearchReply) Dest() string {
	return msg.Destination
}

func (msg *SearchReply) IsTransmittable() bool {
	return msg.HopLimit > 0
}
