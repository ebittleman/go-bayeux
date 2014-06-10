package bayeux

// func (bs *bayeuxServer) Bind(path string) {
//  http.Handle(path)

// http.HandleFunc(path+"/handshake", func(resp http.ResponseWriter, req *http.Request) {
//  client := NewLongPollClient(GenerateNewClientId(), resp, req, bs)
//  bs.RegisterClient(client.GetId(), client)

//  msgs := []*messages.HandshakeRequest{}

//  err := ParseReqBody(req.Body, &msgs)
//  if err != nil {
//      http.Error(resp, "Parse Error, bad request", 406)
//      return
//  }

//  handshakeRequest := msgs[0]

//  Handshake(bs, client.GetId(), handshakeRequest)
// })

// http.HandleFunc(path+"/connect", func(resp http.ResponseWriter, req *http.Request) {
//  msgs := []*messages.ConnectRequest{}

//  err := ParseReqBody(req.Body, &msgs)
//  if err != nil {
//      http.Error(resp, "Parse Error, bad request", 406)
//      return
//  }

//  connectRequest := msgs[0]

//  //TODO Implement Connect

//  Connect(bs, connectRequest)

// })
// http.HandleFunc(path+"/", func(resp http.ResponseWriter, req *http.Request) {
//  <-time.After(time.Minute)
// })
// }
