package files

import (
	"github.com/gregunz/Peerster/models/packets/packets_gossiper"
	"sync"
)

type Searcher struct {
	searchList            []*Search
	finishedSearchIndices map[int]bool
	filesFound            map[string]*SearchMetadata
	MatchChan             ReachableFileChan

	sync.RWMutex
}

func NewSearcher(activateChan bool) *Searcher {
	return &Searcher{
		searchList:            []*Search{},
		finishedSearchIndices: map[int]bool{},
		filesFound:            map[string]*SearchMetadata{},
		MatchChan:             NewMatchChan(activateChan),
	}
}

func (searcher *Searcher) Search(keywords []string, initialBudget uint64) *Search {
	searcher.Lock()
	defer searcher.Unlock()

	newSearch := newSearch(keywords, initialBudget, searcher.MatchChan)
	searcher.searchList = append(searcher.searchList, newSearch)
	return newSearch
}

func (searcher *Searcher) Ack(reply *packets_gossiper.SearchReply) {
	searcher.RLock()
	defer searcher.RUnlock()

	for _, search := range searcher.searchList {
		newFilesFound := search.Ack(reply)
		for _, newFile := range newFilesFound {
			for _, metadata := range newFile.ToSearchMetadata() {
				if _, ok := searcher.filesFound[metadata.Filename]; !ok {
					searcher.filesFound[metadata.Filename] = metadata
					searcher.MatchChan.Push(metadata)
				}
			}
		}
	}
}

func (searcher *Searcher) GetAllSearches() []*Search {
	searcher.RLock()
	defer searcher.RUnlock()

	searches := []*Search{}
	for _, s := range searcher.searchList {
		searches = append(searches, s)
	}
	return searches
}

func (searcher *Searcher) GetAllMetadata() []*SearchMetadata {
	searcher.RLock()
	defer searcher.RUnlock()

	metadataList := []*SearchMetadata{}
	for _, s := range searcher.filesFound {
		metadataList = append(metadataList, s)
	}
	return metadataList
}

func (searcher *Searcher) GetLatestFullSearches() []*Search {
	searcher.RLock()
	defer searcher.RUnlock()

	searches := []*Search{}
	for idx, s := range searcher.searchList {
		if ok := searcher.finishedSearchIndices[idx]; !ok && s.IsFull() {
			searcher.finishedSearchIndices[idx] = true
			searches = append(searches, s)
		}
	}
	return searches
}
