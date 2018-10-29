package www

type client struct {
	IsSubscribedToMessage bool
	IsSubscribedToNode    bool
}

func NewClient() *client {
	return &client{
		IsSubscribedToMessage: false,
		IsSubscribedToNode:    false,
	}
}
