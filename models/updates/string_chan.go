package updates

type StringChan interface {
	Get() string
	Push(s string)
}

type stringChan struct {
	Chan
}

func NewStringChan(activated bool) StringChan {
	return &stringChan{Chan: NewChan(activated)}
}

func (ch *stringChan) Push(s string) {
	ch.Chan.Push(s)
}

func (ch *stringChan) Get() string {
	s, ok := ch.Chan.Get().(string)
	if !ok {
		return ""
	}
	return s
}
