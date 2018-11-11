package updates

type StringChan interface {
	Get() string
	Push(filename string)
}

type stringChan struct {
	ch Chan
}

func NewStringChan(activated bool) StringChan {
	return &stringChan{ch: NewChan(activated)}
}

func (ch *stringChan) Push(s string) {
	ch.ch.Push(s)
}

func (ch *stringChan) Get() string {
	s, ok := ch.ch.Get().(string)
	if !ok {
		return ""
	}
	return s
}
