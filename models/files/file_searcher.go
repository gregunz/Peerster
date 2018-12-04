package files

import "sync"

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
	newSearch := NewSearch(keywords, initialBudget)
	searcher.searches = append(searcher.searches, newSearch)
	return newSearch
}
