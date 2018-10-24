package gossiper

type GossiperMode struct {
	simple bool
}

func NewSimpleMode() *GossiperMode {
	return &GossiperMode{
		simple: true,
	}
}

func NewDefaultMode() *GossiperMode {
	return &GossiperMode{
		simple: false,
	}
}

func (mode *GossiperMode) isSimple() bool {
	return mode.simple
}

func (mode *GossiperMode) isDefault() bool {
	return !mode.simple
}
