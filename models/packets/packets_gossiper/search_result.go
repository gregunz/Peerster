package packets_gossiper

import (
	"fmt"
	"github.com/gregunz/Peerster/utils"
)

type SearchResult struct {
	FileName     string   `json:"file-name"`
	MetafileHash []byte   `json:"metafile-hash"`
	ChunkMap     []uint64 `json:"chunk-map"`
	ChunkCount   uint64   `json:"chunk-count"`
}

func (packet *SearchResult) String() string {
	return fmt.Sprintf("SEARCH RESULT named %s with %d chunks and metafile hash %s",
		packet.FileName, packet.ChunkCount, utils.HashToHex(packet.MetafileHash))
}
