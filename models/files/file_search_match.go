package files

import (
	"github.com/gregunz/Peerster/models/packets/packets_gossiper"
	"github.com/gregunz/Peerster/utils"
	"sync"
)

type SearchMatch struct {
	metafileHash string
	filenames    map[string]bool
	chunkOrigins map[uint64]map[string]bool
	numOfChunks  uint64

	sync.RWMutex
}

func NewSearchMatch(result *packets_gossiper.SearchResult) *SearchMatch {
	match := &SearchMatch{
		filenames:    map[string]bool{},
		metafileHash: utils.HashToHex(result.MetafileHash),
		chunkOrigins: map[uint64]map[string]bool{},
		numOfChunks:  result.ChunkCount,
	}
	match.filenames[result.FileName] = true
	return match
}

func (match *SearchMatch) Ack(origin string, result *packets_gossiper.SearchResult) bool {
	match.Lock()
	defer match.Unlock()

	numChunks := len(match.chunkOrigins)
	for _, chunkIdx := range result.ChunkMap {
		if set, ok := match.chunkOrigins[chunkIdx]; ok {
			match.filenames[result.FileName] = true
			set[origin] = true
		} else {
			set := map[string]bool{}
			set[origin] = true
			match.chunkOrigins[chunkIdx] = set
			match.filenames[result.FileName] = true
		}
	}
	return len(match.chunkOrigins) > numChunks && uint64(len(match.chunkOrigins)) == match.numOfChunks
}

func (match *SearchMatch) AllOrigins() map[string]bool {
	match.RLock()
	defer match.RUnlock()

	originsSet := map[string]bool{}
	for _, origins := range match.chunkOrigins {
		for origin, _ := range origins {
			originsSet[origin] = true
		}
	}
	return originsSet
}

func (match *SearchMatch) IsFull() bool {
	match.RLock()
	defer match.RUnlock()

	return uint64(len(match.chunkOrigins)) == match.numOfChunks
}

type SearchMetadata struct {
	Filename    string
	MetaHash    string
	NumOfChunks uint64
}

func (match *SearchMatch) ToSearchMetadata() []*SearchMetadata {
	match.RLock()
	defer match.RUnlock()

	responses := []*SearchMetadata{}
	for fn := range match.filenames {
		responses = append(responses, &SearchMetadata{
			Filename:    fn,
			MetaHash:    match.metafileHash,
			NumOfChunks: match.numOfChunks,
		})
	}
	return responses
}
