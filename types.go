package bayeux

import (
	"fmt"
	"time"
)

const (
	STATE_UNCONNECTED = 0
	STATE_CONNECTING  = 1
	STATE_CONNECTED   = 2

	CLIENT_WEBSOCKET = "websocket"
	CLIENT_LONGPOLL  = "long-polling"
	CLIENT_CALLBACK  = "callback-polling"
	CLIENT_IFRAME    = "iframe"
	CLIENT_FLASH     = "flase"

	RECONNECT_RETRY     = "retry"
	RECONNECT_HANDSHAKE = "handshake"
	RECONNECT_NONE      = "none"
)

var supportedClients = []string{CLIENT_LONGPOLL}

type BayeuxHandler func(*MessageStruct)

type message struct {
	Channel  string
	ClientId string
	Payload  []byte
}

type request struct {
	Message Message
	Payload []byte
}
type handshakeRequest struct {
	request
}
type connectRequest struct {
	request
}
type Request interface {
	GetMessage() Message
}

type response struct {
	Message Message
	Success bool
	Response
}
type handshakeResponse struct {
	response
}
type connectResponse struct {
	response
}
type Response interface {
	GetMesage() Message
}

type Advice struct {
	Reconnect       string   `json:"reconnect"`
	Interval        int      `json:"interval,omitempty"`
	MultipleClients bool     `json:"multiple-clients,omitempty"`
	Hosts           []string `json:"hosts,omitempty"`
	CallbackPolling struct {
		Reconnect string `json:"reconnect"`
	} `json:"callback-polling,omitempty"`
}

type MessageStruct struct {
	Channel                  string                 `json:"channel"`
	ClientId                 string                 `json:"clientId,omitempty"`
	Successful               bool                   `json:"successful,omitempty"`
	Error                    string                 `json:"error,omitempty"`
	Advice                   Advice                 `json:"advice,omitempty"`
	Data                     map[string]interface{} `json:"data,omitempty"`
	Id                       string                 `json:"id,omitempty"`
	Ext                      interface{}            `json:"ext,omitempty"`
	Version                  string                 `json:"version,omitempty"`
	MinimumVersion           string                 `json:"minimumVersion,omitempty"`
	SupportedConnectionTypes []string               `json:"supportedConnectionTypes,omitempty"`
	AuthSuccessful           bool                   `json:"authSuccessful,omitempty"`
	ConnectionType           string                 `json:"connectionType,omitempty"`
	Timestamp                string                 `json:"timestamp,omitempty"`
}

type Message interface {
	// GetEvent() Event
}

type Event interface{}
type Envelope interface{}
type Channel interface {
	GetName() string
}
type ChannelSegment interface {
	GetToken() string
}
type ChanelSegments []ChannelSegment
type Connection interface{}

type Timestamp struct {
	Time time.Time
}

func NewTimestamp() *Timestamp {
	return &Timestamp{time.Now()}
}

func (t *Timestamp) String() string {
	utc := t.Time.UTC()
	pre := int((float32(utc.Nanosecond()) / 1000000000.0) * 100)

	return fmt.Sprintf("%s.%d", utc.Format("2006-01-02T15:04:05"), pre)
}
