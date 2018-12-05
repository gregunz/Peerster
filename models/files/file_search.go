package files

import (
	"github.com/gregunz/Peerster/models/packets/packets_gossiper"
	"github.com/gregunz/Peerster/utils"
	"strings"
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

func (search *Search) Ack(reply *packets_gossiper.SearchReply) {
	search.RLock()
	defer search.RUnlock()

	for _, result := range reply.Results {
		fileId := utils.HashToHex(result.MetafileHash)
		if match, ok := search.matches[fileId]; ok {
			match.Ack(reply.Origin, result)
		} else {
			keywordMatched := false
			for _, k := range search.Keywords {
				if strings.Contains(result.FileName, k) { // here is where we check for keyword match
					keywordMatched = true
					break
				}
			}
			if keywordMatched {
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
