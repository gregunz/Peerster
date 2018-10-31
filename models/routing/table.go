package routing

import (
	"sync"
)

type Table struct {
	myOrigin   string
	handlers   map[string]*tableHandler
	originChan OriginChan
	mux        sync.Mutex
}

func NewTable(myOrigin string, originChan OriginChan) *Table {
	return &Table{
		myOrigin:   myOrigin,
		handlers:   map[string]*tableHandler{},
		originChan: originChan,
	}
}

func (table *Table) getOrCreateHandler(origin string) *tableHandler {
	h, ok := table.handlers[origin]
	if !ok {
		h = newRoutingTableHandler(origin)
		table.handlers[origin] = h
		if table.myOrigin != origin {
			table.originChan.AddOrigin(origin)
		}
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
