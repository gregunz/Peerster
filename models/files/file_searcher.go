package files

import (
	"github.com/gregunz/Peerster/models/packets/packets_gossiper"
	"sync"
)

type Searcher interface {
	Search(keywords []string, initialBudget uint64) *Search
	Ack(reply *packets_gossiper.SearchReply)
	GetFullSearches() []*Search
}

type searcher struct {
	searches map[*Search]bool

	sync.RWMutex
}

func NewSearcher() *searcher {
	return &searcher{
		searches: map[*Search]bool{},
	}
}

func (searcher *searcher) Search(keywords []string, initialBudget uint64) *Search {
	searcher.Lock()
	defer searcher.Unlock()

	newSearch := newSearch(keywords, initialBudget)
	searcher.searches[newSearch] = true
	return newSearch
}

func (searcher *searcher) Ack(reply *packets_gossiper.SearchReply) {
	searcher.RLock()
	defer searcher.RUnlock()

	for search := range searcher.searches {
		search.Ack(reply)
	}
}

func (searcher *searcher) GetFullSearches() []*Search {
	searcher.RLock()
	defer searcher.RUnlock()

	searches := []*Search{}
	for s := range searcher.searches {
		if s.IsFull() {
			searches = append(searches, s)
		}
	}
	return searches
}
