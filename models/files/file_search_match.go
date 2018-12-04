package files

import (
	"github.com/gregunz/Peerster/models/packets/packets_gossiper"
	"sync"
)

type SearchMatch struct {
	filename     string
	metafileHash []byte
	chunkOrigins map[uint64]map[string]bool
	numOfChunks  uint64

	sync.RWMutex
}

func NewSearchMatch(origin string, result *packets_gossiper.SearchResult) *SearchMatch {
	match := &SearchMatch{
		filename:     result.FileName,
		metafileHash: result.MetafileHash,
		chunkOrigins: map[uint64]map[string]bool{},
		numOfChunks:  result.ChunkCount,
	}
	match.Ack(origin, result)
	return match
}

func (match *SearchMatch) Ack(origin string, result *packets_gossiper.SearchResult) {
	match.Lock()
	defer match.Unlock()

	for _, c := range result.ChunkMap {
		if set, ok := match.chunkOrigins[c]; ok {
			set[origin] = true
		} else {
			set := map[string]bool{}
			set[origin] = true
			match.chunkOrigins[c] = set
		}
	}
}

func (match *SearchMatch) IsFull() bool {
	match.RLock()
	defer match.RUnlock()

	return uint64(len(match.chunkOrigins)) == match.numOfChunks
}
