package origin_handlers

import (
	"github.com/gregunz/Peerster/models/packets/packets_gossiper"
	"sync"
)

type OriginsHandler struct {
	MyOrigin        string
	Map             map[string]*OriginHandlerElem
	latestRumorChan chan *packets_gossiper.RumorMessage
	mux             sync.Mutex
	vectorClockMux  sync.Mutex
}

func NewOriginToHandlers(myOrigin string) *OriginsHandler {
	return &OriginsHandler{
		MyOrigin:        myOrigin,
		Map:             map[string]*OriginHandlerElem{},
		latestRumorChan: make(chan *packets_gossiper.RumorMessage, 1),
	}
}

func (originsHandler *OriginsHandler) ToVectorClock() VectorClock {
	return &ProtoVectorClock{
		originsHandler:  originsHandler,
		latestRumorChan: originsHandler.latestRumorChan,
		mux:             originsHandler.vectorClockMux,
	}
}

func (originsHandler *OriginsHandler) ToRoutingTable() RoutingTable {
	return &ProtoRoutingTable{
		originsHandler: originsHandler,
	}
}

func (originsHandler *OriginsHandler) getOrCreateHandler(origin string) *OriginHandlerElem {
	h, ok := originsHandler.Map[origin]
	if !ok {
		h = originsHandler.NewHandler(origin)
		originsHandler.Map[origin] = h
	}
	return h
}

func (originsHandler *OriginsHandler) GetOrCreateHandler(origin string) *OriginHandlerElem {
	originsHandler.mux.Lock()
	defer originsHandler.mux.Unlock()

	return originsHandler.getOrCreateHandler(origin)
}

func (originsHandler *OriginsHandler) NewHandler(origin string) *OriginHandlerElem {
	return &OriginHandlerElem{
		saveRumor:      NewRumorHandler(origin, originsHandler.latestRumorChan),
		routingHandler: NewRoutingHandler(origin),
	}
}
