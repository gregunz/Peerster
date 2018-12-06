package files

import (
	"github.com/gregunz/Peerster/models/packets/packets_gossiper"
	"sync"
)

type Searcher struct {
	searches  map[*Search]bool
	MatchChan ReachableFileChan

	sync.RWMutex
}

func NewSearcher(activateChan bool) *Searcher {
	return &Searcher{
		searches:  map[*Search]bool{},
		MatchChan: NewMatchChan(activateChan),
	}
}

func (searcher *Searcher) Search(keywords []string, initialBudget uint64) *Search {
	searcher.Lock()
	defer searcher.Unlock()

	newSearch := newSearch(keywords, initialBudget, searcher.MatchChan)
	searcher.searches[newSearch] = true
	return newSearch
}

func (searcher *Searcher) Ack(reply *packets_gossiper.SearchReply) {
	searcher.RLock()
	defer searcher.RUnlock()

	for search := range searcher.searches {
		search.Ack(reply)
	}
}

func (searcher *Searcher) GetAllSearches() []*Search {
	searcher.RLock()
	defer searcher.RUnlock()

	searches := []*Search{}
	for s := range searcher.searches {
		searches = append(searches, s)
	}
	return searches
}

func (searcher *Searcher) GetFullSearches() []*Search {
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
