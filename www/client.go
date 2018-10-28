package www

type client struct {
	IsSubscribedToMessage bool
}

func NewClient() *client {
	return &client{
		IsSubscribedToMessage: false, // default for now is to be subscribed
	}
}
