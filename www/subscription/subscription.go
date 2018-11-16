package subscription

type Sub int

const (
	Message Sub = 1 + iota
	Node
	Origin
	File
)

func (sub Sub) Name() string {
	switch sub {

	case Message:
		return "message"
	case Node:
		return "node"
	case Origin:
		return "origin"
	case File:
		return "files"
	default:
		return "<unnamed subscription>"

	}
}
