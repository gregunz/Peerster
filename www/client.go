package www

type client struct {
	IsSubscribedToMessage bool
	IsSubscribedToNode    bool
	IsSubscribedToOrigin  bool
	IsSubscribedToFiles   bool
}

func NewClient() *client {
	return &client{
		IsSubscribedToMessage: false,
		IsSubscribedToNode:    false,
		IsSubscribedToOrigin:  false,
		IsSubscribedToFiles:   false,
	}
}
