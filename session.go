package main

import (
	"os"
	"sync"

	"github.com/gorilla/websocket"
)

// session.go is the hub. One Session owns one pty and the set of
// browser clients watching it. Model to copy: the gorilla/websocket
// chat example (hub.go + client.go) — termshare IS that example with
// the pty master file substituted for "messages users type".

// Client is one connected browser.
type Client struct {
	conn *websocket.Conn
	// send is buffered. Broadcasts push here; writePump is the ONLY
	// thing that drains it and writes to conn. This is how you obey
	// gorilla's "one concurrent writer per connection" rule.
	send chan []byte
}

func (c *Client) writePump() {
	for b := range c.send {
		c.conn.WriteMessage(websocket.BinaryMessage, b)
	}
}

// readPump forwards WebSocket input to the pty; unregisters on disconnect.
func (c *Client) readPump(s *Session) {
	defer func() { s.unregister <- c }()
	for {
		_, b, err := c.conn.ReadMessage()
		if err != nil {
			break
		}
		s.ptyFile.Write(b)
	}
}

// Session ties the pty to its clients.
type Session struct {
	ptyFile    *os.File
	clients    map[*Client]bool
	register   chan *Client
	unregister chan *Client
	broadcast  chan []byte // bytes read off the pty land here
	mu         sync.Mutex  // only if you guard `clients` directly
	// instead of doing everything through the channels above
}

// NewSession spawns a shell PTY and initializes the hub channels.
// Caller must run go s.run() and go s.readPTY() before registering clients.
func NewSession() (*Session, error) {
	f, _, err := StartPTY()
	if err != nil {
		return nil, err
	}
	return &Session{ptyFile: f, clients: map[*Client]bool{},
		register: make(chan *Client), unregister: make(chan *Client),
		broadcast: make(chan []byte, 256)}, nil
}

// run is the hub loop: register/unregister clients and fan out pty output.
func (s *Session) run() {
	for {
		select {
		case c := <-s.register:
			s.clients[c] = true
		case c := <-s.unregister:
			s.drop(c)
		case b := <-s.broadcast:
			for c := range s.clients {
				select {
				case c.send <- b:
				default:
					// Slow client. Drop inline — sending to
					// s.unregister here deadlocks, since this
					// goroutine is its only receiver.
					s.drop(c)
				}
			}
		}
	}
}

// drop removes a client exactly once. Safe to call twice: readPump's
// deferred unregister can fire after a slow-client drop, and the map
// check stops a double close(c.send) panic.
func (s *Session) drop(c *Client) {
	if _, ok := s.clients[c]; ok {
		delete(s.clients, c)
		close(c.send)
		c.conn.Close()
	}
}

// readPTY reads shell output and pushes copies onto broadcast (buf is reused).
func (s *Session) readPTY() {
	buf := make([]byte, 4096)
	for {
		n, err := s.ptyFile.Read(buf)
		if err != nil {
			return
		}
		data := make([]byte, n)
		copy(data, buf[:n])
		s.broadcast <- data
	}
}
