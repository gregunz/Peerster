package routing

import (
	"github.com/gregunz/Peerster/models/updates"
	"sync"
)

type Table struct {
	myOrigin   string
	handlers   map[string]*tableHandler
	OriginChan OriginChan
	mux        sync.Mutex
}

func NewTable(myOrigin string, activateChan bool) *Table {
	return &Table{
		myOrigin:   myOrigin,
		handlers:   map[string]*tableHandler{},
		OriginChan: updates.NewStringChan(activateChan),
	}
}

func (table *Table) getOrCreateHandler(origin string) *tableHandler {
	if table.myOrigin == origin {
		return nil
	}
	h, ok := table.handlers[origin]
	if !ok {
		h = newRoutingTableHandler(origin)
		table.handlers[origin] = h
	}
	return h
}

func (table *Table) GetOrCreateHandler(origin string) *tableHandler {
	table.mux.Lock()
	defer table.mux.Unlock()

	return table.getOrCreateHandler(origin)
}

func (table *Table) GetOrigins() []string {
	table.mux.Lock()
	defer table.mux.Unlock()

	ls := []string{}
	for o := range table.handlers {
		ls = append(ls, o)
	}
	return ls
}
