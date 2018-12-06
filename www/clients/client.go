package clients

import (
	"github.com/gregunz/Peerster/www/subscription"
	"sync"
)

type Client struct {
	subscriptions map[subscription.Sub]bool

	sync.RWMutex
}

func NewClient() *Client {
	return &Client{
		subscriptions: map[subscription.Sub]bool{},
	}
}

func (c *Client) IsSubscribedTo(sub subscription.Sub) bool {
	c.RLock()
	defer c.RUnlock()
	v, ok := c.subscriptions[sub]
	if ok {
		return v
	}
	return false // default: not being subscribed
}

func (c *Client) SetSubscriptionTo(sub subscription.Sub, v bool) {
	c.Lock()
	defer c.Unlock()
	c.subscriptions[sub] = v
}
