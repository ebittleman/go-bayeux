package bayeux

import (
	"encoding/json"
	"log"
	"net/http"
	"sync"

	"code.google.com/p/go.net/websocket"

	"github.com/ebittleman/go-bayeux/channel"
	"github.com/ebittleman/go-bayeux/messages"
)

type baseClient struct {
	id           string
	server       Server
	channels     map[string]channel.Channel
	channelsLock *sync.Mutex
	responses    chan interface{}
	done         chan struct{}
	logger       *log.Logger
}

type websocketClient struct {
	ws *websocket.Conn
	baseClient
}

type Client interface {
	Close() error
	Wait()
	OnMessage(string, []byte)
	SendMessage(messages.Message)
	GetLogger() *log.Logger
	channel.Subscriber
}

func (c *baseClient) GetId() string {
	return c.id
}

func (c *baseClient) Subscribe(ch channel.Channel) {

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

func (c *baseClient) Publish(ch channel.Channel, msg messages.Message) {
	go ch.Publish(msg)
}

func (c *baseClient) Unsubscribe(ch channel.Channel) {
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

func (c *baseClient) Wait() {}

func (c *baseClient) Close() error {

	for _, channel := range c.channels {
		c.Unsubscribe(channel)
	}

	c.server.UnregisterClient(c.GetId())

	return nil
}

func (c *baseClient) GetLogger() *log.Logger {
	return c.logger
}

func (c *baseClient) OnMessage(ch string, payload []byte) {
	c.server.OnReceiveMessage(ch, c.GetId(), payload)
}

func (c *baseClient) SendMessage(msg messages.Message) {
	go func(c *baseClient, msg messages.Message) {
		c.responses <- msg
	}(c, msg)
}

func NewClient(id string, ws *websocket.Conn, server Server) Client {
	client := &websocketClient{
		ws,
		baseClient{
			id, server,
			make(map[string]channel.Channel),
			&sync.Mutex{},
			make(chan interface{}),
			make(chan struct{}),
			logger,
		},
	}

	go client.IncomingLoop()
	go client.OutgoingLoop()

	return client
}

func (c *websocketClient) Wait() {
	<-c.done
}

func (c *websocketClient) Close() error {

	err := c.baseClient.Close()

	if err != nil {
		return err
	}

	c.GetLogger().Println("Client Disconnected")
	c.ws.Close()

	if c.done == nil {
		return nil
	}

	close(c.done)

	return nil
}

func (c *websocketClient) IncomingLoop() {
	for {
		err := c.ReceiveMesages()
		if err != nil {
			c.Close()
			return
		}
	}
}

func (c *websocketClient) OutgoingLoop() {
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

func (c *websocketClient) ReceiveMesages() error {
	msgs := make([]map[string]interface{}, 0)
	err := websocket.JSON.Receive(c.ws, &msgs)
	for _, msg := range msgs {
		payload, _ := json.Marshal(msg)
		channel := msg["channel"]
		go c.OnMessage(channel.(string), payload)
	}

	return err
}

type longPollClient struct {
	resp     http.ResponseWriter
	req      *http.Request
	respChan chan http.ResponseWriter
	baseClient
}

func NewLongPollClient(id string, resp http.ResponseWriter, req *http.Request, server Server) Client {
	client := &longPollClient{
		resp,
		req,
		make(chan http.ResponseWriter),
		baseClient{
			id,
			server,
			make(map[string]channel.Channel),
			&sync.Mutex{},
			make(chan interface{}),
			make(chan struct{}),
			logger,
		},
	}

	// go client.IncomingLoop()
	go client.OutgoingLoop()
	go func(c *longPollClient, resp http.ResponseWriter) {
		c.respChan <- resp
	}(client, resp)

	return client
}

func (c *longPollClient) OutgoingLoop() {
	for {
		resp := <-c.respChan
		select {
		case msg := <-c.responses:
			output, err := json.Marshal(&[]interface{}{msg})
			if err != nil {
				http.Error(resp, "Error Formatting Response", 406)
				c.Close()
				return
			}
			resp.Write(output)
		case <-c.done:
			return
		}
	}
}
