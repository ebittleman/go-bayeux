package bayeux

import (
	"crypto/md5"
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"os"
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
	incomingCh            chan messages.RawMessage
	done                  chan struct{}

	logger *log.Logger
}

type Server interface {
	// Lifecycle
	Bind(string)
	HandleFunc(string, BayeuxHandler)
	GetHandler(string) BayeuxHandler
	RegisterClient(string, Client)
	GetClient(string) Client
	OnReceiveMessage(string, string, []byte)
	Close() error

	GetLogger() *log.Logger
}

func NewServer() Server {
	logger := log.New(os.Stdout, "go-bayux/server::", log.Ldate|log.Ltime)

	server := &bayeuxServer{
		make(map[string]BayeuxHandler),
		make(map[string]channel.Channel),
		make(map[string]Client),
		&sync.Mutex{},
		&sync.Mutex{},
		&sync.Mutex{},
		make(chan messages.RawMessage),
		make(chan struct{}),
		logger,
	}

	server.HandleFunc("/meta/handshake", func(msg messages.RawMessage) {
		handshakeRequest := &messages.HandshakeRequest{}
		json.Unmarshal(msg.Payload, handshakeRequest)
		Handshake(server, msg.ClientId, handshakeRequest)
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

	go server.Loop()
	return server
}

func (bs *bayeuxServer) Bind(path string) {
	http.Handle(path, websocket.Handler(func(ws *websocket.Conn) {
		client := NewClient(GenerateNewClientId(), ws, bs)
		bs.RegisterClient(client.GetId(), client)
		client.Wait()
	}))
}

func (bs *bayeuxServer) HandleFunc(channel string, handleFunc BayeuxHandler) {
	bs.channelHandlerslMutex.Lock()
	bs.channelHandlers[channel] = handleFunc
	bs.channelHandlerslMutex.Unlock()
}

func (bs *bayeuxServer) GetHandler(channel string) BayeuxHandler {
	var handleFunc BayeuxHandler

	bs.channelHandlerslMutex.Lock()
	handleFunc = bs.channelHandlers[channel]
	bs.channelHandlerslMutex.Unlock()

	return handleFunc
}

func (bs *bayeuxServer) RegisterClient(id string, client Client) {
	bs.clientMutex.Lock()
	bs.clients[id] = client
	bs.clientMutex.Unlock()
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
		fmt.Println("msg received")
	}
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
		// nil,
		&messages.HandshakeResponseAdvice{RECONNECT_RETRY, 0},
		// &messages.HandshakeResponseAdvice{RECONNECT_RETRY, defaultInterval},
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
		&messages.ConnectAdvice{RECONNECT_RETRY, defaultInterval},
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
	}{"Welcome to " + ch.GetName() + " Client '" + client.GetId() + "'"}, "66"})
}

func PublicMessage(bs Server, msg messages.RawMessage) {
}
