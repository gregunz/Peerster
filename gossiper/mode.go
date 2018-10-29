package gossiper

type Mode struct {
	simple bool
}

func NewSimpleMode() *Mode {
	return &Mode{
		simple: true,
	}
}

func NewDefaultMode() *Mode {
	return &Mode{
		simple: false,
	}
}

func (mode *Mode) isSimple() bool {
	return mode.simple
}

func (mode *Mode) isDefault() bool {
	return !mode.simple
}
