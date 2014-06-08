package channel

import (
	"sync"

	"github.com/ebittleman/go-bayeux/messages"
)

type MessageHandler func(messages.Message)

type channel struct {
	name          string
	subscriptions map[string]MessageHandler
	subscribers   map[string]Subscriber
	lock          *sync.Mutex
}

type Channel interface {
	GetName() string
	AddSubscription(Subscriber, MessageHandler)
	RemoveSubscription(Subscriber)
	Publish(messages.Message)
}

type Subscriber interface {
	GetId() string
	Subscribe(Channel)
	Unsubscribe(Channel)
	Publish(Channel, messages.Message)
}

func NewChannel(name string) Channel {
	return &channel{name, make(map[string]MessageHandler), make(map[string]Subscriber), &sync.Mutex{}}
}

func (c *channel) GetName() string {
	return c.name
}

func (c *channel) AddSubscription(subscriber Subscriber, m MessageHandler) {
	c.lock.Lock()

	_, ok := c.subscriptions[subscriber.GetId()]
	if ok {
		c.lock.Unlock()
		return
	}

	c.subscriptions[subscriber.GetId()] = m
	c.subscribers[subscriber.GetId()] = subscriber
	c.lock.Unlock()

	subscriber.Subscribe(c)
}
func (c *channel) RemoveSubscription(subscriber Subscriber) {
	c.lock.Lock()

	_, ok := c.subscriptions[subscriber.GetId()]
	if !ok {
		c.lock.Unlock()
		return
	}

	delete(c.subscriptions, subscriber.GetId())
	delete(c.subscribers, subscriber.GetId())
	c.lock.Unlock()

	subscriber.Unsubscribe(c)
}

func (c *channel) Publish(m messages.Message) {
	c.lock.Lock()
	for _, messageHandler := range c.subscriptions {
		go messageHandler(m)
	}
	c.lock.Unlock()
}
