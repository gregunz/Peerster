package files

import (
	"github.com/gregunz/Peerster/models/packets/packets_gossiper"
	"sync"
)

type Searcher interface {
	Search(keywords []string, initialBudget uint64) *Search
}

type searcher struct {
	searches []*Search

	sync.RWMutex
}

func NewSearcher() *searcher {
	return &searcher{
		searches: []*Search{},
	}
}

func (searcher *searcher) Search(keywords []string, initialBudget uint64) *Search {
	searcher.Lock()
	defer searcher.Unlock()

	newSearch := newSearch(keywords, initialBudget)
	searcher.searches = append(searcher.searches, newSearch)
	return newSearch
}

func (searcher *searcher) Ack(reply *packets_gossiper.SearchReply) {
	searcher.RLock()
	defer searcher.RUnlock()

	for _, search := range searcher.searches {
		search.Ack(reply)
	}
}
