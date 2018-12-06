package routing

type OriginChan interface {
	Get() string
	Push(origin string)
}
