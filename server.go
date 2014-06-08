package bayeux

import (
	"crypto/md5"
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"sync"
	"time"

	"code.google.com/p/go.net/websocket"
)

type bayeuxServer struct {
	channels     map[string]BayeuxHandler
	clients      map[string]Client
	clientMutex  *sync.Mutex
	channelMutex *sync.Mutex
	incomingCh   chan Message
	done         chan struct{}
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

	// Server Attributes
	// ProtocolVersion() string
	// MinimumProtocolVersion() string
	// GetSupportedConnectionTypes() []string

	// Client Communication
	// Handshake()
	// Connect()
	// On(Request)
	// Emit(Response)
}

func NewServer() Server {
	server := &bayeuxServer{
		make(map[string]BayeuxHandler),
		make(map[string]Client),
		&sync.Mutex{},
		&sync.Mutex{},
		make(chan Message),
		make(chan struct{}),
	}

	server.HandleFunc("/meta/handshake", func(msg *MessageStruct) {
		Handshake(server, msg)
	})

	server.HandleFunc("/meta/connect", func(msg *MessageStruct) {
		Connect(server, msg)
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
	bs.channelMutex.Lock()
	bs.channels[channel] = handleFunc
	bs.channelMutex.Unlock()
}

func (bs *bayeuxServer) GetHandler(channel string) BayeuxHandler {
	var handleFunc BayeuxHandler

	bs.channelMutex.Lock()
	handleFunc = bs.channels[channel]
	bs.channelMutex.Unlock()

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
	bs.incomingCh <- message{channel, clientId, payload}
}

func (bs *bayeuxServer) Loop() {
	for {
		select {
		case msg := <-bs.incomingCh:
			go RouteIncomingMsg(bs, msg.(message))
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

func RouteIncomingMsg(bs Server, msg message) {
	handler := bs.GetHandler(msg.Channel)

	if handler == nil {
		fmt.Printf("Not Handler for Channel '%s'\n", msg.Channel)
		return
	}

	msgStruct := &MessageStruct{}
	err := json.Unmarshal(msg.Payload, msgStruct)

	msgStruct.ClientId = msg.ClientId

	if err != nil {
		panic(err)
	}

	handler(msgStruct)
}

func GenerateNewClientId() string {
	rand.Seed(time.Now().UnixNano())
	id := fmt.Sprintf("%d", rand.Int63())
	h := md5.New()

	h.Write([]byte(id))

	return fmt.Sprintf("%x", h.Sum(nil))
}

func Connect(bs Server, msg *MessageStruct) {
	client := bs.GetClient(msg.ClientId)
	if client == nil {
		panic("can't connect someone who does not exist?")
	}
	fmt.Printf("Do Connect\n%v\n", msg)
	fmt.Printf("For Client\n%v\n", client)

	client.SendMessage(&struct {
		Channel    string      `json:"channel"`
		Successful bool        `json:"successful"`
		Error      string      `json:"error"`
		ClientId   string      `json:"clientId"`
		Id         string      `json:"id"`
		Timestamp  string      `json:"timestamp"`
		Advice     interface{} `json:"advice"`
	}{
		msg.Channel,
		true,
		"",
		msg.ClientId,
		msg.Id,
		NewTimestamp().String(),
		struct {
			Reconnect string `json:"reconnect"`
			Interval  int    `json:"interval"`
		}{"retry", 60000},
	})
}

func Handshake(bs Server, msg *MessageStruct) {
	client := bs.GetClient(msg.ClientId)
	if client == nil {
		panic("can't handshake someone who does not exist?")
	}
	fmt.Printf("Do Handshake\n%v\n", msg)
	fmt.Printf("For Client\n%v\n", client)

	client.SendMessage(&struct {
		Channel                  string      `json:"channel"`
		Version                  string      `json:"version"`
		MinimumVersion           string      `json:"minimumVersion"`
		SupportedConnectionTypes []string    `json:"supportedConnectionTypes"`
		ClientId                 string      `json:"clientId"`
		Successful               bool        `json:"successful"`
		AuthSuccessful           bool        `json:"authSuccessful"`
		Advice                   interface{} `json:"advice"`
	}{
		msg.Channel,
		"1.0",
		"1.0",
		[]string{"websocket"},
		msg.ClientId,
		true,
		true,
		struct {
			Reconnect string `json:"reconnect"`
		}{"retry"},
	})
}
