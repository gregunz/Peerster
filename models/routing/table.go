package routing

import (
	"sync"
)

type table struct {
	myOrigin   string
	handlers   map[string]*tableHandler
	originChan OriginChan
	mux        sync.Mutex
}

type Table interface {
	GetOrCreateHandler(origin string) *tableHandler
	GetOrigins() []string
}

func NewTable(myOrigin string, originChan OriginChan) *table {
	return &table{
		myOrigin:   myOrigin,
		handlers:   map[string]*tableHandler{},
		originChan: originChan,
	}
}

func (table *table) getOrCreateHandler(origin string) *tableHandler {
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

func (table *table) GetOrCreateHandler(origin string) *tableHandler {
	table.mux.Lock()
	defer table.mux.Unlock()

	return table.getOrCreateHandler(origin)
}

func (table *table) GetOrigins() []string {
	table.mux.Lock()
	defer table.mux.Unlock()

	ls := []string{}
	for o := range table.handlers {
		ls = append(ls, o)
	}
	return ls
}
