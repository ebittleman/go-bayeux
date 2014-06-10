package main

import (
	"bytes"
	"html/template"
	_ "io/ioutil"
	"log"
	"net/http"
	"os"

	bayeux "github.com/ebittleman/go-bayeux"
)

var listenAddr = "localhost:8080"
var indexBytes []byte

func main() {

	logger := log.New(os.Stdout, "go-bayux::", log.Ldate|log.Ltime)

	rootHandler, err := GetRootHandler(logger)
	if err != nil {
		logger.Fatal(err)
	}

	mux := http.NewServeMux()

	mux.HandleFunc("/", rootHandler)
	mux.Handle("/ws/cometd", bayeux.Handler())

	logger.Printf("Starting webserver on %s", listenAddr)
	err = http.ListenAndServe(listenAddr, mux)
	if err != nil {
		logger.Fatal(err)
	}
}

func GetRootHandler(logger *log.Logger) (func(http.ResponseWriter, *http.Request), error) {
	buf := &bytes.Buffer{}

	tmpl, err := template.ParseFiles("./static/src/index.html")
	if err != nil {
		return nil, err
	}

	tmpl.Execute(buf, map[string]string{"jsUrl": "http://localhost:8888"})
	indexBytes = buf.Bytes()

	return func(w http.ResponseWriter, r *http.Request) {
		logger.Println("Serving Page")
		w.Write(indexBytes)
	}, nil
}
