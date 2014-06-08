package bayeux

import (
	"encoding/json"

	"code.google.com/p/go.net/websocket"
)

type webserverClient struct {
	id        string
	ws        *websocket.Conn
	server    Server
	responses chan interface{}
	done      chan struct{}
}

type Client interface {
	GetId() string
	Close() error
	Wait()
	OnMessage(string, []byte)
	SendMessage(interface{})
}

func NewClient(id string, ws *websocket.Conn, server Server) Client {
	client := &webserverClient{id, ws, server, make(chan interface{}), make(chan struct{})}
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
	c.ws.Close()
	if c.done == nil {
		return nil
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
func (c *webserverClient) SendMessage(msg interface{}) {
	go func(c *webserverClient, msg interface{}) {
		c.responses <- msg
	}(c, msg)
}
