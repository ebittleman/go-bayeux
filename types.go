package bayeux

import (
	"fmt"
	"log"
	"os"
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
	CLIENT_FLASH     = "flash"

	RECONNECT_RETRY     = "retry"
	RECONNECT_HANDSHAKE = "handshake"
	RECONNECT_NONE      = "none"
)

var logger *log.Logger

func init() {
	// f, err := os.OpenFile("log.log", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	// if err != nil {
	// 	panic(err)
	// }
	logger = log.New(os.Stdout, "go-bayux/client::", log.Ldate|log.Ltime)
}

var supportedClients = []string{CLIENT_WEBSOCKET}
var defaultInterval = 60000

type Event interface{}
type Envelope interface{}

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
