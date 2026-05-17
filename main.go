package main

import (
	"flag"
	"log"
	"net/http"

	"github.com/gorilla/websocket"
)

var addr = flag.String("addr", ":8080", "http service address")

func main() {
	flag.Parse()

	s, err := NewSession()
	if err != nil {
		log.Fatal("NewSession:", err)
	}
	session = s
	go session.run()
	go session.readPTY()

	http.Handle("/", http.FileServer(http.Dir("static")))

	// endpoint for the live terminal stream.
	http.HandleFunc("/ws", serveWS)

	log.Printf("termshare listening on %s", *addr)
	if err := http.ListenAndServe(*addr, nil); err != nil {
		log.Fatal("ListenAndServe:", err)
	}
}

var (
	upgrader = websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool { return true },
	}
	// session is the ONE shared terminal. Every browser that hits /ws
	// becomes a viewer of this same session — that is the "share".
	// Created once in main(), never per-connection.
	session *Session
)

func serveWS(w http.ResponseWriter, req *http.Request) {
	conn, err := upgrader.Upgrade(w, req, nil)
	if err != nil {
		return
	}
	client := &Client{conn: conn, send: make(chan []byte, 256)}
	session.register <- client
	go client.writePump()
	go client.readPump(session)
}
