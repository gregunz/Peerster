package routing

type OriginChan interface {
	AddOrigin(string)
	GetOrigin() string
}
