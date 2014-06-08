package main

import (
	"io/ioutil"
	"log"
	"net/http"
	"os"

	bayeux "github.com/ebittleman/go-bayeux"
)

var listenAddr = "localhost:8080"
var indexBytes []byte

func main() {

	logger := log.New(os.Stdout, "go-bayux::", log.Ldate|log.Ltime)
	index, err := ioutil.ReadFile("./static/src/index.html")
	indexBytes = index

	bayeuxServer := bayeux.NewServer()

	if err != nil {
		logger.Fatal(err)
	}

	http.HandleFunc("/", rootHandler)
	bayeuxServer.Bind("/ws/cometd")

	logger.Printf("Starting webserver on %s", listenAddr)
	err = http.ListenAndServe(listenAddr, nil)
	logger.Printf("Started webserver on %s", listenAddr)
	if err != nil {
		logger.Fatal(err)
	}

}

func rootHandler(w http.ResponseWriter, r *http.Request) {
	logger := log.New(os.Stdout, "go-bayux", log.Ldate|log.Ltime)
	logger.Println("Serving Page")
	w.Write(indexBytes)
}
