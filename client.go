package bayeux

import (
	"encoding/json"
	"log"
	"os"
	"sync"

	"code.google.com/p/go.net/websocket"

	"github.com/ebittleman/go-bayeux/channel"
	"github.com/ebittleman/go-bayeux/messages"
)

type webserverClient struct {
	id           string
	ws           *websocket.Conn
	server       Server
	channels     map[string]channel.Channel
	channelsLock *sync.Mutex
	responses    chan interface{}
	done         chan struct{}
	logger       *log.Logger
}

type Client interface {
	Close() error
	Wait()
	OnMessage(string, []byte)
	SendMessage(messages.Message)
	GetLogger() *log.Logger
	channel.Subscriber
}

func NewClient(id string, ws *websocket.Conn, server Server) Client {

	logger := log.New(os.Stdout, "go-bayux/client::", log.Ldate|log.Ltime)

	client := &webserverClient{
		id,
		ws,
		server,
		make(map[string]channel.Channel),
		&sync.Mutex{},
		make(chan interface{}),
		make(chan struct{}),
		logger,
	}

	go client.IncomingLoop()
	go client.OutgoingLoop()

	return client
}

func (c *webserverClient) GetId() string {
	return c.id
}

func (c *webserverClient) Wait() {
	<-c.done
}

func (c *webserverClient) Close() error {
	c.GetLogger().Println("Client Disconnected")
	c.ws.Close()
	if c.done == nil {
		return nil
	}

	for _, channel := range c.channels {
		c.Unsubscribe(channel)
	}

	close(c.done)
	return nil
}

func (c *webserverClient) IncomingLoop() {
	for {
		err := c.ReceiveMesages()
		if err != nil {
			c.Close()
			return
		}
	}
}

func (c *webserverClient) OutgoingLoop() {
	for {
		select {
		case msg := <-c.responses:
			err := websocket.JSON.Send(c.ws, &[]interface{}{msg})
			if err != nil {
				c.Close()
				return
			}
		case <-c.done:
			return
		}
	}
}

func (c *webserverClient) ReceiveMesages() error {
	msgs := make([]map[string]interface{}, 0)
	err := websocket.JSON.Receive(c.ws, &msgs)
	for _, msg := range msgs {
		payload, _ := json.Marshal(msg)
		channel := msg["channel"]
		go c.OnMessage(channel.(string), payload)
	}

	return err
}

func (c *webserverClient) OnMessage(ch string, payload []byte) {
	c.server.OnReceiveMessage(ch, c.GetId(), payload)
}

func (c *webserverClient) SendMessage(msg messages.Message) {
	go func(c *webserverClient, msg messages.Message) {
		c.responses <- msg
	}(c, msg)
}

func (c *webserverClient) GetLogger() *log.Logger {
	return c.logger
}

func (c *webserverClient) Subscribe(ch channel.Channel) {

	c.channelsLock.Lock()

	_, ok := c.channels[ch.GetName()]
	if ok {
		c.channelsLock.Unlock()
		return
	}

	c.channels[ch.GetName()] = ch
	c.channelsLock.Unlock()

	ch.AddSubscription(c, func(msg messages.Message) {
		c.SendMessage(msg)
	})
}

func (c *webserverClient) Publish(ch channel.Channel, msg messages.Message) {
	go ch.Publish(msg)
}

func (c *webserverClient) Unsubscribe(ch channel.Channel) {
	c.channelsLock.Lock()

	_, ok := c.channels[ch.GetName()]
	if !ok {
		c.channelsLock.Unlock()
		return
	}

	delete(c.channels, ch.GetName())
	c.channelsLock.Unlock()

	ch.RemoveSubscription(c)
}
