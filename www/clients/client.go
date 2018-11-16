package clients

import (
	"github.com/gregunz/Peerster/www/subscription"
	"sync"
)

type client struct {
	subscriptions map[subscription.Sub]bool
	mux           sync.RWMutex
}

type Client interface {
	IsSubscribedTo(subscription.Sub) bool
	SetSubscriptionTo(subscription.Sub, bool)
}

func NewClient() Client {
	return &client{
		subscriptions: map[subscription.Sub]bool{},
	}
}

func (c *client) IsSubscribedTo(sub subscription.Sub) bool {
	c.mux.RLock()
	defer c.mux.RLock()
	v, ok := c.subscriptions[sub]
	if ok {
		return v
	}
	return false // default: not being subscribed
}

func (c *client) SetSubscriptionTo(sub subscription.Sub, v bool) {
	c.mux.Lock()
	defer c.mux.Unlock()
	c.subscriptions[sub] = v
}
