package bayeux

import (
	"bytes"
	"crypto/md5"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"math/rand"
	"net/http"
	"sync"
	"time"

	"code.google.com/p/go.net/websocket"

	"github.com/ebittleman/go-bayeux/channel"
	"github.com/ebittleman/go-bayeux/messages"
)

type BayeuxHandler func(msg messages.RawMessage)

type bayeuxServer struct {
	channelHandlers       map[string]BayeuxHandler
	channels              map[string]channel.Channel
	clients               map[string]Client
	clientMutex           *sync.Mutex
	channelHandlerslMutex *sync.Mutex
	channelsMutex         *sync.Mutex
	websocketHandler      http.Handler
	incomingCh            chan messages.RawMessage
	done                  chan struct{}

	logger *log.Logger
}

type Server interface {
	// Lifecycle
	ServeHTTP(http.ResponseWriter, *http.Request)
	HandleFunc(string, BayeuxHandler)
	GetHandler(string) BayeuxHandler
	RegisterClient(string, Client)
	UnregisterClient(string) error
	GetClient(string) Client
	OnReceiveMessage(string, string, []byte)
	Close() error

	Publish(string, messages.Message)

	GetLogger() *log.Logger
}

func (s *bayeuxServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.websocketHandler.ServeHTTP(w, r)
}

func Handler() Server {

	server := &bayeuxServer{
		make(map[string]BayeuxHandler),
		make(map[string]channel.Channel),
		make(map[string]Client),
		&sync.Mutex{},
		&sync.Mutex{},
		&sync.Mutex{},
		nil,
		make(chan messages.RawMessage),
		make(chan struct{}),
		logger,
	}

	server.HandleFunc("/meta/handshake", func(msg messages.RawMessage) {
		handshakeRequest := &messages.HandshakeRequest{}
		json.Unmarshal(msg.Payload, handshakeRequest)
		Handshake(server, msg.ClientId, handshakeRequest)
	})

	server.HandleFunc("/meta/disconnect", func(msg messages.RawMessage) {
		disconnectRequest := &messages.DisconnectRequest{}
		json.Unmarshal(msg.Payload, disconnectRequest)
		Disconnect(server, disconnectRequest)
	})

	server.HandleFunc("/meta/connect", func(msg messages.RawMessage) {
		connectRequest := &messages.ConnectRequest{}
		json.Unmarshal(msg.Payload, connectRequest)
		Connect(server, connectRequest)
	})

	server.HandleFunc("/meta/subscribe", func(msg messages.RawMessage) {
		subscribeRequest := &messages.SubscribeRequest{}
		json.Unmarshal(msg.Payload, subscribeRequest)
		server.HandleSubscribe(subscribeRequest)
	})

	server.HandleFunc("/meta/unsubscribe", func(msg messages.RawMessage) {
		subscribeResponse := &messages.SubscribeResponse{}
		json.Unmarshal(msg.Payload, subscribeResponse)
		server.HandleUnsubscribe(subscribeResponse)
	})

	server.websocketHandler = websocket.Handler(func(ws *websocket.Conn) {
		client := NewClient(GenerateNewClientId(), ws, server)
		server.RegisterClient(client.GetId(), client)
		client.Wait()
	})

	go server.Loop()
	return server
}

func ParseReqBody(body io.Reader, v interface{}) error {
	buf := bytes.Buffer{}
	buf.ReadFrom(body)
	return json.Unmarshal(buf.Bytes(), v)
}

func (bs *bayeuxServer) HandleFunc(channel string, handleFunc BayeuxHandler) {
	bs.channelHandlerslMutex.Lock()
	bs.channelHandlers[channel] = handleFunc
	bs.channelHandlerslMutex.Unlock()
}

func (bs *bayeuxServer) GetHandler(channel string) BayeuxHandler {
	var (
		handleFunc BayeuxHandler
		ok         bool
	)

	bs.channelHandlerslMutex.Lock()
	handleFunc, ok = bs.channelHandlers[channel]
	bs.channelHandlerslMutex.Unlock()

	if !ok {
		bs.channelsMutex.Lock()
		_, ok := bs.channels[channel]
		if ok {
			handleFunc = GeneratePublicMesaageHandler(bs)
		}
		bs.channelsMutex.Unlock()
	}

	return handleFunc
}

func (bs *bayeuxServer) RegisterClient(id string, client Client) {
	bs.clientMutex.Lock()
	bs.clients[id] = client
	bs.clientMutex.Unlock()
}
func (bs *bayeuxServer) UnregisterClient(id string) error {
	bs.clientMutex.Lock()
	delete(bs.clients, id)
	bs.clientMutex.Unlock()

	return nil
}

func (bs *bayeuxServer) GetClient(id string) Client {
	var client Client

	bs.clientMutex.Lock()
	client = bs.clients[id]
	bs.clientMutex.Unlock()

	return client
}

func (bs *bayeuxServer) OnReceiveMessage(channel string, clientId string, payload []byte) {
	bs.incomingCh <- messages.RawMessage{channel, clientId, payload}
}

func (bs *bayeuxServer) Loop() {
	for {
		select {
		case msg := <-bs.incomingCh:
			go RouteIncomingMsg(bs, msg)
		case <-bs.done:
			return
		}
	}
}

func (bs *bayeuxServer) Publish(channelPath string, msg messages.Message) {
	bs.channelsMutex.Lock()
	ch, ok := bs.channels[channelPath]
	if !ok {
		fmt.Printf("Channel Not Found '%s'\n", channelPath)
		return
	}
	bs.channelsMutex.Unlock()
	ch.Publish(msg)
}

