package messages

type Message interface{}

type RawMessage struct {
	Channel  string
	ClientId string
	Payload  []byte
}

type HandshakeRequest struct {
	Channel                  string   `json:"channel"`
	Version                  string   `json:"version"`
	MinimumVersion           string   `json:"minimumVersion"`
	SupportedConnectionTypes []string `json:"supportedConnectionTypes"`
	Advice                   struct {
		Timeout  int `json:"timeout"`
		Interval int `json:"interval"`
	} `json:"advice"`
	Id string `json:"id,omitempty"`
}

type HandshakeResponseAdvice struct {
	Reconnect string `json:"reconnect"`
	Interval  int    `json:"interval,omitempty"`
}

type HandshakeResponse struct {
	Channel                  string                   `json:"channel"`
	Version                  string                   `json:"version"`
	MinimumVersion           string                   `json:"minimumVersion"`
	SupportedConnectionTypes []string                 `json:"supportedConnectionTypes"`
	ClientId                 string                   `json:"clientId"`
	Successful               bool                     `json:"successful"`
	AuthSuccessful           bool                     `json:"authSuccessful"`
	Error                    string                   `json:"error,omitempty"`
	Id                       string                   `json:"id,omitempty"`
	Advice                   *HandshakeResponseAdvice `json:"advice,omitempty"`
}

type ConnectRequest struct {
	Channel        string `json:"channel"`
	ClientId       string `json:"clientId"`
	ConnectionType string `json:"connectionType"`
	Id             string `json:"id,omitempty"`
}

type ConnectAdvice struct {
	Reconnect string `json:"reconnect"`
	Interval  int    `json:"interval"`
}

type ConnectResponse struct {
	Channel    string         `json:"channel"`
	Successful bool           `json:"successful"`
	Error      string         `json:"error"`
	ClientId   string         `json:"clientId"`
	Timestamp  string         `json:"timestamp"`
	Id         string         `json:"id,omitempty"`
	Advice     *ConnectAdvice `json:"advice"`
}

type DisconnectRequest struct {
	Channel  string `json:"channel"`
	ClientId string `json:"clientId"`
	Id       string `json:"id,omitempty"`
}

type DisconnectResponse struct {
	Channel    string `json:"channel"`
	ClientId   string `json:"clientId"`
	Successful bool   `json:"successful"`
	Id         string `json:"id,omitempty"`
}

type SubscribeRequest struct {
	Channel      string `json:"channel"`
	ClientId     string `json:"clientId"`
	Subscription string `json:"subscription"`
	Id           string `json:"id,omitempty"`
}

type SubscribeResponse struct {
	Channel      string `json:"channel"`
	ClientId     string `json:"clientId"`
	Subscription string `json:"subscription"`
	Successful   bool   `json:"successful"`
	Error        string `json:"error,omitempty"`
	Timestamp    string `json:"timestamp"`
	Id           string `json:"id,omitempty"`
}

type UnsubscribeRequest SubscribeRequest
type UnsubscribeResponse SubscribeResponse

type PublishRequest struct {
	Channel  string      `json:"channel"`
	Data     interface{} `json:"data"`
	ClientId string      `json:"clientId"`
	Id       string      `json:"id,omitempty"`
}

type PublishResponse struct {
	Channel    string `json:"channel"`
	Successful bool   `json:"successful"`
	Error      string `json:"error,omitempty"`
	Id         string `json:"id,omitempty"`
}

type EventMessage struct {
	Channel string      `json:"channel"`
	Data    interface{} `json:"data"`
	Id      string      `json:"id,omitempty"`
}
