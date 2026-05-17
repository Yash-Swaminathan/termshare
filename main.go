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

	// Serve the xterm.js frontend.
	http.Handle("/", http.FileServer(http.Dir("static")))

	// WebSocket endpoint for the live terminal stream.
	http.HandleFunc("/ws", serveWS)

	log.Printf("termshare listening on %s", *addr)
	if err := http.ListenAndServe(*addr, nil); err != nil {
		log.Fatal("ListenAndServe:", err)
	}
}

// upgrader turns an incoming HTTP request into a WebSocket connection.
// CheckOrigin returning true accepts any origin — fine for local/trusted
// use (see Claude.md "no auth"); tighten this before exposing it widely.
var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
}

// serveWS is the bridge: HTTP request -> WebSocket -> attached to the
// Session as a Client. This is the only "endpoint that connects both
// sides" — there is one HTTP server, not two.
func serveWS(w http.ResponseWriter, req *http.Request) {
	// TODO 1: conn, err := upgrader.Upgrade(w, req, nil)
	//         on err, just return (Upgrade already wrote the HTTP error).
	//
	// TODO 2: client := &Client{conn: conn, send: make(chan []byte, 256)}
	//
	// TODO 3: register the client with the session:
	//         session.register <- client
	//
	// TODO 4: go client.writePump()
	//         The ONLY goroutine allowed to call conn.WriteMessage.
	//         It ranges over client.send and writes each []byte as a
	//         websocket.BinaryMessage. (gorilla: one writer per conn.)
	//
	// TODO 5: client.readPump()   // blocks until the conn closes
	//         Loop conn.ReadMessage(); write the bytes into the pty
	//         (session.ptyFile.Write). Skip this write for read-only
	//         viewers — Claude.md says viewers are read-only by default.
	//         On read error: session.unregister <- client; return.
	//
	// Keep this http.Error until the steps above replace it, so the
	// build stays green.
	http.Error(w, "websocket handler not implemented yet", http.StatusNotImplemented)
	_ = upgrader
}
