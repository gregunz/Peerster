package origin_handlers

type OriginHandlerElem struct {
	saveRumor      *saveRumorHandler
	routingHandler *routingHandler
}

func NewHandler(origin string) *OriginHandlerElem {
	return &OriginHandlerElem{
		saveRumor:      NewRumorHandler(origin),
		routingHandler: NewRoutingHandler(origin),
	}
}
