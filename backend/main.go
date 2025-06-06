package main

import (
	"flag"
	"log"
	"net/http"
)

// define address pointer
var addr = flag.String("addr", ":8080", "http server address")

func main() {
	flag.Parse()
	wsServer := NewWebsocketServer()
	// run as goroutine
	go wsServer.Run()
	http.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		// deal with client requests
		ServeWs(wsServer, w, r)
	})

	log.Fatal(http.ListenAndServe(*addr, nil))
}
