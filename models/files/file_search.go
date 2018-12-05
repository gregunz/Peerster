package files

import (
	"github.com/gregunz/Peerster/models/packets/packets_client"
	"github.com/gregunz/Peerster/models/packets/packets_gossiper"
	"github.com/gregunz/Peerster/utils"
	"sync"
)

const (
	maxBudget         = 32
	minNumFullMatches = 2
)

type Search struct {
	Keywords []string
	Budget   uint64
	matches  map[string]*SearchMatch
	sync.RWMutex
}

func newSearch(keywords []string, initialBudget uint64) *Search {
	return &Search{
		Keywords: keywords,
		Budget:   initialBudget,
		matches:  map[string]*SearchMatch{},
	}
}

func (search *Search) DoubleBudget() bool {
	search.Lock()
	defer search.Unlock()

	if search.Budget < maxBudget {
		search.Budget = utils.Min_uint64(search.Budget*2, maxBudget)
		return true
	}
	return false //cannot double Budget when reached `maxBudget`
}

func (search *Search) Match(filename string) bool {
	search.RLock()
	defer search.RUnlock()
	return utils.Match(filename, search.Keywords)
}

func (search *Search) Ack(reply *packets_gossiper.SearchReply) {
	search.RLock()
	defer search.RUnlock()

	for _, result := range reply.Results {
		fileId := utils.HashToHex(result.MetafileHash)
		if match, ok := search.matches[fileId]; ok {
			match.Ack(reply.Origin, result)
		} else {
			if search.Match(result.FileName) {
				search.Lock()
				search.matches[fileId] = NewSearchMatch(reply.Origin, result)
				search.Unlock()
			}
		}
	}
}

func (search *Search) IsFull() bool {
	search.RLock()
	defer search.RUnlock()

	numFullMatches := 0
	for _, m := range search.matches {
		if m.IsFull() {
			numFullMatches += 1
		}
	}
	return numFullMatches >= minNumFullMatches
}

func (search *Search) ToRequestFiles() []*packets_client.RequestFilePacket {
	search.RLock()
	defer search.RUnlock()

	requests := []*packets_client.RequestFilePacket{}
	// adding metafile requests first
	for _, match := range search.matches {
		for origin, _ := range match.AllOrigins() {
			requests = append(requests, &packets_client.RequestFilePacket{
				Destination: origin,
				Filename:    match.filename,
				Request:     utils.HashToHex(match.metafileHash),
			})
		}
	}
	return requests
}