func (bs *bayeuxServer) Close() error {
	if bs.done == nil {
		return nil
	}
	close(bs.done)
	bs.done = nil
	return nil
}

func (bs *bayeuxServer) GetLogger() *log.Logger {
	return bs.logger
}

func RouteIncomingMsg(bs Server, msg messages.RawMessage) {
	handler := bs.GetHandler(msg.Channel)

	if handler == nil {
		fmt.Printf("Not Handler for Channel '%s'\n", msg.Channel)
		return
	}

	handler(msg)
}

func GenerateNewClientId() string {
	rand.Seed(time.Now().UnixNano())
	id := fmt.Sprintf("%d", rand.Int63())
	h := md5.New()

	h.Write([]byte(id))

	return fmt.Sprintf("%x", h.Sum(nil))
}

func Handshake(bs Server, ClientId string, msg *messages.HandshakeRequest) {
	client := bs.GetClient(ClientId)
	if client == nil {
		panic("can't handshake someone who does not exist?")
	}

	bs.GetLogger().Printf("Do Handshake\n%v\n", msg)
	bs.GetLogger().Printf("For Client\n%v\n", client)

	client.SendMessage(&messages.HandshakeResponse{
		msg.Channel,
		"1.0",
		"1.0",
		supportedClients,
		ClientId,
		true,
		true,
		"",
		msg.Id,
		&messages.HandshakeResponseAdvice{RECONNECT_RETRY, 0},
	})

}

func Disconnect(bs Server, msg *messages.DisconnectRequest) {
	client := bs.GetClient(msg.ClientId)
	if client == nil {
		return
	}

	bs.GetLogger().Printf("Do Dissconnect\n%v\n", msg)
	bs.GetLogger().Printf("For Client\n%v\n", client)

	client.Close()

	client.SendMessage(&messages.DisconnectResponse{
		msg.Channel,
		msg.ClientId,
		true,
		msg.Id,
	})
}

func Connect(bs Server, msg *messages.ConnectRequest) {
	client := bs.GetClient(msg.ClientId)
	if client == nil {
		panic("can't connect someone who does not exist?")
	}

	bs.GetLogger().Printf("Do Connect\n%v\n", msg)
	bs.GetLogger().Printf("For Client\n%v\n", client)

	//TODO Implement Connect

	client.SendMessage(&messages.ConnectResponse{
		msg.Channel,
		true,
		"",
		msg.ClientId,
		NewTimestamp().String(),
		msg.Id,
		&messages.ConnectAdvice{RECONNECT_HANDSHAKE, 120000},
	})
}

func (bs *bayeuxServer) HandleSubscribe(msg *messages.SubscribeRequest) {
	client := bs.GetClient(msg.ClientId)
	if client == nil {
		panic("can't handshake someone who does not exist?")
	}
	bs.GetLogger().Printf("Do Subscribe\n%v\n", msg)
	bs.GetLogger().Printf("For Client\n%v\n", client)

	bs.channelsMutex.Lock()
	ch, ok := bs.channels[msg.Subscription]
	if !ok {
		ch = channel.NewChannel(msg.Subscription)
		bs.channels[msg.Subscription] = ch
	}
	bs.channelsMutex.Unlock()

	client.Subscribe(ch)

	client.SendMessage(&messages.SubscribeResponse{
		msg.Channel,
		msg.ClientId,
		msg.Subscription,
		true,
		"",
		NewTimestamp().String(),
		msg.Id,
	})

	ch.Publish(&messages.EventMessage{ch.GetName(), struct {
		Msg string `json:"msg"`
	}{"Welcome to " + ch.GetName() + " Client '" + client.GetId() + "'"}, ""})
}

func (bs *bayeuxServer) HandleUnsubscribe(msg *messages.SubscribeResponse) {
	client := bs.GetClient(msg.ClientId)
	if client == nil {
		panic("can't publish someone whois not connected")
	}

	bs.GetLogger().Printf("Unsubscribe Message\n%v\n", msg)
	bs.GetLogger().Printf("For Client\n%v\n", client)

	bs.channelsMutex.Lock()
	ch := bs.channels[msg.Subscription]
	bs.channelsMutex.Unlock()

	if ch != nil {
		client.Unsubscribe(ch)
	}

	client.SendMessage(&messages.UnsubscribeResponse{
		msg.Channel,
		client.GetId(),
		msg.Subscription,
		true,
		"",
		NewTimestamp().String(),
		msg.Id,
	})
}

func GeneratePublicMesaageHandler(bs Server) BayeuxHandler {
	return func(msg messages.RawMessage) {
		PublicMessage(bs, msg)
	}
}

func PublicMessage(bs Server, msg messages.RawMessage) {
	client := bs.GetClient(msg.ClientId)
	if client == nil {
		panic("can't publish someone whois not connected")
	}

	bs.GetLogger().Printf("Publish Message\n%v\n", msg)
	bs.GetLogger().Printf("For Client\n%v\n", client)

	payload := make(map[string]interface{})
	json.Unmarshal(msg.Payload, &payload)

	Id, ok := payload["id"]

	if !ok {
		Id = interface{}("")
	}

	client.SendMessage(&messages.PublishResponse{
		msg.Channel,
		true,
		"",
		Id.(string),
	})

	bs.Publish(msg.Channel, payload)
}
