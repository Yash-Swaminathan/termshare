package main

import (
	"crypto/rand"
	"encoding/hex"
	"flag"
	"log"
	"net/http"

	"github.com/gorilla/websocket"
)

var (
	addr    = flag.String("addr", ":8080", "http service address")
	hostKey = flag.String("host-key", "", "key that grants host (write) access; random if empty")
)

func main() {
	flag.Parse()

	key := *hostKey
	if key == "" {
		key = randomKey()
	}

	s, err := NewSession(key)
	if err != nil {
		log.Fatal("NewSession:", err)
	}
	session = s
	go session.run()
	go session.readPTY()

	http.Handle("/", http.FileServer(http.Dir("static")))
	http.HandleFunc("/ws", serveWS)

	log.Printf("termshare listening on %s", *addr)
	log.Printf("host  (can type): http://localhost%s/?key=%s", *addr, key)
	log.Printf("viewer (read-only): http://localhost%s/", *addr)
	if err := http.ListenAndServe(*addr, nil); err != nil {
		log.Fatal("ListenAndServe:", err)
	}
}

func randomKey() string {
	b := make([]byte, 8)
	if _, err := rand.Read(b); err != nil {
		log.Fatal("randomKey:", err)
	}
	return hex.EncodeToString(b)
}

var (
	upgrader = websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool { return true },
	}
	// session is the one shared terminal every /ws client views.
	session *Session
)

func serveWS(w http.ResponseWriter, req *http.Request) {
	conn, err := upgrader.Upgrade(w, req, nil)
	if err != nil {
		return
	}
	// Host access is granted by a matching ?key=; viewers get canWrite at register.
	isHost := session.hostKey != "" && req.URL.Query().Get("key") == session.hostKey
	client := &Client{conn: conn, send: make(chan outMsg, 256), isHost: isHost}
	client.canWrite.Store(isHost)
	session.register <- client
	go client.writePump()
	go client.readPump(session)
}
