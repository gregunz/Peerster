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

func newSearch(keywords []string, initialBudget uint64, FileChan ReachableFileChan) *Search {
	return &Search{
		Keywords: keywords,
		Budget:   initialBudget,
		matches:  map[string]*SearchMatch{},
	}
}

func (search *Search) DoubleBudget() bool {
	search.Lock()
	defer search.Unlock()

	prevBudget := search.Budget
	if search.Budget < maxBudget {
		search.Budget = utils.Min(prevBudget*2, maxBudget)
		search.Budget = utils.Max(search.Budget, prevBudget+1) // in case it was set to zero (which is stupid but anyway)
		return true
	}
	return false //cannot double Budget when reached `maxBudget`
}

func (search *Search) Match(filename string) bool {
	search.RLock()
	defer search.RUnlock()
	return utils.Match(filename, search.Keywords)
}

func (search *Search) Ack(reply *packets_gossiper.SearchReply) []*SearchMatch {
	search.Lock()
	defer search.Unlock()

	newFilesFound := []*SearchMatch{}
	for _, result := range reply.Results {
		fileId := utils.HashToHex(result.MetafileHash)

		var match *SearchMatch
		var ok bool

		if match, ok = search.matches[fileId]; !ok {
			if utils.Match(result.FileName, search.Keywords) {
				match = NewSearchMatch(result)
				search.matches[fileId] = match
			}
		}
		if match != nil && match.Ack(reply.Origin, result) {
			newFilesFound = append(newFilesFound, match)
		}
	}
	return newFilesFound
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

func (search *Search) GetAllMatches() []*SearchMatch {
	search.RLock()
	defer search.RUnlock()

	matches := []*SearchMatch{}
	for _, m := range search.matches {
		matches = append(matches, m)
	}
	return matches
}

func (search *Search) ToRequestFiles(filename, metahash string) []*packets_client.RequestFilePacket {
	search.RLock()
	defer search.RUnlock()

	requests := []*packets_client.RequestFilePacket{}
	// adding metafile requests first
	for _, match := range search.matches {
		match.RLock()
		if match.metafileHash == metahash {
			for origin, _ := range match.AllOrigins() {
				requests = append(requests, &packets_client.RequestFilePacket{
					Destination: origin,
					Filename:    filename,
					Request:     match.metafileHash,
				})
			}
		}
		match.RUnlock()
	}
	return requests
}
